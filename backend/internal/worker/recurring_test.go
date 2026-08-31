package worker

import (
	"context"
	"testing"
	"time"
)

func TestRecurringRunnerRequiresLeaseBeforeProcessing(t *testing.T) {
	processor := &recurringProcessorStub{acquired: true}
	now := time.Date(2026, 8, 31, 2, 0, 0, 0, time.UTC)
	count, err := (RecurringRunner{Processor: processor, Owner: "worker-a", Now: func() time.Time { return now }}).RunOnce(context.Background())
	if err != nil {
		t.Fatalf("run once: %v", err)
	}
	if count != 3 || processor.leaseKey != "recurring-schedules" || processor.processOwner != "worker-a" {
		t.Fatalf("runner did not lease/process correctly: %#v count=%d", processor, count)
	}
}

func TestRecurringRunnerSkipsProcessingWithoutLease(t *testing.T) {
	processor := &recurringProcessorStub{}
	count, err := (RecurringRunner{Processor: processor, Owner: "worker-b"}).RunOnce(context.Background())
	if err != nil {
		t.Fatalf("run once: %v", err)
	}
	if count != 0 || processor.processed {
		t.Fatalf("runner should skip processing without lease: %#v count=%d", processor, count)
	}
}

type recurringProcessorStub struct {
	acquired     bool
	processed    bool
	leaseKey     string
	processOwner string
}

func (s *recurringProcessorStub) AcquireWorkerLease(_ context.Context, leaseKey string, _ string, _ time.Duration, _ time.Time) (bool, error) {
	s.leaseKey = leaseKey
	return s.acquired, nil
}

func (s *recurringProcessorStub) ProcessDueRecurringSchedules(_ context.Context, owner string, _ time.Time, _ int) (int, error) {
	s.processed = true
	s.processOwner = owner
	return 3, nil
}
