package repository

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	contract "github.com/mypocket/backend/internal/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func advisorRuntimeTestDB(t *testing.T) (*gorm.DB, string) {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires migrated local TEST_DATABASE_URL")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if !db.Migrator().HasColumn(&entity.AdvisorRun{}, "DeadlineAt") {
		t.Skip("requires advisor runtime migration 000041")
	}
	owner := uuid.NewString()
	if err := db.Create(&entity.User{ID: owner, Email: owner + "@advisor-runtime.test", GoogleSubject: owner}).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Where("id = ?", owner).Delete(&entity.User{}).Error })
	return db, owner
}

func TestAdvisorLeaseSingleClaim(t *testing.T) {
	db, owner := advisorRuntimeTestDB(t)
	store := NewAdvisorPostgresRepository(db)
	run, _, _, err := store.StartRun(context.Background(), contract.AdvisorRunInput{OwnerID: owner, CredentialKind: "session", CredentialID: "test", RequestID: uuid.NewString(), PayloadHash: "hash", Parts: []entity.AdvisorPart{{Type: "text", Text: "summary"}}})
	if err != nil {
		t.Fatal(err)
	}
	first, err := store.(*AdvisorPostgresRepository).ClaimRun(context.Background(), owner, run.ID, "token-a")
	if err != nil || first.Token != "token-a" {
		t.Fatalf("first claim failed: %+v %v", first, err)
	}
	if _, err := store.(*AdvisorPostgresRepository).ClaimRun(context.Background(), owner, run.ID, "token-b"); !errors.Is(err, contract.ErrAdvisorLeaseLost) {
		t.Fatalf("second executor must not steal lease: %v", err)
	}
}

func TestAdvisorCompletionAckLostReplay(t *testing.T) {
	db, owner := advisorRuntimeTestDB(t)
	store := NewAdvisorPostgresRepository(db).(*AdvisorPostgresRepository)
	run, _, _, err := store.StartRun(context.Background(), contract.AdvisorRunInput{OwnerID: owner, CredentialKind: "session", CredentialID: "test", RequestID: uuid.NewString(), PayloadHash: "hash", Parts: []entity.AdvisorPart{{Type: "text", Text: "summary"}}})
	if err != nil {
		t.Fatal(err)
	}
	lease, err := store.ClaimRun(context.Background(), owner, run.ID, "token-a")
	if err != nil {
		t.Fatal(err)
	}
	completion := entity.AdvisorCompletion{Parts: []entity.AdvisorPart{{Type: "text", Text: "validated answer"}}, Usage: map[string]any{"total_tokens": int64(3)}}
	first, err := store.CompleteRun(context.Background(), lease, completion)
	if err != nil || first.Replayed || first.Message == nil {
		t.Fatalf("completion failed: %+v %v", first, err)
	}
	replay, err := store.CompleteRun(context.Background(), lease, completion)
	if err != nil || !replay.Replayed || replay.Message.ID != first.Message.ID {
		t.Fatalf("completion replay mismatch: first=%+v replay=%+v err=%v", first, replay, err)
	}
	var count int64
	if err := db.Model(&entity.AdvisorMessage{}).Where("run_id = ? AND role = ?", run.ID, entity.AdvisorRoleAssistant).Count(&count).Error; err != nil || count != 1 {
		t.Fatalf("replay inserted duplicate assistant message: count=%d err=%v", count, err)
	}
}

