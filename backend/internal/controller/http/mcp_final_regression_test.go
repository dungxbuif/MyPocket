package httpapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
	"github.com/mypocket/backend/internal/usecase"
)

func mcpFeedbackPNG(size int) []byte {
	data := make([]byte, size)
	copy(data, "\x89PNG\r\n\x1a\n")
	return data
}

func TestMCPFeedbackScreenshotMultipartCanonicalParser(t *testing.T) {
	for _, size := range []int{8, 1 << 20} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			png := mcpFeedbackPNG(size)
			sink := &mcpAuditCapture{}
			var calls int
			session := mcpTransportSession(t, NewMCPHandler(DefaultExternalOperationCatalog(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				c, _ := gin.CreateTestContext(w)
				c.Request = r
				input, err := parseFeedbackInput(c)
				if err != nil {
					t.Error(err)
					w.WriteHeader(400)
					return
				}
				if r.MultipartForm != nil {
					defer r.MultipartForm.RemoveAll()
				}
				if input.Screenshot == nil {
					t.Error("canonical screenshot missing")
					return
				}
				got, _ := io.ReadAll(input.Screenshot)
				if !bytes.Equal(got, png) || input.ScreenshotSize != int64(size) || input.ScreenshotMIME != "image/png" || input.Type != "bug" || input.Title != "Private title" || input.Description != "Private description" {
					t.Errorf("canonical fields/bytes changed: %+v", input)
				}
				w.WriteHeader(201)
				_, _ = io.WriteString(w, `{"data":{"id":"feedback"}}`)
			}), MCPHandlerOptions{Audit: sink}), "Bearer private-bearer")
			listed, err := session.ListTools(context.Background(), nil)
			if err != nil {
				t.Fatal(err)
			}
			var schema string
			for _, tool := range listed.Tools {
				if tool.Name == "feedback.create" {
					schema = string(mustMarshalMCP(t, tool.InputSchema))
				}
			}
			if !strings.Contains(schema, `"screenshot"`) || !strings.Contains(schema, `"base64"`) {
				t.Errorf("discoverable screenshot schema missing: %s", schema)
			}
			result, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "feedback.create", Arguments: map[string]any{
				"type": "bug", "title": "Private title", "description": "Private description",
				"screenshot": map[string]any{"filename": "screen.png", "mime_type": "image/png", "base64": base64.StdEncoding.EncodeToString(png)},
			}})
			if err != nil || result.IsError || mcpTransportBody(t, result)["status"] != float64(201) || calls != 1 {
				t.Fatalf("screenshot not dispatched: result=%+v err=%v calls=%d", result, err, calls)
			}
			for _, secret := range []string{"private-bearer", "Private title", "Private description", "iVBOR"} {
				if strings.Contains(string(mustMarshalMCP(t, sink.await(t, 1))), secret) {
					t.Fatalf("audit contains %s", secret)
				}
			}
		})
	}
}

func TestMCPFeedbackScreenshotRejectsInvalidUpload(t *testing.T) {
	valid := base64.StdEncoding.EncodeToString(mcpFeedbackPNG(8))
	for _, tc := range []struct{ name, filename, mime, encoded string }{
		{"too-large", "screen.png", "image/png", base64.StdEncoding.EncodeToString(mcpFeedbackPNG((1 << 20) + 1))},
		{"empty", "screen.png", "image/png", ""},
		{"invalid-base64", "screen.png", "image/png", "private invalid bytes"},
		{"mismatched-mime", "screen.png", "image/jpeg", valid},
		{"pdf", "screen.png", "image/png", base64.StdEncoding.EncodeToString([]byte("%PDF-1.4\n"))},
		{"path", "../screen.png", "image/png", valid},
		{"long-name", strings.Repeat("a", 256), "image/png", valid},
	} {
		t.Run(tc.name, func(t *testing.T) {
			op, _ := DefaultExternalOperationCatalog().Lookup("feedback.create")
			if _, err := op.Encode(map[string]any{"type": "bug", "title": "title", "description": "description", "screenshot": map[string]any{"filename": tc.filename, "mime_type": tc.mime, "base64": tc.encoded}}, nil); err == nil {
				t.Fatal("accepted invalid screenshot")
			}
		})
	}
	op, _ := DefaultExternalOperationCatalog().Lookup("feedback.create")
	for _, extra := range []string{"owner_id", "userId", "provider_secret"} {
		if _, err := op.Encode(map[string]any{"screenshot": map[string]any{"filename": "screen.png", "mime_type": "image/png", "base64": valid, extra: "foreign"}}, nil); err == nil {
			t.Fatalf("accepted nested field %s", extra)
		}
	}
	if _, err := op.Encode(map[string]any{"description": strings.Repeat("x", 2<<20), "screenshot": map[string]any{"filename": "screen.png", "mime_type": "image/png", "base64": valid}}, nil); err == nil {
		t.Fatal("accepted multipart beyond canonical body bound")
	}
}

