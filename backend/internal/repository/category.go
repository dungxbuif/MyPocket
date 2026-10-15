package repository

import (
	"errors"

	"github.com/mypocket/backend/internal/entity"
)

var ErrCategoryInactive = errors.New("category is inactive")

type CategoryRepository interface {
	EnsurePersonalDefaults(ownerID string) error
	ListVisible(ownerID string) ([]entity.Category, error)
	FindVisible(ownerID, id string) (*entity.Category, error)
	FindPersonal(ownerID, id string) (*entity.Category, error)
	SetActive(ownerID, id string, active bool) (*entity.Category, error)
	Create(ownerID string, category *entity.Category) error
	Update(ownerID string, category *entity.Category) error
	Delete(ownerID, id string) error
	ReplaceWallets(ownerID string, category *entity.Category, walletIDs []string) error
	ValidateWallets(ownerID string, walletIDs []string) error
}
