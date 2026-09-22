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

func TestFeedbackPostgresOwnerScopeAndChangelogAtomicity(t *testing.T) {
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
		if err := db.Create(&entity.User{ID: id, Email: id + "@feedback.test", GoogleSubject: id}).Error; err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		db.Where("user_id IN ?", []string{owner, other}).Delete(&entity.Feedback{})
		db.Where("id IN ?", []string{owner, other}).Delete(&entity.User{})
	})

	feedbackRepo := NewFeedbackPostgresRepository(db)
	changelogRepo := NewChangelogPostgresRepository(db)
	first := &entity.Feedback{ID: uuid.NewString(), UserID: owner, Type: entity.FeedbackTypeBug, Title: "Owner issue", Description: "Private details", Status: entity.FeedbackStatusOpen}
	second := &entity.Feedback{ID: uuid.NewString(), UserID: other, Type: entity.FeedbackTypeBug, Title: "Other issue", Description: "Other private details", Status: entity.FeedbackStatusOpen}
	if err := feedbackRepo.Create(context.Background(), first); err != nil {
		t.Fatal(err)
	}
	if err := feedbackRepo.Create(context.Background(), second); err != nil {
		t.Fatal(err)
	}
	rows, err := feedbackRepo.ListByOwner(context.Background(), owner)
	if err != nil || len(rows) != 1 || rows[0].ID != first.ID {
		t.Fatalf("owner scope mismatch: rows=%+v err=%v", rows, err)
	}
	if _, err := feedbackRepo.FindByOwner(context.Background(), other, first.ID); err == nil {
		t.Fatal("cross-owner feedback read must fail")
	}

	changelog := &entity.Changelog{ID: uuid.NewString(), Version: "1.4.2-test", Title: "Feedback fix", Description: "Fixed the issue", PublishedAt: time.Now().UTC()}
	if err := changelogRepo.Publish(context.Background(), changelog, []string{first.ID}); err != nil {
		t.Fatal(err)
	}
	fixed, err := feedbackRepo.FindByOwner(context.Background(), owner, first.ID)
	if err != nil || fixed.Status != entity.FeedbackStatusFixed || fixed.ChangelogID == nil || *fixed.ChangelogID != changelog.ID || fixed.FixedAt == nil {
		t.Fatalf("feedback was not finalized: %+v err=%v", fixed, err)
	}
	if err := changelogRepo.Publish(context.Background(), &entity.Changelog{ID: uuid.NewString(), Version: "1.4.2-test", Title: "Duplicate", Description: "Duplicate", PublishedAt: time.Now().UTC()}, []string{second.ID}); err == nil {
		t.Fatal("duplicate changelog version must fail")
	}
}