func TestMCPFeedbackTextOnlyRetainsJSON(t *testing.T) {
	op, _ := DefaultExternalOperationCatalog().Lookup("feedback.create")
	r, err := op.Encode(map[string]any{"type": "bug", "title": "Text", "description": "Text only"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if r.Header.Get("Content-Type") != "application/json" {
		t.Fatal("text-only flow stopped using JSON")
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = r
	input, err := parseFeedbackInput(c)
	if err != nil || input.Screenshot != nil || input.Title != "Text" || input.Description != "Text only" {
		t.Fatalf("text-only canonical input=%+v err=%v", input, err)
	}
}

type mcpFinalScreenshotRepo struct{ repository.FeedbackRepository }

func (*mcpFinalScreenshotRepo) FindByOwner(_ context.Context, owner, id string) (*entity.Feedback, error) {
	if owner != "owner-1" || id != "owned-feedback" {
		return nil, repository.ErrNotFound
	}
	return &entity.Feedback{ID: id, UserID: owner, ScreenshotObjectKey: "private-owner-key"}, nil
}

type mcpFinalScreenshotStorage struct{ usecase.FeedbackStorage }

func (*mcpFinalScreenshotStorage) SignedGet(_ context.Context, key string) (string, error) {
	if key != "private-owner-key" {
		return "", fmt.Errorf("unexpected storage key")
	}
	return "https://private.example/screenshot?signature=private-owner-signature", nil
}

func TestMCPFeedbackScreenshotCanonicalOwnerBoundary(t *testing.T) {
	service := usecase.NewFeedbackServiceWithStorage(&mcpFinalScreenshotRepo{}, nil, &mcpFinalScreenshotStorage{}, nil)
	router, bearer := mcpFinalRouter(t, &mcpAuditCapture{}, time.Second, &FeedbackHandler{Service: service})
	session := mcpTransportSession(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = "/api/v1/mcp"
		router.Engine.ServeHTTP(w, r)
	}), bearer)
	for _, id := range []string{"owned-feedback", "foreign-feedback"} {
		result, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "feedback.screenshot", Arguments: map[string]any{"id": id}})
		if err != nil {
			t.Fatal(err)
		}
		want := float64(302)
		if id == "foreign-feedback" {
			want = 404
		}
		if mcpTransportBody(t, result)["status"] != want || result.IsError != (want == 404) {
			t.Fatalf("owner boundary id=%s result=%+v", id, result)
		}
		if strings.Contains(string(mustMarshalMCP(t, result.StructuredContent)), "private-owner") {
			t.Fatal("private storage value escaped structured content")
		}
	}
}

func TestMCPFeedbackScreenshotResourceLinkAndDenial(t *testing.T) {
	for _, status := range []int{302, 404} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			const signed = "https://private.example/feedback?signature=private-signed-secret"
			sink := &mcpAuditCapture{}
			session := mcpTransportSession(t, NewMCPHandler(DefaultExternalOperationCatalog(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if status == 302 {
					http.Redirect(w, r, signed, 302)
				} else {
					w.Header().Set("Content-Type", "application/problem+json")
					w.WriteHeader(status)
					_, _ = io.WriteString(w, `{"code":"NOT_FOUND","status":404}`)
				}
			}), MCPHandlerOptions{Audit: sink}), "Bearer owner")
			result, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "feedback.screenshot", Arguments: map[string]any{"id": "feedback"}})
			if err != nil || result.IsError != (status == 404) {
				t.Fatalf("screenshot status=%d result=%+v err=%v", status, result, err)
			}
			var link *mcp.ResourceLink
			for _, content := range result.Content {
				if resource, ok := content.(*mcp.ResourceLink); ok {
					link = resource
				} else if strings.Contains(string(mustMarshalMCP(t, content)), "signature") {
					t.Fatal("signed URL leaked into text")
				}
			}
			if status == 302 && (link == nil || link.URI != signed || !strings.Contains(string(mustMarshalMCP(t, result.StructuredContent)), `"path":"/api/v1/feedback/feedback/screenshot"`)) {
				t.Fatal("successful screenshot resource/fallback missing")
			}
			if status == 404 && link != nil {
				t.Fatal("owner denial returned resource link")
			}
			if strings.Contains(string(mustMarshalMCP(t, result.StructuredContent)), "signature") || strings.Contains(string(mustMarshalMCP(t, sink.await(t, 1))), "private-signed-secret") {
				t.Fatal("signed URL leaked into structured content/audit")
			}
		})
	}
}

