package repository

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestMonthNoteIsOwnerScopedAndCanBeCleared(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires migrated local TEST_DATABASE_URL")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	owner, otherOwner := uuid.NewString(), uuid.NewString()
	users := []entity.User{
		{ID: owner, Email: owner + "@test.invalid", GoogleSubject: owner},
		{ID: otherOwner, Email: otherOwner + "@test.invalid", GoogleSubject: otherOwner},
	}
	for _, user := range users {
		if err = db.Create(&user).Error; err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		db.Where("owner_id IN ?", []string{owner, otherOwner}).Delete(&entity.MonthNote{})
		for _, id := range []string{owner, otherOwner} {
			db.Where("id = ?", id).Delete(&entity.User{})
		}
	})

	repo := NewMonthPostgresRepository(db)
	if err = repo.Save(owner, "2026-08", "My August note"); err != nil {
		t.Fatal(err)
	}
	if err = repo.Save(otherOwner, "2026-08", "Other account note"); err != nil {
		t.Fatal(err)
	}
	got, err := repo.Find(owner, "2026-08")
	if err != nil || got != "My August note" {
		t.Fatalf("owner note mismatch: got %q, err=%v", got, err)
	}
	got, err = repo.Find(otherOwner, "2026-08")
	if err != nil || got != "Other account note" {
		t.Fatalf("other owner note mismatch: got %q, err=%v", got, err)
	}
	if err = repo.Save(owner, "2026-08", ""); err != nil {
		t.Fatal(err)
	}
	got, err = repo.Find(owner, "2026-08")
	if err != nil || got != "" {
		t.Fatalf("cleared note was retained: got %q, err=%v", got, err)
	}
	got, err = repo.Find(otherOwner, "2026-08")
	if err != nil || got != "Other account note" {
		t.Fatalf("clearing one owner's note affected another: got %q, err=%v", got, err)
	}
}
