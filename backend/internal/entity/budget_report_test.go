package entity

import (
	"errors"
	"testing"
	"time"
)

func reportBudget(t *testing.T, start, end string) Budget {
	t.Helper()
	from, err := ParseCalendarDate(start)
	if err != nil {
		t.Fatal(err)
	}
	to, err := ParseCalendarDate(end)
	if err != nil {
		t.Fatal(err)
	}
	return Budget{ID: "budget", OwnerID: "owner", LimitAmount: 4000000, StartDate: from, EndDate: to}
}

func TestBudgetReportLinearEstimateAndZeroDays(t *testing.T) {
	b := reportBudget(t, "2026-09-01", "2026-09-30")
	now := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	txs := []Transaction{{ID: "1", OwnerID: "owner", Type: TransactionTypeExpense, Amount: 300000, IncludedInReports: true, OccurredAt: now.AddDate(0, 0, -1)}}
	r, err := CalculateBudgetReport(b, txs, nil, now, time.UTC)
	if err != nil {
		t.Fatal(err)
	}
	if r.Spent != 300000 || r.Forecast.Status != "estimated" || r.Forecast.ElapsedDays != 3 || r.Forecast.TotalDays != 30 || r.Forecast.ProjectedSpent == nil || *r.Forecast.ProjectedSpent != 3000000 {
		t.Fatalf("incorrect estimate: %+v", r)
	}
	if len(r.Daily) != 3 || r.Daily[0].Spent != 0 || r.Daily[1].Spent != 300000 || r.Daily[2].Spent != 0 {
		t.Fatalf("missing zero days: %+v", r.Daily)
	}
}

func TestBudgetReportFutureEndedAndCeiling(t *testing.T) {
	b := reportBudget(t, "2026-09-01", "2026-09-30")
	before := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	r, err := CalculateBudgetReport(b, nil, nil, before, time.UTC)
	if err != nil || r.Forecast.Status != "not_started" || r.Forecast.ProjectedSpent != nil || len(r.Daily) != 0 {
		t.Fatalf("future fabricated: %+v %v", r, err)
	}
	now := time.Date(2026, 9, 7, 12, 0, 0, 0, time.UTC)
	txs := []Transaction{{ID: "1", OwnerID: "owner", Type: TransactionTypeExpense, Amount: 1, IncludedInReports: true, OccurredAt: now}}
	r, err = CalculateBudgetReport(b, txs, nil, now, time.UTC)
	if err != nil || r.Forecast.ProjectedSpent == nil || *r.Forecast.ProjectedSpent != 5 {
		t.Fatalf("estimate must ceil 30/7: %+v %v", r, err)
	}
	r, err = CalculateBudgetReport(b, txs, nil, now.AddDate(0, 1, 0), time.UTC)
	if err != nil || r.Forecast.Status != "actual" || *r.Forecast.ProjectedSpent != 1 || len(r.Daily) != 30 {
		t.Fatalf("ended must be actual: %+v %v", r, err)
	}
}

func TestBudgetReportMatchingDescendantsOwnerExclusionsAndFuture(t *testing.T) {
	b := reportBudget(t, "2026-09-01", "2026-09-30")
	wallet, parent, child, transfer := "wallet", "parent", "child", "pair"
	b.WalletID = &wallet
	b.CategoryID = &parent
	now := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	good := Transaction{ID: "good", OwnerID: "owner", WalletID: wallet, CategoryID: &child, Type: TransactionTypeExpense, Amount: 10, IncludedInReports: true, OccurredAt: now}
	rows := []Transaction{good}
	for _, kind := range []string{"foreign", "wallet", "category", "excluded", "transfer", "future", "income", "before"} {
		row := good
		row.ID = kind
		switch kind {
		case "foreign":
			row.OwnerID = "other"
		case "wallet":
			row.WalletID = "other"
		case "category":
			row.CategoryID = nil
		case "excluded":
			row.IncludedInReports = false
		case "transfer":
			row.TransferID = &transfer
		case "future":
			row.OccurredAt = now.Add(time.Second)
		case "income":
			row.Type = TransactionTypeIncome
		case "before":
			row.OccurredAt = now.AddDate(0, -1, 0)
		}
		rows = append(rows, row)
	}
	cats := []Category{{ID: child, ParentID: &parent}, {ID: parent}}
	matched := MatchBudgetTransactions(b, rows, cats, now, time.UTC)
	if len(matched) != 1 || matched[0].ID != "good" {
		t.Fatalf("scope mismatch: %+v", matched)
	}
	r, err := CalculateBudgetReport(b, rows, cats, now, time.UTC)
	if err != nil || r.Spent != 10 {
		t.Fatalf("bad total: %+v %v", r, err)
	}
}

func TestBudgetReportCalendarDSTAndCheckedMoney(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	b := reportBudget(t, "2026-03-01", "2026-03-31")
	now := time.Date(2026, 3, 9, 0, 30, 0, 0, loc)
	r, err := CalculateBudgetReport(b, nil, nil, now, loc)
	if err != nil || r.Forecast.ElapsedDays != 9 || len(r.Daily) != 9 || r.Forecast.TotalDays != 31 {
		t.Fatalf("DST changed calendar days: %+v %v", r, err)
	}
	txs := []Transaction{{ID: "large", OwnerID: "owner", Type: TransactionTypeExpense, Amount: MaxSafeMoneyAmount, IncludedInReports: true, OccurredAt: now}}
	if _, err = CalculateBudgetReport(b, txs, nil, now, loc); !errors.Is(err, ErrFinanceAmountInvalid) {
		t.Fatalf("overflow estimate not rejected: %v", err)
	}
}
