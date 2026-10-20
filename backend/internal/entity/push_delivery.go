package entity

import "time"

// PushDelivery is the durable, owner-scoped hand-off from an in-app
// notification to a provider token. Payload must contain only an app route and
// notification ID; financial amounts and lock-screen text are intentionally
// excluded by the producer contract.
type PushDelivery struct {
	ID             string     `gorm:"type:uuid;primaryKey" json:"id"`
	OwnerID        string     `gorm:"not null;index" json:"-"`
	NotificationID string     `gorm:"not null" json:"notification_id"`
	PushTokenID    string     `gorm:"type:uuid;not null" json:"push_token_id"`
	Environment    string     `gorm:"not null" json:"environment"`
	CollapseID     string     `gorm:"not null" json:"collapse_id"`
	Payload        []byte     `gorm:"type:jsonb;not null" json:"-"`
	Status         string     `gorm:"not null" json:"status"`
	Attempts       int        `gorm:"not null;default:0" json:"attempts"`
	AvailableAt    time.Time  `gorm:"not null" json:"available_at"`
	LeaseUntil     *time.Time `json:"-"`
	LastError      string     `json:"-"`
	DeliveredAt    *time.Time `json:"delivered_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (PushDelivery) TableName() string { return "push_deliveries" }
