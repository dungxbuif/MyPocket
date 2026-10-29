package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestFeedbackAgentMiddlewareRequiresScopedBearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	middleware := NewFeedbackAgentMiddleware("feedback-test-token")
	for _, tc := range []struct {
		name   string
		header string
		status int
	}{
		{name: "missing", status: http.StatusUnauthorized},
		{name: "wrong scheme", header: "Basic feedback-test-token", status: http.StatusUnauthorized},
		{name: "wrong token", header: "Bearer another-token", status: http.StatusUnauthorized},
		{name: "valid", header: "Bearer feedback-test-token", status: http.StatusOK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			r.Use(middleware.RequireFeedbackAgent)
			r.GET("/", func(c *gin.Context) {
				if c.GetString(feedbackAgentActorKindKey) != feedbackAgentActorKind {
					t.Fatal("agent actor marker missing")
				}
				c.Status(http.StatusOK)
			})
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			res := httptest.NewRecorder()
			r.ServeHTTP(res, req)
			if res.Code != tc.status {
				t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
			}
			if tc.status != http.StatusOK && res.Body.String() == "" {
				t.Fatal("unauthorized response must contain problem details")
			}
		})
	}
}
