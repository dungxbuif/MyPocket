package repository

import (
	"errors"

	"github.com/mypocket/backend/internal/entity"
)

var ErrCategoryInactive = errors.New("category is inactive")
var ErrCategoryVersionConflict = errors.New("category version conflict")

// CategoryCreateMutationRepository is the durable receipt boundary for
// idempotent category creation. The response is persisted with the category
// write so a retried advisor confirmation cannot create a duplicate.
type CategoryCreateMutationRepository interface {
	CreateIdempotent(ownerID string, category *entity.Category, idempotencyKey, requestHash string) (*entity.Category, bool, error)
}

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

// VersionedCategoryRepository provides atomic compare-and-swap updates.
type VersionedCategoryRepository interface {
	CategoryRepository
	UpdateWithVersion(ownerID, id string, expectedVersion int64, category *entity.Category, walletIDs []string) error
}
