package httpapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"mypocket/internal/identity"
	"mypocket/internal/notification"
	"mypocket/internal/platform/httpapi"
)

type notificationRepoStub struct {
	listUser string
	readID   string
	items    []notification.Notice
}

func (s *notificationRepoStub) List(_ context.Context, userID string, _ int, _ time.Time) ([]notification.Notice, bool, error) {
	s.listUser = userID
	return s.items, false, nil
}
func (s *notificationRepoStub) MarkRead(_ context.Context, userID, id string) error {
	s.readID = userID + ":" + id
	return nil
}
func (*notificationRepoStub) UpsertPushSubscription(context.Context, string, notification.CreatePushSubscriptionInput) (notification.PushSubscription, error) {
	return notification.PushSubscription{ID: "push-1"}, nil
}
func (*notificationRepoStub) DeletePushSubscription(context.Context, string, string) error {
	return nil
}

func TestNotificationsAPIScopesAndMarksRead(t *testing.T) {
	repo := &notificationRepoStub{items: []notification.Notice{{ID: "notice-1", UserID: "user_123", Title: "Budget", CreatedAt: time.Now()}}}
	handler := httpapi.NewRouter(authTestConfig(), httpapi.Dependencies{IdentityRepository: &authRepoStub{user: identity.User{ID: "user_123", Email: "a@example.com", EmailVerified: true}}, NotificationRepository: repo})
	request := authenticatedRequest(t, http.MethodGet, "/api/v1/notifications", "")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || repo.listUser != "user_123" || !strings.Contains(response.Body.String(), "notice-1") {
		t.Fatalf("list failed: %d %s", response.Code, response.Body.String())
	}
	request = authenticatedRequest(t, http.MethodPatch, "/api/v1/notifications/notice-1/read", "")
	addCSRF(request)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || repo.readID != "user_123:notice-1" {
		t.Fatalf("mark read failed: %d %s", response.Code, response.Body.String())
	}
}

func TestNotificationsAPIUsesAPIKeyOwner(t *testing.T) {
	cfg := authTestConfig()
	cfg.APIKeyHashSecret = "change-this-development-api-key-hash-secret-32-bytes"
	identityRepo := &authRepoStub{user: identity.User{ID: "user_123", Email: "agent@example.com", EmailVerified: true}}
	repo := &notificationRepoStub{}
	handler := httpapi.NewRouter(cfg, httpapi.Dependencies{IdentityRepository: identityRepo, APIKeyRepository: identityRepo, NotificationRepository: repo, AuthCache: &authCacheStub{}})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications", nil)
	req.Header.Set("Authorization", "Bearer mpk_test")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK || repo.listUser != "user_123" {
		t.Fatalf("expected notifications scoped to API key owner, code=%d user=%q body=%s", res.Code, repo.listUser, res.Body.String())
	}
}
