package repository

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
)

func TestFinanceBudgetToolSelectedPeriodRetainsPriorCarryover(t *testing.T) {
	db, owner, seriesRepo, now := budgetSeriesFixture(t)
	start, _ := entity.ParseCalendarDate("2026-10-01")
	series := entity.BudgetSeries{ID: uuid.NewString(), Cadence: "monthly", StartDate: start}
	cfg := entity.BudgetSeriesConfig{Name: "Test", LimitAmount: 100, CarryoverEnabled: true}
	if _, _, err := seriesRepo.Mutate(owner, &series, &cfg, "create", "create-key", "hash", 0, now); err != nil {
		t.Fatal(err)
	}
	asOf := now.AddDate(0, 1, 0)
	if _, err := seriesRepo.Ensure(owner, asOf); err != nil {
		t.Fatal(err)
	}
	budgets, err := NewBudgetPostgresRepository(db).List(owner)
	if err != nil || len(budgets) != 2 {
		t.Fatalf("periods %v %v", budgets, err)
	}
	selected := budgets[0]
	reader := &FinanceQueryPostgresRepository{db: db}
	result, err := reader.readBudgetProgress(context.Background(), db, owner, time.UTC, asOf, entity.NormalizedQuery{Key: "q", ObjectIDs: []string{selected.ID}})
	if err != nil {
		t.Fatal(err)
	}
	var summary entity.BudgetSummary
	if err = json.Unmarshal(result.View, &summary); err != nil {
		t.Fatal(err)
	}
	if len(summary.Items) != 1 || summary.Items[0].ID != selected.ID || summary.Items[0].CarryoverAmount != 100 || summary.Items[0].EffectiveLimitAmount != 200 || summary.LimitAmount != 200 {
		t.Fatalf("tool lost carry source: %+v", summary)
	}
}
