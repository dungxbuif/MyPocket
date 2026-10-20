package repository

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
	"testing"
)

func TestSharedCategoryAssignmentsStayWithinEachOwner(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires migrated TEST_DATABASE_URL")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	a, b := uuid.NewString(), uuid.NewString()
	wa := entity.Wallet{ID: uuid.NewString(), OwnerID: a, Name: "a", Type: "basic", Currency: "VND"}
	wb := entity.Wallet{ID: uuid.NewString(), OwnerID: b, Name: "b", Type: "basic", Currency: "VND"}
	c := entity.Category{ID: uuid.NewString(), Name: "shared test", Kind: "expense", IsSystem: true, IconKey: "tag"}
	t.Cleanup(func() {
		db.Where("id IN ?", []string{wa.ID, wb.ID}).Delete(&entity.Wallet{})
		db.Where("id = ?", c.ID).Delete(&entity.Category{})
	})
	for _, item := range []any{&wa, &wb, &c, &entity.CategoryWallet{CategoryID: c.ID, WalletID: wa.ID}, &entity.CategoryWallet{CategoryID: c.ID, WalletID: wb.ID}} {
		if err = db.Create(item).Error; err != nil {
			t.Fatal(err)
		}
	}
	r := NewCategoryPostgresRepository(db)
	items, err := r.ListVisible(a)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range items {
		if item.ID == c.ID && (len(item.WalletIDs) != 1 || item.WalletIDs[0] != wa.ID) {
			t.Fatalf("other owner's wallet exposed: %v", item.WalletIDs)
		}
	}
	if err = r.ReplaceWallets(a, &c, nil); err != nil {
		t.Fatal(err)
	}
	var assignments []entity.CategoryWallet
	db.Where("category_id = ?", c.ID).Find(&assignments)
	if len(assignments) != 1 || assignments[0].WalletID != wb.ID {
		t.Fatalf("other owner's choice removed: %+v", assignments)
	}
}

func TestCreateCategoryPersistsOwnerWalletScope(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires migrated TEST_DATABASE_URL")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	ownerID := uuid.NewString()
	wallet := entity.Wallet{ID: uuid.NewString(), OwnerID: ownerID, Name: "category scope test", Type: "basic", Currency: "VND"}
	category := entity.Category{
		ID:        uuid.NewString(),
		Name:      "category scope test",
		Kind:      "expense",
		IconKey:   "tag",
		WalletIDs: []string{wallet.ID},
	}
	t.Cleanup(func() {
		db.Where("category_id = ?", category.ID).Delete(&entity.CategoryWallet{})
		db.Where("id = ?", category.ID).Delete(&entity.Category{})
		db.Where("id = ?", wallet.ID).Delete(&entity.Wallet{})
	})
	if err := db.Create(&wallet).Error; err != nil {
		t.Fatal(err)
	}

	repository := NewCategoryPostgresRepository(db)
	if err := repository.Create(ownerID, &category); err != nil {
		t.Fatal(err)
	}

	var assignments []entity.CategoryWallet
	if err := db.Where("category_id = ?", category.ID).Find(&assignments).Error; err != nil {
		t.Fatal(err)
	}
	if len(assignments) != 1 || assignments[0].WalletID != wallet.ID {
		t.Fatalf("category creation lost wallet scope: got %+v, want wallet %s", assignments, wallet.ID)
	}
}

func TestCategoryArchiveMigrationDefaultsActive(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires migrated TEST_DATABASE_URL")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	categoryID := uuid.NewString()
	t.Cleanup(func() { db.Exec("DELETE FROM categories WHERE id = ?", categoryID) })

	var active bool
	if err := db.Raw(
		"INSERT INTO categories (id, kind, name, is_system) VALUES (?, 'expense', ?, false) RETURNING is_active",
		categoryID,
		"archive default "+categoryID,
	).Scan(&active).Error; err != nil {
		t.Fatal(err)
	}
	if !active {
		t.Fatal("category inserted without is_active did not inherit the migration's active default")
	}
}

func TestCategoryArchiveIsOwnerScopedIdempotentAndPreservesHistory(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires migrated TEST_DATABASE_URL")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	ownerID, otherOwnerID := uuid.NewString(), uuid.NewString()
	wallet := entity.Wallet{ID: uuid.NewString(), OwnerID: ownerID, Name: "archive history test", Type: "basic", Currency: "VND"}
	category := entity.Category{ID: uuid.NewString(), OwnerID: &ownerID, Kind: "expense", Name: "archive history test", IconKey: "tag", IsActive: true}
	systemCategory := entity.Category{ID: uuid.NewString(), Kind: "expense", Name: "archive system test", IconKey: "tag", IsSystem: true}
	transaction := entity.Transaction{ID: uuid.NewString(), OwnerID: ownerID, WalletID: wallet.ID, CategoryID: &category.ID, Type: entity.TransactionTypeExpense, Amount: 100, OccurredAt: time.Now()}
	t.Cleanup(func() {
		db.Where("id = ?", transaction.ID).Delete(&entity.Transaction{})
		db.Where("id IN ?", []string{category.ID, systemCategory.ID}).Delete(&entity.Category{})
		db.Where("id = ?", wallet.ID).Delete(&entity.Wallet{})
	})
	for _, item := range []any{&wallet, &category, &systemCategory, &transaction} {
		if err := db.Create(item).Error; err != nil {
			t.Fatal(err)
		}
	}

	repository := NewCategoryPostgresRepository(db)
	archived, err := repository.SetActive(ownerID, category.ID, false)
	if err != nil || archived.IsActive || archived.Version != 2 {
		t.Fatalf("archive result = %+v, err = %v; want inactive category at version 2", archived, err)
	}
	if _, err := repository.SetActive(ownerID, category.ID, false); err != nil {
		t.Fatalf("repeating archive must be idempotent: %v", err)
	}
	if _, err := repository.SetActive(otherOwnerID, category.ID, true); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("other owner status update error = %v, want not found", err)
	}
	if _, err := repository.SetActive(ownerID, systemCategory.ID, false); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("system category status update error = %v, want not found", err)
	}
	var historyCount int64
	if err := db.Model(&entity.Transaction{}).Where("id = ? AND category_id = ?", transaction.ID, category.ID).Count(&historyCount).Error; err != nil {
		t.Fatal(err)
	}
	if historyCount != 1 {
		t.Fatalf("archiving category changed historical transaction count to %d", historyCount)
	}
	restored, err := repository.SetActive(ownerID, category.ID, true)
	if err != nil || !restored.IsActive || restored.Version != 3 {
		t.Fatalf("restore result = %+v, err = %v; want active category at version 3", restored, err)
	}
}
