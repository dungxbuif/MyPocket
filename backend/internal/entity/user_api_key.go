package entity

import (
	"errors"
	"strings"
	"time"
)

const (
	APIKeyScopeFinanceRead = "finance:read"
	APIKeyScopeAdvisorRead = "advisor:read"
	APIKeyScopeAdvisorChat = "advisor:chat"
)

var ErrAPIKeyInvalid = errors.New("api key is invalid")

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

func NormalizeAPIKeyScopes(scopes []string) ([]string, error) {
	seen := make(map[string]struct{}, len(scopes)+2)
	result := make([]string, 0, len(scopes)+2)
	for _, scope := range scopes {
		scope = strings.TrimSpace(scope)
		if scope == "" {
			continue
		}
		switch scope {
		case APIKeyScopeFinanceRead, APIKeyScopeAdvisorRead, APIKeyScopeAdvisorChat:
		default:
			return nil, ErrAPIKeyInvalid
		}
		if _, ok := seen[scope]; ok {
			continue
		}
		seen[scope] = struct{}{}
		result = append(result, scope)
	}
	if containsScope(result, APIKeyScopeAdvisorChat) {
		if _, ok := seen[APIKeyScopeAdvisorRead]; !ok {
			result = append(result, APIKeyScopeAdvisorRead)
		}
		if _, ok := seen[APIKeyScopeFinanceRead]; !ok {
			result = append(result, APIKeyScopeFinanceRead)
		}
	}
	if len(result) == 0 {
		return nil, ErrAPIKeyInvalid
	}
	return result, nil
}

func (k UserAPIKey) HasScope(scope string) bool { return containsScope(k.Scopes, scope) }

func (k UserAPIKey) ActiveAt(now time.Time) bool {
	if k.RevokedAt != nil && !k.RevokedAt.After(now) {
		return false
	}
	return k.ExpiresAt == nil || k.ExpiresAt.After(now)
}

func containsScope(scopes []string, wanted string) bool {
	for _, scope := range scopes {
		if scope == wanted {
			return true
		}
	}
	return false
}
