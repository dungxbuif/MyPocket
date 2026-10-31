package usecase

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
)

type advisorEntryTaskRepositoryFake struct {
	repository.AdvisorEntryTaskRepository
	task        *entity.AdvisorEntryTask
	createCalls int
	completed   *entity.AdvisorEntryBatchPart
	failed      string
	listEvents  []entity.AdvisorEntryTaskEvent
}

func (f *advisorEntryTaskRepositoryFake) GetTaskByRequest(_ context.Context, ownerID, conversationID string, generation int64, requestID string) (*entity.AdvisorEntryTask, error) {
	if f.task == nil || f.task.OwnerID != ownerID || f.task.ConversationID != conversationID || f.task.Generation != generation || f.task.ClientRequestID != requestID {
		return nil, repository.ErrNotFound
	}
	return f.task, nil
}

func (f *advisorEntryTaskRepositoryFake) CreateTask(_ context.Context, task entity.AdvisorEntryTask) (*entity.AdvisorEntryTask, error) {
	f.createCalls++
	f.task = &task
	return f.task, nil
}

func (f *advisorEntryTaskRepositoryFake) GetTask(_ context.Context, ownerID, taskID string) (*entity.AdvisorEntryTask, error) {
	if f.task == nil || f.task.OwnerID != ownerID || f.task.ID != taskID {
		return nil, repository.ErrNotFound
	}
	return f.task, nil
}

func (f *advisorEntryTaskRepositoryFake) CompleteTask(_ context.Context, _ repository.AdvisorEntryTaskLease, batch entity.AdvisorEntryBatchPart) error {
	f.completed = &batch
	if f.task != nil {
		f.task.Status = entity.AdvisorEntryTaskReviewReady
		f.task.BatchID = batch.BatchID
		f.task.BatchVersion = batch.Version
	}
	return nil
}

func (f *advisorEntryTaskRepositoryFake) FailTask(_ context.Context, _ repository.AdvisorEntryTaskLease, code string) error {
	f.failed = code
	return nil
}

func (f *advisorEntryTaskRepositoryFake) ListEvents(context.Context, string, string, int64, int) ([]entity.AdvisorEntryTaskEvent, error) {
	return f.listEvents, nil
}

type advisorEntryAIRepositoryFake struct {
	repository.AIEntryRepository
	session       *entity.AIEntrySession
	attachments   []entity.AIEntryAttachment
	beginCalls    int
	finishCalls   int
	decideCalls   int
	decideOwner   string
	decideID      string
	decideVersion int
	decideApprove bool
}

func (f *advisorEntryAIRepositoryFake) CreateProcess(context.Context, string, string) error {
	return nil
}

func (f *advisorEntryAIRepositoryFake) ListAttachments(context.Context, string, string) ([]entity.AIEntryAttachment, error) {
	return append([]entity.AIEntryAttachment(nil), f.attachments...), nil
}

func (f *advisorEntryAIRepositoryFake) Session(_ context.Context, ownerID, sessionID string) (*entity.AIEntrySession, error) {
	if f.session == nil || f.session.OwnerID != "" && f.session.OwnerID != ownerID || f.session.ID != sessionID {
		return nil, repository.ErrNotFound
	}
	return f.session, nil
}

func (f *advisorEntryAIRepositoryFake) BeginMessage(context.Context, string, string, string, string, string) (string, bool, error) {
	f.beginCalls++
	return "entry-token", true, nil
}

func (f *advisorEntryAIRepositoryFake) UpdateAttachmentOCR(context.Context, string, string, []string, bool) error {
	return nil
}

func (f *advisorEntryAIRepositoryFake) FinishMessage(_ context.Context, _, _, token string, output entity.AIExtractOutput) error {
	if token != "entry-token" {
		return repository.ErrAIConflict
	}
	f.finishCalls++
	f.session.Proposals = make([]entity.AIEntryProposal, 0, len(output.Drafts))
	for index, draft := range output.Drafts {
		f.session.Proposals = append(f.session.Proposals, entity.AIEntryProposal{
			ID: "proposal-" + string(rune('1'+index)), SessionID: f.session.ID, Version: 1,
			Status: entity.AIProposalPending, Draft: draft.AIEntryDraft, Questions: draft.Questions,
		})
	}
	f.session.Processing = false
	return nil
}

func (f *advisorEntryAIRepositoryFake) DecideProposal(_ context.Context, ownerID, proposalID string, version int, approve bool) (*entity.AIEntryProposal, error) {
	f.decideCalls++
	f.decideOwner, f.decideID, f.decideVersion, f.decideApprove = ownerID, proposalID, version, approve
	for index := range f.session.Proposals {
		proposal := &f.session.Proposals[index]
		if proposal.ID == proposalID {
			if proposal.Version != version {
				return nil, repository.ErrAIConflict
			}
			if approve {
				proposal.Status = entity.AIProposalApplied
			} else {
				proposal.Status = entity.AIProposalRejected
			}
			return proposal, nil
		}
	}
	return nil, repository.ErrNotFound
}