func TestAdvisorCompletionRejectsCancelledOrForeignLease(t *testing.T) {
	db, owner := advisorRuntimeTestDB(t)
	store := NewAdvisorPostgresRepository(db).(*AdvisorPostgresRepository)
	run, _, _, err := store.StartRun(context.Background(), contract.AdvisorRunInput{OwnerID: owner, CredentialKind: "session", CredentialID: "test", RequestID: uuid.NewString(), PayloadHash: "hash", Parts: []entity.AdvisorPart{{Type: "text", Text: "summary"}}})
	if err != nil {
		t.Fatal(err)
	}
	lease, err := store.ClaimRun(context.Background(), owner, run.ID, "token-a")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.MarkRunStatus(context.Background(), owner, run.ID, entity.AdvisorStatusCancelled, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	if _, err := store.CompleteRun(context.Background(), lease, entity.AdvisorCompletion{Parts: []entity.AdvisorPart{{Type: "text", Text: "late"}}}); !errors.Is(err, contract.ErrAdvisorLeaseLost) {
		t.Fatalf("cancelled run accepted completion: %v", err)
	}
	foreign := lease
	foreign.OwnerID = uuid.NewString()
	if _, err := store.CompleteRun(context.Background(), foreign, entity.AdvisorCompletion{Parts: []entity.AdvisorPart{{Type: "text", Text: "foreign"}}}); !errors.Is(err, contract.ErrNotFound) && !errors.Is(err, contract.ErrAdvisorLeaseLost) {
		t.Fatalf("foreign lease returned unsafe error: %v", err)
	}
}

func TestAdvisorReaperClosesExpiredExecutorLeaseWithoutReplay(t *testing.T) {
	db, owner := advisorRuntimeTestDB(t)
	store := NewAdvisorPostgresRepository(db).(*AdvisorPostgresRepository)
	run, _, _, err := store.StartRun(context.Background(), contract.AdvisorRunInput{
		OwnerID: owner, CredentialKind: "session", CredentialID: "test",
		RequestID: uuid.NewString(), PayloadHash: "hash",
		Parts: []entity.AdvisorPart{{Type: "text", Text: "summary"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.ClaimRun(context.Background(), owner, run.ID, "executor-that-died"); err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&entity.AdvisorRun{}).Where("id = ? AND owner_id = ?", run.ID, owner).Updates(map[string]any{
		"lease_until": gorm.Expr("NOW() - INTERVAL '1 second'"),
	}).Error; err != nil {
		t.Fatal(err)
	}
	count, err := store.ReapExpiredRuns(context.Background(), 100)
	if err != nil || count != 1 {
		t.Fatalf("reaper must close exactly one lost executor run: count=%d err=%v", count, err)
	}
	var persisted entity.AdvisorRun
	if err := db.Where("id = ? AND owner_id = ?", run.ID, owner).First(&persisted).Error; err != nil {
		t.Fatal(err)
	}
	if persisted.Status != entity.AdvisorStatusInterrupted || persisted.ErrorCode != "executor_lost" {
		t.Fatalf("lost executor must become interrupted without replay: status=%s error=%s", persisted.Status, persisted.ErrorCode)
	}
	var eventCount int64
	if err := db.Model(&entity.AdvisorEvent{}).Where("run_id = ? AND type = ?", run.ID, "run.interrupted").Count(&eventCount).Error; err != nil || eventCount != 1 {
		t.Fatalf("reaper must record one terminal event: count=%d err=%v", eventCount, err)
	}
}

func TestAdvisorReaperRechecksCandidateLeaseUnderLock(t *testing.T) {
	db, owner := advisorRuntimeTestDB(t)
	store := NewAdvisorPostgresRepository(db).(*AdvisorPostgresRepository)
	run, _, _, err := store.StartRun(context.Background(), contract.AdvisorRunInput{OwnerID: owner, CredentialKind: "session", CredentialID: "test", RequestID: uuid.NewString(), PayloadHash: "hash", Parts: []entity.AdvisorPart{{Type: "text", Text: "test"}}})
	if err != nil {
		t.Fatal(err)
	}
	lease, err := store.ClaimRun(context.Background(), owner, run.ID, "renewed-executor")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.HeartbeatRun(context.Background(), lease); err != nil {
		t.Fatal(err)
	}
	// A candidate selected before renewal must be harmless after acquiring locks.
	updated, err := store.reapOne(context.Background(), struct{ ID, OwnerID, ConversationID string }{run.ID, owner, run.ConversationID})
	if err != nil || updated {
		t.Fatalf("reaper killed a live renewed lease: updated=%v err=%v", updated, err)
	}
	got, err := store.GetRun(context.Background(), owner, run.ID)
	if err != nil || got.Status != entity.AdvisorStatusRunning {
		t.Fatalf("live run changed: %+v %v", got, err)
	}
}

func TestAdvisorContextStatePersistsWithOwnerGenerationIsolation(t *testing.T) {
	db, owner := advisorRuntimeTestDB(t)
	store := NewAdvisorPostgresRepository(db).(*AdvisorPostgresRepository)
	run, _, _, err := store.StartRun(context.Background(), contract.AdvisorRunInput{OwnerID: owner, CredentialKind: "session", CredentialID: "test", RequestID: uuid.NewString(), PayloadHash: "hash", Parts: []entity.AdvisorPart{{Type: "text", Text: "test"}}})
	if err != nil {
		t.Fatal(err)
	}
	lease, err := store.ClaimRun(context.Background(), owner, run.ID, "context-executor")
	if err != nil {
		t.Fatal(err)
	}
	completed, err := store.CompleteRun(context.Background(), lease, entity.AdvisorCompletion{Parts: []entity.AdvisorPart{{Type: "text", Text: "answer"}}, ContextState: map[string]any{"version": 1, "generation": run.Generation, "date_from": "2026-09-01", "date_to": "2026-09-30", "timezone": "Asia/Ho_Chi_Minh"}})
	if err != nil {
		t.Fatal(err)
	}
	state, err := store.LoadAdvisorContextState(context.Background(), owner, run.ConversationID, run.Generation)
	if err != nil || state.DateFrom != "2026-09-01" || state.ThroughSeq != completed.Message.Seq {
		t.Fatalf("state lost or wrong sequence: %+v %v", state, err)
	}
	if _, err := store.LoadAdvisorContextState(context.Background(), uuid.NewString(), run.ConversationID, run.Generation); !errors.Is(err, contract.ErrNotFound) {
		t.Fatalf("foreign owner read state: %v", err)
	}
	if _, err := store.LoadAdvisorContextState(context.Background(), owner, run.ConversationID, run.Generation+1); !errors.Is(err, contract.ErrNotFound) {
		t.Fatalf("wrong generation read state: %v", err)
	}
}

func TestAdvisorRuntimeMigrationAllowsLegacyWriterDuringRollingUpdate(t *testing.T) {
	db, owner := advisorRuntimeTestDB(t)
	store := NewAdvisorPostgresRepository(db).(*AdvisorPostgresRepository)
	run, _, _, err := store.StartRun(context.Background(), contract.AdvisorRunInput{OwnerID: owner, CredentialKind: "session", CredentialID: "test", RequestID: uuid.NewString(), PayloadHash: "hash", Parts: []entity.AdvisorPart{{Type: "text", Text: "test"}}})
	if err != nil {
		t.Fatal(err)
	}
	legacy := *run
	legacy.ID = uuid.NewString()
	legacy.ClientRequestID = uuid.NewString()
	legacy.Status = entity.AdvisorStatusInterrupted
	legacy.DeadlineAt = nil
	if err := db.Omit("DeadlineAt").Create(&legacy).Error; err != nil {
		t.Fatalf("old API writer fails after additive migration: %v", err)
	}
	var stored entity.AdvisorRun
	if err := db.Where("id = ?", legacy.ID).First(&stored).Error; err != nil {
		t.Fatal(err)
	}
	if stored.DeadlineAt == nil || !stored.DeadlineAt.After(stored.CreatedAt) {
		t.Fatalf("legacy writer has no server deadline: %+v", stored.DeadlineAt)
	}
}
