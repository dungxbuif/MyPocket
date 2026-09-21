package repository

import (
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
