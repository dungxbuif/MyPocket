package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
	"github.com/mypocket/backend/internal/usecase"
)

type middlewareAPIKeyRepo struct{ key *entity.UserAPIKey }

type middlewareAuditStub struct{ events []repository.AuditEvent }

func (s *middlewareAuditStub) Record(_ context.Context, event repository.AuditEvent) error {
	s.events = append(s.events, event)
	return nil
}

func (r *middlewareAPIKeyRepo) Create(_ context.Context, key *entity.UserAPIKey) error {
	r.key = key
	return nil
}
func (r *middlewareAPIKeyRepo) FindByLookup(_ context.Context, lookup string) (*entity.UserAPIKey, error) {
	if r.key == nil || r.key.LookupID != lookup {
		return nil, repository.ErrNotFound
	}
	return r.key, nil
}
func (r *middlewareAPIKeyRepo) FindByID(_ context.Context, id string) (*entity.UserAPIKey, error) {
	if r.key == nil || r.key.ID != id {
		return nil, repository.ErrNotFound
	}
	return r.key, nil
}
func (r *middlewareAPIKeyRepo) ListByOwner(context.Context, string) ([]entity.UserAPIKey, error) {
	return nil, nil
}
func (r *middlewareAPIKeyRepo) Revoke(context.Context, string, string, time.Time) error { return nil }
func (r *middlewareAPIKeyRepo) TouchLastUsed(context.Context, string, time.Time) error  { return nil }

func TestRequireAdvisorAuthAcceptsUserAPIKeyWithoutJWTFallback(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &middlewareAPIKeyRepo{}
	service := usecase.NewUserAPIKeyService(repo)
	created, err := service.Create(context.Background(), usecase.APIKeyCreateInput{OwnerID: "owner-1", Name: "CI", Scopes: []string{entity.APIKeyScopeAdvisorChat}})
	if err != nil {
		t.Fatal(err)
	}
	audit := &middlewareAuditStub{}
	middleware := &AuthMiddleware{APIKeys: service, Audit: audit}
	engine := gin.New()
	engine.GET(advisorRoute+"/overview", middleware.RequireAdvisorAuth, func(c *gin.Context) {
		value, exists := c.Get(contextAdvisorPrincipalKey)
		principal, ok := value.(usecase.Principal)
		if !exists || !ok || principal.OwnerID != "owner-1" {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusNoContent)
	})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, advisorRoute+"/overview", nil)
	request.Header.Set("Authorization", "Bearer "+created.Secret)
	engine.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("valid API key should authenticate: %d %s", recorder.Code, recorder.Body.String())
	}
	if len(audit.events) != 1 || !audit.events[0].Allowed || audit.events[0].ActorKind != "user_api_key" || audit.events[0].CredentialID != created.Key.ID {
		t.Fatalf("advisor access audit mismatch: %+v", audit.events)
	}
	badRecorder := httptest.NewRecorder()
	badRequest := httptest.NewRequest(http.MethodGet, advisorRoute+"/overview", nil)
	badRequest.Header.Set("Authorization", "Bearer mpk_invalid.invalid")
	engine.ServeHTTP(badRecorder, badRequest)
	if badRecorder.Code != http.StatusUnauthorized {
		t.Fatalf("invalid mpk key must not fall back to JWT: %d", badRecorder.Code)
	}
	if len(audit.events) != 2 || audit.events[1].Allowed || audit.events[1].Reason != "denied" {
		t.Fatalf("invalid-key audit mismatch: %+v", audit.events)
	}
}
