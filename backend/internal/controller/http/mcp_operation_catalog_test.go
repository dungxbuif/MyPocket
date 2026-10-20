package httpapi

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// Deliberately independent from the production descriptor builder. New routes
// must be classified here, not silently inherited from a prefix or Swagger.
const mcpAuthenticatedRouteInventory = `
GET /api/v1/ai/advisor/capabilities
GET /api/v1/ai/advisor/overview
GET /api/v1/ai/advisor/conversation/messages
POST /api/v1/ai/advisor/conversation/reset
POST /api/v1/ai/advisor/messages
POST /api/v1/ai/advisor/messages/async
POST /api/v1/ai/advisor/actions/{action_id}/confirm
GET /api/v1/ai/advisor/runs/{id}
GET /api/v1/ai/advisor/runs/{id}/events
POST /api/v1/ai/advisor/runs/{id}/cancel
GET /api/v1/ai/entry/capabilities
POST /api/v1/ai/entry/process
GET /api/v1/ai/entry/requests/{request_id}
PATCH /api/v1/ai/entry/proposals/{id}
POST /api/v1/ai/entry/proposals/{id}/approve
POST /api/v1/ai/entry/proposals/{id}/reject
GET /api/v1/api-keys
POST /api/v1/api-keys
DELETE /api/v1/api-keys/{id}
POST /api/v1/auth/devices
DELETE /api/v1/auth/devices/{id}
POST /api/v1/auth/logout
GET /api/v1/auth/profile
PATCH /api/v1/auth/profile
GET /api/v1/home
GET /api/v1/budget-series
POST /api/v1/budget-series
PATCH /api/v1/budget-series/{id}
POST /api/v1/budget-series/{id}/stop
GET /api/v1/budgets
POST /api/v1/budgets
PATCH /api/v1/budgets/{id}
DELETE /api/v1/budgets/{id}
GET /api/v1/budgets/{id}/report
GET /api/v1/budgets/{id}/transactions
GET /api/v1/categories
POST /api/v1/categories
PATCH /api/v1/categories/{id}
PATCH /api/v1/categories/{id}/status
PATCH /api/v1/categories/{id}/wallets
DELETE /api/v1/categories/{id}
POST /api/v1/credit/wallets/{wallet_id}/entries
POST /api/v1/credit/wallets/{wallet_id}/payments
GET /api/v1/credit/wallets/{wallet_id}/statement
GET /api/v1/feedback
POST /api/v1/feedback
GET /api/v1/feedback/{id}
GET /api/v1/feedback/{id}/screenshot
PATCH /api/v1/feedback/{id}/resolution
GET /api/v1/jars
POST /api/v1/jars
PUT /api/v1/jars/{jar_id}/months/{month}
DELETE /api/v1/jars/{jar_id}/months/{month}
GET /api/v1/jars/{jar_id}/report
GET /api/v1/months/{month}
PUT /api/v1/months/{month}/note
DELETE /api/v1/months/{month}/note
GET /api/v1/notifications
POST /api/v1/notifications/{id}/read
POST /api/v1/notifications/devices
DELETE /api/v1/notifications/devices/{id}
GET /api/v1/portfolio/summary
GET /api/v1/portfolio/assets
POST /api/v1/portfolio/assets
PATCH /api/v1/portfolio/assets/{id}
DELETE /api/v1/portfolio/assets/{id}
PATCH /api/v1/portfolio/assets/{id}/price
GET /api/v1/portfolio/assets/{id}/trades
POST /api/v1/portfolio/assets/{id}/trades
POST /api/v1/portfolio/assets/{id}/trade-corrections/preview
POST /api/v1/portfolio/assets/{id}/trade-corrections
GET /api/v1/recurring
POST /api/v1/recurring
PATCH /api/v1/recurring/{id}
DELETE /api/v1/recurring/{id}
POST /api/v1/recurring/run-due
GET /api/v1/reports/insider
GET /api/v1/reports/savings
GET /api/v1/reports/savings/transactions
GET /api/v1/sync/capabilities
GET /api/v1/transactions
POST /api/v1/transactions
POST /api/v1/transactions/adjustment
POST /api/v1/transactions/bulk-delete
POST /api/v1/transactions/transfer
PATCH /api/v1/transactions/transfer/{transfer_id}
DELETE /api/v1/transactions/transfer/{transfer_id}
PATCH /api/v1/transactions/{id}
DELETE /api/v1/transactions/{id}
PATCH /api/v1/transactions/{id}/travel
GET /api/v1/transactions/{id}/attachments/{attachmentId}/download
GET /api/v1/travel/events
POST /api/v1/travel/events
PATCH /api/v1/travel/events/{id}
DELETE /api/v1/travel/events/{id}
POST /api/v1/travel/events/{id}/activate
POST /api/v1/travel/events/{id}/deactivate
GET /api/v1/wallets
POST /api/v1/wallets
PATCH /api/v1/wallets/{id}
DELETE /api/v1/wallets/{id}
`

