package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestUserAPIKeyPostgresOwnerIsolationAndRevoke(t *testing.T) {
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
		if err := db.Create(&entity.User{ID: id, Email: id + "@key.test", GoogleSubject: id, Timezone: "Asia/Ho_Chi_Minh"}).Error; err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		db.Where("owner_id IN ?", []string{owner, other}).Delete(&entity.UserAPIKey{})
		db.Where("id IN ?", []string{owner, other}).Delete(&entity.User{})
	})
	store := NewUserAPIKeyPostgresRepository(db)
	key := &entity.UserAPIKey{ID: uuid.NewString(), OwnerID: owner, LookupID: uuid.NewString(), Name: "CLI", SecretHash: uuid.NewString(), Scopes: []string{entity.APIKeyScopeFinanceRead}, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	if err := store.Create(context.Background(), key); err != nil {
		t.Fatal(err)
	}
	found, err := store.FindByLookup(context.Background(), key.LookupID)
	if err != nil || found.OwnerID != owner || found.SecretHash == "" {
		t.Fatalf("unexpected key lookup: %+v %v", found, err)
	}
	keys, err := store.ListByOwner(context.Background(), other)
	if err != nil || len(keys) != 0 {
		t.Fatalf("other owner must not list key: %+v %v", keys, err)
	}
	if err := store.Revoke(context.Background(), other, key.ID, time.Now().UTC()); err == nil {
		t.Fatal("cross-owner revoke must fail")
	}
	if err := store.Revoke(context.Background(), owner, key.ID, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	if err := store.TouchLastUsed(context.Background(), key.ID, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
}
