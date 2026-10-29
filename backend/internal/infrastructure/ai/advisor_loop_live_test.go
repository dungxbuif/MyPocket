package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mypocket/backend/internal/config"
	"github.com/mypocket/backend/internal/usecase"
)

// Guard the evaluation before dispatch: a nil finance dependency must never
// receive a model-selected read, and no live owner data is available here.
type evalRecordingProvider struct {
	provider usecase.AdvisorProvider
	allowed  string
	calls    []usecase.AdvisorToolCall
}

func (p *evalRecordingProvider) Chat(ctx context.Context, request usecase.AdvisorRequest) (usecase.AdvisorResponse, error) {
	response, err := p.provider.Chat(ctx, request)
	if err != nil {
		return response, err
	}
	for _, call := range response.ToolCalls {
		if call.Name != p.allowed {
			return usecase.AdvisorResponse{}, fmt.Errorf("evaluation rejected unexpected tool %q", call.Name)
		}
		p.calls = append(p.calls, call)
	}
	return response, nil
}

type evalResponseProvider struct{ response usecase.AdvisorResponse }

func (p evalResponseProvider) Chat(context.Context, usecase.AdvisorRequest) (usecase.AdvisorResponse, error) {
	return p.response, nil
}

func TestAdvisorEvalRejectsUnexpectedToolBeforeDispatch(t *testing.T) {
	provider := &evalRecordingProvider{
		provider: evalResponseProvider{usecase.AdvisorResponse{ToolCalls: []usecase.AdvisorToolCall{
			{ID: "synthetic", Name: "get_wallet_balances", Arguments: json.RawMessage(`{}`)},
		}}},
		allowed: "get_product_help",
	}
	orchestrator := usecase.NewAdvisorOrchestrator(provider, usecase.NewAdvisorToolRegistry(nil))
	_, err := orchestrator.Answer(context.Background(), usecase.Principal{OwnerID: "synthetic-owner", Timezone: "Asia/Ho_Chi_Minh"},
		[]usecase.AdvisorChatMessage{{Role: "user", Content: "Trợ lý ở đâu?"}})
	if err == nil || !strings.Contains(err.Error(), "evaluation rejected") {
		t.Fatalf("unexpected finance tool must be rejected before dispatch: %v", err)
	}
	if len(provider.calls) != 0 {
		t.Fatal("rejected call must not be recorded as accepted")
	}
}

func TestLiveAdvisorApplicationLoop(t *testing.T) {
	if os.Getenv("AI_ADVISOR_LOOP_LIVE") != "1" {
		t.Skip("set AI_ADVISOR_LOOP_LIVE=1 for synthetic real-orchestrator evaluation")
	}
	if err := config.LoadLocalEnv(filepath.Join("..", "..", "..", ".env.local")); err != nil {
		t.Fatal("load local provider configuration")
	}
	cfg := config.Load()
	client := NewAdvisorClient(AdvisorClientConfig{BaseURL: cfg.AIBaseURL, APIKey: cfg.AIAPIKey, Model: cfg.AIModel})
	if !client.Available() {
		t.Fatal("provider configuration unavailable")
	}
	cases := []struct {
		name, prompt, tool, part string
		fields                   map[string]any
	}{
		{"navigation", "Tôi mở phần Trợ lý tài chính ở đâu trong app?", "get_product_help", "product_help", nil},
		{"jar", "Đề xuất tạo hũ tên Cà phê cho tháng 2026-09, phân bổ cố định 500000 đồng. Tôi sẽ xác nhận trên thẻ.", "propose_create_jar", "jar_confirmation",
			map[string]any{"name": "Cà phê", "month": "2026-09", "allocation_mode": "fixed", "allocation_amount": float64(500000)}},
		{"wallet", "Đề xuất tạo ví thường tên Tiền mặt thử nghiệm, số dư ban đầu 0 đồng và tính vào tổng cộng. Tôi sẽ xác nhận trên thẻ.", "propose_create_wallet", "wallet_confirmation",
			map[string]any{"name": "Tiền mặt thử nghiệm", "type": "basic", "opening_balance": float64(0), "is_in_total": true}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
			defer cancel()
			recorder := &evalRecordingProvider{provider: client, allowed: tc.tool}
			orchestrator := usecase.NewAdvisorOrchestrator(recorder, usecase.NewAdvisorToolRegistry(nil))
			answer, err := orchestrator.Answer(ctx, usecase.Principal{OwnerID: "synthetic-owner", Timezone: "Asia/Ho_Chi_Minh"},
				[]usecase.AdvisorChatMessage{{Role: "user", Content: tc.prompt}})
			if err != nil {
				t.Fatalf("application loop failed: %v", err)
			}
			if len(recorder.calls) != 1 || len(answer.Results) != 1 || strings.TrimSpace(answer.Text) == "" {
				t.Fatalf("expected one relevant tool result and final text; calls=%d results=%d", len(recorder.calls), len(answer.Results))
			}
			var view map[string]any
			if err := json.Unmarshal(answer.Results[0].View, &view); err != nil {
				t.Fatal("invalid structured result")
			}
			for key, expected := range tc.fields {
				if view[key] != expected {
					t.Errorf("synthetic field %s: got %v want %v", key, view[key], expected)
				}
			}
			if tc.name == "navigation" {
				sources, ok := view["sources"].([]any)
				if !ok || len(sources) == 0 || !strings.Contains(strings.ToLower(answer.Text), "trợ lý") {
					t.Error("navigation requires sources and relevant answer")
				}
			}
			found := false
			for _, part := range usecase.BuildAdvisorParts(answer) {
				if part.Type == tc.part {
					found = true
				}
			}
			if !found {
				t.Errorf("missing structured part %s", tc.part)
			}
			t.Logf("model=%s tool=%s part=%s", cfg.AIModel, tc.tool, tc.part)
		})
	}
}
