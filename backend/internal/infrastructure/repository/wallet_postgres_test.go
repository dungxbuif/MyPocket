package repository

import (
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	walletrepo "github.com/mypocket/backend/internal/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestWalletCreatePersistsExplicitlyExcludedFromTotal(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires migrated local TEST_DATABASE_URL")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	ownerID := uuid.NewString()
	if err := db.Create(&entity.User{ID: ownerID, Email: ownerID + "@wallet.test", GoogleSubject: ownerID, Timezone: "Asia/Ho_Chi_Minh"}).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Where("owner_id = ?", ownerID).Delete(&entity.Wallet{})
		db.Where("id = ?", ownerID).Delete(&entity.User{})
	})

	wallet := entity.Wallet{ID: uuid.NewString(), OwnerID: ownerID, Name: "Excluded", Type: entity.WalletTypeBasic, Currency: entity.WalletCurrencyVND, IsInTotal: false}
	repository := NewWalletPostgresRepository(db)
	if err := repository.Create(&wallet); err != nil {
		t.Fatal(err)
	}
	var persisted entity.Wallet
	if err := db.Where("id = ? AND owner_id = ?", wallet.ID, ownerID).First(&persisted).Error; err != nil {
		t.Fatal(err)
	}
	if persisted.IsInTotal {
		t.Fatalf("explicit is_in_total=false was stored as true: %#v", persisted)
	}
}

func TestWalletCreateIdempotentReplaysAndRejectsPayloadConflict(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires migrated local TEST_DATABASE_URL")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	ownerID := uuid.NewString()
	if err := db.Create(&entity.User{ID: ownerID, Email: ownerID + "@wallet-idempotency.test", GoogleSubject: ownerID, Timezone: "Asia/Ho_Chi_Minh"}).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Where("owner_id = ? AND operation = ?", ownerID, "wallet.create").Delete(&entity.TransactionMutation{})
		db.Where("owner_id = ?", ownerID).Delete(&entity.Wallet{})
		db.Where("id = ?", ownerID).Delete(&entity.User{})
	})

	repository := NewWalletPostgresRepository(db)
	first := &entity.Wallet{ID: uuid.NewString(), OwnerID: ownerID, Name: "Du lịch", Type: entity.WalletTypeGoal, Currency: entity.WalletCurrencyVND, IsInTotal: true, TargetAmount: int64Ptr(1000000)}
	created, replayed, err := repository.(interface {
		CreateIdempotent(*entity.Wallet, string, string) (*entity.Wallet, bool, error)
	}).CreateIdempotent(first, "wallet-confirm-1", "payload-1")
	if err != nil || replayed || created == nil {
		t.Fatalf("first create: wallet=%#v replayed=%v err=%v", created, replayed, err)
	}

	replay, replayed, err := repository.(interface {
		CreateIdempotent(*entity.Wallet, string, string) (*entity.Wallet, bool, error)
	}).CreateIdempotent(&entity.Wallet{ID: uuid.NewString(), OwnerID: ownerID, Name: "Du lịch", Type: entity.WalletTypeGoal, Currency: entity.WalletCurrencyVND, IsInTotal: true, TargetAmount: int64Ptr(1000000)}, "wallet-confirm-1", "payload-1")
	if err != nil || !replayed || replay == nil || replay.ID != first.ID {
		t.Fatalf("replay: wallet=%#v replayed=%v err=%v", replay, replayed, err)
	}

	_, _, err = repository.(interface {
		CreateIdempotent(*entity.Wallet, string, string) (*entity.Wallet, bool, error)
	}).CreateIdempotent(&entity.Wallet{ID: uuid.NewString(), OwnerID: ownerID, Name: "Khác", Type: entity.WalletTypeBasic, Currency: entity.WalletCurrencyVND}, "wallet-confirm-1", "payload-2")
	if !errors.Is(err, walletrepo.ErrMutationConflict) {
		t.Fatalf("expected payload conflict, got %v", err)
	}
	var count int64
	if err := db.Model(&entity.Wallet{}).Where("owner_id = ?", ownerID).Count(&count).Error; err != nil || count != 1 {
		t.Fatalf("wallet count=%d err=%v", count, err)
	}
}

func TestCalculateCurrentBalancesAppliesIncomeAndExpensePerWallet(t *testing.T) {
	wallets := []entity.Wallet{
		{ID: "wallet-1", OpeningBalance: 100000},
		{ID: "wallet-2", OpeningBalance: -50000},
	}
	transactions := []entity.Transaction{
		{WalletID: "wallet-1", Type: entity.TransactionTypeIncome, Amount: 25000},
		{WalletID: "wallet-1", Type: entity.TransactionTypeExpense, Amount: 10000},
		{WalletID: "wallet-2", Type: entity.TransactionTypeIncome, Amount: 5000},
	}

	calculateCurrentBalances(wallets, transactions)

	if wallets[0].CurrentBalance != 115000 {
		t.Fatalf("wallet-1 current balance = %d", wallets[0].CurrentBalance)
	}
	if wallets[1].CurrentBalance != -45000 {
		t.Fatalf("wallet-2 current balance = %d", wallets[1].CurrentBalance)
	}
}

func TestCalculateCurrentBalancesAppliesExplicitAdjustments(t *testing.T) {
	wallets := []entity.Wallet{{ID: "wallet-1", OpeningBalance: 100000}}
	increase := entity.AdjustmentDirectionIncrease
	decrease := entity.AdjustmentDirectionDecrease
	transactions := []entity.Transaction{
		{WalletID: "wallet-1", Type: entity.TransactionTypeAdjustment, Amount: 90000, AdjustmentDirection: &increase},
		{WalletID: "wallet-1", Type: entity.TransactionTypeAdjustment, Amount: 10000, AdjustmentDirection: &decrease},
	}

	calculateCurrentBalances(wallets, transactions)

	if wallets[0].CurrentBalance != 180000 {
		t.Fatalf("current balance = %d", wallets[0].CurrentBalance)
	}
}
