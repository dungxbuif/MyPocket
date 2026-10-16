package entity

import (
	"errors"
	"math"
	"testing"
	"time"
)

func savingsTime(s string) time.Time { t, _ := time.Parse(time.RFC3339, s); return t }
func savingsString(s string) *string { return &s }
func savingsAmount(v int64) *int64   { return &v }

func TestSavingsOpeningBalanceStartsWhenWalletIsCreated(t *testing.T) {
	now := savingsTime("2026-10-16T12:00:00Z")
	wallet := Wallet{ID: "new-goal", OwnerID: "owner", Type: WalletTypeGoal, OpeningBalance: 5000000, CreatedAt: now.Add(-time.Hour)}
	previous, err := CalculateSavingsReport("owner", []Wallet{wallet}, nil, "2026-09", "", now, time.UTC)
	if err != nil || previous.Goals.OpeningBalance != 0 || previous.Goals.ClosingBalance != 0 {
		t.Fatalf("future wallet baseline leaked into past: %+v %v", previous, err)
	}
	current, err := CalculateSavingsReport("owner", []Wallet{wallet}, nil, "2026-10", "", now, time.UTC)
	if err != nil || current.Goals.OpeningBalance != 0 || current.Goals.ClosingBalance != 5000000 || current.Goals.InitialBalanceAdded != 5000000 || current.Goals.Adjustments != 0 || current.Surplus.Net != 0 {
		t.Fatalf("creation baseline must be separate from ledger income/adjustment: %+v %v", current, err)
	}
	wallet.CreatedAt = savingsTime("2026-08-31T12:00:00Z")
	previous, err = CalculateSavingsReport("owner", []Wallet{wallet}, nil, "2026-09", "", now, time.UTC)
	if err != nil || previous.Goals.OpeningBalance != 5000000 || previous.Goals.InitialBalanceAdded != 0 {
		t.Fatalf("earlier baseline missing: %+v %v", previous, err)
	}
}

func TestSavingsSeparatesSurplusAndTransfers(t *testing.T) {
	now := savingsTime("2026-10-16T12:00:00Z")
	wallets := []Wallet{{ID: "cash", OwnerID: "owner", Type: WalletTypeBasic}, {ID: "goal1", OwnerID: "owner", Type: WalletTypeGoal, TargetAmount: savingsAmount(10000000)}, {ID: "goal2", OwnerID: "owner", Type: WalletTypeGoal, TargetAmount: savingsAmount(10000000)}}
	rows := []Transaction{
		{ID: "income", OwnerID: "owner", WalletID: "cash", Type: TransactionTypeIncome, Amount: 20000000, OccurredAt: now, IncludedInReports: true},
		{ID: "expense", OwnerID: "owner", WalletID: "cash", Type: TransactionTypeExpense, Amount: 12000000, OccurredAt: now, IncludedInReports: true},
		{ID: "out", OwnerID: "owner", WalletID: "cash", Type: TransactionTypeExpense, Amount: 5000000, OccurredAt: now, TransferID: savingsString("external")},
		{ID: "in", OwnerID: "owner", WalletID: "goal1", Type: TransactionTypeIncome, Amount: 5000000, OccurredAt: now, TransferID: savingsString("external")},
		{ID: "gout", OwnerID: "owner", WalletID: "goal1", Type: TransactionTypeExpense, Amount: 2000000, OccurredAt: now, TransferID: savingsString("internal")},
		{ID: "gin", OwnerID: "owner", WalletID: "goal2", Type: TransactionTypeIncome, Amount: 2000000, OccurredAt: now, TransferID: savingsString("internal")},
	}
	r, err := CalculateSavingsReport("owner", wallets, rows, "2026-10", "", now, time.UTC)
	if err != nil {
		t.Fatal(err)
	}
	if r.Surplus.Net != 8000000 || r.Surplus.TransactionCount != 2 || r.Surplus.SavingsRateBPS == nil || *r.Surplus.SavingsRateBPS != 4000 {
		t.Fatalf("surplus: %+v", r.Surplus)
	}
	if r.Goals.ClosingBalance != 5000000 || r.Goals.InternalTransfers != 5000000 || r.Goals.Deposits != 0 || r.Goals.Withdrawals != 0 {
		t.Fatalf("goals: %+v", r.Goals)
	}
	filtered, err := CalculateSavingsReport("owner", wallets, rows, "2026-10", "goal2", now, time.UTC)
	if err != nil || filtered.Goals.InternalTransfers != 2000000 || filtered.Surplus.Net != 8000000 {
		t.Fatalf("filtered=%+v err=%v", filtered, err)
	}
	internal, err := CalculateSavingsReport("owner", wallets, rows[4:], "2026-10", "", now, time.UTC)
	if err != nil || internal.Goals.InternalTransfers != 0 || internal.Goals.NetChange != 0 {
		t.Fatalf("internal=%+v err=%v", internal, err)
	}
}

