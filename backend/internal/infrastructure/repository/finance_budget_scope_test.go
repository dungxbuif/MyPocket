package repository

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	"testing"
	"time"
)

// Catches historical rows crowding the card, active_on acting only as a clock,
// and totals still describing budgets removed by the selection.
func TestFinanceBudgetProgressFiltersBeforePresentation(t *testing.T) {
	db := openReportsPostgresTestDB(t)
	owner := insertReportsTestUser(t, db, "Asia/Ho_Chi_Minh")
	foreign := insertReportsTestUser(t, db, "Asia/Ho_Chi_Minh")
	location, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		t.Fatal(err)
	}
	wallet := budgetScopeWallet(t, db, owner)
	ids := map[string]string{}
	for _, row := range []struct {
		name, start, end, owner string
		limit                   int64
	}{
		{"old", "2026-01-01", "2026-01-31", owner, 100},
		{"monthly", "2026-11-01", "2026-11-30", owner, 1000},
		{"weekly", "2026-11-02", "2026-11-08", owner, 200},
		{"late", "2026-11-20", "2026-11-25", owner, 300},
		{"foreign", "2026-11-01", "2026-11-30", foreign, 9000},
	} {
		b := entity.Budget{ID: uuid.NewString(), OwnerID: row.owner, Name: row.name, LimitAmount: row.limit, StartDate: budgetScopeDate(t, row.start), EndDate: budgetScopeDate(t, row.end)}
		if err := db.Create(&b).Error; err != nil {
			t.Fatal(err)
		}
		ids[row.name] = b.ID
	}
	asOf := time.Date(2026, 11, 22, 12, 0, 0, 0, location)
	budgetScopeExpense(t, db, owner, wallet, nil, 50, time.Date(2026, 11, 2, 23, 0, 0, 0, location))
	budgetScopeExpense(t, db, owner, wallet, nil, 70, time.Date(2026, 11, 10, 12, 0, 0, 0, location))
	reader := &FinanceQueryPostgresRepository{db: db}
	for _, tc := range []struct {
		name         string
		query        entity.NormalizedQuery
		names        []string
		limit, spent int64
	}{
		{"default current month", entity.NormalizedQuery{}, []string{"monthly", "weekly", "late"}, 1500, 120},
		{"requested name", entity.NormalizedQuery{Month: "2026-11", BudgetNameQuery: " WEEK "}, []string{"weekly"}, 200, 50},
		{"requested month", entity.NormalizedQuery{Month: "2026-11"}, []string{"monthly", "weekly", "late"}, 1500, 120},
		{"active day includes entire local day", entity.NormalizedQuery{ActiveOn: "2026-11-02"}, []string{"monthly", "weekly"}, 1200, 50},
		{"specific budget", entity.NormalizedQuery{Month: "2026-11", ObjectIDs: []string{ids["weekly"]}}, []string{"weekly"}, 200, 50},
		{"explicit historical ID", entity.NormalizedQuery{ObjectIDs: []string{ids["old"]}}, []string{"old"}, 100, 0},
		{"ID and month intersect", entity.NormalizedQuery{Month: "2026-11", ObjectIDs: []string{ids["old"]}}, nil, 0, 0},
		{"foreign ID", entity.NormalizedQuery{ObjectIDs: []string{ids["foreign"]}}, nil, 0, 0},
		{"empty month", entity.NormalizedQuery{Month: "2026-02"}, nil, 0, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.query.Key = "q"
			result, err := reader.readBudgetProgress(context.Background(), db, owner, location, asOf, tc.query)
			if err != nil {
				t.Fatal(err)
			}
			var got entity.BudgetSummary
			if err := json.Unmarshal(result.View, &got); err != nil {
				t.Fatal(err)
			}
			if len(got.Items) != len(tc.names) || got.LimitAmount != tc.limit || got.Spent != tc.spent {
				t.Fatalf("view=%+v, want %v limit=%d spent=%d", got, tc.names, tc.limit, tc.spent)
			}
			wanted := map[string]bool{}
			for _, name := range tc.names {
				wanted[ids[name]] = true
			}
			for _, b := range got.Items {
				if !wanted[b.ID] {
					t.Fatalf("unrelated budget=%+v", b)
				}
			}
			for _, fact := range result.Facts {
				if fact.ID == "q.remaining" && (fact.Integer == nil || *fact.Integer != tc.limit-tc.spent) {
					t.Fatalf("unscoped remaining=%+v", fact)
				}
			}
		})
	}
	for _, query := range []entity.NormalizedQuery{{Month: "2026-13"}, {ActiveOn: "2026-02-30"}} {
		if _, err := reader.readBudgetProgress(context.Background(), db, owner, location, asOf, query); err == nil {
			t.Fatalf("invalid date accepted: %+v", query)
		}
	}
}
