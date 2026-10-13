package entity

import "time"

// DeviceCredential stores only a digest of the device's Keychain secret.
// It is independent of any access/refresh session and never serializes secrets.
type DeviceCredential struct {
	ID                  string     `gorm:"type:uuid;primaryKey" json:"id"`
	OwnerID             string     `gorm:"type:text;not null" json:"-"`
	DeviceID            string     `gorm:"not null" json:"-"`
	DeviceName          string     `json:"device_name"`
	SecretHash          string     `json:"-"`
	PasswordFingerprint string     `json:"-"`
	ExpiresAt           time.Time  `json:"expires_at"`
	RevokedAt           *time.Time `json:"revoked_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}
