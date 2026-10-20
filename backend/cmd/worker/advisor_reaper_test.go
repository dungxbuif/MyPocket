package main

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/mypocket/backend/internal/entity"
)

type reaperStoreStub struct {
	mu    sync.Mutex
	calls int
}

func (s *reaperStoreStub) ClaimRun(context.Context, string, string, string) (entity.AdvisorRunLease, error) {
	return entity.AdvisorRunLease{}, nil
}
func (s *reaperStoreStub) HeartbeatRun(context.Context, entity.AdvisorRunLease) error { return nil }
func (s *reaperStoreStub) CompleteRun(context.Context, entity.AdvisorRunLease, entity.AdvisorCompletion) (entity.AdvisorCompletionResult, error) {
	return entity.AdvisorCompletionResult{}, nil
}
func (s *reaperStoreStub) FinishRun(context.Context, entity.AdvisorRunLease, string, string, map[string]any) error {
	return nil
}
func (s *reaperStoreStub) ReapExpiredRuns(context.Context, int) (int, error) {
	s.mu.Lock()
	s.calls++
	s.mu.Unlock()
	return 1, nil
}

func TestAdvisorReaperRunsIndependentlyAndStops(t *testing.T) {
	store := &reaperStoreStub{}
	stop := startAdvisorReaperWithInterval(context.Background(), store, time.Millisecond)
	time.Sleep(5 * time.Millisecond)
	stop()
	store.mu.Lock()
	calls := store.calls
	store.mu.Unlock()
	if calls == 0 {
		t.Fatal("reaper did not tick")
	}
}
