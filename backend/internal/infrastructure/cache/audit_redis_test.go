package cache

import (
	"context"
	"encoding/json"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/repository"
)

type mcpAuditOutboxCapture struct {
	repository.AuditOutboxRepository
	payload []byte
}

func (s *mcpAuditOutboxCapture) EnqueueAudit(_ context.Context, _, _ string, payload []byte, _ time.Time) error {
	s.payload = append([]byte(nil), payload...)
	return nil
}

func TestRedisMCPAuditByteCountsSurviveOutboxAndReplay(t *testing.T) {
	raw := os.Getenv("TEST_REDIS_URL")
	if raw == "" {
		t.Skip("requires disposable loopback TEST_REDIS_URL")
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Hostname() != "127.0.0.1" && u.Hostname() != "localhost" && u.Hostname() != "::1") {
		t.Fatal("Redis audit proof requires loopback URL")
	}
	r, err := NewRedis(raw)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = r.Close() })
	outbox := &mcpAuditOutboxCapture{}
	r.SetAuditOutbox(outbox)
	requestID := "mcp-audit-count-" + uuid.NewString()
	event := repository.AuditEvent{RequestID: requestID, Action: "mcp.wallets.list", Allowed: true, Status: "200", RequestBytes: 2, ResponseBytes: 11}
	if err := r.Record(context.Background(), event); err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(outbox.payload, &payload); err != nil {
		t.Fatal(err)
	}
	if payload["request_bytes"] != "2" || payload["response_bytes"] != "11" {
		t.Fatalf("outbox counts=%+v", payload)
	}
	if err := r.PublishAudit(context.Background(), feedbackAuditStream, "mcp-count-replay-"+requestID, outbox.payload); err != nil {
		t.Fatal(err)
	}
	delete(payload, "request_bytes")
	delete(payload, "response_bytes")
	legacy, _ := json.Marshal(payload)
	if err := r.PublishAudit(context.Background(), feedbackAuditStream, "mcp-count-legacy-"+requestID, legacy); err != nil {
		t.Fatal(err)
	}
	entries, err := r.client.XRange(context.Background(), feedbackAuditStream, "-", "+").Result()
	if err != nil {
		t.Fatal(err)
	}
	matched := 0
	for _, entry := range entries {
		if entry.Values["request_id"] == requestID {
			matched++
			_, _ = r.client.XDel(context.Background(), feedbackAuditStream, entry.ID).Result()
			wantRequest, wantResponse := "2", "11"
			if entry.Values["event_id"] == "mcp-count-legacy-"+requestID {
				wantRequest, wantResponse = "0", "0"
			}
			if entry.Values["request_bytes"] != wantRequest || entry.Values["response_bytes"] != wantResponse {
				t.Fatalf("stream count metadata=%+v", entry.Values)
			}
		}
	}
	if matched != 3 {
		t.Fatalf("record/replay entries=%d want=3", matched)
	}
}

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

func TestRedisRecordRetryDoesNotDuplicateTheSameEvent(t *testing.T) {
	redisURL := os.Getenv("TEST_REDIS_URL")
	if redisURL == "" {
		t.Skip("TEST_REDIS_URL is required for Redis integration proof")
	}
	r, err := NewRedis(redisURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = r.Close() })
	requestID := "audit-retry-" + uuid.NewString()
	event := repository.AuditEvent{
		RequestID: requestID, ActorKind: "user_api_key", CredentialID: "key-test",
		Action: "advisor.overview", Allowed: true, Status: "200", Reason: "allowed",
	}
	if err := r.Record(context.Background(), event); err != nil {
		t.Fatal(err)
	}
	if err := r.Record(context.Background(), event); err != nil {
		t.Fatal(err)
	}
	entries, err := r.client.XRange(context.Background(), advisorAuditStream, "-", "+").Result()
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, entry := range entries {
		if entry.Values["request_id"] == requestID {
			count++
			_, _ = r.client.XDel(context.Background(), advisorAuditStream, entry.ID).Result()
		}
	}
	if count != 1 {
		t.Fatalf("same audit event was recorded %d times, want exactly once", count)
	}
}
