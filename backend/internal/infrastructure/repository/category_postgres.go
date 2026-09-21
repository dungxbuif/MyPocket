package repository

import (
	"errors"
	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	categoryrepo "github.com/mypocket/backend/internal/repository"
	"github.com/mypocket/backend/internal/usecase"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CategoryPostgresRepository struct{ db *gorm.DB }

func NewCategoryPostgresRepository(db *gorm.DB) categoryrepo.CategoryRepository {
	return &CategoryPostgresRepository{db: db}
}

func (r *CategoryPostgresRepository) EnsurePersonalDefaults(ownerID string) error {
	var templates []entity.Category
	if err := r.db.Where("owner_id IS NULL AND is_system = ?", false).Find(&templates).Error; err != nil {
		return err
	}
	if len(templates) == 0 {
		return nil
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		var personal []entity.Category
		if err := tx.Where("owner_id = ?", ownerID).Find(&personal).Error; err != nil {
			return err
		}
		existing := map[string]bool{}
		for _, item := range personal {
			existing[item.Name+"|"+item.Kind] = true
		}
		ids := map[string]string{}
		for _, item := range templates {
			if !existing[item.Name+"|"+item.Kind] {
				ids[item.ID] = uuid.NewString()
			}
		}
		for _, item := range templates {
			id := ids[item.ID]
			if id == "" {
				continue
			}
			copy := entity.Category{ID: id, OwnerID: &ownerID, Kind: item.Kind, Name: item.Name, IconKey: item.IconKey}
			if item.ParentID != nil {
				if parent := ids[*item.ParentID]; parent != "" {
					copy.ParentID = &parent
				}
			}
			if err := tx.Create(&copy).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *CategoryPostgresRepository) Create(ownerID string, category *entity.Category) error {
	category.OwnerID, category.IsSystem = &ownerID, false
	return r.db.Create(category).Error
}

func (r *CategoryPostgresRepository) FindPersonal(ownerID, id string) (*entity.Category, error) {
	var item entity.Category
	if err := r.db.Where("id = ? AND owner_id = ? AND is_system = ?", id, ownerID, false).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *CategoryPostgresRepository) FindVisible(ownerID, id string) (*entity.Category, error) {
	var item entity.Category
	if err := r.db.Where("id = ? AND (is_system = ? OR owner_id = ?)", id, true, ownerID).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *CategoryPostgresRepository) Update(ownerID string, category *entity.Category) error {
	return r.db.Where("id = ? AND owner_id = ? AND is_system = ?", category.ID, ownerID, false).Save(category).Error
}

func (r *CategoryPostgresRepository) Delete(ownerID, id string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&entity.Category{}).Where("parent_id = ? AND owner_id = ? AND is_system = ?", id, ownerID, false).Update("parent_id", nil).Error; err != nil {
			return err
		}
		result := tx.Where("id = ? AND owner_id = ? AND is_system = ?", id, ownerID, false).Delete(&entity.Category{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("category not found")
		}
		return nil
	})
}

func (r *CategoryPostgresRepository) ListVisible(ownerID string) ([]entity.Category, error) {
	var categories []entity.Category
	err := r.db.Where("is_system = ? OR owner_id = ?", true, ownerID).Order("kind, parent_id NULLS FIRST, name").Find(&categories).Error
	if err != nil || len(categories) == 0 {
		return categories, err
	}
	categoryIDs := make([]string, len(categories))
	for index, category := range categories {
		categoryIDs[index] = category.ID
	}
	var links []entity.CategoryWallet
	if err = r.db.Model(&entity.CategoryWallet{}).Select("category_wallets.*").Joins("JOIN wallets ON wallets.id = category_wallets.wallet_id").Where("category_id IN ? AND wallets.owner_id = ?", categoryIDs, ownerID).Find(&links).Error; err != nil {
		return nil, err
	}
	walletsByCategory := make(map[string][]string)
	for _, link := range links {
		walletsByCategory[link.CategoryID] = append(walletsByCategory[link.CategoryID], link.WalletID)
	}
	for index := range categories {
		categories[index].WalletIDs = append([]string{}, walletsByCategory[categories[index].ID]...)
	}
	return categories, nil
}

func (r *CategoryPostgresRepository) ReplaceWallets(ownerID string, category *entity.Category, walletIDs []string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var locked entity.Category
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND (is_system = true OR owner_id = ?)", category.ID, ownerID).First(&locked).Error; err != nil {
			return err
		}
		if len(walletIDs) > 0 {
			var count int64
			if err := tx.Model(&entity.Wallet{}).Where("id IN ? AND owner_id = ?", walletIDs, ownerID).Count(&count).Error; err != nil {
				return err
			}
			if count != int64(len(walletIDs)) {
				return usecase.ErrCategoryWalletInvalid
			}
		}
		if err := tx.Where("category_id = ? AND wallet_id IN (SELECT id FROM wallets WHERE owner_id = ?)", category.ID, ownerID).Delete(&entity.CategoryWallet{}).Error; err != nil {
			return err
		}
		for _, walletID := range walletIDs {
			if err := tx.Create(&entity.CategoryWallet{CategoryID: category.ID, WalletID: walletID}).Error; err != nil {
				return err
			}
		}
		category.WalletIDs = walletIDs
		return nil
	})
}

func (r *CategoryPostgresRepository) ValidateWallets(ownerID string, walletIDs []string) error {
	if len(walletIDs) == 0 {
		return nil
	}
	var count int64
	if err := r.db.Model(&entity.Wallet{}).Where("owner_id = ? AND id IN ?", ownerID, walletIDs).Count(&count).Error; err != nil {
		return err
	}
	if count != int64(len(walletIDs)) {
		return usecase.ErrCategoryWalletInvalid
	}
	return nil
}
