package repository

import (
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// The owner's example: a 15,000,000 monthly jar with 10,000,000 spent in October shows
// 20,000,000 still plannable in November, and the plan is reconciled against real wallets.
func TestJarMonthShowsCumulativeRemainingPlanAndWalletReconciliation(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires migrated local TEST_DATABASE_URL")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	owner := uuid.NewString()
	if err = db.Create(&entity.User{ID: owner, Email: owner + "@test.invalid", GoogleSubject: owner, Timezone: "Asia/Ho_Chi_Minh", TimezoneConfirmed: true}).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Where("owner_id = ?", owner).Delete(&entity.Transaction{})
		db.Where("owner_id = ?", owner).Delete(&entity.JarMonthConfig{})
		db.Where("owner_id = ?", owner).Delete(&entity.JarMonth{})
		db.Where("owner_id = ?", owner).Delete(&entity.Jar{})
		db.Where("owner_id = ?", owner).Delete(&entity.Wallet{})
		db.Where("owner_id = ?", owner).Delete(&entity.Category{})
		db.Where("id = ?", owner).Delete(&entity.User{})
	})
	location, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
	wallet := entity.Wallet{ID: uuid.NewString(), OwnerID: owner, Name: "Main", Type: entity.WalletTypeBasic, Currency: "VND", OpeningBalance: 40_000_000, IsInTotal: true}
	category := entity.Category{ID: uuid.NewString(), OwnerID: &owner, Kind: entity.TransactionTypeExpense, Name: "Du lịch", IconKey: "tag"}
	jar := entity.Jar{ID: uuid.NewString(), OwnerID: owner}
	for _, row := range []any{&wallet, &category, &jar} {
		if err = db.Create(row).Error; err != nil {
			t.Fatal(err)
		}
	}
	for _, label := range []string{"2026-10-01", "2026-11-01"} {
		month, _ := entity.ParseCalendarDate(label)
		amount := int64(15_000_000)
		if err = db.Create(&entity.JarMonth{OwnerID: owner, Month: month}).Error; err != nil {
			t.Fatal(err)
		}
		if err = db.Create(&entity.JarMonthConfig{OwnerID: owner, Month: month, JarID: jar.ID, Name: "Du lịch", AllocationMode: entity.JarAllocationFixed, AllocationAmount: &amount, Active: true}).Error; err != nil {
			t.Fatal(err)
		}
	}
	expense := func(amount int64, month time.Month, day int, jarID *string) {
		row := entity.Transaction{ID: uuid.NewString(), OwnerID: owner, WalletID: wallet.ID, CategoryID: &category.ID, JarID: jarID, Type: entity.TransactionTypeExpense, Amount: amount, OccurredAt: time.Date(2026, month, day, 12, 0, 0, 0, location), IncludedInReports: true}
		if err := db.Create(&row).Error; err != nil {
			t.Fatal(err)
		}
	}
	expense(10_000_000, time.October, 5, &jar.ID)
	expense(400_000, time.November, 2, nil)
	repo := NewJarPostgresRepository(db)

	summary, err := repo.ListMonth(owner, "2026-11", "Asia/Ho_Chi_Minh")
	if err != nil {
		t.Fatal(err)
	}
	if len(summary.Items) != 1 {
		t.Fatalf("items = %+v", summary.Items)
	}
	item := summary.Items[0]
	if item.RemainingPlan == nil || *item.RemainingPlan != 20_000_000 || item.CumulativeAllocated != 30_000_000 || item.CumulativeSpent != 10_000_000 ||
		item.AllocationMonths != 2 || item.MonthsInRange != 2 || item.FirstConfigMonth != "2026-10" || item.MonthVariance == nil || *item.MonthVariance != 15_000_000 {
		t.Fatalf("owner example: want 20,000,000 remaining plan from 30,000,000 planned and 10,000,000 spent, got %+v", item)
	}
	if summary.RemainingPlanTotal != 20_000_000 || summary.UnassignedCumulative != 400_000 || summary.WalletsInTotalBalance != 40_000_000-10_000_000-400_000 || summary.PlanExceedsWalletsBy != 0 {
		t.Fatalf("summary totals = %+v", summary)
	}

	// Overspending: November spend 17,000,000 leaves 3,000,000 of plan; the month itself is over by 2,000,000.
	expense(17_000_000, time.November, 9, &jar.ID)
	summary, err = repo.ListMonth(owner, "2026-11", "Asia/Ho_Chi_Minh")
	if err != nil {
		t.Fatal(err)
	}
	item = summary.Items[0]
	if item.RemainingPlan == nil || *item.RemainingPlan != 3_000_000 || item.MonthVariance == nil || *item.MonthVariance != -2_000_000 {
		t.Fatalf("after overspending: %+v", item)
	}

	// Plan beyond real money is reported, never blocked.
	if err = db.Model(&entity.Wallet{}).Where("id = ?", wallet.ID).Update("opening_balance", 30_000_000).Error; err != nil {
		t.Fatal(err)
	}
	summary, err = repo.ListMonth(owner, "2026-11", "Asia/Ho_Chi_Minh")
	if err != nil {
		t.Fatal(err)
	}
	balance := int64(30_000_000 - 10_000_000 - 400_000 - 17_000_000)
	if summary.WalletsInTotalBalance != balance || summary.PlanExceedsWalletsBy != 3_000_000-balance {
		t.Fatalf("reconciliation = balance %d exceeds by %d", summary.WalletsInTotalBalance, summary.PlanExceedsWalletsBy)
	}

	// Viewing October only sees October: remaining plan 5,000,000 and no November spend.
	october, err := repo.ListMonth(owner, "2026-10", "Asia/Ho_Chi_Minh")
	if err != nil {
		t.Fatal(err)
	}
	if got := october.Items[0]; got.RemainingPlan == nil || *got.RemainingPlan != 5_000_000 || got.MonthsInRange != 1 {
		t.Fatalf("October view = %+v", got)
	}
}
