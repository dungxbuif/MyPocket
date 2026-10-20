package sync_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"mypocket/internal/finance"
	"mypocket/internal/portfolio"
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

func TestServiceReturnsConflictWhenVersionChangesAfterPrecheck(t *testing.T) {
	store := newStoreStub()
	key := string(mysync.EntityTransaction) + ":" + fixedTransactionID
	store.entityVersions[key] = versionPayload{version: 1, payload: rawJSON(`{"id":"` + fixedTransactionID + `","version":1}`)}
	financeRepo := &financeStub{updateTransaction: func() (finance.Transaction, error) {
		store.entityVersions[key] = versionPayload{version: 2, payload: rawJSON(`{"id":"` + fixedTransactionID + `","version":2,"note":"other device"}`)}
		return finance.Transaction{}, finance.ErrConflict
	}}
	service := mysync.NewService(store, financeRepo)

	result, err := service.ApplyMutations(context.Background(), fixedUserID, []mysync.Mutation{transactionUpdateMutation("mut_race", 1, 1)})
	if err != nil {
		t.Fatalf("post-precheck version race should be a conflict result, got %v", err)
	}
	if result[0].State != mysync.ResultConflict || result[0].Conflict == nil || result[0].Conflict.ServerVersion != 2 {
		t.Fatalf("unexpected race conflict result: %#v", result[0])
	}
}

func TestServiceAppliesIdempotentCategoryActivationWithoutCategoryVersionConflict(t *testing.T) {
	store := newStoreStub()
	store.entityVersions[string(mysync.EntityCategory)+":"+fixedCategoryID] = versionPayload{version: 7, payload: rawJSON(`{"id":"` + fixedCategoryID + `","version":7}`)}
	financeRepo := &financeStub{}
	service := mysync.NewService(store, financeRepo)

	result, err := service.ApplyMutations(context.Background(), fixedUserID, []mysync.Mutation{{
		MutationID:  "mut_category_active",
		DeviceID:    "device_1",
		Sequence:    1,
		EntityType:  mysync.EntityCategory,
		EntityID:    fixedCategoryID,
		Operation:   mysync.OperationSetCategoryActive,
		BaseVersion: 0,
		Payload:     rawJSON(`{"wallet_id":"` + fixedWalletID + `","active":false}`),
	}})
	if err != nil {
		t.Fatalf("apply category activation: %v", err)
	}
	if result[0].State != mysync.ResultApplied || financeRepo.setCategoryActiveCalls != 1 {
		t.Fatalf("category activation should apply once without comparing category version: result=%#v calls=%d", result[0], financeRepo.setCategoryActiveCalls)
	}
}