func TestSavingsCalendarBalancesAndExclusions(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
	now := savingsTime("2024-02-29T12:00:00Z")
	wallets := []Wallet{{ID: "g", OwnerID: "owner", Type: WalletTypeGoal, OpeningBalance: 100, TargetAmount: savingsAmount(1000)}, {ID: "foreign", OwnerID: "other", Type: WalletTypeGoal, OpeningBalance: 999999}}
	rows := []Transaction{
		{ID: "before", OwnerID: "owner", WalletID: "g", Type: TransactionTypeIncome, Amount: 40, OccurredAt: savingsTime("2024-01-31T16:59:59Z")},
		{ID: "start", OwnerID: "owner", WalletID: "g", Type: TransactionTypeIncome, Amount: 100, OccurredAt: savingsTime("2024-01-31T17:00:00Z"), IncludedInReports: true},
		{ID: "withdraw", OwnerID: "owner", WalletID: "g", Type: TransactionTypeExpense, Amount: 30, OccurredAt: now, IncludedInReports: false},
		{ID: "adjust", OwnerID: "owner", WalletID: "g", Type: TransactionTypeAdjustment, AdjustmentDirection: savingsString(AdjustmentDirectionDecrease), Amount: 10, OccurredAt: now, IncludedInReports: true},
		{ID: "future", OwnerID: "owner", WalletID: "g", Type: TransactionTypeIncome, Amount: 10000, OccurredAt: now.Add(time.Second), IncludedInReports: true},
		{ID: "next-month", OwnerID: "owner", WalletID: "g", Type: TransactionTypeIncome, Amount: 10000, OccurredAt: savingsTime("2024-02-29T17:00:00Z"), IncludedInReports: true},
		{ID: "missing-date", OwnerID: "owner", WalletID: "g", Type: TransactionTypeIncome, Amount: 10000, IncludedInReports: true},
		{ID: "foreign-owner", OwnerID: "other", WalletID: "g", Type: TransactionTypeIncome, Amount: 10000, OccurredAt: now, IncludedInReports: true},
		{ID: "missing-wallet", OwnerID: "owner", WalletID: "deleted", Type: TransactionTypeIncome, Amount: 10000, OccurredAt: now, IncludedInReports: true},
		{ID: "principal", OwnerID: "owner", WalletID: "g", Type: TransactionTypeExpense, CreditKind: savingsString(CreditKindPayment), Amount: 20, OccurredAt: now, IncludedInReports: true},
	}
	r, err := CalculateSavingsReport("owner", wallets, rows, "2024-02", "", now, loc)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Goals.Items) != 1 || r.Goals.OpeningBalance != 140 || r.Goals.ClosingBalance != 180 || r.Goals.Deposits != 100 || r.Goals.Withdrawals != 50 || r.Goals.Adjustments != -10 || r.Goals.NetChange != 40 {
		t.Fatalf("goals: %+v", r.Goals)
	}
	if r.Goals.Items[0].Remaining != 820 || r.Goals.Items[0].ProgressBPS != 1800 || r.Goals.Items[0].Reached || r.Goals.Items[0].TransactionCount != 4 {
		t.Fatalf("item: %+v", r.Goals.Items[0])
	}
	if r.Surplus.Income != 100 || r.Surplus.Expense != 0 {
		t.Fatalf("surplus: %+v", r.Surplus)
	}
	page, err := SavingsReportTransactions("owner", wallets, rows, "2024-02", "", "goals", "all", 2, 1, now, loc)
	if err != nil || page.Total != 4 || len(page.Items) != 2 || page.Items[0].ID != "principal" || page.Items[1].ID != "adjust" {
		t.Fatalf("page=%+v err=%v", page, err)
	}
	page, err = SavingsReportTransactions("owner", wallets, rows, "2024-02", "", "surplus", "all", 50, 0, now, loc)
	if err != nil || page.Total != 1 || page.Items[0].ID != "start" {
		t.Fatalf("surplus page=%+v err=%v", page, err)
	}
}

