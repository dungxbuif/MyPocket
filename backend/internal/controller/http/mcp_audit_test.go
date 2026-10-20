package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/mypocket/backend/internal/repository"
	"github.com/mypocket/backend/internal/usecase"
)

type mcpAuditCapture struct {
	mu     sync.Mutex
	events []repository.AuditEvent
	err    error
}

func TestMCPAuditRecordsExactBoundedByteCounts(t *testing.T) {
	for _, tc := range []struct {
		name  string
		limit int64
		want  int64
	}{
		{"complete", 64, 11}, {"overflow", 7, 7},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sink := &mcpAuditCapture{}
			session := mcpTransportSession(t, NewMCPHandler(DefaultExternalOperationCatalog(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = io.WriteString(w, `{"data":[]}`) }), MCPHandlerOptions{Audit: sink, MaxResponseBytes: tc.limit}), "Bearer owner")
			_, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "wallets.list", Arguments: map[string]any{}})
			if err != nil {
				t.Fatal(err)
			}
			events := sink.await(t, 1)
			var event map[string]any
			if err := json.Unmarshal(mustMarshalMCP(t, events[0]), &event); err != nil {
				t.Fatal(err)
			}
			if event["RequestBytes"] != float64(2) || event["ResponseBytes"] != float64(tc.want) {
				t.Fatalf("audit count metadata=%+v", event)
			}
		})
	}
}

type mcpStalledAuditSink struct {
	started chan struct{}
	release chan struct{}
	calls   atomic.Int32
}

func (s *mcpStalledAuditSink) Record(context.Context, repository.AuditEvent) error {
	s.calls.Add(1)
	select {
	case s.started <- struct{}{}:
	default:
	}
	<-s.release // Deliberately ignores context, like a blocked Redis socket.
	return nil
}

func TestMCPAuditStalledSinkDoesNotBlockBusinessResponse(t *testing.T) {
	sink := &mcpStalledAuditSink{started: make(chan struct{}, 1), release: make(chan struct{})}
	release := sync.OnceFunc(func() { close(sink.release) })
	defer release()
	session := mcpTransportSession(t, NewMCPHandler(DefaultExternalOperationCatalog(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = io.WriteString(w, `{"data":[]}`) }), MCPHandlerOptions{Audit: sink}), "Bearer owner")
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()
	result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "wallets.list", Arguments: map[string]any{}})
	if err != nil || result.IsError {
		t.Fatalf("stalled sink held business response: %v", err)
	}
	select {
	case <-sink.started:
	case <-time.After(time.Second):
		t.Fatal("audit delivery never started")
	}
	// Saturate the per-handler queue while its only worker is blocked. Every
	// subsequent business response must still complete without another sink call.
	for i := 0; i < 80; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
		result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "wallets.list", Arguments: map[string]any{}})
		cancel()
		if err != nil || result.IsError {
			t.Fatalf("queue saturation delayed call %d: %v", i, err)
		}
	}
	if sink.calls.Load() != 1 {
		t.Fatalf("stalled sink has %d concurrent deliveries", sink.calls.Load())
	}
	release()
	deadline := time.Now().Add(time.Second)
	for sink.calls.Load() < 65 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	// One in-flight event plus 64 queued events can be delivered. Overflow
	// must have been dropped rather than retained as an unbounded backlog.
	time.Sleep(10 * time.Millisecond)
	if got := sink.calls.Load(); got != 65 {
		t.Fatalf("delivered %d events after saturation, want 65", got)
	}
}

func (s *mcpAuditCapture) await(t *testing.T, count int) []repository.AuditEvent {
	t.Helper()
	deadline := time.NewTimer(time.Second)
	defer deadline.Stop()
	tick := time.NewTicker(time.Millisecond)
	defer tick.Stop()
	for {
		events := s.snapshot()
		if len(events) >= count {
			return events
		}
		select {
		case <-tick.C:
		case <-deadline.C:
			t.Fatalf("audit delivered %d want=%d", len(events), count)
		}
	}
}

func TestMCPAuditDeniedErrorsAndLimits(t *testing.T) {
	for _, tc := range []struct {
		name                   string
		status                 int
		args                   map[string]any
		limit                  bool
		wantStatus, wantReason string
	}{
		{"owner-not-found", 404, nil, false, "404", "denied"},
		{"conflict", 409, nil, false, "409", "denied"},
		{"unavailable", 503, nil, false, "503", "error"},
		{"invalid-owner-selector", 200, map[string]any{"owner_id": "foreign"}, false, "400", "denied"},
		{"output-limit", 200, nil, true, "502", "error"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.args == nil {
				tc.args = map[string]any{}
			}
			sink := &mcpAuditCapture{}
			opts := MCPHandlerOptions{Audit: sink}
			if tc.limit {
				opts.MaxResponseBytes = 32
			}
			session := mcpTransportSession(t, NewMCPHandler(DefaultExternalOperationCatalog(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = io.WriteString(w, strings.Repeat("private financial payload", 3))
			}), opts), "Bearer owner")
			result, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "wallets.list", Arguments: tc.args})
			if err != nil || !result.IsError {
				t.Fatalf("call=%+v %v", result, err)
			}
			events := sink.await(t, 1)
			if len(events) != 1 || events[0].Allowed || events[0].Status != tc.wantStatus || events[0].Reason != tc.wantReason {
				t.Fatalf("denial audit=%+v", events)
			}
		})
	}
}

