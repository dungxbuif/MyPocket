package entity

import "time"

// MarketProduct is a backend-configured catalog identity. Provider key and
// external product ID, not display name, define the source identity.
type MarketProduct struct {
	ID                string    `json:"id" gorm:"primaryKey"`
	ProviderKey       string    `json:"provider_key" gorm:"not null;index"`
	ExternalProductID string    `json:"external_product_id" gorm:"not null"`
	Name              string    `json:"name" gorm:"not null"`
	Kind              string    `json:"kind" gorm:"not null"`
	Unit              string    `json:"unit" gorm:"not null"`
	Currency          string    `json:"currency" gorm:"not null;default:VND"`
	ValuationSide     string    `json:"valuation_side" gorm:"not null;default:buyback"`
	Enabled           bool      `json:"enabled" gorm:"not null;default:false"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// MarketQuote is the latest accepted quote per catalog product. ObservedAt is
// source time; FetchedAt is transport time and is never used as price time.
type MarketQuote struct {
	ProductID  string    `json:"product_id" gorm:"primaryKey"`
	Price      int64     `json:"price" gorm:"not null"`
	ObservedAt time.Time `json:"observed_at" gorm:"not null"`
	FetchedAt  time.Time `json:"fetched_at" gorm:"not null"`
}

type MarketPollClaim struct {
	ProviderKey string    `json:"provider_key" gorm:"primaryKey;column:provider_key"`
	SlotAt      time.Time `json:"slot_at" gorm:"primaryKey;column:slot_at"`
	LeaseToken  string    `json:"-" gorm:"not null"`
	LeaseUntil  time.Time `json:"-" gorm:"not null"`
	Status      string    `json:"status" gorm:"not null"`
	Error       *string   `json:"error,omitempty" gorm:"column:error_message"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`
}
