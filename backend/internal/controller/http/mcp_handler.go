package httpapi

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/mypocket/backend/internal/repository"
	"github.com/mypocket/backend/internal/usecase"
)

const (
	mcpImplementationName          = "mypocket"
	mcpImplementationVersion       = "1.0.0"
	mcpDefaultRequestBytes   int64 = 140 << 20 // Includes base64 overhead for the REST multipart limit.
	mcpDefaultResponseBytes  int64 = 8 << 20
	mcpDefaultRequestTimeout       = 2 * time.Minute
	mcpInvalidArgumentsCode        = "MCP_INVALID_ARGUMENTS"
	mcpResponseLimitCode           = "MCP_RESPONSE_TOO_LARGE"
	mcpDeadlineCode                = "MCP_DEADLINE_EXCEEDED"
	mcpCancelledCode               = "MCP_REQUEST_CANCELLED"
	mcpReplayLimit                 = 100
	mcpAuditQueueCapacity          = 64
)

// MCPHandlerOptions bounds transport resources, without changing owner authority.
// Origins without an Origin header are allowed for non-browser MCP clients.
type MCPHandlerOptions struct {
	Audit            repository.AuditSink
	AllowedOrigins   []string
	MaxRequestBytes  int64
	MaxResponseBytes int64
	RequestTimeout   time.Duration
}

type mcpHTTPRequestContextKey struct{}
type mcpPrincipalContextKey struct{}
type mcpAuditMetricsContextKey struct{}

// Only server-side in-process dispatch can create this key. HTTP headers and
// tool arguments cannot opt out of canonical REST audit.
type mcpAuditedDispatchContextKey struct{}

type mcpAuditMetrics struct {
	requestBytes  int64
	responseBytes int64
}

// The fixed ring contains redacted scalars only, never request contexts,
// arguments or response payloads. A stalled sink pins at most this ring and one
// drain worker. Idle workers exit; enqueue never waits for sink I/O.
type mcpAuditQueue struct {
	mu         sync.Mutex
	sink       repository.AuditSink
	events     [mcpAuditQueueCapacity]repository.AuditEvent
	head, size int
	running    bool
}

func (q *mcpAuditQueue) enqueue(event repository.AuditEvent) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.size == len(q.events) {
		return
	}
	q.events[(q.head+q.size)%len(q.events)] = event
	q.size++
	if !q.running {
		q.running = true
		go q.drain()
	}
}

func (q *mcpAuditQueue) drain() {
	for {
		q.mu.Lock()
		if q.size == 0 {
			q.running = false
			q.mu.Unlock()
			return
		}
		event := q.events[q.head]
		q.events[q.head] = repository.AuditEvent{}
		q.head = (q.head + 1) % len(q.events)
		q.size--
		q.mu.Unlock()
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		_ = q.sink.Record(ctx, event)
		cancel()
	}
}

// Feedback domain events remain distinct from operation access audits. Only
// audited MCP dispatch uses the bounded delivery queue; direct REST retains
// its synchronous sink contract. Composition copies the service, so callers'
// shared dependencies are never mutated.
type mcpFeedbackAuditSink struct {
	direct repository.AuditSink
	queue  *mcpAuditQueue
}

func (s *mcpFeedbackAuditSink) Record(ctx context.Context, event repository.AuditEvent) error {
	if audited, _ := ctx.Value(mcpAuditedDispatchContextKey{}).(bool); audited {
		// User feedback dispatch emits these two scalar-only domain events.
		// Bound and copy caller-derived identifiers so a stalled sink cannot
		// retain a large header allocation; never queue service-agent ID lists.
		s.queue.enqueue(repository.AuditEvent{
			RequestID: boundedMCPAuditID(event.RequestID), ActorKind: boundedMCPAuditID(event.ActorKind),
			Action: boundedMCPAuditID(event.Action), FeedbackID: boundedMCPAuditID(event.FeedbackID), Reason: boundedMCPAuditID(event.Reason),
		})
		return nil
	}
	return s.direct.Record(ctx, event)
}

func mcpFeedbackHandler(handler *FeedbackHandler) *FeedbackHandler {
	if handler.Service == nil || handler.Service.Audit == nil {
		return handler
	}
	copyHandler, copyService := *handler, *handler.Service
	sink := copyService.Audit
	copyService.Audit = &mcpFeedbackAuditSink{direct: sink, queue: &mcpAuditQueue{sink: sink}}
	copyHandler.Service = &copyService
	return &copyHandler
}

