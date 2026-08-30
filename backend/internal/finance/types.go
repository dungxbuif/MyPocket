package finance

import (
	"strings"
	"time"
)

type WalletType string

const (
	WalletCash    WalletType = "cash"
	WalletBank    WalletType = "bank"
	WalletCredit  WalletType = "credit"
	WalletEWallet WalletType = "e_wallet"
	WalletSavings WalletType = "savings"
	WalletDebt    WalletType = "debt"
)

type CreateWalletInput struct {
	Name           string
	Type           WalletType
	CreditLimitVND *int64
	StatementDay   *int
	PaymentDueDay  *int
}

type Wallet struct {
	ID             string
	UserID         string
	Name           string
	Type           WalletType
	BalanceVND     int64
	IncludeInTotal bool
	IsDefaultAI    bool
	Version        int64
}

type UpdateWalletInput struct {
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
	ID        string
	UserID    string
	ParentID  string
	Kind      CategoryKind
	IsSystem  bool
	SystemKey string
	Name      string
}

type CreateCategoryInput struct {
	Kind        CategoryKind
	Name        string
	ParentDepth *int
}

type UpdateCategoryInput struct {
	Name string
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
	IdempotencyKey      string
	Type                TransactionType
	SourceWalletID      string
	DestinationWalletID string
	CategoryID          string
	AmountVND           int64
	TargetBalanceVND    *int64
	OccurredAt          time.Time
	Note                string
	WithPerson          string
	EventRef            string
	ExcludedFromReports bool
}

type UpdateTransactionInput struct {
	Type                TransactionType
	SourceWalletID      string
	DestinationWalletID string
	CategoryID          string
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
