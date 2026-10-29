package repository

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestNotificationPostgresOwnerScopeAndIdempotentRead(t *testing.T) {
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
		if err := db.Create(&entity.User{ID: id, Email: id + "@notification.test", GoogleSubject: id}).Error; err != nil {
			t.Fatal(err)
		}
	}
	item := &entity.Notification{ID: uuid.NewString(), OwnerID: owner, Kind: "budget", Title: "Budget", Body: "Near limit", DedupeKey: uuid.NewString()}
	if err := db.Create(item).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Where("id = ?", item.ID).Delete(&entity.Notification{})
		db.Where("id IN ?", []string{owner, other}).Delete(&entity.User{})
	})

	repo := NewNotificationPostgresRepository(db)
	rows, err := repo.List(context.Background(), owner)
	if err != nil || len(rows) != 1 || rows[0].ID != item.ID {
		t.Fatalf("unexpected owner list: rows=%+v err=%v", rows, err)
	}
	if rows, err := repo.List(context.Background(), other); err != nil || len(rows) != 0 {
		t.Fatalf("cross-owner notification leak: rows=%+v err=%v", rows, err)
	}
	if err := repo.MarkRead(context.Background(), owner, item.ID); err != nil {
		t.Fatal(err)
	}
	if err := repo.MarkRead(context.Background(), owner, item.ID); err != nil {
		t.Fatalf("read retry must be idempotent: %v", err)
	}
	if err := repo.MarkRead(context.Background(), other, item.ID); err == nil {
		t.Fatal("foreign owner must not mark notification read")
	}
}

func TestNotificationPostgresCreateIfAbsentIsOwnerScopedAndReplaySafe(t *testing.T) {
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
		if err := db.Create(&entity.User{ID: id, Email: id + "@notification-producer.test", GoogleSubject: id}).Error; err != nil {
			t.Fatal(err)
		}
	}
	item := &entity.Notification{ID: uuid.NewString(), OwnerID: owner, Kind: "budget.threshold.v1", Title: "Budget", Body: "Near limit", DedupeKey: "budget-1:2026-10:80"}
	t.Cleanup(func() {
		db.Where("owner_id IN ?", []string{owner, other}).Delete(&entity.Notification{})
		db.Where("id IN ?", []string{owner, other}).Delete(&entity.User{})
	})

	repo := NewNotificationPostgresRepository(db)
	created, err := repo.CreateIfAbsent(context.Background(), item)
	if err != nil || !created {
		t.Fatalf("first insert created=%v err=%v", created, err)
	}
	replayed := *item
	replayed.ID = uuid.NewString()
	replayed.Body = "changed body must not replace the original"
	created, err = repo.CreateIfAbsent(context.Background(), &replayed)
	if err != nil || created {
		t.Fatalf("duplicate insert created=%v err=%v", created, err)
	}
	rows, err := repo.List(context.Background(), owner)
	if err != nil || len(rows) != 1 || rows[0].Body != item.Body {
		t.Fatalf("dedupe must preserve one original row: rows=%+v err=%v", rows, err)
	}
	foreign := *item
	foreign.ID = uuid.NewString()
	foreign.OwnerID = other
	created, err = repo.CreateIfAbsent(context.Background(), &foreign)
	if err != nil || !created {
		t.Fatalf("same dedupe key must be scoped per owner: created=%v err=%v", created, err)
	}
}
