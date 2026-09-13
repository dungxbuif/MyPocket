package repository

import "github.com/mypocket/backend/internal/entity"

type TransactionRepository interface {
	List(ownerID string) ([]entity.Transaction, error)
	Create(transaction *entity.Transaction) error
	Update(ownerID, id string, updates map[string]any) (*entity.Transaction, error)
	Delete(ownerID, id string) error
}