func mcpFinalRouter(t *testing.T, sink repository.AuditSink, timeout time.Duration, feedback *FeedbackHandler) (*Router, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	keys := usecase.NewUserAPIKeyService(&feedbackRouteAPIKeyRepo{})
	key, err := keys.Create(context.Background(), usecase.APIKeyCreateInput{OwnerID: "owner-1", Name: "Synthetic MCP"})
	if err != nil {
		t.Fatal(err)
	}
	router := NewApplicationRouter(ApplicationHandlers{Middleware: &AuthMiddleware{APIKeys: keys, Audit: sink}, Feedback: feedback}, ApplicationRouterOptions{MCPOptions: MCPHandlerOptions{RequestTimeout: timeout}})
	return router, "Bearer " + key.Secret
}

func TestMCPAdvisorCanonicalStalledAuditDoesNotBlock(t *testing.T) {
	for _, deadline := range []bool{false, true} {
		t.Run(fmt.Sprint(deadline), func(t *testing.T) {
			sink := &mcpStalledAuditSink{started: make(chan struct{}, 1), release: make(chan struct{})}
			release := sync.OnceFunc(func() { close(sink.release) })
			defer release()
			router, bearer := mcpFinalRouter(t, sink, 30*time.Millisecond, nil)
			if deadline {
				// A cooperative canonical operation waits for its deadline;
				// its advisor middleware must still not block on audit afterward.
				router.Engine.GET(advisorRoute+"/deadline-test", router.AuthMiddleware.RequireAdvisorAuth, func(c *gin.Context) { <-c.Request.Context().Done() })
				op, _ := DefaultExternalOperationCatalog().Lookup("advisor.capabilities")
				op.PathTemplate = advisorRoute + "/deadline-test"
				catalog, err := NewExternalOperationCatalog([]ExternalOperation{op})
				if err != nil {
					t.Fatal(err)
				}
				router.Engine.POST("/mcp-deadline", router.AuthMiddleware.RequireAuth, MCPAuthenticatedHandler(NewMCPHandler(catalog, router.Engine, MCPHandlerOptions{Audit: sink, RequestTimeout: 30 * time.Millisecond})))
			}
			var session *mcp.ClientSession
			if deadline {
				session = mcpTransportSession(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					r.URL.Path = "/mcp-deadline"
					router.Engine.ServeHTTP(w, r)
				}), bearer)
			} else {
				session = mcpTransportSession(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					r.URL.Path = "/api/v1/mcp"
					router.Engine.ServeHTTP(w, r)
				}), bearer)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
			defer cancel()
			result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "advisor.capabilities", Arguments: map[string]any{}})
			if err != nil {
				release()
				t.Fatalf("canonical audit stalled MCP: %v", err)
			}
			want := float64(200)
			if deadline {
				want = 504
			}
			if mcpTransportBody(t, result)["status"] != want {
				t.Fatalf("result=%+v want=%v", result, want)
			}
			select {
			case <-sink.started:
			case <-time.After(time.Second):
				t.Fatal("outer MCP audit missing")
			}
			if sink.calls.Load() != 1 {
				t.Fatalf("duplicate audit calls=%d", sink.calls.Load())
			}
		})
	}
}

