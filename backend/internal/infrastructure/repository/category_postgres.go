package repository

import (
	"encoding/json"
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
			copy := entity.Category{ID: id, OwnerID: &ownerID, Kind: item.Kind, Name: item.Name, IconKey: item.IconKey, IsActive: true}
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
	return r.db.Transaction(func(tx *gorm.DB) error {
		return r.createInTx(tx, ownerID, category)
	})
}

func (r *CategoryPostgresRepository) createInTx(tx *gorm.DB, ownerID string, category *entity.Category) error {
	category.OwnerID, category.IsSystem, category.IsActive = &ownerID, false, true
	if category.Version < 1 {
		category.Version = 1
	}
	if len(category.WalletIDs) > 0 {
		var count int64
		if err := tx.Model(&entity.Wallet{}).Where("id IN ? AND owner_id = ?", category.WalletIDs, ownerID).Count(&count).Error; err != nil {
			return err
		}
		if count != int64(len(category.WalletIDs)) {
			return usecase.ErrCategoryWalletInvalid
		}
	}
	if err := tx.Create(category).Error; err != nil {
		return err
	}
	for _, walletID := range category.WalletIDs {
		if err := tx.Create(&entity.CategoryWallet{CategoryID: category.ID, WalletID: walletID}).Error; err != nil {
			return err
		}
	}
	return nil
}

// CreateIdempotent atomically creates a category and records its exact result
// in the shared durable mutation ledger. A retry with the same owner,
// operation, key, and request hash replays the stored category; a different
// hash is rejected as a key conflict.
func (r *CategoryPostgresRepository) CreateIdempotent(ownerID string, category *entity.Category, idempotencyKey, requestHash string) (*entity.Category, bool, error) {
	if category == nil || ownerID == "" || idempotencyKey == "" || requestHash == "" {
		return nil, false, categoryrepo.ErrMutationInvalid
	}
	var result entity.Category
	var replayed bool
	err := r.db.Transaction(func(tx *gorm.DB) error {
		mutation := entity.TransactionMutation{
			ID: uuid.NewString(), OwnerID: ownerID, Operation: "category.create",
			IdempotencyKey: idempotencyKey, RequestHash: requestHash,
			Status: "pending", ResponseJSON: "{}",
		}
		inserted := tx.Exec(`INSERT INTO transaction_mutations (id, owner_id, operation, idempotency_key, request_hash, status, response_json) VALUES (?, ?, ?, ?, ?, ?, ?::jsonb) ON CONFLICT (owner_id, operation, idempotency_key) DO NOTHING`, mutation.ID, mutation.OwnerID, mutation.Operation, mutation.IdempotencyKey, mutation.RequestHash, mutation.Status, mutation.ResponseJSON)
		if inserted.Error != nil {
			return inserted.Error
		}
		if inserted.RowsAffected == 0 {
			var persisted entity.TransactionMutation
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("owner_id = ? AND operation = ? AND idempotency_key = ?", ownerID, mutation.Operation, idempotencyKey).First(&persisted).Error; err != nil {
				return err
			}
			if persisted.RequestHash != requestHash {
				return categoryrepo.ErrMutationConflict
			}
			if persisted.Status != categoryrepo.TransactionMutationCompleted {
				return categoryrepo.ErrMutationPending
			}
			if err := json.Unmarshal([]byte(persisted.ResponseJSON), &result); err != nil {
				return err
			}
			replayed = true
			return nil
		}

		candidate := *category
		candidate.WalletIDs = append([]string(nil), category.WalletIDs...)
		if err := r.createInTx(tx, ownerID, &candidate); err != nil {
			return err
		}
		result = candidate
		encoded, err := json.Marshal(result)
		if err != nil {
			return err
		}
		updated := tx.Model(&entity.TransactionMutation{}).Where("id = ?", mutation.ID).Updates(map[string]any{
			"status": categoryrepo.TransactionMutationCompleted, "response_json": string(encoded),
		})
		if updated.Error != nil {
			return updated.Error
		}
		if updated.RowsAffected != 1 {
			return categoryrepo.ErrMutationPending
		}
		return nil
	})
	if err != nil {
		return nil, replayed, err
	}
	return &result, replayed, nil
}

