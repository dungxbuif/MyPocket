package repository

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestFinanceCursorRejectsDifferentScope(t *testing.T) {
	issued, err := encodeFinanceCursor(time.Date(2026, 9, 10, 4, 0, 0, 0, time.UTC), "tx-1", "scope-a")
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := decodeFinanceCursor(issued, "scope-a")
	if err != nil || decoded.ID != "tx-1" {
		t.Fatalf("cursor did not decode: %+v %v", decoded, err)
	}
	if _, err := decodeFinanceCursor(issued, "scope-b"); err == nil {
		t.Fatal("cursor from another scope must be rejected")
	}
}

func TestFinanceQueryPostgresSummaryUsesOwnerAndReportScope(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires migrated local TEST_DATABASE_URL")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	owner, other := uuid.NewString(), uuid.NewString()
	walletID := uuid.NewString()
	categoryID := uuid.NewString()
	childCategoryID := uuid.NewString()
	otherCategoryID := uuid.NewString()
	now := time.Date(2026, 9, 10, 4, 0, 0, 0, time.UTC)
	if err := db.Create(&entity.User{ID: owner, Email: owner + "@finance.test", GoogleSubject: owner, Timezone: "Asia/Ho_Chi_Minh", TimezoneConfirmed: true}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&entity.User{ID: other, Email: other + "@finance.test", GoogleSubject: other, Timezone: "Asia/Ho_Chi_Minh", TimezoneConfirmed: true}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&entity.Wallet{ID: walletID, OwnerID: owner, Name: "Cash", Type: entity.WalletTypeBasic, Currency: entity.WalletCurrencyVND, IsInTotal: true}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&entity.Category{ID: categoryID, OwnerID: &owner, Kind: entity.TransactionTypeExpense, Name: "Food"}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&entity.Category{ID: childCategoryID, OwnerID: &owner, ParentID: &categoryID, Kind: entity.TransactionTypeExpense, Name: "Coffee"}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&entity.Category{ID: otherCategoryID, OwnerID: &other, Kind: entity.TransactionTypeExpense, Name: "Private"}).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Where("owner_id IN ?", []string{owner, other}).Delete(&entity.Transaction{})
		db.Where("id IN ?", []string{walletID}).Delete(&entity.Wallet{})
		db.Where("id IN ?", []string{categoryID, childCategoryID, otherCategoryID}).Delete(&entity.Category{})
		db.Where("id IN ?", []string{owner, other}).Delete(&entity.User{})
	})
	transactionID := uuid.NewString()
	for _, tx := range []entity.Transaction{
		{ID: transactionID, OwnerID: owner, WalletID: walletID, CategoryID: &categoryID, Type: entity.TransactionTypeIncome, Amount: 1000, OccurredAt: now, IncludedInReports: true},
		{ID: uuid.NewString(), OwnerID: owner, WalletID: walletID, CategoryID: &categoryID, Type: entity.TransactionTypeExpense, Amount: 100, OccurredAt: now.Add(time.Hour), IncludedInReports: true},
		{ID: uuid.NewString(), OwnerID: owner, WalletID: walletID, CategoryID: &categoryID, Type: entity.TransactionTypeExpense, Amount: 50, OccurredAt: now.Add(2 * time.Hour), IncludedInReports: true},
		{ID: uuid.NewString(), OwnerID: owner, WalletID: walletID, CategoryID: &categoryID, Type: entity.TransactionTypeExpense, Amount: 20, OccurredAt: now.Add(3 * time.Hour), IncludedInReports: false},
		{ID: uuid.NewString(), OwnerID: other, WalletID: walletID, CategoryID: &otherCategoryID, Type: entity.TransactionTypeExpense, Amount: 999, OccurredAt: now, IncludedInReports: true},
	} {
		excluded := tx.Amount == 20
		if err := db.Create(&tx).Error; err != nil {
			t.Fatal(err)
		}
		if excluded {
			if err := db.Exec("UPDATE transactions SET included_in_reports = false WHERE id = ?", tx.ID).Error; err != nil {
				t.Fatal(err)
			}
		}
	}
	var includedCount, excludedCount int64
	if err := db.Model(&entity.Transaction{}).Where("owner_id = ? AND included_in_reports = true", owner).Count(&includedCount).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&entity.Transaction{}).Where("owner_id = ? AND included_in_reports = false", owner).Count(&excludedCount).Error; err != nil {
		t.Fatal(err)
	}
	if includedCount != 3 || excludedCount != 1 {
		t.Fatalf("fixture report flags incorrect: included=%d excluded=%d", includedCount, excludedCount)
	}

	reader := NewFinanceQueryPostgresRepository(db)
	filter, err := entity.NormalizeFinanceFilter(entity.FinanceFilter{Range: entity.DateRange{From: "2026-09-01", To: "2026-09-30"}}, time.FixedZone("ICT", 7*60*60))
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := reader.ReadBundle(context.Background(), owner, []entity.NormalizedQuery{{Key: "q1", Kind: "get_finance_summary", Filter: filter}})
	if err != nil {
		t.Fatal(err)
	}
	if len(bundle.Results) != 1 || bundle.Results[0].Status != "ok" {
		t.Fatalf("unexpected bundle: %+v", bundle)
	}
	var summary struct {
		Income  int64 `json:"income"`
		Expense int64 `json:"expense"`
		Net     int64 `json:"net"`
	}
	if err := json.Unmarshal(bundle.Results[0].View, &summary); err != nil {
		t.Fatal(err)
	}
	if summary.Income != 1000 || summary.Expense != 150 || summary.Net != 850 {
		t.Fatalf("incorrect report: %+v", summary)
	}
	detailBundle, err := reader.ReadBundle(context.Background(), owner, []entity.NormalizedQuery{{Key: "q-detail", Kind: "get_transaction", ObjectIDs: []string{transactionID}}})
	if err != nil {
		t.Fatal(err)
	}
	if detailBundle.Results[0].Status != "ok" || detailBundle.Results[0].ViewKind != "transaction_detail" {
		t.Fatalf("transaction detail should be readable regardless of report scope: %+v", detailBundle.Results[0])
	}
	var detail financeTransactionDetailView
	if err := json.Unmarshal(detailBundle.Results[0].View, &detail); err != nil {
		t.Fatal(err)
	}
	if detail.ID != transactionID || detail.Amount != 1000 || detail.WalletName != "Cash" || detail.CategoryName != "Food" {
		t.Fatalf("unexpected transaction detail: %+v", detail)
	}
	otherDetail, err := reader.ReadBundle(context.Background(), other, []entity.NormalizedQuery{{Key: "q-detail-other", Kind: "get_transaction", ObjectIDs: []string{transactionID}}})
	if err != nil {
		t.Fatal(err)
	}
	if otherDetail.Results[0].Status != "not_found" {
		t.Fatalf("transaction detail must not cross owner boundary: %+v", otherDetail.Results[0])
	}
	for i := 0; i < 51; i++ {
		row := entity.Transaction{ID: uuid.NewString(), OwnerID: owner, WalletID: walletID, Type: entity.TransactionTypeExpense, Amount: int64(i + 1), OccurredAt: now.Add(time.Duration(i+4) * time.Hour), IncludedInReports: true}
		if err := db.Create(&row).Error; err != nil {
			t.Fatal(err)
		}
	}
	searchFilter := filter
	searchFilter.Limit = 50
	searchQuery := entity.NormalizedQuery{Key: "q2", Kind: "search_transactions", Filter: searchFilter}
	searchBundle, err := reader.ReadBundle(context.Background(), owner, []entity.NormalizedQuery{searchQuery})
	if err != nil {
		t.Fatal(err)
	}
	var page financeSearchView
	if err := json.Unmarshal(searchBundle.Results[0].View, &page); err != nil {
		t.Fatal(err)
	}
	if page.TotalCount != 55 || len(page.Items) != 50 || page.NextCursor == "" {
		t.Fatalf("aggregate must not use page count: %+v", page)
	}
	searchQuery.Cursor = page.NextCursor
	secondPage, err := reader.ReadBundle(context.Background(), owner, []entity.NormalizedQuery{searchQuery})
	if err != nil {
		t.Fatal(err)
	}
	var tail financeSearchView
	if err := json.Unmarshal(secondPage.Results[0].View, &tail); err != nil {
		t.Fatal(err)
	}
	if len(tail.Items) != 5 || tail.TotalCount != 55 {
		t.Fatalf("cursor page mismatch: %+v", tail)
	}
	searchQuery.Cursor = page.NextCursor
	searchQuery.Filter.WalletIDs = []string{"different-scope"}
	if _, err := reader.ReadBundle(context.Background(), owner, []entity.NormalizedQuery{searchQuery}); !errors.Is(err, ErrFinanceCursor) {
		t.Fatalf("cursor scope mismatch should fail, got %v", err)
	}
	childRow := entity.Transaction{ID: uuid.NewString(), OwnerID: owner, WalletID: walletID, CategoryID: &childCategoryID, Type: entity.TransactionTypeExpense, Amount: 30, OccurredAt: now.Add(100 * time.Hour), IncludedInReports: true}
	if err := db.Create(&childRow).Error; err != nil {
		t.Fatal(err)
	}
	categoryFilter := filter
	categoryFilter.CategoryIDs = []string{categoryID}
	categoryFilter.IncludeChildren = true
	categoryBundle, err := reader.ReadBundle(context.Background(), owner, []entity.NormalizedQuery{{Key: "q3", Kind: "get_finance_summary", Filter: categoryFilter}})
	if err != nil {
		t.Fatal(err)
	}
	var categorySummary financeSummaryView
	if err := json.Unmarshal(categoryBundle.Results[0].View, &categorySummary); err != nil {
		t.Fatal(err)
	}
	if categorySummary.Expense != 180 {
		t.Fatalf("category subtree duplicated/missing: %+v", categorySummary)
	}
	compareBundle, err := reader.ReadBundle(context.Background(), owner, []entity.NormalizedQuery{{Key: "q-compare", Kind: "compare_spending_periods", Filter: filter, CompareRange: &entity.DateRange{From: "2026-08-01", To: "2026-08-31"}}})
	if err != nil {
		t.Fatal(err)
	}
	var comparison financeComparisonView
	if err := json.Unmarshal(compareBundle.Results[0].View, &comparison); err != nil {
		t.Fatal(err)
	}
	if comparison.CurrentExpense <= 0 || comparison.PreviousExpense != 0 || comparison.ChangeBPS != nil {
		t.Fatalf("zero-baseline comparison must be explicit: %+v", comparison)
	}
	var monthsBefore, configsBefore int64
	if err := db.Model(&entity.JarMonth{}).Where("owner_id = ?", owner).Count(&monthsBefore).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&entity.JarMonthConfig{}).Where("owner_id = ?", owner).Count(&configsBefore).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := reader.ReadBundle(context.Background(), owner, []entity.NormalizedQuery{{Key: "q4", Kind: "get_jar_progress", Month: "2026-09"}}); err != nil {
		t.Fatal(err)
	}
	var monthsAfter, configsAfter int64
	if err := db.Model(&entity.JarMonth{}).Where("owner_id = ?", owner).Count(&monthsAfter).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&entity.JarMonthConfig{}).Where("owner_id = ?", owner).Count(&configsAfter).Error; err != nil {
		t.Fatal(err)
	}
	if monthsBefore != monthsAfter || configsBefore != configsAfter {
		t.Fatalf("read-only jar query mutated config rows: before=%d/%d after=%d/%d", monthsBefore, configsBefore, monthsAfter, configsAfter)
	}
}
