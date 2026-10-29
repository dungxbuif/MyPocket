package repository

import (
	"errors"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	transactionrepo "github.com/mypocket/backend/internal/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// The runtime capability assertion lets this test fail against the existing
// implementation before adding the private repository boundary.
type offlineCreateRepository interface {
	CreateIdempotent(*entity.Transaction, string, string) (*entity.Transaction, bool, error)
	CreateTransferIdempotent(string, *entity.Transaction, *entity.Transaction, string, string) ([]entity.Transaction, bool, error)
}

func offlineReceiptDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires migrated disposable TEST_DATABASE_URL")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := any(NewTransactionPostgresRepository(db)).(offlineCreateRepository); !ok {
		t.Fatal("transaction repository lacks durable create/replay capability")
	}
	return db
}

func offlineOwner(t *testing.T, db *gorm.DB) (string, string, string) {
	t.Helper()
	owner := uuid.NewString()
	if err := db.Create(&entity.User{ID: owner, Email: owner + "@offline.test", GoogleSubject: owner, Timezone: "Asia/Ho_Chi_Minh"}).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		for _, model := range []any{&entity.Transaction{}, &entity.TransactionMutation{}, &entity.Wallet{}} {
			if err := db.Where("owner_id = ?", owner).Delete(model).Error; err != nil {
				t.Error(err)
			}
		}
		if err := db.Where("id = ?", owner).Delete(&entity.User{}).Error; err != nil {
			t.Error(err)
		}
	})
	wallet := entity.Wallet{ID: uuid.NewString(), OwnerID: owner, Name: "Offline fixture", Type: entity.WalletTypeBasic, Currency: entity.WalletCurrencyVND}
	if err := db.Create(&wallet).Error; err != nil {
		t.Fatal(err)
	}
	var category entity.Category
	if err := db.Where("is_system = ? AND system_key = ?", true, "expense_general").First(&category).Error; err != nil {
		t.Fatal(err)
	}
	return owner, wallet.ID, category.ID
}

func offlineTransaction(owner, wallet, category string) *entity.Transaction {
	return &entity.Transaction{ID: uuid.NewString(), OwnerID: owner, WalletID: wallet, CategoryID: &category, Type: entity.TransactionTypeExpense, Amount: 12500, OccurredAt: time.Date(2026, 9, 1, 3, 0, 0, 0, time.UTC), IncludedInReports: false}
}

func TestOfflineCreateReceiptReplayConflictAndOwnerIsolation(t *testing.T) {
	db := offlineReceiptDB(t)
	owner, wallet, category := offlineOwner(t, db)
	repo := any(NewTransactionPostgresRepository(db)).(offlineCreateRepository)
	key := uuid.NewString()
	first, replay, err := repo.CreateIdempotent(offlineTransaction(owner, wallet, category), key, "same-request-hash")
	if err != nil || replay || first == nil {
		t.Fatalf("first create: replay=%v err=%v", replay, err)
	}
	second, replay, err := repo.CreateIdempotent(offlineTransaction(owner, wallet, category), key, "same-request-hash")
	if err != nil || !replay || second == nil || second.ID != first.ID || second.Amount != 12500 || second.IncludedInReports {
		t.Fatalf("replay lost original result: replay=%v row=%+v err=%v", replay, second, err)
	}
	if _, _, err = repo.CreateIdempotent(offlineTransaction(owner, wallet, category), key, "changed-request-hash"); !errors.Is(err, transactionrepo.ErrMutationConflict) {
		t.Fatalf("changed request must conflict: %v", err)
	}
	other, otherWallet, otherCategory := offlineOwner(t, db)
	row, replay, err := repo.CreateIdempotent(offlineTransaction(other, otherWallet, otherCategory), key, "same-request-hash")
	if err != nil || replay || row == nil || row.OwnerID != other || row.ID == first.ID {
		t.Fatalf("owner isolation: replay=%v row=%+v err=%v", replay, row, err)
	}
	var count int64
	if err := db.Model(&entity.Transaction{}).Where("owner_id = ?", owner).Count(&count).Error; err != nil || count != 1 {
		t.Fatalf("duplicate financial rows=%d err=%v", count, err)
	}
	var receipt entity.TransactionMutation
	if err := db.Where("owner_id = ? AND operation = ? AND idempotency_key = ?", owner, "transaction.create", key).First(&receipt).Error; err != nil {
		t.Fatal(err)
	}
	if receipt.Status != "completed" || receipt.CreatedAt.IsZero() || receipt.UpdatedAt.IsZero() || receipt.ResponseJSON == "{}" {
		t.Fatal("completed receipt lacks durable result or retention timestamps")
	}
}

