package repository

import "github.com/mypocket/backend/internal/entity"

type CategoryRepository interface {
	ListVisible(ownerID string) ([]entity.Category, error)
	Create(ownerID string, category *entity.Category) error
	Update(ownerID, id string, updates map[string]any) (*entity.Category, error)
	Delete(ownerID, id string) error
}
