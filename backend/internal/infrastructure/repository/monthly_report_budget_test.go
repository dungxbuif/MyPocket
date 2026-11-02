package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	"gorm.io/gorm"
)

func budgetScopeDate(t *testing.T, value string) entity.CalendarDate {
	t.Helper()
	date, err := entity.ParseCalendarDate(value)
	if err != nil {
		t.Fatal(err)
	}
	return date
}

func budgetScopeWallet(t *testing.T, db *gorm.DB, owner string) string {
	t.Helper()
	wallet := entity.Wallet{ID: uuid.NewString(), OwnerID: owner, Name: "Budget scope", Type: entity.WalletTypeBasic, Currency: "VND"}
	if err := db.Create(&wallet).Error; err != nil {
		t.Fatal(err)
	}
	return wallet.ID
}

func budgetScopeExpense(t *testing.T, db *gorm.DB, owner, wallet string, category *string, amount int64, at time.Time) entity.Transaction {
	t.Helper()
	if category == nil {
		fallback := entity.Category{ID: uuid.NewString(), OwnerID: &owner, Kind: entity.TransactionTypeExpense, Name: "Fallback expense"}
		if err := db.Create(&fallback).Error; err != nil {
			t.Fatal(err)
		}
		category = &fallback.ID
	}
	row := entity.Transaction{ID: uuid.NewString(), OwnerID: owner, WalletID: wallet, CategoryID: category, Type: entity.TransactionTypeExpense, Amount: amount, OccurredAt: at, IncludedInReports: true}
	if err := db.Create(&row).Error; err != nil {
		t.Fatal(err)
	}
	return row
}

// Catches first-of-month selection, month-wide spend, UTC date conversion and
// evidence that drops the budget wallet or descendant category scope.
func TestMonthlyBudgetPreviewCanonicalPeriodAndExactEvidence(t *testing.T) {
	db := openReportsPostgresTestDB(t)
	owner := insertReportsTestUser(t, db, "Asia/Ho_Chi_Minh")
	foreign := insertReportsTestUser(t, db, "Asia/Ho_Chi_Minh")
	location, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		t.Fatal(err)
	}
	wallet, otherWallet := budgetScopeWallet(t, db, owner), budgetScopeWallet(t, db, owner)
	root := entity.Category{ID: uuid.NewString(), OwnerID: &owner, Kind: "expense", Name: "Food"}
	child := entity.Category{ID: uuid.NewString(), OwnerID: &owner, ParentID: &root.ID, Kind: "expense", Name: "Lunch"}
	for _, category := range []entity.Category{root, child} {
		if err := db.Create(&category).Error; err != nil {
			t.Fatal(err)
		}
	}
	at := func(day, hour int) time.Time { return time.Date(2026, 11, day, hour, 0, 0, 0, location) }
	rows := []struct {
		start, end string
		spent      int64
	}{
		{"2026-11-02", "2026-11-08", 70},
		{"2026-11-15", "2026-11-20", 30},
		{"2026-10-29", "2026-11-04", 110},
	}
	budgets := []entity.Budget{}
	for _, row := range rows {
		b := entity.Budget{ID: uuid.NewString(), OwnerID: owner, Name: row.start, LimitAmount: 100, WalletID: &wallet, CategoryID: &root.ID, StartDate: budgetScopeDate(t, row.start), EndDate: budgetScopeDate(t, row.end)}
		if err := db.Create(&b).Error; err != nil {
			t.Fatal(err)
		}
		budgets = append(budgets, b)
	}
	for _, row := range []struct{ owner, start, end string }{{owner, "2026-10-01", "2026-10-31"}, {owner, "2026-12-01", "2026-12-31"}, {foreign, "2026-11-01", "2026-11-30"}} {
		b := entity.Budget{ID: uuid.NewString(), OwnerID: row.owner, Name: "not visible", LimitAmount: 10, StartDate: budgetScopeDate(t, row.start), EndDate: budgetScopeDate(t, row.end)}
		if err := db.Create(&b).Error; err != nil {
			t.Fatal(err)
		}
	}
	budgetScopeExpense(t, db, owner, wallet, &child.ID, 40, time.Date(2026, 10, 30, 12, 0, 0, 0, location))
	budgetScopeExpense(t, db, owner, wallet, &child.ID, 70, at(2, 0))
	budgetScopeExpense(t, db, owner, wallet, &child.ID, 30, at(15, 12))
	budgetScopeExpense(t, db, owner, wallet, &child.ID, 900, at(9, 0))
	budgetScopeExpense(t, db, owner, otherWallet, &child.ID, 800, at(3, 12))
	budgetScopeExpense(t, db, owner, wallet, nil, 700, at(3, 12))
	excluded := budgetScopeExpense(t, db, owner, wallet, &child.ID, 600, at(3, 12))
	if err := db.Model(&excluded).Update("included_in_reports", false).Error; err != nil {
		t.Fatal(err)
	}
	repo := NewMonthlyReportPostgresRepository(db)
	scope, err := repo.resolveScope(context.Background(), db, owner, entity.MonthlyReportScope{Month: "2026-11"}, time.Date(2026, 12, 10, 0, 0, 0, 0, location))
	if err != nil {
		t.Fatal(err)
	}
	got, err := repo.budgetPreviews(context.Background(), db, owner, scope, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("previews=%+v, want three overlapping owned periods", got)
	}
	byID := map[string]entity.ReportBudgetPreview{}
	for _, preview := range got {
		byID[preview.ID] = preview
	}
	for i, b := range budgets {
		preview := byID[b.ID]
		start := time.Date(b.StartDate.Year(), b.StartDate.Month(), b.StartDate.Day(), 0, 0, 0, 0, location)
		end := time.Date(b.EndDate.Year(), b.EndDate.Month(), b.EndDate.Day(), 0, 0, 0, 0, location).AddDate(0, 0, 1)
		if preview.Spent != rows[i].spent || preview.Remaining != 100-rows[i].spent || preview.OverLimit != (rows[i].spent > 100) || !preview.From.Equal(start) || !preview.EndExclusive.Equal(end) {
			t.Fatalf("period preview=%+v", preview)
		}
		d := preview.Drilldown
		if d.TargetID != "budget:"+b.ID || d.WalletID != wallet || d.Type != "expense" || d.CategoryBucket != "root" || len(d.CategoryIDs) != 2 || !d.From.Equal(start) || !d.EndExclusive.Equal(end) {
			t.Fatalf("evidence=%+v", d)
		}
		var sum int64
		if err := db.Model(&entity.Transaction{}).Select("COALESCE(SUM(amount),0)").Where("owner_id = ? AND wallet_id = ? AND occurred_at >= ? AND occurred_at < ? AND type = ? AND category_id IN ? AND included_in_reports = TRUE AND transfer_id IS NULL AND credit_payment_id IS NULL", owner, d.WalletID, d.From, d.EndExclusive, d.Type, d.CategoryIDs).Scan(&sum).Error; err != nil {
			t.Fatal(err)
		}
		if sum != preview.Spent {
			t.Fatalf("evidence sum=%d, preview=%d", sum, preview.Spent)
		}
		page, err := repo.ReadMonthlyTransactionsByDrilldown(context.Background(), owner, scope.Scope, d, 100, 0, time.Date(2026, 12, 10, 0, 0, 0, 0, location))
		if err != nil {
			t.Fatal(err)
		}
		var pageSum int64
		for _, row := range page.Items {
			pageSum += row.Amount
		}
		if pageSum != preview.Spent {
			t.Fatalf("actual drilldown=%d preview=%d", pageSum, preview.Spent)
		}
	}
	// A wallet report must not include another wallet's budget.
	scope.Scope.WalletID = otherWallet
	got, err = repo.budgetPreviews(context.Background(), db, owner, scope, 0)
	if err != nil || len(got) != 0 {
		t.Fatalf("wallet isolation: %+v %v", got, err)
	}
}

