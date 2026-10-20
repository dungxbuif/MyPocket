package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/mypocket/backend/internal/entity"
)

type advisorProviderStub struct {
	responses []AdvisorResponse
	calls     int
}

type capturingAdvisorProvider struct {
	response AdvisorResponse
	request  AdvisorRequest
}

func (p *capturingAdvisorProvider) Chat(_ context.Context, request AdvisorRequest) (AdvisorResponse, error) {
	p.request = request
	return p.response, nil
}

type advisorPrincipalValidatorStub struct {
	calls int
	err   error
}

func (v *advisorPrincipalValidatorStub) Validate(context.Context, Principal) error {
	v.calls++
	return v.err
}

func (p *advisorProviderStub) Chat(_ context.Context, _ AdvisorRequest) (AdvisorResponse, error) {
	if p.calls >= len(p.responses) {
		return AdvisorResponse{}, errors.New("script exhausted")
	}
	response := p.responses[p.calls]
	p.calls++
	return response, nil
}

func TestAdvisorOrchestratorResolvesToolCallBeforeAnswer(t *testing.T) {
	result := entity.FinanceResult{QueryKey: "q1", Status: "ok", ViewKind: "finance_summary", View: json.RawMessage(`{"expense":150}`)}
	reader := &financeReaderStub{bundle: entity.FactBundle{Results: []entity.FinanceResult{result}}}
	call := AdvisorToolCall{ID: "call-1", Name: "get_finance_summary", Arguments: json.RawMessage(`{"range":{"from":"2026-09-01","to":"2026-09-30"}}`)}
	provider := &advisorProviderStub{responses: []AdvisorResponse{{ToolCalls: []AdvisorToolCall{call}}, {Text: "Bạn đã chi 150 ₫."}}}
	orchestrator := NewAdvisorOrchestrator(provider, NewAdvisorToolRegistry(NewFinanceQueryService(reader)))
	answer, err := orchestrator.Answer(context.Background(), Principal{OwnerID: "owner-1", Timezone: "Asia/Ho_Chi_Minh"}, []AdvisorChatMessage{{Role: "user", Content: "Tôi chi bao nhiêu?"}})
	if err != nil {
		t.Fatal(err)
	}
	if answer.Text != "Bạn đã chi 150 ₫." || provider.calls != 2 || reader.owner != "owner-1" {
		t.Fatalf("unexpected grounded answer: %+v calls=%d owner=%q", answer, provider.calls, reader.owner)
	}
}

func TestAdvisorOrchestratorIndexesEvidenceByProviderCall(t *testing.T) {
	result := entity.FinanceResult{QueryKey: "q1", Status: "ok", ViewKind: "finance_summary", Facts: []entity.Fact{{ID: "q1.expense", Kind: "money", Integer: int64Ptr(150)}}}
	reader := &financeReaderStub{bundle: entity.FactBundle{Results: []entity.FinanceResult{result}}}
	call := AdvisorToolCall{ID: "evidence-call", Name: "get_finance_summary", Arguments: json.RawMessage(`{"range":{"from":"2026-09-01","to":"2026-09-30"}}`)}
	provider := &advisorProviderStub{responses: []AdvisorResponse{{ToolCalls: []AdvisorToolCall{call}}, {Text: "ok"}}}
	answer, err := NewAdvisorOrchestrator(provider, NewAdvisorToolRegistry(NewFinanceQueryService(reader))).Answer(context.Background(), Principal{OwnerID: "owner-1", Timezone: "Asia/Ho_Chi_Minh"}, []AdvisorChatMessage{{Role: "user", Content: "summary"}})
	if err != nil {
		t.Fatal(err)
	}
	if answer.Evidence == nil {
		t.Fatal("advisor answer must retain run-local evidence index")
	}
	if _, err := answer.Evidence.Resolve("evidence-call.q1.expense"); err != nil {
		t.Fatalf("provider call evidence missing: %v", err)
	}
}

