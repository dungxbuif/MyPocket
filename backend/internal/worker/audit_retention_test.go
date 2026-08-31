package worker_test

import (
	"context"
	"testing"

	"mypocket/internal/worker"
)

func TestAuditRetentionRunnerPurgesOnlyWithLease(t *testing.T) {
	repo := &auditRetentionRepoStub{acquired: true}
	purged, err := (worker.AuditRetentionRunner{Repo: repo, Owner: "worker-a", RetentionDays: 90, Limit: 25}).RunOnce(context.Background())
	if err != nil {
		t.Fatalf("run audit retention: %v", err)
	}
	if purged != 7 || repo.leaseKey != "audit-retention" || repo.retentionDays != 90 || repo.limit != 25 {
		t.Fatalf("unexpected audit retention run: %#v purged=%d", repo, purged)
	}
}

func TestAuditRetentionRunnerSkipsWithoutLease(t *testing.T) {
	repo := &auditRetentionRepoStub{}
	purged, err := (worker.AuditRetentionRunner{Repo: repo, Owner: "worker-b"}).RunOnce(context.Background())
	if err != nil {
		t.Fatalf("run audit retention: %v", err)
	}
	if purged != 0 || repo.purged {
		t.Fatalf("runner should skip purge without lease: %#v purged=%d", repo, purged)
	}
}

type auditRetentionRepoStub struct {
	acquired      bool
	purged        bool
	leaseKey      string
	retentionDays int
	limit         int
}

func (r *auditRetentionRepoStub) AcquireWorkerLease(_ context.Context, leaseKey string, _ string, _ int64) (bool, error) {
	r.leaseKey = leaseKey
	return r.acquired, nil
}

func (r *auditRetentionRepoStub) PurgeExpired(_ context.Context, retentionDays int, limit int) (int, error) {
	r.purged = true
	r.retentionDays = retentionDays
	r.limit = limit
	return 7, nil
}
