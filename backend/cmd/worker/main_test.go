package main

import (
	"context"
	"os"
	"testing"
	"time"
)

type advisorExpiryRunnerStub struct {
	limit int
	now   time.Time
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