type advisorEntryStorageFake struct {
	data      map[string][]byte
	readCalls int
	putCalls  int
}

func (f *advisorEntryStorageFake) Key(_, _, attachment, filename string) (string, error) {
	return attachment + "/" + filename, nil
}
func (f *advisorEntryStorageFake) Put(context.Context, string, string, []byte) error {
	f.putCalls++
	return nil
}
func (*advisorEntryStorageFake) Delete(context.Context, string) error              { return nil }
func (*advisorEntryStorageFake) SignedGet(context.Context, string) (string, error) { return "", nil }
func (f *advisorEntryStorageFake) Read(_ context.Context, key string, _ int64) ([]byte, error) {
	f.readCalls++
	data, ok := f.data[key]
	if !ok {
		return nil, errors.New("missing private object")
	}
	return data, nil
}

func TestAdvisorEntryAcceptIsIdempotentAndPayloadFenced(t *testing.T) {
	tasks := &advisorEntryTaskRepositoryFake{}
	service := &AdvisorEntryService{
		Tasks:   tasks,
		Entries: &AIEntryService{Entries: &entryRepoStub{}},
		Now:     func() time.Time { return time.Date(2026, 11, 1, 1, 2, 3, 0, time.UTC) },
	}
	input := AdvisorEntryInput{ClientRequestID: "request-1", ConversationID: "conversation-1", Generation: 3, Text: "  ăn sáng  ", Timezone: "UTC"}
	principal := Principal{OwnerID: "owner-1", Timezone: "Asia/Ho_Chi_Minh"}

	first, err := service.Accept(context.Background(), principal, input)
	if err != nil {
		t.Fatal(err)
	}
	if first.InputText != "ăn sáng" || first.Timezone != "UTC" || first.Status != entity.AdvisorEntryTaskAccepted {
		t.Fatalf("accepted task did not retain durable private input: %+v", first)
	}

	replay, err := service.Accept(context.Background(), principal, input)
	if err != nil {
		t.Fatal(err)
	}
	if replay.ID != first.ID || tasks.createCalls != 1 {
		t.Fatalf("same request was not idempotent: first=%s replay=%s creates=%d", first.ID, replay.ID, tasks.createCalls)
	}

	input.Text = "khác nội dung"
	if _, err = service.Accept(context.Background(), principal, input); !errors.Is(err, repository.ErrAdvisorEntryConflict) {
		t.Fatalf("changed payload error = %v, want ErrAdvisorEntryConflict", err)
	}
}

func TestAdvisorEntryProcessClaimedReadsPrivateAttachmentsWithoutReupload(t *testing.T) {
	requestID := "00000000-0000-4000-8000-000000000001"
	tasks := &advisorEntryTaskRepositoryFake{task: &entity.AdvisorEntryTask{
		ID: "task-1", OwnerID: "owner-1", ConversationID: "conversation-1", Generation: 3,
		ClientRequestID: requestID, SessionID: "session-1", InputText: "receipt", Timezone: "UTC",
		Status: entity.AdvisorEntryTaskProcessing,
	}}
	entries := &advisorEntryAIRepositoryFake{
		session:     &entity.AIEntrySession{ID: "session-1", OwnerID: "owner-1", Messages: []entity.AIEntryMessage{}},
		attachments: []entity.AIEntryAttachment{{ID: "attachment-1", OwnerID: "owner-1", ProcessID: "session-1", ObjectKey: "private/receipt.pdf", Filename: "receipt.pdf", MIMEType: "application/pdf", OCRStatus: "pending"}},
	}
	storage := &advisorEntryStorageFake{data: map[string][]byte{"private/receipt.pdf": []byte("private receipt bytes")}}
	seen := entity.AIExtractInput{}
	service := &AdvisorEntryService{
		Tasks: tasks,
		Entries: &AIEntryService{
			Entries: entries, Wallets: entryWallets{}, Categories: entryCategories{}, Storage: storage,
			Extractor: entryExtractor{configured: true, ocrConfigured: true, seen: &seen, out: entity.AIExtractOutput{
				AttachmentTexts: []string{"receipt evidence"}, OCRComplete: true,
				Drafts: []entity.AIExtractDraft{{AIEntryDraft: entity.AIEntryDraft{Type: entity.TransactionTypeExpense, Amount: 35000, WalletID: "w", OccurredAt: "2026-11-01T08:00:00Z"}}},
			}},
		},
	}
	lease := repository.AdvisorEntryTaskLease{OwnerID: "owner-1", ConversationID: "conversation-1", Generation: 3, TaskID: "task-1", Token: "lease-1", LeaseUntil: time.Now().Add(time.Minute)}
	if err := service.ProcessClaimed(context.Background(), lease); err != nil {
		t.Fatal(err)
	}
	if storage.readCalls != 1 || storage.putCalls != 0 {
		t.Fatalf("recovery must read existing private input once without re-upload: reads=%d puts=%d", storage.readCalls, storage.putCalls)
	}
	if len(seen.Images) != 1 || seen.Images[0].AttachmentID != "attachment-1" {
		t.Fatalf("worker did not pass stored attachment identity to extractor: %+v", seen.Images)
	}
	if entries.beginCalls != 1 || entries.finishCalls != 1 || tasks.completed == nil || !reflect.DeepEqual(tasks.completed.ItemIDs, []string{"proposal-1"}) {
		t.Fatalf("recovery did not complete authoritative batch: begin=%d finish=%d batch=%+v", entries.beginCalls, entries.finishCalls, tasks.completed)
	}
	if tasks.failed != "" {
		t.Fatalf("successful recovery was marked failed: %q", tasks.failed)
	}
}