func TestOfflineCreateReceiptConcurrentSingleWriteAndRollback(t *testing.T) {
	db := offlineReceiptDB(t)
	owner, wallet, category := offlineOwner(t, db)
	repo := any(NewTransactionPostgresRepository(db)).(offlineCreateRepository)
	key := uuid.NewString()
	type outcome struct {
		row    *entity.Transaction
		replay bool
		err    error
	}
	results := make(chan outcome, 8)
	var group sync.WaitGroup
	for i := 0; i < 8; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			row, replay, err := repo.CreateIdempotent(offlineTransaction(owner, wallet, category), key, "concurrent-hash")
			results <- outcome{row, replay, err}
		}()
	}
	group.Wait()
	close(results)
	firstID := ""
	applied := 0
	for result := range results {
		if result.err != nil || result.row == nil {
			t.Fatalf("concurrent request failed: %v", result.err)
		}
		if firstID == "" {
			firstID = result.row.ID
		}
		if result.row.ID != firstID {
			t.Fatal("concurrent retry created a different transaction")
		}
		if !result.replay {
			applied++
		}
	}
	if applied != 1 {
		t.Fatalf("want exactly one applied call, got%d", applied)
	}
	rollbackKey := uuid.NewString()
	rollback := errors.New("abort outer transaction")
	err := db.Transaction(func(tx *gorm.DB) error {
		transactional := any(NewTransactionPostgresRepository(tx)).(offlineCreateRepository)
		if _, _, err := transactional.CreateIdempotent(offlineTransaction(owner, wallet, category), rollbackKey, "rollback-hash"); err != nil {
			return err
		}
		return rollback
	})
	if !errors.Is(err, rollback) {
		t.Fatal(err)
	}
	var count int64
	if err := db.Model(&entity.TransactionMutation{}).Where("owner_id = ? AND idempotency_key = ?", owner, rollbackKey).Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("receipt escaped rollback: %d %v", count, err)
	}
	if err := db.Model(&entity.Transaction{}).Where("owner_id = ?", owner).Count(&count).Error; err != nil || count != 1 {
		t.Fatalf("financial row escaped rollback: %d %v", count, err)
	}
	bad := offlineTransaction(owner, wallet, category)
	bad.CategoryID = nil
	failureKey := uuid.NewString()
	if _, _, err := repo.CreateIdempotent(bad, failureKey, "invalid-hash"); err == nil {
		t.Fatal("invalid create unexpectedly succeeded")
	}
	if _, replay, err := repo.CreateIdempotent(offlineTransaction(owner, wallet, category), failureKey, "corrected-hash"); err != nil || replay {
		t.Fatalf("failed create left a reserved receipt: replay=%v err=%v", replay, err)
	}
}

func TestOfflineTransferCreateReceiptReplaysExactlyTwoExcludedRows(t *testing.T) {
	db := offlineReceiptDB(t)
	owner, sourceWallet, _ := offlineOwner(t, db)
	destination := entity.Wallet{ID: uuid.NewString(), OwnerID: owner, Name: "Destination", Type: entity.WalletTypeBasic, Currency: entity.WalletCurrencyVND}
	if err := db.Create(&destination).Error; err != nil {
		t.Fatal(err)
	}
	var sourceCategory, destinationCategory entity.Category
	if err := db.Where("system_key = ?", "expense_transfer_out").First(&sourceCategory).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Where("system_key = ?", "income_transfer_in").First(&destinationCategory).Error; err != nil {
		t.Fatal(err)
	}
	makePair := func() (*entity.Transaction, *entity.Transaction) {
		pairID := uuid.NewString()
		source := offlineTransaction(owner, sourceWallet, sourceCategory.ID)
		source.TransferID = &pairID
		target := offlineTransaction(owner, destination.ID, destinationCategory.ID)
		target.TransferID = &pairID
		target.Type = entity.TransactionTypeIncome
		return source, target
	}
	repo := any(NewTransactionPostgresRepository(db)).(offlineCreateRepository)
	source, target := makePair()
	key := uuid.NewString()
	first, replay, err := repo.CreateTransferIdempotent(owner, source, target, key, "pair-hash")
	if err != nil || replay || len(first) != 2 {
		t.Fatalf("create pair: replay=%v len=%d err=%v", replay, len(first), err)
	}
	source, target = makePair()
	second, replay, err := repo.CreateTransferIdempotent(owner, source, target, key, "pair-hash")
	if err != nil || !replay || len(second) != 2 {
		t.Fatalf("replay pair: replay=%v len=%d err=%v", replay, len(second), err)
	}
	for i := range second {
		if second[i].ID != first[i].ID || second[i].IncludedInReports || second[i].JarID != nil || second[i].TransferID == nil || *second[i].TransferID != *first[i].TransferID {
			t.Fatal("replayed pair lost original identity or exclusion")
		}
	}
	if _, _, err := repo.CreateTransferIdempotent(owner, source, target, key, "different-pair-hash"); !errors.Is(err, transactionrepo.ErrMutationConflict) {
		t.Fatalf("changed pair must conflict: %v", err)
	}
	var count int64
	if err := db.Model(&entity.Transaction{}).Where("owner_id = ?", owner).Count(&count).Error; err != nil || count != 2 {
		t.Fatalf("want2 rows, got%d err=%v", count, err)
	}
}

