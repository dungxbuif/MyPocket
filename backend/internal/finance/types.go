package finance

import "strings"

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

func trimmed(value string) string {
	return strings.TrimSpace(value)
}
