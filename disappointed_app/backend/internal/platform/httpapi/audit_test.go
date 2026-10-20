package httpapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"mypocket/internal/audit"
	"mypocket/internal/identity"
	"mypocket/internal/platform/httpapi"
)

func TestAuditEventsRequiresExactVerifiedViewerEmail(t *testing.T) {
	cfg := authTestConfig()
	cfg.AuditViewerEmail = "owner@example.com"
	cfg.AuditHashSecret = "01234567890123456789012345678901"
	viewer := &authRepoStub{user: identity.User{ID: "user_123", Email: "owner@example.com", EmailVerified: true}}
	audits := &auditRepoStub{events: []audit.Event{{ID: "audit_1", CorrelationID: "req_1", Action: "wallet.create", Outcome: audit.OutcomeSuccess, Severity: audit.SeverityInfo, Source: audit.SourceAPI}}}
	handler := httpapi.NewRouter(cfg, httpapi.Dependencies{IdentityRepository: viewer, AuditRepository: audits})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/audit/events?limit=10", nil)
	addAuthCookie(t, req, cfg.CookieSecret, "user_123")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected audit viewer 200, got %d: %s", res.Code, res.Body.String())
	}
	if !strings.Contains(res.Body.String(), "wallet.create") {
		t.Fatalf("expected audit event in response, got %s", res.Body.String())
	}

	blocked := &authRepoStub{user: identity.User{ID: "user_123", Email: "other@example.com", EmailVerified: true}}
	audits = &auditRepoStub{}
	handler = httpapi.NewRouter(cfg, httpapi.Dependencies{IdentityRepository: blocked, AuditRepository: audits})
	req = httptest.NewRequest(http.MethodGet, "/api/v1/audit/events", nil)
	addAuthCookie(t, req, cfg.CookieSecret, "user_123")
	res = httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden audit viewer, got %d: %s", res.Code, res.Body.String())
	}
	if len(audits.appended) != 1 || audits.appended[0].Action != "audit.viewer.denied" {
		t.Fatalf("expected denied audit event, got %#v", audits.appended)
	}
}

func TestAuditAccessCheckUsesSameViewerAuthorization(t *testing.T) {
	cfg := authTestConfig()
	cfg.AuditViewerEmail = "owner@example.com"
	viewer := &authRepoStub{user: identity.User{ID: "user_123", Email: "owner@example.com", EmailVerified: true}}
	handler := httpapi.NewRouter(cfg, httpapi.Dependencies{IdentityRepository: viewer})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/audit/access", nil)
	addAuthCookie(t, req, cfg.CookieSecret, "user_123")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK || !strings.Contains(res.Body.String(), `"allowed":true`) {
		t.Fatalf("expected allowed audit access response, got %d: %s", res.Code, res.Body.String())
	}

	blocked := &authRepoStub{user: identity.User{ID: "user_123", Email: "other@example.com", EmailVerified: true}}
	handler = httpapi.NewRouter(cfg, httpapi.Dependencies{IdentityRepository: blocked})
	req = httptest.NewRequest(http.MethodGet, "/api/v1/audit/access", nil)
	addAuthCookie(t, req, cfg.CookieSecret, "user_123")
	res = httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden audit access response, got %d: %s", res.Code, res.Body.String())
	}
}

func TestMutatingRequestAppendsGenericAuditEvent(t *testing.T) {
	cfg := authTestConfig()
	cfg.AuditHashSecret = "01234567890123456789012345678901"
	audits := &auditRepoStub{}
	handler := httpapi.NewRouter(cfg, httpapi.Dependencies{IdentityRepository: &authRepoStub{}, AuditRepository: audits})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.Header.Set("X-Correlation-ID", "req_logout")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusForbidden {
		t.Fatalf("expected csrf failure, got %d", res.Code)
	}
	if len(audits.appended) != 1 {
		t.Fatalf("expected one audit event, got %#v", audits.appended)
	}
	event := audits.appended[0]
	if event.CorrelationID != "req_logout" || event.Outcome != audit.OutcomeDenied || event.EntityType != "auth" {
		t.Fatalf("unexpected audit event: %#v", event)
	}
}

func TestAPIKeyMutatingRequestAuditsActorFromAuthContext(t *testing.T) {
	cfg := authTestConfig()
	cfg.AuditHashSecret = "01234567890123456789012345678901"
	cfg.APIKeyHashSecret = "change-this-development-api-key-hash-secret-32-bytes"
	audits := &auditRepoStub{}
	repo := &authRepoStub{user: identity.User{ID: "user_123", Email: "agent@example.com", EmailVerified: true}}
	handler := httpapi.NewRouter(cfg, httpapi.Dependencies{IdentityRepository: repo, APIKeyRepository: repo, AuditRepository: audits, AuthCache: &authCacheStub{}})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer mpk_test")
	req.Header.Set("X-Correlation-ID", "req_api_key_logout")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected api-key-authenticated mutation to bypass browser csrf, got %d: %s", res.Code, res.Body.String())
	}
	found := false
	for _, event := range audits.appended {
		if event.Action == "http.post.auth/logout" {
			found = true
			if event.ActorUserID != "user_123" || event.CorrelationID != "req_api_key_logout" {
				t.Fatalf("mutation audit did not include api key actor: %#v", event)
			}
		}
	}
	if !found {
		t.Fatalf("missing generic mutation audit event: %#v", audits.appended)
	}
}

func TestPanicRecoveryReturnsSafeCorrelatedError(t *testing.T) {
	cfg := authTestConfig()
	cfg.AuditHashSecret = "01234567890123456789012345678901"
	audits := &auditRepoStub{}
	handler := httpapi.NewRouter(cfg, httpapi.Dependencies{
		AuditRepository: audits,
		ReadyCheck:      func() error { panic("database password=secret stack") },
	})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health/ready", nil)
	req.Header.Set("X-Correlation-ID", "req_panic")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", res.Code, res.Body.String())
	}
	if !strings.Contains(res.Body.String(), `"correlation_id":"req_panic"`) || strings.Contains(res.Body.String(), "secret") {
		t.Fatalf("panic response was not safe/correlated: %s", res.Body.String())
	}
	if len(audits.appended) != 1 || audits.appended[0].Action != "http.panic" {
		t.Fatalf("expected panic audit event, got %#v", audits.appended)
	}
}

type auditRepoStub struct {
	events   []audit.Event
	appended []audit.Event
}

func (r *auditRepoStub) Append(_ context.Context, event audit.Event) error {
	r.appended = append(r.appended, event)
	return nil
}

func (r *auditRepoStub) List(context.Context, audit.Query) ([]audit.Event, error) {
	return r.events, nil
}

func (r *auditRepoStub) PurgeExpired(context.Context, int, int) (int, error) {
	return 0, nil
}

func addAuthCookie(t *testing.T, req *http.Request, secret string, userID string) {
	t.Helper()
	value, err := identity.NewCookieSigner([]byte(secret)).Sign(userID, time.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	req.AddCookie(&http.Cookie{Name: identity.AuthCookieName, Value: value})
}