func (r *CategoryPostgresRepository) FindPersonal(ownerID, id string) (*entity.Category, error) {
	var item entity.Category
	if err := r.db.Where("id = ? AND owner_id = ? AND is_system = ?", id, ownerID, false).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *CategoryPostgresRepository) SetActive(ownerID, id string, active bool) (*entity.Category, error) {
	var item entity.Category
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var current entity.Category
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND owner_id = ? AND is_system = ?", id, ownerID, false).First(&current).Error; err != nil {
			return err
		}
		if current.IsActive == active {
			item = current
			return tx.Table("category_wallets").
				Select("category_wallets.wallet_id").
				Joins("JOIN wallets ON wallets.id = category_wallets.wallet_id").
				Where("category_wallets.category_id = ? AND wallets.owner_id = ?", id, ownerID).
				Order("category_wallets.wallet_id").
				Pluck("category_wallets.wallet_id", &item.WalletIDs).Error
		}
		result := tx.Model(&entity.Category{}).
			Where("id = ? AND owner_id = ? AND is_system = ? AND version = ?", id, ownerID, false, maxCategoryVersion(current.Version)).
			Updates(map[string]any{"is_active": active, "version": gorm.Expr("version + 1")})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return categoryrepo.ErrCategoryVersionConflict
		}
		if err := tx.Where("id = ? AND owner_id = ? AND is_system = ?", id, ownerID, false).First(&item).Error; err != nil {
			return err
		}
		item.WalletIDs = make([]string, 0)
		return tx.Table("category_wallets").
			Select("category_wallets.wallet_id").
			Joins("JOIN wallets ON wallets.id = category_wallets.wallet_id").
			Where("category_wallets.category_id = ? AND wallets.owner_id = ?", id, ownerID).
			Order("category_wallets.wallet_id").
			Pluck("category_wallets.wallet_id", &item.WalletIDs).Error
	})
	if err != nil {
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
	return r.updateWithVersion(ownerID, category.ID, maxCategoryVersion(category.Version), category, category.WalletIDs)
}

func maxCategoryVersion(version int64) int64 {
	if version < 1 {
		return 1
	}
	return version
}

func (r *CategoryPostgresRepository) UpdateWithVersion(ownerID, id string, expectedVersion int64, category *entity.Category, walletIDs []string) error {
	return r.updateWithVersion(ownerID, id, expectedVersion, category, walletIDs)
}

func (r *CategoryPostgresRepository) updateWithVersion(ownerID, id string, expectedVersion int64, category *entity.Category, walletIDs []string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var current entity.Category
		if err := tx.Where("id = ? AND owner_id = ? AND is_system = ?", id, ownerID, false).First(&current).Error; err != nil {
			return err
		}
		if maxCategoryVersion(current.Version) != expectedVersion {
			return categoryrepo.ErrCategoryVersionConflict
		}
		updates := map[string]any{"name": category.Name, "kind": category.Kind, "parent_id": category.ParentID, "icon_key": category.IconKey, "version": gorm.Expr("version + 1")}
		result := tx.Model(&entity.Category{}).Where("id = ? AND owner_id = ? AND is_system = ? AND version = ?", id, ownerID, false, expectedVersion).Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return categoryrepo.ErrCategoryVersionConflict
		}
		if err := tx.Where("id = ? AND owner_id = ? AND is_system = ?", id, ownerID, false).First(category).Error; err != nil {
			return err
		}
		return r.replaceWalletsTx(tx, ownerID, category, walletIDs)
	})
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
		return r.replaceWalletsTx(tx, ownerID, category, walletIDs)
	})
}

func (r *CategoryPostgresRepository) replaceWalletsTx(tx *gorm.DB, ownerID string, category *entity.Category, walletIDs []string) error {
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
