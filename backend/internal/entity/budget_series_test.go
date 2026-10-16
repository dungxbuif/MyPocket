package entity

import (
	"testing"
	"time"
)

func TestBudgetCadenceCalendar(t *testing.T) {
	for _, tc := range []struct{ cadence, start, next string }{{"weekly", "2024-02-26", "2024-03-04"}, {"monthly", "2024-02-01", "2024-03-01"}, {"quarterly", "2024-10-01", "2025-01-01"}, {"yearly", "2024-01-01", "2025-01-01"}} {
		start, _ := ParseCalendarDate(tc.start)
		next, err := NextBudgetPeriod(start, tc.cadence)
		if err != nil || next.Format("2006-01-02") != tc.next {
			t.Fatalf("%+v: %v %v", tc, next, err)
		}
	}
	bad, _ := ParseCalendarDate("2024-02-02")
	if _, err := NextBudgetPeriod(bad, "monthly"); err == nil {
		t.Fatal("noncanonical start accepted")
	}
}

func TestBudgetCarryChainRecalculatesAndNeverCarriesDebt(t *testing.T) {
	sid := "series"
	jan, _ := ParseCalendarDate("2026-01-01")
	janEnd, _ := ParseCalendarDate("2026-01-31")
	feb, _ := ParseCalendarDate("2026-02-01")
	febEnd, _ := ParseCalendarDate("2026-02-28")
	rows := []Budget{{ID: "feb", SeriesID: &sid, StartDate: feb, EndDate: febEnd, LimitAmount: 100, CarryoverEnabled: true}, {ID: "jan", SeriesID: &sid, StartDate: jan, EndDate: janEnd, LimitAmount: 100, CarryoverEnabled: true}}
	txs := []Transaction{{ID: "expense", Type: TransactionTypeExpense, Amount: 30, IncludedInReports: true, OccurredAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)}}
	got, err := CalculateBudgetsChecked(rows, txs, nil, time.Date(2026, 2, 3, 0, 0, 0, 0, time.UTC), time.UTC)
	if err != nil || got.Items[0].CarryoverAmount != 70 || got.LimitAmount != 170 {
		t.Fatalf("%+v %v", got, err)
	}
	txs[0].Amount = 130
	got, err = CalculateBudgetsChecked(rows, txs, nil, time.Date(2026, 2, 3, 0, 0, 0, 0, time.UTC), time.UTC)
	if err != nil || got.Items[0].CarryoverAmount != 0 {
		t.Fatalf("debt: %+v %v", got, err)
	}
	rows[0].LimitAmount = MaxBudgetMoney
	txs[0].Amount = 0
	if _, err = CalculateBudgetsChecked(rows, txs, nil, time.Date(2026, 2, 3, 0, 0, 0, 0, time.UTC), time.UTC); err == nil {
		t.Fatal("overflow accepted")
	}
}

func TestBudgetCarryResetsOnScopeChangeAndDisabledPeriod(t *testing.T) {
	sid := "series"
	first, _ := ParseCalendarDate("2026-01-01")
	end, _ := ParseCalendarDate("2026-01-31")
	second, _ := ParseCalendarDate("2026-02-01")
	secondEnd, _ := ParseCalendarDate("2026-02-28")
	a, b := "a", "b"
	for _, tc := range []struct {
		name              string
		previous, current Budget
	}{
		{"wallet", Budget{WalletID: &a}, Budget{WalletID: &b, CarryoverEnabled: true}},
		{"category", Budget{CategoryID: &a}, Budget{CategoryID: &b, CarryoverEnabled: true}},
		{"nil scope", Budget{WalletID: &a}, Budget{CarryoverEnabled: true}},
		{"disabled", Budget{}, Budget{CarryoverEnabled: false}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			prev, current := tc.previous, tc.current
			prev.ID = "jan"
			prev.SeriesID = &sid
			prev.StartDate = first
			prev.EndDate = end
			prev.LimitAmount = 100
			current.ID = "feb"
			current.SeriesID = &sid
			current.StartDate = second
			current.EndDate = secondEnd
			current.LimitAmount = 100
			got, err := CalculateBudgetsChecked([]Budget{prev, current}, nil, nil, time.Date(2026, 2, 3, 0, 0, 0, 0, time.UTC), time.UTC)
			if err != nil || got.Items[1].CarryoverAmount != 0 {
				t.Fatalf("carry transferred scopes: %+v %v", got, err)
			}
		})
	}
}

func TestBudgetCarryGapDoesNotBridgeMissingPeriods(t *testing.T) {
	sid := "series"
	jan, _ := ParseCalendarDate("2026-01-01")
	janEnd, _ := ParseCalendarDate("2026-01-31")
	march, _ := ParseCalendarDate("2026-03-01")
	marchEnd, _ := ParseCalendarDate("2026-03-31")
	rows := []Budget{{ID: "jan", SeriesID: &sid, StartDate: jan, EndDate: janEnd, LimitAmount: 100, CarryoverEnabled: true}, {ID: "march", SeriesID: &sid, StartDate: march, EndDate: marchEnd, LimitAmount: 100, CarryoverEnabled: true}}
	got, err := CalculateBudgetsChecked(rows, nil, nil, time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC), time.UTC)
	if err != nil || got.Items[1].CarryoverAmount != 0 {
		t.Fatalf("gap %+v %v", got, err)
	}
}
