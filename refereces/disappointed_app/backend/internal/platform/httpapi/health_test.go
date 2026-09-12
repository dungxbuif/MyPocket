package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"mypocket/internal/platform/config"
	"mypocket/internal/platform/httpapi"
)

func TestLiveHealthReturnsOKWithCorrelationID(t *testing.T) {
	handler := httpapi.NewRouter(config.Config{AppEnv: "test"}, httpapi.Dependencies{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health/live", nil)
	req.Header.Set("X-Correlation-ID", "req_live")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	if got := res.Header().Get("X-Correlation-ID"); got != "req_live" {
		t.Fatalf("expected response correlation id req_live, got %q", got)
	}
	if !strings.Contains(res.Body.String(), `"status":"ok"`) {
		t.Fatalf("expected ok body, got %s", res.Body.String())
	}
}

func TestReadyHealthReportsDependencyFailureSafely(t *testing.T) {
	handler := httpapi.NewRouter(config.Config{AppEnv: "test"}, httpapi.Dependencies{
		ReadyCheck: func() error { return assertErr("database password=secret unavailable") },
	})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health/ready", nil)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", res.Code, res.Body.String())
	}
	if strings.Contains(res.Body.String(), "secret") {
		t.Fatalf("readiness leaked secret: %s", res.Body.String())
	}
	if !strings.Contains(res.Body.String(), `"code":"INTERNAL_RETRYABLE"`) {
		t.Fatalf("expected retryable error code, got %s", res.Body.String())
	}
}

type assertErr string

func (e assertErr) Error() string {
	return string(e)
}
