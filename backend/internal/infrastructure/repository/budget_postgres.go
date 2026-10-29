package repository

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	contract "github.com/mypocket/backend/internal/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type BudgetPostgresRepository struct{ db *gorm.DB }

func NewBudgetPostgresRepository(db *gorm.DB) contract.BudgetRepository {
	return &BudgetPostgresRepository{db}
}
func (r *BudgetPostgresRepository) List(owner string) ([]entity.Budget, error) {
	items := []entity.Budget{}
	err := r.db.Where("owner_id = ?", owner).Order("start_date DESC, created_at DESC").Find(&items).Error
	return items, err
}
func (r *BudgetPostgresRepository) Save(owner string, budget *entity.Budget, creating bool) error {
	return r.db.Transaction(func(tx *gorm.DB) error { return r.saveInTx(tx, owner, budget, creating) })
}

func (r *BudgetPostgresRepository) saveInTx(tx *gorm.DB, owner string, budget *entity.Budget, creating bool) error {
	var user entity.User
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", owner).First(&user).Error; err != nil {
		return err
	}
	var existing entity.Budget
	if !creating {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("owner_id = ? AND id = ?", owner, budget.ID).First(&existing).Error; err != nil {
			return err
		}
		if existing.SeriesID != nil {
			return contract.ErrBudgetSeriesManaged
		}
		var user entity.User
		if err := tx.Where("id = ?", owner).First(&user).Error; err != nil {
			return err
		}
		location, err := time.LoadLocation(user.Timezone)
		if err != nil {
			return err
		}
		today := time.Now().In(location)
		end := existing.EndDate.Time
		if today.Year() > end.Year() || (today.Year() == end.Year() && today.Month() > end.Month()) || (today.Year() == end.Year() && today.Month() == end.Month() && today.Day() > end.Day()) {
			return contract.ErrBudgetEnded
		}
		budget.CreatedAt = existing.CreatedAt
	}
	if budget.CategoryID != nil {
		var category entity.Category
		if err := tx.Clauses(clause.Locking{Strength: "SHARE"}).Where("id = ? AND (owner_id = ? OR is_system = true)", *budget.CategoryID, owner).First(&category).Error; err != nil {
			return err
		}
		keepArchivedReference := !creating && existing.CategoryID != nil && *existing.CategoryID == category.ID
		if category.Kind != entity.TransactionTypeExpense {
			return gorm.ErrInvalidData
		}
		if !category.IsActive && !keepArchivedReference {
			return contract.ErrCategoryInactive
		}
	}
	var count int64
	err := tx.Model(&entity.Budget{}).Where("owner_id = ? AND id <> ? AND wallet_id IS NOT DISTINCT FROM ? AND category_id IS NOT DISTINCT FROM ? AND start_date <= ? AND end_date >= ?", owner, budget.ID, budget.WalletID, budget.CategoryID, budget.EndDate, budget.StartDate).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return contract.ErrBudgetConflict
	}
	budget.OwnerID = owner
	if creating {
		if err := tx.Create(budget).Error; err != nil {
			return err
		}
		budget.BaseLimitAmount = budget.LimitAmount
		budget.EffectiveLimitAmount = budget.LimitAmount
		return nil
	}
	result := tx.Model(&entity.Budget{}).Where("id = ? AND owner_id = ?", budget.ID, owner).Updates(map[string]any{"name": budget.Name, "limit_amount": budget.LimitAmount, "wallet_id": budget.WalletID, "category_id": budget.CategoryID, "start_date": budget.StartDate, "end_date": budget.EndDate})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	if err := tx.Where("id = ? AND owner_id = ?", budget.ID, owner).First(budget).Error; err != nil {
		return err
	}
	budget.BaseLimitAmount = budget.LimitAmount
	budget.EffectiveLimitAmount = budget.LimitAmount
	return nil
}

// CreateIdempotent atomically creates a budget and records the response in the
// shared mutation ledger. It is used by confirmed assistant actions.
func (r *BudgetPostgresRepository) CreateIdempotent(owner string, budget *entity.Budget, idempotencyKey, requestHash string) (*entity.Budget, bool, error) {
	if owner == "" || budget == nil || idempotencyKey == "" || requestHash == "" {
		return nil, false, contract.ErrMutationInvalid
	}
	var result entity.Budget
	replayed := false
	err := r.db.Transaction(func(tx *gorm.DB) error {
		mutation := entity.TransactionMutation{ID: uuid.NewString(), OwnerID: owner, Operation: "budget.create", IdempotencyKey: idempotencyKey, RequestHash: requestHash, Status: "pending", ResponseJSON: "{}"}
		inserted := tx.Exec(`INSERT INTO transaction_mutations (id, owner_id, operation, idempotency_key, request_hash, status, response_json) VALUES (?, ?, ?, ?, ?, ?, ?::jsonb) ON CONFLICT (owner_id, operation, idempotency_key) DO NOTHING`, mutation.ID, mutation.OwnerID, mutation.Operation, mutation.IdempotencyKey, mutation.RequestHash, "pending", "{}")
		if inserted.Error != nil {
			return inserted.Error
		}
		if inserted.RowsAffected == 0 {
			var persisted entity.TransactionMutation
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("owner_id = ? AND operation = ? AND idempotency_key = ?", owner, mutation.Operation, idempotencyKey).First(&persisted).Error; err != nil {
				return err
			}
			if persisted.RequestHash != requestHash {
				return contract.ErrMutationConflict
			}
			if persisted.Status != contract.TransactionMutationCompleted {
				return contract.ErrMutationPending
			}
			if err := json.Unmarshal([]byte(persisted.ResponseJSON), &result); err != nil {
				return err
			}
			replayed = true
			return nil
		}
		candidate := *budget
		if err := r.saveInTx(tx, owner, &candidate, true); err != nil {
			return err
		}
		result = candidate
		encoded, err := json.Marshal(result)
		if err != nil {
			return err
		}
		updated := tx.Model(&entity.TransactionMutation{}).Where("id = ?", mutation.ID).Updates(map[string]any{"status": contract.TransactionMutationCompleted, "response_json": string(encoded)})
		if updated.Error != nil {
			return updated.Error
		}
		if updated.RowsAffected != 1 {
			return contract.ErrMutationPending
		}
		return nil
	})
	if err != nil {
		return nil, replayed, err
	}
	return &result, replayed, nil
}
func (r *BudgetPostgresRepository) Delete(owner, id string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var user entity.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", owner).First(&user).Error; err != nil {
			return err
		}
		var budget entity.Budget
		if err := tx.Where("owner_id = ? AND id = ?", owner, id).First(&budget).Error; err != nil {
			return err
		}
		if budget.SeriesID != nil {
			return contract.ErrBudgetSeriesManaged
		}
		result := tx.Where("owner_id = ? AND id = ?", owner, id).Delete(&entity.Budget{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}
