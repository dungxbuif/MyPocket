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
