package usecase

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mypocket/backend/internal/entity"
)

type advisorProviderStub struct {
	responses []AdvisorResponse
	calls     int
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
