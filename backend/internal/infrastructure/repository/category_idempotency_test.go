package repository

import (
	"errors"
	"os"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	categoryrepo "github.com/mypocket/backend/internal/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type categoryCreateMutationRepository interface {
	CreateIdempotent(string, *entity.Category, string, string) (*entity.Category, bool, error)
}

func categoryMutationDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires migrated TEST_DATABASE_URL")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := any(NewCategoryPostgresRepository(db)).(categoryCreateMutationRepository); !ok {
		t.Fatal("category repository lacks durable create/replay capability")
	}
	return db
}

func categoryMutationOwner(t *testing.T, db *gorm.DB) string {
	t.Helper()
	owner := uuid.NewString()
	if err := db.Create(&entity.User{ID: owner, Email: owner + "@category-mutation.test", GoogleSubject: owner, Timezone: "Asia/Ho_Chi_Minh"}).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Exec("DELETE FROM category_wallets WHERE category_id IN (SELECT id FROM categories WHERE owner_id = ?)", owner)
		db.Where("owner_id = ?", owner).Delete(&entity.Category{})
		db.Where("owner_id = ?", owner).Delete(&entity.TransactionMutation{})
		db.Where("id = ?", owner).Delete(&entity.User{})
	})
	return owner
}

func categoryMutationCandidate(owner, name string) *entity.Category {
	return &entity.Category{ID: uuid.NewString(), OwnerID: &owner, Kind: entity.TransactionTypeExpense, Name: name, IconKey: "tag", IsActive: true}
}

func TestCategoryCreateMutationReplaysConflictsAndScopesOwners(t *testing.T) {
	db := categoryMutationDB(t)
	owner := categoryMutationOwner(t, db)
	otherOwner := categoryMutationOwner(t, db)
	repo := any(NewCategoryPostgresRepository(db)).(categoryCreateMutationRepository)
	key := uuid.NewString()

	first, replay, err := repo.CreateIdempotent(owner, categoryMutationCandidate(owner, "Coffee"), key, "category-hash")
	if err != nil || replay || first == nil {
		t.Fatalf("first create: replay=%v row=%+v err=%v", replay, first, err)
	}
	second, replay, err := repo.CreateIdempotent(owner, categoryMutationCandidate(owner, "Different name"), key, "category-hash")
	if err != nil || !replay || second == nil || second.ID != first.ID || second.Name != first.Name {
		t.Fatalf("replay changed persisted result: replay=%v row=%+v first=%+v err=%v", replay, second, first, err)
	}
	if _, _, err = repo.CreateIdempotent(owner, categoryMutationCandidate(owner, "Changed"), key, "changed-hash"); !errors.Is(err, categoryrepo.ErrMutationConflict) {
		t.Fatalf("changed payload must conflict: %v", err)
	}
	other, replay, err := repo.CreateIdempotent(otherOwner, categoryMutationCandidate(otherOwner, "Coffee"), key, "category-hash")
	if err != nil || replay || other == nil || other.ID == first.ID || other.OwnerID == nil || *other.OwnerID != otherOwner {
		t.Fatalf("owner isolation failed: replay=%v row=%+v err=%v", replay, other, err)
	}

	var categoryCount, mutationCount int64
	if err := db.Model(&entity.Category{}).Where("owner_id = ? AND id = ?", owner, first.ID).Count(&categoryCount).Error; err != nil || categoryCount != 1 {
		t.Fatalf("category rows=%d err=%v", categoryCount, err)
	}
	if err := db.Model(&entity.TransactionMutation{}).Where("owner_id = ? AND operation = ? AND idempotency_key = ?", owner, "category.create", key).Count(&mutationCount).Error; err != nil || mutationCount != 1 {
		t.Fatalf("owner mutation rows=%d err=%v", mutationCount, err)
	}
}

func TestCategoryCreateMutationConcurrentConfirmationWritesOnce(t *testing.T) {
	db := categoryMutationDB(t)
	owner := categoryMutationOwner(t, db)
	repo := any(NewCategoryPostgresRepository(db)).(categoryCreateMutationRepository)
	key := uuid.NewString()

	type outcome struct {
		row    *entity.Category
		replay bool
		err    error
	}
	results := make(chan outcome, 8)
	var group sync.WaitGroup
	for i := 0; i < 8; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			row, replay, err := repo.CreateIdempotent(owner, categoryMutationCandidate(owner, "Concurrent"), key, "concurrent-hash")
			results <- outcome{row: row, replay: replay, err: err}
		}()
	}
	group.Wait()
	close(results)

	firstID := ""
	applied := 0
	for result := range results {
		if result.err != nil || result.row == nil {
			t.Fatalf("concurrent confirmation failed: row=%+v err=%v", result.row, result.err)
		}
		if firstID == "" {
			firstID = result.row.ID
		}
		if result.row.ID != firstID {
			t.Fatalf("concurrent retry returned a different category: %s vs %s", result.row.ID, firstID)
		}
		if !result.replay {
			applied++
		}
	}
	if applied != 1 {
		t.Fatalf("want exactly one applied confirmation, got %d", applied)
	}
	var categoryCount, mutationCount int64
	if err := db.Model(&entity.Category{}).Where("owner_id = ? AND name = ?", owner, "Concurrent").Count(&categoryCount).Error; err != nil || categoryCount != 1 {
		t.Fatalf("concurrent category rows=%d err=%v", categoryCount, err)
	}
	if err := db.Model(&entity.TransactionMutation{}).Where("owner_id = ? AND operation = ? AND idempotency_key = ?", owner, "category.create", key).Count(&mutationCount).Error; err != nil || mutationCount != 1 {
		t.Fatalf("concurrent mutation rows=%d err=%v", mutationCount, err)
	}
}

func TestCategoryCreateMutationRollsBackCategoryLinksAndReceipt(t *testing.T) {
	db := categoryMutationDB(t)
	owner := categoryMutationOwner(t, db)
	wallet := entity.Wallet{ID: uuid.NewString(), OwnerID: owner, Name: "Synthetic category wallet", Type: entity.WalletTypeBasic, Currency: "VND"}
	if err := db.Create(&wallet).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Where("id = ?", wallet.ID).Delete(&entity.Wallet{}) })
	const callback = "test:reject_category_wallet_insert"
	if err := db.Callback().Create().Before("gorm:create").Register(callback, func(tx *gorm.DB) {
		if tx.Statement.Table == "category_wallets" {
			tx.AddError(errors.New("synthetic category link failure"))
		}
	}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Callback().Create().Remove(callback) })
	candidate := categoryMutationCandidate(owner, "Rollback")
	candidate.WalletIDs = []string{wallet.ID}
	key := uuid.NewString()
	repo := any(NewCategoryPostgresRepository(db)).(categoryCreateMutationRepository)
	if _, _, err := repo.CreateIdempotent(owner, candidate, key, "rollback-hash"); err == nil {
		t.Fatal("link failure did not abort create")
	}
	for _, query := range []struct {
		model any
		where string
		value string
	}{
		{&entity.Category{}, "id = ?", candidate.ID},
		{&entity.CategoryWallet{}, "category_id = ?", candidate.ID},
		{&entity.TransactionMutation{}, "idempotency_key = ?", key},
	} {
		var count int64
		if err := db.Model(query.model).Where(query.where, query.value).Count(&count).Error; err != nil || count != 0 {
			t.Fatalf("partial category mutation survived: model=%T count=%d err=%v", query.model, count, err)
		}
	}
}
