package repository

import (
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestJarSpendRemainsVisibleWhenTimezoneRebucketsLinkedTransaction(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires migrated local TEST_DATABASE_URL")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	owner, walletID, jarID, transactionID := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	user := entity.User{ID: owner, Email: owner + "@test.invalid", GoogleSubject: owner, Timezone: "Asia/Ho_Chi_Minh", TimezoneConfirmed: true}
	if err = db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Where("owner_id = ?", owner).Delete(&entity.Transaction{})
		db.Where("owner_id = ?", owner).Delete(&entity.Wallet{})
		db.Where("id = ?", owner).Delete(&entity.User{})
	})
	wallet := entity.Wallet{ID: walletID, OwnerID: owner, Name: "timezone test", Type: entity.WalletTypeBasic, Currency: "VND"}
	if err = db.Create(&wallet).Error; err != nil {
		t.Fatal(err)
	}
	jar := entity.Jar{ID: jarID, OwnerID: owner}
	if err = db.Create(&jar).Error; err != nil {
		t.Fatal(err)
	}
	month, err := entity.ParseCalendarDate("2026-02-01")
	if err != nil {
		t.Fatal(err)
	}
	if err = db.Create(&entity.JarMonth{OwnerID: owner, Month: month}).Error; err != nil {
		t.Fatal(err)
	}
	if err = db.Create(&entity.JarMonthConfig{OwnerID: owner, Month: month, JarID: jarID, Name: "Ăn uống", AllocationMode: entity.JarAllocationNone, Active: true}).Error; err != nil {
		t.Fatal(err)
	}
	instant := time.Date(2026, time.February, 1, 0, 30, 0, 0, time.FixedZone("ICT", 7*60*60))
	transaction := entity.Transaction{ID: transactionID, OwnerID: owner, WalletID: walletID, JarID: &jarID, Type: entity.TransactionTypeExpense, Amount: 1200, OccurredAt: instant, IncludedInReports: true}
	if err = db.Create(&transaction).Error; err != nil {
		t.Fatal(err)
	}
	if err = db.Model(&entity.User{}).Where("id = ?", owner).Update("timezone", "America/Los_Angeles").Error; err != nil {
		t.Fatal(err)
	}
	repo := NewTransactionPostgresRepository(db)
	if _, err = repo.Update(owner, transactionID, map[string]any{"note": "metadata edit", "jar_id": jarID, "occurred_at": instant}); err != nil {
		t.Fatalf("metadata edit must preserve the existing stable jar link after timezone change: %v", err)
	}
	jarRepo := NewJarPostgresRepository(db)
	summary, err := jarRepo.ListMonth(owner, "2026-01", "America/Los_Angeles")
	if err != nil {
		t.Fatal(err)
	}
	if summary.TotalSpent != 1200 || summary.UnassignedSpent != 0 || len(summary.Items) != 1 {
		t.Fatalf("re-bucketed linked spend must remain attributed: %+v", summary)
	}
	item := summary.Items[0]
	if item.JarID != jarID || item.Name != "Ăn uống" || item.Spent != 1200 || item.Active {
		t.Fatalf("missing month config should show a readable, inactive historical jar item: %+v", item)
	}
	if len(summary.Jars) != 1 || summary.Jars[0].JarID != jarID || summary.Jars[0].Name != "Ăn uống" {
		t.Fatalf("month API must expose stable jar identities for cumulative navigation: %+v", summary.Jars)
	}
	cumulative, err := jarRepo.Cumulative(owner, jarID, "", "2026-01", "America/Los_Angeles")
	if err != nil {
		t.Fatalf("omitting the start month should include the entire jar history through the selected month: %v", err)
	}
	if cumulative.FromMonth != "2026-01" || cumulative.ToMonth != "2026-01" || cumulative.TotalSpent != 1200 || len(cumulative.Months) != 1 {
		t.Fatalf("all-history cumulative summary omitted rebucketed spend: %+v", cumulative)
	}
}

