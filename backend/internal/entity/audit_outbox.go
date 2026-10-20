package entity

import "time"

// AuditOutbox is intentionally not exposed through an HTTP response. It is a
// durable, redacted hand-off for operational audit delivery only.
type AuditOutbox struct {
	ID          string    `gorm:"primaryKey"`
	Stream      string    `gorm:"not null"`
	EventID     string    `gorm:"not null"`
	Payload     []byte    `gorm:"type:jsonb;not null"`
	Attempts    int       `gorm:"not null;default:0"`
	AvailableAt time.Time `gorm:"not null"`
	LeaseUntil  *time.Time
	DeliveredAt *time.Time
	LastError   string
	CreatedAt   time.Time
}

func (AuditOutbox) TableName() string { return "audit_outbox" }