func TestSavingsValidatesMonthsScopesAndOverflow(t *testing.T) {
	now := savingsTime("2026-10-16T12:00:00Z")
	wallets := []Wallet{{ID: "g", OwnerID: "owner", Type: WalletTypeGoal, OpeningBalance: math.MaxInt64}, {ID: "cash", OwnerID: "owner", Type: WalletTypeBasic}}
	for _, tc := range []struct {
		month, wallet string
		want          error
	}{{"2026-13", "", ErrSavingsInvalidQuery}, {"2026-1", "", ErrSavingsInvalidQuery}, {"2026-11", "", ErrSavingsFutureMonth}, {"2026-10", "cash", ErrSavingsWalletNotFound}, {"2026-10", "foreign", ErrSavingsWalletNotFound}} {
		_, err := CalculateSavingsReport("owner", wallets, nil, tc.month, tc.wallet, now, time.UTC)
		if !errors.Is(err, tc.want) {
			t.Fatalf("%+v: %v", tc, err)
		}
	}
	r, err := CalculateSavingsReport("owner", nil, nil, "", "", now, time.UTC)
	if err != nil || r.Month != "2026-10" || r.Surplus.SavingsRateBPS != nil || r.Goals.Items == nil {
		t.Fatalf("empty=%+v err=%v", r, err)
	}
	_, err = CalculateSavingsReport("owner", wallets, []Transaction{{OwnerID: "owner", WalletID: "g", Type: TransactionTypeIncome, Amount: 1, OccurredAt: now}}, "2026-10", "", now, time.UTC)
	if !errors.Is(err, ErrSavingsOverflow) {
		t.Fatalf("expected overflow, got %v", err)
	}
	for _, tc := range []struct {
		section, kind, wallet string
		limit, offset         int
	}{{"surplus", "deposit", "", 50, 0}, {"goals", "income", "", 50, 0}, {"surplus", "all", "g", 50, 0}, {"goals", "all", "", 0, 0}, {"goals", "all", "", 101, 0}, {"goals", "all", "", 50, -1}} {
		_, err := SavingsReportTransactions("owner", wallets, nil, "2026-10", tc.wallet, tc.section, tc.kind, tc.limit, tc.offset, now, time.UTC)
		if !errors.Is(err, ErrSavingsInvalidQuery) {
			t.Fatalf("invalid query accepted %+v: %v", tc, err)
		}
	}
}

func TestSavingsHistoricalMonthUsesCompleteLocalBoundsAndCurrentTargets(t *testing.T) {
	loc, _ := time.LoadLocation("America/New_York")
	now := savingsTime("2024-04-03T12:00:00Z")
	date, _ := ParseCalendarDate("2024-12-31")
	wallets := []Wallet{{ID: "g", OwnerID: "owner", Type: WalletTypeGoal, TargetAmount: savingsAmount(100), TargetDate: &date}}
	rows := []Transaction{
		{ID: "opening", OwnerID: "owner", WalletID: "g", Type: TransactionTypeIncome, Amount: 20, OccurredAt: savingsTime("2024-03-01T04:59:59Z")},
		{ID: "first", OwnerID: "owner", WalletID: "g", Type: TransactionTypeIncome, Amount: 80, OccurredAt: savingsTime("2024-03-01T05:00:00Z"), IncludedInReports: true},
		{ID: "last", OwnerID: "owner", WalletID: "g", Type: TransactionTypeIncome, Amount: 50, OccurredAt: savingsTime("2024-04-01T03:59:59Z"), IncludedInReports: true},
		{ID: "next", OwnerID: "owner", WalletID: "g", Type: TransactionTypeIncome, Amount: 9000, OccurredAt: savingsTime("2024-04-01T04:00:00Z"), IncludedInReports: true},
	}
	r, err := CalculateSavingsReport("owner", wallets, rows, "2024-03", "", now, loc)
	if err != nil || r.Goals.OpeningBalance != 20 || r.Goals.ClosingBalance != 150 || r.Goals.NetChange != 130 || r.Goals.Items[0].Remaining != 0 || r.Goals.Items[0].ProgressBPS != 15000 || !r.Goals.Items[0].Reached || r.Goals.Items[0].TargetDate.String() != "2024-12-31" {
		t.Fatalf("report=%+v err=%v", r, err)
	}
	p, err := SavingsReportTransactions("owner", wallets, rows, "2024-03", "", "goals", "deposit", 50, 0, now, loc)
	if err != nil || p.Total != 2 || p.Items[0].ID != "last" || p.Items[1].ID != "first" {
		t.Fatalf("page=%+v err=%v", p, err)
	}
}

func TestSavingsZeroIncomeAndNegativeBalances(t *testing.T) {
	now := savingsTime("2026-10-16T12:00:00Z")
	wallets := []Wallet{{ID: "g", OwnerID: "owner", Type: WalletTypeGoal, TargetAmount: savingsAmount(100)}}
	rows := []Transaction{{ID: "expense", OwnerID: "owner", WalletID: "g", Type: TransactionTypeExpense, Amount: 20, OccurredAt: now, IncludedInReports: true}}
	r, err := CalculateSavingsReport("owner", wallets, rows, "2026-10", "", now, time.UTC)
	if err != nil || r.Surplus.Net != -20 || r.Surplus.SavingsRateBPS != nil || r.Goals.Items[0].Remaining != 120 || r.Goals.Items[0].ProgressBPS != 0 || r.Goals.Items[0].Reached {
		t.Fatalf("report=%+v err=%v", r, err)
	}
}