// MCPAuthenticatedHandler bridges the already validated Gin principal into
// SDK request context. It must run after RequireAuth and never resolves bearer
// credentials itself.
func MCPAuthenticatedHandler(handler http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, _ := advisorPrincipalFromContext(c)
		ctx := context.WithValue(c.Request.Context(), mcpPrincipalContextKey{}, principal)
		request := c.Request.Clone(ctx)
		if id := requestID(c); id != "" {
			request.Header.Set(requestIDHeader, id)
		}
		handler.ServeHTTP(c.Writer, request)
	}
}

// NewMCPHandler exposes only catalog tools. Mount this handler behind the same
// AuthMiddleware.RequireAuth as REST, and pass the canonical router as dispatcher.
// The SDK owns all MCP envelopes, negotiation and protocol errors.
//
// @Summary MCP authenticated user-operation transport
// @Description Official Go SDK v1.8.0 stateless Streamable HTTP with JSON responses. JWT and mpk_ API keys have identical same-owner user authority without scopes. POST carries initialize, tools/list and tools/call; tools/list publishes authoritative tool names and inputSchema values. Ordinary tools may write directly; Assistant prepared actions retain their canonical confirmation contract. GET and DELETE normally return 405 in stateless mode. Credential bootstrap/exchange and service-token routes are excluded. See docs/ai/mcp-api-parity.md for recovery, limits, errors and audit privacy.
// @Tags MCP
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object false "SDK-owned MCP JSON-RPC request (POST only); discover business schemas through tools/list"
// @Success 200 {object} map[string]interface{} "SDK JSON-RPC response; inspect result.isError and structuredContent.status/body for domain outcomes"
// @Success 202 {string} string "Accepted SDK notification"
// @Failure 400 {object} map[string]interface{} "SDK invalid protocol request"
// @Failure 401 {object} Problem "Missing, invalid, expired or revoked owner credential"
// @Failure 403 {string} string "Origin is not allowed"
// @Failure 405 {string} string "Stateless transport accepts POST"
// @Failure 413 {string} string "Outer request exceeds configured input bound"
// @Router /api/v1/mcp [post]
// @Router /api/v1/mcp [get]
// @Router /api/v1/mcp [delete]
func NewMCPHandler(catalog *ExternalOperationCatalog, dispatcher http.Handler, options MCPHandlerOptions) http.Handler {
	if catalog == nil || dispatcher == nil {
		panic("MCP catalog and canonical dispatcher are required")
	}
	if options.MaxRequestBytes <= 0 {
		options.MaxRequestBytes = mcpDefaultRequestBytes
	}
	if options.MaxResponseBytes <= 0 {
		options.MaxResponseBytes = mcpDefaultResponseBytes
	}
	if options.RequestTimeout <= 0 {
		options.RequestTimeout = mcpDefaultRequestTimeout
	}
	server := mcp.NewServer(&mcp.Implementation{Name: mcpImplementationName, Version: mcpImplementationVersion}, nil)
	if options.Audit != nil {
		server.AddReceivingMiddleware(mcpAuditMiddleware(catalog, &mcpAuditQueue{sink: options.Audit}, options.MaxRequestBytes))
	}
	for _, operation := range catalog.List() {
		server.AddTool(&mcp.Tool{Name: operation.Name, Description: operation.Description, InputSchema: operation.InputSchema}, func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			ctx, cancel := context.WithTimeout(ctx, options.RequestTimeout)
			defer cancel()
			// Older SDK-negotiated protocols intentionally detach HTTP cancellation.
			// Preserve their protocol behavior while keeping synchronous REST writes
			// tied to this HTTP caller as required by MyPocket's dispatch contract.
			if outer, ok := ctx.Value(mcpHTTPRequestContextKey{}).(context.Context); ok {
				stop := context.AfterFunc(outer, cancel)
				defer stop()
				if outer.Err() != nil {
					cancel()
				}
			}
			arguments := make(map[string]any)
			if len(req.Params.Arguments) != 0 {
				decoder := json.NewDecoder(bytes.NewReader(req.Params.Arguments))
				decoder.UseNumber()
				if err := decoder.Decode(&arguments); err != nil || arguments == nil {
					return mcpDispatchProblem(http.StatusBadRequest, mcpInvalidArgumentsCode, "Tool arguments must be a JSON object."), nil
				}
			}
			// The SDK attaches negotiation metadata to each call. Only application
			// metadata belongs in the REST encoder; leave protocol fields SDK-owned.
			metadata := make(map[string]any)
			for key, value := range req.Params.Meta {
				switch key {
				case mcp.MetaKeyProtocolVersion, mcp.MetaKeyClientInfo, mcp.MetaKeyServerInfo, mcp.MetaKeyClientCapabilities, mcp.MetaKeyLogLevel, mcp.MetaKeySubscriptionID, "progressToken":
					continue
				default:
					metadata[key] = value
				}
			}
			inner, err := operation.Encode(arguments, metadata)
			if err != nil {
				// Encoder errors may contain invalid user input. Keep diagnostics
				// bounded and do not copy arguments into protocol/audit errors.
				return mcpDispatchProblem(http.StatusBadRequest, mcpInvalidArgumentsCode, "Tool arguments are invalid for this operation."), nil
			}
			if options.Audit != nil {
				ctx = context.WithValue(ctx, mcpAuditedDispatchContextKey{}, true)
			}
			inner = inner.WithContext(ctx)
			if req.Extra != nil {
				inner.Header.Set(authHeader, req.Extra.Header.Get(authHeader))
				inner.Header.Set(requestIDHeader, req.Extra.Header.Get(requestIDHeader))
			}
			if err := ctx.Err(); err != nil {
				return mcpContextProblem(err), nil
			}
			response := &mcpResponseBuffer{header: make(http.Header), limit: options.MaxResponseBytes}
			defer func() {
				if metrics, ok := ctx.Value(mcpAuditMetricsContextKey{}).(*mcpAuditMetrics); ok {
					metrics.responseBytes = response.observedBytes
				}
			}()
			dispatcher.ServeHTTP(response, inner)
			if err := ctx.Err(); err != nil {
				return mcpContextProblem(err), nil
			}
			if response.overflow {
				return mcpDispatchProblem(http.StatusBadGateway, mcpResponseLimitCode, "The operation response exceeds the MCP output limit."), nil
			}
			status := response.status
			if status == 0 {
				status = http.StatusOK
			}
			var body any
			decoder := json.NewDecoder(bytes.NewReader(response.body.Bytes()))
			decoder.UseNumber()
			if decoder.Decode(&body) != nil {
				body = response.body.String()
			}
			id := response.header.Get(requestIDHeader)
			if id == "" {
				if problem, ok := body.(map[string]any); ok {
					id, _ = problem["request_id"].(string)
				}
			}
			if id == "" {
				id = inner.Header.Get(requestIDHeader)
			}
			return mcpOperationResult(operation, inner, response, status, body, id), nil
		})
	}
	transport := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, &mcp.StreamableHTTPOptions{
		Stateless: true, JSONResponse: true, MaxRequestBodyBytes: options.MaxRequestBytes, PropagateRequestCancellation: true,
	})
	origins := make(map[string]struct{}, len(options.AllowedOrigins))
	for _, origin := range options.AllowedOrigins {
		if origin = strings.TrimSpace(origin); origin != "" {
			origins[origin] = struct{}{}
		}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			if _, ok := origins[origin]; !ok {
				http.Error(w, "MCP origin is not allowed", http.StatusForbidden)
				return
			}
		}
		if r.ContentLength > options.MaxRequestBytes {
			http.Error(w, "MCP request exceeds the input limit", http.StatusRequestEntityTooLarge)
			return
		}
		// Only the tool operation gets RequestTimeout. Expiring the SDK's HTTP
		// context at that deadline would prevent it from serializing the timeout
		// result. Actual caller cancellation still cancels both lifecycles.
		ctx := context.WithValue(r.Context(), mcpHTTPRequestContextKey{}, r.Context())
		if r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, options.MaxRequestBytes)
		}
		transport.ServeHTTP(w, r.WithContext(ctx))
	})
}

