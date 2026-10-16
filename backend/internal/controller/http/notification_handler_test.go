package httpapi

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mypocket/backend/internal/entity"
)

type notificationRouteStub struct {
	items     []entity.Notification
	listOwner string
	readOwner string
	read      string
}

func (s *notificationRouteStub) List(_ context.Context, owner string) ([]entity.Notification, error) {
	s.listOwner = owner
	return s.items, nil
}

func (s *notificationRouteStub) MarkRead(_ context.Context, owner, id string) error {
	s.readOwner = owner
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

func TestNotificationHandlersForwardAuthenticatedOwner(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &notificationRouteStub{items: []entity.Notification{{ID: "notice-1", OwnerID: "owner-1"}}}
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set(contextUserIDKey, "owner-1"); c.Next() })
	h := &NotificationHandler{Notifications: stub}
	router.GET("/notifications", h.List)
	router.POST("/notifications/:id/read", h.MarkRead)

	listResponse := httptest.NewRecorder()
	router.ServeHTTP(listResponse, httptest.NewRequest("GET", "/notifications", nil))
	if listResponse.Code != 200 || stub.listOwner != "owner-1" {
		t.Fatalf("list owner scope missing: status=%d owner=%q body=%s", listResponse.Code, stub.listOwner, listResponse.Body.String())
	}
	readResponse := httptest.NewRecorder()
	router.ServeHTTP(readResponse, httptest.NewRequest("POST", "/notifications/notice-1/read", nil))
	if readResponse.Code != 204 || stub.readOwner != "owner-1" || stub.read != "notice-1" {
		t.Fatalf("read owner scope missing: status=%d owner=%q id=%q", readResponse.Code, stub.readOwner, stub.read)
	}
}
