package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	"testing"
	"time"
)

func TestMonthlyContributorEvidenceUsesDistinctTargetAndMatchedPeriod(t *testing.T) {
	db := openReportsPostgresTestDB(t)
	owner := insertReportsTestUser(t, db, "UTC")
	wallet := budgetScopeWallet(t, db, owner)
	category := entity.Category{ID: uuid.NewString(), OwnerID: &owner, Kind: "expense", Name: "Food"}
	if err := db.Create(&category).Error; err != nil {
		t.Fatal(err)
	}
	secondCategory := entity.Category{ID: uuid.NewString(), OwnerID: &owner, Kind: entity.TransactionTypeExpense, Name: "Transport"}
	if err := db.Create(&secondCategory).Error; err != nil {
		t.Fatal(err)
	}
	for _, categoryID := range []*string{&category.ID, &secondCategory.ID} {
		budgetScopeExpense(t, db, owner, wallet, categoryID, 100, time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC))
		budgetScopeExpense(t, db, owner, wallet, categoryID, 200, time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC))
		budgetScopeExpense(t, db, owner, wallet, categoryID, 700, time.Date(2026, 11, 2, 1, 0, 0, 0, time.UTC))
	}
	asOf := time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC)
	matchedEnd := time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC)
	repo := NewMonthlyReportPostgresRepository(db)
	scope, err := repo.resolveScope(context.Background(), db, owner, entity.MonthlyReportScope{Month: "2026-11", WalletID: wallet}, asOf)
	if err != nil {
		t.Fatal(err)
	}
	categories, contributors, err := repo.categoriesAndContributors(context.Background(), db, owner, scope)
	if err != nil {
		t.Fatal(err)
	}
	if len(categories) != 2 || len(contributors) != 2 {
		t.Fatalf("categories=%+v contributors=%+v", categories, contributors)
	}
	targets := map[string]bool{}
	for _, row := range categories {
		targets[row.TargetID] = true
		if row.Amount != 900 || !row.Drilldown.EndExclusive.Equal(asOf) {
			t.Fatalf("category must retain full range: %+v", row)
		}
	}
	for _, row := range contributors {
		page, err := repo.ReadMonthlyTransactionsByDrilldown(context.Background(), owner, scope.Scope, row.Drilldown, 100, 0, asOf)
		if err != nil {
			t.Fatal(err)
		}
		var sum int64
		for _, transaction := range page.Items {
			sum += transaction.Amount
		}
		if sum != row.CurrentAmount {
			t.Fatalf("actual matched drilldown=%d expected=%d", sum, row.CurrentAmount)
		}
		if row.CategoryID == nil || (*row.CategoryID != category.ID && *row.CategoryID != secondCategory.ID) {
			t.Fatalf("unexpected contributor category: %+v", row)
		}
		suffix := *row.CategoryID
		if row.TargetID != "comparison:category:"+suffix || targets[row.TargetID] || row.Drilldown.TargetID != row.TargetID || !row.Drilldown.EndExclusive.Equal(matchedEnd) || !row.Drilldown.From.Equal(time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)) || row.Drilldown.WalletID != wallet || row.CurrentAmount != 200 || row.PreviousAmount != 100 {
			t.Fatalf("comparison evidence mismatch: %+v", row)
		}
	}
}