func TestEmptyJarMonthReturnsEmptyCollections(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires migrated local TEST_DATABASE_URL")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	owner := uuid.NewString()
	user := entity.User{ID: owner, Email: owner + "@test.invalid", GoogleSubject: owner, Timezone: "Asia/Ho_Chi_Minh"}
	if err = db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Where("id = ?", owner).Delete(&entity.User{}) })

	summary, err := NewJarPostgresRepository(db).ListMonth(owner, "2026-09", user.Timezone)
	if err != nil {
		t.Fatal(err)
	}
	if summary.Items == nil || summary.Jars == nil {
		t.Fatalf("empty jar API collections must serialize as arrays, got items=%#v jars=%#v", summary.Items, summary.Jars)
	}
}

func TestJarMonthInitializationCopiesActiveConfigOnceAndStaysOwnerScoped(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires migrated local TEST_DATABASE_URL")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	owner, otherOwner := uuid.NewString(), uuid.NewString()
	users := []entity.User{
		{ID: owner, Email: owner + "@test.invalid", GoogleSubject: owner, Timezone: "Asia/Ho_Chi_Minh"},
		{ID: otherOwner, Email: otherOwner + "@test.invalid", GoogleSubject: otherOwner, Timezone: "Asia/Ho_Chi_Minh"},
	}
	for _, user := range users {
		if err = db.Create(&user).Error; err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		for _, id := range []string{owner, otherOwner} {
			db.Where("owner_id = ?", id).Delete(&entity.JarMonthConfig{})
			db.Where("owner_id = ?", id).Delete(&entity.JarMonth{})
			db.Where("owner_id = ?", id).Delete(&entity.Jar{})
			db.Where("id = ?", id).Delete(&entity.User{})
		}
	})

	priorMonth, err := entity.ParseCalendarDate("2026-01-01")
	if err != nil {
		t.Fatal(err)
	}
	if err = db.Create(&entity.JarMonth{OwnerID: owner, Month: priorMonth}).Error; err != nil {
		t.Fatal(err)
	}
	activeJarID, inactiveJarID := uuid.NewString(), uuid.NewString()
	for _, jarID := range []string{activeJarID, inactiveJarID} {
		if err = db.Create(&entity.Jar{ID: jarID, OwnerID: owner}).Error; err != nil {
			t.Fatal(err)
		}
	}
	configs := []entity.JarMonthConfig{
		{OwnerID: owner, Month: priorMonth, JarID: activeJarID, Name: "Ăn uống", AllocationMode: entity.JarAllocationFixed, AllocationAmount: ptrInt64(100000), Active: true},
		{OwnerID: owner, Month: priorMonth, JarID: inactiveJarID, Name: "Đã gỡ", AllocationMode: entity.JarAllocationNone, Active: false},
	}
	if err = db.Create(&configs).Error; err != nil {
		t.Fatal(err)
	}
	if err = db.Model(&entity.JarMonthConfig{}).Where("owner_id = ? AND month = ? AND jar_id = ?", owner, priorMonth, inactiveJarID).Update("active", false).Error; err != nil {
		t.Fatal(err)
	}

	repo := NewJarPostgresRepository(db)
	var wait sync.WaitGroup
	errors := make(chan error, 2)
	for range 2 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			_, listErr := repo.ListMonth(owner, "2026-02", "Asia/Ho_Chi_Minh")
			errors <- listErr
		}()
	}
	wait.Wait()
	close(errors)
	for listErr := range errors {
		if listErr != nil {
			t.Fatal(listErr)
		}
	}

	var copied []entity.JarMonthConfig
	if err = db.Where("owner_id = ? AND month = ?", owner, "2026-02-01").Find(&copied).Error; err != nil {
		t.Fatal(err)
	}
	if len(copied) != 1 || copied[0].JarID != activeJarID || copied[0].Name != "Ăn uống" || !copied[0].Active {
		t.Fatalf("month copy must be idempotent and copy active configs only: %+v", copied)
	}
	var otherOwnerConfigCount int64
	if err = db.Model(&entity.JarMonthConfig{}).Where("owner_id = ?", otherOwner).Count(&otherOwnerConfigCount).Error; err != nil {
		t.Fatal(err)
	}
	if otherOwnerConfigCount != 0 {
		t.Fatalf("month initialization leaked configuration across owners: %d", otherOwnerConfigCount)
	}
}

func ptrInt64(value int64) *int64 { return &value }