// Each exclusion is a full method/path, with a reason checked below.
var mcpExcludedRouteInventory = map[string]string{
	"POST /api/v1/login":                          "credential establishment",
	"POST /api/v1/auth/refresh":                   "credential exchange",
	"POST /api/v1/auth/device-login":              "credential establishment",
	"GET /api/v1/auth/google":                     "credential establishment",
	"GET /api/v1/auth/google/callback":            "credential exchange",
	"POST /api/v1/auth/mobile/google/exchange":    "credential exchange",
	"GET /api/v1/health":                          "public health",
	"GET /api/v1/changelog":                       "public changelog",
	"GET /api/v1/changelog/{id}":                  "public changelog",
	"GET /api/v1/agent/feedback":                  "feedback-agent service token",
	"GET /api/v1/agent/feedback/{id}/screenshot":  "feedback-agent service token",
	"PATCH /api/v1/internal/feedback/{id}/status": "feedback-agent service token",
	"POST /api/v1/internal/changelog":             "feedback-agent service token",
	"GET /api/v1/mcp":                             "SDK protocol entry point; never a recursive tool",
	"POST /api/v1/mcp":                            "SDK protocol entry point; never a recursive tool",
	"DELETE /api/v1/mcp":                          "SDK protocol entry point; never a recursive tool",
	"GET /api/v1/docs/*any":                       "public OpenAPI documentation",
	"GET /docs":                                   "public documentation redirect",
	"GET /docs/":                                  "public documentation redirect",
}

func mcpInventorySet() map[string]bool {
	result := map[string]bool{}
	for _, line := range strings.Split(strings.TrimSpace(mcpAuthenticatedRouteInventory), "\n") {
		result[line] = true
	}
	return result
}

func mcpParityRouter() *Router {
	gin.SetMode(gin.TestMode)
	return NewApplicationRouter(ApplicationHandlers{}, ApplicationRouterOptions{})
}

