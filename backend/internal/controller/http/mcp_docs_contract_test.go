package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"

	swaggerdocs "github.com/mypocket/backend/docs"
)

// Omitting transport auth from generated Swagger makes an agent/release audit
// treat an owner-only endpoint as public. Discovery owns tool schemas, rather
// than a second, hand-maintained set of OpenAPI business DTOs.
func TestMCPTransportDocsDeclareAuthenticatedDiscovery(t *testing.T) {
	data, err := os.ReadFile("../../../docs/swagger.json")
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Paths map[string]map[string]struct {
			Security    []map[string][]string `json:"security"`
			Description string                `json:"description"`
		} `json:"paths"`
		SecurityDefinitions map[string]struct {
			Type string `json:"type"`
			In   string `json:"in"`
			Name string `json:"name"`
		} `json:"securityDefinitions"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	var checkedIn, served map[string]any
	if err := json.Unmarshal(data, &checkedIn); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(swaggerdocs.SwaggerInfo.ReadDoc()), &served); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(checkedIn["paths"], served["paths"]) || !reflect.DeepEqual(checkedIn["securityDefinitions"], served["securityDefinitions"]) {
		t.Fatal("runtime Swagger template differs from checked-in JSON; regenerate all artifacts")
	}
	bearer := doc.SecurityDefinitions["BearerAuth"]
	if bearer.Type != "apiKey" || bearer.In != "header" || bearer.Name != "Authorization" {
		t.Fatal("OpenAPI bearer header security is missing")
	}
	for _, method := range []string{"get", "post", "delete"} {
		op, exists := doc.Paths["/api/v1/mcp"][method]
		if !exists {
			t.Errorf("OpenAPI omits MCP %s", method)
			continue
		}
		authenticated := false
		for _, security := range op.Security {
			_, ok := security["BearerAuth"]
			authenticated = authenticated || ok
		}
		if !authenticated || !strings.Contains(op.Description, "tools/list") {
			t.Errorf("MCP %s must declare bearer auth and dynamic schema discovery", method)
		}
	}
}

// The JSON block is the agent-facing locator and exclusion manifest. A missing
// exclusion could imply user credentials authorize bootstrap/service routes.
func TestMCPAgentDocsCatalogLocatorAndExclusions(t *testing.T) {
	data, err := os.ReadFile("../../../../docs/ai/mcp-api-parity.md")
	if err != nil {
		t.Fatal(err)
	}
	_, block, found := strings.Cut(string(data), "```json\n")
	if !found {
		t.Fatal("agent reference omits JSON contract locator")
	}
	block, _, found = strings.Cut(block, "\n```")
	if !found {
		t.Fatal("unterminated JSON contract locator")
	}
	var contract struct {
		Endpoint           string   `json:"endpoint"`
		CatalogMethod      string   `json:"catalog_method"`
		ToolCount          int      `json:"tool_count"`
		ExcludedOperations []string `json:"excluded_operations"`
	}
	if err := json.Unmarshal([]byte(block), &contract); err != nil {
		t.Fatal(err)
	}
	if contract.Endpoint != "/api/v1/mcp" || contract.CatalogMethod != "tools/list" {
		t.Fatal("agent catalog locator does not resolve to SDK discovery")
	}
	if contract.ToolCount != 101 || len(mcpInventorySet()) != contract.ToolCount {
		t.Fatalf("published tool count=%d; authenticated inventory=%d", contract.ToolCount, len(mcpInventorySet()))
	}
	if len(contract.ExcludedOperations) != 19 || len(mcpExcludedRouteInventory) != 19 {
		t.Fatal("published exclusions must classify all 19 non-tool routes")
	}
	seen := map[string]bool{}
	for _, excluded := range contract.ExcludedOperations {
		if seen[excluded] || mcpExcludedRouteInventory[excluded] == "" {
			t.Errorf("duplicate or unclassified published exclusion: %s", excluded)
		}
		seen[excluded] = true
	}
	for excluded := range mcpExcludedRouteInventory {
		if !seen[excluded] {
			t.Errorf("agent contract omits exclusion: %s", excluded)
		}
	}
}

// Exercise the advertised catalog through the real SDK client. Removing a
// mounted tool or its usable schema must break discovery for API consumers.
func TestMCPMachineCatalogDocsPublishEveryRouteAndSchema(t *testing.T) {
	session := mcpTransportSession(t, NewMCPHandler(DefaultExternalOperationCatalog(), http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("discovery must not invoke business operations")
	}), MCPHandlerOptions{}), "Bearer placeholder")
	listed, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	want := mcpInventorySet()
	if len(listed.Tools) != 101 || len(listed.Tools) != len(want) {
		t.Fatalf("SDK tools/list publishes %d tools; want 101 authenticated operations", len(listed.Tools))
	}
	names, routes := map[string]bool{}, map[string]bool{}
	for _, tool := range listed.Tools {
		if names[tool.Name] || !externalOperationName.MatchString(tool.Name) || !want[tool.Description] || routes[tool.Description] {
			t.Errorf("invalid/duplicate discovery entry: %s %s", tool.Name, tool.Description)
		}
		names[tool.Name], routes[tool.Description] = true, true
		var schema map[string]any
		if err := json.Unmarshal(mustMarshalMCP(t, tool.InputSchema), &schema); err != nil {
			t.Fatal(err)
		}
		if schema["type"] != "object" || schema["additionalProperties"] != false || schema["properties"] == nil {
			t.Errorf("%s does not publish a closed usable input object", tool.Name)
		}
	}
	for route := range want {
		if !routes[route] {
			t.Errorf("SDK discovery omits authenticated operation: %s", route)
		}
	}
}
