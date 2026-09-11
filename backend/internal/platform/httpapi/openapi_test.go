package httpapi

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestOpenAPICoversEveryVersionedRoute(t *testing.T) {
	var doc map[string]any
	if err := json.Unmarshal(openAPIDocument(), &doc); err != nil {
		t.Fatal(err)
	}
	paths := doc["paths"].(map[string]any)
	for _, route := range RegisteredRoutesForTest() {
		path := strings.TrimPrefix(route.Path, "/api/v1")
		operationPath, ok := paths[path].(map[string]any)
		if !ok {
			t.Errorf("missing path %s", route.Path)
			continue
		}
		operation, ok := operationPath[strings.ToLower(route.Method)].(map[string]any)
		if !ok {
			t.Errorf("missing operation %s %s", route.Method, route.Path)
			continue
		}
		responses := operation["responses"].(map[string]any)
		if _, ok := responses["429"]; !ok {
			t.Errorf("missing rate-limit response %s %s", route.Method, route.Path)
		}
		if route.Owned {
			if _, ok := operation["security"]; !ok {
				t.Errorf("missing security %s %s", route.Method, route.Path)
			}
		}
		if route.Idempotent {
			parameters, _ := operation["parameters"].([]any)
			if len(parameters) == 0 {
				t.Errorf("missing idempotency header %s %s", route.Method, route.Path)
			}
		}
	}
}

func TestEmbeddedOpenAPIHasStableErrorEnvelopeAndNoProductionSecret(t *testing.T) {
	body := string(openAPIDocument())
	for _, required := range []string{`"openapi":"3.1.0"`, `ErrorEnvelope`, `correlation_id`, `bearerAuth`, `csrfHeader`} {
		if !strings.Contains(body, required) {
			t.Fatalf("missing %s", required)
		}
	}
	if strings.Contains(body, "mpk_") || strings.Contains(body, "PRIVATE KEY") {
		t.Fatal("contract contains secret-like material")
	}
}
