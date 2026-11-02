package usecase

import (
	"context"
	"encoding/json"
	"github.com/mypocket/backend/internal/entity"
	"testing"
)

func TestAdvisorBudgetToolPreservesRequestedMonthAndBudget(t *testing.T) {
	reader := &financeReaderStub{bundle: entity.FactBundle{Results: []entity.FinanceResult{{QueryKey: "q1"}}}}
	tool, _ := NewAdvisorToolRegistry(NewFinanceQueryService(reader)).Lookup("get_budget_progress")
	_, err := tool.Execute(context.Background(), Principal{OwnerID: "owner"}, json.RawMessage(`{"month":"2026-11","budget_ids":["11111111-1111-4111-8111-111111111111"],"active_on":"2026-11-02"}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(reader.queries) != 1 || reader.queries[0].Month != "2026-11" || reader.queries[0].ActiveOn != "2026-11-02" || len(reader.queries[0].ObjectIDs) != 1 {
		t.Fatalf("lost requested scope: %+v", reader.queries)
	}
}

func TestAdvisorBudgetToolNormalizesSingularIDAndNameQuery(t *testing.T) {
	reader := &financeReaderStub{bundle: entity.FactBundle{Results: []entity.FinanceResult{{QueryKey: "q1"}}}}
	tool, _ := NewAdvisorToolRegistry(NewFinanceQueryService(reader)).Lookup("get_budget_progress")
	_, err := tool.Execute(context.Background(), Principal{OwnerID: "owner"}, json.RawMessage(`{"month":"2026-11","budget_id":"11111111-1111-4111-8111-111111111111","query":" Shopping "}`))
	if err != nil {
		t.Fatal(err)
	}
	q := reader.queries[0]
	if q.Month != "2026-11" || q.BudgetNameQuery != "Shopping" || len(q.ObjectIDs) != 1 || q.ObjectIDs[0] != "11111111-1111-4111-8111-111111111111" {
		t.Fatalf("lost singular/name scope: %+v", q)
	}
	_, err = tool.Execute(context.Background(), Principal{OwnerID: "owner"}, json.RawMessage(`{"budget_id":"11111111-1111-4111-8111-111111111111","budget_ids":["22222222-2222-4222-8222-222222222222"]}`))
	if err == nil {
		t.Fatal("conflicting singular/plural IDs must not broaden selection")
	}
}