func mcpAuditMiddleware(catalog *ExternalOperationCatalog, queue *mcpAuditQueue, requestLimit int64) mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (result mcp.Result, err error) {
			params, ok := req.GetParams().(*mcp.CallToolParamsRaw)
			if method != "tools/call" || !ok || params == nil {
				return next(ctx, method, req)
			}
			// Never copy an unrecognized caller-supplied name into durable logs.
			name := "unknown"
			if operation, exists := catalog.Lookup(params.Name); exists {
				name = operation.Name
			}
			started := time.Now()
			metrics := &mcpAuditMetrics{requestBytes: min(int64(len(params.Arguments)), requestLimit)}
			ctx = context.WithValue(ctx, mcpAuditMetricsContextKey{}, metrics)
			defer func() {
				status := http.StatusBadRequest
				allowed := false
				if call, ok := result.(*mcp.CallToolResult); ok && call != nil {
					if value, ok := call.Meta["status"].(int); ok {
						status = value
					}
					allowed = err == nil && !call.IsError
				}
				reason := "denied"
				if allowed {
					reason = "allowed"
				} else if status >= http.StatusInternalServerError {
					reason = "error"
				}
				principal, _ := ctx.Value(mcpPrincipalContextKey{}).(usecase.Principal)
				actor := principal.CredentialKind
				if actor == "" {
					actor = "unknown"
				}
				id := ""
				if extra := req.GetExtra(); extra != nil {
					id = extra.Header.Get(requestIDHeader)
				}
				// Bound externally supplied identifiers before queueing. The queue
				// retains no caller context or private payload, and drops when full.
				queue.enqueue(repository.AuditEvent{RequestID: boundedMCPAuditID(id), ActorKind: boundedMCPAuditID(actor), CredentialID: boundedMCPAuditID(principal.CredentialID), Action: "mcp." + name, Allowed: allowed, Status: strconv.Itoa(status), Reason: reason, LatencyMS: time.Since(started).Milliseconds(), RequestBytes: metrics.requestBytes, ResponseBytes: metrics.responseBytes})
			}()
			return next(ctx, method, req)
		}
	}
}