func TestAdvisorOrchestratorGroundingRequiresTypedFinishAnswer(t *testing.T) {
	amount := int64(150)
	result := entity.FinanceResult{QueryKey: "q1", Status: "ok", ViewKind: "finance_summary", Facts: []entity.Fact{{ID: "q1.expense", Kind: "money", Integer: &amount, Currency: stringPtr(entity.WalletCurrencyVND)}}}
	reader := &financeReaderStub{bundle: entity.FactBundle{Results: []entity.FinanceResult{result}}}
	lookup := AdvisorToolCall{ID: "lookup", Name: "get_finance_summary", Arguments: json.RawMessage(`{"range":{"from":"2026-09-01","to":"2026-09-30"}}`)}
	finish := AdvisorToolCall{ID: "finish", Name: "finish_answer", Arguments: json.RawMessage(`{"mode":"finance","template_id":"summary","claims":[{"slot":"Chi tiêu","fact_ref":"lookup.q1.expense"}]}`)}
	provider := &advisorProviderStub{responses: []AdvisorResponse{{ToolCalls: []AdvisorToolCall{lookup}}, {ToolCalls: []AdvisorToolCall{finish}}}}
	orchestrator := NewAdvisorOrchestrator(provider, NewAdvisorToolRegistry(NewFinanceQueryService(reader)))
	orchestrator.GroundingEnabled = true
	answer, err := orchestrator.Answer(context.Background(), Principal{OwnerID: "owner-1", Timezone: "Asia/Ho_Chi_Minh"}, []AdvisorChatMessage{{Role: "user", Content: "Tôi chi bao nhiêu?"}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(answer.Text, "Chi tiêu: 150") || !strings.Contains(answer.Text, entity.WalletCurrencyVND) {
		t.Fatalf("typed finish answer must render server evidence: %q", answer.Text)
	}
}

func TestAdvisorOrchestratorGroundingRejectsFreeTextFinal(t *testing.T) {
	provider := &advisorProviderStub{responses: []AdvisorResponse{{Text: "Bạn đã chi 150 ₫."}, {Text: "Bạn đã chi 150 ₫."}}}
	orchestrator := NewAdvisorOrchestrator(provider, NewAdvisorToolRegistry(NewFinanceQueryService(&financeReaderStub{})))
	orchestrator.GroundingEnabled = true
	if _, err := orchestrator.Answer(context.Background(), Principal{OwnerID: "owner-1"}, []AdvisorChatMessage{{Role: "user", Content: "summary"}}); !errors.Is(err, ErrAdvisorGroundingRequired) {
		t.Fatalf("grounding must reject untyped free text, got %v", err)
	}
}

func TestAdvisorOrchestratorGroundingRepairsOnceBeforeFallback(t *testing.T) {
	amount := int64(150)
	result := entity.FinanceResult{QueryKey: "q1", Status: "ok", ViewKind: "finance_summary", Facts: []entity.Fact{{ID: "q1.expense", Kind: "money", Integer: &amount, Currency: stringPtr(entity.WalletCurrencyVND)}}}
	reader := &financeReaderStub{bundle: entity.FactBundle{Results: []entity.FinanceResult{result}}}
	lookup := AdvisorToolCall{ID: "lookup", Name: "get_finance_summary", Arguments: json.RawMessage(`{"range":{"from":"2026-09-01","to":"2026-09-30"}}`)}
	finish := AdvisorToolCall{ID: "finish", Name: "finish_answer", Arguments: json.RawMessage(`{"mode":"finance","template_id":"summary","claims":[{"slot":"Chi tiêu","fact_ref":"lookup.q1.expense"}]}`)}
	provider := &advisorProviderStub{responses: []AdvisorResponse{{ToolCalls: []AdvisorToolCall{lookup}}, {Text: "Bạn đã chi ba triệu đồng."}, {ToolCalls: []AdvisorToolCall{finish}}}}
	orchestrator := NewAdvisorOrchestrator(provider, NewAdvisorToolRegistry(NewFinanceQueryService(reader)))
	orchestrator.GroundingEnabled = true
	answer, err := orchestrator.Answer(context.Background(), Principal{OwnerID: "owner-1", Timezone: "Asia/Ho_Chi_Minh"}, []AdvisorChatMessage{{Role: "user", Content: "summary"}})
	if err != nil || !strings.Contains(answer.Text, "150") || provider.calls != 3 {
		t.Fatalf("grounding repair did not recover typed final: answer=%+v err=%v calls=%d", answer, err, provider.calls)
	}
}

func TestAdvisorOrchestratorGroundingFallsBackToValidatedCards(t *testing.T) {
	amount := int64(150)
	result := entity.FinanceResult{QueryKey: "q1", Status: "ok", ViewKind: "finance_summary", Facts: []entity.Fact{{ID: "q1.expense", Kind: "money", Integer: &amount}}}
	reader := &financeReaderStub{bundle: entity.FactBundle{Results: []entity.FinanceResult{result}}}
	lookup := AdvisorToolCall{ID: "lookup", Name: "get_finance_summary", Arguments: json.RawMessage(`{"range":{"from":"2026-09-01","to":"2026-09-30"}}`)}
	badFinish := AdvisorToolCall{ID: "finish", Name: "finish_answer", Arguments: json.RawMessage(`{"mode":"finance","template_id":"summary","claims":[{"slot":"x","fact_ref":"forged"}]}`)}
	provider := &advisorProviderStub{responses: []AdvisorResponse{{ToolCalls: []AdvisorToolCall{lookup}}, {ToolCalls: []AdvisorToolCall{badFinish}}, {ToolCalls: []AdvisorToolCall{badFinish}}}}
	orchestrator := NewAdvisorOrchestrator(provider, NewAdvisorToolRegistry(NewFinanceQueryService(reader)))
	orchestrator.GroundingEnabled = true
	answer, err := orchestrator.Answer(context.Background(), Principal{OwnerID: "owner-1", Timezone: "Asia/Ho_Chi_Minh"}, []AdvisorChatMessage{{Role: "user", Content: "summary"}})
	if err != nil || !strings.Contains(answer.Text, "thẻ kết quả") || len(answer.Results) != 1 {
		t.Fatalf("invalid final should fall back to validated cards: answer=%+v err=%v", answer, err)
	}
}

func TestAdvisorOrchestratorAggregatesBoundedPerCallUsage(t *testing.T) {
	call := AdvisorToolCall{ID: "call-1", Name: "get_finance_summary", Arguments: json.RawMessage(`{"range":{"from":"2026-09-01","to":"2026-09-30"}}`)}
	provider := &advisorProviderStub{responses: []AdvisorResponse{
		{ToolCalls: []AdvisorToolCall{call}, Usage: map[string]any{
			"provider": "openai_compatible", "model": "advisor-model", "request_id": "request-1", "latency_ms": int64(12), "finish_reason": "tool_calls",
			"prompt_tokens": int64(10), "completion_tokens": int64(3), "total_tokens": int64(13), "secret": "must-not-survive",
		}},
		{Text: "Bạn đã chi 150 ₫.", Usage: map[string]any{
			"provider": "openai_compatible", "model": "advisor-model", "request_id": "request-2", "latency_ms": int64(8), "finish_reason": "stop",
			"prompt_tokens": int64(14), "completion_tokens": int64(5), "total_tokens": int64(19), "nested": map[string]any{"prompt": "must-not-survive"},
		}},
	}}
	reader := &financeReaderStub{bundle: entity.FactBundle{Results: []entity.FinanceResult{{QueryKey: "q1", Status: "ok", ViewKind: "finance_summary", View: json.RawMessage(`{"expense":150}`)}}}}
	orchestrator := NewAdvisorOrchestrator(provider, NewAdvisorToolRegistry(NewFinanceQueryService(reader)))
	answer, err := orchestrator.Answer(context.Background(), Principal{OwnerID: "owner-1", Timezone: "Asia/Ho_Chi_Minh"}, []AdvisorChatMessage{{Role: "user", Content: "Tôi chi bao nhiêu?"}})
	if err != nil {
		t.Fatal(err)
	}
	if answer.Usage == nil {
		t.Fatal("aggregated usage must be retained")
	}
	for key, want := range map[string]int64{"provider_calls": 2, "prompt_tokens": 24, "completion_tokens": 8, "total_tokens": 32} {
		got, ok := advisorUsageNumber(answer.Usage[key])
		if !ok || got != want {
			t.Fatalf("aggregate %s=%d, want %d: %#v", key, got, want, answer.Usage)
		}
	}
	calls, ok := answer.Usage["calls"].([]map[string]any)
	if !ok || len(calls) != 2 {
		t.Fatalf("per-call usage missing or unbounded: %#v", answer.Usage["calls"])
	}
	if calls[0]["request_id"] != "request-1" || calls[0]["finish_reason"] != "tool_calls" || calls[1]["request_id"] != "request-2" || calls[1]["finish_reason"] != "stop" {
		t.Fatalf("per-call metadata lost: %#v", calls)
	}
	for _, callUsage := range calls {
		if _, ok := callUsage["secret"]; ok {
			t.Fatal("unknown per-call usage field must not survive aggregation")
		}
		if _, ok := callUsage["nested"]; ok {
			t.Fatal("nested per-call usage field must not survive aggregation")
		}
	}
}

func TestAdvisorOrchestratorRejectsToolCallsWithoutCorrelationID(t *testing.T) {
	result := entity.FinanceResult{QueryKey: "q1", Status: "ok", ViewKind: "finance_summary", View: json.RawMessage(`{"expense":150}`)}
	reader := &financeReaderStub{bundle: entity.FactBundle{Results: []entity.FinanceResult{result}}}
	provider := &advisorProviderStub{responses: []AdvisorResponse{
		{ToolCalls: []AdvisorToolCall{{Name: "get_finance_summary", Arguments: json.RawMessage(`{"range":{"from":"2026-09-01","to":"2026-09-30"}}`)}}},
		{Text: "unexpected"},
	}}
	orchestrator := NewAdvisorOrchestrator(provider, NewAdvisorToolRegistry(NewFinanceQueryService(reader)))
	if _, err := orchestrator.Answer(context.Background(), Principal{OwnerID: "owner-1", Timezone: "Asia/Ho_Chi_Minh"}, []AdvisorChatMessage{{Role: "user", Content: "summary"}}); !errors.Is(err, ErrAdvisorToolInvalid) {
		t.Fatalf("expected malformed tool call to be rejected, got %v", err)
	}
	if reader.owner != "" {
		t.Fatalf("malformed provider response must not execute a tool, reader owner=%q", reader.owner)
	}
}

func TestAdvisorOrchestratorRejectsDuplicateToolCorrelationIDs(t *testing.T) {
	result := entity.FinanceResult{QueryKey: "q1", Status: "ok", ViewKind: "finance_summary", View: json.RawMessage(`{"expense":150}`)}
	reader := &financeReaderStub{bundle: entity.FactBundle{Results: []entity.FinanceResult{result}}}
	calls := []AdvisorToolCall{
		{ID: "duplicate", Name: "get_finance_summary", Arguments: json.RawMessage(`{"range":{"from":"2026-09-01","to":"2026-09-30"}}`)},
		{ID: "duplicate", Name: "get_finance_summary", Arguments: json.RawMessage(`{"range":{"from":"2026-09-01","to":"2026-09-30"}}`)},
	}
	provider := &advisorProviderStub{responses: []AdvisorResponse{{ToolCalls: calls}, {Text: "unexpected"}}}
	orchestrator := NewAdvisorOrchestrator(provider, NewAdvisorToolRegistry(NewFinanceQueryService(reader)))
	if _, err := orchestrator.Answer(context.Background(), Principal{OwnerID: "owner-1", Timezone: "Asia/Ho_Chi_Minh"}, []AdvisorChatMessage{{Role: "user", Content: "summary"}}); !errors.Is(err, ErrAdvisorToolInvalid) {
		t.Fatalf("expected duplicate tool call IDs to be rejected, got %v", err)
	}
	if reader.owner != "" {
		t.Fatalf("malformed provider response must not execute a tool, reader owner=%q", reader.owner)
	}
}

func TestAdvisorOrchestratorPrependsProductHandbook(t *testing.T) {
	provider := &capturingAdvisorProvider{response: AdvisorResponse{Text: "ok"}}
	orchestrator := NewAdvisorOrchestrator(provider, NewAdvisorToolRegistry(NewFinanceQueryService(&financeReaderStub{})))
	if _, err := orchestrator.Answer(context.Background(), Principal{OwnerID: "owner-1", Timezone: "Asia/Ho_Chi_Minh"}, []AdvisorChatMessage{{Role: "user", Content: "Tính năng ví ở đâu?"}}); err != nil {
		t.Fatal(err)
	}
	if len(provider.request.Messages) != 2 || provider.request.Messages[0].Role != "system" {
		t.Fatalf("advisor must prepend a system handbook: %+v", provider.request.Messages)
	}
	content := provider.request.Messages[0].Content
	for _, required := range []string{"read-only", "Giao dịch", "tool", "edit or delete", "propose_create_jar", "propose_update_jar", "propose_create_budget", "propose_update_budget", "xác nhận", "không có quyền", "không từ chối", "ĐƯỢC PHÉP chuẩn bị đề xuất tạo hũ"} {
		if !strings.Contains(content, required) {
			t.Fatalf("handbook missing %q: %s", required, content)
		}
	}
	jarToolFound := false
	for _, definition := range provider.request.Tools {
		if definition.Name == "propose_create_jar" {
			jarToolFound = true
			if !strings.Contains(definition.Description, "never create it directly") {
				t.Fatalf("jar proposal tool must advertise its non-mutating boundary: %q", definition.Description)
			}
		}
	}
	if !jarToolFound {
		t.Fatalf("advisor request must expose propose_create_jar to the provider: %+v", provider.request.Tools)
	}
	if strings.Contains(content, "cannot directly create/edit/delete categories, budgets, jars") {
		t.Fatalf("handbook still contradicts the implemented jar proposal tool: %s", content)
	}
}

func TestAdvisorOrchestratorHandbookMatchesPrimaryAssistantTab(t *testing.T) {
	provider := &capturingAdvisorProvider{response: AdvisorResponse{Text: "ok"}}
	orchestrator := NewAdvisorOrchestrator(provider, NewAdvisorToolRegistry(NewFinanceQueryService(&financeReaderStub{})))
	if _, err := orchestrator.Answer(context.Background(), Principal{OwnerID: "owner-1", Timezone: "Asia/Ho_Chi_Minh"}, []AdvisorChatMessage{{Role: "user", Content: "Trợ lý ở đâu?"}}); err != nil {
		t.Fatal(err)
	}
	content := provider.request.Messages[0].Content
	if !strings.Contains(content, "Trợ lý là tab chính") {
		t.Fatalf("handbook must describe the primary Assistant tab: %s", content)
	}
	if strings.Contains(content, "More includes") || strings.Contains(content, "More chứa") {
		t.Fatalf("handbook must not place Assistant under More: %s", content)
	}
}

func TestAdvisorOrchestratorCapsToolCalls(t *testing.T) {
	call := AdvisorToolCall{ID: "call-1", Name: "get_finance_summary", Arguments: json.RawMessage(`{"range":{"from":"2026-09-01","to":"2026-09-30"}}`)}
	provider := &advisorProviderStub{responses: []AdvisorResponse{{ToolCalls: []AdvisorToolCall{call}}}}
	orchestrator := NewAdvisorOrchestrator(provider, NewAdvisorToolRegistry(NewFinanceQueryService(&financeReaderStub{})))
	if _, err := orchestrator.Answer(context.Background(), Principal{OwnerID: "owner-1", Timezone: "Asia/Ho_Chi_Minh"}, nil); err == nil {
		t.Fatal("tool loop must stop when provider repeats beyond budget")
	}
}

func TestAdvisorOrchestratorAllowsFinalAnswerAfterFourToolCalls(t *testing.T) {
	call := AdvisorToolCall{ID: "call-1", Name: "get_wallet_balances", Arguments: json.RawMessage(`{"wallet_ids":[]}`)}
	responses := []AdvisorResponse{
		{ToolCalls: []AdvisorToolCall{call}},
		{ToolCalls: []AdvisorToolCall{call}},
		{ToolCalls: []AdvisorToolCall{call}},
		{ToolCalls: []AdvisorToolCall{call}},
		{Text: "Đây là câu trả lời cuối."},
	}
	for index := 0; index < 4; index++ {
		responses[index].ToolCalls[0].ID = fmt.Sprintf("call-%d", index+1)
	}
	provider := &advisorProviderStub{responses: responses}
	reader := &financeReaderStub{bundle: entity.FactBundle{Results: []entity.FinanceResult{{QueryKey: "q1", ViewKind: "wallet_balances", View: json.RawMessage(`{"items":[]}`)}}}}
	orchestrator := NewAdvisorOrchestrator(provider, NewAdvisorToolRegistry(NewFinanceQueryService(reader)))
	answer, err := orchestrator.Answer(context.Background(), Principal{OwnerID: "owner-1", Timezone: "Asia/Ho_Chi_Minh"}, []AdvisorChatMessage{{Role: "user", Content: "Tóm tắt ví"}})
	if err != nil {
		t.Fatal(err)
	}
	if answer.Text != "Đây là câu trả lời cuối." || provider.calls != 5 {
		t.Fatalf("expected final answer after four tool calls, answer=%+v calls=%d", answer, provider.calls)
	}
}

func TestAdvisorOrchestratorAllowsFiveToolCallsInOneProviderResponse(t *testing.T) {
	calls := make([]AdvisorToolCall, 0, 5)
	for index := 1; index <= 5; index++ {
		calls = append(calls, AdvisorToolCall{ID: fmt.Sprintf("call-%d", index), Name: "get_wallet_balances", Arguments: json.RawMessage(`{"wallet_ids":[]}`)})
	}
	provider := &advisorProviderStub{responses: []AdvisorResponse{
		{ToolCalls: calls},
		{Text: "Đây là câu trả lời cuối."},
	}}
	reader := &financeReaderStub{bundle: entity.FactBundle{Results: []entity.FinanceResult{{QueryKey: "q1", ViewKind: "wallet_balances", View: json.RawMessage(`{"items":[]}`)}}}}
	orchestrator := NewAdvisorOrchestrator(provider, NewAdvisorToolRegistry(NewFinanceQueryService(reader)))
	answer, err := orchestrator.Answer(context.Background(), Principal{OwnerID: "owner-1", Timezone: "Asia/Ho_Chi_Minh"}, []AdvisorChatMessage{{Role: "user", Content: "Tóm tắt ví"}})
	if err != nil {
		t.Fatalf("five tool calls returned an error: %v", err)
	}
	if answer.Text != "Đây là câu trả lời cuối." || provider.calls != 2 {
		t.Fatalf("expected final answer after five tool calls, answer=%+v calls=%d", answer, provider.calls)
	}
}

func TestAdvisorOrchestratorRevalidatesPrincipalBeforeProviderAndTool(t *testing.T) {
	result := entity.FinanceResult{QueryKey: "q1", Status: "ok", ViewKind: "finance_summary", View: json.RawMessage(`{"expense":150}`)}
	reader := &financeReaderStub{bundle: entity.FactBundle{Results: []entity.FinanceResult{result}}}
	call := AdvisorToolCall{ID: "call-1", Name: "get_finance_summary", Arguments: json.RawMessage(`{"range":{"from":"2026-09-01","to":"2026-09-30"}}`)}
	provider := &advisorProviderStub{responses: []AdvisorResponse{{ToolCalls: []AdvisorToolCall{call}}, {Text: "ok"}}}
	validator := &advisorPrincipalValidatorStub{}
	orchestrator := NewAdvisorOrchestrator(provider, NewAdvisorToolRegistry(NewFinanceQueryService(reader)))
	orchestrator.Validator = validator
	if _, err := orchestrator.Answer(context.Background(), Principal{OwnerID: "owner-1", Timezone: "Asia/Ho_Chi_Minh"}, []AdvisorChatMessage{{Role: "user", Content: "summary"}}); err != nil {
		t.Fatal(err)
	}
	if validator.calls < 3 {
		t.Fatalf("principal should be validated before each provider/tool boundary, calls=%d", validator.calls)
	}
}
