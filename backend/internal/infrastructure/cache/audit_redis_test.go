package cache

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/repository"
)

func TestRedisRecordAdvisorUsesDedicatedStream(t *testing.T) {
	redisURL := os.Getenv("TEST_REDIS_URL")
	if redisURL == "" {
		t.Skip("TEST_REDIS_URL is required for Redis integration proof")
	}
	r, err := NewRedis(redisURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = r.Close() })
	requestID := "audit-test-" + uuid.NewString()
	if err := r.Record(context.Background(), repository.AuditEvent{
		RequestID: requestID, ActorKind: "user_api_key", CredentialID: "key-test",
		Action: "advisor.overview", Allowed: true, Status: "200", Reason: "allowed",
	}); err != nil {
		t.Fatal(err)
	}
	entries, err := r.client.XRange(context.Background(), advisorAuditStream, "-", "+").Result()
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.Values["request_id"] == requestID {
			if entry.Values["credential_id"] != "key-test" || entry.Values["allowed"] != "1" {
				t.Fatalf("unexpected advisor audit entry: %#v", entry.Values)
			}
			_, _ = r.client.XDel(context.Background(), advisorAuditStream, entry.ID).Result()
			return
		}
	}
	t.Fatalf("advisor audit request %q not found", requestID)
}
