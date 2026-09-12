package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mypocket/internal/platform/config"
	"mypocket/internal/platform/httpapi"
)

func TestDocsHandler_RequiresAuth_WhenUnauthenticated(t *testing.T) {
	cfg := authTestConfig()
	cfg.DocsDir = "testdata/docs"

	handler := http.HandlerFunc(httpapi.DocsHandler(cfg))

	tmpDir := t.TempDir()
	cfg.DocsDir = tmpDir
	os.WriteFile(filepath.Join(tmpDir, "index.html"), []byte("<html>docs</html>"), 0644)

	handler = http.HandlerFunc(httpapi.DocsHandler(cfg))

	req := httptest.NewRequest(http.MethodGet, "/docs/", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusSeeOther {
		t.Fatalf("expected 303 redirect for unauthenticated request, got %d", res.Code)
	}
	loc := res.Header().Get("Location")
	if loc != "/" {
		t.Fatalf("expected redirect to /, got %s", loc)
	}
}

func TestDocsHandler_ServesIndexHTML_WhenAuthenticated(t *testing.T) {
	cfg := authTestConfig()
	tmpDir := t.TempDir()
	cfg.DocsDir = tmpDir
	os.WriteFile(filepath.Join(tmpDir, "index.html"), []byte("<html>docs</html>"), 0644)

	handler := http.HandlerFunc(httpapi.DocsHandler(cfg))

	req := authenticatedRequest(t, http.MethodGet, "/docs/", "")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	if !strings.Contains(res.Body.String(), "docs") {
		t.Fatalf("expected docs content, got %s", res.Body.String())
	}
}

func TestDocsHandler_ServesSubPath_WhenAuthenticated(t *testing.T) {
	cfg := authTestConfig()
	tmpDir := t.TempDir()
	cfg.DocsDir = tmpDir
	os.WriteFile(filepath.Join(tmpDir, "index.html"), []byte("<html>docs</html>"), 0644)
	os.MkdirAll(filepath.Join(tmpDir, "api"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "api", "wallets.html"), []byte("<html>wallets api</html>"), 0644)

	handler := http.HandlerFunc(httpapi.DocsHandler(cfg))

	req := authenticatedRequest(t, http.MethodGet, "/docs/api/wallets.html", "")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	if !strings.Contains(res.Body.String(), "wallets api") {
		t.Fatalf("expected wallets api content, got %s", res.Body.String())
	}
}

func TestDocsHandler_Returns404_WhenDocsDirNotSet(t *testing.T) {
	cfg := config.Config{}

	handler := http.HandlerFunc(httpapi.DocsHandler(cfg))

	req := httptest.NewRequest(http.MethodGet, "/docs/", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf("expected 404 when DOCS_DIR is empty, got %d", res.Code)
	}
}

func TestDocsHandler_FallsBackToIndexHTML_ForUnknownPath(t *testing.T) {
	cfg := authTestConfig()
	tmpDir := t.TempDir()
	cfg.DocsDir = tmpDir
	os.WriteFile(filepath.Join(tmpDir, "index.html"), []byte("<html>fallback</html>"), 0644)

	handler := http.HandlerFunc(httpapi.DocsHandler(cfg))

	req := authenticatedRequest(t, http.MethodGet, "/docs/nonexistent/page", "")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200 with fallback, got %d: %s", res.Code, res.Body.String())
	}
	if !strings.Contains(res.Body.String(), "fallback") {
		t.Fatalf("expected fallback content, got %s", res.Body.String())
	}
}
