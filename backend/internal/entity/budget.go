package entity

import (
	"math"
	"time"
)

type Budget struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	OwnerID       string    `json:"owner_id"`
	Name          string    `json:"name"`
	LimitAmount   int64     `json:"limit_amount"`
	WalletID      *string   `json:"wallet_id"`
	CategoryID    *string   `json:"category_id"`
	StartAt       time.Time `json:"start_at"`
	EndAt         time.Time `json:"end_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Spent         int64     `json:"spent" gorm:"-"`
	DaysRemaining int       `json:"days_remaining" gorm:"-"`
	Ended         bool      `json:"ended" gorm:"-"`
}

type BudgetSummary struct {
	Items       []Budget `json:"items"`
	LimitAmount int64    `json:"limit_amount"`
	Spent       int64    `json:"spent"`
}

func CalculateBudgets(budgets []Budget, transactions []Transaction, categories []Category, now time.Time) BudgetSummary {
	result := BudgetSummary{Items: append([]Budget{}, budgets...)}
	parents := map[string]string{}
	for _, c := range categories {
		if c.ParentID != nil {
			parents[c.ID] = *c.ParentID
		}
	}
	counted := map[string]bool{}
	for i := range result.Items {
		budget := &result.Items[i]
		budget.Spent = 0
		budget.Ended = !now.Before(budget.EndAt)
		remainingStart := now
		if now.Before(budget.StartAt) {
			remainingStart = budget.StartAt
		}
		budget.DaysRemaining = max(0, int(math.Ceil(budget.EndAt.Sub(remainingStart).Hours()/24)))
		active := !now.Before(budget.StartAt) && now.Before(budget.EndAt)
		if active {
			result.LimitAmount += budget.LimitAmount
		}
		for _, tx := range transactions {
			if tx.Type != TransactionTypeExpense || !tx.IncludedInReports || tx.OccurredAt.Before(budget.StartAt) || !tx.OccurredAt.Before(budget.EndAt) {
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
