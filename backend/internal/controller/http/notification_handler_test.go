package httpapi

import (
	"context"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mypocket/backend/internal/entity"
)

type notificationRouteStub struct {
	items []entity.Notification
	read  string
}

func (s *notificationRouteStub) List(context.Context, string) ([]entity.Notification, error) {
	return s.items, nil
}

func (s *notificationRouteStub) MarkRead(_ context.Context, _, id string) error {
	s.read = id
	return nil
}

func TestNotificationRoutesExposeOwnerScopedInbox(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := &Router{Engine: gin.New(), AuthMiddleware: &AuthMiddleware{}}
	router.RegisterNotificationRoutes(&NotificationHandler{})
	routes := map[string]bool{}
	for _, route := range router.Engine.Routes() {
		routes[route.Method+" "+route.Path] = true
	}
	for _, expected := range []string{"GET /api/v1/notifications", "POST /api/v1/notifications/:id/read"} {
		if !routes[expected] {
			t.Fatalf("missing route %s", expected)
		}
	}
	for route := range routes {
		if strings.Contains(route, "public") {
			t.Fatalf("notification route must not be public: %s", route)
		}
	}
}
