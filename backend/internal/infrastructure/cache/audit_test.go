package cache

import (
	"testing"

	"github.com/mypocket/backend/internal/repository"
)

func TestAuditFieldsContainOnlyRedactedFeedbackMetadata(t *testing.T) {
	fields := auditFields(repository.AuditEvent{
		RequestID: "req-1", ActorKind: "agent", Action: "feedback.created", FeedbackID: "fb-1", Version: "1.4.2",
	})
	if fields["request_id"] != "req-1" || fields["action"] != "feedback.created" || fields["feedback_id"] != "fb-1" {
		t.Fatalf("missing audit metadata: %#v", fields)
	}
	for _, forbidden := range []string{"title", "description", "token", "email", "jwt"} {
		if _, ok := fields[forbidden]; ok {
			t.Fatalf("sensitive audit field %q leaked: %#v", forbidden, fields)
		}
	}
}

func TestAuditFieldsContainAdvisorAccessMetadataOnly(t *testing.T) {
	fields := auditFields(repository.AuditEvent{
		RequestID: "req-2", ActorKind: "user_api_key", CredentialID: "key-1",
		Action: "advisor.overview", Allowed: true, Status: "200", Reason: "allowed", LatencyMS: 12,
	})
	for key, want := range map[string]string{
		"request_id": "req-2", "actor_kind": "user_api_key", "credential_id": "key-1",
		"action": "advisor.overview", "status": "200", "reason": "allowed",
	} {
		if fields[key] != want {
			t.Fatalf("audit field %q = %#v, want %q", key, fields[key], want)
		}
	}
	if fields["allowed"] != "1" || fields["latency_ms"] != "12" {
		t.Fatalf("missing advisor decision metadata: %#v", fields)
	}
	for _, forbidden := range []string{"secret", "prompt", "text", "note", "amount", "token", "jwt"} {
		if _, ok := fields[forbidden]; ok {
			t.Fatalf("sensitive audit field %q leaked: %#v", forbidden, fields)
		}
	}
}
