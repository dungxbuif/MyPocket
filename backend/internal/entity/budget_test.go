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

func TestBudgetProgressUsesAccountLocalInclusiveCalendarInterval(t *testing.T) {
	location, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		t.Fatal(err)
	}
	start := CalendarDate{Time: dateAt(2026, time.September, 1)}
	end := CalendarDate{Time: dateAt(2026, time.September, 30)}
	categoryID := "food"
	budget := Budget{ID: "september", LimitAmount: 10000, CategoryID: &categoryID, StartDate: start, EndDate: end}
	transactions := []Transaction{
		// 23:59:59 on September 30 in the account timezone: included.
		{ID: "last-local-day", CategoryID: &categoryID, Type: TransactionTypeExpense, Amount: 1200, IncludedInReports: true, OccurredAt: time.Date(2026, time.September, 30, 16, 59, 59, 0, time.UTC)},
		// 00:00:00 on October 1 in the account timezone: excluded.
		{ID: "next-local-day", CategoryID: &categoryID, Type: TransactionTypeExpense, Amount: 3400, IncludedInReports: true, OccurredAt: time.Date(2026, time.September, 30, 17, 0, 0, 0, time.UTC)},
	}
	result := CalculateBudgetsInLocation([]Budget{budget}, transactions, []Category{{ID: categoryID}}, time.Date(2026, time.September, 30, 12, 0, 0, 0, time.UTC), location)
	if result.Items[0].Spent != 1200 || result.Spent != 1200 {
		t.Fatalf("budget must include only the account-local inclusive interval: %+v", result)
	}
}
