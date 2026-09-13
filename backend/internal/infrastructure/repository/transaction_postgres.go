package repository

import (
	"errors"

	"github.com/mypocket/backend/internal/entity"
	transactionrepo "github.com/mypocket/backend/internal/repository"
	"gorm.io/gorm"
)

type TransactionPostgresRepository struct{ db *gorm.DB }

func NewTransactionPostgresRepository(db *gorm.DB) transactionrepo.TransactionRepository {
	return &TransactionPostgresRepository{db: db}
}

func (r *TransactionPostgresRepository) List(ownerID string) ([]entity.Transaction, error) {
	var transactions []entity.Transaction
	err := r.db.Where("owner_id = ?", ownerID).Order("occurred_at DESC, created_at DESC").Find(&transactions).Error
	return transactions, err
}

func (r *TransactionPostgresRepository) Create(transaction *entity.Transaction) error { return r.db.Create(transaction).Error }

func (r *TransactionPostgresRepository) Update(ownerID, id string, updates map[string]any) (*entity.Transaction, error) {
	var transaction entity.Transaction
	if err := r.db.Where("id = ? AND owner_id = ?", id, ownerID).First(&transaction).Error; err != nil {
		return nil, err
	}
	if err := r.db.Model(&transaction).Updates(updates).Error; err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *TransactionPostgresRepository) Delete(ownerID, id string) error {
	result := r.db.Where("id = ? AND owner_id = ?", id, ownerID).Delete(&entity.Transaction{})
	if result.Error != nil { return result.Error }
	if result.RowsAffected == 0 { return errors.New("transaction not found") }
	return nil
}