func TestServiceAppliesAssetMutations(t *testing.T) {
	store := newStoreStub()
	portfolioRepo := &portfolioStub{
		created: portfolio.Position{ID: fixedAssetID, UserID: fixedUserID, Type: portfolio.AssetGold, Name: "SJC", Unit: "tael", PricingMode: portfolio.PricingManual, IncludeInNetWorth: true, Version: 1},
		priced:  portfolio.Position{ID: fixedAssetID, UserID: fixedUserID, Type: portfolio.AssetGold, Name: "SJC", Unit: "tael", PricingMode: portfolio.PricingManual, IncludeInNetWorth: true, Version: 2},
	}
	service := mysync.NewService(store, &financeStub{}, portfolioRepo)

	createResult, err := service.ApplyMutations(context.Background(), fixedUserID, []mysync.Mutation{{
		MutationID:  "mut_asset_create",
		DeviceID:    "device_1",
		Sequence:    1,
		EntityType:  mysync.EntityAsset,
		EntityID:    fixedAssetID,
		Operation:   mysync.OperationCreate,
		BaseVersion: 0,
		Payload:     rawJSON(`{"type":"gold","name":"SJC","unit":"tael","pricing_mode":"manual","include_in_net_worth":true}`),
	}})
	if err != nil {
		t.Fatalf("apply asset create: %v", err)
	}
	priceResult, err := service.ApplyMutations(context.Background(), fixedUserID, []mysync.Mutation{{
		MutationID:  "mut_asset_price",
		DeviceID:    "device_1",
		Sequence:    2,
		EntityType:  mysync.EntityAsset,
		EntityID:    fixedAssetID,
		Operation:   mysync.OperationAddPrice,
		BaseVersion: 1,
		Payload:     rawJSON(`{"unit_price_vnd":75000000,"priced_at":"2026-08-31T00:00:00Z","source":"manual","base_version":1}`),
	}})
	if err != nil {
		t.Fatalf("apply asset price: %v", err)
	}
	if createResult[0].State != mysync.ResultApplied || priceResult[0].State != mysync.ResultApplied {
		t.Fatalf("unexpected asset sync states: create=%s price=%s", createResult[0].State, priceResult[0].State)
	}
	if portfolioRepo.createCalls != 1 || portfolioRepo.priceCalls != 1 {
		t.Fatalf("asset calls create=%d price=%d", portfolioRepo.createCalls, portfolioRepo.priceCalls)
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
	fixedAssetID       = "00000000-0000-4000-8000-000000000401"
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
	setCategoryActiveCalls int
	createdTransaction     finance.Transaction
	updateTransaction      func() (finance.Transaction, error)
}

func (f *financeStub) ListWallets(context.Context, string) ([]finance.Wallet, error) { return nil, nil }
func (f *financeStub) CreateWallet(context.Context, string, finance.CreateWalletInput) (finance.Wallet, error) {
	return finance.Wallet{ID: fixedWalletID, Version: 1}, nil
}
func (f *financeStub) UpdateWallet(context.Context, string, string, finance.UpdateWalletInput) (finance.Wallet, error) {
	return finance.Wallet{ID: fixedWalletID, Version: 2}, nil
}
func (f *financeStub) ArchiveWallet(context.Context, string, string, int64) error      { return nil }
func (f *financeStub) SetDefaultAIWallet(context.Context, string, string, int64) error { return nil }
func (f *financeStub) ListCategories(context.Context, string) ([]finance.Category, error) {
	return nil, nil
}
func (f *financeStub) CreateCategory(context.Context, string, finance.CreateCategoryInput) (finance.Category, error) {
	return finance.Category{ID: fixedCategoryID, Version: 1}, nil
}
func (f *financeStub) UpdateCategory(context.Context, string, string, finance.UpdateCategoryInput) (finance.Category, error) {
	return finance.Category{ID: fixedCategoryID, Version: 2}, nil
}
func (f *financeStub) ArchiveCategory(context.Context, string, string, int64) error { return nil }
func (f *financeStub) SetWalletCategoryActive(context.Context, string, string, string, bool) error {
	f.setCategoryActiveCalls++
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
	if f.updateTransaction != nil {
		return f.updateTransaction()
	}
	return finance.Transaction{ID: fixedTransactionID, Version: 2}, nil
}
func (f *financeStub) ArchiveTransaction(context.Context, string, string, int64) error { return nil }

type portfolioStub struct {
	createCalls int
	priceCalls  int
	created     portfolio.Position
	priced      portfolio.Position
}

func (p *portfolioStub) ListPositions(context.Context, string, bool) ([]portfolio.Position, error) {
	return []portfolio.Position{p.created}, nil
}
func (p *portfolioStub) CreatePosition(_ context.Context, _ string, input portfolio.CreatePositionInput) (portfolio.Position, error) {
	p.createCalls++
	p.created.ID = input.ID
	return p.created, nil
}
func (p *portfolioStub) GetPosition(context.Context, string, string) (portfolio.Position, error) {
	return p.created, nil
}
func (p *portfolioStub) ArchivePosition(context.Context, string, string, int64) error { return nil }
func (p *portfolioStub) AddTrade(context.Context, string, string, portfolio.AddTradeInput) (portfolio.Position, error) {
	return p.priced, nil
}
func (p *portfolioStub) UpdateTrade(context.Context, string, string, string, portfolio.UpdateTradeInput) (portfolio.Position, error) {
	return p.priced, nil
}
func (p *portfolioStub) ArchiveTrade(context.Context, string, string, string, int64) (portfolio.Position, error) {
	return p.priced, nil
}
func (p *portfolioStub) AddPrice(context.Context, string, string, portfolio.AddPriceInput) (portfolio.Position, error) {
	p.priceCalls++
	return p.priced, nil
}

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
