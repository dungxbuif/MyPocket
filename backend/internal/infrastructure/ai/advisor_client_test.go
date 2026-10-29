package ai

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mypocket/backend/internal/usecase"
)

func TestAdvisorClientRoundTripsToolCall(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer secret" {
			t.Fatalf("provider key missing")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":null,"tool_calls":[{"id":"call-1","type":"function","function":{"name":"get_finance_summary","arguments":"{\"range\":{\"from\":\"2026-09-01\",\"to\":\"2026-09-30\"}}"}}]}}]}`))
	}))
	defer server.Close()
	client := NewAdvisorClient(AdvisorClientConfig{BaseURL: server.URL, APIKey: "secret", Model: "test"})
	response, err := client.Chat(context.Background(), usecase.AdvisorRequest{Messages: []usecase.AdvisorChatMessage{{Role: "user", Content: "summary"}}, Tools: []usecase.ToolDefinition{{Name: "get_finance_summary"}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(response.ToolCalls) != 1 || response.ToolCalls[0].Name != "get_finance_summary" {
		t.Fatalf("unexpected tool response: %+v", response)
	}
}

func TestAdvisorClientEncodesOpenAICompatibleToolCalls(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Messages []struct {
				Role       string `json:"role"`
				ToolCallID string `json:"tool_call_id"`
				ToolCalls  []struct {
					ID       string `json:"id"`
					Type     string `json:"type"`
					Function struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					} `json:"function"`
				} `json:"tool_calls"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if len(body.Messages) != 3 || len(body.Messages[1].ToolCalls) != 1 {
			t.Fatalf("unexpected message wire shape: %+v", body.Messages)
		}
		call := body.Messages[1].ToolCalls[0]
		if call.Type != "function" || call.Function.Name != "get_finance_summary" || call.ID != "call-1" {
			t.Fatalf("tool call must use OpenAI function shape: %+v", call)
		}
		if body.Messages[2].Role != "tool" || body.Messages[2].ToolCallID != "call-1" {
			t.Fatalf("tool result must retain tool_call_id: %+v", body.Messages[2])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"ok"}}]}`))
	}))
	defer server.Close()
	client := NewAdvisorClient(AdvisorClientConfig{BaseURL: server.URL, APIKey: "secret", Model: "test"})
	_, err := client.Chat(context.Background(), usecase.AdvisorRequest{Messages: []usecase.AdvisorChatMessage{
		{Role: "user", Content: "summary"},
		{Role: "assistant", ToolCalls: []usecase.AdvisorToolCall{{ID: "call-1", Name: "get_finance_summary", Arguments: json.RawMessage(`{"range":{"from":"2026-09-01","to":"2026-09-30"}}`)}}},
		{Role: "tool", ToolCallID: "call-1", Content: `{"expense":150}`},
	}})
	if err != nil {
		t.Fatal(err)
	}
}

func TestAdvisorClientRejectsMissingConfiguration(t *testing.T) {
	client := NewAdvisorClient(AdvisorClientConfig{})
	if _, err := client.Chat(context.Background(), usecase.AdvisorRequest{}); err == nil {
		t.Fatal("unconfigured advisor provider must fail closed")
	}
}

func TestAdvisorClientMapsUpstreamFailureToProviderUnavailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"detail":"upstream credentials or body must never escape"}`))
	}))
	defer server.Close()
	client := NewAdvisorClient(AdvisorClientConfig{BaseURL: server.URL, APIKey: "secret", Model: "test"})
	_, err := client.Chat(context.Background(), usecase.AdvisorRequest{Messages: []usecase.AdvisorChatMessage{{Role: "user", Content: "summary"}}})
	if !errors.Is(err, usecase.ErrAdvisorProviderUnavailable) {
		t.Fatalf("upstream failure must map to provider unavailable, got %v", err)
	}
	if strings.Contains(err.Error(), "credentials") || strings.Contains(err.Error(), server.URL) {
		t.Fatalf("provider failure leaked sensitive detail: %v", err)
	}
}

func TestAdvisorClientNormalizesBoundedUsageMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"provider-request-1","choices":[{"finish_reason":"tool_calls","message":{"content":"ok"}}],"usage":{"prompt_tokens":12,"completion_tokens":4,"total_tokens":16,"secret":"do-not-store","prompt_tokens_details":{"cached_tokens":7}}}`))
	}))
	defer server.Close()

	client := NewAdvisorClient(AdvisorClientConfig{BaseURL: server.URL, APIKey: "secret", Model: "advisor-model", StoreUsage: true})
	response, err := client.Chat(context.Background(), usecase.AdvisorRequest{Messages: []usecase.AdvisorChatMessage{{Role: "user", Content: "summary"}}})
	if err != nil {
		t.Fatal(err)
	}
	if response.Usage == nil {
		t.Fatal("usage must be retained when StoreUsage is enabled")
	}
	if response.Usage["provider"] != "openai_compatible" || response.Usage["model"] != "advisor-model" || response.Usage["request_id"] != "provider-request-1" || response.Usage["finish_reason"] != "tool_calls" {
		t.Fatalf("bounded provider metadata missing: %#v", response.Usage)
	}
	for key, want := range map[string]int64{"prompt_tokens": 12, "completion_tokens": 4, "total_tokens": 16} {
		if got := advisorUsageInt(response.Usage, key); got != want {
			t.Fatalf("%s=%d, want %d: %#v", key, got, want, response.Usage)
		}
	}
	if _, ok := response.Usage["secret"]; ok {
		t.Fatal("unknown provider usage field must not be persisted")
	}
	if _, ok := response.Usage["prompt_tokens_details"]; ok {
		t.Fatal("nested provider usage details must not be persisted")
	}
}

func TestAdvisorClientStoreUsageOptOut(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"provider-request-2","choices":[{"finish_reason":"stop","message":{"content":"ok"}}],"usage":{"prompt_tokens":12,"completion_tokens":4,"total_tokens":16}}`))
	}))
	defer server.Close()

	client := NewAdvisorClient(AdvisorClientConfig{BaseURL: server.URL, APIKey: "secret", Model: "advisor-model", StoreUsage: false})
	response, err := client.Chat(context.Background(), usecase.AdvisorRequest{Messages: []usecase.AdvisorChatMessage{{Role: "user", Content: "summary"}}})
	if err != nil {
		t.Fatal(err)
	}
	if response.Usage != nil {
		t.Fatalf("usage must be omitted when StoreUsage is disabled: %#v", response.Usage)
	}
}
