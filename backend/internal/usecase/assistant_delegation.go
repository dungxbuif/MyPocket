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

var ErrAssistantDelegationInvalid = errors.New("assistant delegation is invalid")

var DefaultAssistantOperations = []string{
	"GET /api/v1/home", "GET /api/v1/wallets", "GET /api/v1/transactions",
	"GET /api/v1/budgets", "GET /api/v1/jars", "GET /api/v1/categories",
}

type AssistantDelegationService struct {
	Delegations repository.AssistantDelegationRepository
	Now         func() time.Time
}

type AssistantDelegationCreated struct {
	Delegation entity.AssistantDelegation
	Secret     string
}

func NewAssistantDelegationService(repo repository.AssistantDelegationRepository) *AssistantDelegationService {
	return &AssistantDelegationService{Delegations: repo, Now: func() time.Time { return time.Now().UTC() }}
}

func (s *AssistantDelegationService) Create(ctx context.Context, ownerID string, ttl time.Duration) (AssistantDelegationCreated, error) {
	if s == nil || s.Delegations == nil || strings.TrimSpace(ownerID) == "" || ttl <= 0 {
		return AssistantDelegationCreated{}, ErrAssistantDelegationInvalid
	}
	lookupBytes, secretBytes := make([]byte, 10), make([]byte, 32)
	if _, err := rand.Read(lookupBytes); err != nil {
		return AssistantDelegationCreated{}, err
	}
	if _, err := rand.Read(secretBytes); err != nil {
		return AssistantDelegationCreated{}, err
	}
	lookup := base64.RawURLEncoding.EncodeToString(lookupBytes)
	secret := "mpa_" + lookup + "." + base64.RawURLEncoding.EncodeToString(secretBytes)
	now := s.now()
	expires := now.Add(ttl)
	value := entity.AssistantDelegation{ID: uuid.NewString(), OwnerID: ownerID, LookupID: lookup, SecretHash: hashAssistantSecret(secret), AllowedOperations: append([]string(nil), DefaultAssistantOperations...), ExpiresAt: &expires, CreatedAt: now, UpdatedAt: now}
	if err := s.Delegations.Create(ctx, &value); err != nil {
		return AssistantDelegationCreated{}, err
	}
	return AssistantDelegationCreated{Delegation: value, Secret: secret}, nil
}

func (s *AssistantDelegationService) Authenticate(ctx context.Context, secret, method, path string) (Principal, error) {
	if s == nil || s.Delegations == nil || !strings.HasPrefix(secret, "mpa_") {
		return Principal{}, ErrAssistantDelegationInvalid
	}
	parts := strings.SplitN(strings.TrimPrefix(secret, "mpa_"), ".", 2)
	if len(parts) != 2 || len(parts[1]) < 40 {
		return Principal{}, ErrAssistantDelegationInvalid
	}
	value, err := s.Delegations.FindByLookup(ctx, parts[0])
	if err != nil || value == nil || subtle.ConstantTimeCompare([]byte(value.SecretHash), []byte(hashAssistantSecret(secret))) != 1 || !value.ActiveAt(s.now()) {
		return Principal{}, ErrAssistantDelegationInvalid
	}
	wanted := strings.ToUpper(strings.TrimSpace(method)) + " " + strings.TrimSpace(path)
	allowed := false
	for _, operation := range value.AllowedOperations {
		if operation == wanted {
			allowed = true
			break
		}
	}
	if !allowed {
		return Principal{}, ErrAssistantDelegationInvalid
	}
	_ = s.Delegations.TouchLastUsed(ctx, value.ID, s.now())
	return Principal{OwnerID: value.OwnerID, CredentialID: value.ID, CredentialKind: "assistant_delegation", ExpiresAt: valueOrZero(value.ExpiresAt)}, nil
}

func (s *AssistantDelegationService) Revoke(ctx context.Context, ownerID, id string) error {
	return s.Delegations.Revoke(ctx, ownerID, id, s.now())
}
func (s *AssistantDelegationService) now() time.Time {
	if s.Now != nil {
		return s.Now().UTC()
	}
	return time.Now().UTC()
}
func hashAssistantSecret(value string) string {
	digest := sha256.Sum256([]byte(value))
	return base64.RawURLEncoding.EncodeToString(digest[:])
}
