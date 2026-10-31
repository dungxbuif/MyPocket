package entity

import (
	"encoding/json"
	"testing"
)

func validAdvisorEntryTask() AdvisorEntryTask {
	return AdvisorEntryTask{
		ID: "task-1", OwnerID: "owner-1", ConversationID: "conversation-1",
		Generation: 2, ClientRequestID: "request-1", PayloadHash: "sha256:abc",
		SessionID: "session-1", Status: AdvisorEntryTaskAccepted,
	}
}

func TestAdvisorEntryTaskRequiresOwnerGenerationRequestAndHash(t *testing.T) {
	base := validAdvisorEntryTask()
	if err := base.Validate(); err != nil {
		t.Fatalf("valid task: %v", err)
	}
	for name, mutate := range map[string]func(*AdvisorEntryTask){
		"owner":        func(task *AdvisorEntryTask) { task.OwnerID = "" },
		"conversation": func(task *AdvisorEntryTask) { task.ConversationID = "" },
		"generation":   func(task *AdvisorEntryTask) { task.Generation = 0 },
		"request":      func(task *AdvisorEntryTask) { task.ClientRequestID = "" },
		"hash":         func(task *AdvisorEntryTask) { task.PayloadHash = "" },
		"session":      func(task *AdvisorEntryTask) { task.SessionID = "" },
	} {
		t.Run(name, func(t *testing.T) {
			task := base
			mutate(&task)
			if err := task.Validate(); err == nil {
				t.Fatal("accepted missing identity or payload binding")
			}
		})
	}
}

func TestAdvisorEntryTaskTransitions(t *testing.T) {
	for _, tc := range []struct {
		from, to string
		allowed  bool
	}{
		{AdvisorEntryTaskAccepted, AdvisorEntryTaskProcessing, true},
		{AdvisorEntryTaskAccepted, AdvisorEntryTaskCancelled, true},
		{AdvisorEntryTaskProcessing, AdvisorEntryTaskReviewReady, true},
		{AdvisorEntryTaskProcessing, AdvisorEntryTaskFailed, true},
		{AdvisorEntryTaskProcessing, AdvisorEntryTaskCancelled, true},
		{AdvisorEntryTaskReviewReady, AdvisorEntryTaskProcessing, false},
		{AdvisorEntryTaskFailed, AdvisorEntryTaskProcessing, false},
		{AdvisorEntryTaskCancelled, AdvisorEntryTaskReviewReady, false},
		{AdvisorEntryTaskAccepted, AdvisorEntryTaskReviewReady, false},
	} {
		if got := CanTransitionAdvisorEntryTask(tc.from, tc.to); got != tc.allowed {
			t.Errorf("%s -> %s = %v, want %v", tc.from, tc.to, got, tc.allowed)
		}
	}
}

func TestAdvisorEntryTaskBatchVersionMustBePositive(t *testing.T) {
	task := validAdvisorEntryTask()
	task.BatchID = "batch-1"
	if err := task.Validate(); err == nil {
		t.Fatal("accepted batch without positive version")
	}
	task.BatchVersion = 1
	if err := task.Validate(); err != nil {
		t.Fatalf("positive version: %v", err)
	}
}

func TestAdvisorEntryBatchPartExposesReferencesWithoutDraft(t *testing.T) {
	part := AdvisorEntryBatchPart{TaskID: "task-1", BatchID: "batch-1", Version: 1, ItemIDs: []string{"item-1"}, ItemCount: 1, Summary: "One candidate"}
	encoded, err := json.Marshal(part)
	if err != nil {
		t.Fatal(err)
	}
	var object map[string]any
	if err := json.Unmarshal(encoded, &object); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"task_id", "batch_id", "version", "item_ids", "item_count", "summary"} {
		if _, ok := object[key]; !ok {
			t.Errorf("missing %s", key)
		}
	}
	if _, ok := object["draft"]; ok {
		t.Fatal("entry batch exposed authoritative draft")
	}
}
