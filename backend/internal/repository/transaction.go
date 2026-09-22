package repository

import (
	"errors"

	"github.com/mypocket/backend/internal/entity"
)

var (
	ErrTransferInvalid       = errors.New("invalid internal transfer")
	ErrTransferWalletInvalid = errors.New("invalid transfer wallet")
)

type TransactionRepository interface {
	List(ownerID string) ([]entity.Transaction, error)
	Find(ownerID, id string) (*entity.Transaction, error)
	Create(transaction *entity.Transaction) error
	Update(ownerID, id string, updates map[string]any) (*entity.Transaction, error)
	Delete(ownerID, id string) error
}
