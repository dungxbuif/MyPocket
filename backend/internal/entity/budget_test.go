package entity

import (
	"testing"
	"time"
)

func TestBudgetProgressUsesScopeChildrenAndDeduplicatesSummary(t *testing.T) {
	start := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)
	parent, child, wallet := "food", "coffee", "w1"
	budgets := []Budget{{ID: "parent", LimitAmount: 1000, CategoryID: &parent, StartAt: start, EndAt: end}, {ID: "child", LimitAmount: 500, CategoryID: &child, WalletID: &wallet, StartAt: start, EndAt: end}}
	categories := []Category{{ID: parent}, {ID: child, ParentID: &parent}}
	txs := []Transaction{
		{ID: "1", WalletID: "w1", CategoryID: &child, Type: "expense", Amount: 100, OccurredAt: start, IncludedInReports: true},
		{ID: "2", WalletID: "w2", CategoryID: &child, Type: "expense", Amount: 200, OccurredAt: start, IncludedInReports: true},
		{ID: "3", WalletID: "w1", CategoryID: &child, Type: "income", Amount: 900, OccurredAt: start, IncludedInReports: true},
		{ID: "4", WalletID: "w1", CategoryID: &child, Type: "expense", Amount: 800, OccurredAt: start, IncludedInReports: false},
		{ID: "5", WalletID: "w1", CategoryID: &child, Type: "expense", Amount: 700, OccurredAt: end, IncludedInReports: true},
	}
	result := CalculateBudgets(budgets, txs, categories, start)
	if result.Items[0].Spent != 300 || result.Items[1].Spent != 100 || result.Spent != 300 || result.LimitAmount != 1500 {
		t.Fatalf("wrong progress: %+v", result)
	}
	if result.Items[0].DaysRemaining != 30 {
		t.Fatalf("days=%d", result.Items[0].DaysRemaining)
	}
	txs[0].Amount = 50
	result = CalculateBudgets(budgets, txs, categories, end)
	if result.Items[0].Spent != 250 || !result.Items[0].Ended || result.Items[0].DaysRemaining != 0 {
		t.Fatalf("historical edit not reflected: %+v", result)
	}
}
