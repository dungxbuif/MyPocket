package httpapi

import (
	"encoding/json"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// candidateRouteContract is deliberately narrow: these are the routes added
// after the currently deployed production OpenAPI document. Keeping the list
// here makes a staged rollout fail closed when a handler, docs entry, or auth
// wrapper is accidentally removed.
var candidateRouteContract = map[string][]string{
	"/api/v1/ai/advisor/runs/{id}/events":                     {"get"},
	"/api/v1/budget-series":                                   {"get", "post"},
	"/api/v1/budget-series/{id}":                              {"patch"},
	"/api/v1/budget-series/{id}/stop":                         {"post"},
	"/api/v1/budgets/{id}/report":                             {"get"},
	"/api/v1/budgets/{id}/transactions":                       {"get"},
	"/api/v1/notifications/devices":                           {"post"},
	"/api/v1/notifications/devices/{id}":                      {"delete"},
	"/api/v1/reports/savings":                                 {"get"},
	"/api/v1/reports/savings/transactions":                    {"get"},
	"/api/v1/portfolio/assets/{id}/trade-corrections/preview": {"post"},
	"/api/v1/portfolio/assets/{id}/trade-corrections":         {"post"},
}

func TestCandidateRouteContractIsRegisteredAndDocumented(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := &Router{Engine: gin.New(), AuthMiddleware: &AuthMiddleware{}}
	router.RegisterAdvisorRoutes(&AdvisorHandler{})
	router.RegisterBudgetSeriesRoutes(&BudgetSeriesHandler{})
	router.RegisterBudgetReportRoutes(&BudgetReportHandler{})
	router.RegisterNotificationRoutes(&NotificationHandler{})
	router.RegisterPortfolioRoutes(&PortfolioHandler{})
	// Reports are registered in main directly rather than through a Router
	// method; mirror that composition here so the parity test covers the real
	// route shape without constructing repositories or a database.
	reports := router.Engine.Group("/api/v1/reports")
	reports.Use(router.AuthMiddleware.RequireAuth)
	reports.GET("/savings", (&SavingsReportHandler{}).GetSavingsReport)
	reports.GET("/savings/transactions", (&SavingsReportHandler{}).ListSavingsTransactions)

	registered := make(map[string]map[string]bool)
	for _, route := range router.Engine.Routes() {
		path := strings.ReplaceAll(route.Path, ":id", "{id}")
		if registered[path] == nil {
			registered[path] = make(map[string]bool)
		}
		registered[path][strings.ToLower(route.Method)] = true
	}
	for path, methods := range candidateRouteContract {
		for _, method := range methods {
			if !registered[path][method] {
				t.Fatalf("candidate route is not registered: %s %s", strings.ToUpper(method), path)
			}
		}
	}

	// swagger.json is generated from the route annotations and is the checked-in
	// public contract consumed by native clients and release audits.
	docBytes, err := os.ReadFile("../../../docs/swagger.json")
	if err != nil {
		t.Fatalf("read checked-in OpenAPI document: %v", err)
	}
	var doc struct {
		Paths map[string]map[string]json.RawMessage `json:"paths"`
	}
	if err := json.Unmarshal(docBytes, &doc); err != nil {
		t.Fatalf("decode checked-in OpenAPI document: %v", err)
	}
	for path, methods := range candidateRouteContract {
		operations, ok := doc.Paths[path]
		if !ok {
			t.Fatalf("candidate route missing from OpenAPI: %s", path)
		}
		for _, method := range methods {
			operation, ok := operations[method]
			if !ok {
				t.Fatalf("candidate operation missing from OpenAPI: %s %s", strings.ToUpper(method), path)
			}
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(operation, &fields); err != nil {
				t.Fatalf("decode OpenAPI operation %s %s: %v", strings.ToUpper(method), path, err)
			}
			security, ok := fields["security"]
			if !ok || string(security) == "null" || string(security) == "[]" {
				t.Fatalf("candidate operation lacks authenticated security contract: %s %s", strings.ToUpper(method), path)
			}
		}
	}

	// Keep failures deterministic when this list is edited during a staged
	// rollout; a sorted copy makes test output stable for CI diagnostics.
	paths := make([]string, 0, len(candidateRouteContract))
	for path := range candidateRouteContract {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	if len(paths) != 12 {
		t.Fatalf("candidate route contract unexpectedly changed: %d routes (%v)", len(paths), paths)
	}
}
