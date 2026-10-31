package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
	"github.com/mypocket/backend/internal/usecase"
)

type advisorEntryWorkerTaskFake struct {
	repository.AdvisorEntryTaskRepository
	task       *entity.AdvisorEntryTask
	lease      repository.AdvisorEntryTaskLease
	claimCalls int
	completed  int
	failed     string
}

func (f *advisorEntryWorkerTaskFake) ClaimNextTask(context.Context, string) (repository.AdvisorEntryTaskLease, error) {
	f.claimCalls++
	if f.claimCalls > 1 {
		return repository.AdvisorEntryTaskLease{}, repository.ErrNotFound
	}
	return f.lease, nil
}

func (f *advisorEntryWorkerTaskFake) GetTask(_ context.Context, ownerID, taskID string) (*entity.AdvisorEntryTask, error) {
	if f.task == nil || f.task.OwnerID != ownerID || f.task.ID != taskID {
		return nil, repository.ErrNotFound
	}
	return f.task, nil
}

func (*advisorEntryWorkerTaskFake) HeartbeatTask(context.Context, repository.AdvisorEntryTaskLease) error {
	return nil
}

func (f *advisorEntryWorkerTaskFake) CompleteTask(_ context.Context, _ repository.AdvisorEntryTaskLease, batch entity.AdvisorEntryBatchPart) error {
	f.completed++
	f.task.Status = entity.AdvisorEntryTaskReviewReady
	f.task.BatchID = batch.BatchID
	f.task.BatchVersion = batch.Version
	return nil
}

func (f *advisorEntryWorkerTaskFake) FailTask(_ context.Context, _ repository.AdvisorEntryTaskLease, code string) error {
	f.failed = code
	return nil
}

type advisorEntryWorkerAIEntryFake struct {
	repository.AIEntryRepository
	session *entity.AIEntrySession
}

func (f *advisorEntryWorkerAIEntryFake) Session(_ context.Context, ownerID, sessionID string) (*entity.AIEntrySession, error) {
	if f.session == nil || f.session.OwnerID != ownerID || f.session.ID != sessionID {
		return nil, repository.ErrNotFound
	}
	return f.session, nil
}

func (f *advisorEntryWorkerAIEntryFake) ListAttachments(context.Context, string, string) ([]entity.AIEntryAttachment, error) {
	return []entity.AIEntryAttachment{}, nil
}

func (f *advisorEntryWorkerAIEntryFake) BeginMessage(context.Context, string, string, string, string, string) (string, bool, error) {
	return "worker-token", true, nil
}

func (*advisorEntryWorkerAIEntryFake) UpdateAttachmentOCR(context.Context, string, string, []string, bool) error {
	return nil
}

func (f *advisorEntryWorkerAIEntryFake) FinishMessage(_ context.Context, _, _, token string, output entity.AIExtractOutput) error {
	if token != "worker-token" {
		return repository.ErrAIConflict
	}
	f.session.Proposals = []entity.AIEntryProposal{{ID: "proposal-1", SessionID: f.session.ID, Version: 1, Status: entity.AIProposalPending, Draft: output.Drafts[0].AIEntryDraft}}
	return nil
}

type advisorEntryWorkerWallets struct{ repository.WalletRepository }

func (advisorEntryWorkerWallets) List(string) ([]entity.Wallet, error) {
	return []entity.Wallet{{ID: "wallet-1", OwnerID: "owner-1", Type: entity.WalletTypeBasic}}, nil
}

type advisorEntryWorkerCategories struct{ repository.CategoryRepository }

func (advisorEntryWorkerCategories) ListVisible(string) ([]entity.Category, error) {
	return []entity.Category{}, nil
}

type advisorEntryWorkerExtractor struct{}

func (advisorEntryWorkerExtractor) Configured() bool                     { return true }
func (advisorEntryWorkerExtractor) OCRConfigured() bool                  { return true }
func (advisorEntryWorkerExtractor) ValidateFiles([]entity.AIImage) error { return nil }
func (advisorEntryWorkerExtractor) Extract(context.Context, entity.AIExtractInput) (entity.AIExtractOutput, error) {
	return entity.AIExtractOutput{Drafts: []entity.AIExtractDraft{{AIEntryDraft: entity.AIEntryDraft{
		Type: entity.TransactionTypeExpense, Amount: 35000, WalletID: "wallet-1", OccurredAt: "2026-11-01T08:00:00Z",
	}}}}, nil
}

type advisorEntryWorkerStorage struct{}

func (advisorEntryWorkerStorage) Key(string, string, string, string) (string, error) { return "", nil }
func (advisorEntryWorkerStorage) Put(context.Context, string, string, []byte) error  { return nil }
func (advisorEntryWorkerStorage) Delete(context.Context, string) error               { return nil }
func (advisorEntryWorkerStorage) SignedGet(context.Context, string) (string, error)  { return "", nil }
func (advisorEntryWorkerStorage) Read(context.Context, string, int64) ([]byte, error) {
	return nil, errors.New("unexpected private object read")
}

func TestAdvisorEntryWorkerClaimsProcessesAndCompletesOneTask(t *testing.T) {
	requestID := "00000000-0000-4000-8000-000000000002"
	task := &entity.AdvisorEntryTask{
		ID: "task-1", OwnerID: "owner-1", ConversationID: "conversation-1", Generation: 2,
		ClientRequestID: requestID, SessionID: "session-1", InputText: "coffee", Timezone: "UTC", Status: entity.AdvisorEntryTaskProcessing,
	}
	lease := repository.AdvisorEntryTaskLease{OwnerID: "owner-1", ConversationID: "conversation-1", Generation: 2, TaskID: "task-1", Token: "lease-1", LeaseUntil: time.Now().Add(time.Minute)}
	tasks := &advisorEntryWorkerTaskFake{task: task, lease: lease}
	entries := &advisorEntryWorkerAIEntryFake{session: &entity.AIEntrySession{ID: "session-1", OwnerID: "owner-1"}}
	service := &usecase.AdvisorEntryService{Tasks: tasks, Entries: &usecase.AIEntryService{
		Entries: entries, Wallets: advisorEntryWorkerWallets{}, Categories: advisorEntryWorkerCategories{},
		Extractor: advisorEntryWorkerExtractor{}, Storage: advisorEntryWorkerStorage{},
	}}

	processed, err := (&advisorEntryWorker{Tasks: tasks, Service: service}).Run(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if processed != 1 || tasks.claimCalls != 1 || tasks.completed != 1 || task.Status != entity.AdvisorEntryTaskReviewReady {
		t.Fatalf("worker did not complete claimed task: processed=%d claims=%d completed=%d task=%+v", processed, tasks.claimCalls, tasks.completed, task)
	}
	if tasks.failed != "" {
		t.Fatalf("worker marked successful task failed: %q", tasks.failed)
	}
}
