package worker

import (
	"context"
	"fmt"
	"time"
)

type AuditRetentionRepository interface {
	AcquireWorkerLease(ctx context.Context, leaseKey string, owner string, ttlSeconds int64) (bool, error)
	PurgeExpired(ctx context.Context, retentionDays int, limit int) (int, error)
}

type AuditRetentionRunner struct {
	Repo          AuditRetentionRepository
	Owner         string
	RetentionDays int
	Limit         int
}

func (r AuditRetentionRunner) RunOnce(ctx context.Context) (int, error) {
	if r.Repo == nil {
		return 0, fmt.Errorf("audit retention repository is required")
	}
	owner := r.Owner
	if owner == "" {
		owner = defaultOwner()
	}
	acquired, err := r.Repo.AcquireWorkerLease(ctx, "audit-retention", owner, int64(time.Hour.Seconds()))
	if err != nil {
		return 0, err
	}
	if !acquired {
		return 0, nil
	}
	retentionDays := r.RetentionDays
	if retentionDays <= 0 {
		retentionDays = 180
	}
	limit := r.Limit
	if limit <= 0 {
		limit = 1000
	}
	return r.Repo.PurgeExpired(ctx, retentionDays, limit)
}
