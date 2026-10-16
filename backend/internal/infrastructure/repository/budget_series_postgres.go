package repository

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	contract "github.com/mypocket/backend/internal/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
	"time"
)

type BudgetSeriesPostgresRepository struct{ db *gorm.DB }

func NewBudgetSeriesPostgresRepository(db *gorm.DB) contract.BudgetSeriesRepository {
	return &BudgetSeriesPostgresRepository{db}
}

func (r *BudgetSeriesPostgresRepository) List(owner string, now time.Time) ([]entity.BudgetSeries, error) {
	var user entity.User
	if err := r.db.Where("id = ?", owner).First(&user).Error; err != nil {
		return nil, err
	}
	location, err := time.LoadLocation(user.Timezone)
	if err != nil {
		return nil, err
	}
	rows := []entity.BudgetSeries{}
	if err = r.db.Where("owner_id = ?", owner).Order("created_at DESC, id DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	for i := range rows {
		if err = loadBudgetSeriesConfigs(r.db, &rows[i], now.In(location)); err != nil {
			return nil, err
		}
	}
	return rows, nil
}

func loadBudgetSeriesConfigs(tx *gorm.DB, s *entity.BudgetSeries, localNow time.Time) error {
	configs := []entity.BudgetSeriesConfig{}
	if err := tx.Where("series_id = ?", s.ID).Order("effective_start_date ASC").Find(&configs).Error; err != nil {
		return err
	}
	today, _ := entity.ParseCalendarDate(localNow.Format("2006-01-02"))
	s.CurrentConfig = nil
	s.PendingConfig = nil
	for i := range configs {
		cfg := configs[i]
		if !cfg.EffectiveStartDate.After(today.Time) {
			s.CurrentConfig = &cfg
		} else {
			s.PendingConfig = &cfg
			break
		}
	}
	if s.CurrentConfig == nil && len(configs) > 0 {
		cfg := configs[0]
		s.CurrentConfig = &cfg
		s.PendingConfig = nil
		if len(configs) > 1 {
			pending := configs[1]
			s.PendingConfig = &pending
		}
	}
	return nil
}

