package repository

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	recurringrepo "github.com/mypocket/backend/internal/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RecurringPostgresRepository struct{ db *gorm.DB }

func NewRecurringPostgresRepository(db *gorm.DB) recurringrepo.RecurringRepository {
	return &RecurringPostgresRepository{db: db}
}

func (r *RecurringPostgresRepository) List(ownerID string) ([]entity.RecurringSchedule, error) {
	var rows []entity.RecurringSchedule
	if err := r.db.Where("owner_id = ?", ownerID).Order("active DESC, next_run_at ASC, created_at ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *RecurringPostgresRepository) Find(ownerID, id string) (*entity.RecurringSchedule, error) {
	var row entity.RecurringSchedule
	if err := r.db.Where("owner_id = ? AND id = ?", ownerID, id).First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *RecurringPostgresRepository) Create(ownerID string, schedule *entity.RecurringSchedule) error {
	if err := validateRecurring(schedule, ownerID); err != nil {
		return err
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := validateRecurringWallet(tx, ownerID, schedule.WalletID); err != nil {
			return err
		}
		if err := validateRecurringCategory(tx, ownerID, schedule.CategoryID, schedule.Type); err != nil {
			return err
		}
		return tx.Create(schedule).Error
	})
}

func (r *RecurringPostgresRepository) Update(ownerID, id string, updates recurringrepo.RecurringUpdate) (*entity.RecurringSchedule, error) {
	if updates.Name == "" || updates.Amount <= 0 || updates.Interval <= 0 || updates.NextRunAt.IsZero() || !validRecurringType(updates.Type) || !validRecurringFrequency(updates.Frequency) {
		return nil, recurringrepo.ErrRecurringInvalid
	}
	returnValue := &entity.RecurringSchedule{}
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var current entity.RecurringSchedule
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("owner_id = ? AND id = ?", ownerID, id).First(&current).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return recurringrepo.ErrRecurringNotFound
			}
			return err
		}
		if err := validateRecurringWallet(tx, ownerID, current.WalletID); err != nil {
			return err
		}
		if err := validateRecurringCategory(tx, ownerID, updates.CategoryID, updates.Type); err != nil {
			return err
		}
		updates.AnchorDay = updates.NextRunAt.In(time.UTC).Day()
		updates.AnchorMonth = int(updates.NextRunAt.In(time.UTC).Month())
		if err := tx.Model(&current).Updates(map[string]any{
			"name": updates.Name, "amount": updates.Amount, "category_id": updates.CategoryID, "type": updates.Type,
			"note": updates.Note, "frequency": updates.Frequency, "interval": updates.Interval,
			"next_run_at": updates.NextRunAt, "ends_at": updates.EndsAt, "anchor_day": updates.AnchorDay,
			"anchor_month": updates.AnchorMonth, "active": updates.Active,
		}).Error; err != nil {
			return err
		}
		if err := tx.Where("id = ?", current.ID).First(&current).Error; err != nil {
			return err
		}
		returnValue = &current
		return nil
	})
	if err != nil {
		return nil, err
	}
	return returnValue, nil
}

