package httpapi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestMCPOwnerCoverageQueryAndBodyDispatch(t *testing.T) {
	cases := []struct {
		name, method, path, query string
		args                      map[string]any
	}{
		{"transactions.list", "GET", "/api/v1/transactions", "jar_id=unassigned", map[string]any{"jar_id": "unassigned"}},
		{"reports.monthly", "GET", "/api/v1/reports/monthly", "granularity=day&month=2026-11&wallet_id=wallet", map[string]any{"month": "2026-11", "wallet_id": "wallet", "granularity": "day"}},
		{"reports.monthly.transactions", "GET", "/api/v1/reports/monthly/transactions", "limit=10&offset=2&snapshot_id=snapshot&target_id=target", map[string]any{"snapshot_id": "snapshot", "target_id": "target", "limit": 10, "offset": 2}},
		{"reports.analysis.latest", "GET", "/api/v1/reports/analyses/latest", "snapshot_id=snapshot", map[string]any{"snapshot_id": "snapshot"}},
		{"reports.analysis.create", "POST", "/api/v1/reports/analyses", "", map[string]any{"snapshot_id": "snapshot", "client_request_id": "request"}},
		{"advisor.action.replace", "PATCH", "/api/v1/ai/advisor/actions/action/replace", "", map[string]any{"action_id": "action", "conversation_id": "conversation", "generation": 2, "revision": 3, "patch": map[string]any{"name": "Updated", "amount": 123, "assignments": []any{map[string]any{"transaction_id": "transaction", "jar_id": nil}}}}},
		{"advisor.entry.decide", "POST", "/api/v1/ai/advisor/entries/task/decisions", "", map[string]any{"id": "task", "task_id": "task", "batch_id": "batch", "batch_version": 2, "item_id": "item", "item_version": 1, "decision": "approve", "client_request_id": "request"}},
		{"portfolio.asset.update_unit", "PATCH", "/api/v1/portfolio/assets/asset/unit", "", map[string]any{"id": "asset", "expected_version": 2, "unit_label": "chỉ"}},
		{"portfolio.asset.update_price_source", "PATCH", "/api/v1/portfolio/assets/asset/price-source", "", map[string]any{"id": "asset", "expected_version": 2, "price_mode": "manual", "product_id": nil}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			session := mcpTransportSession(t, NewMCPHandler(DefaultExternalOperationCatalog(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				if r.Method != tc.method || r.URL.Path != tc.path || r.URL.RawQuery != tc.query {
					t.Errorf("request = %s %s?%s", r.Method, r.URL.Path, r.URL.RawQuery)
				}
				if tc.method != "GET" {
					var body map[string]any
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Fatal(err)
					}
					if _, exists := body["id"]; exists {
						t.Error("path id leaked into JSON body")
					}
					if tc.name == "advisor.action.replace" && body["patch"].(map[string]any)["amount"] != float64(123) {
						t.Error("patch JSON changed")
					}
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, `{"data":{}}`)
			}), MCPHandlerOptions{}), "Bearer owner")
			result, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: tc.name, Arguments: tc.args})
			if err != nil || result.IsError || !called {
				t.Fatalf("dispatch failed: result=%+v error=%v called=%v", result, err, called)
			}
		})
	}
}

func TestMCPAdvisorEntryMultipartPreservesCanonicalFields(t *testing.T) {
	called := false
	session := mcpTransportSession(t, NewMCPHandler(DefaultExternalOperationCatalog(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.URL.Path != "/api/v1/ai/advisor/entries" {
			t.Errorf("path=%s", r.URL.Path)
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatal(err)
		}
		defer r.MultipartForm.RemoveAll()
		for key, want := range map[string]string{"client_request_id": "request", "conversation_id": "conversation", "generation": "2", "text": "receipt", "timezone": "UTC"} {
			if r.FormValue(key) != want {
				t.Errorf("%s=%q want %q", key, r.FormValue(key), want)
			}
		}
		if len(r.MultipartForm.File["files"]) != 1 {
			t.Fatal("missing file")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, _ = io.WriteString(w, `{"data":{"id":"task","status":"queued"}}`)
	}), MCPHandlerOptions{}), "Bearer owner")
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "advisor.entry.create", Arguments: map[string]any{"client_request_id": "request", "conversation_id": "conversation", "generation": 2, "text": "receipt", "timezone": "UTC", "files": []any{map[string]any{"filename": "receipt.pdf", "mime_type": "application/pdf", "base64": base64.StdEncoding.EncodeToString([]byte("%PDF-1.4\nfixture"))}}}})
	if err != nil || result.IsError || !called {
		t.Fatalf("multipart dispatch failed: result=%+v error=%v called=%v", result, err, called)
	}
}
