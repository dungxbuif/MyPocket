package main

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/mypocket/backend/internal/entity"
)

type advisorExpiryRunnerStub struct {
	limit int
	now   time.Time
}

type attachmentCleanupRunnerStub struct {
	limit int
	n     int
	err   error
}

func (s *attachmentCleanupRunnerStub) Run(_ context.Context, limit int) (int, error) {
	s.limit = limit
	return s.n, s.err
}

func (s *advisorExpiryRunnerStub) ExpireIdleBatch(_ context.Context, now time.Time, limit int) (int, error) {
	s.now = now
	s.limit = limit
	return 2, nil
}

func TestWorkerIntervalDefaultsAndOverrides(t *testing.T) {
	t.Setenv("WORKER_INTERVAL_SECONDS", "")
	if got := workerInterval(); got != 60*time.Second {
		t.Fatalf("default interval=%s", got)
	}
	t.Setenv("WORKER_INTERVAL_SECONDS", "15")
	if got := workerInterval(); got != 15*time.Second {
		t.Fatalf("override interval=%s", got)
	}
	_ = os.Unsetenv("WORKER_INTERVAL_SECONDS")
}

func TestWorkerOnceFlag(t *testing.T) {
	t.Setenv("WORKER_ONCE", "true")
	if !workerOnce() {
		t.Fatal("expected once mode")
	}
	t.Setenv("WORKER_ONCE", "no")
	if workerOnce() {
		t.Fatal("expected loop mode")
	}
}

func TestAdvisorExpiryPassUsesBoundedBatch(t *testing.T) {
	runner := &advisorExpiryRunnerStub{}
	now := time.Date(2026, 10, 15, 10, 0, 0, 0, time.UTC)
	count, err := runAdvisorExpiryPass(context.Background(), runner, now)
	if err != nil || count != 2 {
		t.Fatalf("expiry pass count=%d err=%v", count, err)
	}
	if runner.limit != 100 || !runner.now.Equal(now) {
		t.Fatalf("expiry pass must use the bounded worker contract: limit=%d now=%s", runner.limit, runner.now)
	}
}

func TestAdvisorExpiryPassIsNoopWithoutRunner(t *testing.T) {
	count, err := runAdvisorExpiryPass(context.Background(), nil, time.Now().UTC())
	if err != nil || count != 0 {
		t.Fatalf("nil expiry runner should be a no-op: count=%d err=%v", count, err)
	}
}

func TestAttachmentCleanupPassUsesBoundedBatch(t *testing.T) {
	runner := &attachmentCleanupRunnerStub{n: 3}
	count, err := runAttachmentCleanupPass(context.Background(), runner)
	if err != nil || count != 3 {
		t.Fatalf("attachment cleanup pass count=%d err=%v", count, err)
	}
	if runner.limit != 100 {
		t.Fatalf("attachment cleanup pass must use bounded batch: limit=%d", runner.limit)
	}
}

func TestAttachmentCleanupPassIsNoopWithoutRunner(t *testing.T) {
	count, err := runAttachmentCleanupPass(context.Background(), nil)
	if err != nil || count != 0 {
		t.Fatalf("nil attachment cleanup runner should be a no-op: count=%d err=%v", count, err)
	}
}

func TestBudgetNotificationCandidatesUseEffectiveLimitAndStableDedupe(t *testing.T) {
	start := entity.CalendarDate{Time: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)}
	end := entity.CalendarDate{Time: time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)}
	cat := "food"
	budget := entity.Budget{ID: "budget-1", OwnerID: "owner-1", Name: "Ăn uống", LimitAmount: 1000, CarryoverAmount: 200, EffectiveLimitAmount: 1200, StartDate: start, EndDate: end, CategoryID: &cat}
	now := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	txs := []entity.Transaction{
		{ID: "expense", OwnerID: "owner-1", Type: entity.TransactionTypeExpense, Amount: 960, CategoryID: &cat, OccurredAt: now, IncludedInReports: true},
		{ID: "transfer", OwnerID: "owner-1", Type: entity.TransactionTypeExpense, Amount: 1000, CategoryID: &cat, OccurredAt: now, IncludedInReports: true, TransferID: stringPtr("transfer-1")},
	}
	candidates, err := budgetNotificationCandidates(budget, txs, []entity.Category{{ID: cat, Kind: entity.TransactionTypeExpense}}, now, time.UTC)
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) != 1 || candidates[0].Threshold != 80 {
		t.Fatalf("expected only 80%% threshold for 960/1200, got %+v", candidates)
	}
	if candidates[0].DedupeKey != "budget-1:2026-10-01:2026-10-31:80" {
		t.Fatalf("unexpected stable dedupe key: %q", candidates[0].DedupeKey)
	}
}

func stringPtr(value string) *string { return &value }
