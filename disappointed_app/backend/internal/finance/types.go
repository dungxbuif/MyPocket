package finance

import (
	"strings"
	"time"
)

type WalletType string

const (
	WalletBasic  WalletType = "basic"
	WalletGoal   WalletType = "goal"
	WalletCredit WalletType = "credit"
)

type CreateWalletInput struct {
	ID             string
	Name           string
	Type           WalletType
	BalanceVND     int64
	IncludeInTotal *bool
	CreditLimitVND *int64
	StatementDay   *int
	PaymentDueDay  *int
	GoalTargetVND  *int64
	GoalDeadlineOn *string
}

type Wallet struct {
	ID             string     `json:"id"`
	UserID         string     `json:"user_id"`
	Name           string     `json:"name"`
	Type           WalletType `json:"type"`
	BalanceVND     int64      `json:"balance_vnd"`
	IncludeInTotal bool       `json:"include_in_total"`
	IsDefaultAI    bool       `json:"is_default_ai"`
	CreditLimitVND *int64     `json:"credit_limit_vnd,omitempty"`
	StatementDay   *int       `json:"statement_day,omitempty"`
	PaymentDueDay  *int       `json:"payment_due_day,omitempty"`
	GoalTargetVND  *int64     `json:"goal_target_vnd,omitempty"`
	GoalDeadlineOn string     `json:"goal_deadline_on,omitempty"`
	Version        int64      `json:"version"`
}

type UpdateWalletInput struct {
	BaseVersion    int64
	Name           string
	IncludeInTotal *bool
}

type CategoryKind string

const (
	CategoryExpense CategoryKind = "expense"
	CategoryIncome  CategoryKind = "income"
	CategoryDebt    CategoryKind = "debt"
)

type Category struct {
	ID        string       `json:"id"`
	UserID    string       `json:"user_id"`
	ParentID  string       `json:"parent_id,omitempty"`
	Kind      CategoryKind `json:"kind"`
	IsSystem  bool         `json:"is_system"`
	SystemKey string       `json:"system_key,omitempty"`
	Name      string       `json:"name"`
	Version   int64        `json:"version"`
}

type CreateCategoryInput struct {
	ID          string
	Kind        CategoryKind
	Name        string
	ParentID    string
	ParentDepth *int
}

type UpdateCategoryInput struct {
	Name        string
	ParentID    *string
	BaseVersion int64
}

type WalletCategorySetting struct {
	Category Category `json:"category"`
	Active   bool     `json:"active"`
}

type TransactionType string

const (
	TransactionIncome     TransactionType = "income"
	TransactionExpense    TransactionType = "expense"
	TransactionTransfer   TransactionType = "transfer"
	TransactionAdjustment TransactionType = "adjustment"
)

type AccountingInput struct {
	Type                  TransactionType
	AmountVND             int64
	SourceWalletID        string
	DestinationWalletID   string
	SourceBalanceVND      int64
	DestinationBalanceVND int64
	TargetBalanceVND      *int64
}

type AccountingEffect struct {
	SourceBalanceVND      int64
	DestinationBalanceVND int64
	SourceDeltaVND        int64
	DestinationDeltaVND   int64
}

type CreateTransactionInput struct {
	ID                  string
	IdempotencyKey      string
	Type                TransactionType
	SourceWalletID      string
	DestinationWalletID string
	CategoryID          string
	BudgetID            string
	ReceiptObjectID     string
	AmountVND           int64
	TargetBalanceVND    *int64
	OccurredAt          time.Time
	Note                string
	WithPerson          string
	EventRef            string
	ExcludedFromReports bool
}

type UpdateTransactionInput struct {
	BaseVersion         int64
	Type                TransactionType
	SourceWalletID      string
	DestinationWalletID string
	CategoryID          string
	BudgetID            string
	ReceiptObjectID     string
	AmountVND           int64
	TargetBalanceVND    *int64
	OccurredAt          time.Time
	Note                string
	WithPerson          string
	EventRef            string
	ExcludedFromReports bool
}

type TransactionFilters struct {
	WalletID             string
	CategoryID           string
	Type                 TransactionType
	DateFrom             *time.Time
	DateTo               *time.Time
	Query                string
	ExcludedFromReports  *bool
	IncludeArchivedItems bool
}

type Transaction struct {
	ID                  string          `json:"id"`
	UserID              string          `json:"user_id"`
	Type                TransactionType `json:"type"`
	SourceWalletID      string          `json:"source_wallet_id"`
	DestinationWalletID string          `json:"destination_wallet_id,omitempty"`
	CategoryID          string          `json:"category_id,omitempty"`
	BudgetID            string          `json:"budget_id,omitempty"`
	ReceiptObjectID     string          `json:"receipt_object_id,omitempty"`
	AmountVND           int64           `json:"amount_vnd"`
	BalanceAfterVND     int64           `json:"balance_after_vnd"`
	SourceDeltaVND      int64           `json:"source_delta_vnd"`
	DestinationDeltaVND int64           `json:"destination_delta_vnd"`
	OccurredAt          time.Time       `json:"occurred_at"`
	Note                string          `json:"note"`
	WithPerson          string          `json:"with_person"`
	EventRef            string          `json:"event_ref"`
	ExcludedFromReports bool            `json:"excluded_from_reports"`
	Version             int64           `json:"version"`
}

type ReceiptObject struct {
	ID               string    `json:"id"`
	UserID           string    `json:"user_id"`
	ObjectKey        string    `json:"object_key"`
	ContentType      string    `json:"content_type"`
	SizeBytes        int64     `json:"size_bytes"`
	ChecksumSHA256   string    `json:"checksum_sha256"`
	OriginalFilename string    `json:"original_filename"`
	CreatedAt        time.Time `json:"created_at"`
}

type CreateReceiptObjectInput struct {
	ObjectKey        string
	ContentType      string
	SizeBytes        int64
	ChecksumSHA256   string
	OriginalFilename string
}

func trimmed(value string) string {
	return strings.TrimSpace(value)
}
