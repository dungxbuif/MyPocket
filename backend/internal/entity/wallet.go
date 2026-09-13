package entity

import "time"

const (
	WalletTypeBasic   = "basic"
	WalletTypeGoal    = "goal"
	WalletTypeCredit  = "credit"
	WalletCurrencyVND = "VND"
)

type Wallet struct {
	ID                   string     `json:"id" gorm:"primaryKey"`
	OwnerID              string     `json:"owner_id" gorm:"index;not null"`
	Name                 string     `json:"name" gorm:"not null"`
	Type                 string     `json:"type" gorm:"not null"`
	Currency             string     `json:"currency" gorm:"not null;default:VND"`
	OpeningBalance       int64      `json:"opening_balance"`
	IsInTotal            bool       `json:"is_in_total" gorm:"not null;default:true"`
	Description          *string    `json:"description,omitempty"`
	TargetAmount         *int64     `json:"target_amount,omitempty"`
	TargetDate           *time.Time `json:"target_date,omitempty"`
	CreditLimit          *int64     `json:"credit_limit,omitempty"`
	LastStatementBalance *int64     `json:"last_statement_balance,omitempty"`
	StatementDay         *int       `json:"statement_day,omitempty"`
	PaymentDueDay        *int       `json:"payment_due_day,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type Category struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	OwnerID   *string   `json:"owner_id,omitempty" gorm:"index"`
	ParentID  *string   `json:"parent_id,omitempty" gorm:"index"`
	Kind      string    `json:"kind" gorm:"not null"`
	Name      string    `json:"name" gorm:"not null"`
	SystemKey *string   `json:"system_key,omitempty" gorm:"uniqueIndex"`
	IsSystem  bool      `json:"is_system" gorm:"not null;default:false"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CategoryWallet struct {
	CategoryID string    `json:"category_id" gorm:"primaryKey"`
	WalletID   string    `json:"wallet_id" gorm:"primaryKey"`
	CreatedAt  time.Time `json:"created_at"`
}
