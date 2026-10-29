package ai

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mypocket/backend/internal/config"
	"github.com/mypocket/backend/internal/usecase"
)

type observedReviewProvider struct {
	provider usecase.AdvisorProvider
	t        *testing.T
}

func (p observedReviewProvider) Chat(ctx context.Context, request usecase.AdvisorRequest) (usecase.AdvisorResponse, error) {
	response, err := p.provider.Chat(ctx, request)
	for _, call := range response.ToolCalls {
		if call.Name == "finish_answer" {
			var shape struct {
				Mode, TemplateID, SourceID string
				Claims                     []struct {
					FactRef string `json:"fact_ref"`
				} `json:"claims"`
			}
			var raw map[string]json.RawMessage
			_ = json.Unmarshal(call.Arguments, &raw)
			_ = json.Unmarshal(raw["mode"], &shape.Mode)
			_ = json.Unmarshal(raw["template_id"], &shape.TemplateID)
			_ = json.Unmarshal(raw["claims"], &shape.Claims)
			p.t.Logf("synthetic final mode=%s template=%s claim_count=%d", shape.Mode, shape.TemplateID, len(shape.Claims))
		} else {
			p.t.Logf("synthetic tool=%s", call.Name)
		}
	}
	return response, err
}

// Opt-in synthetic provider acceptance for the new terminal protocol. It has
// no DB, mutation handler or production owner; no raw answers are logged.
func TestLiveAdvisorReviewGroundedProtocol(t *testing.T) {
	if os.Getenv("AI_ADVISOR_REVIEW_LIVE") != "1" {
		t.Skip("requires AI_ADVISOR_REVIEW_LIVE=1 and approved provider")
	}
	if err := config.LoadLocalEnv(filepath.Join("..", "..", "..", ".env.local")); err != nil {
		t.Fatal("load provider config")
	}
	cfg := config.Load()
	client := NewAdvisorClient(AdvisorClientConfig{BaseURL: cfg.AIBaseURL, APIKey: cfg.AIAPIKey, Model: cfg.AIModel})
	if !client.Available() {
		t.Fatal("provider unavailable")
	}
	for _, tc := range []struct {
		name, prompt string
		check        func(*testing.T, usecase.AdvisorAnswer)
	}{
		{"finance", "Tháng 9 năm 2026 tôi đã chi bao nhiêu? Đọc công cụ và trả tổng chi từ evidence.", func(t *testing.T, a usecase.AdvisorAnswer) {
			if !strings.Contains(a.Text, "1234500") || len(a.Results) == 0 {
				t.Fatal("financial final did not render synthetic evidence")
			}
		}},
		{"handbook", "Tôi tạo ngân sách ở màn nào? Tra handbook và hướng dẫn tôi.", func(t *testing.T, a usecase.AdvisorAnswer) {
			if !strings.Contains(a.Text, "Tài khoản > Ngân sách") {
				t.Fatal("help did not render handbook source")
			}
		}},
		{"proposal", "Chuẩn bị đề xuất tạo ví thường tên Ví thử nghiệm, số dư ban đầu 0; tôi sẽ xác nhận trên thẻ.", func(t *testing.T, a usecase.AdvisorAnswer) {
			found := false
			for _, r := range a.Results {
				if r.Status == "proposed" && r.ViewKind == "wallet_creation_proposal" {
					found = true
				}
			}
			if !found || !strings.Contains(a.Text, "xác nhận") {
				t.Fatal("proposal did not preserve confirmation boundary")
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reader := &qualityFinanceReader{}
			o := usecase.NewAdvisorOrchestrator(observedReviewProvider{client, t}, usecase.NewAdvisorToolRegistry(usecase.NewFinanceQueryService(reader)))
			o.GroundingEnabled = true
			ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
			defer cancel()
			answer, err := o.Answer(ctx, usecase.Principal{OwnerID: "synthetic-review-owner", Timezone: "Asia/Ho_Chi_Minh"}, []usecase.AdvisorChatMessage{{Role: "user", Content: tc.prompt}})
			if err != nil {
				t.Fatalf("grounded protocol failed: %v", err)
			}
			tc.check(t, answer)
			for _, owner := range reader.owners {
				if owner != "synthetic-review-owner" {
					t.Fatal("owner scope changed")
				}
			}
			t.Logf("model=%s query_count=%d result_count=%d", cfg.AIModel, len(reader.queries), len(answer.Results))
		})
	}
}