func mcpOpenAPI(t *testing.T) map[string]map[string]struct {
	Security []map[string][]string `json:"security"`
} {
	t.Helper()
	data, err := os.ReadFile("../../../docs/swagger.json")
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Paths map[string]map[string]struct {
			Security []map[string][]string `json:"security"`
		} `json:"paths"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	return doc.Paths
}

func TestMCPAuthenticatedRouteParity(t *testing.T) {
	catalog := DefaultExternalOperationCatalog()
	want := mcpInventorySet()
	actual := map[string]bool{}
	for _, operation := range catalog.List() {
		actual[operation.Method+" "+operation.PathTemplate] = true
	}
	if !reflect.DeepEqual(actual, want) {
		t.Fatalf("catalog inventory mismatch:\ngot %v\nwant %v", actual, want)
	}
	doc := mcpOpenAPI(t)
	registered := map[string]bool{}
	for _, route := range mcpParityRouter().Engine.Routes() {
		parts := strings.Split(route.Path, "/")
		for i, part := range parts {
			if strings.HasPrefix(part, ":") {
				parts[i] = "{" + part[1:] + "}"
			}
		}
		key := route.Method + " " + strings.Join(parts, "/")
		if !want[key] && mcpExcludedRouteInventory[key] == "" {
			t.Errorf("unclassified registered route: %s", key)
		}
		registered[key] = true
	}
	for key := range want {
		if !registered[key] {
			t.Errorf("missing registered route: %s", key)
		}
		parts := strings.SplitN(key, " ", 2)
		if len(doc[parts[1]][strings.ToLower(parts[0])].Security) == 0 {
			t.Errorf("missing OpenAPI security: %s", key)
		}
	}
	for path, methods := range doc {
		for method, operation := range methods {
			if len(operation.Security) == 0 {
				continue
			}
			key := strings.ToUpper(method) + " " + path
			if !want[key] && mcpExcludedRouteInventory[key] == "" {
				t.Errorf("unclassified authenticated OpenAPI operation: %s", key)
			}
		}
	}
}

func TestMCPRouteExclusions(t *testing.T) {
	catalog := DefaultExternalOperationCatalog()
	for _, operation := range catalog.List() {
		if reason := mcpExcludedRouteInventory[operation.Method+" "+operation.PathTemplate]; reason != "" {
			t.Errorf("exposed excluded operation: %s (%s)", operation.Name, reason)
		}
	}
	// Execute the real middleware composition: an unauthenticated caller
	// cannot reach even a nil-backed business handler on any catalog route.
	r := mcpParityRouter()
	for key := range mcpInventorySet() {
		parts := strings.SplitN(key, " ", 2)
		path := parts[1]
		for strings.Contains(path, "{") {
			start := strings.Index(path, "{")
			end := strings.Index(path, "}")
			path = path[:start] + "fixture" + path[end+1:]
		}
		recorder := httptest.NewRecorder()
		r.Engine.ServeHTTP(recorder, httptest.NewRequest(parts[0], path, nil))
		if recorder.Code != http.StatusUnauthorized {
			t.Errorf("%s auth status = %d, want 401", key, recorder.Code)
		}
	}
}

func TestMCPCatalogStableNamesAndOrdinaryWrites(t *testing.T) {
	catalog := DefaultExternalOperationCatalog()
	want := map[string]string{
		"account.profile.get": "GET /api/v1/auth/profile", "account.profile.update_timezone": "PATCH /api/v1/auth/profile", "account.api_key.create": "POST /api/v1/api-keys", "account.logout": "POST /api/v1/auth/logout",
		"wallets.create": "POST /api/v1/wallets", "categories.update": "PATCH /api/v1/categories/{id}", "transactions.create": "POST /api/v1/transactions", "transactions.transfer.delete": "DELETE /api/v1/transactions/transfer/{transfer_id}",
		"budgets.create": "POST /api/v1/budgets", "budget_series.stop": "POST /api/v1/budget-series/{id}/stop", "jars.month.update": "PUT /api/v1/jars/{jar_id}/months/{month}", "months.note.save": "PUT /api/v1/months/{month}/note",
		"recurring.run_due": "POST /api/v1/recurring/run-due", "credit.payment.create": "POST /api/v1/credit/wallets/{wallet_id}/payments", "travel.activate": "POST /api/v1/travel/events/{id}/activate", "portfolio.trade.correct": "POST /api/v1/portfolio/assets/{id}/trade-corrections",
		"reports.savings.transactions": "GET /api/v1/reports/savings/transactions", "notifications.device.register": "POST /api/v1/notifications/devices", "feedback.create": "POST /api/v1/feedback", "ai_entry.process": "POST /api/v1/ai/entry/process", "advisor.messages_async": "POST /api/v1/ai/advisor/messages/async", "sync.capabilities": "GET /api/v1/sync/capabilities",
	}
	for name, route := range want {
		op, ok := catalog.Lookup(name)
		if !ok || op.Method+" "+op.PathTemplate != route {
			t.Errorf("%s missing or wrong route: %+v", name, op)
		}
	}
	list := catalog.List()
	names := make([]string, len(list))
	for i, op := range list {
		names[i] = op.Name
	}
	if !sort.StringsAreSorted(names) {
		t.Fatal("catalog listing must have stable name order")
	}
	if op, _ := catalog.Lookup("ai_entry.process"); op.BodyEncoding != "multipart" || !op.Async {
		t.Fatal("AI entry must describe multipart/async behavior")
	}
	if op, _ := catalog.Lookup("advisor.messages_async"); !op.Async {
		t.Fatal("async marker missing")
	}
	if op, _ := catalog.Lookup("transactions.attachment.download"); !op.SensitiveResult {
		t.Fatal("signed attachment result must be sensitive")
	}
	if op, _ := catalog.Lookup("transactions.transfer.update"); op.Idempotency != "required" {
		t.Fatal("transfer update requires canonical receipt key")
	}
}

func TestMCPCatalogRejectsInvalidDescriptors(t *testing.T) {
	catalog := DefaultExternalOperationCatalog()
	base, _ := catalog.Lookup("wallets.update")
	for _, tc := range []struct {
		name   string
		mutate func(*ExternalOperation)
	}{
		{"invalid name", func(op *ExternalOperation) { op.Name = "call_rest" }},
		{"invalid method", func(op *ExternalOperation) { op.Method = "CONNECT" }},
		{"external URL", func(op *ExternalOperation) { op.PathTemplate = "https://example.com/api/v1/wallets/{id}" }},
		{"internal route", func(op *ExternalOperation) { op.PathTemplate = "/api/v1/internal/wallets/{id}" }},
		{"recursive transport", func(op *ExternalOperation) { op.PathTemplate = "/api/v1/mcp" }},
		{"path traversal", func(op *ExternalOperation) { op.PathTemplate = "/api/v1/../wallets/{id}" }},
		{"missing schema", func(op *ExternalOperation) { op.InputSchema = nil }},
		{"missing path field", func(op *ExternalOperation) { delete(op.InputSchema["properties"].(map[string]any), "id") }},
		{"missing encoding", func(op *ExternalOperation) { op.BodyEncoding = "" }},
		{"owner schema", func(op *ExternalOperation) {
			op.InputSchema["properties"].(map[string]any)["owner_id"] = map[string]any{"type": "string"}
		}},
		{"invalid idempotency", func(op *ExternalOperation) { op.Idempotency = "automatic" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			copy, _ := catalog.Lookup(base.Name)
			tc.mutate(&copy)
			if _, err := NewExternalOperationCatalog([]ExternalOperation{copy}); err == nil {
				t.Fatal("accepted invalid descriptor")
			}
		})
	}
	if _, err := NewExternalOperationCatalog([]ExternalOperation{base, base}); err == nil {
		t.Fatal("accepted duplicate name")
	}
	duplicate := base
	duplicate.Name = "wallets.other"
	if _, err := NewExternalOperationCatalog([]ExternalOperation{base, duplicate}); err == nil {
		t.Fatal("accepted duplicate method/path")
	}
}

func TestMCPCatalogEncodeCanonicalRequest(t *testing.T) {
	catalog := DefaultExternalOperationCatalog()
	op, _ := catalog.Lookup("transactions.update")
	request, err := op.Encode(map[string]any{"id": "owned-id", "amount": float64(1250), "expected_version": float64(4), "note": "tea"}, map[string]any{"idempotency_key": "client-stable-key"})
	if err != nil {
		t.Fatal(err)
	}
	if request.Method != "PATCH" || request.URL.RequestURI() != "/api/v1/transactions/owned-id" || request.Header.Get("Idempotency-Key") != "client-stable-key" || request.Header.Get("Content-Type") != "application/json" {
		t.Fatalf("wrong canonical request: %+v", request)
	}
	var body map[string]any
	if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(body, map[string]any{"amount": float64(1250), "expected_version": float64(4), "note": "tea"}) {
		t.Fatalf("path/meta leaked into body: %v", body)
	}
	op, _ = catalog.Lookup("reports.savings.transactions")
	request, err = op.Encode(map[string]any{"month": "2026-10", "section": "surplus", "limit": float64(25), "offset": float64(0)}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if request.URL.RequestURI() != "/api/v1/reports/savings/transactions?limit=25&month=2026-10&offset=0&section=surplus" {
		t.Fatalf("query encoding: %s", request.URL.RequestURI())
	}
	if request.Body != nil {
		data, _ := io.ReadAll(request.Body)
		if len(data) > 0 {
			t.Fatalf("query operation sent a body: %s", data)
		}
	}
}

func TestMCPCatalogEncodeRejectsOwnerSelectorsAndMalformedArguments(t *testing.T) {
	catalog := DefaultExternalOperationCatalog()
	for _, tc := range []struct {
		name, operation string
		arguments, meta map[string]any
	}{
		{"owner", "wallets.create", map[string]any{"owner_id": "foreign"}, nil},
		{"user", "reports.savings", map[string]any{"user_id": "foreign"}, nil},
		{"nested owner", "feedback.create", map[string]any{"metadata": map[string]any{"owner_id": "foreign"}}, nil},
		{"nested array owner", "transactions.bulk_delete", map[string]any{"transaction_ids": []any{map[string]any{"user_id": "foreign"}}}, nil},
		{"metadata owner", "wallets.create", map[string]any{}, map[string]any{"owner_id": "foreign"}},
		{"unknown field", "wallets.create", map[string]any{"arbitrary": true}, nil},
		{"wrong type", "wallets.create", map[string]any{"name": 12}, nil},
		{"missing path", "wallets.update", map[string]any{"name": "new"}, nil},
		{"traversal", "wallets.delete", map[string]any{"id": "../foreign"}, nil},
		{"encoded slash", "wallets.delete", map[string]any{"id": "a%2fb"}, nil},
		{"query injection", "wallets.delete", map[string]any{"id": "a?owner_id=foreign"}, nil},
		{"body key", "wallets.create", map[string]any{"idempotency_key": "wrong"}, map[string]any{"idempotency_key": "right"}},
		{"invalid metadata", "wallets.create", nil, map[string]any{"headers": map[string]any{"Authorization": "other"}}},
		{"invalid key", "wallets.create", nil, map[string]any{"idempotency_key": "line\nbreak"}},
		{"fractional integer", "transactions.create", map[string]any{"amount": 1.5}, nil},
		{"confirmation payload", "advisor.action.confirm", map[string]any{"action_id": "action", "conversation_id": "conv", "generation": float64(1), "amount": 100}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			op, ok := catalog.Lookup(tc.operation)
			if !ok {
				t.Fatal("missing operation")
			}
			if request, err := op.Encode(tc.arguments, tc.meta); err == nil {
				t.Fatalf("accepted unsafe arguments: %+v", request)
			}
		})
	}
}

func TestMCPCatalogCopiesCannotMutateDispatchAuthority(t *testing.T) {
	catalog := DefaultExternalOperationCatalog()
	op, _ := catalog.Lookup("wallets.update")
	op.PathTemplate = "/api/v1/internal/feedback"
	op.InputSchema["properties"].(map[string]any)["owner_id"] = map[string]any{"type": "string"}
	list := catalog.List()
	for i := range list {
		if list[i].Name == "wallets.update" {
			list[i].InputSchema["properties"].(map[string]any)["user_id"] = map[string]any{"type": "string"}
		}
	}
	op, _ = catalog.Lookup("wallets.update")
	if op.PathTemplate != "/api/v1/wallets/{id}" {
		t.Fatal("caller changed catalog authority")
	}
	properties := op.InputSchema["properties"].(map[string]any)
	if properties["owner_id"] != nil || properties["user_id"] != nil {
		t.Fatal("caller changed catalog schema")
	}
}

func TestMCPCatalogRejectsUnusableSchemas(t *testing.T) {
	catalog := DefaultExternalOperationCatalog()
	for _, tc := range []struct {
		name   string
		schema map[string]any
	}{
		{"unsupported type", map[string]any{"type": "future-type"}},
		{"unresolved reference", map[string]any{"$ref": "#/definitions/Secret"}},
		{"array without item schema", map[string]any{"type": "array"}},
		{"open nested object", map[string]any{"type": "object", "properties": map[string]any{}, "additionalProperties": true}},
		{"required unknown field", map[string]any{"type": "object", "properties": map[string]any{}, "additionalProperties": false, "required": []string{"unknown"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			op, _ := catalog.Lookup("wallets.create")
			op.InputSchema["properties"].(map[string]any)["name"] = tc.schema
			if _, err := NewExternalOperationCatalog([]ExternalOperation{op}); err == nil {
				t.Fatal("accepted unusable schema")
			}
		})
	}
}

func TestMCPCatalogPreservesCanonicalNullableListsAndASCIIKeys(t *testing.T) {
	catalog := DefaultExternalOperationCatalog()
	op, _ := catalog.Lookup("categories.set_wallets")
	request, err := op.Encode(map[string]any{"id": "owned", "wallet_ids": nil}, map[string]any{"idempotency_key": "client key with spaces"})
	if err != nil {
		t.Fatal(err)
	}
	data, _ := io.ReadAll(request.Body)
	if string(data) != `{"wallet_ids":null}` || request.Header.Get("Idempotency-Key") != "client key with spaces" {
		t.Fatalf("canonical nullable list or key changed: %s, %q", data, request.Header.Get("Idempotency-Key"))
	}
	op, _ = catalog.Lookup("transactions.create")
	request, err = op.Encode(map[string]any{"amount": json.Number("9007199254740993")}, nil)
	if err != nil {
		t.Fatal(err)
	}
	data, _ = io.ReadAll(request.Body)
	if string(data) != `{"amount":9007199254740993}` {
		t.Fatalf("int64 precision changed: %s", data)
	}
}

func TestMCPCatalogIdempotencyDoesNotPromiseUnsupportedReceipts(t *testing.T) {
	catalog := DefaultExternalOperationCatalog()
	for name, want := range map[string]string{
		"wallets.create": "optional", "categories.create": "optional", "budgets.create": "optional", "jars.create": "optional", "transactions.create": "optional", "transactions.transfer.create": "optional",
		"transactions.transfer.update": "required", "transactions.transfer.delete": "required", "budget_series.create": "required", "budget_series.update": "required", "budget_series.stop": "required", "portfolio.trade.correct": "required", "advisor.reset": "required",
		"wallets.update": "none", "transactions.update": "none", "transactions.delete": "none", "feedback.create": "none", "recurring.create": "none", "credit.payment.create": "none", "travel.create": "none",
	} {
		op, ok := catalog.Lookup(name)
		if !ok || op.Idempotency != want {
			t.Errorf("%s idempotency = %q, want %q", name, op.Idempotency, want)
		}
	}
	// Absence of a receipt key is passed through unchanged. The canonical
	// handler decides the original status/problem and never gets an invented key.
	op, _ := catalog.Lookup("budget_series.create")
	request, err := op.Encode(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if request.Header.Get("Idempotency-Key") != "" {
		t.Fatal("encoder invented a retry key")
	}
}

func TestMCPCatalogMultipartPreservesFileOrderAndFields(t *testing.T) {
	catalog := DefaultExternalOperationCatalog()
	op, _ := catalog.Lookup("ai_entry.process")
	file := func(name, text string) map[string]any {
		return map[string]any{"filename": name, "mime_type": "application/pdf", "base64": base64.StdEncoding.EncodeToString([]byte(text))}
	}
	request, err := op.Encode(map[string]any{"request_id": "client-id", "text": "two receipts", "timezone": "Asia/Ho_Chi_Minh", "files": []any{file("first.pdf", "%PDF-1.7\nfirst"), file("second.pdf", "%PDF-1.7\nsecond")}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	reader, err := request.MultipartReader()
	if err != nil {
		t.Fatal(err)
	}
	var names, contents []string
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		data, _ := io.ReadAll(part)
		if part.FileName() == "" {
			names = append(names, part.FormName())
			contents = append(contents, string(data))
		} else {
			names = append(names, part.FileName())
			contents = append(contents, string(data))
			if part.Header.Get("Content-Type") != "application/pdf" {
				t.Fatal("file MIME was lost")
			}
		}
	}
	if !reflect.DeepEqual(names, []string{"request_id", "text", "timezone", "first.pdf", "second.pdf"}) || !reflect.DeepEqual(contents, []string{"client-id", "two receipts", "Asia/Ho_Chi_Minh", "%PDF-1.7\nfirst", "%PDF-1.7\nsecond"}) {
		t.Fatalf("multipart changed ordering/fields: %v %v", names, contents)
	}
}

func TestMCPCatalogMultipartLimitsAndMIME(t *testing.T) {
	catalog := DefaultExternalOperationCatalog()
	op, _ := catalog.Lookup("ai_entry.process")
	pdf := []byte("%PDF-1.7\nreceipt")
	file := func(data []byte) map[string]any {
		return map[string]any{"filename": "receipt.pdf", "mime_type": "application/pdf", "base64": base64.StdEncoding.EncodeToString(data)}
	}
	t.Run("20 files accepted", func(t *testing.T) {
		files := make([]any, 20)
		for i := range files {
			files[i] = file(pdf)
		}
		if _, err := op.Encode(map[string]any{"files": files}, nil); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("5 MiB accepted", func(t *testing.T) {
		data := append([]byte("%PDF-1.7\n"), bytes.Repeat([]byte("a"), 5*1024*1024-len("%PDF-1.7\n"))...)
		if _, err := op.Encode(map[string]any{"files": []any{file(data)}}, nil); err != nil {
			t.Fatal(err)
		}
	})
	for _, tc := range []struct {
		name  string
		files []any
	}{
		{"21 files", make([]any, 21)},
		{"oversized file", []any{file(bytes.Repeat([]byte("a"), 5*1024*1024+1))}},
		{"unsupported bytes", []any{file([]byte("plain text pretending PDF"))}},
		{"mismatched MIME", []any{map[string]any{"filename": "receipt.png", "mime_type": "image/png", "base64": base64.StdEncoding.EncodeToString(pdf)}}},
		{"bad base64", []any{map[string]any{"filename": "receipt.pdf", "mime_type": "application/pdf", "base64": "%%%"}}},
		{"invalid filename", []any{map[string]any{"filename": "../receipt.pdf", "mime_type": "application/pdf", "base64": base64.StdEncoding.EncodeToString(pdf)}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.name == "21 files" {
				for i := range tc.files {
					tc.files[i] = file(pdf)
				}
			}
			if _, err := op.Encode(map[string]any{"files": tc.files}, nil); err == nil {
				t.Fatal("accepted invalid upload")
			}
		})
	}
	t.Run("total multipart bytes", func(t *testing.T) {
		if _, _, err := encodeExternalMultipart(map[string]any{"text": strings.Repeat("x", 101*1024*1024)}); err == nil {
			t.Fatal("accepted oversized multipart body")
		}
	})
}