func (r *RecurringPostgresRepository) Delete(ownerID, id string) error {
	result := r.db.Where("owner_id = ? AND id = ?", ownerID, id).Delete(&entity.RecurringSchedule{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return recurringrepo.ErrRecurringNotFound
	}
	return nil
}

func (r *RecurringPostgresRepository) RunDue(ownerID string, now time.Time) (int, error) {
	if ownerID == "" || now.IsZero() {
		return 0, recurringrepo.ErrRecurringInvalid
	}
	created := 0
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var user entity.User
		if err := tx.Where("id = ?", ownerID).First(&user).Error; err != nil {
			return err
		}
		location, err := time.LoadLocation(user.Timezone)
		if err != nil || user.Timezone == "Local" {
			return recurringrepo.ErrRecurringInvalid
		}
		var schedules []entity.RecurringSchedule
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("owner_id = ? AND active = true AND next_run_at <= ?", ownerID, now.UTC()).Order("next_run_at ASC, id ASC").Limit(100).Find(&schedules).Error; err != nil {
			return err
		}
		for index := range schedules {
			schedule := &schedules[index]
			for schedule.Active && !schedule.NextRunAt.After(now.UTC()) && created < 100 {
				due := schedule.NextRunAt
				if schedule.EndsAt != nil && due.After(*schedule.EndsAt) {
					schedule.Active = false
					break
				}
				transactionID := uuid.NewString()
				note := schedule.Note
				if note == nil || *note == "" {
					defaultNote := "Giao dịch định kỳ — " + schedule.Name
					note = &defaultNote
				}
				transaction := &entity.Transaction{ID: transactionID, OwnerID: ownerID, WalletID: schedule.WalletID, CategoryID: schedule.CategoryID, Type: schedule.Type, Amount: schedule.Amount, OccurredAt: due, Note: note, IncludedInReports: true}
				var existing entity.RecurringOccurrence
				lookupErr := tx.Where("schedule_id = ? AND due_at = ?", schedule.ID, due).First(&existing).Error
				if errors.Is(lookupErr, gorm.ErrRecordNotFound) {
					if err := tx.Create(transaction).Error; err != nil {
						return err
					}
					occurrence := &entity.RecurringOccurrence{ID: uuid.NewString(), ScheduleID: schedule.ID, DueAt: due, TransactionID: transactionID}
					if err := tx.Create(occurrence).Error; err != nil {
						return err
					}
					created++
				} else if lookupErr != nil {
					return lookupErr
				}
				schedule.NextRunAt = advanceRecurring(schedule, due, location)
				if schedule.EndsAt != nil && schedule.NextRunAt.After(*schedule.EndsAt) {
					schedule.Active = false
				}
			}
			if err := tx.Model(schedule).Updates(map[string]any{"next_run_at": schedule.NextRunAt, "active": schedule.Active}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return created, err
}

func validateRecurring(schedule *entity.RecurringSchedule, ownerID string) error {
	if schedule == nil || schedule.OwnerID != ownerID || schedule.ID == "" || schedule.Name == "" || schedule.Amount <= 0 || schedule.Interval <= 0 || schedule.NextRunAt.IsZero() || !validRecurringType(schedule.Type) || !validRecurringFrequency(schedule.Frequency) || schedule.AnchorDay < 1 || schedule.AnchorDay > 31 || schedule.AnchorMonth < 1 || schedule.AnchorMonth > 12 {
		return recurringrepo.ErrRecurringInvalid
	}
	if schedule.EndsAt != nil && schedule.EndsAt.Before(schedule.NextRunAt) {
		return recurringrepo.ErrRecurringInvalid
	}
	return nil
}

func validateRecurringWallet(tx *gorm.DB, ownerID, walletID string) error {
	var wallet entity.Wallet
	if err := tx.Where("owner_id = ? AND id = ?", ownerID, walletID).First(&wallet).Error; err != nil {
		return recurringrepo.ErrRecurringWalletInvalid
	}
	if wallet.Type == entity.WalletTypeCredit {
		return recurringrepo.ErrRecurringWalletInvalid
	}
	return nil
}

func validateRecurringCategory(tx *gorm.DB, ownerID string, categoryID *string, kind string) error {
	if categoryID == nil {
		return nil
	}
	var category entity.Category
	if err := tx.Where("id = ? AND (owner_id = ? OR owner_id IS NULL)", *categoryID, ownerID).First(&category).Error; err != nil || category.Kind != kind {
		return recurringrepo.ErrRecurringCategoryInvalid
	}
	return nil
}

func validRecurringType(value string) bool {
	return value == entity.TransactionTypeIncome || value == entity.TransactionTypeExpense
}
func validRecurringFrequency(value string) bool {
	return value == "daily" || value == "weekly" || value == "monthly" || value == "yearly"
}

func advanceRecurring(schedule *entity.RecurringSchedule, due time.Time, location *time.Location) time.Time {
	local := due.In(location)
	switch schedule.Frequency {
	case "daily":
		return local.AddDate(0, 0, schedule.Interval).UTC()
	case "weekly":
		return local.AddDate(0, 0, 7*schedule.Interval).UTC()
	case "monthly":
		monthStart := time.Date(local.Year(), local.Month(), 1, local.Hour(), local.Minute(), local.Second(), local.Nanosecond(), location).AddDate(0, schedule.Interval, 0)
		day := schedule.AnchorDay
		lastDay := monthStart.AddDate(0, 1, -1).Day()
		if day > lastDay {
			day = lastDay
		}
		return time.Date(monthStart.Year(), monthStart.Month(), day, local.Hour(), local.Minute(), local.Second(), local.Nanosecond(), location).UTC()
	case "yearly":
		yearStart := time.Date(local.Year()+schedule.Interval, time.Month(schedule.AnchorMonth), 1, local.Hour(), local.Minute(), local.Second(), local.Nanosecond(), location)
		day := schedule.AnchorDay
		lastDay := yearStart.AddDate(0, 1, -1).Day()
		if day > lastDay {
			day = lastDay
		}
		return time.Date(yearStart.Year(), yearStart.Month(), day, local.Hour(), local.Minute(), local.Second(), local.Nanosecond(), location).UTC()
	default:
		return due.AddDate(0, 0, schedule.Interval).UTC()
	}
}
