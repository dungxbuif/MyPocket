package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
	"github.com/mypocket/backend/internal/usecase"
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
	for _, expected := range []string{"GET /api/v1/notifications", "POST /api/v1/notifications/:id/read", "POST /api/v1/notifications/devices", "DELETE /api/v1/notifications/devices/:id"} {
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

func TestPushTokenHandlersForwardOwnerAndNeverReturnToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &pushTokenHTTPRepo{items: map[string]entity.PushToken{}}
	h := &NotificationHandler{PushTokens: usecase.NewPushTokenInteractor(repo)}
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set(contextUserIDKey, "owner-1"); c.Next() })
	r.POST("/notifications/devices", h.RegisterPushToken)
	r.DELETE("/notifications/devices/:id", h.RevokePushToken)
	body, _ := json.Marshal(usecase.PushTokenInput{DeviceID: "iphone", DeviceName: "iPhone", Platform: "ios", Environment: "production", Token: strings.Repeat("a", 64)})
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, httptest.NewRequest(http.MethodPost, "/notifications/devices", bytes.NewReader(body)))
	if resp.Code != http.StatusOK || strings.Contains(resp.Body.String(), strings.Repeat("a", 64)) {
		t.Fatalf("token registration leaked/failed: status=%d body=%s", resp.Code, resp.Body.String())
	}
	var envelope struct {
		Data entity.PushToken `json:"data"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &envelope); err != nil || envelope.Data.ID == "" {
		t.Fatalf("missing registration metadata: %s", resp.Body.String())
	}
	revoke := httptest.NewRecorder()
	r.ServeHTTP(revoke, httptest.NewRequest(http.MethodDelete, "/notifications/devices/"+envelope.Data.ID, nil))
	if revoke.Code != http.StatusNoContent || repo.items[envelope.Data.ID].RevokedAt == nil {
		t.Fatalf("owner revoke failed: status=%d", revoke.Code)
	}
}

type pushTokenHTTPRepo struct{ items map[string]entity.PushToken }

func (r *pushTokenHTTPRepo) Save(_ context.Context, item *entity.PushToken) error {
	for id, old := range r.items {
		if old.OwnerID == item.OwnerID && old.DeviceID == item.DeviceID && old.Platform == item.Platform && old.Environment == item.Environment {
			item.ID = id
		}
	}
	r.items[item.ID] = *item
	return nil
}
func (r *pushTokenHTTPRepo) Revoke(_ context.Context, owner, id string) error {
	item, ok := r.items[id]
	if !ok || item.OwnerID != owner {
		return repository.ErrNotFound
	}
	now := time.Now()
	item.RevokedAt = &now
	r.items[id] = item
	return nil
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