func TestMCPAdvisorDirectRESTAuditCannotBeForgedAway(t *testing.T) {
	sink := &mcpAuditCapture{}
	router, bearer := mcpFinalRouter(t, sink, time.Second, nil)
	request := httptest.NewRequest("GET", advisorRoute+"/capabilities", nil)
	request.Header.Set("Authorization", bearer)
	request.Header.Set("X-MCP-Dispatch", "true")
	recorder := httptest.NewRecorder()
	router.Engine.ServeHTTP(recorder, request)
	events := sink.snapshot()
	if recorder.Code != 200 || len(events) != 1 || events[0].Action != "advisor.get.capabilities" || !events[0].Allowed {
		t.Fatalf("REST advisor audit changed: status=%d events=%+v", recorder.Code, events)
	}
}

type mcpFinalFeedbackRepo struct{ repository.FeedbackRepository }

func (*mcpFinalFeedbackRepo) Create(_ context.Context, row *entity.Feedback) error { return nil }

func (*mcpFinalFeedbackRepo) UpdateResolution(_ context.Context, owner, id string, resolved bool, _ time.Time) (*entity.Feedback, error) {
	return &entity.Feedback{ID: id, UserID: owner, Resolved: resolved}, nil
}

type mcpFinalDeadlineFeedbackRepo struct{ mcpFinalFeedbackRepo }

func (*mcpFinalDeadlineFeedbackRepo) Create(ctx context.Context, _ *entity.Feedback) error {
	<-ctx.Done()
	return nil // Acceptance raced the caller's deadline; domain audit still runs.
}

func TestMCPFeedbackCanonicalDeadlineWithStalledAudit(t *testing.T) {
	sink := &mcpStalledAuditSink{started: make(chan struct{}, 1), release: make(chan struct{})}
	release := sync.OnceFunc(func() { close(sink.release) })
	defer release()
	router, bearer := mcpFinalRouter(t, sink, 30*time.Millisecond, &FeedbackHandler{Service: usecase.NewFeedbackService(&mcpFinalDeadlineFeedbackRepo{}, sink, nil)})
	session := mcpTransportSession(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = "/api/v1/mcp"
		router.Engine.ServeHTTP(w, r)
	}), bearer)
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()
	result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "feedback.create", Arguments: map[string]any{"type": "bug", "title": "Synthetic", "description": "Synthetic description"}})
	if err != nil {
		release()
		t.Fatalf("feedback audit blocked deadline result: %v", err)
	}
	if mcpTransportBody(t, result)["status"] != float64(504) || !strings.Contains(string(mustMarshalMCP(t, result)), "MCP_DEADLINE_EXCEEDED") {
		t.Fatalf("feedback deadline result=%+v", result)
	}
}

func TestMCPFeedbackCanonicalStalledDomainAuditDoesNotBlock(t *testing.T) {
	sink := &mcpStalledAuditSink{started: make(chan struct{}, 1), release: make(chan struct{})}
	release := sync.OnceFunc(func() { close(sink.release) })
	defer release()
	router, bearer := mcpFinalRouter(t, sink, 30*time.Millisecond, &FeedbackHandler{Service: usecase.NewFeedbackService(&mcpFinalFeedbackRepo{}, sink, nil)})
	session := mcpTransportSession(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = "/api/v1/mcp"
		router.Engine.ServeHTTP(w, r)
	}), bearer)
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()
	result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "feedback.create", Arguments: map[string]any{"type": "bug", "title": "Synthetic feedback", "description": "Synthetic description"}})
	if err != nil {
		release()
		t.Fatalf("feedback domain audit stalled MCP: %v", err)
	}
	if result.IsError || mcpTransportBody(t, result)["status"] != float64(201) {
		t.Fatalf("feedback result=%+v", result)
	}
	// Both domain and MCP queues must stay bounded under the same blocked sink.
	select {
	case <-sink.started:
	case <-time.After(time.Second):
		t.Fatal("audit missing")
	}
	for i := 0; i < 80; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
		result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "feedback.set_resolution", Arguments: map[string]any{"id": "feedback", "resolved": true}})
		cancel()
		if err != nil || result.IsError {
			release()
			t.Fatalf("saturated domain audit stalled call %d: %v", i, err)
		}
	}
	if sink.calls.Load() != 2 {
		t.Fatalf("want one worker per domain/MCP queue, got %d", sink.calls.Load())
	}
	release()
	end := time.Now().Add(time.Second)
	for sink.calls.Load() < 130 && time.Now().Before(end) {
		time.Sleep(time.Millisecond)
	}
	if sink.calls.Load() != 130 {
		t.Fatalf("bounded domain/MCP delivery=%d want 130", sink.calls.Load())
	}
}

