package repository

import (
	"errors"
	"time"

	"github.com/mypocket/backend/internal/entity"
)

var (
	ErrTransferInvalid         = errors.New("invalid internal transfer")
	ErrTransferWalletInvalid   = errors.New("invalid transfer wallet")
	ErrTransferNotFound        = errors.New("transfer not found")
	ErrTransferPairInvalid     = errors.New("invalid transfer pair")
	ErrAdjustmentInvalid       = errors.New("invalid balance adjustment")
	ErrAdjustmentWalletInvalid = errors.New("invalid adjustment wallet")
	ErrBulkDeleteInvalid       = errors.New("invalid bulk deletion")
	ErrBulkDeleteNotFound      = errors.New("bulk deletion transaction not found")
	ErrBulkDeleteLinked        = errors.New("bulk deletion includes linked transfer")
	ErrBulkDeleteAdjustment    = errors.New("bulk deletion includes immutable adjustment")
	ErrBulkDeleteCredit        = errors.New("bulk deletion includes credit ledger row")
	ErrCreditInvalid           = errors.New("invalid credit operation")
	ErrCreditWalletInvalid     = errors.New("invalid credit wallet")
	ErrCreditPaymentInvalid    = errors.New("invalid credit payment")
)

type TransferUpdate struct {
	Amount     int64
	OccurredAt time.Time
	Note       *string
}

type TransactionRepository interface {
	List(ownerID string) ([]entity.Transaction, error)
	Find(ownerID, id string) (*entity.Transaction, error)
	Create(transaction *entity.Transaction) error
	CreateAdjustment(transaction *entity.Transaction) error
	BulkDelete(ownerID string, ids []string) error
	Update(ownerID, id string, updates map[string]any) (*entity.Transaction, error)
	Delete(ownerID, id string) error
}

type CreditEntryInput struct {
	WalletID   string
	Kind       string
	CategoryID *string
	Amount     int64
	OccurredAt time.Time
	Note       *string
}

type CreditRepository interface {
	CreateCreditEntry(ownerID string, input CreditEntryInput) (*entity.Transaction, error)
	CreateCreditPayment(ownerID, creditWalletID, sourceWalletID string, amount int64, occurredAt time.Time, note *string) ([]entity.Transaction, error)
	ListCreditStatement(ownerID, walletID string, from, to *time.Time) (entity.CreditStatement, error)
}
