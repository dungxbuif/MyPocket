package entity

import (
	"encoding/json"
	"testing"
	"time"
)

func TestAdvisorMessageRejectsOversizedUserPart(t *testing.T) {
	message := AdvisorMessage{Role: AdvisorRoleUser, Parts: []AdvisorPart{{Type: "text", Text: string(make([]byte, 8193))}}}
	if err := message.Validate(); err == nil {
		t.Fatal("oversized user message must be rejected")
	}
}

func TestAdvisorRunAllowsOnlyKnownTransitions(t *testing.T) {
	if !CanTransitionAdvisorRun(AdvisorStatusQueued, AdvisorStatusRunning) {
		t.Fatal("queued run should be claimable")
	}
	if CanTransitionAdvisorRun(AdvisorStatusCompleted, AdvisorStatusRunning) {
		t.Fatal("terminal run must not be restarted")
	}
}

func TestAdvisorMessageRejectsOversizedPartData(t *testing.T) {
	message := AdvisorMessage{Role: AdvisorRoleAssistant, Parts: []AdvisorPart{{Type: "metric_group", Data: string(make([]byte, 64*1024+1))}}}
	if err := message.Validate(); err == nil {
		t.Fatal("oversized structured part must be rejected")
	}
}

func TestAdvisorConversationIdleExpiryUsesServerBoundary(t *testing.T) {
	base := time.Date(2026, 10, 15, 10, 0, 0, 0, time.UTC)
	conversation := AdvisorConversation{LastUserMessageAt: &base}
	if conversation.IsIdleExpired(base.Add(30*time.Minute - time.Nanosecond)) {
		t.Fatal("conversation must remain active just before the 30-minute boundary")
	}
	if !conversation.IsIdleExpired(base.Add(30 * time.Minute)) {
		t.Fatal("conversation must expire at the inclusive 30-minute boundary")
	}
	if conversation.IsIdleExpired(base.Add(29 * time.Minute)) {
		t.Fatal("conversation must not expire while active")
	}
}

func TestAdvisorConversationWithoutUserInputHasNoIdleDeadline(t *testing.T) {
	conversation := AdvisorConversation{}
	if conversation.IsIdleExpired(time.Now().UTC().Add(24 * time.Hour)) {
		t.Fatal("empty conversation must not expire")
	}
	if conversation.IdleExpiresAt() != nil {
		t.Fatal("empty conversation must not expose an expiry")
	}
}

func TestAdvisorHistoryMetaUsesServerExpiryWithoutExtendingOnRead(t *testing.T) {
	lastUser := time.Date(2026, 10, 15, 10, 0, 0, 0, time.UTC)
	serverTime := lastUser.Add(5 * time.Minute)
	meta := (AdvisorConversation{ID: "conversation-1", Generation: 3, LastUserMessageAt: &lastUser}).HistoryMeta(serverTime)
	if meta.ConversationID != "conversation-1" || meta.Generation != 3 {
		t.Fatalf("unexpected history identity: %+v", meta)
	}
	if meta.LastUserMessageAt == nil || !meta.LastUserMessageAt.Equal(lastUser) {
		t.Fatalf("last user timestamp changed: %+v", meta.LastUserMessageAt)
	}
	if meta.ExpiresAt == nil || !meta.ExpiresAt.Equal(lastUser.Add(AdvisorIdleTTL)) {
		t.Fatalf("unexpected expiry: %+v", meta.ExpiresAt)
	}
	if !meta.ServerTime.Equal(serverTime) {
		t.Fatalf("unexpected server time: %v", meta.ServerTime)
	}
}

func TestAdvisorResetResultDoesNotExposeInternalPurgedRunIDs(t *testing.T) {
	payload, err := json.Marshal(AdvisorResetResult{ConversationID: "conversation-1", Generation: 2, PurgedRunIDs: []string{"run-private"}})
	if err != nil {
		t.Fatal(err)
	}
	if string(payload) == "" || string(payload) == "null" {
		t.Fatal("reset result must remain a JSON object")
	}
	if string(payload) == "" || containsJSONField(payload, "PurgedRunIDs") || containsJSONField(payload, "purged_run_ids") {
		t.Fatalf("internal purged run IDs must not cross the API boundary: %s", payload)
	}
}

func containsJSONField(payload []byte, field string) bool {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(payload, &object); err != nil {
		return false
	}
	_, ok := object[field]
	return ok
}
