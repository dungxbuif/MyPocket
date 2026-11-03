package usecase

import (
	"encoding/json"
	"testing"

	"github.com/mypocket/backend/internal/entity"
)

func TestBuildAdvisorPartsKeepsNumbersOutOfFreeTextContract(t *testing.T) {
	answer := AdvisorAnswer{
		Text: "Bạn đã chi 150 ₫.",
		Results: []entity.FinanceResult{{
			Status:   "ok",
			ViewKind: "finance_summary",
			View:     json.RawMessage(`{"income":1000,"expense":150,"net":850,"count":3}`),
		}},
	}
	parts := BuildAdvisorParts(answer)
	if len(parts) != 2 || parts[0].Type != "text" || parts[1].Type != "metric_group" {
		t.Fatalf("unexpected parts: %+v", parts)
	}
	if parts[1].Data == nil {
		t.Fatal("metric part must carry server view data")
	}
}

func TestBuildAdvisorPartsRejectsUnknownAndMalformedViews(t *testing.T) {
	parts := BuildAdvisorParts(AdvisorAnswer{Text: "ok", Results: []entity.FinanceResult{{ViewKind: "future_card", View: json.RawMessage(`{"amount":999}`)}, {ViewKind: "finance_summary", View: json.RawMessage(`not-json`)}}})
	if len(parts) != 1 || parts[0].Type != "text" {
		t.Fatalf("unknown or malformed views must not reach the renderer: %+v", parts)
	}
}

func TestBuildAdvisorPartsMapsProductHelpToStructuredCard(t *testing.T) {
	parts := BuildAdvisorParts(AdvisorAnswer{Results: []entity.FinanceResult{{Status: "ok", ViewKind: "product_help", View: json.RawMessage(`{"version":"1.1","sources":[{"id":"handbook.navigation.v1.1","title":"Điều hướng","excerpt":"Mở Tài khoản."}]}`)}}})
	if len(parts) != 1 || parts[0].Type != "product_help" {
		t.Fatalf("product help must remain a structured part: %+v", parts)
	}
}

func TestBuildAdvisorPartsCarriesServerActionContextForWalletProposal(t *testing.T) {
	parts := BuildAdvisorPartsForRun(AdvisorAnswer{Results: []entity.FinanceResult{{
		Status:   "proposed",
		ViewKind: "wallet_creation_proposal",
		View:     json.RawMessage(`{"action_id":"action-1","action":"create_wallet","name":"Du lịch","type":"basic","opening_balance":0,"is_in_total":true}`),
	}}}, "conversation-1", 7, "run-1")
	if len(parts) != 1 || parts[0].Type != "wallet_confirmation" {
		t.Fatalf("unexpected wallet proposal parts: %+v", parts)
	}
	data, ok := parts[0].Data.(map[string]any)
	if !ok {
		t.Fatalf("wallet proposal data type = %T", parts[0].Data)
	}
	view, ok := data["view"].(map[string]any)
	generation, generationOK := view["generation"].(int64)
	if !ok || view["conversation_id"] != "conversation-1" || !generationOK || generation != 7 || view["run_id"] != "run-1" {
		t.Fatalf("proposal context missing: %+v", data)
	}
}

func TestBuildAdvisorPartsCarriesServerActionContextForJarProposal(t *testing.T) {
	parts := BuildAdvisorPartsForRun(AdvisorAnswer{Results: []entity.FinanceResult{{
		Status: "proposed", ViewKind: "jar_creation_proposal",
		View: json.RawMessage(`{"action_id":"jar-action-1","action":"create_jar","month":"2026-09","name":"Ăn uống","allocation_mode":"fixed","allocation_amount":400000}`),
	}}}, "conversation-1", 7, "run-1")
	if len(parts) != 1 || parts[0].Type != "jar_confirmation" {
		t.Fatalf("unexpected jar proposal parts: %+v", parts)
	}
	data, ok := parts[0].Data.(map[string]any)
	if !ok {
		t.Fatalf("jar proposal data type = %T", parts[0].Data)
	}
	view, ok := data["view"].(map[string]any)
	generation, generationOK := view["generation"].(int64)
	if !ok || view["conversation_id"] != "conversation-1" || !generationOK || generation != 7 || view["run_id"] != "run-1" {
		t.Fatalf("jar proposal context missing: %+v", data)
	}
}

