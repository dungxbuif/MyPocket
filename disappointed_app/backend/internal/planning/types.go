package planning

import (
	"strings"
	"time"

	"mypocket/internal/finance"
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
	BaseVersion int64
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
	BaseVersion int64
	Name        string
	StartsOn    string
	EndsOn      string
	Note        string
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
	BaseVersion  int64
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

type RecurrenceFrequency string

const (
	RecurrenceDaily   RecurrenceFrequency = "daily"
	RecurrenceWeekly  RecurrenceFrequency = "weekly"
	RecurrenceMonthly RecurrenceFrequency = "monthly"
)

type RecurringPostingMode string

const (
	RecurringPostingDraft    RecurringPostingMode = "draft"
	RecurringPostingAutoPost RecurringPostingMode = "auto_post"
)

type CreateRecurringScheduleInput struct {
	Name                string
	Frequency           RecurrenceFrequency
	Timezone            string
	StartsAt            string
	EndsAt              string
	PostingMode         RecurringPostingMode
	Type                finance.TransactionType
	SourceWalletID      string
	DestinationWalletID string
	CategoryID          string
	BudgetID            string
	AmountVND           int64
	Note                string
}

type UpdateRecurringScheduleInput struct {
	BaseVersion         int64
	Name                string
	Frequency           RecurrenceFrequency
	Timezone            string
	StartsAt            string
	EndsAt              string
	PostingMode         RecurringPostingMode
	Type                finance.TransactionType
	SourceWalletID      string
	DestinationWalletID string
	CategoryID          string
	BudgetID            string
	AmountVND           int64
	Note                string
}

type RecurringSchedule struct {
	ID                  string                  `json:"id"`
	UserID              string                  `json:"user_id"`
	Name                string                  `json:"name"`
	Frequency           RecurrenceFrequency     `json:"frequency"`
	Timezone            string                  `json:"timezone"`
	StartsAt            time.Time               `json:"starts_at"`
	NextOccursAt        time.Time               `json:"next_occurs_at"`
	EndsAt              *time.Time              `json:"ends_at,omitempty"`
	PausedAt            *time.Time              `json:"paused_at,omitempty"`
	PostingMode         RecurringPostingMode    `json:"posting_mode"`
	Type                finance.TransactionType `json:"type"`
	SourceWalletID      string                  `json:"source_wallet_id"`
	DestinationWalletID string                  `json:"destination_wallet_id,omitempty"`
	CategoryID          string                  `json:"category_id,omitempty"`
	BudgetID            string                  `json:"budget_id,omitempty"`
	AmountVND           int64                   `json:"amount_vnd"`
	Note                string                  `json:"note"`
	Version             int64                   `json:"version"`
}

type TransactionDraft struct {
	ID                     string                  `json:"id"`
	UserID                 string                  `json:"user_id"`
	ScheduleID             string                  `json:"schedule_id,omitempty"`
	OccurrenceKey          string                  `json:"occurrence_key"`
	Type                   finance.TransactionType `json:"type"`
	SourceWalletID         string                  `json:"source_wallet_id"`
	DestinationWalletID    string                  `json:"destination_wallet_id,omitempty"`
	CategoryID             string                  `json:"category_id,omitempty"`
	BudgetID               string                  `json:"budget_id,omitempty"`
	AmountVND              int64                   `json:"amount_vnd"`
	OccurredAt             time.Time               `json:"occurred_at"`
	Note                   string                  `json:"note"`
	Status                 string                  `json:"status"`
	ConfirmedTransactionID string                  `json:"confirmed_transaction_id,omitempty"`
	Version                int64                   `json:"version"`
}

type ConfirmTransactionDraftInput struct {
	Version        int64
	AmountVND      int64
	Note           string
	IdempotencyKey string
}

type RejectTransactionDraftInput struct {
	Version int64
}

type TransactionDraftDecision struct {
	Draft       TransactionDraft
	Transaction *finance.Transaction
}

func trimmed(value string) string {
	return strings.TrimSpace(value)
}
