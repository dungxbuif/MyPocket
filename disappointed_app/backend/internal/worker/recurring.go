package worker

import (
	"context"
	"fmt"
	"os"
	"time"
)

type RecurringProcessor interface {
	AcquireWorkerLease(ctx context.Context, leaseKey string, owner string, ttl time.Duration, now time.Time) (bool, error)
	ProcessDueRecurringSchedules(ctx context.Context, workerID string, now time.Time, limit int) (int, error)
}

type RecurringRunner struct {
	Processor RecurringProcessor
	Owner     string
	Now       func() time.Time
}

func (r RecurringRunner) RunOnce(ctx context.Context) (int, error) {
	if r.Processor == nil {
		return 0, fmt.Errorf("recurring processor is required")
	}
	owner := r.Owner
	if owner == "" {
		owner = defaultOwner()
	}
	now := time.Now
	if r.Now != nil {
		now = r.Now
	}
	current := now().UTC()
	acquired, err := r.Processor.AcquireWorkerLease(ctx, "recurring-schedules", owner, time.Minute, current)
	if err != nil {
		return 0, err
	}
	if !acquired {
		return 0, nil
	}
	return r.Processor.ProcessDueRecurringSchedules(ctx, owner, current, 50)
}

func defaultOwner() string {
	host, err := os.Hostname()
	if err != nil || host == "" {
		return "worker"
	}
	return host
}
