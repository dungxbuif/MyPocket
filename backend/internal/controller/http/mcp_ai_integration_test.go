package httpapi

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// These fixtures exercise the SDK/catalog/REST mapping boundary. Provider and
// persistence lifecycle semantics remain covered by canonical handler tests.
func TestMCPAdvisorAsyncHandleWithoutResubmission(t *testing.T) {
	var calls atomic.Int32
	session := mcpTransportSession(t, NewMCPHandler(DefaultExternalOperationCatalog(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.URL.Path != "/api/v1/ai/advisor/messages/async" {
			t.Error(r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(202)
		_, _ = io.WriteString(w, `{"data":{"run_id":"opaque-run","conversation_id":"conversation","status":"queued"}}`)
	}), MCPHandlerOptions{}), "Bearer owner")
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "advisor.messages_async", Arguments: map[string]any{"text": "private prompt", "client_request_id": "request"}})
	if err != nil {
		t.Fatal(err)
	}
	body := mcpTransportBody(t, result)
	handle, ok := body["async"].(map[string]any)
	if result.IsError || body["status"] != float64(202) || !ok || handle["id"] != "opaque-run" || handle["poll_tool"] != "advisor.run.get" || handle["cancel_tool"] != "advisor.run.cancel" || calls.Load() != 1 {
		t.Fatalf("async result=%+v calls=%d", result, calls.Load())
	}
}

func TestMCPAdvisorFiniteEventPage(t *testing.T) {
	session := mcpTransportSession(t, NewMCPHandler(DefaultExternalOperationCatalog(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("after_seq") != "7" {
			t.Error("cursor not forwarded")
		}
		w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
		_, _ = io.WriteString(w, "id: 8\nevent: completed\ndata: {\"v\":1,\"run_id\":\"run\",\"seq\":8,\"type\":\"completed\",\"payload\":{\"text\":\"answer\"}}\n\n")
	}), MCPHandlerOptions{}), "Bearer owner")
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "advisor.run.events", Arguments: map[string]any{"id": "run", "after_seq": 7}})
	if err != nil {
		t.Fatal(err)
	}
	body, ok := mcpTransportBody(t, result)["body"].(map[string]any)
	if !ok || body["after_seq"] != float64(8) || body["replay_only"] != true {
		t.Fatalf("event page=%+v", result)
	}
	events, ok := body["events"].([]any)
	if !ok || len(events) != 1 || events[0].(map[string]any)["type"] != "completed" {
		t.Fatalf("events=%+v", body)
	}
}

func TestMCPAdvisorEventLimitAndMalformedData(t *testing.T) {
	var overLimit strings.Builder
	for i := 1; i <= 101; i++ {
		fmt.Fprintf(&overLimit, "id: %d\nevent: progress\ndata: {\"seq\":%d}\n\n", i, i)
	}
	for _, tc := range []struct{ name, payload string }{
		{"limit", overLimit.String()},
		{"malformed", "id: 1\nevent: progress\ndata: private malformed event\n\n"},
		{"trailing-data", "id: 1\nevent: progress\ndata: {\"seq\":1} private malformed event\n\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			session := mcpTransportSession(t, NewMCPHandler(DefaultExternalOperationCatalog(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/event-stream")
				_, _ = io.WriteString(w, tc.payload)
			}), MCPHandlerOptions{}), "Bearer owner")
			result, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "advisor.run.events", Arguments: map[string]any{"id": "run"}})
			if err != nil {
				t.Fatal(err)
			}
			if !result.IsError || strings.Contains(string(mustMarshalMCP(t, result)), "private malformed") {
				t.Fatalf("invalid replay returned: %+v", result)
			}
		})
	}
}