func TestBuildAdvisorPartsMapsJarUpdateProposalToConfirmation(t *testing.T) {
	parts := BuildAdvisorPartsForRun(AdvisorAnswer{Results: []entity.FinanceResult{{
		Status: "proposed", ViewKind: "jar_update_proposal",
		View: json.RawMessage(`{"action_id":"jar-update-1","action":"update_jar","jar_id":"jar-1","month":"2026-09","name":"Ăn uống","allocation_mode":"fixed","allocation_amount":400000}`),
	}}}, "conversation-1", 7, "run-1")
	if len(parts) != 1 || parts[0].Type != "jar_confirmation" {
		t.Fatalf("unexpected jar update proposal parts: %+v", parts)
	}
	data, ok := parts[0].Data.(map[string]any)
	if !ok {
		t.Fatalf("jar update proposal data type = %T", parts[0].Data)
	}
	view, ok := data["view"].(map[string]any)
	if !ok || view["action"] != "update_jar" || view["jar_id"] != "jar-1" {
		t.Fatalf("jar update proposal payload missing action: %+v", data)
	}
}

func TestBuildAdvisorPartsMapsBudgetProposalToConfirmation(t *testing.T) {
	answer := AdvisorAnswer{Results: []entity.FinanceResult{{
		Status: "proposed", ViewKind: "budget_creation_proposal",
		View: json.RawMessage(`{"action_id":"budget-1","action":"create_budget","name":"Ăn uống","limit_amount":4000000,"start_date":"2026-09-01","end_date":"2026-09-30"}`),
	}}}
	parts := BuildAdvisorPartsForRun(answer, "conversation-1", 4, "run-1")
	if len(parts) != 1 || parts[0].Type != "budget_confirmation" {
		t.Fatalf("unexpected budget proposal parts: %+v", parts)
	}
	envelope, ok := parts[0].Data.(map[string]any)
	if !ok {
		t.Fatalf("budget proposal data is not an envelope: %+v", parts[0].Data)
	}
	view, ok := envelope["view"].(map[string]any)
	if !ok || view["action"] != "create_budget" || view["conversation_id"] != "conversation-1" {
		t.Fatalf("budget proposal lost binding: %+v", parts[0])
	}
}

func TestSanitizeAdvisorTextStripsLeakedProposalJSON(t *testing.T) {
	leaked := "{\"version\":3,\"amount\":120000,\"id\":\"abc\"}"
	if got := sanitizeAdvisorText(leaked, true); got != advisorProposalCardText {
		t.Fatalf("json-only with card = %q", got)
	}
	if got := sanitizeAdvisorText(leaked, false); got != "" {
		t.Fatalf("json-only without card = %q", got)
	}
	mixed := "Mình đề xuất 2 khoản.\n```json\n{\"version\":1,\"amount\":5,\"id\":\"x\"}\n```\n\"version\": 2,\n<tool_call>{\"id\":1}</tool_call>\nBạn xác nhận nhé."
	if got := sanitizeAdvisorText(mixed, true); got != "Mình đề xuất 2 khoản.\n\nBạn xác nhận nhé." {
		t.Fatalf("mixed = %q", got)
	}
	plain := "Tháng này bạn chi 3 {khoản}."
	if got := sanitizeAdvisorText(plain, true); got != plain {
		t.Fatalf("plain changed: %q", got)
	}
	parts := BuildAdvisorParts(AdvisorAnswer{Text: leaked, Results: []entity.FinanceResult{{ViewKind: "jar_assignment_proposal", View: []byte(`{"month":"2026-11"}`), Status: "ok"}}})
	if len(parts) < 1 || parts[0].Type != "text" || parts[0].Text != advisorProposalCardText {
		t.Fatalf("parts = %+v", parts)
	}
}
