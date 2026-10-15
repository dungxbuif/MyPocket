package repository

import (
	"context"
	"errors"
	"os"
	"strings"
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

func advisorCancellationTestDB(t *testing.T) (*gorm.DB, string) {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires migrated local TEST_DATABASE_URL")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	owner := uuid.NewString()
	if err := db.Create(&entity.User{ID: owner, Email: owner + "@advisor-cancel.test", GoogleSubject: owner}).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		// Only this synthetic owner and its FK-cascaded advisor rows.
		if err := db.Where("id = ?", owner).Delete(&entity.User{}).Error; err != nil {
			t.Error(err)
		}
	})
	return db, owner
}

func TestAdvisorPostgresCancelledRunCannotMutateNewRun(t *testing.T) {
	db, owner := advisorCancellationTestDB(t)
	store := NewAdvisorPostgresRepository(db)
	ctx := context.Background()
	input := repository.AdvisorRunInput{OwnerID: owner, CredentialKind: "session", CredentialID: "synthetic-session", RequestID: "first", PayloadHash: "first", Parts: []entity.AdvisorPart{{Type: "text", Text: "synthetic question"}}}
	old, _, _, err := store.StartRun(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.MarkRunStatus(ctx, uuid.NewString(), old.ID, entity.AdvisorStatusCancelled, time.Now()); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("foreign owner cancellation: %v", err)
	}
	if err := store.MarkRunStatus(ctx, owner, old.ID, entity.AdvisorStatusCancelled, time.Now()); err != nil {
		t.Fatal(err)
	}
	input.RequestID, input.PayloadHash = "second", "second"
	newRun, newMessage, replay, err := store.StartRun(ctx, input)
	if err != nil || replay {
		t.Fatalf("new run after cancellation: replay=%t err=%v", replay, err)
	}
	if newMessage.Seq != 2 {
		t.Fatalf("new user message sequence=%d, want2", newMessage.Seq)
	}
	if err := store.MarkRunStatus(ctx, owner, old.ID, entity.AdvisorStatusCompleted, time.Now()); !errors.Is(err, entity.ErrAdvisorRunTransition) {
		t.Fatalf("late completion accepted: %v", err)
	}
	if _, err := store.CreateAssistantMessage(ctx, owner, old.ID, old.Generation, []entity.AdvisorPart{{Type: "text", Text: "late old answer"}}); !errors.Is(err, repository.ErrAdvisorLeaseLost) {
		t.Fatalf("late old answer accepted: %v", err)
	}
	fresh, err := store.GetRun(ctx, owner, newRun.ID)
	if err != nil || fresh.Status != entity.AdvisorStatusQueued {
		t.Fatalf("new run changed: run=%+v err=%v", fresh, err)
	}
	rows, err := store.ListMessages(ctx, owner, newRun.ConversationID, 0, 50)
	if err != nil || len(rows) != 2 {
		t.Fatalf("unexpected persisted history count=%d err=%v", len(rows), err)
	}
	for _, row := range rows {
		if row.Role != entity.AdvisorRoleUser {
			t.Fatal("cancelled answer persisted")
		}
	}
}