func TestMCPFeedbackDomainEventsPreservedAndDirectRESTSynchronous(t *testing.T) {
	t.Run("MCP preserves domain events", func(t *testing.T) {
		sink := &mcpAuditCapture{}
		service := usecase.NewFeedbackService(&mcpFinalFeedbackRepo{}, sink, nil)
		router, bearer := mcpFinalRouter(t, sink, time.Second, &FeedbackHandler{Service: service})
		session := mcpTransportSession(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.URL.Path = "/api/v1/mcp"
			router.Engine.ServeHTTP(w, r)
		}), bearer)
		for _, call := range []*mcp.CallToolParams{
			{Name: "feedback.create", Arguments: map[string]any{"type": "bug", "title": "Synthetic", "description": "Synthetic description"}},
			{Name: "feedback.set_resolution", Arguments: map[string]any{"id": "feedback", "resolved": true}},
		} {
			result, err := session.CallTool(context.Background(), call)
			if err != nil || result.IsError {
				t.Fatalf("feedback call=%+v %v", result, err)
			}
		}
		events := sink.await(t, 4)
		actions := map[string]repository.AuditEvent{}
		for _, event := range events {
			actions[event.Action] = event
		}
		if len(events) != 4 || len(actions) != 4 || actions["feedback.created"].FeedbackID == "" || actions["feedback.owner_resolution"].FeedbackID != "feedback" || actions["feedback.owner_resolution"].Reason != "resolved" || !actions["mcp.feedback.create"].Allowed || !actions["mcp.feedback.set_resolution"].Allowed {
			t.Fatalf("domain/MCP audit events=%+v", events)
		}
		if service.Audit != sink {
			t.Fatal("router mutated supplied feedback service")
		}
	})
	t.Run("REST remains synchronous", func(t *testing.T) {
		sink := &mcpStalledAuditSink{started: make(chan struct{}, 1), release: make(chan struct{})}
		release := sync.OnceFunc(func() { close(sink.release) })
		defer release()
		router, bearer := mcpFinalRouter(t, sink, time.Second, &FeedbackHandler{Service: usecase.NewFeedbackService(&mcpFinalFeedbackRepo{}, sink, nil)})
		done := make(chan int, 1)
		go func() {
			r := httptest.NewRequest("POST", "/api/v1/feedback", strings.NewReader(`{"type":"bug","title":"Synthetic","description":"Synthetic description"}`))
			r.Header.Set("Content-Type", "application/json")
			r.Header.Set("Authorization", bearer)
			r.Header.Set("X-MCP-Dispatch", "true")
			w := httptest.NewRecorder()
			router.Engine.ServeHTTP(w, r)
			done <- w.Code
		}()
		select {
		case <-sink.started:
		case <-time.After(time.Second):
			t.Fatal("REST audit missing")
		}
		select {
		case <-done:
			t.Fatal("REST audit became asynchronous")
		default:
		}
		release()
		select {
		case status := <-done:
			if status != 201 {
				t.Fatalf("REST status=%d", status)
			}
		case <-time.After(time.Second):
			t.Fatal("REST did not finish")
		}
	})
}

func TestMCPFeedbackDomainAuditBoundsRequestIdentifier(t *testing.T) {
	sink := &mcpAuditCapture{}
	router, bearer := mcpFinalRouter(t, sink, time.Second, &FeedbackHandler{Service: usecase.NewFeedbackService(&mcpFinalFeedbackRepo{}, sink, nil)})
	session := mcpTransportSession(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = "/api/v1/mcp"
		r.Header.Set(requestIDHeader, strings.Repeat("a", 4096))
		router.Engine.ServeHTTP(w, r)
	}), bearer)
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "feedback.create", Arguments: map[string]any{"type": "bug", "title": "Synthetic", "description": "Synthetic description"}})
	if err != nil || result.IsError {
		t.Fatalf("feedback=%+v %v", result, err)
	}
	for _, event := range sink.await(t, 2) {
		if len(event.RequestID) != 128 {
			t.Fatalf("queued %s retained request identifier length=%d", event.Action, len(event.RequestID))
		}
	}
}
