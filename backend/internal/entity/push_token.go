package entity

import "time"

// PushToken is an owner-scoped provider address. Token is deliberately omitted
// from JSON and logs; only a future delivery adapter may read it server-side.
type PushToken struct {
	ID          string     `gorm:"type:uuid;primaryKey" json:"id"`
	OwnerID     string     `gorm:"not null;index" json:"-"`
	DeviceID    string     `gorm:"not null" json:"device_id"`
	DeviceName  string     `gorm:"not null" json:"device_name"`
	Platform    string     `gorm:"not null" json:"platform"`
	Environment string     `gorm:"not null" json:"environment"`
	Token       string     `gorm:"not null" json:"-"`
	RevokedAt   *time.Time `json:"revoked_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
