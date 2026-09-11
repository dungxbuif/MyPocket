package httpapi

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"strings"
)

//go:embed openapi.json
var openAPIBase []byte

type RouteDescriptor struct {
	Method, Path, Summary       string
	Owned, Mutating, Idempotent bool
}

func route(method, path, summary string, owned, mutating, idempotent bool) RouteDescriptor {
	return RouteDescriptor{method, path, summary, owned, mutating, idempotent}
}

var versionedRoutes = []RouteDescriptor{
	route("GET", "/api/v1/auth/google", "Start Google authentication", false, false, false), route("GET", "/api/v1/auth/google/callback", "Complete Google authentication", false, false, false), route("POST", "/api/v1/auth/logout", "Log out", true, true, false), route("GET", "/api/v1/me", "Get current user", true, false, false),
	route("GET", "/api/v1/api-keys", "List API keys", true, false, false), route("POST", "/api/v1/api-keys", "Create API key", true, true, false), route("POST", "/api/v1/api-keys/{id}/revoke", "Revoke API key", true, true, false),
	route("GET", "/api/v1/wallets", "List wallets", true, false, false), route("POST", "/api/v1/wallets", "Create wallet", true, true, true), route("PATCH", "/api/v1/wallets/{id}", "Update wallet", true, true, true), route("POST", "/api/v1/wallets/{id}/archive", "Archive wallet", true, true, true), route("POST", "/api/v1/wallets/{id}/default-ai", "Set default agent wallet", true, true, true), route("GET", "/api/v1/wallets/{id}/detail", "Get wallet detail", true, false, false), route("GET", "/api/v1/wallets/{id}/category-settings", "List wallet category settings", true, false, false), route("PUT", "/api/v1/wallets/{id}/categories/{categoryId}", "Set wallet category state", true, true, false),
	route("GET", "/api/v1/categories", "List categories", true, false, false), route("POST", "/api/v1/categories", "Create category", true, true, true), route("PATCH", "/api/v1/categories/{id}", "Update category", true, true, true), route("POST", "/api/v1/categories/{id}/archive", "Archive category", true, true, true),
	route("GET", "/api/v1/transactions", "List transactions", true, false, false), route("POST", "/api/v1/transactions", "Create transaction", true, true, true), route("PATCH", "/api/v1/transactions/{id}", "Update transaction", true, true, true), route("POST", "/api/v1/transactions/{id}/archive", "Archive transaction", true, true, true),
	route("POST", "/api/v1/files/presign", "Create receipt upload", true, true, false), route("GET", "/api/v1/files/{id}/download", "Download receipt", true, false, false),
	route("GET", "/api/v1/budgets", "List budgets", true, false, false), route("POST", "/api/v1/budgets", "Create budget", true, true, false), route("PATCH", "/api/v1/budgets/{id}", "Update budget", true, true, false), route("POST", "/api/v1/budgets/{id}/archive", "Archive budget", true, true, false),
	route("GET", "/api/v1/events", "List events", true, false, false), route("POST", "/api/v1/events", "Create event", true, true, false), route("PATCH", "/api/v1/events/{id}", "Update event", true, true, false), route("POST", "/api/v1/events/{id}/archive", "Archive event", true, true, false), route("POST", "/api/v1/events/{id}/transactions/{transactionId}", "Link event transaction", true, true, false),
	route("GET", "/api/v1/obligations", "List obligations", true, false, false), route("POST", "/api/v1/obligations", "Create obligation", true, true, false), route("PATCH", "/api/v1/obligations/{id}", "Update obligation", true, true, false), route("POST", "/api/v1/obligations/{id}/archive", "Archive obligation", true, true, false), route("POST", "/api/v1/obligations/{id}/repayments/{transactionId}", "Link repayment", true, true, false),
	route("GET", "/api/v1/recurring-schedules", "List recurring schedules", true, false, false), route("POST", "/api/v1/recurring-schedules", "Create recurring schedule", true, true, false), route("POST", "/api/v1/recurring-schedules/{id}/archive", "Archive recurring schedule", true, true, false), route("GET", "/api/v1/transaction-drafts", "List transaction drafts", true, false, false), route("POST", "/api/v1/transaction-drafts/{id}/confirm", "Confirm draft", true, true, true), route("POST", "/api/v1/transaction-drafts/{id}/reject", "Reject draft", true, true, false),
	route("GET", "/api/v1/notifications", "List notifications", true, false, false), route("PATCH", "/api/v1/notifications/{id}/read", "Mark notification read", true, true, false), route("POST", "/api/v1/push-subscriptions", "Create push subscription", true, true, false), route("DELETE", "/api/v1/push-subscriptions/{id}", "Delete push subscription", true, true, false),
	route("GET", "/api/v1/dashboard", "Get dashboard", true, false, false), route("GET", "/api/v1/assets", "List assets", true, false, false), route("POST", "/api/v1/assets", "Create asset", true, true, false), route("GET", "/api/v1/assets/{id}", "Get asset", true, false, false), route("POST", "/api/v1/assets/{id}/archive", "Archive asset", true, true, false), route("POST", "/api/v1/assets/{id}/trades", "Add trade", true, true, false), route("PATCH", "/api/v1/assets/{id}/trades/{tradeId}", "Update trade", true, true, false), route("POST", "/api/v1/assets/{id}/archive-trade/{tradeId}", "Archive trade", true, true, false), route("POST", "/api/v1/assets/{id}/prices", "Add asset price", true, true, false), route("GET", "/api/v1/portfolio/summary", "Get portfolio summary", true, false, false),
	route("GET", "/api/v1/reports/cash-flow", "Get cash-flow report", true, false, false), route("GET", "/api/v1/reports/categories", "Get category report", true, false, false), route("GET", "/api/v1/reports/daily", "Get daily report", true, false, false), route("GET", "/api/v1/reports/cumulative", "Get cumulative report", true, false, false), route("GET", "/api/v1/reports/comparison", "Get six-period comparison", true, false, false), route("GET", "/api/v1/reports/insider", "Get spending insight", true, false, false), route("GET", "/api/v1/search", "Search user data", true, false, false),
	route("POST", "/api/v1/sync/mutations", "Apply offline mutations", true, true, true), route("GET", "/api/v1/sync/changes", "Read change feed", true, false, false), route("POST", "/api/v1/sync/resync", "Get authoritative snapshot", true, true, false), route("GET", "/api/v1/audit/events", "List authorized audit events", true, false, false), route("GET", "/api/v1/audit/access", "Check audit access", true, false, false),
	route("POST", "/api/v1/imports", "Create import", true, true, true), route("GET", "/api/v1/imports/{id}", "Get import", true, false, false), route("POST", "/api/v1/imports/{id}/confirm", "Confirm import", true, true, false), route("POST", "/api/v1/exports", "Create export", true, true, true), route("GET", "/api/v1/exports/{id}", "Get export", true, false, false), route("GET", "/api/v1/exports/{id}/download", "Get export download URL", true, false, false), route("POST", "/api/v1/account/reset", "Preview or confirm reset", true, true, true), route("POST", "/api/v1/account/delete", "Preview or confirm delete", true, true, true), route("GET", "/api/v1/account/jobs/{id}", "Get lifecycle job", true, false, false),
	route("GET", "/api/v1/health/live", "Liveness", false, false, false), route("GET", "/api/v1/health/ready", "Readiness", false, false, false), route("GET", "/api/v1/openapi.json", "OpenAPI contract", false, false, false),
}

