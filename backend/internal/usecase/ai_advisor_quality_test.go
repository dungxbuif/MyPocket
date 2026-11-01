package usecase

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/mypocket/backend/internal/entity"
)

// Script the provider boundary, but execute the real proposal validator,
// evidence renderer and persisted-card builder.
func TestAdvisorQualityProposalReplyDoesNotRepeatCard(t *testing.T) {
	for _, mode := range []string{"legacy", "grounded", "grounded_fallback"} {
		t.Run(mode, func(t *testing.T) {
			proposal := AdvisorResponse{ToolCalls: []AdvisorToolCall{{ID: "jar", Name: "propose_create_jar", Arguments: json.RawMessage(`{"month":"2026-11","name":"Du lịch","allocation_mode":"fixed","allocation_amount":2000000}`)}}}
			verbose := AdvisorResponse{Text: "Đã tạo hũ Du lịch. Mỗi tháng tự động chuyển 2000000 VND. allocation_mode=fixed. Bạn cần xác nhận."}
			responses := []AdvisorResponse{proposal, verbose}
			if mode == "grounded" {
				responses[1] = AdvisorResponse{ToolCalls: []AdvisorToolCall{{ID: "finish", Name: "finish_answer", Arguments: json.RawMessage(`{"mode":"finance","template_id":"proposal","claims":[]}`)}}}
			} else if mode == "grounded_fallback" {
				responses = append(responses, verbose)
			}
			provider := &advisorProviderStub{responses: responses}
			orchestrator := NewAdvisorOrchestrator(provider, NewAdvisorToolRegistry(NewFinanceQueryService(&financeReaderStub{})))
			orchestrator.GroundingEnabled = mode != "legacy"
			answer, err := orchestrator.Answer(context.Background(), Principal{OwnerID: "owner-1", Timezone: "Asia/Ho_Chi_Minh"}, []AdvisorChatMessage{{Role: "user", Content: "Tạo hũ Du lịch tháng 11/2026, dành 2 triệu đồng."}})
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(answer.Text, "xác nhận") || len([]rune(answer.Text)) > 120 || strings.Contains(answer.Text, "2000000") || strings.Contains(answer.Text, "Đã tạo") {
				t.Fatalf("expected concise proposal-only acknowledgement, got %q", answer.Text)
			}
			parts := BuildAdvisorPartsForRun(answer, "conversation-1", 7, "run-1")
			if len(parts) != 2 || parts[1].Type != "jar_confirmation" {
				t.Fatalf("missing proposal card: %+v", parts)
			}
			view := parts[1].Data.(map[string]any)["view"].(map[string]any)
			if view["allocation_amount"] != float64(2000000) || view["month"] != "2026-11" || view["action_id"] == "" || view["action"] != "create_jar" || view["status"] != entity.AIProposalPending || view["generation"] != int64(7) || view["conversation_id"] != "conversation-1" || view["run_id"] != "run-1" {
				t.Fatalf("proposal or confirmation provenance changed: %+v", view)
			}
		})
	}
}

func TestAdvisorQualityGroundedJarMonthClarification(t *testing.T) {
	provider := &advisorProviderStub{responses: []AdvisorResponse{{ToolCalls: []AdvisorToolCall{{ID: "finish", Name: "finish_answer", Arguments: json.RawMessage(`{"mode":"clarification","question_id":"jar_month"}`)}}}}}
	orchestrator := NewAdvisorOrchestrator(provider, NewAdvisorToolRegistry(NewFinanceQueryService(&financeReaderStub{})))
	orchestrator.GroundingEnabled = true
	answer, err := orchestrator.Answer(context.Background(), Principal{OwnerID: "owner-1"}, []AdvisorChatMessage{{Role: "user", Content: "Tạo hũ Du lịch, mỗi tháng dành 2 triệu đồng."}})
	if err != nil || !strings.Contains(answer.Text, "tháng nào") || len(answer.Results) != 0 {
		t.Fatalf("missing month must support a focused clarification without a proposal: %+v, %v", answer, err)
	}
}

func TestAdvisorQualityMissingJarMonthIsNotDefaulted(t *testing.T) {
	_, err := (&jarProposalTool{}).Execute(context.Background(), Principal{OwnerID: "owner-1"}, json.RawMessage(`{"name":"Du lịch","allocation_mode":"fixed","allocation_amount":2000000}`))
	if err == nil {
		t.Fatal("missing month must not silently become a current-month or recurring allocation")
	}
}

func TestAdvisorQualityPendingAmountEditDoesNotClaimMutation(t *testing.T) {
	provider := &advisorProviderStub{responses: []AdvisorResponse{{ToolCalls: []AdvisorToolCall{{ID: "finish", Name: "finish_answer", Arguments: json.RawMessage(`{"mode":"clarification","question_id":"proposal_edit"}`)}}}}}
	orchestrator := NewAdvisorOrchestrator(provider, NewAdvisorToolRegistry(NewFinanceQueryService(&financeReaderStub{})))
	orchestrator.GroundingEnabled = true
	answer, err := orchestrator.Answer(context.Background(), Principal{OwnerID: "owner-1"}, []AdvisorChatMessage{
		{Role: "user", Content: "Tạo hũ Du lịch tháng 11/2026, dành 2 triệu."},
		{Role: "assistant", Content: "Đã chuẩn bị đề xuất. Bạn kiểm tra và xác nhận trên thẻ để lưu."},
		{Role: "user", Content: "đổi thành 3 triệu"},
	})
	if err != nil || len(answer.Results) != 0 || !strings.Contains(answer.Text, "chưa thể sửa") || !strings.Contains(answer.Text, "đừng xác nhận thẻ cũ") {
		t.Fatalf("pending edit must disclose limitation without creating another jar: %+v, %v", answer, err)
	}
}

func TestAdvisorQualityFallbackRetainsMixedResultsAndWarnings(t *testing.T) {
	results := []entity.FinanceResult{
		{Status: "ok", ViewKind: "jar_progress", View: json.RawMessage(`{"month":"2026-11","items":[]}`)},
		{Status: "proposed", ViewKind: "jar_creation_proposal", View: json.RawMessage(`{"action_id":"a1","action":"create_jar"}`), WarningCodes: []string{"review_required"}},
	}
	answer := groundedFallbackAnswer(results, nil, false, NewAdvisorEvidenceIndex())
	if !strings.Contains(answer.Text, "thẻ kết quả") || len(answer.Results) != 2 {
		t.Fatalf("mixed result fallback must retain read results: %+v", answer)
	}
	parts := BuildAdvisorParts(answer)
	if len(parts) != 3 || parts[2].Data.(map[string]any)["warnings"].([]string)[0] != "review_required" {
		t.Fatalf("card warnings lost: %+v", parts)
	}
}
