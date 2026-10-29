package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
)

var (
	ErrAPIKeyInvalid       = errors.New("api key is invalid")
	ErrAPIKeyUnavailable   = errors.New("api key service unavailable")
	ErrAPIKeyNameInvalid   = errors.New("api key name is invalid")
	ErrAPIKeyExpiryInvalid = errors.New("api key expiry is invalid")
)

type UserAPIKeyService struct {
	Keys repository.UserAPIKeyRepository
	Now  func() time.Time
}

type APIKeyCreateInput struct {
	OwnerID   string
	Name      string
	Scopes    []string // Deprecated input ignored; all keys have full same-owner access.
	ExpiresAt *time.Time
}

type APIKeyCreated struct {
	Key    entity.UserAPIKey
	Secret string
}

func NewUserAPIKeyService(keys repository.UserAPIKeyRepository) *UserAPIKeyService {
	return &UserAPIKeyService{Keys: keys, Now: func() time.Time { return time.Now().UTC() }}
}

func (s *UserAPIKeyService) Create(ctx context.Context, input APIKeyCreateInput) (APIKeyCreated, error) {
	ownerID := strings.TrimSpace(input.OwnerID)
	if s == nil || s.Keys == nil || ownerID == "" {
		return APIKeyCreated{}, ErrAPIKeyUnavailable
	}
	name := strings.TrimSpace(input.Name)
	if name == "" || len(name) > 80 {
		return APIKeyCreated{}, ErrAPIKeyNameInvalid
	}
	now := s.now()
	if input.ExpiresAt != nil && !input.ExpiresAt.After(now) {
		return APIKeyCreated{}, ErrAPIKeyExpiryInvalid
	}
	lookupBytes := make([]byte, 10)
	secretBytes := make([]byte, 32)
	if _, err := rand.Read(lookupBytes); err != nil {
		return APIKeyCreated{}, ErrAPIKeyUnavailable
	}
	if _, err := rand.Read(secretBytes); err != nil {
		return APIKeyCreated{}, ErrAPIKeyUnavailable
	}
	lookup := base64.RawURLEncoding.EncodeToString(lookupBytes)
	secret := "mpk_" + lookup + "." + base64.RawURLEncoding.EncodeToString(secretBytes)
	key := entity.UserAPIKey{ID: uuid.NewString(), OwnerID: ownerID, LookupID: lookup, Name: name, SecretHash: hashAPIKey(secret), Scopes: []string{}, ExpiresAt: input.ExpiresAt, CreatedAt: now, UpdatedAt: now}
	if err := s.Keys.Create(ctx, &key); err != nil {
		return APIKeyCreated{}, err
	}
	return APIKeyCreated{Key: key, Secret: secret}, nil
}

// The final legacy argument is ignored; there is no scope authorization model.
func (s *UserAPIKeyService) Authenticate(ctx context.Context, secret string, _ []string) (Principal, error) {
	if s == nil || s.Keys == nil {
		return Principal{}, ErrAPIKeyUnavailable
	}
	secret = strings.TrimSpace(secret)
	if !strings.HasPrefix(secret, "mpk_") {
		return Principal{}, ErrAPIKeyInvalid
	}
	body := strings.TrimPrefix(secret, "mpk_")
	parts := strings.SplitN(body, ".", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" || len(parts[1]) < 40 {
		return Principal{}, ErrAPIKeyInvalid
	}
	key, err := s.Keys.FindByLookup(ctx, parts[0])
	if err != nil || key == nil || strings.TrimSpace(key.OwnerID) == "" || subtle.ConstantTimeCompare([]byte(key.SecretHash), []byte(hashAPIKey(secret))) != 1 || !key.ActiveAt(s.now()) {
		return Principal{}, ErrAPIKeyInvalid
	}
	_ = s.Keys.TouchLastUsed(ctx, key.ID, s.now())
	return Principal{OwnerID: key.OwnerID, CredentialID: key.ID, CredentialKind: "user_api_key", ExpiresAt: valueOrZero(key.ExpiresAt)}, nil
}

// ValidatePrincipal re-reads the credential so long-running advisor work does
// not continue after an API key has been revoked or expired.
func (s *UserAPIKeyService) ValidatePrincipal(ctx context.Context, principal Principal) error {
	if s == nil || s.Keys == nil || principal.CredentialKind != "user_api_key" || strings.TrimSpace(principal.CredentialID) == "" || strings.TrimSpace(principal.OwnerID) == "" {
		return ErrAPIKeyInvalid
	}
	key, err := s.Keys.FindByID(ctx, principal.CredentialID)
	if err != nil || key == nil || key.OwnerID != principal.OwnerID || !key.ActiveAt(s.now()) {
		return ErrAPIKeyInvalid
	}
	return nil
}

func (s *UserAPIKeyService) List(ctx context.Context, ownerID string) ([]entity.UserAPIKey, error) {
	if s == nil || s.Keys == nil || strings.TrimSpace(ownerID) == "" {
		return nil, ErrAPIKeyUnavailable
	}
	keys, err := s.Keys.ListByOwner(ctx, ownerID)
	for index := range keys {
		keys[index].Scopes = []string{}
	}
	return keys, err
}

func (s *UserAPIKeyService) Revoke(ctx context.Context, ownerID, id string) error {
	if s == nil || s.Keys == nil || strings.TrimSpace(ownerID) == "" || strings.TrimSpace(id) == "" {
		return ErrAPIKeyUnavailable
	}
	return s.Keys.Revoke(ctx, ownerID, id, s.now())
}

func (s *UserAPIKeyService) now() time.Time {
	if s.Now != nil {
		return s.Now().UTC()
	}
	return time.Now().UTC()
}

func hashAPIKey(value string) string {
	digest := sha256.Sum256([]byte(value))
	return base64.RawURLEncoding.EncodeToString(digest[:])
}

func valueOrZero(value *time.Time) time.Time {
	if value == nil {
		return time.Time{}
	}
	return *value
}
