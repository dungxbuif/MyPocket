package repository

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	travelrepo "github.com/mypocket/backend/internal/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TravelPostgresRepository struct{ db *gorm.DB }

func NewTravelPostgresRepository(db *gorm.DB) travelrepo.TravelRepository {
	return &TravelPostgresRepository{db: db}
}

func (r *TravelPostgresRepository) List(ownerID string) ([]entity.TravelEvent, error) {
	var rows []entity.TravelEvent
	if err := r.db.Where("owner_id = ?", ownerID).Order("active DESC, starts_on NULLS LAST, created_at DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *TravelPostgresRepository) Find(ownerID, id string) (*entity.TravelEvent, error) {
	var row entity.TravelEvent
	if err := r.db.Where("owner_id = ? AND id = ?", ownerID, id).First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *TravelPostgresRepository) Active(ownerID string) (*entity.TravelEvent, error) {
	var row entity.TravelEvent
	if err := r.db.Where("owner_id = ? AND active = true", ownerID).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *TravelPostgresRepository) Create(ownerID string, event *entity.TravelEvent) error {
	if err := validateTravelEvent(ownerID, event); err != nil {
		return err
	}
	if event.ID == "" {
		event.ID = uuid.NewString()
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		if event.Active {
			if err := tx.Model(&entity.TravelEvent{}).Where("owner_id = ? AND active = true", ownerID).Update("active", false).Error; err != nil {
				return err
			}
		}
		return tx.Create(event).Error
	})
}

func (r *TravelPostgresRepository) Update(ownerID, id string, updates travelrepo.TravelUpdate) (*entity.TravelEvent, error) {
	if strings.TrimSpace(updates.Name) == "" || !travelDatesValid(updates.StartsOn, updates.EndsOn) {
		return nil, travelrepo.ErrTravelInvalid
	}
	var result entity.TravelEvent
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var current entity.TravelEvent
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("owner_id = ? AND id = ?", ownerID, id).First(&current).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return travelrepo.ErrTravelNotFound
			}
			return err
		}
		if err := tx.Model(&current).Updates(map[string]any{"name": strings.TrimSpace(updates.Name), "context": updates.Context, "starts_on": updates.StartsOn, "ends_on": updates.EndsOn}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", current.ID).First(&result).Error
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *TravelPostgresRepository) Delete(ownerID, id string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var event entity.TravelEvent
		if err := tx.Where("owner_id = ? AND id = ?", ownerID, id).First(&event).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return travelrepo.ErrTravelNotFound
			}
			return err
		}
		if err := tx.Model(&entity.Transaction{}).Where("owner_id = ? AND travel_event_id = ?", ownerID, id).Update("travel_event_id", nil).Error; err != nil {
			return err
		}
		return tx.Delete(&event).Error
	})
}

func (r *TravelPostgresRepository) SetActive(ownerID, id string, active bool) (*entity.TravelEvent, error) {
	var result entity.TravelEvent
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var event entity.TravelEvent
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("owner_id = ? AND id = ?", ownerID, id).First(&event).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return travelrepo.ErrTravelNotFound
			}
			return err
		}
		if active {
			if err := tx.Model(&entity.TravelEvent{}).Where("owner_id = ? AND id <> ? AND active = true", ownerID, id).Update("active", false).Error; err != nil {
				return err
			}
		}
		if err := tx.Model(&event).Update("active", active).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", id).First(&result).Error
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *TravelPostgresRepository) LinkTransaction(ownerID, transactionID string, eventID *string) (*entity.Transaction, error) {
	var result entity.Transaction
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var row entity.Transaction
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("owner_id = ? AND id = ?", ownerID, transactionID).First(&row).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return travelrepo.ErrTravelTransaction
			}
			return err
		}
		if row.TransferID != nil || row.CreditKind != nil || row.Type == entity.TransactionTypeAdjustment {
			return travelrepo.ErrTravelTransaction
		}
		if eventID != nil {
			var event entity.TravelEvent
			if err := tx.Where("owner_id = ? AND id = ?", ownerID, strings.TrimSpace(*eventID)).First(&event).Error; err != nil {
				return travelrepo.ErrTravelNotFound
			}
		}
		if err := tx.Model(&row).Update("travel_event_id", eventID).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", row.ID).First(&result).Error
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func validateTravelEvent(ownerID string, event *entity.TravelEvent) error {
	if event == nil || strings.TrimSpace(ownerID) == "" || event.OwnerID != ownerID || strings.TrimSpace(event.Name) == "" || !travelDatesValid(event.StartsOn, event.EndsOn) {
		return travelrepo.ErrTravelInvalid
	}
	return nil
}

func travelDatesValid(startsOn, endsOn *entity.CalendarDate) bool {
	if startsOn != nil && !startsOn.IsZero() && endsOn != nil && !endsOn.IsZero() && endsOn.Time.Before(startsOn.Time) {
		return false
	}
	return true
}