func (r *BudgetSeriesPostgresRepository) Mutate(owner string, series *entity.BudgetSeries, config *entity.BudgetSeriesConfig, operation, key, hash string, expected int64, now time.Time) (*entity.BudgetSeries, bool, error) {
	if series == nil || owner == "" || key == "" || hash == "" || (operation != "create" && operation != "update" && operation != "stop") {
		return nil, false, contract.ErrMutationInvalid
	}
	var result entity.BudgetSeries
	replayed := false
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var user entity.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", owner).First(&user).Error; err != nil {
			return err
		}
		op := "budget_series." + operation
		var prior entity.TransactionMutation
		lookup := tx.Where("owner_id = ? AND operation = ? AND idempotency_key = ?", owner, op, key).Find(&prior)
		if lookup.Error != nil {
			return lookup.Error
		}
		if lookup.RowsAffected > 0 {
			if prior.RequestHash != hash {
				return contract.ErrMutationConflict
			}
			if prior.Status != contract.TransactionMutationCompleted {
				return contract.ErrMutationPending
			}
			replayed = true
			return json.Unmarshal([]byte(prior.ResponseJSON), &result)
		}
		location, err := time.LoadLocation(user.Timezone)
		if err != nil {
			return err
		}
		local := now.In(location)
		if operation == "create" {
			result = *series
			result.OwnerID = owner
			result.Active = true
			result.Version = 1
			result.NextStartDate = result.StartDate
			result.GenerationError = nil
			result.CreatedAt = time.Time{}
			result.UpdatedAt = time.Time{}
			if _, err := entity.NextBudgetPeriod(result.StartDate, result.Cadence); err != nil {
				return gorm.ErrInvalidData
			}
			current, err := entity.BudgetPeriodStart(local, result.Cadence)
			if err != nil || result.StartDate.Before(current.Time) {
				return gorm.ErrInvalidData
			}
			if result.ID == "" {
				result.ID = uuid.NewString()
			}
			if err = validateBudgetSeriesConfig(tx, owner, config, nil); err != nil {
				return err
			}
			if err = tx.Create(&result).Error; err != nil {
				return err
			}
			cfg := *config
			cfg.ID = uuid.NewString()
			cfg.SeriesID = result.ID
			cfg.EffectiveStartDate = result.StartDate
			cfg.CreatedAt = time.Time{}
			if err = tx.Create(&cfg).Error; err != nil {
				return err
			}
		} else {
			if err = tx.Where("owner_id = ? AND id = ?", owner, series.ID).First(&result).Error; err != nil {
				return err
			}
			if result.Version != expected {
				return contract.ErrBudgetSeriesVersion
			}
			if !result.Active {
				return contract.ErrBudgetSeriesStopped
			}
			if operation == "update" {
				if err = loadBudgetSeriesConfigs(tx, &result, local); err != nil {
					return err
				}
				previous := result.CurrentConfig
				if config != nil && config.CategoryID != nil && result.PendingConfig != nil && result.PendingConfig.CategoryID != nil && *config.CategoryID == *result.PendingConfig.CategoryID {
					previous = result.PendingConfig
				}
				if err = validateBudgetSeriesConfig(tx, owner, config, previous); err != nil {
					return err
				}
				current, err := entity.BudgetPeriodStart(local, result.Cadence)
				if err != nil {
					return err
				}
				effective, err := entity.NextBudgetPeriod(current, result.Cadence)
				if err != nil {
					return err
				}
				if effective.Before(result.StartDate.Time) {
					effective = result.StartDate
				}
				cfg := *config
				cfg.ID = uuid.NewString()
				cfg.SeriesID = result.ID
				cfg.EffectiveStartDate = effective
				cfg.CreatedAt = time.Time{}
				if err = tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "series_id"}, {Name: "effective_start_date"}}, DoUpdates: clause.AssignmentColumns([]string{"name", "limit_amount", "wallet_id", "category_id", "carryover_enabled"})}).Create(&cfg).Error; err != nil {
					return err
				}
			}
			updates := map[string]any{"version": result.Version + 1, "updated_at": now}
			if operation == "stop" {
				updates["active"] = false
			}
			if err = tx.Model(&entity.BudgetSeries{}).Where("id = ? AND owner_id = ?", result.ID, owner).Updates(updates).Error; err != nil {
				return err
			}
		}
		if _, err = r.ensureInTx(tx, owner, local); err != nil {
			return err
		}
		if err = tx.Where("id = ? AND owner_id = ?", result.ID, owner).First(&result).Error; err != nil {
			return err
		}
		if err = loadBudgetSeriesConfigs(tx, &result, local); err != nil {
			return err
		}
		response, err := json.Marshal(result)
		if err != nil {
			return err
		}
		receipt := entity.TransactionMutation{ID: uuid.NewString(), OwnerID: owner, Operation: op, IdempotencyKey: key, RequestHash: hash, Status: contract.TransactionMutationCompleted, ResponseJSON: string(response)}
		return tx.Create(&receipt).Error
	})
	if err != nil {
		return nil, replayed, err
	}
	return &result, replayed, nil
}

func validateBudgetSeriesConfig(tx *gorm.DB, owner string, cfg, previous *entity.BudgetSeriesConfig) error {
	if cfg == nil || strings.TrimSpace(cfg.Name) == "" || len(cfg.Name) > 200 || cfg.LimitAmount <= 0 || cfg.LimitAmount > entity.MaxBudgetMoney {
		return gorm.ErrInvalidData
	}
	if cfg.WalletID != nil {
		var wallet entity.Wallet
		if err := tx.Clauses(clause.Locking{Strength: "SHARE"}).Where("id = ? AND owner_id = ?", *cfg.WalletID, owner).First(&wallet).Error; err != nil {
			return err
		}
		if wallet.Type == entity.WalletTypeCredit {
			return gorm.ErrInvalidData
		}
	}
	if cfg.CategoryID != nil {
		var category entity.Category
		if err := tx.Clauses(clause.Locking{Strength: "SHARE"}).Where("id = ? AND (owner_id = ? OR is_system = true)", *cfg.CategoryID, owner).First(&category).Error; err != nil {
			return err
		}
		if category.Kind != entity.TransactionTypeExpense {
			return gorm.ErrInvalidData
		}
		if !category.IsActive && (previous == nil || previous.CategoryID == nil || *previous.CategoryID != category.ID) {
			return contract.ErrCategoryInactive
		}
	}
	return nil
}

