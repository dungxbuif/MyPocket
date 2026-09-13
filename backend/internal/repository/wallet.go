package repository

import "github.com/mypocket/backend/internal/entity"

type WalletRepository interface {
	List(ownerID string) ([]entity.Wallet, error)
	Find(ownerID, id string) (*entity.Wallet, error)
	Create(wallet *entity.Wallet) error
	Update(ownerID, id string, updates map[string]any) (*entity.Wallet, error)
	Delete(ownerID, id string) error
}