func boundedMCPAuditID(value string) string {
	// Clone so a short identifier cannot retain an oversized header's backing
	// allocation for the lifetime of a stalled queue.
	return strings.Clone(value[:min(len(value), 128)])
}

func mcpOperationResult(operation ExternalOperation, request *http.Request, response *mcpResponseBuffer, status int, body any, id string) *mcp.CallToolResult {
	if (operation.Name == "transactions.attachment.download" || operation.Name == "feedback.screenshot") && status == http.StatusFound {
		location := response.header.Get("Location")
		parsed, err := url.Parse(location)
		if err != nil || parsed.Host == "" || parsed.User != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || len(location) > 8192 || int64(len(location)) > response.limit {
			return mcpDispatchProblem(http.StatusBadGateway, "MCP_ATTACHMENT_UNAVAILABLE", "The attachment resource is unavailable.")
		}
		// Gin's redirect HTML embeds the signed URL. Discard it completely; only
		// the SDK ResourceLink carries that ephemeral credential to the caller.
		result := mcpHTTPResult(status, map[string]any{"rest_fallback": map[string]any{"method": "GET", "path": request.URL.Path}}, "application/json", id)
		result.IsError = false
		name := "transaction-attachment"
		if operation.Name == "feedback.screenshot" {
			name = "feedback-screenshot"
		}
		result.Content = append([]mcp.Content{&mcp.ResourceLink{URI: location, Name: name, Description: "Short-lived private attachment; use the authenticated REST fallback to renew access."}}, result.Content...)
		return result
	}
	if operation.Name == "advisor.run.events" && status >= 200 && status < 300 {
		after, _ := strconv.ParseInt(request.URL.Query().Get("after_seq"), 10, 64)
		page, err := mcpReplayPage(response.body.Bytes(), after)
		if err != nil {
			return mcpDispatchProblem(http.StatusBadGateway, "MCP_INVALID_EVENT_REPLAY", "The operation event replay is invalid or exceeds the event limit.")
		}
		body = page
	}
	result := mcpHTTPResult(status, body, response.header.Get("Content-Type"), id)
	if status < 200 || status >= 300 {
		return result
	}
	envelope, _ := body.(map[string]any)
	data, _ := envelope["data"].(map[string]any)
	var handle map[string]any
	switch operation.Name {
	case "advisor.messages_async":
		if run, ok := data["run_id"].(string); ok && run != "" {
			handle = map[string]any{"kind": "advisor_run", "id": run, "status": data["status"], "conversation_id": data["conversation_id"], "poll_tool": "advisor.run.get", "events_tool": "advisor.run.events", "cancel_tool": "advisor.run.cancel"}
		}
	case "ai_entry.process", "ai_entry.request.get":
		if entry, ok := data["id"].(string); ok && entry != "" && (status == http.StatusAccepted || data["processing"] == true) {
			handle = map[string]any{"kind": "ai_entry_request", "id": entry, "processing": data["processing"], "poll_tool": "ai_entry.request.get"}
		}
	}
	if handle != nil {
		structured := result.StructuredContent.(map[string]any)
		structured["async"] = handle
		encoded, _ := json.Marshal(structured)
		result.Content = []mcp.Content{&mcp.TextContent{Text: string(encoded)}}
	}
	return result
}

