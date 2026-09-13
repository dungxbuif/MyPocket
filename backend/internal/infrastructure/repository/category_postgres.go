package repository

import (
	"errors"
	"github.com/mypocket/backend/internal/entity"
	categoryrepo "github.com/mypocket/backend/internal/repository"
	"gorm.io/gorm"
)

type CategoryPostgresRepository struct{ db *gorm.DB }

func NewCategoryPostgresRepository(db *gorm.DB) categoryrepo.CategoryRepository {
	return &CategoryPostgresRepository{db: db}
}

func (r *CategoryPostgresRepository) Create(ownerID string, category *entity.Category) error { category.OwnerID = &ownerID; category.IsSystem = false; return r.db.Create(category).Error }
func (r *CategoryPostgresRepository) Update(ownerID, id string, updates map[string]any) (*entity.Category, error) { var item entity.Category; if err := r.db.Where("id = ? AND owner_id = ? AND is_system = ?", id, ownerID, false).First(&item).Error; err != nil { return nil, err }; if err := r.db.Model(&item).Updates(updates).Error; err != nil { return nil, err }; return &item, nil }
func (r *CategoryPostgresRepository) Delete(ownerID, id string) error { result := r.db.Where("id = ? AND owner_id = ? AND is_system = ?", id, ownerID, false).Delete(&entity.Category{}); if result.Error != nil { return result.Error }; if result.RowsAffected == 0 { return errors.New("category not found") }; return nil }

func (r *CategoryPostgresRepository) ListVisible(ownerID string) ([]entity.Category, error) {
	var categories []entity.Category
	err := r.db.Where("is_system = ? OR owner_id = ?", true, ownerID).Order("kind, parent_id NULLS FIRST, name").Find(&categories).Error
	return categories, err
}
