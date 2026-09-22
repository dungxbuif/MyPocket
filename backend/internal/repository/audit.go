package repository

import "context"

type AuditEvent struct {
	RequestID   string
	ActorKind   string
	Action      string
	FeedbackIDs []string
	FeedbackID  string
	ChangelogID string
	OldStatus   string
	NewStatus   string
	Version     string
}

type AuditSink interface {
	Record(context.Context, AuditEvent) error
}
