package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type mcpTransportHeaders struct {
	base  http.RoundTripper
	token string
}

func (h mcpTransportHeaders) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	r.Header.Set("Authorization", h.token)
	r.Header.Set("X-Request-ID", "mcp-transport-test")
	return h.base.RoundTrip(r)
}

func mcpTransportSession(t *testing.T, handler http.Handler, token string, protocolVersion ...string) *mcp.ClientSession {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	client := mcp.NewClient(&mcp.Implementation{Name: "mypocket-transport-test", Version: "1"}, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var options *mcp.ClientSessionOptions
	if len(protocolVersion) > 0 {
		options = &mcp.ClientSessionOptions{ProtocolVersion: protocolVersion[0]}
	}
	session, err := client.Connect(ctx, &mcp.StreamableClientTransport{
		Endpoint:             srv.URL,
		HTTPClient:           &http.Client{Transport: mcpTransportHeaders{base: srv.Client().Transport, token: token}},
		DisableStandaloneSSE: true, MaxRetries: -1,
	}, options)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close() })
	return session
}

func mcpTransportBody(t *testing.T, result *mcp.CallToolResult) map[string]any {
	t.Helper()
	encoded, err := json.Marshal(result.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}
	var body map[string]any
	if err := json.Unmarshal(encoded, &body); err != nil {
		t.Fatal(err)
	}
	return body
}

func TestMCPInitializeToolsAndCanonicalDispatch(t *testing.T) {
	var calls atomic.Int32
	dispatch := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.Method != "GET" || r.URL.Path != "/api/v1/wallets" {
			t.Errorf("dispatch = %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer owner-jwt" {
			t.Error("bearer not propagated")
		}
		if r.Header.Get("X-Request-ID") != "mcp-transport-test" {
			t.Error("request ID not propagated")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Request-ID", "canonical-request")
		_, _ = io.WriteString(w, `{"data":[{"id":"owned-wallet"}]}`)
	})
	session := mcpTransportSession(t, NewMCPHandler(DefaultExternalOperationCatalog(), dispatch, MCPHandlerOptions{}), "Bearer owner-jwt")
	listed, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	foundRead, foundWrite := false, false
	for _, tool := range listed.Tools {
		foundRead = foundRead || tool.Name == "wallets.list"
		foundWrite = foundWrite || tool.Name == "wallets.create"
	}
	if !foundRead || !foundWrite {
		t.Fatal("ordinary read/write tools missing")
	}
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "wallets.list", Arguments: map[string]any{}})
	if err != nil {
		t.Fatal(err)
	}
	body := mcpTransportBody(t, result)
	if result.IsError || body["status"] != float64(200) || body["request_id"] != "canonical-request" {
		t.Fatalf("result = %+v", result)
	}
	if !strings.Contains(string(mustMarshalMCP(t, body["body"])), "owned-wallet") {
		t.Fatal("canonical body missing")
	}
	if len(result.Content) == 0 || calls.Load() != 1 {
		t.Fatal("missing content or unexpected dispatch count")
	}
}

func mustMarshalMCP(t *testing.T, value any) []byte {
	t.Helper()
	b, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestMCPCallPreservesDomainProblem(t *testing.T) {
	session := mcpTransportSession(t, NewMCPHandler(DefaultExternalOperationCatalog(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(409)
		_, _ = io.WriteString(w, `{"status":409,"code":"VERSION_CONFLICT","detail":"stale version","request_id":"conflict-request"}`)
	}), MCPHandlerOptions{}), "Bearer owner")
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "wallets.list", Arguments: map[string]any{}})
	if err != nil {
		t.Fatal(err)
	}
	body := mcpTransportBody(t, result)
	problem, ok := body["body"].(map[string]any)
	if !result.IsError || body["status"] != float64(409) || !ok || problem["code"] != "VERSION_CONFLICT" || problem["detail"] != "stale version" {
		t.Fatalf("problem = %+v", result)
	}
}

func TestMCPToolsUnknownAndArbitraryPathNeverDispatch(t *testing.T) {
	var calls atomic.Int32
	session := mcpTransportSession(t, NewMCPHandler(DefaultExternalOperationCatalog(), http.HandlerFunc(func(http.ResponseWriter, *http.Request) { calls.Add(1) }), MCPHandlerOptions{}), "Bearer owner")
	_, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "call_rest", Arguments: map[string]any{"method": "POST", "path": "/api/v1/internal/feedback"}})
	if err == nil || calls.Load() != 0 {
		t.Fatal("unknown tool escaped SDK/catalog boundary")
	}
}

func TestMCPDispatchIdempotencyMetadata(t *testing.T) {
	session := mcpTransportSession(t, NewMCPHandler(DefaultExternalOperationCatalog(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Idempotency-Key") != "stable-key" {
			t.Error("idempotency metadata missing")
		}
		body, _ := io.ReadAll(r.Body)
		if strings.Contains(string(body), "idempotency_key") {
			t.Error("adapter metadata leaked into business body")
		}
		w.WriteHeader(201)
		_, _ = io.WriteString(w, `{"data":{"id":"created"}}`)
	}), MCPHandlerOptions{}), "Bearer owner")
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "wallets.create", Arguments: map[string]any{"name": "Travel", "type": "basic", "opening_balance": 0}, Meta: mcp.Meta{"idempotency_key": "stable-key"}})
	if err != nil || result.IsError {
		t.Fatalf("call = %+v, %v", result, err)
	}
}

