package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mypocket/backend/internal/repository"
)

type auditOutboxStub struct {
	rows       []repository.AuditOutboxEntry
	delivered  []string
	failedID   string
	nextRetry  time.Time
	claimError error
}

func (s *auditOutboxStub) EnqueueAudit(context.Context, string, string, []byte, time.Time) error {
	return nil
}
func (s *auditOutboxStub) ClaimAudit(context.Context, time.Time, int, time.Duration) ([]repository.AuditOutboxEntry, error) {
	if s.claimError != nil {
		return nil, s.claimError
	}
	return s.rows, nil
}
func (s *auditOutboxStub) MarkAuditDelivered(_ context.Context, id string, _ time.Time) error {
	s.delivered = append(s.delivered, id)
	return nil
}
func (s *auditOutboxStub) MarkAuditFailed(_ context.Context, id string, retry time.Time, _ string) error {
	s.failedID, s.nextRetry = id, retry
	return nil
}

type auditPublisherStub struct {
	err        error
	failOnCall int
	calls      int
}

func (s *auditPublisherStub) PublishAudit(context.Context, string, string, []byte) error {
	s.calls++
	if s.failOnCall > 0 && s.calls == s.failOnCall {
		return s.err
	}
	return nil
}

func TestRunAuditOutboxPassMarksDeliveredAndRetriesFailures(t *testing.T) {
	now := time.Date(2026, 10, 19, 12, 0, 0, 0, time.UTC)
	outbox := &auditOutboxStub{rows: []repository.AuditOutboxEntry{{ID: "ok", Stream: "mypocket:audit:feedback", EventID: "e1", Payload: []byte(`{"action":"x"}`), Attempts: 1}, {ID: "bad", Stream: "mypocket:audit:feedback", EventID: "e2", Payload: []byte(`{"action":"y"}`), Attempts: 2}}}
	publisher := &auditPublisherStub{}
	// The fake publisher fails only the second event, proving one bad delivery
	// does not prevent the rest of the bounded batch from being processed.
	publisher.err, publisher.failOnCall = errors.New("redis down"), 2
	count, err := runAuditOutboxPass(context.Background(), outbox, publisher, now)
	if err == nil || count != 1 || len(outbox.delivered) != 1 || publisher.calls != 2 || outbox.failedID != "bad" {
		t.Fatalf("pass count=%d err=%v delivered=%v calls=%d", count, err, outbox.delivered, publisher.calls)
	}

	outbox = &auditOutboxStub{rows: []repository.AuditOutboxEntry{{ID: "bad", Attempts: 3}}}
	publisher = &auditPublisherStub{err: errors.New("redis down"), failOnCall: 1}
	count, err = runAuditOutboxPass(context.Background(), outbox, publisher, now)
	if count != 0 || err == nil || outbox.failedID != "bad" || !outbox.nextRetry.Equal(now.Add(4*time.Second)) {
		t.Fatalf("failure retry count=%d err=%v failed=%q retry=%s", count, err, outbox.failedID, outbox.nextRetry)
	}
}

func TestAuditRetryDelayIsBounded(t *testing.T) {
	if got := auditRetryDelay(1); got != time.Second {
		t.Fatalf("first delay=%s", got)
	}
	if got := auditRetryDelay(20); got != 5*time.Minute {
		t.Fatalf("bounded delay=%s", got)
	}
}
