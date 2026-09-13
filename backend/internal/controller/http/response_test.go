package httpapi

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestOKUsesDataEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/", func(c *gin.Context) { OK(c, map[string]string{"name": "demo"}) })
	res := httptest.NewRecorder()
	r.ServeHTTP(res, httptest.NewRequest("GET", "/", nil))
	if res.Code != 200 || res.Body.String() != `{"data":{"name":"demo"},"meta":null}` {
		t.Fatalf("unexpected response: %d %s", res.Code, res.Body.String())
	}
}

func TestFailUsesProblemDetails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/", func(c *gin.Context) {
		Fail(c, 401, Problem{Code: "AUTH_REQUIRED", Title: "Unauthorized", Detail: "chưa đăng nhập"})
	})
	res := httptest.NewRecorder()
	r.ServeHTTP(res, httptest.NewRequest("GET", "/", nil))
	if res.Code != 401 || res.Header().Get("Content-Type") != "application/problem+json" {
		t.Fatalf("unexpected response: %d %s", res.Code, res.Header().Get("Content-Type"))
	}
}