func TestMCPAdvisorEventBoundaryAndEmptyPage(t *testing.T) {
	for _, count := range []int{0, 100} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			var replay strings.Builder
			for i := 8; i < 8+count; i++ {
				fmt.Fprintf(&replay, "id: %d\nevent: progress\ndata: {\"seq\":%d}\n\n", i, i)
			}
			session := mcpTransportSession(t, NewMCPHandler(DefaultExternalOperationCatalog(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/event-stream")
				_, _ = io.WriteString(w, replay.String())
			}), MCPHandlerOptions{}), "Bearer owner")
			result, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "advisor.run.events", Arguments: map[string]any{"id": "run", "after_seq": 7}})
			if err != nil {
				t.Fatal(err)
			}
			body, ok := mcpTransportBody(t, result)["body"].(map[string]any)
			if result.IsError || !ok || body["after_seq"] != float64(7+count) || len(body["events"].([]any)) != count {
				t.Fatalf("event boundary=%+v", result)
			}
		})
	}
}

func TestMCPAIAndAdvisorCanonicalFlows(t *testing.T) {
	for _, tc := range []struct {
		name, method, path string
		args               map[string]any
		status             int
	}{
		{"advisor.capabilities", "GET", "/api/v1/ai/advisor/capabilities", nil, 200},
		{"advisor.overview", "GET", "/api/v1/ai/advisor/overview", nil, 200},
		{"advisor.history", "GET", "/api/v1/ai/advisor/conversation/messages", map[string]any{"conversation_id": "conversation"}, 200},
		{"advisor.reset", "POST", "/api/v1/ai/advisor/conversation/reset", nil, 200},
		{"advisor.messages", "POST", "/api/v1/ai/advisor/messages", map[string]any{"text": "question", "client_request_id": "request"}, 200},
		{"advisor.run.get", "GET", "/api/v1/ai/advisor/runs/run", map[string]any{"id": "run"}, 200},
		{"advisor.run.cancel", "POST", "/api/v1/ai/advisor/runs/run/cancel", map[string]any{"id": "run"}, 204},
		{"advisor.action.confirm", "POST", "/api/v1/ai/advisor/actions/action/confirm", map[string]any{"action_id": "action", "conversation_id": "conversation", "generation": 2}, 409},
		{"ai_entry.capabilities", "GET", "/api/v1/ai/entry/capabilities", nil, 200},
		{"ai_entry.proposal.edit", "PATCH", "/api/v1/ai/entry/proposals/proposal", map[string]any{"id": "proposal", "version": 1, "draft": map[string]any{}}, 400},
		{"ai_entry.proposal.approve", "POST", "/api/v1/ai/entry/proposals/proposal/approve", map[string]any{"id": "proposal", "version": 1}, 200},
		{"ai_entry.proposal.reject", "POST", "/api/v1/ai/entry/proposals/proposal/reject", map[string]any{"id": "proposal", "version": 1}, 200},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.args == nil {
				tc.args = map[string]any{}
			}
			var calls atomic.Int32
			session := mcpTransportSession(t, NewMCPHandler(DefaultExternalOperationCatalog(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				if r.Method != tc.method || r.URL.Path != tc.path {
					t.Errorf("dispatch=%s %s", r.Method, r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				if tc.status != 204 {
					_, _ = io.WriteString(w, `{"code":"canonical","data":{"id":"owned"}}`)
				}
			}), MCPHandlerOptions{}), "Bearer owner")
			result, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: tc.name, Arguments: tc.args, Meta: mcp.Meta{"idempotency_key": "stable-reset"}})
			if err != nil {
				t.Fatal(err)
			}
			if mcpTransportBody(t, result)["status"] != float64(tc.status) || result.IsError != (tc.status >= 400) || calls.Load() != 1 {
				t.Fatalf("canonical result=%+v calls=%d", result, calls.Load())
			}
		})
	}
}