func TestMCPDispatchCallerCancellation(t *testing.T) {
	for _, version := range []string{"2026-07-28", "2025-11-25"} {
		t.Run(version, func(t *testing.T) {
			started, stopped := make(chan struct{}), make(chan struct{})
			session := mcpTransportSession(t, NewMCPHandler(DefaultExternalOperationCatalog(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				close(started)
				<-r.Context().Done()
				close(stopped)
			}), MCPHandlerOptions{}), "Bearer owner", version)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			finished := make(chan struct{})
			go func() {
				defer close(finished)
				_, _ = session.CallTool(ctx, &mcp.CallToolParams{Name: "wallets.list", Arguments: map[string]any{}})
			}()
			select {
			case <-started:
			case <-time.After(3 * time.Second):
				t.Fatal("dispatch did not start")
			}
			cancel()
			select {
			case <-stopped:
			case <-time.After(3 * time.Second):
				t.Fatal("caller cancellation detached from canonical operation")
			}
			<-finished
		})
	}
}

func TestMCPDispatchDeadline(t *testing.T) {
	for _, version := range []string{"2026-07-28", "2025-11-25"} {
		t.Run(version, func(t *testing.T) {
			stopped := make(chan struct{})
			session := mcpTransportSession(t, NewMCPHandler(DefaultExternalOperationCatalog(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				<-r.Context().Done()
				close(stopped)
			}), MCPHandlerOptions{RequestTimeout: 100 * time.Millisecond}), "Bearer owner", version)
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "wallets.list", Arguments: map[string]any{}})
			if err != nil {
				t.Fatalf("operation timeout must arrive as a structured MCP result: %v", err)
			}
			body := mcpTransportBody(t, result)
			problem, ok := body["body"].(map[string]any)
			if !result.IsError || body["status"] != float64(504) || !ok || problem["code"] != "MCP_DEADLINE_EXCEEDED" || problem["status"] != float64(504) {
				t.Fatalf("timeout result = %+v", result)
			}
			select {
			case <-stopped:
			case <-time.After(time.Second):
				t.Fatal("deadline did not stop dispatch")
			}
			// A call timeout must not terminate the still-authenticated transport.
			if _, err := session.ListTools(ctx, nil); err != nil {
				t.Fatalf("transport unusable after operation timeout: %v", err)
			}
		})
	}
}

func TestMCPLimitsAndOrigin(t *testing.T) {
	handler := NewMCPHandler(DefaultExternalOperationCatalog(), http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Error("must not dispatch") }), MCPHandlerOptions{MaxRequestBytes: 128, AllowedOrigins: []string{"https://app.example"}})
	for _, tc := range []struct {
		name, origin, body string
		status             int
	}{
		{"origin", "https://evil.example", `{}`, 403},
		{"size", "", strings.Repeat("x", 129), 413},
		{"chunked-size", "https://app.example", strings.Repeat("x", 129), 413},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest("POST", "http://localhost/api/v1/mcp", strings.NewReader(tc.body))
			if tc.name == "chunked-size" {
				r.ContentLength = -1
			}
			r.Header.Set("Origin", tc.origin)
			r.Header.Set("Content-Type", "application/json")
			r.Header.Set("Accept", "application/json, text/event-stream")
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			if w.Code != tc.status {
				t.Fatalf("status=%d, want=%d, body=%s", w.Code, tc.status, w.Body.String())
			}
		})
	}
}

func TestMCPResponseLimit(t *testing.T) {
	session := mcpTransportSession(t, NewMCPHandler(DefaultExternalOperationCatalog(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, strings.Repeat("private-financial-data", 1024))
	}), MCPHandlerOptions{MaxResponseBytes: 128}), "Bearer owner")
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "wallets.list", Arguments: map[string]any{}})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError || strings.Contains(string(mustMarshalMCP(t, result)), "private-financial-data") {
		t.Fatal("oversized partial private payload leaked")
	}
}

func TestMCPDispatchPreservesJSONNumbers(t *testing.T) {
	session := mcpTransportSession(t, NewMCPHandler(DefaultExternalOperationCatalog(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"data":{"version":9007199254740993,"amount":1234567890.123456789}}`)
	}), MCPHandlerOptions{}), "Bearer owner")
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "wallets.list", Arguments: map[string]any{}})
	if err != nil {
		t.Fatal(err)
	}
	text, ok := result.Content[0].(*mcp.TextContent)
	if !ok || !strings.Contains(text.Text, "9007199254740993") || !strings.Contains(text.Text, "1234567890.123456789") {
		t.Fatalf("canonical JSON numbers changed: %+v", result.Content)
	}
}

func TestMCPDispatchFiniteResponseFlush(t *testing.T) {
	session := mcpTransportSession(t, NewMCPHandler(DefaultExternalOperationCatalog(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "event: completed\ndata: {}\n\n")
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Error("canonical finite SSE handler requires http.Flusher")
			return
		}
		flusher.Flush()
	}), MCPHandlerOptions{}), "Bearer owner")
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "wallets.list", Arguments: map[string]any{}})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError || !strings.Contains(string(mustMarshalMCP(t, result)), "completed") {
		t.Fatalf("finite replay lost: %+v", result)
	}
}

func TestMCPAuthWrapperRejectsBeforeProtocol(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/v1/mcp", (&AuthMiddleware{}).RequireAuth, gin.WrapH(NewMCPHandler(DefaultExternalOperationCatalog(), http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Error("must not dispatch") }), MCPHandlerOptions{})))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest("POST", "/api/v1/mcp", strings.NewReader("{}")))
	if w.Code != 401 || !strings.Contains(w.Body.String(), "AUTH_REQUIRED") {
		t.Fatalf("unauthenticated response = %d %s", w.Code, w.Body.String())
	}
}
