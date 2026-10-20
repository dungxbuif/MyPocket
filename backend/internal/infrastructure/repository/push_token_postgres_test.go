package repository

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	domain "github.com/mypocket/backend/internal/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestPushTokenPostgresOwnerIsolationAndReregistration(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires migrated local TEST_DATABASE_URL")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	owner, other := uuid.NewString(), uuid.NewString()
	for _, id := range []string{owner, other} {
		if err := db.Create(&entity.User{ID: id, Email: id + "@push-token.test", GoogleSubject: id}).Error; err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		db.Where("owner_id IN ?", []string{owner, other}).Delete(&entity.PushToken{})
		db.Where("id IN ?", []string{owner, other}).Delete(&entity.User{})
	})
	repo := NewPushTokenPostgresRepository(db)
	item := &entity.PushToken{ID: uuid.NewString(), OwnerID: owner, DeviceID: "iphone", DeviceName: "iPhone", Platform: "ios", Environment: "production", Token: strings.Repeat("a", 64), CreatedAt: nowUTC(), UpdatedAt: nowUTC()}
	if err := repo.Save(context.Background(), item); err != nil {
		t.Fatal(err)
	}
	firstID := item.ID
	item.Token = strings.Repeat("b", 64)
	item.DeviceName = "Renamed iPhone"
	if err := repo.Save(context.Background(), item); err != nil {
		t.Fatal(err)
	}
	if item.ID != firstID {
		t.Fatalf("same owner/device re-registration changed id: %s -> %s", firstID, item.ID)
	}
	foreign := *item
	foreign.ID = uuid.NewString()
	foreign.OwnerID = other
	foreign.Token = strings.Repeat("c", 64)
	if err := repo.Save(context.Background(), &foreign); err != nil {
		t.Fatalf("foreign distinct token should register: %v", err)
	}
	foreign.Token = item.Token
	if err := repo.Save(context.Background(), &foreign); err != domain.ErrPushTokenConflict {
		t.Fatalf("foreign duplicate token should conflict: %v", err)
	}
	if err := repo.Revoke(context.Background(), other, item.ID); err != domain.ErrNotFound {
		t.Fatalf("foreign owner revoked registration: %v", err)
	}
	if err := repo.Revoke(context.Background(), owner, item.ID); err != nil {
		t.Fatal(err)
	}
}

func nowUTC() time.Time { return time.Now().UTC() }
