package repository

import "context"

type AuditEvent struct {
	RequestID    string
	ActorKind    string
	CredentialID string
	Action       string
	Allowed      bool
	Status       string
	Reason       string
	LatencyMS    int64
	FeedbackIDs  []string
	FeedbackID   string
	ChangelogID  string
	OldStatus    string
	NewStatus    string
	Version      string
}

type AuditSink interface {
	Record(context.Context, AuditEvent) error
}
