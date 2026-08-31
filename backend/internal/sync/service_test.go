package sync_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"mypocket/internal/finance"
	mysync "mypocket/internal/sync"
)

func TestServiceAppliesAndReplaysMutationOnce(t *testing.T) {
	store := newStoreStub()
	financeRepo := &financeStub{createdTransaction: finance.Transaction{ID: fixedTransactionID, UserID: fixedUserID, Type: finance.TransactionExpense, SourceWalletID: fixedWalletID, CategoryID: fixedCategoryID, AmountVND: 42000, OccurredAt: fixedTime(), Version: 1}}
	service := mysync.NewService(store, financeRepo)
	mutation := transactionCreateMutation("mut_1", 1, 42000)

	first, err := service.ApplyMutations(context.Background(), fixedUserID, []mysync.Mutation{mutation})
	if err != nil {
		t.Fatalf("apply first: %v", err)
	}
	second, err := service.ApplyMutations(context.Background(), fixedUserID, []mysync.Mutation{mutation})
	if err != nil {
		t.Fatalf("apply replay: %v", err)
	}

	if first[0].State != mysync.ResultApplied || second[0].State != mysync.ResultReplayed {
		t.Fatalf("unexpected states first=%s second=%s", first[0].State, second[0].State)
	}
	if financeRepo.createTransactionCalls != 1 {
		t.Fatalf("expected one accounting apply, got %d", financeRepo.createTransactionCalls)
	}
}

func TestServiceRejectsDuplicateMutationIDWithDifferentHash(t *testing.T) {
	store := newStoreStub()
	financeRepo := &financeStub{createdTransaction: finance.Transaction{ID: fixedTransactionID, UserID: fixedUserID, Type: finance.TransactionExpense, SourceWalletID: fixedWalletID, CategoryID: fixedCategoryID, AmountVND: 1, OccurredAt: fixedTime(), Version: 1}}
	service := mysync.NewService(store, financeRepo)

	if _, err := service.ApplyMutations(context.Background(), fixedUserID, []mysync.Mutation{transactionCreateMutation("mut_1", 1, 1)}); err != nil {
		t.Fatalf("apply original: %v", err)
	}
	result, err := service.ApplyMutations(context.Background(), fixedUserID, []mysync.Mutation{transactionCreateMutation("mut_1", 1, 2)})
	if err != nil {
		t.Fatalf("apply duplicate: %v", err)
	}

	if result[0].State != mysync.ResultRejected || result[0].Reason == "" {
		t.Fatalf("expected rejected duplicate hash result, got %#v", result[0])
	}
	if financeRepo.createTransactionCalls != 1 {
		t.Fatalf("duplicate hash should not apply accounting again, got %d calls", financeRepo.createTransactionCalls)
	}
}

func TestServiceReturnsConflictForStaleBaseVersion(t *testing.T) {
	store := newStoreStub()
	store.entityVersions[string(mysync.EntityTransaction)+":"+fixedTransactionID] = versionPayload{version: 3, payload: rawJSON(`{"id":"` + fixedTransactionID + `","version":3,"note":"server"}`)}
	service := mysync.NewService(store, &financeStub{})
	mutation := transactionUpdateMutation("mut_2", 2, 1)

	result, err := service.ApplyMutations(context.Background(), fixedUserID, []mysync.Mutation{mutation})
	if err != nil {
		t.Fatalf("apply stale update: %v", err)
	}

	if result[0].State != mysync.ResultConflict || result[0].Conflict == nil {
		t.Fatalf("expected conflict result, got %#v", result[0])
	}
	if result[0].Conflict.ServerVersion != 3 || result[0].Conflict.BaseVersion != 1 {
		t.Fatalf("wrong conflict versions: %#v", result[0].Conflict)
	}
}

func TestServiceValidatesOrderedBatch(t *testing.T) {
	service := mysync.NewService(newStoreStub(), &financeStub{})
	_, err := service.ApplyMutations(context.Background(), fixedUserID, []mysync.Mutation{
		transactionCreateMutation("mut_2", 2, 1),
		transactionCreateMutation("mut_1", 1, 1),
	})

	if !errors.Is(err, mysync.ErrValidation) {
		t.Fatalf("expected validation error for unordered batch, got %v", err)
	}
}

const (
	fixedUserID        = "00000000-0000-4000-8000-000000000001"
	fixedWalletID      = "00000000-0000-4000-8000-000000000101"
	fixedCategoryID    = "00000000-0000-4000-8000-000000000201"
	fixedTransactionID = "00000000-0000-4000-8000-000000000301"
)

type storeStub struct {
	mutations      map[string]storedMutation
	entityVersions map[string]versionPayload
}

type storedMutation struct {
	hash   string
	result mysync.MutationResult
}

type versionPayload struct {
	version int64
	payload json.RawMessage
}

func newStoreStub() *storeStub {
	return &storeStub{mutations: map[string]storedMutation{}, entityVersions: map[string]versionPayload{}}
}

