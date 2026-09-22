package entity

import (
	"math"
	"time"
)

type Budget struct {
	ID            string       `json:"id" gorm:"primaryKey"`
	OwnerID       string       `json:"owner_id"`
	Name          string       `json:"name"`
	LimitAmount   int64        `json:"limit_amount"`
	WalletID      *string      `json:"wallet_id"`
	CategoryID    *string      `json:"category_id"`
	StartDate     CalendarDate `json:"start_date" gorm:"column:start_date;type:date;not null"`
	EndDate       CalendarDate `json:"end_date" gorm:"column:end_date;type:date;not null"`
	StartAt       time.Time    `json:"-" gorm:"-"`
	EndAt         time.Time    `json:"-" gorm:"-"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
	Spent         int64        `json:"spent" gorm:"-"`
	DaysRemaining int          `json:"days_remaining" gorm:"-"`
	Ended         bool         `json:"ended" gorm:"-"`
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
	if location == nil {
		location = time.UTC
	}
	result := BudgetSummary{Items: append([]Budget{}, budgets...)}
	parents := map[string]string{}
	for _, c := range categories {
		if c.ParentID != nil {
			parents[c.ID] = *c.ParentID
		}
	}
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
		if active {
			result.LimitAmount += budget.LimitAmount
		}
		for _, tx := range transactions {
			txLocal := tx.OccurredAt.In(location)
			txDate := CalendarDate{Time: dateAt(txLocal.Year(), txLocal.Month(), txLocal.Day())}
			if tx.Type != TransactionTypeExpense || !tx.IncludedInReports || txDate.Before(startDate.Time) || txDate.After(endDate.Time) {
				continue
			}
			if budget.WalletID != nil && tx.WalletID != *budget.WalletID {
				continue
			}
			if budget.CategoryID != nil {
				if tx.CategoryID == nil {
					continue
				}
				id := *tx.CategoryID
				seen := map[string]bool{}
				for id != "" && id != *budget.CategoryID && !seen[id] {
					seen[id] = true
					id = parents[id]
				}
				if id != *budget.CategoryID {
					continue
				}
			}
			budget.Spent += tx.Amount
			if active && !counted[tx.ID] {
				result.Spent += tx.Amount
				counted[tx.ID] = true
			}
		}
	}
	return result
}

func dateAt(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
