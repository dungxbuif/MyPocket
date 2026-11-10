package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
)

type assistantDelegationRepoStub struct{ value *entity.AssistantDelegation }

func (r *assistantDelegationRepoStub) Create(_ context.Context, value *entity.AssistantDelegation) error {
	copy := *value
	r.value = &copy
	return nil
}
func (r *assistantDelegationRepoStub) FindByLookup(_ context.Context, lookup string) (*entity.AssistantDelegation, error) {
	if r.value == nil || r.value.LookupID != lookup {
		return nil, repository.ErrNotFound
	}
	copy := *r.value
	return &copy, nil
}
func (r *assistantDelegationRepoStub) FindByID(context.Context, string) (*entity.AssistantDelegation, error) {
	return r.value, nil
}
func (r *assistantDelegationRepoStub) ListByOwner(context.Context, string) ([]entity.AssistantDelegation, error) {
	return nil, nil
}
func (r *assistantDelegationRepoStub) Revoke(_ context.Context, owner, id string, at time.Time) error {
	if r.value == nil || r.value.OwnerID != owner || r.value.ID != id {
		return repository.ErrNotFound
	}
	r.value.RevokedAt = &at
	return nil
}
func (r *assistantDelegationRepoStub) TouchLastUsed(context.Context, string, time.Time) error {
	return nil
}

func TestAssistantDelegationIsOwnerBoundScopedAndRevocable(t *testing.T) {
	repo := &assistantDelegationRepoStub{}
	service := NewAssistantDelegationService(repo)
	now := time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC)
	service.Now = func() time.Time { return now }
	created, err := service.Create(context.Background(), "owner-a", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	principal, err := service.Authenticate(context.Background(), created.Secret, "GET", "/api/v1/wallets")
	if err != nil || principal.OwnerID != "owner-a" || principal.CredentialKind != "assistant_delegation" {
		t.Fatalf("principal=%#v err=%v", principal, err)
	}
	if _, err := service.Authenticate(context.Background(), created.Secret, "POST", "/api/v1/transactions"); err == nil {
		t.Fatal("read-only assistant delegation accepted a financial write")
	}
	if _, err := service.Authenticate(context.Background(), created.Secret, "DELETE", "/api/v1/wallets/other-owner"); err == nil {
		t.Fatal("out-of-scope operation accepted")
	}
	if err := service.Revoke(context.Background(), "owner-a", created.Delegation.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Authenticate(context.Background(), created.Secret, "GET", "/api/v1/wallets"); err == nil {
		t.Fatal("revoked credential accepted")
	}
}