func (s *storeStub) LoadMutationResult(_ context.Context, _ string, mutationID string) (string, mysync.MutationResult, bool, error) {
	stored, ok := s.mutations[mutationID]
	return stored.hash, stored.result, ok, nil
}

func (s *storeStub) StoreMutationResult(_ context.Context, _ string, mutationID string, requestHash string, result mysync.MutationResult) error {
	s.mutations[mutationID] = storedMutation{hash: requestHash, result: result}
	return nil
}

func (s *storeStub) AppendChange(_ context.Context, _ string, entityType mysync.EntityType, entityID string, operation mysync.Operation, version int64, payload json.RawMessage) (mysync.Change, error) {
	return mysync.Change{Cursor: int64(len(s.mutations) + 1), EntityType: entityType, EntityID: entityID, Operation: operation, Version: version, Payload: payload, CreatedAt: fixedTime()}, nil
}

func (s *storeStub) ListChanges(_ context.Context, _ string, after int64, _ int) (mysync.ChangesResult, error) {
	return mysync.ChangesResult{NextCursor: after}, nil
}

func (s *storeStub) CurrentCursor(context.Context, string) (int64, error) {
	return 0, nil
}

func (s *storeStub) EntityVersionAndPayload(_ context.Context, _ string, entityType mysync.EntityType, entityID string) (int64, json.RawMessage, bool, error) {
	record, ok := s.entityVersions[string(entityType)+":"+entityID]
	return record.version, record.payload, ok, nil
}

type financeStub struct {
	createTransactionCalls int
	createdTransaction     finance.Transaction
}

func (f *financeStub) ListWallets(context.Context, string) ([]finance.Wallet, error) { return nil, nil }
func (f *financeStub) CreateWallet(context.Context, string, finance.CreateWalletInput) (finance.Wallet, error) {
	return finance.Wallet{ID: fixedWalletID, Version: 1}, nil
}
func (f *financeStub) UpdateWallet(context.Context, string, string, finance.UpdateWalletInput) (finance.Wallet, error) {
	return finance.Wallet{ID: fixedWalletID, Version: 2}, nil
}
func (f *financeStub) ArchiveWallet(context.Context, string, string) error      { return nil }
func (f *financeStub) SetDefaultAIWallet(context.Context, string, string) error { return nil }
func (f *financeStub) ListCategories(context.Context, string) ([]finance.Category, error) {
	return nil, nil
}
func (f *financeStub) CreateCategory(context.Context, string, finance.CreateCategoryInput) (finance.Category, error) {
	return finance.Category{ID: fixedCategoryID, Version: 1}, nil
}
func (f *financeStub) UpdateCategory(context.Context, string, string, finance.UpdateCategoryInput) (finance.Category, error) {
	return finance.Category{ID: fixedCategoryID, Version: 2}, nil
}
func (f *financeStub) ArchiveCategory(context.Context, string, string) error { return nil }
func (f *financeStub) SetWalletCategoryActive(context.Context, string, string, string, bool) error {
	return nil
}
func (f *financeStub) ListTransactions(context.Context, string, finance.TransactionFilters) ([]finance.Transaction, error) {
	return nil, nil
}
func (f *financeStub) CreateTransaction(_ context.Context, _ string, input finance.CreateTransactionInput) (finance.Transaction, error) {
	f.createTransactionCalls++
	f.createdTransaction.ID = input.ID
	return f.createdTransaction, nil
}
func (f *financeStub) UpdateTransaction(context.Context, string, string, finance.UpdateTransactionInput) (finance.Transaction, error) {
	return finance.Transaction{ID: fixedTransactionID, Version: 2}, nil
}
func (f *financeStub) ArchiveTransaction(context.Context, string, string) error { return nil }

func transactionCreateMutation(mutationID string, sequence int64, amount int64) mysync.Mutation {
	return mysync.Mutation{
		MutationID:  mutationID,
		DeviceID:    "device_1",
		Sequence:    sequence,
		EntityType:  mysync.EntityTransaction,
		EntityID:    fixedTransactionID,
		Operation:   mysync.OperationCreate,
		BaseVersion: 0,
		Payload:     rawJSON(`{"type":"expense","source_wallet_id":"` + fixedWalletID + `","category_id":"` + fixedCategoryID + `","amount_vnd":` + intString(amount) + `,"occurred_at":"2026-08-31T00:00:00Z"}`),
	}
}

func transactionUpdateMutation(mutationID string, sequence int64, baseVersion int64) mysync.Mutation {
	mutation := transactionCreateMutation(mutationID, sequence, 42000)
	mutation.Operation = mysync.OperationUpdate
	mutation.BaseVersion = baseVersion
	return mutation
}

func rawJSON(value string) json.RawMessage {
	return json.RawMessage(value)
}

func intString(value int64) string {
	body, _ := json.Marshal(value)
	return string(body)
}

func fixedTime() time.Time {
	return time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)
}
