package httpapi

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/mypocket/backend/internal/usecase"
)

func TestApplicationRouterPreservesPublicDocumentation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewApplicationRouter(ApplicationHandlers{}, ApplicationRouterOptions{DocsHandler: func(c *gin.Context) { c.String(http.StatusOK, "application docs") }})
	for _, path := range []string{"/docs", "/docs/"} {
		response := httptest.NewRecorder()
		router.Engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusFound || response.Header().Get("Location") != "/api/v1/docs/index.html" {
			t.Fatalf("documentation redirect %s = %d %q", path, response.Code, response.Header().Get("Location"))
		}
	}
	response := httptest.NewRecorder()
	router.Engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/docs/index.html", nil))
	if response.Code != http.StatusOK || response.Body.String() != "application docs" {
		t.Fatalf("configured documentation handler was lost: %d %s", response.Code, response.Body.String())
	}
}

func TestApplicationRouterMCPUsesExistingOwnerAuthentication(t *testing.T) {
	router := mcpParityRouter()
	for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodDelete} {
		response := httptest.NewRecorder()
		router.Engine.ServeHTTP(response, httptest.NewRequest(method, "/api/v1/mcp", nil))
		if response.Code != http.StatusUnauthorized || response.Header().Get(requestIDHeader) == "" {
			t.Fatalf("MCP %s must authenticate with a request ID: %d %s", method, response.Code, response.Body.String())
		}
	}
}

func TestApplicationRouterMCPInheritsAuditPrincipalAndOrigins(t *testing.T) {
	gin.SetMode(gin.TestMode)
	keys := usecase.NewUserAPIKeyService(&middlewareAPIKeyRepo{})
	created, err := keys.Create(context.Background(), usecase.APIKeyCreateInput{OwnerID: "application-router-owner", Name: "application-router"})
	if err != nil {
		t.Fatal(err)
	}
	audit := &mcpAuditCapture{}
	middleware := &AuthMiddleware{APIKeys: keys, Audit: audit}
	router := NewApplicationRouter(ApplicationHandlers{Middleware: middleware, APIKey: &APIKeyHandler{Service: keys}}, ApplicationRouterOptions{AllowedOrigins: []string{"https://application.example"}})
	server := httptest.NewServer(router.Engine)
	defer server.Close()
	client := mcp.NewClient(&mcp.Implementation{Name: "application-router-test", Version: "1"}, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	session, err := client.Connect(ctx, &mcp.StreamableClientTransport{
		Endpoint:             server.URL + "/api/v1/mcp",
		HTTPClient:           &http.Client{Transport: mcpTransportHeaders{base: server.Client().Transport, token: "Bearer " + created.Secret}},
		DisableStandaloneSSE: true, MaxRetries: -1,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "account.api_key.list", Arguments: map[string]any{}})
	if err != nil || result.IsError {
		t.Fatalf("factory SDK call = %+v %v", result, err)
	}
	found := false
	for _, event := range audit.snapshot() {
		if event.Action == "mcp.account.api_key.list" {
			found = true
			if !event.Allowed || event.Status != "200" || event.ActorKind != "user_api_key" || event.CredentialID != created.Key.ID {
				t.Fatalf("factory dropped audit/principal: %+v", event)
			}
		}
	}
	if !found {
		t.Fatal("factory did not use the existing middleware audit sink")
	}
	for _, tc := range []struct {
		origin    string
		forbidden bool
	}{{"https://application.example", false}, {"https://foreign.example", true}} {
		request, err := http.NewRequestWithContext(ctx, http.MethodPost, server.URL+"/api/v1/mcp", strings.NewReader("{}"))
		if err != nil {
			t.Fatal(err)
		}
		request.Header.Set("Authorization", "Bearer "+created.Secret)
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Origin", tc.origin)
		response, err := server.Client().Do(request)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = io.Copy(io.Discard, response.Body)
		response.Body.Close()
		if (response.StatusCode == http.StatusForbidden) != tc.forbidden {
			t.Fatalf("origin %s status = %d", tc.origin, response.StatusCode)
		}
	}
}
