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

type CreateEventInput struct {
	Name     string
	StartsOn string
	EndsOn   string
	Note     string
}

type UpdateEventInput struct {
	Name     string
	StartsOn string
	EndsOn   string
	Note     string
}

type EventSummary struct {
	ID               string `json:"id"`
	UserID           string `json:"user_id"`
	Name             string `json:"name"`
	StartsOn         string `json:"starts_on"`
	EndsOn           string `json:"ends_on"`
	Note             string `json:"note"`
	TotalVND         int64  `json:"total_vnd"`
	TransactionCount int64  `json:"transaction_count"`
	Version          int64  `json:"version"`
}

type ObligationDirection string

const (
	ObligationBorrowed ObligationDirection = "borrowed"
	ObligationLent     ObligationDirection = "lent"
)

type CreateObligationInput struct {
	Direction    ObligationDirection
	PrincipalVND int64
	Counterparty string
	DueOn        string
	Note         string
}

type UpdateObligationInput struct {
	Direction    ObligationDirection
	PrincipalVND int64
	Counterparty string
	DueOn        string
	Note         string
}

type ObligationSummary struct {
	ID           string              `json:"id"`
	UserID       string              `json:"user_id"`
	Direction    ObligationDirection `json:"direction"`
	PrincipalVND int64               `json:"principal_vnd"`
	Counterparty string              `json:"counterparty"`
	DueOn        string              `json:"due_on"`
	Note         string              `json:"note"`
	RepaidVND    int64               `json:"repaid_vnd"`
	RemainingVND int64               `json:"remaining_vnd"`
	Version      int64               `json:"version"`
}

func trimmed(value string) string {
	return strings.TrimSpace(value)
}
