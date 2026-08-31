package planning

import (
	"strings"
	"time"
)

type BudgetPeriodType string

const (
	BudgetWeekly    BudgetPeriodType = "weekly"
	BudgetMonthly   BudgetPeriodType = "monthly"
	BudgetQuarterly BudgetPeriodType = "quarterly"
	BudgetYearly    BudgetPeriodType = "yearly"
	BudgetCustom    BudgetPeriodType = "custom"
)

type CreateBudgetInput struct {
	Name        string
	PeriodType  BudgetPeriodType
	AmountVND   int64
	CategoryIDs []string
	CustomStart *time.Time
	CustomEnd   *time.Time
}

type UpdateBudgetInput struct {
	Name        string
	PeriodType  BudgetPeriodType
	AmountVND   int64
	CategoryIDs []string
	CustomStart *time.Time
	CustomEnd   *time.Time
}

type Budget struct {
	ID            string           `json:"id"`
	UserID        string           `json:"user_id"`
	Name          string           `json:"name"`
	PeriodType    BudgetPeriodType `json:"period_type"`
	AmountVND     int64            `json:"amount_vnd"`
	CategoryIDs   []string         `json:"category_ids"`
	AllCategories bool             `json:"all_categories"`
	CustomStart   string           `json:"custom_start,omitempty"`
	CustomEnd     string           `json:"custom_end,omitempty"`
	Version       int64            `json:"version"`
}

type BudgetProgress struct {
	Budget       Budget `json:"budget"`
	PeriodStart  string `json:"period_start"`
	PeriodEnd    string `json:"period_end"`
	SpentVND     int64  `json:"spent_vnd"`
	RemainingVND int64  `json:"remaining_vnd"`
	Percent      int64  `json:"percent"`
	Alert80      bool   `json:"alert_80"`
	Alert100     bool   `json:"alert_100"`
}

func trimmed(value string) string {
	return strings.TrimSpace(value)
}
