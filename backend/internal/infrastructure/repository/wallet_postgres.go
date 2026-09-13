package repository

import (
	"errors"

	"github.com/mypocket/backend/internal/entity"
	walletrepo "github.com/mypocket/backend/internal/repository"
	"gorm.io/gorm"
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
	return r.db.Create(wallet).Error
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