func TestMCPAdvisorConfirmRejectsCallerProposal(t *testing.T) {
	var calls atomic.Int32
	session := mcpTransportSession(t, NewMCPHandler(DefaultExternalOperationCatalog(), http.HandlerFunc(func(http.ResponseWriter, *http.Request) { calls.Add(1) }), MCPHandlerOptions{}), "Bearer owner")
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "advisor.action.confirm", Arguments: map[string]any{"action_id": "action", "conversation_id": "conversation", "generation": 2, "amount": 123, "draft": map[string]any{"wallet_id": "foreign"}}})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError || calls.Load() != 0 {
		t.Fatal("caller proposal crossed server-owned action provenance boundary")
	}
}

func TestMCPAttachmentResourceLinkAndFallback(t *testing.T) {
	const signed = "https://private.example/object?signature=private-secret"
	sink := &mcpAuditCapture{}
	session := mcpTransportSession(t, NewMCPHandler(DefaultExternalOperationCatalog(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", signed)
		w.WriteHeader(302)
		_, _ = fmt.Fprintf(w, "<a href=%q>Found</a>", signed)
	}), MCPHandlerOptions{Audit: sink}), "Bearer owner")
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "transactions.attachment.download", Arguments: map[string]any{"id": "transaction", "attachmentId": "attachment"}})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("attachment redirect became failure: %+v", result)
	}
	var link *mcp.ResourceLink
	for _, content := range result.Content {
		if value, ok := content.(*mcp.ResourceLink); ok {
			link = value
		}
	}
	if link == nil || link.URI != signed {
		t.Fatal("SDK resource link missing")
	}
	structured := string(mustMarshalMCP(t, result.StructuredContent))
	if strings.Contains(structured, "signature") || !strings.Contains(structured, "/api/v1/transactions/transaction/attachments/attachment/download") {
		t.Fatal("signed URL leaked into structured fallback or fallback missing")
	}
	events := sink.await(t, 1)
	if len(events) != 1 || !events[0].Allowed || events[0].Status != "302" || strings.Contains(string(mustMarshalMCP(t, events)), "private-secret") {
		t.Fatalf("attachment audit missing or leaked signed URL: %+v", events)
	}
}

func TestMCPAIEntryMultipartRecoveryAndHandle(t *testing.T) {
	var submissions atomic.Int32
	session := mcpTransportSession(t, NewMCPHandler(DefaultExternalOperationCatalog(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			submissions.Add(1)
			if err := r.ParseMultipartForm(1 << 20); err != nil {
				t.Error(err)
				return
			}
			defer r.MultipartForm.RemoveAll()
			images, err := parseAIEntryFiles(r.MultipartForm.File["files"])
			if err != nil || len(images) != 20 || images[0].Name != "00.pdf" || images[19].Name != "19.pdf" {
				t.Errorf("multipart order/count invalid: %v", err)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"data":{"id":"entry-request","processing":true,"proposals":[]}}`)
	}), MCPHandlerOptions{}), "Bearer owner")
	files := []any{}
	for i := 0; i < 20; i++ {
		files = append(files, map[string]any{"filename": fmt.Sprintf("%02d.pdf", i), "mime_type": "application/pdf", "base64": base64.StdEncoding.EncodeToString([]byte("%PDF-1.4\nsynthetic"))})
	}
	for _, p := range []*mcp.CallToolParams{
		{Name: "ai_entry.process", Arguments: map[string]any{"request_id": "entry-request", "timezone": "UTC", "files": files}},
		{Name: "ai_entry.request.get", Arguments: map[string]any{"request_id": "entry-request"}},
	} {
		result, err := session.CallTool(context.Background(), p)
		if err != nil {
			t.Fatal(err)
		}
		handle, ok := mcpTransportBody(t, result)["async"].(map[string]any)
		if result.IsError || !ok || handle["id"] != "entry-request" || handle["poll_tool"] != "ai_entry.request.get" {
			t.Fatalf("entry handle=%+v", result)
		}
	}
	if submissions.Load() != 1 {
		t.Fatal("recovery resubmitted process")
	}
}
