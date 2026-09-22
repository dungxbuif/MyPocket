package repository

import (
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
	return r.db.Transaction(func(tx *gorm.DB) error {
		var user entity.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", owner).First(&user).Error; err != nil {
			return err
		}
		if !creating {
			var existing entity.Budget
			if err := tx.Where("owner_id = ? AND id = ?", owner, budget.ID).First(&existing).Error; err != nil {
				return err
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
			return tx.Create(budget).Error
		}
		result := tx.Model(&entity.Budget{}).Where("id = ? AND owner_id = ?", budget.ID, owner).Updates(map[string]any{"name": budget.Name, "limit_amount": budget.LimitAmount, "wallet_id": budget.WalletID, "category_id": budget.CategoryID, "start_date": budget.StartDate, "end_date": budget.EndDate})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return tx.Where("id = ? AND owner_id = ?", budget.ID, owner).First(budget).Error
	})
}
func (r *BudgetPostgresRepository) Delete(owner, id string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var user entity.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", owner).First(&user).Error; err != nil {
			return err
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
