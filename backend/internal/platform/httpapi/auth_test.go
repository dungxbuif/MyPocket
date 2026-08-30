package httpapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"mypocket/internal/identity"
	"mypocket/internal/platform/config"
	"mypocket/internal/platform/httpapi"
)

func TestOAuthFixtureCallbackSetsSecureAuthCookie(t *testing.T) {
	repo := &authRepoStub{user: identity.User{ID: "user_123", Email: "a@example.com", EmailVerified: true, DisplayName: "A"}}
	handler := httpapi.NewRouter(authTestConfig(), httpapi.Dependencies{IdentityRepository: repo})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/google/callback?subject=google-sub-1&email=a@example.com&email_verified=true&name=A", nil)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusFound {
		t.Fatalf("expected redirect, got %d: %s", res.Code, res.Body.String())
	}
	if repo.profile.Subject != "google-sub-1" || repo.profile.Email != "a@example.com" {
		t.Fatalf("wrong fixture profile: %#v", repo.profile)
	}
	authCookie := findCookie(t, res.Result().Cookies(), identity.AuthCookieName)
	if !authCookie.HttpOnly || !authCookie.Secure || authCookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("auth cookie missing secure flags: %#v", authCookie)
	}
	csrfCookie := findCookie(t, res.Result().Cookies(), identity.CSRFCookieName)
	if csrfCookie.HttpOnly || !csrfCookie.Secure || csrfCookie.Value == "" {
		t.Fatalf("csrf cookie missing browser-readable secure contract: %#v", csrfCookie)
	}
}

func TestCurrentUserReturnsAuthenticatedUser(t *testing.T) {
	repo := &authRepoStub{user: identity.User{ID: "user_123", Email: "a@example.com", EmailVerified: true, DisplayName: "A"}}
	cfg := authTestConfig()
	signer := identity.NewCookieSigner([]byte(cfg.CookieSecret))
	value, err := signer.Sign("user_123", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	handler := httpapi.NewRouter(cfg, httpapi.Dependencies{IdentityRepository: repo})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.AddCookie(&http.Cookie{Name: identity.AuthCookieName, Value: value})
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	body := res.Body.String()
	if !strings.Contains(body, `"email":"a@example.com"`) || strings.Contains(body, "google-sub") || strings.Contains(body, "token") {
		t.Fatalf("unexpected current user body: %s", body)
	}
}

func TestLogoutRequiresCSRF(t *testing.T) {
	handler := httpapi.NewRouter(authTestConfig(), httpapi.Dependencies{IdentityRepository: &authRepoStub{}})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.Header.Set("X-Correlation-ID", "req_logout")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", res.Code, res.Body.String())
	}
	if !strings.Contains(res.Body.String(), "CSRF_REQUIRED") {
		t.Fatalf("expected csrf code, got %s", res.Body.String())
	}
	if !strings.Contains(res.Body.String(), `"correlation_id":"req_logout"`) {
		t.Fatalf("expected correlation id, got %s", res.Body.String())
	}
}

func TestCORSAllowsPublicWebOriginCredentials(t *testing.T) {
	handler := httpapi.NewRouter(authTestConfig(), httpapi.Dependencies{IdentityRepository: &authRepoStub{}})
	req := httptest.NewRequest(http.MethodOptions, "/api/v1/auth/logout", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "X-CSRF-Token, Idempotency-Key")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusNoContent {
		t.Fatalf("expected 204 preflight, got %d: %s", res.Code, res.Body.String())
	}
	if got := res.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("wrong allow origin: %q", got)
	}
	if got := res.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("wrong allow credentials: %q", got)
	}
	if got := res.Header().Get("Vary"); !strings.Contains(got, "Origin") {
		t.Fatalf("expected Origin vary header, got %q", got)
	}
	if got := res.Header().Get("Access-Control-Allow-Headers"); !strings.Contains(got, "Idempotency-Key") {
		t.Fatalf("expected idempotency header to be allowed, got %q", got)
	}
}

type authRepoStub struct {
	user    identity.User
	profile identity.GoogleProfile
}

func (r *authRepoStub) FindOrCreateGoogleUser(_ context.Context, profile identity.GoogleProfile) (identity.User, error) {
	r.profile = profile
	return r.user, nil
}

func (r *authRepoStub) FindByID(_ context.Context, id string) (identity.User, error) {
	if id != r.user.ID {
		return identity.User{}, identity.ErrUserNotFound
	}
	return r.user, nil
}

func authTestConfig() config.Config {
	return config.Config{
		AppEnv:           "test",
		PublicWebURL:     "http://localhost:5173",
		CookieSecret:     "01234567890123456789012345678901",
		CSRFSecret:       "abcdefghijklmnopqrstuvwxyz123456",
		OAuthFixtureMode: true,
	}
}

func findCookie(t *testing.T, cookies []*http.Cookie, name string) *http.Cookie {
	t.Helper()

	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie
		}
	}
	t.Fatalf("missing cookie %s in %#v", name, cookies)
	return nil
}