func TestAdvisorEntryBatchHydrationAndDecisionUseAuthoritativeProposalRows(t *testing.T) {
	transactionID := "transaction-1"
	entries := &advisorEntryAIRepositoryFake{session: &entity.AIEntrySession{
		ID: "batch-1", OwnerID: "owner-1", Proposals: []entity.AIEntryProposal{
			{ID: "proposal-1", Version: 1, Status: entity.AIProposalPending, Draft: entity.AIEntryDraft{Type: entity.TransactionTypeExpense, Amount: 12000, WalletID: "w"}},
			{ID: "proposal-2", Version: 2, Status: entity.AIProposalApproved, Draft: entity.AIEntryDraft{Type: entity.TransactionTypeIncome, Amount: 50000, WalletID: "w"}, TransactionID: &transactionID},
			{ID: "proposal-3", Version: 1, Status: entity.AIProposalRejected, Draft: entity.AIEntryDraft{Type: entity.TransactionTypeExpense, Amount: 7000, WalletID: "w"}},
		},
	}}
	tasks := &advisorEntryTaskRepositoryFake{task: &entity.AdvisorEntryTask{
		ID: "task-1", OwnerID: "owner-1", ConversationID: "conversation-1", Generation: 3,
		SessionID: "batch-1", BatchID: "batch-1", BatchVersion: 1, Status: entity.AdvisorEntryTaskReviewReady,
	}}
	service := &AdvisorEntryService{Tasks: tasks, Entries: &AIEntryService{Entries: entries}}

	view, err := service.Entries.Batch(context.Background(), "owner-1", "batch-1", 1)
	if err != nil {
		t.Fatal(err)
	}
	if got := []string{view.Items[0].Status, view.Items[1].Status, view.Items[2].Status}; !reflect.DeepEqual(got, []string{entity.AIProposalPending, entity.AIProposalApplied, entity.AIProposalRejected}) {
		t.Fatalf("batch did not normalize authoritative statuses: %v", got)
	}
	if view.Items[1].ResourceID != transactionID || view.Items[1].ResourceReceipt["resource_id"] != transactionID {
		t.Fatalf("applied item did not expose resource receipt: %+v", view.Items[1])
	}
	if _, err := service.Entries.Batch(context.Background(), "owner-1", "batch-1", 2); !errors.Is(err, repository.ErrAIConflict) {
		t.Fatalf("unsupported batch version error = %v, want ErrAIConflict", err)
	}

	original := []entity.AdvisorMessage{{Parts: []entity.AdvisorPart{{Type: "entry_batch", Data: map[string]any{"batch_id": "batch-1", "version": float64(1)}}}}}
	hydrated, err := service.Entries.HydrateAdvisorMessages(context.Background(), "owner-1", original)
	if err != nil {
		t.Fatal(err)
	}
	data := hydrated[0].Parts[0].Data.(map[string]any)
	if _, ok := data["batch"]; !ok || data["batch"].(*entity.AdvisorEntryBatchView).Items[1].Status != entity.AIProposalApplied {
		t.Fatalf("hydrated entry batch missing authoritative view: %#v", data)
	}
	if _, mutated := original[0].Parts[0].Data.(map[string]any)["batch"]; mutated {
		t.Fatal("hydration mutated the caller's persisted-history value")
	}

	result, err := service.Decide(context.Background(), "owner-1", "task-1", "batch-1", 1, "proposal-1", 1, "apply")
	if err != nil {
		t.Fatal(err)
	}
	if entries.decideCalls != 1 || entries.decideOwner != "owner-1" || entries.decideID != "proposal-1" || entries.decideVersion != 1 || !entries.decideApprove {
		t.Fatalf("decision did not use owner/version fenced proposal mutation: calls=%d owner=%s id=%s version=%d approve=%v", entries.decideCalls, entries.decideOwner, entries.decideID, entries.decideVersion, entries.decideApprove)
	}
	if result.Batch == nil || result.Batch.Items[0].Status != entity.AIProposalApplied {
		t.Fatalf("decision response did not rehydrate updated authoritative batch: %+v", result.Batch)
	}
	if _, err := service.Decide(context.Background(), "owner-1", "task-1", "batch-1", 1, "proposal-1", 2, "apply"); !errors.Is(err, repository.ErrAIConflict) {
		t.Fatalf("stale item version error = %v, want ErrAIConflict", err)
	}
}
