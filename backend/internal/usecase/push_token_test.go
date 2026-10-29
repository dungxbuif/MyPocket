package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
)

type pushTokenRepoStub struct {
	items map[string]entity.PushToken
}

func (r *pushTokenRepoStub) Save(_ context.Context, item *entity.PushToken) error {
	if r.items == nil {
		r.items = map[string]entity.PushToken{}
	}
	for id, prior := range r.items {
		if prior.OwnerID == item.OwnerID && prior.DeviceID == item.DeviceID && prior.Platform == item.Platform && prior.Environment == item.Environment {
			item.ID = id
		}
		if prior.Token == item.Token && prior.Platform == item.Platform && prior.Environment == item.Environment && prior.OwnerID != item.OwnerID {
			return repository.ErrPushTokenConflict
		}
	}
	r.items[item.ID] = *item
	return nil
}
func (r *pushTokenRepoStub) Revoke(_ context.Context, owner, id string) error {
	item, ok := r.items[id]
	if !ok || item.OwnerID != owner {
		return repository.ErrNotFound
	}
	now := time.Now().UTC()
	item.RevokedAt = &now
	r.items[id] = item
	return nil
}

func TestPushTokenRegistrationValidatesAndIsOwnerScoped(t *testing.T) {
	repo := &pushTokenRepoStub{}
	svc := NewPushTokenInteractor(repo)
	first, err := svc.Register(context.Background(), "owner-a", PushTokenInput{DeviceID: "iphone-a", DeviceName: "iPhone", Platform: "ios", Environment: "production", Token: strings.Repeat("a", 64)})
	if err != nil {
		t.Fatal(err)
	}
	if first.Token != "" {
		t.Fatal("registration response must not expose the provider token")
	}
	second, err := svc.Register(context.Background(), "owner-a", PushTokenInput{DeviceID: "iphone-a", DeviceName: "iPhone renamed", Platform: "ios", Environment: "production", Token: strings.Repeat("b", 64)})
	if err != nil || second.ID != first.ID {
		t.Fatalf("same owner/device must update registration: first=%+v second=%+v err=%v", first, second, err)
	}
	if err := svc.Revoke(context.Background(), "other-owner", first.ID); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("foreign owner revoked token: %v", err)
	}
	if err := svc.Revoke(context.Background(), "owner-a", first.ID); err != nil {
		t.Fatal(err)
	}
	if repo.items[first.ID].RevokedAt == nil {
		t.Fatal("owner revoke did not persist")
	}
}

func TestPushTokenRegistrationRejectsInvalidAndForeignDuplicate(t *testing.T) {
	repo := &pushTokenRepoStub{}
	svc := NewPushTokenInteractor(repo)
	for _, in := range []PushTokenInput{
		{DeviceID: "", Platform: "ios", Environment: "production", Token: strings.Repeat("a", 64)},
		{DeviceID: "device", Platform: "android", Environment: "production", Token: strings.Repeat("a", 64)},
		{DeviceID: "device", Platform: "ios", Environment: "test", Token: strings.Repeat("a", 64)},
		{DeviceID: "device", Platform: "ios", Environment: "production", Token: "short"},
	} {
		if _, err := svc.Register(context.Background(), "owner-a", in); !errors.Is(err, ErrInvalidPushToken) {
			t.Fatalf("invalid input accepted: %+v err=%v", in, err)
		}
	}
	valid := PushTokenInput{DeviceID: "device-a", Platform: "ios", Environment: "production", Token: strings.Repeat("a", 64)}
	if _, err := svc.Register(context.Background(), "owner-a", valid); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Register(context.Background(), "owner-b", PushTokenInput{DeviceID: "device-b", Platform: "ios", Environment: "production", Token: valid.Token}); !errors.Is(err, repository.ErrPushTokenConflict) {
		t.Fatalf("duplicate token crossed owners without conflict: %v", err)
	}
}
