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
)

// TestLiveProductHelpGrounding is an opt-in provider contract probe. It does
// not call the application orchestrator, read owner data, or persist anything:
// the handbook result is synthetic and is supplied as a tool result to the
// configured OpenAI-compatible provider.
func TestLiveProductHelpGrounding(t *testing.T) {
	if os.Getenv("AI_RAG_LIVE") != "1" {
		t.Skip("set AI_RAG_LIVE=1 to exercise live product-help grounding")
	}
	if err := config.LoadLocalEnv(filepath.Join("..", "..", "..", ".env.local")); err != nil {
		t.Fatal("load local provider configuration:", err)
	}
	cfg := config.Load()
	client := NewClient(Config{BaseURL: cfg.AIBaseURL, APIKey: cfg.AIAPIKey, Model: cfg.AIModel})
	if !client.Configured() {
		t.Fatal("AI provider configuration is incomplete")
	}

	tool := map[string]any{
		"type": "function",
		"function": map[string]any{
			"name":        "get_product_help",
			"description": "Search the versioned My Pocket handbook; this tool has no financial data and no side effects.",
			"strict":      true,
			"parameters": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query":       map[string]any{"type": "string"},
					"max_results": map[string]any{"type": "integer", "minimum": 1, "maximum": 3},
				},
				"required":             []string{"query", "max_results"},
				"additionalProperties": false,
			},
		},
	}
	messages := []map[string]any{
		{"role": "system", "content": "You answer in Vietnamese. For product navigation questions, call get_product_help before answering. Never invent finance data or claim an action was performed."},
		{"role": "user", "content": "Tôi mở phần Trợ lý tài chính ở đâu?"},
	}
	request := map[string]any{
		"model":       cfg.AIModel,
		"temperature": 0,
		"max_tokens":  300,
		"messages":    messages,
		"tools":       []any{tool},
		"tool_choice": map[string]any{"type": "function", "function": map[string]string{"name": "get_product_help"}},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	started := time.Now()
	body, _, err := client.request(ctx, "POST", cfg.AIBaseURL+"/chat/completions", cfg.AIAPIKey, request, "AI product-help grounding probe")
	if err != nil {
		t.Fatal("live product-help tool call:", err)
	}
	var first struct {
		Choices []struct {
			Message struct {
				ToolCalls []struct {
					ID       string `json:"id"`
					Type     string `json:"type"`
					Function struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					} `json:"function"`
				} `json:"tool_calls"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &first); err != nil || len(first.Choices) != 1 || len(first.Choices[0].Message.ToolCalls) != 1 {
		t.Fatalf("expected one product-help tool call, response_bytes=%d err=%v", len(body), err)
	}
	call := first.Choices[0].Message.ToolCalls[0]
	if call.Function.Name != "get_product_help" || call.ID == "" {
		t.Fatalf("unexpected product-help call: %#v", call)
	}
	var arguments struct {
		Query      string `json:"query"`
		MaxResults int    `json:"max_results"`
	}
	if err := json.Unmarshal([]byte(call.Function.Arguments), &arguments); err != nil || strings.TrimSpace(arguments.Query) == "" || arguments.MaxResults < 1 || arguments.MaxResults > 3 {
		t.Fatalf("invalid product-help arguments: %q", call.Function.Arguments)
	}

	toolResult := `{"status":"ok","version":"1.1","sources":[{"id":"handbook.navigation.v1.1","title":"Điều hướng","excerpt":"Trợ lý là tab chính ở thanh điều hướng phía dưới. Tài khoản chứa các thiết lập và tính năng phụ."}]}`
	followUp := map[string]any{
		"model":       cfg.AIModel,
		"temperature": 0,
		"max_tokens":  250,
		"messages": []any{
			messages[0], messages[1],
			map[string]any{"role": "assistant", "content": nil, "tool_calls": []any{map[string]any{"id": call.ID, "type": call.Type, "function": map[string]string{"name": call.Function.Name, "arguments": call.Function.Arguments}}}},
			map[string]any{"role": "tool", "tool_call_id": call.ID, "content": toolResult},
		},
	}
	answerBody, _, err := client.request(ctx, "POST", cfg.AIBaseURL+"/chat/completions", cfg.AIAPIKey, followUp, "AI product-help answer probe")
	if err != nil {
		t.Fatal("live product-help grounded answer:", err)
	}
	var answer struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(answerBody, &answer); err != nil || len(answer.Choices) != 1 || strings.TrimSpace(answer.Choices[0].Message.Content) == "" {
		t.Fatalf("empty grounded answer, response_bytes=%d err=%v", len(answerBody), err)
	}
	content := answer.Choices[0].Message.Content
	if !strings.Contains(strings.ToLower(content), "trợ lý") {
		t.Fatalf("grounded answer did not mention the requested product surface: %q", content)
	}
	t.Logf("product_help_grounding pass=true tool=%s answer_bytes=%d first_latency=%s", call.Function.Name, len(content), time.Since(started).Round(time.Millisecond))
}
