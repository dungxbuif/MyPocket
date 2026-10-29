package entity

import (
	"math/big"
	"time"
)

type InsiderCategoryTotal struct {
	CategoryID *string `json:"category_id,omitempty"`
	Name       string  `json:"name"`
	Amount     int64   `json:"amount"`
	Count      int     `json:"count"`
}

type InsiderExpenseRow struct {
	ID           string    `json:"id"`
	WalletID     string    `json:"wallet_id"`
	WalletName   string    `json:"wallet_name,omitempty"`
	CategoryID   *string   `json:"category_id,omitempty"`
	CategoryName string    `json:"category_name,omitempty"`
	Amount       int64     `json:"amount"`
	OccurredAt   time.Time `json:"occurred_at"`
	Note         *string   `json:"note,omitempty"`
}

type InsiderSummary struct {
	Month                  string                 `json:"month"`
	Timezone               string                 `json:"timezone"`
	StartAt                time.Time              `json:"start_at"`
	NextStartAt            time.Time              `json:"next_start_at"`
	IsCurrent              bool                   `json:"is_current"`
	Income                 int64                  `json:"income"`
	Expense                int64                  `json:"expense"`
	Net                    int64                  `json:"net"`
	PreviousExpense        int64                  `json:"previous_expense"`
	ExpenseDelta           int64                  `json:"expense_delta"`
	ExpenseChangeBPS       *int64                 `json:"expense_change_bps,omitempty"`
	AverageDailyExpense    int64                  `json:"average_daily_expense"`
	DaysConsidered         int                    `json:"days_considered"`
	SpendingIncomeRatioBPS *int64                 `json:"spending_income_ratio_bps,omitempty"`
	TopCategories          []InsiderCategoryTotal `json:"top_categories"`
	TopExpenses            []InsiderExpenseRow    `json:"top_expenses"`
	Estimated              bool                   `json:"estimated"`
	WalletFilter           string                 `json:"wallet_filter,omitempty"`
	CategoryFilter         string                 `json:"category_filter,omitempty"`
}

func SpendingIncomeRatioBPS(expense, income int64) *int64 {
	if expense < 0 || income <= 0 {
		return nil
	}
	numerator := new(big.Int).Mul(big.NewInt(expense), big.NewInt(10000))
	value := new(big.Int).Quo(numerator, big.NewInt(income))
	if !value.IsInt64() {
		return nil
	}
	ratio := value.Int64()
	return &ratio
}

func AverageDailySpend(expense int64, days int) int64 {
	if expense <= 0 || days <= 0 {
		return 0
	}
	return expense / int64(days)
}
