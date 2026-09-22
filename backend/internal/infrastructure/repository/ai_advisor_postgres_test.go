package repository

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestAdvisorPostgresStartRunIsIdempotentAndBusyScoped(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires migrated local TEST_DATABASE_URL")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	owner := uuid.NewString()
	if err := db.Create(&entity.User{ID: owner, Email: owner + "@advisor.test", GoogleSubject: owner}).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Where("owner_id = ?", owner).Delete(&entity.AdvisorFactBundle{})
		db.Exec("DELETE FROM advisor_events WHERE run_id IN (SELECT id FROM advisor_runs WHERE owner_id = ?)", owner)
		db.Exec("DELETE FROM advisor_messages WHERE run_id IN (SELECT id FROM advisor_runs WHERE owner_id = ?)", owner)
		db.Where("owner_id = ?", owner).Delete(&entity.AdvisorRun{})
		db.Where("owner_id = ?", owner).Delete(&entity.AdvisorConversation{})
		db.Where("id = ?", owner).Delete(&entity.User{})
	})

	store := NewAdvisorPostgresRepository(db)
	parts := []entity.AdvisorPart{{Type: "text", Text: "Tôi đã chi bao nhiêu?"}}
	first, message, replay, err := store.StartRun(context.Background(), repository.AdvisorRunInput{OwnerID: owner, CredentialKind: "session", CredentialID: "session-1", RequestID: "req-1", PayloadHash: "hash-1", Parts: parts})
	if err != nil || replay || first.Status != entity.AdvisorStatusQueued || message.Seq != 1 {
		t.Fatalf("first run mismatch: run=%+v message=%+v replay=%t err=%v", first, message, replay, err)
	}
	replayed, replayMessage, replay, err := store.StartRun(context.Background(), repository.AdvisorRunInput{OwnerID: owner, CredentialKind: "session", CredentialID: "session-1", RequestID: "req-1", PayloadHash: "hash-1", Parts: parts})
	if err != nil || !replay || replayed.ID != first.ID || replayMessage.ID != message.ID {
		t.Fatalf("same request must replay: run=%+v message=%+v replay=%t err=%v", replayed, replayMessage, replay, err)
	}
	if _, _, _, err := store.StartRun(context.Background(), repository.AdvisorRunInput{OwnerID: owner, CredentialKind: "session", CredentialID: "session-1", RequestID: "req-1", PayloadHash: "changed", Parts: parts}); !errors.Is(err, repository.ErrAdvisorRequestConflict) {
		t.Fatalf("changed request payload must conflict, got %v", err)
	}
	if _, _, _, err := store.StartRun(context.Background(), repository.AdvisorRunInput{OwnerID: owner, CredentialKind: "session", CredentialID: "session-1", RequestID: "req-2", PayloadHash: "hash-2", Parts: parts}); !errors.Is(err, repository.ErrAdvisorBusy) {
		t.Fatalf("second active request must be busy, got %v", err)
	}
	if err := store.MarkRunStatus(context.Background(), owner, first.ID, entity.AdvisorStatusCancelled, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	if _, err := store.CreateAssistantMessage(context.Background(), owner, first.ID, first.Generation, []entity.AdvisorPart{{Type: "text", Text: "late answer"}}); !errors.Is(err, repository.ErrAdvisorLeaseLost) {
		t.Fatalf("cancelled run must not accept a late assistant message, got %v", err)
	}
}
