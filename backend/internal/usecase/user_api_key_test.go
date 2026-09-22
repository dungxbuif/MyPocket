package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mypocket/backend/internal/entity"
)

type apiKeyRepoStub struct {
	key      *entity.UserAPIKey
	created  *entity.UserAPIKey
	touched  bool
	revoked  bool
	ownerIDs []string
}

func (r *apiKeyRepoStub) Create(_ context.Context, key *entity.UserAPIKey) error {
	r.created = key
	r.key = key
	return nil
}
func (r *apiKeyRepoStub) FindByLookup(_ context.Context, lookup string) (*entity.UserAPIKey, error) {
	if r.key == nil || r.key.LookupID != lookup {
		return nil, errors.New("not found")
	}
	return r.key, nil
}
func (r *apiKeyRepoStub) FindByID(_ context.Context, id string) (*entity.UserAPIKey, error) {
	if r.key == nil || r.key.ID != id {
		return nil, errors.New("not found")
	}
	return r.key, nil
}
func (r *apiKeyRepoStub) ListByOwner(_ context.Context, owner string) ([]entity.UserAPIKey, error) {
	r.ownerIDs = append(r.ownerIDs, owner)
	if r.key == nil || r.key.OwnerID != owner {
		return nil, nil
	}
	return []entity.UserAPIKey{*r.key}, nil
}
func (r *apiKeyRepoStub) Revoke(_ context.Context, owner, id string, now time.Time) error {
	if r.key == nil || r.key.OwnerID != owner || r.key.ID != id {
		return errors.New("not found")
	}
	r.revoked = true
	r.key.RevokedAt = &now
	return nil
}
func (r *apiKeyRepoStub) TouchLastUsed(_ context.Context, _ string, _ time.Time) error {
	r.touched = true
	return nil
}

func TestUserAPIKeyCreateReturnsSecretButStoresOnlyDigest(t *testing.T) {
	repo := &apiKeyRepoStub{}
	service := NewUserAPIKeyService(repo)
	service.Now = func() time.Time { return time.Date(2026, 9, 22, 8, 0, 0, 0, time.UTC) }
	created, err := service.Create(context.Background(), APIKeyCreateInput{OwnerID: "owner-1", Name: "Mobile", Scopes: []string{entity.APIKeyScopeAdvisorChat}})
	if err != nil {
		t.Fatal(err)
	}
	if created.Secret == "" || repo.created == nil || repo.created.SecretHash == created.Secret || repo.created.SecretHash == "" {
		t.Fatal("secret must be returned once and stored only as a digest")
	}
	if !repo.created.HasScope(entity.APIKeyScopeFinanceRead) || !repo.created.HasScope(entity.APIKeyScopeAdvisorRead) {
		t.Fatalf("advisor chat key missing implied read scopes: %v", repo.created.Scopes)
	}
}

func TestUserAPIKeyAuthenticateChecksOwnerIndependentScopesAndRevocation(t *testing.T) {
	repo := &apiKeyRepoStub{}
	service := NewUserAPIKeyService(repo)
	service.Now = func() time.Time { return time.Date(2026, 9, 22, 8, 0, 0, 0, time.UTC) }
	created, err := service.Create(context.Background(), APIKeyCreateInput{OwnerID: "owner-1", Name: "Read", Scopes: []string{entity.APIKeyScopeFinanceRead}})
	if err != nil {
		t.Fatal(err)
	}
	principal, err := service.Authenticate(context.Background(), created.Secret, []string{entity.APIKeyScopeFinanceRead})
	if err != nil || principal.OwnerID != "owner-1" || principal.CredentialKind != "user_api_key" || !repo.touched {
		t.Fatalf("unexpected principal: %+v err=%v touched=%t", principal, err, repo.touched)
	}
	if _, err := service.Authenticate(context.Background(), created.Secret, []string{entity.APIKeyScopeAdvisorChat}); !errors.Is(err, ErrAPIKeyScopeDenied) {
		t.Fatalf("missing scope must be denied, got %v", err)
	}
	if err := service.Revoke(context.Background(), "other-owner", repo.key.ID); err == nil {
		t.Fatal("cross-owner revoke must fail")
	}
	if err := service.Revoke(context.Background(), "owner-1", repo.key.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Authenticate(context.Background(), created.Secret, []string{entity.APIKeyScopeFinanceRead}); err == nil {
		t.Fatal("revoked key must not authenticate")
	}
}

func TestUserAPIKeyValidatePrincipalRejectsRevokedKey(t *testing.T) {
	repo := &apiKeyRepoStub{}
	service := NewUserAPIKeyService(repo)
	service.Now = func() time.Time { return time.Date(2026, 9, 22, 8, 0, 0, 0, time.UTC) }
	created, err := service.Create(context.Background(), APIKeyCreateInput{OwnerID: "owner-1", Name: "Advisor", Scopes: []string{entity.APIKeyScopeAdvisorChat}})
	if err != nil {
		t.Fatal(err)
	}
	principal, err := service.Authenticate(context.Background(), created.Secret, []string{entity.APIKeyScopeAdvisorChat})
	if err != nil {
		t.Fatal(err)
	}
	if err := service.Revoke(context.Background(), "owner-1", created.Key.ID); err != nil {
		t.Fatal(err)
	}
	if err := service.ValidatePrincipal(context.Background(), principal); !errors.Is(err, ErrAPIKeyInvalid) {
		t.Fatalf("revoked key must fail revalidation, got %v", err)
	}
}
