package repository

import (
	"errors"
	"time"

	"github.com/mypocket/backend/internal/entity"
)

var (
	ErrTransferInvalid       = errors.New("invalid internal transfer")
	ErrTransferWalletInvalid = errors.New("invalid transfer wallet")
	ErrTransferNotFound      = errors.New("transfer not found")
	ErrTransferPairInvalid   = errors.New("invalid transfer pair")
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
	Update(ownerID, id string, updates map[string]any) (*entity.Transaction, error)
	Delete(ownerID, id string) error
}