// mcpReplayPage consumes finite, persisted REST replay only. A malformed or
// oversized page fails closed without returning partial event payloads.
func mcpReplayPage(raw []byte, after int64) (map[string]any, error) {
	events := make([]any, 0)
	data := make([]string, 0)
	flush := func() error {
		if len(data) == 0 {
			return nil
		}
		if len(events) >= mcpReplayLimit {
			return errors.New("event limit")
		}
		var event map[string]any
		decoder := json.NewDecoder(strings.NewReader(strings.Join(data, "\n")))
		decoder.UseNumber()
		if err := decoder.Decode(&event); err != nil || event == nil {
			return errors.New("invalid event")
		}
		if decoder.Decode(new(any)) != io.EOF {
			return errors.New("trailing event data")
		}
		sequence, ok := event["seq"].(json.Number)
		seq, err := sequence.Int64()
		if !ok || err != nil || seq <= after {
			return errors.New("invalid event sequence")
		}
		after = seq
		events = append(events, event)
		data = data[:0]
		return nil
	}
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	scanner.Buffer(make([]byte, 4096), len(raw)+1)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			if err := flush(); err != nil {
				return nil, err
			}
		} else if strings.HasPrefix(line, "data:") {
			data = append(data, strings.TrimPrefix(strings.TrimPrefix(line, "data:"), " "))
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if err := flush(); err != nil {
		return nil, err
	}
	return map[string]any{"events": events, "after_seq": after, "replay_only": true}, nil
}

func mcpContextProblem(err error) *mcp.CallToolResult {
	if errors.Is(err, context.DeadlineExceeded) {
		return mcpDispatchProblem(http.StatusGatewayTimeout, mcpDeadlineCode, "The operation deadline was exceeded.")
	}
	return mcpDispatchProblem(http.StatusRequestTimeout, mcpCancelledCode, "The operation was cancelled.")
}

func mcpDispatchProblem(status int, code, detail string) *mcp.CallToolResult {
	return mcpHTTPResult(status, map[string]any{"status": status, "code": code, "detail": detail}, "application/problem+json", "")
}

func mcpHTTPResult(status int, body any, contentType, id string) *mcp.CallToolResult {
	structured := map[string]any{"status": status, "body": body, "content_type": contentType, "request_id": id}
	encoded, _ := json.Marshal(structured)
	return &mcp.CallToolResult{
		IsError:           status < http.StatusOK || status >= http.StatusMultipleChoices,
		StructuredContent: structured,
		Content:           []mcp.Content{&mcp.TextContent{Text: string(encoded)}},
		Meta:              mcp.Meta{"status": status, "content_type": contentType, "request_id": id},
	}
}

// mcpResponseBuffer retains at most limit bytes, even when a canonical handler
// ignores a short write. It never returns a partial financial response to MCP.
type mcpResponseBuffer struct {
	header        http.Header
	status        int
	body          bytes.Buffer
	limit         int64
	overflow      bool
	observedBytes int64
}

func (w *mcpResponseBuffer) Header() http.Header { return w.header }

// Flush is a no-op because REST replay is collected before producing one MCP
// result; Gin's finite SSE handlers still require the http.Flusher interface.
func (w *mcpResponseBuffer) Flush() {
	if w.status == 0 {
		w.status = http.StatusOK
	}
}
func (w *mcpResponseBuffer) WriteHeader(status int) {
	if w.status == 0 {
		w.status = status
	}
}
func (w *mcpResponseBuffer) Write(p []byte) (int, error) {
	w.observedBytes += min(int64(len(p)), w.limit-w.observedBytes)
	if w.status == 0 {
		w.status = http.StatusOK
	}
	if w.overflow || int64(w.body.Len())+int64(len(p)) > w.limit {
		w.overflow = true
		return 0, errors.New("MCP output limit exceeded")
	}
	return w.body.Write(p)
}