func (r *BudgetSeriesPostgresRepository) Ensure(owner string, now time.Time) (int, error) {
	count := 0
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var user entity.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", owner).First(&user).Error; err != nil {
			return err
		}
		location, err := time.LoadLocation(user.Timezone)
		if err != nil {
			return err
		}
		count, err = r.ensureInTx(tx, owner, now.In(location))
		return err
	})
	return count, err
}

func (r *BudgetSeriesPostgresRepository) ensureInTx(tx *gorm.DB, owner string, localNow time.Time) (int, error) {
	today, _ := entity.ParseCalendarDate(localNow.Format("2006-01-02"))
	rows := []entity.BudgetSeries{}
	if err := tx.Where("owner_id = ? AND active = true AND next_start_date <= ?", owner, today).Order("next_start_date ASC, id ASC").Find(&rows).Error; err != nil {
		return 0, err
	}
	count := 0
	for _, s := range rows {
		for !s.NextStartDate.After(today.Time) && count < 100 {
			next, err := entity.NextBudgetPeriod(s.NextStartDate, s.Cadence)
			if err != nil {
				return count, err
			}
			var cfg entity.BudgetSeriesConfig
			if err = tx.Where("series_id = ? AND effective_start_date <= ?", s.ID, s.NextStartDate).Order("effective_start_date DESC").First(&cfg).Error; err != nil {
				return count, err
			}
			end := entity.CalendarDate{Time: next.AddDate(0, 0, -1)}
			var conflicts int64
			if err = tx.Model(&entity.Budget{}).Where("owner_id = ? AND wallet_id IS NOT DISTINCT FROM ? AND category_id IS NOT DISTINCT FROM ? AND start_date <= ? AND end_date >= ? AND (series_id IS NULL OR series_id <> ?)", owner, cfg.WalletID, cfg.CategoryID, end, s.NextStartDate, s.ID).Count(&conflicts).Error; err != nil {
				return count, err
			}
			if conflicts > 0 {
				if err = tx.Model(&entity.BudgetSeries{}).Where("id = ?", s.ID).Updates(map[string]any{"generation_error": "BUDGET_OVERLAP", "updated_at": localNow}).Error; err != nil {
					return count, err
				}
				break
			}
			// Revalidate persisted scopes; a deletion must never become an all-scope row.
			if cfg.WalletID != nil {
				var found int64
				if err = tx.Model(&entity.Wallet{}).Where("id = ? AND owner_id = ?", *cfg.WalletID, owner).Count(&found).Error; err != nil {
					return count, err
				}
				if found == 0 {
					return count, gorm.ErrRecordNotFound
				}
			}
			if cfg.CategoryID != nil {
				var found int64
				if err = tx.Model(&entity.Category{}).Where("id = ? AND (owner_id = ? OR is_system = true)", *cfg.CategoryID, owner).Count(&found).Error; err != nil {
					return count, err
				}
				if found == 0 {
					return count, gorm.ErrRecordNotFound
				}
			}
			row := entity.Budget{ID: uuid.NewString(), OwnerID: owner, Name: cfg.Name, LimitAmount: cfg.LimitAmount, WalletID: cfg.WalletID, CategoryID: cfg.CategoryID, StartDate: s.NextStartDate, EndDate: end, SeriesID: &s.ID, CarryoverEnabled: cfg.CarryoverEnabled}
			inserted := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&row)
			if inserted.Error != nil {
				return count, inserted.Error
			}
			count += int(inserted.RowsAffected)
			s.NextStartDate = next
			if err = tx.Model(&entity.BudgetSeries{}).Where("id = ? AND owner_id = ?", s.ID, owner).Updates(map[string]any{"next_start_date": next, "generation_error": nil, "updated_at": localNow}).Error; err != nil {
				return count, err
			}
		}
		if count >= 100 {
			break
		}
	}
	return count, nil
}