func TestMonthlyBudgetPreviewCarryoverAndExclusiveAsOf(t *testing.T) {
	db, owner, seriesRepo, now := budgetSeriesFixture(t)
	wallet := budgetScopeWallet(t, db, owner)
	series := entity.BudgetSeries{ID: uuid.NewString(), Cadence: "monthly", StartDate: budgetScopeDate(t, "2026-10-01")}
	cfg := entity.BudgetSeriesConfig{Name: "Carry", LimitAmount: 1000, CarryoverEnabled: true, WalletID: &wallet}
	if _, _, err := seriesRepo.Mutate(owner, &series, &cfg, "create", "carry-preview", "hash", 0, now); err != nil {
		t.Fatal(err)
	}
	asOf := time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC)
	if _, err := seriesRepo.Ensure(owner, asOf); err != nil {
		t.Fatal(err)
	}
	budgetScopeExpense(t, db, owner, wallet, nil, 100, time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC))
	budgetScopeExpense(t, db, owner, wallet, nil, 350, asOf.Add(-time.Hour))
	budgetScopeExpense(t, db, owner, wallet, nil, 500, asOf)
	repo := NewMonthlyReportPostgresRepository(db)
	scope, err := repo.resolveScope(context.Background(), db, owner, entity.MonthlyReportScope{Month: "2026-11"}, asOf)
	if err != nil {
		t.Fatal(err)
	}
	got, err := repo.budgetPreviews(context.Background(), db, owner, scope, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].LimitAmount != 1900 || got[0].Spent != 350 || got[0].Remaining != 1550 || !got[0].EndExclusive.Equal(asOf) {
		t.Fatalf("carryover/as-of=%+v", got)
	}
}
