package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestChangelogPublicRouteDoesNotRequireUserContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := &Router{Engine: gin.New(), AuthMiddleware: &AuthMiddleware{}}
	router.RegisterFeedbackRoutes(&FeedbackHandler{}, &ChangelogHandler{}, NewFeedbackAgentMiddleware("token"))
	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/changelog", nil)
	router.Engine.ServeHTTP(res, req)
	if res.Code == http.StatusUnauthorized {
		t.Fatal("public changelog must not be behind user JWT middleware")
	}
}
