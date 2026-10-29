package repository

import (
	"context"
	"time"
)

type AuditEvent struct {
	RequestID    string
	ActorKind    string
	CredentialID string
	Action       string
	Allowed      bool
	Status       string
	Reason       string
	LatencyMS    int64
	// MCP counts argument JSON and canonical REST body bytes, saturated at
	// configured capture limits. Other audit producers may leave these zero.
	RequestBytes  int64
	ResponseBytes int64
	FeedbackIDs   []string
	FeedbackID    string
	ChangelogID   string
	OldStatus     string
	NewStatus     string
	Version       string
}

type AuditSink interface {
	Record(context.Context, AuditEvent) error
}

// AuditOutboxEntry is a redacted audit payload persisted before delivery to
// Redis. Payload must contain only the fields produced by auditFields; it must
// never contain prompts, OCR text, credentials or financial rows.
type AuditOutboxEntry struct {
	ID          string
	Stream      string
	EventID     string
	Payload     []byte
	Attempts    int
	AvailableAt time.Time
	LeaseUntil  *time.Time
}

// AuditOutboxRepository provides a durable hand-off between request handling
// and the Redis audit stream. Implementations must make Enqueue idempotent on
// (stream,event_id), and Claim use a lease so multiple workers are safe.
type AuditOutboxRepository interface {
	EnqueueAudit(context.Context, string, string, []byte, time.Time) error
	ClaimAudit(context.Context, time.Time, int, time.Duration) ([]AuditOutboxEntry, error)
	MarkAuditDelivered(context.Context, string, time.Time) error
	MarkAuditFailed(context.Context, string, time.Time, string) error
}
