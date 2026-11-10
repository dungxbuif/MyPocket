package entity

import "time"

type AssistantDelegation struct {
	ID                string     `json:"id" gorm:"primaryKey"`
	OwnerID           string     `json:"owner_id" gorm:"not null;index"`
	LookupID          string     `json:"-" gorm:"not null;uniqueIndex"`
	SecretHash        string     `json:"-" gorm:"not null"`
	AllowedOperations []string   `json:"allowed_operations" gorm:"serializer:json;type:jsonb"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty"`
	RevokedAt         *time.Time `json:"revoked_at,omitempty"`
	LastUsedAt        *time.Time `json:"last_used_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

func (AssistantDelegation) TableName() string { return "assistant_delegations" }

func (d AssistantDelegation) ActiveAt(now time.Time) bool {
	if d.RevokedAt != nil && !d.RevokedAt.After(now) {
		return false
	}
	return d.ExpiresAt == nil || d.ExpiresAt.After(now)
}
