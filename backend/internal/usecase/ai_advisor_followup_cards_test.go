package usecase

import (
	"testing"

	"github.com/mypocket/backend/internal/entity"
)

func transactionListPart(rows int) entity.AdvisorPart {
	items := make([]any, rows)
	for i := range items {
		items[i] = map[string]any{"id": i}
	}
	return entity.AdvisorPart{Type: "transaction_list", Data: map[string]any{"view": map[string]any{"items": items}}}
}

func TestNarrowReportFollowUpCardsKeepsOnlyTheMostSpecificList(t *testing.T) {
	withContext := &entity.AdvisorMessage{Parts: []entity.AdvisorPart{{Type: "text", Text: "q"}, {Type: "report_context"}}}
	parts := []entity.AdvisorPart{{Type: "text", Text: "answer"}, {Type: "budget_status"}, transactionListPart(18), transactionListPart(1)}
	got := narrowReportFollowUpCards(parts, withContext)
	if len(got) != 3 || got[2].Type != "transaction_list" || len(got[2].Data.(map[string]any)["view"].(map[string]any)["items"].([]any)) != 1 {
		t.Fatalf("narrowed parts = %+v, want text, budget and the one-row list", got)
	}
	plain := &entity.AdvisorMessage{Parts: []entity.AdvisorPart{{Type: "text", Text: "q"}}}
	if all := narrowReportFollowUpCards(parts, plain); len(all) != 4 {
		t.Fatalf("without report context every card is kept, got %d", len(all))
	}
	single := []entity.AdvisorPart{{Type: "text", Text: "a"}, transactionListPart(18)}
	if kept := narrowReportFollowUpCards(single, withContext); len(kept) != 2 {
		t.Fatalf("a single list must not be removed, got %d", len(kept))
	}
}

func TestNarrowReportFollowUpCardsFiltersRowsToTheFindingScope(t *testing.T) {
	context := entity.AdvisorPart{Type: "report_context", Data: map[string]any{"findings": []any{map[string]any{"drilldown": map[string]any{
		"category_ids": []any{"shopping"}, "type": "expense", "from": "2026-10-31T17:00:00Z", "end_exclusive": "2026-11-02T14:00:00Z"}}}}}
	message := &entity.AdvisorMessage{Parts: []entity.AdvisorPart{{Type: "text", Text: "q"}, context}}
	list := entity.AdvisorPart{Type: "transaction_list", Data: map[string]any{"view": map[string]any{"next_cursor": "x", "total_count": 3, "items": []any{
		map[string]any{"category_id": "shopping", "type": "expense", "occurred_at": "2026-11-01T12:15:00+07:00"},
		map[string]any{"category_id": "food", "type": "expense", "occurred_at": "2026-11-01T12:15:00+07:00"},
		map[string]any{"category_id": "shopping", "type": "expense", "occurred_at": "2026-10-20T12:15:00+07:00"},
	}}}}
	got := narrowReportFollowUpCards([]entity.AdvisorPart{{Type: "text", Text: "a"}, list}, message)
	view := got[1].Data.(map[string]any)["view"].(map[string]any)
	if len(got) != 2 || len(view["items"].([]any)) != 1 || view["total_count"] != 1 || view["next_cursor"] != nil {
		t.Fatalf("scope filter = %+v, want the single in-scope row without a stale cursor", got)
	}
	unrelated := entity.AdvisorPart{Type: "transaction_list", Data: map[string]any{"view": map[string]any{"items": []any{map[string]any{"category_id": "food", "type": "expense", "occurred_at": "2026-11-01T12:15:00+07:00"}}}}}
	if dropped := narrowReportFollowUpCards([]entity.AdvisorPart{{Type: "text", Text: "a"}, unrelated}, message); len(dropped) != 1 {
		t.Fatalf("a list with no in-scope rows must be dropped, got %d parts", len(dropped))
	}
}
