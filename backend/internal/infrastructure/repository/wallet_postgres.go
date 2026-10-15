package repository

import (
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	walletrepo "github.com/mypocket/backend/internal/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WalletPostgresRepository struct{ db *gorm.DB }

func NewWalletPostgresRepository(db *gorm.DB) walletrepo.WalletRepository {
	return &WalletPostgresRepository{db: db}
}

func (r *WalletPostgresRepository) List(ownerID string) ([]entity.Wallet, error) {
	var wallets []entity.Wallet
	if err := r.db.Where("owner_id = ?", ownerID).Order("created_at ASC").Find(&wallets).Error; err != nil {
		return nil, err
	}
	var transactions []entity.Transaction
	if err := r.db.Where("owner_id = ?", ownerID).Find(&transactions).Error; err != nil {
		return nil, err
	}
	calculateCurrentBalances(wallets, transactions)
	for index := range wallets {
		if wallets[index].Type == entity.WalletTypeCredit && wallets[index].CreditLimit != nil {
			wallets[index].AvailableCredit = *wallets[index].CreditLimit + wallets[index].CurrentBalance
		}
	}
	return wallets, nil
}

func calculateCurrentBalances(wallets []entity.Wallet, transactions []entity.Transaction) {
	indexByID := make(map[string]int, len(wallets))
	for index := range wallets {
		wallets[index].CurrentBalance = wallets[index].OpeningBalance
		indexByID[wallets[index].ID] = index
	}
	for _, transaction := range transactions {
		index, ok := indexByID[transaction.WalletID]
		if !ok {
			continue
		}
		switch transaction.Type {
		case entity.TransactionTypeIncome:
			wallets[index].CurrentBalance += transaction.Amount
		case entity.TransactionTypeExpense:
			wallets[index].CurrentBalance -= transaction.Amount
		case entity.TransactionTypeAdjustment:
			if transaction.AdjustmentDirection != nil && *transaction.AdjustmentDirection == entity.AdjustmentDirectionIncrease {
				wallets[index].CurrentBalance += transaction.Amount
			} else if transaction.AdjustmentDirection != nil && *transaction.AdjustmentDirection == entity.AdjustmentDirectionDecrease {
				wallets[index].CurrentBalance -= transaction.Amount
			}
		}
	}
}

func (r *WalletPostgresRepository) Find(ownerID, id string) (*entity.Wallet, error) {
	var wallet entity.Wallet
	if err := r.db.Where("id = ? AND owner_id = ?", id, ownerID).First(&wallet).Error; err != nil {
		return nil, err
	}
	return &wallet, nil
}

func (r *WalletPostgresRepository) Create(wallet *entity.Wallet) error {
	isInTotal := wallet.IsInTotal
	// `is_in_total` defaults to true. A plain Create treats false as a zero
	// value for GORM's default-tag handling. Persist the explicit caller value
	// in the same transaction so a deliberate opt-out cannot be lost.
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(wallet).Error; err != nil {
			return err
		}
		if err := tx.Model(&entity.Wallet{}).
			Where("id = ? AND owner_id = ?", wallet.ID, wallet.OwnerID).
			UpdateColumn("is_in_total", isInTotal).Error; err != nil {
			return err
		}
		wallet.IsInTotal = isInTotal
		return nil
	})
}

// CreateIdempotent atomically creates a wallet and records its response in the
// existing mutation ledger. The same owner/operation/key can therefore replay
// a confirmed write after a lost response without creating a second wallet.
func (r *WalletPostgresRepository) CreateIdempotent(wallet *entity.Wallet, idempotencyKey, requestHash string) (*entity.Wallet, bool, error) {
	if wallet == nil || wallet.OwnerID == "" || idempotencyKey == "" || requestHash == "" {
		return nil, false, walletrepo.ErrMutationInvalid
	}
	var replayed bool
	var result entity.Wallet
	err := r.db.Transaction(func(tx *gorm.DB) error {
		mutation := entity.TransactionMutation{
			ID: uuid.NewString(), OwnerID: wallet.OwnerID, Operation: "wallet.create",
			IdempotencyKey: idempotencyKey, RequestHash: requestHash, Status: "pending", ResponseJSON: "{}",
		}
		inserted := tx.Exec(`INSERT INTO transaction_mutations (id, owner_id, operation, idempotency_key, request_hash, status, response_json) VALUES (?, ?, ?, ?, ?, ?, ?::jsonb) ON CONFLICT (owner_id, operation, idempotency_key) DO NOTHING`, mutation.ID, mutation.OwnerID, mutation.Operation, mutation.IdempotencyKey, mutation.RequestHash, mutation.Status, mutation.ResponseJSON)
		if inserted.Error != nil {
			return inserted.Error
		}
		if inserted.RowsAffected == 0 {
			var persisted entity.TransactionMutation
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("owner_id = ? AND operation = ? AND idempotency_key = ?", wallet.OwnerID, mutation.Operation, idempotencyKey).First(&persisted).Error; err != nil {
				return err
			}
			if persisted.RequestHash != requestHash {
				return walletrepo.ErrMutationConflict
			}
			if persisted.Status != walletrepo.TransactionMutationCompleted {
				return walletrepo.ErrMutationPending
			}
			if err := json.Unmarshal([]byte(persisted.ResponseJSON), &result); err != nil {
				return err
			}
			replayed = true
			return nil
		}

		if err := tx.Create(wallet).Error; err != nil {
			return err
		}
		if err := tx.Model(&entity.Wallet{}).Where("id = ? AND owner_id = ?", wallet.ID, wallet.OwnerID).UpdateColumn("is_in_total", wallet.IsInTotal).Error; err != nil {
			return err
		}
		result = *wallet
		encoded, err := json.Marshal(result)
		if err != nil {
			return err
		}
		updated := tx.Model(&entity.TransactionMutation{}).Where("id = ?", mutation.ID).Updates(map[string]any{"status": walletrepo.TransactionMutationCompleted, "response_json": string(encoded)})
		if updated.Error != nil {
			return updated.Error
		}
		if updated.RowsAffected != 1 {
			return walletrepo.ErrMutationPending
		}
		return nil
	})
	if err != nil {
		return nil, replayed, err
	}
	return &result, replayed, nil
}

func (r *WalletPostgresRepository) Update(ownerID, id string, updates map[string]any) (*entity.Wallet, error) {
	var wallet entity.Wallet
	if err := r.db.Where("id = ? AND owner_id = ?", id, ownerID).First(&wallet).Error; err != nil {
		return nil, err
	}
	if err := r.db.Model(&wallet).Updates(updates).Error; err != nil {
		return nil, err
	}
	return &wallet, nil
}

func (r *WalletPostgresRepository) Delete(ownerID, id string) error {
	result := r.db.Where("id = ? AND owner_id = ?", id, ownerID).Delete(&entity.Wallet{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("wallet not found")
	}
	return nil
}
