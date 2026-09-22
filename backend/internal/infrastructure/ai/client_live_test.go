package ai

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mypocket/backend/internal/config"
	"github.com/mypocket/backend/internal/entity"
)

func TestLiveExtractionWhenExplicitlyEnabled(t *testing.T) {
	if os.Getenv("AI_LIVE_TEST") != "1" {
		t.Skip("set AI_LIVE_TEST=1 to exercise the configured text model")
	}
	if err := config.LoadLocalEnv(filepath.Join("..", "..", "..", ".env.local")); err != nil {
		t.Fatal(err)
	}
	cfg := config.Load()
	c := NewClient(Config{BaseURL: cfg.AIBaseURL, APIKey: cfg.AIAPIKey, Model: cfg.AIModel})
	out, err := c.Extract(context.Background(), Input{Text: "Tôi vừa ăn trưa hết 35k bằng Cash.", Timezone: "Asia/Ho_Chi_Minh", Now: time.Date(2026, 9, 21, 10, 0, 0, 0, time.FixedZone("ICT", 7*3600)), Wallets: []entity.Wallet{{ID: "cash", Name: "Cash", Type: entity.WalletTypeBasic, Currency: "VND"}}, Categories: []entity.Category{{ID: "food", Name: "Ăn uống", Kind: entity.TransactionTypeExpense}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Drafts) != 1 || out.Drafts[0].Type != entity.TransactionTypeExpense || out.Drafts[0].Amount != 35000 || out.Drafts[0].WalletID != "cash" {
		t.Fatalf("unexpected live draft: %#v", out.Drafts)
	}
}

// TestLiveToolCallingCompatibility is a read-only capability probe. It sends
// one synthetic expense and never executes or persists the proposed function.
func TestLiveToolCallingCompatibility(t *testing.T) {
	if os.Getenv("AI_TOOL_CALL_PROBE") != "1" {
		t.Skip("set AI_TOOL_CALL_PROBE=1 to probe OpenAI-compatible tool calling")
	}
	if err := config.LoadLocalEnv(filepath.Join("..", "..", "..", ".env.local")); err != nil {
		t.Fatal("load local provider configuration")
	}
	cfg := config.Load()
	client := NewClient(Config{BaseURL: cfg.AIBaseURL, APIKey: cfg.AIAPIKey, Model: cfg.AIModel})
	if !client.Configured() {
		t.Fatal("AI provider configuration is incomplete")
	}

	draftSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"type":                map[string]any{"type": "string", "enum": []string{"income", "expense", "transfer", "unknown"}},
			"amount":              map[string]any{"type": "integer", "minimum": 0},
			"wallet_id":           map[string]any{"type": "string"},
			"category_id":         map[string]any{"type": []string{"string", "null"}},
			"occurred_at":         map[string]any{"type": "string"},
			"note":                map[string]any{"type": "string"},
			"included_in_reports": map[string]any{"type": "boolean"},
			"questions":           map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
		},
		"required":             []string{"type", "amount", "wallet_id", "category_id", "occurred_at", "note", "included_in_reports", "questions"},
		"additionalProperties": false,
	}
	resultSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"reply":  map[string]any{"type": "string"},
			"drafts": map[string]any{"type": "array", "items": draftSchema},
		},
		"required":             []string{"reply", "drafts"},
		"additionalProperties": false,
	}
	request := map[string]any{
		"model": cfg.AIModel,
		"messages": []map[string]string{
			{"role": "system", "content": "Produce a transaction proposal only. Call propose_transactions exactly once. Never write or confirm a transaction."},
			{"role": "user", "content": "Today I spent 35000 VND for lunch from wallet cash; category food. Return one expense draft."},
		},
		"tools": []map[string]any{{
			"type": "function",
			"function": map[string]any{
				"name":        "propose_transactions",
				"description": "Return reviewable transaction drafts only; this function has no side effects.",
				"strict":      true,
				"parameters":  resultSchema,
			},
		}},
		"tool_choice": map[string]any{"type": "function", "function": map[string]string{"name": "propose_transactions"}},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	started := time.Now()
	body, _, err := client.request(ctx, "POST", cfg.AIBaseURL+"/chat/completions", cfg.AIAPIKey, request, "AI tool probe")
	if err != nil {
		t.Logf("tool_call_probe transport_or_capability_error=%q latency=%s", err.Error(), time.Since(started).Round(time.Millisecond))
		return
	}
	var envelope struct {
		Choices []struct {
			FinishReason string `json:"finish_reason"`
			Message      struct {
				Content   string `json:"content"`
				ToolCalls []struct {
					Type     string `json:"type"`
					Function struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					} `json:"function"`
				} `json:"tool_calls"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil || len(envelope.Choices) != 1 {
		t.Logf("tool_call_probe envelope=invalid response_bytes=%d latency=%s", len(body), time.Since(started).Round(time.Millisecond))
		return
	}
	choice := envelope.Choices[0]
	var candidate string
	toolName := ""
	for _, call := range choice.Message.ToolCalls {
		toolName = call.Function.Name
		if call.Function.Name == "propose_transactions" {
			candidate = call.Function.Arguments
		}
	}
	if candidate == "" {
		candidate = choice.Message.Content
	}
	_, schemaErr := parseOutput([]byte(candidate))
	if mismatch := new(SchemaMismatchError); errors.As(schemaErr, &mismatch) {
		issues, _ := json.Marshal(mismatch.Issues)
		t.Logf("tool_call_probe finish_reason=%s tool_call_count=%d tool_name=%s schema_valid=false issues=%s response_bytes=%d candidate_bytes=%d prompt_tokens=%d completion_tokens=%d latency=%s", choice.FinishReason, len(choice.Message.ToolCalls), toolName, issues, len(body), len(candidate), envelope.Usage.PromptTokens, envelope.Usage.CompletionTokens, time.Since(started).Round(time.Millisecond))
		return
	}
	t.Logf("tool_call_probe finish_reason=%s tool_call_count=%d tool_name=%s schema_valid=%t response_bytes=%d candidate_bytes=%d prompt_tokens=%d completion_tokens=%d latency=%s", choice.FinishReason, len(choice.Message.ToolCalls), toolName, schemaErr == nil, len(body), len(candidate), envelope.Usage.PromptTokens, envelope.Usage.CompletionTokens, time.Since(started).Round(time.Millisecond))
}

// TestLiveJSONSchemaCompatibility checks the OpenAI-compatible structured
// output mode independently from MyPocket's production JSON-object prompt.
// It sends synthetic data and only logs response shape, never values.
func TestLiveJSONSchemaCompatibility(t *testing.T) {
	if os.Getenv("AI_JSON_SCHEMA_PROBE") != "1" {
		t.Skip("set AI_JSON_SCHEMA_PROBE=1 to probe OpenAI-compatible JSON Schema output")
	}
	if err := config.LoadLocalEnv(filepath.Join("..", "..", "..", ".env.local")); err != nil {
		t.Fatal("load local provider configuration")
	}
	cfg := config.Load()
	client := NewClient(Config{BaseURL: cfg.AIBaseURL, APIKey: cfg.AIAPIKey, Model: cfg.AIModel})
	if !client.Configured() {
		t.Fatal("AI provider configuration is incomplete")
	}
	client.http.Timeout = 60 * time.Second

	draftSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"type":                map[string]any{"type": "string", "enum": []string{"income", "expense", "transfer", "unknown"}},
			"amount":              map[string]any{"type": "integer", "minimum": 0},
			"wallet_id":           map[string]any{"type": "string"},
			"category_id":         map[string]any{"type": []string{"string", "null"}},
			"occurred_at":         map[string]any{"type": "string"},
			"note":                map[string]any{"type": "string"},
			"included_in_reports": map[string]any{"type": "boolean"},
			"questions":           map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
		},
		"required":             []string{"type", "amount", "wallet_id", "category_id", "occurred_at", "note", "included_in_reports", "questions"},
		"additionalProperties": false,
	}
	resultSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"reply":  map[string]any{"type": "string"},
			"drafts": map[string]any{"type": "array", "items": draftSchema},
		},
		"required":             []string{"reply", "drafts"},
		"additionalProperties": false,
	}
	type evalCase struct {
		name, text, wantType, wantWallet, wantCategory, datePrefix string
		amounts                                                    []int64
		needsQuestion                                              bool
		walletMustBeEmpty                                          bool
	}
	caseWallets := []entity.Wallet{
		{ID: "wallet-cash", Name: "Ví tiền mặt", Type: entity.WalletTypeBasic, Currency: "VND"},
		{ID: "wallet-bank", Name: "Tài khoản ngân hàng", Type: entity.WalletTypeBasic, Currency: "VND"},
	}
	caseCategories := []entity.Category{
		{ID: "category-food", Name: "Ăn uống", Kind: entity.TransactionTypeExpense},
		{ID: "category-coffee", Name: "Cà phê", Kind: entity.TransactionTypeExpense},
		{ID: "category-salary", Name: "Lương", Kind: entity.TransactionTypeIncome},
	}
	cases := []evalCase{
		{name: "single expense", text: "Hôm nay tôi ăn trưa hết 35.000đ bằng Ví tiền mặt.", wantType: "expense", wantWallet: "wallet-cash", wantCategory: "category-food", datePrefix: "2026-09-22", amounts: []int64{35000}},
		{name: "two expenses yesterday", text: "Hôm qua ăn sáng 45.000đ và uống cà phê 30.000đ bằng Ví tiền mặt.", wantType: "expense", wantWallet: "wallet-cash", datePrefix: "2026-09-21", amounts: []int64{45000, 30000}},
		{name: "income explicit date", text: "Ngày 20/09/2026 nhận lương 15.000.000đ vào Tài khoản ngân hàng.", wantType: "income", wantWallet: "wallet-bank", wantCategory: "category-salary", datePrefix: "2026-09-20", amounts: []int64{15000000}},
		{name: "internal transfer safety", text: "Chuyển 500.000đ từ Tài khoản ngân hàng sang Ví tiền mặt.", wantType: "transfer", amounts: []int64{500000}, needsQuestion: true},
		{name: "missing wallet clarification", text: "Hôm nay tôi ăn trưa hết 80.000đ nhưng quên mất dùng ví nào.", wantType: "expense", amounts: []int64{80000}, needsQuestion: true, walletMustBeEmpty: true},
	}
	now := time.Date(2026, 9, 22, 10, 0, 0, 0, time.FixedZone("ICT", 7*60*60))
	for _, tc := range cases {
		input := Input{Text: tc.text, Timezone: "Asia/Ho_Chi_Minh", Now: now, Wallets: caseWallets, Categories: caseCategories}
		messages, err := buildMessages(input, input.Text)
		if err != nil {
			t.Fatalf("case %q build production extraction prompt: %v", tc.name, err)
		}
		request := map[string]any{
			"model": cfg.AIModel, "temperature": 0, "max_tokens": 600, "messages": messages,
			"response_format": map[string]any{"type": "json_schema", "json_schema": map[string]any{"name": "transaction_proposal", "strict": true, "schema": resultSchema}},
		}
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		started := time.Now()
		body, _, err := client.request(ctx, "POST", cfg.AIBaseURL+"/chat/completions", cfg.AIAPIKey, request, "AI JSON Schema probe")
		cancel()
		latency := time.Since(started)
		if err != nil {
			t.Errorf("case %q request failed latency=%s: %v", tc.name, latency.Round(time.Millisecond), err)
			continue
		}
		var envelope struct {
			Choices []struct {
				FinishReason string `json:"finish_reason"`
				Message      struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
			Usage struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
			} `json:"usage"`
		}
		if err := json.Unmarshal(body, &envelope); err != nil || len(envelope.Choices) != 1 {
			t.Errorf("case %q invalid response envelope bytes=%d latency=%s", tc.name, len(body), latency.Round(time.Millisecond))
			continue
		}
		choice := envelope.Choices[0]
		out, schemaErr := parseOutput([]byte(choice.Message.Content))
		if schemaErr != nil {
			var issues any
			if mismatch := new(SchemaMismatchError); errors.As(schemaErr, &mismatch) {
				issues, _ = json.Marshal(mismatch.Issues)
			}
			t.Errorf("case %q invalid output schema issues=%s latency=%s", tc.name, issues, latency.Round(time.Millisecond))
			continue
		}
		amounts := make(map[int64]int)
		typeOK, walletOK, categoryOK, dateOK, questionOK := true, true, true, true, true
		for _, draft := range out.Drafts {
			amounts[draft.Amount]++
			typeOK = typeOK && draft.Type == tc.wantType
			if tc.wantWallet != "" {
				walletOK = walletOK && draft.WalletID == tc.wantWallet
			}
			if tc.walletMustBeEmpty {
				walletOK = walletOK && draft.WalletID == ""
			}
			if tc.wantCategory != "" {
				categoryOK = categoryOK && draft.CategoryID != nil && *draft.CategoryID == tc.wantCategory
			}
			if tc.datePrefix != "" {
				dateOK = dateOK && strings.HasPrefix(draft.OccurredAt, tc.datePrefix)
			}
		}
		amountOK := len(out.Drafts) == len(tc.amounts)
		for _, amount := range tc.amounts {
			if amounts[amount] == 0 {
				amountOK = false
			}
		}
		if tc.needsQuestion {
			questionOK = false
			for _, draft := range out.Drafts {
				questionOK = questionOK || len(draft.Questions) > 0
			}
		}
		pass := len(out.Drafts) == len(tc.amounts) && amountOK && typeOK && walletOK && categoryOK && dateOK && questionOK
		t.Logf("json_schema_case=%q pass=%t drafts=%d amount_ok=%t type_ok=%t wallet_ok=%t category_ok=%t date_ok=%t question_ok=%t prompt_tokens=%d completion_tokens=%d latency=%s", tc.name, pass, len(out.Drafts), amountOK, typeOK, walletOK, categoryOK, dateOK, questionOK, envelope.Usage.PromptTokens, envelope.Usage.CompletionTokens, latency.Round(time.Millisecond))
		if !pass {
			t.Errorf("case %q failed semantic checks", tc.name)
		}
	}
}
