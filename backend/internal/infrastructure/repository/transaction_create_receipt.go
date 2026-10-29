package repository

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/mypocket/backend/internal/entity"
	transactionrepo "github.com/mypocket/backend/internal/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ transactionrepo.TransactionCreateMutationRepository = (*TransactionPostgresRepository)(nil)

// Peek a committed receipt before mutable wallet/category validation. A miss
// does not reserve a key; the atomic writer still resolves concurrent callers.
func (r *TransactionPostgresRepository) ReplayCreateReceipt(owner, operation, key, requestHash string) (json.RawMessage, bool, error) {
	if !validCreateReceipt(owner, key, requestHash) || (operation != "transaction.create" && operation != "transfer.create") {
		return nil, false, transactionrepo.ErrMutationInvalid
	}
	var receipt entity.TransactionMutation
	err := r.db.Where("owner_id = ? AND operation = ? AND idempotency_key = ?", owner, operation, key).Take(&receipt).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	if receipt.RequestHash != requestHash {
		return nil, false, transactionrepo.ErrMutationConflict
	}
	if receipt.Status != transactionrepo.TransactionMutationCompleted || !json.Valid([]byte(receipt.ResponseJSON)) {
		return nil, false, transactionrepo.ErrMutationPending
	}
	return json.RawMessage(receipt.ResponseJSON), true, nil
}

func validCreateReceipt(owner, key, hash string) bool {
	if strings.TrimSpace(owner) == "" || strings.TrimSpace(hash) == "" || strings.TrimSpace(key) == "" || len(key) > 128 {
		return false
	}
	for _, character := range key {
		if character < 32 || character > 126 {
			return false
		}
	}
	return true
}

func (r *TransactionPostgresRepository) CreateIdempotent(item *entity.Transaction, key, requestHash string) (*entity.Transaction, bool, error) {
	if item == nil || !validCreateReceipt(item.OwnerID, key, requestHash) || item.Amount <= 0 || item.OccurredAt.IsZero() ||
		(item.Type != entity.TransactionTypeIncome && item.Type != entity.TransactionTypeExpense) || item.TransferID != nil || item.CreditKind != nil {
		return nil, false, transactionrepo.ErrMutationInvalid
	}
	var result entity.Transaction
	replayed := false
	err := r.db.Transaction(func(tx *gorm.DB) error {
		mutation, created, err := r.reserveMutation(tx, item.OwnerID, "transaction.create", key, requestHash)
		if err != nil {
			return err
		}
		if !created {
			if err := json.Unmarshal([]byte(mutation.ResponseJSON), &result); err != nil {
				return err
			}
			replayed = true
			return nil
		}
		var wallet entity.Wallet
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND owner_id = ?", item.WalletID, item.OwnerID).First(&wallet).Error; err != nil {
			return err
		}
		if wallet.Type == entity.WalletTypeCredit {
			return transactionrepo.ErrCreditWalletInvalid
		}
		result = *item
		// The existing domain writer retains category/travel/jar validation and
		// explicit report exclusion. Its nested transaction is a savepoint;
		// the outer transaction still owns ledger plus receipt durability.
		if err := (&TransactionPostgresRepository{db: tx}).Create(&result); err != nil {
			return err
		}
		return r.completeMutation(tx, mutation.ID, result)
	})
	if err != nil {
		return nil, false, err
	}
	return &result, replayed, nil
}

func (r *TransactionPostgresRepository) CreateTransferIdempotent(owner string, source, destination *entity.Transaction, key, requestHash string) ([]entity.Transaction, bool, error) {
	if !validCreateReceipt(owner, key, requestHash) || source == nil || destination == nil || source.OwnerID != owner || destination.OwnerID != owner {
		return nil, false, transactionrepo.ErrMutationInvalid
	}
	var result []entity.Transaction
	replayed := false
	err := r.db.Transaction(func(tx *gorm.DB) error {
		mutation, created, err := r.reserveMutation(tx, owner, "transfer.create", key, requestHash)
		if err != nil {
			return err
		}
		if !created {
			if err := json.Unmarshal([]byte(mutation.ResponseJSON), &result); err != nil {
				return err
			}
			replayed = true
			return nil
		}
		from, to := *source, *destination
		if err := (&TransactionPostgresRepository{db: tx}).CreateTransfer(owner, &from, &to); err != nil {
			return err
		}
		result = []entity.Transaction{from, to}
		return r.completeMutation(tx, mutation.ID, result)
	})
	if err != nil {
		return nil, false, err
	}
	return result, replayed, nil
}
