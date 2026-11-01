package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	contract "github.com/mypocket/backend/internal/repository"
	"gorm.io/gorm"
)

func advisorActionJSONFixture(t *testing.T) (*gorm.DB, *AdvisorPostgresRepository, string, *entity.AdvisorRun, *entity.AdvisorMessage, string) {
	t.Helper()
	db, owner := advisorCancellationTestDB(t)
	store := NewAdvisorPostgresRepository(db).(*AdvisorPostgresRepository)
	run, _, _, err := store.StartRun(context.Background(), contract.AdvisorRunInput{
		OwnerID: owner, CredentialKind: "session", CredentialID: "synthetic-json",
		RequestID: "json-action", PayloadHash: "json-action",
		Parts: []entity.AdvisorPart{{Type: "text", Text: "Tạo hũ Du lịch tháng 11/2026"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	actionID := uuid.NewString()
	message, err := store.CreateAssistantMessage(context.Background(), owner, run.ID, run.Generation, []entity.AdvisorPart{
		{Type: "text", Text: "Bạn kiểm tra và xác nhận trên thẻ để lưu."},
		{Type: "jar_confirmation", Data: map[string]any{"status": "proposed", "warnings": []string{"review_required"}, "view": map[string]any{
			"action_id": actionID, "action": "create_jar", "status": entity.AIProposalPending, "revision": 1,
			"conversation_id": run.ConversationID, "generation": run.Generation, "run_id": run.ID,
			"name": "Du lịch", "month": "2026-11", "allocation_mode": "fixed", "allocation_amount": 2000000,
		}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return db, store, owner, run, message, actionID
}

func reloadAdvisorActionJSON(t *testing.T, db *gorm.DB, original *entity.AdvisorMessage, parts int) []entity.AdvisorPart {
	t.Helper()
	var persisted entity.AdvisorMessage
	if err := db.First(&persisted, "id = ?", original.ID).Error; err != nil {
		t.Fatal(err)
	}
	if persisted.ConversationID != original.ConversationID || persisted.Generation != original.Generation || persisted.RunID != original.RunID || persisted.Role != original.Role || persisted.Seq != original.Seq || len(persisted.Parts) != parts || persisted.Parts[0].Text != original.Parts[0].Text {
		t.Fatalf("parts update changed message metadata or sibling text: %+v", persisted)
	}
	var jsonType string
	if err := db.Raw("SELECT jsonb_typeof(parts) FROM advisor_messages WHERE id = ?", original.ID).Scan(&jsonType).Error; err != nil || jsonType != "array" {
		t.Fatalf("parts must persist as a JSON array, type=%q err=%v", jsonType, err)
	}
	return persisted.Parts
}

func TestAdvisorPostgresMarkActionSerializesParts(t *testing.T) {
	db, store, owner, run, message, actionID := advisorActionJSONFixture(t)
	ctx := context.Background()
	if err := store.MarkAction(ctx, uuid.NewString(), run.ConversationID, run.Generation, actionID, entity.AIProposalApplying, nil); !errors.Is(err, contract.ErrNotFound) {
		t.Fatalf("foreign owner must be rejected: %v", err)
	}
	if err := store.MarkAction(ctx, owner, run.ConversationID, run.Generation+1, actionID, entity.AIProposalApplying, nil); !errors.Is(err, contract.ErrAdvisorEntryConflict) {
		t.Fatalf("stale generation must be rejected: %v", err)
	}
	if err := store.MarkAction(ctx, owner, run.ConversationID, run.Generation, actionID, entity.AIProposalApplying, nil); err != nil {
		t.Fatalf("mark applying must serialize JSON parts: %v", err)
	}
	parts := reloadAdvisorActionJSON(t, db, message, 2)
	_, view, ok := advisorProposalView(parts[1])
	if !ok || view["status"] != entity.AIProposalApplying || view["name"] != "Du lịch" || numberValue(view["allocation_amount"]) != 2000000 {
		t.Fatalf("applying state or proposal payload lost: %+v", view)
	}
	receipt := map[string]any{"resource_id": uuid.NewString(), "action": "create_jar"}
	if err := store.MarkAction(ctx, owner, run.ConversationID, run.Generation, actionID, entity.AIProposalApplied, receipt); err != nil {
		t.Fatal(err)
	}
	// Confirmation replay must not erase the receipt or regress applied state.
	for _, status := range []string{entity.AIProposalApplying, entity.AIProposalApplied} {
		if err := store.MarkAction(ctx, owner, run.ConversationID, run.Generation, actionID, status, nil); err != nil {
			t.Fatal(err)
		}
	}
	parts = reloadAdvisorActionJSON(t, db, message, 2)
	envelope, view, _ := advisorProposalView(parts[1])
	if view["status"] != entity.AIProposalApplied || view["receipt"].(map[string]any)["resource_id"] != receipt["resource_id"] || view["resource_receipt"].(map[string]any)["resource_id"] != receipt["resource_id"] || envelope["warnings"].([]any)[0] != "review_required" {
		t.Fatalf("applied receipt or warnings lost: %+v", envelope)
	}
	if err := store.MarkAction(ctx, owner, run.ConversationID, run.Generation, actionID, entity.AIProposalRejected, nil); !errors.Is(err, contract.ErrAdvisorEntryConflict) {
		t.Fatalf("terminal applied state must not be overwritten: %v", err)
	}
}

func TestAdvisorPostgresReplaceActionSerializesParts(t *testing.T) {
	db, store, owner, run, message, actionID := advisorActionJSONFixture(t)
	ctx := context.Background()
	patch := map[string]any{"allocation_amount": 3000000}
	if _, err := store.ReplaceAction(ctx, uuid.NewString(), run.ConversationID, run.Generation, actionID, 1, patch); !errors.Is(err, contract.ErrNotFound) {
		t.Fatalf("foreign owner must be rejected: %v", err)
	}
	if _, err := store.ReplaceAction(ctx, owner, run.ConversationID, run.Generation+1, actionID, 1, patch); !errors.Is(err, contract.ErrAdvisorEntryConflict) {
		t.Fatalf("stale generation must be rejected: %v", err)
	}
	replacement, err := store.ReplaceAction(ctx, owner, run.ConversationID, run.Generation, actionID, 1, patch)
	if err != nil {
		t.Fatalf("revision must serialize JSON parts: %v", err)
	}
	newID := stringValue(replacement["action_id"])
	if newID == "" || newID == actionID {
		t.Fatalf("replacement must have a new action identity: %+v", replacement)
	}
	parts := reloadAdvisorActionJSON(t, db, message, 3)
	_, old, _ := advisorProposalView(parts[1])
	envelope, current, _ := advisorProposalView(parts[2])
	if old["status"] != entity.AIProposalSuperseded || old["superseded_by"] != newID || numberValue(old["allocation_amount"]) != 2000000 || current["status"] != entity.AIProposalPending || numberValue(current["revision"]) != 2 || numberValue(current["allocation_amount"]) != 3000000 || current["name"] != "Du lịch" || current["month"] != "2026-11" || current["conversation_id"] != run.ConversationID || envelope["warnings"].([]any)[0] != "review_required" {
		t.Fatalf("revision did not preserve old/new payloads: old=%+v new=%+v", old, current)
	}
	if _, err := store.ReplaceAction(ctx, owner, run.ConversationID, run.Generation, actionID, 1, patch); !errors.Is(err, contract.ErrAdvisorEntryConflict) {
		t.Fatalf("superseded action must reject replayed replacement: %v", err)
	}
	if _, err := store.ReplaceAction(ctx, owner, run.ConversationID, run.Generation, newID, 1, patch); !errors.Is(err, contract.ErrAdvisorEntryConflict) {
		t.Fatalf("stale revision must be rejected: %v", err)
	}
	if err := store.MarkAction(ctx, owner, run.ConversationID, run.Generation, actionID, entity.AIProposalApplying, nil); !errors.Is(err, contract.ErrAdvisorEntryConflict) {
		t.Fatalf("superseded action must not be confirmable: %v", err)
	}
	reloadAdvisorActionJSON(t, db, message, 3)
}
