package entity

import (
	"errors"
	"time"
)

var ErrBudgetCadence = errors.New("invalid cadence or noncanonical period start")

type BudgetSeries struct {
	ID              string              `json:"id" gorm:"primaryKey"`
	OwnerID         string              `json:"owner_id"`
	Cadence         string              `json:"cadence"`
	StartDate       CalendarDate        `json:"start_date" gorm:"type:date"`
	NextStartDate   CalendarDate        `json:"-" gorm:"type:date"`
	Active          bool                `json:"active"`
	Version         int64               `json:"version"`
	GenerationError *string             `json:"generation_error"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
	CurrentConfig   *BudgetSeriesConfig `json:"current_config" gorm:"-"`
	PendingConfig   *BudgetSeriesConfig `json:"pending_config" gorm:"-"`
}

type BudgetSeriesConfig struct {
	ID                 string       `json:"-" gorm:"primaryKey"`
	SeriesID           string       `json:"-"`
	EffectiveStartDate CalendarDate `json:"effective_start_date" gorm:"type:date"`
	Name               string       `json:"name"`
	LimitAmount        int64        `json:"limit_amount"`
	WalletID           *string      `json:"wallet_id"`
	CategoryID         *string      `json:"category_id"`
	CarryoverEnabled   bool         `json:"carryover_enabled"`
	CreatedAt          time.Time    `json:"-"`
}

func NextBudgetPeriod(start CalendarDate, cadence string) (CalendarDate, error) {
	if start.IsZero() {
		return CalendarDate{}, ErrBudgetCadence
	}
	var next time.Time
	switch cadence {
	case "weekly":
		if start.Weekday() != time.Monday {
			return CalendarDate{}, ErrBudgetCadence
		}
		next = start.AddDate(0, 0, 7)
	case "monthly":
		if start.Day() != 1 {
			return CalendarDate{}, ErrBudgetCadence
		}
		next = start.AddDate(0, 1, 0)
	case "quarterly":
		if start.Day() != 1 || (int(start.Month())-1)%3 != 0 {
			return CalendarDate{}, ErrBudgetCadence
		}
		next = start.AddDate(0, 3, 0)
	case "yearly":
		if start.Day() != 1 || start.Month() != time.January {
			return CalendarDate{}, ErrBudgetCadence
		}
		next = start.AddDate(1, 0, 0)
	default:
		return CalendarDate{}, ErrBudgetCadence
	}
	if next.Year() > 9999 {
		return CalendarDate{}, ErrBudgetCadence
	}
	return CalendarDate{Time: next}, nil
}

// BudgetPeriodStart returns the canonical period containing account-local now.
func BudgetPeriodStart(now time.Time, cadence string) (CalendarDate, error) {
	start := dateAt(now.Year(), now.Month(), now.Day())
	switch cadence {
	case "weekly":
		start = start.AddDate(0, 0, -(int(start.Weekday())+6)%7)
	case "monthly":
		start = dateAt(now.Year(), now.Month(), 1)
	case "quarterly":
		start = dateAt(now.Year(), time.Month((int(now.Month())-1)/3*3+1), 1)
	case "yearly":
		start = dateAt(now.Year(), 1, 1)
	default:
		return CalendarDate{}, ErrBudgetCadence
	}
	return CalendarDate{Time: start}, nil
}