func TestAdvisorPostgresCancellationWhileAnswerWaitsMustRejectLateInsert(t *testing.T) {
	db, owner := advisorCancellationTestDB(t)
	store := NewAdvisorPostgresRepository(db)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	run, _, _, err := store.StartRun(ctx, repository.AdvisorRunInput{OwnerID: owner, CredentialKind: "session", CredentialID: "synthetic-session", RequestID: "race", PayloadHash: "race", Parts: []entity.AdvisorPart{{Type: "text", Text: "synthetic question"}}})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.MarkRunStatus(ctx, owner, run.ID, entity.AdvisorStatusRunning, time.Now()); err != nil {
		t.Fatal(err)
	}
	blocker := db.WithContext(ctx).Begin()
	if blocker.Error != nil {
		t.Fatal(blocker.Error)
	}
	defer blocker.Rollback()
	if err := blocker.Exec("SELECT id FROM advisor_conversations WHERE id = ? FOR UPDATE", run.ConversationID).Error; err != nil {
		t.Fatal(err)
	}
	writer := db.WithContext(ctx).Begin()
	if writer.Error != nil {
		t.Fatal(writer.Error)
	}
	defer writer.Rollback()
	var writerPID int
	if err := writer.Raw("SELECT pg_backend_pid()").Scan(&writerPID).Error; err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		defer close(done)
		_, err := NewAdvisorPostgresRepository(writer).CreateAssistantMessage(ctx, owner, run.ID, run.Generation, []entity.AdvisorPart{{Type: "text", Text: "late raced answer"}})
		if err == nil {
			err = writer.Commit().Error
		} else {
			_ = writer.Rollback().Error
		}
		done <- err
	}()
	defer func() {
		cancel()
		_ = blocker.Rollback().Error
		select {
		case <-done:
		case <-time.After(time.Second):
		}
	}()
	deadline := time.Now().Add(2 * time.Second)
	waiting := false
	for time.Now().Before(deadline) {
		var activity struct {
			WaitEventType string
			Query         string
		}
		if err := db.WithContext(ctx).Raw("SELECT coalesce(wait_event_type,'') AS wait_event_type, query FROM pg_stat_activity WHERE pid = ?", writerPID).Scan(&activity).Error; err != nil {
			t.Fatal(err)
		}
		if activity.WaitEventType == "Lock" && strings.Contains(activity.Query, "advisor_conversations") {
			waiting = true
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !waiting {
		t.Fatal("answer writer did not reach controlled conversation-lock barrier")
	}
	if err := store.MarkRunStatus(ctx, owner, run.ID, entity.AdvisorStatusCancelled, time.Now()); err != nil {
		t.Fatal(err)
	}
	if err := blocker.Rollback().Error; err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		if !errors.Is(err, repository.ErrAdvisorLeaseLost) {
			t.Errorf("answer insertion after committed cancellation must lose lease; got %v", err)
		}
	case <-ctx.Done():
		t.Fatal("answer writer did not finish after lock release")
	}
	rows, err := store.ListMessages(ctx, owner, run.ConversationID, 0, 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Role != entity.AdvisorRoleUser {
		t.Errorf("cancelled run must retain only original question; rows=%+v", rows)
	}
}

func TestAdvisorPostgresResetIsIdempotentAndPreservesNewQuestion(t *testing.T) {
	db, owner := advisorCancellationTestDB(t)
	if !db.Migrator().HasTable("advisor_reset_requests") {
		t.Skip("requires migration 000034_advisor_reset_requests")
	}
	store := NewAdvisorPostgresRepository(db)
	lifecycle, ok := store.(repository.AdvisorLifecycleRepository)
	if !ok {
		t.Fatal("postgres advisor store must expose lifecycle reset")
	}
	ctx := context.Background()
	first, _, _, err := store.StartRun(ctx, repository.AdvisorRunInput{
		OwnerID: owner, CredentialKind: "session", CredentialID: "synthetic-session",
		RequestID: "first", PayloadHash: "first", Parts: []entity.AdvisorPart{{Type: "text", Text: "old question"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	reset, err := lifecycle.ResetConversation(ctx, owner, "reset-1", "empty-body")
	if err != nil || reset.Replayed || reset.Generation != first.Generation+1 {
		t.Fatalf("first reset mismatch: result=%+v err=%v", reset, err)
	}
	if len(reset.PurgedRunIDs) != 1 || reset.PurgedRunIDs[0] != first.ID {
		t.Fatalf("first reset must return the purged active run for in-process cancellation: %+v", reset.PurgedRunIDs)
	}
	rows, err := store.ListMessages(ctx, owner, first.ConversationID, 0, 50)
	if err != nil || len(rows) != 0 {
		t.Fatalf("reset must clear visible history: rows=%d err=%v", len(rows), err)
	}
	newRun, _, _, err := store.StartRun(ctx, repository.AdvisorRunInput{
		OwnerID: owner, CredentialKind: "session", CredentialID: "synthetic-session",
		RequestID: "second", PayloadHash: "second", Parts: []entity.AdvisorPart{{Type: "text", Text: "new question"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	replayed, err := lifecycle.ResetConversation(ctx, owner, "reset-1", "empty-body")
	if err != nil || !replayed.Replayed || replayed.Generation != reset.Generation || !replayed.ServerTime.Equal(reset.ServerTime) {
		t.Fatalf("same reset key must replay original result: result=%+v err=%v", replayed, err)
	}
	if len(replayed.PurgedRunIDs) != 0 {
		t.Fatalf("replayed reset must not report old runs for cancellation: %+v", replayed.PurgedRunIDs)
	}
	rows, err = store.ListMessages(ctx, owner, newRun.ConversationID, 0, 50)
	if err != nil || len(rows) != 1 || rows[0].Parts[0].Text != "new question" {
		t.Fatalf("replayed reset must preserve newer question: rows=%+v err=%v", rows, err)
	}
	if _, err := lifecycle.ResetConversation(ctx, owner, "reset-1", "changed-body"); !errors.Is(err, repository.ErrAdvisorRequestConflict) {
		t.Fatalf("changed reset payload must conflict, got %v", err)
	}
}