func RegisteredRoutesForTest() []RouteDescriptor {
	return append([]RouteDescriptor(nil), versionedRoutes...)
}

func openAPIDocument() []byte {
	var document map[string]any
	if json.Unmarshal(openAPIBase, &document) != nil {
		return openAPIBase
	}
	paths := document["paths"].(map[string]any)
	for _, descriptor := range versionedRoutes {
		path := strings.TrimPrefix(descriptor.Path, "/api/v1")
		if path == "" {
			path = "/"
		}
		entry, ok := paths[path].(map[string]any)
		if !ok {
			entry = map[string]any{}
			paths[path] = entry
		}
		operation := map[string]any{
			"operationId": operationID(descriptor), "summary": descriptor.Summary,
			"responses": map[string]any{
				"200": map[string]any{"description": "Success", "headers": map[string]any{"X-Correlation-ID": map[string]any{"schema": map[string]any{"type": "string"}}}, "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"type": "object"}}}},
				"400": map[string]any{"description": "Validation error", "content": errorContent()}, "401": map[string]any{"description": "Authentication required", "content": errorContent()},
				"404": map[string]any{"description": "Resource not found", "content": errorContent()}, "409": map[string]any{"description": "Version conflict", "content": errorContent()},
				"429": map[string]any{"description": "Rate limited", "headers": map[string]any{"Retry-After": map[string]any{"schema": map[string]any{"type": "integer"}}}, "content": errorContent()},
				"503": map[string]any{"description": "Retryable failure", "content": errorContent()},
			},
		}
		if descriptor.Owned {
			operation["security"] = []any{map[string]any{"cookieAuth": []any{}, "csrfHeader": []any{}}, map[string]any{"bearerAuth": []any{}}}
			if !descriptor.Mutating {
				operation["security"] = []any{map[string]any{"cookieAuth": []any{}}, map[string]any{"bearerAuth": []any{}}}
			}
			if descriptor.Path == "/api/v1/account/reset" || descriptor.Path == "/api/v1/account/delete" {
				operation["security"] = []any{map[string]any{"cookieAuth": []any{}, "csrfHeader": []any{}}}
			}
		}
		parameters := []any{}
		for _, segment := range strings.Split(path, "/") {
			if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") {
				parameters = append(parameters, map[string]any{"name": strings.Trim(segment, "{}"), "in": "path", "required": true, "schema": map[string]any{"type": "string", "format": "uuid"}})
			}
		}
		if descriptor.Idempotent {
			parameters = append(parameters, map[string]any{"name": "Idempotency-Key", "in": "header", "required": true, "schema": map[string]any{"type": "string"}})
		}
		if len(parameters) > 0 {
			operation["parameters"] = parameters
		}
		if descriptor.Mutating {
			operation["requestBody"] = map[string]any{"required": true, "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"type": "object"}}}}
		}
		entry[strings.ToLower(descriptor.Method)] = operation
	}
	body, err := json.Marshal(document)
	if err != nil {
		return openAPIBase
	}
	return body
}
func errorContent() map[string]any {
	return map[string]any{"application/json": map[string]any{"schema": map[string]any{"$ref": "#/components/schemas/ErrorEnvelope"}}}
}
func operationID(route RouteDescriptor) string {
	return strings.ToLower(route.Method) + strings.NewReplacer("/", "_", "{", "", "}", "", "-", "_").Replace(strings.Trim(route.Path, "/"))
}
func openAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
		return
	}
	w.Header().Set("Content-Type", "application/vnd.oai.openapi+json")
	_, _ = w.Write(openAPIDocument())
}
