package entity

import (
	"errors"
	"time"
)

// Repository input validation, unrelated to removed per-operation scopes.
var ErrAPIKeyInvalid = errors.New("api key is invalid")

// Deprecated metadata labels, retained only for existing stored records/clients.
// None of these values grants or restricts authenticated user operations.
const (
	APIKeyScopeFinanceRead   = "finance:read"
	APIKeyScopeAdvisorRead   = "advisor:read"
	APIKeyScopeAdvisorChat   = "advisor:chat"
	APIKeyScopeFeedbackRead  = "feedback:read"
	APIKeyScopeFeedbackWrite = "feedback:write"
)

type UserAPIKey struct {
	ID         string     `json:"id" gorm:"primaryKey"`
	OwnerID    string     `json:"owner_id" gorm:"not null;index"`
	LookupID   string     `json:"-" gorm:"not null;uniqueIndex"`
	Name       string     `json:"name" gorm:"not null"`
	SecretHash string     `json:"-" gorm:"not null"`
	Scopes     []string   `json:"scopes" gorm:"serializer:json;type:jsonb"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

func (UserAPIKey) TableName() string { return "user_api_keys" }

func (k UserAPIKey) ActiveAt(now time.Time) bool {
	if k.RevokedAt != nil && !k.RevokedAt.After(now) {
		return false
	}
	return k.ExpiresAt == nil || k.ExpiresAt.After(now)
}
