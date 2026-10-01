package httpapi

import (
	"context"
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

type feedbackRouteAPIKeyRepo struct{ key *entity.UserAPIKey }

func (r *feedbackRouteAPIKeyRepo) Create(_ context.Context, key *entity.UserAPIKey) error {
	r.key = key
	return nil
}
func (r *feedbackRouteAPIKeyRepo) FindByLookup(_ context.Context, lookup string) (*entity.UserAPIKey, error) {
	if r.key == nil || r.key.LookupID != lookup {
		return nil, repository.ErrNotFound
	}
	return r.key, nil
}
func (r *feedbackRouteAPIKeyRepo) FindByID(_ context.Context, id string) (*entity.UserAPIKey, error) {
	if r.key == nil || r.key.ID != id {
		return nil, repository.ErrNotFound
	}
	return r.key, nil
}
func (r *feedbackRouteAPIKeyRepo) ListByOwner(context.Context, string) ([]entity.UserAPIKey, error) {
	return nil, nil
}
func (r *feedbackRouteAPIKeyRepo) Revoke(context.Context, string, string, time.Time) error {
	return nil
}
func (r *feedbackRouteAPIKeyRepo) TouchLastUsed(context.Context, string, time.Time) error { return nil }

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

func TestFeedbackRoutesAuthenticateUserAPIKeyBeforeHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &feedbackRouteAPIKeyRepo{}
	keys := usecase.NewUserAPIKeyService(repo)
	created, err := keys.Create(context.Background(), usecase.APIKeyCreateInput{OwnerID: "owner-1", Name: "Feedback", Scopes: []string{entity.APIKeyScopeFeedbackRead}})
	if err != nil {
		t.Fatal(err)
	}
	engine := gin.New()
	router := &Router{Engine: engine, AuthMiddleware: &AuthMiddleware{APIKeys: keys}}
	router.RegisterFeedbackRoutes(&FeedbackHandler{}, &ChangelogHandler{}, NewFeedbackAgentMiddleware("token"))
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/feedback", nil)
	request.Header.Set("Authorization", "Bearer "+created.Secret)
	engine.ServeHTTP(recorder, request)
	if recorder.Code == http.StatusUnauthorized || recorder.Code == http.StatusForbidden {
		t.Fatalf("feedback route rejected a valid scoped API key before handler: %d %s", recorder.Code, recorder.Body.String())
	}
}
