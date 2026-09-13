package repository

import "github.com/mypocket/backend/internal/entity"

type CategoryRepository interface {
	EnsurePersonalDefaults(ownerID string) error
	ListVisible(ownerID string) ([]entity.Category, error)
	FindVisible(ownerID, id string) (*entity.Category, error)
	FindPersonal(ownerID, id string) (*entity.Category, error)
	Create(ownerID string, category *entity.Category) error
	Update(ownerID string, category *entity.Category) error
	Delete(ownerID, id string) error
	ReplaceWallets(ownerID string, category *entity.Category, walletIDs []string) error
	ValidateWallets(ownerID string, walletIDs []string) error
}
