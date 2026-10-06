package entity

import "time"

// Notification is an owner-scoped, in-app notification. Delivery channels are
// intentionally outside this read model; a worker can create rows later using
// a stable dedupe key without changing the client contract.
type Notification struct {
	ID        string     `json:"id" gorm:"primaryKey"`
	OwnerID   string     `json:"owner_id" gorm:"index;not null"`
	Kind      string     `json:"kind" gorm:"not null"`
	Title     string     `json:"title" gorm:"not null"`
	Body      string     `json:"body" gorm:"not null"`
	DedupeKey string     `json:"-" gorm:"not null"`
	ReadAt    *time.Time `json:"read_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}
