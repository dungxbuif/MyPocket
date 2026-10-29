package entity

import (
	"errors"
	"math"
	"sort"
	"time"
)

type Budget struct {
	ID                   string       `json:"id" gorm:"primaryKey"`
	OwnerID              string       `json:"owner_id"`
	Name                 string       `json:"name"`
	LimitAmount          int64        `json:"limit_amount"`
	SeriesID             *string      `json:"series_id,omitempty"`
	BaseLimitAmount      int64        `json:"base_limit_amount" gorm:"-"`
	CarryoverAmount      int64        `json:"carryover_amount" gorm:"-"`
	EffectiveLimitAmount int64        `json:"effective_limit_amount" gorm:"-"`
	CarryoverEnabled     bool         `json:"carryover_enabled"`
	WalletID             *string      `json:"wallet_id"`
	CategoryID           *string      `json:"category_id"`
	StartDate            CalendarDate `json:"start_date" gorm:"column:start_date;type:date;not null"`
	EndDate              CalendarDate `json:"end_date" gorm:"column:end_date;type:date;not null"`
	StartAt              time.Time    `json:"-" gorm:"-"`
	EndAt                time.Time    `json:"-" gorm:"-"`
	CreatedAt            time.Time    `json:"created_at"`
	UpdatedAt            time.Time    `json:"updated_at"`
	Spent                int64        `json:"spent" gorm:"-"`
	DaysRemaining        int          `json:"days_remaining" gorm:"-"`
	Ended                bool         `json:"ended" gorm:"-"`
}

type BudgetSummary struct {
	Items       []Budget `json:"items"`
	LimitAmount int64    `json:"limit_amount"`
	Spent       int64    `json:"spent"`
}

func CalculateBudgets(budgets []Budget, transactions []Transaction, categories []Category, now time.Time) BudgetSummary {
	return CalculateBudgetsInLocation(budgets, transactions, categories, now, time.UTC)
}

func CalculateBudgetsInLocation(budgets []Budget, transactions []Transaction, categories []Category, now time.Time, location *time.Location) BudgetSummary {
	result, _ := calculateBudgets(budgets, transactions, categories, now, location, false)
	return result
}

func CalculateBudgetsChecked(budgets []Budget, transactions []Transaction, categories []Category, now time.Time, location *time.Location) (BudgetSummary, error) {
	return calculateBudgets(budgets, transactions, categories, now, location, true)
}

func calculateBudgets(budgets []Budget, transactions []Transaction, categories []Category, now time.Time, location *time.Location, strictAsOf bool) (BudgetSummary, error) {
	if location == nil {
		location = time.UTC
	}
	result := BudgetSummary{Items: append([]Budget{}, budgets...)}
	counted := map[string]bool{}
	today := CalendarDate{Time: dateAt(now.In(location).Year(), now.In(location).Month(), now.In(location).Day())}
	for i := range result.Items {
		budget := &result.Items[i]
		startDate, endDate := budget.StartDate, budget.EndDate
		if startDate.IsZero() && !budget.StartAt.IsZero() {
			startDate = CalendarDate{Time: dateAt(budget.StartAt.UTC().Year(), budget.StartAt.UTC().Month(), budget.StartAt.UTC().Day())}
		}
		if endDate.IsZero() && !budget.EndAt.IsZero() {
			endDate = CalendarDate{Time: dateAt(budget.EndAt.UTC().Add(-time.Microsecond).Year(), budget.EndAt.UTC().Add(-time.Microsecond).Month(), budget.EndAt.UTC().Add(-time.Microsecond).Day())}
		}
		budget.Spent = 0
		budget.Ended = today.After(endDate.Time)
		remainingStart := today.Time
		if today.Before(startDate.Time) {
			remainingStart = startDate.Time
		}
		budget.DaysRemaining = max(0, int(math.Ceil(endDate.Time.Sub(remainingStart).Hours()/24))+1)
		active := !today.Before(startDate.Time) && !today.After(endDate.Time)
		budget.StartDate, budget.EndDate = startDate, endDate
		cutoff := now
		if !strictAsOf {
			cutoff = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)
		}
		for _, tx := range MatchBudgetTransactions(*budget, transactions, categories, cutoff, location) {
			if tx.Amount < 0 || budget.Spent > MaxBudgetMoney-tx.Amount {
				return BudgetSummary{}, ErrBudgetMoneyOverflow
			}
			budget.Spent += tx.Amount
			if active && !counted[tx.ID] {
				if result.Spent > MaxBudgetMoney-tx.Amount {
					return BudgetSummary{}, ErrBudgetMoneyOverflow
				}
				result.Spent += tx.Amount
				counted[tx.ID] = true
			}
		}
	}
	indices := make([]int, len(result.Items))
	for i := range indices {
		indices[i] = i
	}
	sort.SliceStable(indices, func(i, j int) bool {
		return result.Items[indices[i]].StartDate.Before(result.Items[indices[j]].StartDate.Time)
	})
	previous := map[string]Budget{}
	for _, i := range indices {
		b := &result.Items[i]
		b.BaseLimitAmount = b.LimitAmount
		b.CarryoverAmount = 0
		if b.LimitAmount < 0 || b.LimitAmount > MaxBudgetMoney {
			return BudgetSummary{}, ErrBudgetMoneyOverflow
		}
		if b.SeriesID != nil {
			if prev, ok := previous[*b.SeriesID]; ok && b.CarryoverEnabled && sameBudgetScope(prev.WalletID, b.WalletID) && sameBudgetScope(prev.CategoryID, b.CategoryID) && prev.EndDate.AddDate(0, 0, 1).Equal(b.StartDate.Time) {
				b.CarryoverAmount = max(int64(0), prev.EffectiveLimitAmount-prev.Spent)
			}
		}
		if b.LimitAmount > MaxBudgetMoney-b.CarryoverAmount {
			return BudgetSummary{}, ErrBudgetMoneyOverflow
		}
		b.EffectiveLimitAmount = b.LimitAmount + b.CarryoverAmount
		if b.SeriesID != nil {
			previous[*b.SeriesID] = *b
		}
		if !today.Before(b.StartDate.Time) && !today.After(b.EndDate.Time) {
			if result.LimitAmount > MaxBudgetMoney-b.EffectiveLimitAmount {
				return BudgetSummary{}, ErrBudgetMoneyOverflow
			}
			result.LimitAmount += b.EffectiveLimitAmount
		}
	}
	return result, nil
}

const MaxBudgetMoney int64 = 9007199254740991

var ErrBudgetMoneyOverflow = errors.New("budget money exceeds supported range")

func sameBudgetScope(a, b *string) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

func dateAt(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
