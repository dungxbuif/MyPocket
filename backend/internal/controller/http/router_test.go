package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestPublicDocsRedirectsToVersionedSwaggerUI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	RegisterPublicDocsRedirect(engine)

	for _, path := range []string{"/docs", "/docs/"} {
		t.Run(path, func(t *testing.T) {
			record := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, path, nil)
			engine.ServeHTTP(record, request)

			if record.Code != http.StatusFound {
				t.Fatalf("status = %d, want %d", record.Code, http.StatusFound)
			}
			if got := record.Header().Get("Location"); got != "/api/v1/docs/index.html" {
				t.Fatalf("location = %q, want versioned Swagger UI", got)
			}
		})
	}
}