func TestMCPAuditUnknownToolNameIsRedacted(t *testing.T) {
	sink := &mcpAuditCapture{}
	session := mcpTransportSession(t, NewMCPHandler(DefaultExternalOperationCatalog(), http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Error("unknown tool dispatched") }), MCPHandlerOptions{Audit: sink}), "Bearer owner")
	_, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "private-prompt-secret", Arguments: map[string]any{}})
	if err == nil {
		t.Fatal("unknown tool accepted")
	}
	events := sink.await(t, 1)
	if len(events) != 1 || events[0].Action != "mcp.unknown" || events[0].Allowed || strings.Contains(string(mustMarshalMCP(t, events)), "private-prompt-secret") {
		t.Fatalf("unknown audit=%+v", events)
	}
}

func TestMCPAuditDeadlineUsesBoundedLiveContext(t *testing.T) {
	sink := &mcpAuditContextSink{events: make(chan context.Context, 1)}
	session := mcpTransportSession(t, NewMCPHandler(DefaultExternalOperationCatalog(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done() }), MCPHandlerOptions{Audit: sink, RequestTimeout: 100 * time.Millisecond}), "Bearer owner")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, _ = session.CallTool(ctx, &mcp.CallToolParams{Name: "wallets.list", Arguments: map[string]any{}})
	select {
	case auditCtx := <-sink.events:
		if deadline, ok := auditCtx.Deadline(); !ok || time.Until(deadline) > time.Second {
			t.Fatal("audit context has no short deadline")
		}
	case <-time.After(time.Second):
		t.Fatal("deadline failure omitted audit")
	}
}

type mcpAuditContextSink struct{ events chan context.Context }

func (s *mcpAuditContextSink) Record(ctx context.Context, _ repository.AuditEvent) error {
	if ctx.Err() != nil {
		return errors.New("audit inherited caller cancellation")
	}
	s.events <- ctx
	return nil
}
func (s *mcpAuditCapture) Record(_ context.Context, event repository.AuditEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, event)
	return s.err
}
func (s *mcpAuditCapture) snapshot() []repository.AuditEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]repository.AuditEvent(nil), s.events...)
}

func TestMCPAuditPrivacyAndBestEffort(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "success", true: "sink-failure"}[fail], func(t *testing.T) {
			sink := &mcpAuditCapture{}
			if fail {
				sink.err = errors.New("audit unavailable")
			}
			options := MCPHandlerOptions{Audit: sink}
			handler := NewMCPHandler(DefaultExternalOperationCatalog(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, `{"data":{"run_id":"run","conversation_id":"conversation","status":"queued","private":"amount=123456 OCR base64 private-provider-usage"}}`)
			}), options)
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.Any("/", func(c *gin.Context) {
				c.Set(contextAdvisorPrincipalKey, usecase.Principal{OwnerID: "owner", CredentialKind: "api_key", CredentialID: "credential-id"})
				c.Set(requestIDContextKey, "canonical-generated-id")
				c.Request.Header.Del(requestIDHeader)
				c.Next()
			}, MCPAuthenticatedHandler(handler))
			session := mcpTransportSession(t, router, "Bearer mpk_private-secret")
			result, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "advisor.messages_async", Arguments: map[string]any{"text": "private-prompt-secret", "client_request_id": "request"}})
			if err != nil || result.IsError {
				t.Fatalf("audit altered business response: %+v %v", result, err)
			}
			events := sink.await(t, 1)
			if len(events) != 1 {
				t.Fatalf("audit events=%d want=1", len(events))
			}
			event := events[0]
			if event.Action != "mcp.advisor.messages_async" || event.RequestID != "canonical-generated-id" || event.ActorKind != "api_key" || event.CredentialID != "credential-id" || event.Status != "200" || !event.Allowed || event.Reason != "allowed" || event.LatencyMS < 0 {
				t.Fatalf("audit event=%+v", event)
			}
			raw := string(mustMarshalMCP(t, event))
			for _, secret := range []string{"mpk_private-secret", "private-prompt-secret", "amount=", "123456", "OCR", "base64", "private-provider-usage", "signature="} {
				if strings.Contains(raw, secret) {
					t.Fatalf("private value leaked into audit: %s", secret)
				}
			}
		})
	}
}
