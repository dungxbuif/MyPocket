package httpapi

import (
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestFeedbackRoutesExposeOwnerAgentAndPublicBoundaries(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := &Router{Engine: gin.New(), AuthMiddleware: &AuthMiddleware{}}
	router.RegisterFeedbackRoutes(&FeedbackHandler{}, &ChangelogHandler{}, NewFeedbackAgentMiddleware("token"))
	routes := map[string]bool{}
	for _, route := range router.Engine.Routes() {
		routes[route.Method+" "+route.Path] = true
	}
	for _, expected := range []string{
		"GET /api/v1/feedback", "POST /api/v1/feedback", "GET /api/v1/feedback/:id",
		"GET /api/v1/feedback/:id/screenshot", "GET /api/v1/agent/feedback", "GET /api/v1/agent/feedback/:id/screenshot",
		"PATCH /api/v1/internal/feedback/:id/status", "POST /api/v1/internal/changelog",
		"GET /api/v1/changelog", "GET /api/v1/changelog/:id",
	} {
		if !routes[expected] {
			t.Fatalf("missing route %s: %+v", expected, routes)
		}
	}
	for route := range routes {
		if strings.Contains(route, "/public/feedback") {
			t.Fatalf("raw public feedback route must not exist: %s", route)
		}
	}
}