func TestOfflineReceiptRejectsInvalidKeysAndForeignWalletBeforeWriting(t *testing.T) {
	db := offlineReceiptDB(t)
	owner, wallet, category := offlineOwner(t, db)
	repo := any(NewTransactionPostgresRepository(db)).(offlineCreateRepository)
	for _, key := range []string{"", "   ", "bad\nkey", "khóa", strings.Repeat("x", 129)} {
		if _, _, err := repo.CreateIdempotent(offlineTransaction(owner, wallet, category), key, "hash"); err == nil {
			t.Fatalf("invalid key accepted: length%d", len(key))
		}
	}
	_, foreignWallet, _ := offlineOwner(t, db)
	if _, _, err := repo.CreateIdempotent(offlineTransaction(owner, foreignWallet, category), uuid.NewString(), "foreign-wallet-hash"); err == nil {
		t.Fatal("cross-owner wallet was accepted")
	}
	for _, model := range []any{&entity.Transaction{}, &entity.TransactionMutation{}} {
		var count int64
		if err := db.Model(model).Where("owner_id = ?", owner).Count(&count).Error; err != nil || count != 0 {
			t.Fatalf("rejected mutation left state: count%d err%v", count, err)
		}
	}
}

func TestOfflineReceiptSharedTransferPatchAndDeleteReplayStoredResult(t *testing.T) {
	db, repo, owner, pairID, _, _ := transferFixture(t)
	t.Cleanup(func() {
		if err := db.Where("owner_id = ?", owner).Delete(&entity.TransactionMutation{}).Error; err != nil {
			t.Error(err)
		}
	})
	update := transactionrepo.TransferUpdate{Amount: 23000, OccurredAt: time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC)}
	key := uuid.NewString()
	first, replay, err := repo.UpdateTransferIdempotent(owner, pairID, update, key, "patch-hash")
	if err != nil || replay || len(first) != 2 {
		t.Fatalf("patch: replay%v len%d err%v", replay, len(first), err)
	}
	second, replay, err := repo.UpdateTransferIdempotent(owner, pairID, update, key, "patch-hash")
	if err != nil || !replay || len(second) != 2 || second[0].ID != first[0].ID || second[0].Amount != 23000 {
		t.Fatalf("patch replay: replay%v len%d err%v", replay, len(second), err)
	}
	deleteKey := uuid.NewString()
	if replay, err := repo.DeleteTransferIdempotent(owner, pairID, deleteKey, "delete-hash"); err != nil || replay {
		t.Fatalf("delete: replay%v err%v", replay, err)
	}
	if replay, err := repo.DeleteTransferIdempotent(owner, pairID, deleteKey, "delete-hash"); err != nil || !replay {
		t.Fatalf("delete retry must succeed after pair is gone: replay%v err%v", replay, err)
	}
	// A stored patch receipt remains a replay, not a second write after deletion.
	if rows, replay, err := repo.UpdateTransferIdempotent(owner, pairID, update, key, "patch-hash"); err != nil || !replay || len(rows) != 2 {
		t.Fatalf("patch after delete: replay%v len%d err%v", replay, len(rows), err)
	}
	var count int64
	if err := db.Model(&entity.Transaction{}).Where("owner_id = ?", owner).Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("replay resurrected deleted rows: %d %v", count, err)
	}
}
