package entity

import "time"

// TransactionMutation stores the durable replay envelope for a financial
// mutation. ResponseJSON is internal state and is never exposed by API routes.
type TransactionMutation struct {
	ID             string    `json:"-" gorm:"primaryKey"`
	OwnerID        string    `json:"-" gorm:"index;not null"`
	Operation      string    `json:"-" gorm:"not null"`
	IdempotencyKey string    `json:"-" gorm:"not null"`
	RequestHash    string    `json:"-" gorm:"not null"`
	Status         string    `json:"-" gorm:"not null"`
	ResponseJSON   string    `json:"-" gorm:"type:jsonb"`
	CreatedAt      time.Time `json:"-"`
	UpdatedAt      time.Time `json:"-"`
}
