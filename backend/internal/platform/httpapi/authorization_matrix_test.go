package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"mypocket/internal/identity"
	"mypocket/internal/platform/config"
)

type limiterStub struct {
	allowed bool
	retry   time.Duration
	err     error
}

type matrixAuthRepo struct{ user identity.User }

func (r *matrixAuthRepo) FindOrCreateGoogleUser(context.Context, identity.GoogleProfile) (identity.User, error) {
	return r.user, nil
}
func (r *matrixAuthRepo) FindByID(context.Context, string) (identity.User, error) { return r.user, nil }
func (r *matrixAuthRepo) CreateAPIKey(context.Context, string, string, string) (identity.CreatedAPIKey, error) {
	return identity.CreatedAPIKey{}, nil
}
func (r *matrixAuthRepo) ListAPIKeys(context.Context, string) ([]identity.APIKey, error) {
	return nil, nil
}
func (r *matrixAuthRepo) RevokeAPIKey(context.Context, string, string) (identity.APIKey, error) {
	return identity.APIKey{}, nil
}
func (r *matrixAuthRepo) AuthenticateAPIKey(_ context.Context, token, _ string) (identity.User, identity.APIKey, error) {
	if token != "mpk_test" {
		return identity.User{}, identity.APIKey{}, identity.ErrUserNotFound
	}
	return r.user, identity.APIKey{ID: "key-1", UserID: r.user.ID}, nil
}

type matrixCache struct{}

func (matrixCache) Get(context.Context, string) (string, bool, error)        { return "", false, nil }
func (matrixCache) Set(context.Context, string, string, time.Duration) error { return nil }
func (matrixCache) Delete(context.Context, string) error                     { return nil }

func (l limiterStub) Allow(context.Context, string, int, time.Duration) (bool, time.Duration, error) {
	return l.allowed, l.retry, l.err
}

func TestAuthorizationInventoryClassifiesEveryUserRoute(t *testing.T) {
	for _, route := range RegisteredRoutesForTest() {
		if strings.Contains(route.Path, "/health/") || route.Path == "/api/v1/openapi.json" || strings.HasPrefix(route.Path, "/api/v1/auth/google") {
			continue
		}
		if !route.Owned {
			t.Errorf("route lacks ownership classification: %s %s", route.Method, route.Path)
		}
	}
}

func TestBearerRateLimitReturns429AndFailsClosed(t *testing.T) {
	cfg := config.Config{AppEnv: "test", CookieSecret: strings.Repeat("c", 32), CSRFSecret: strings.Repeat("s", 32), APIKeyHashSecret: strings.Repeat("k", 32)}
	cfg.APIRateLimitPerMinute = 1
	repo := &matrixAuthRepo{user: identity.User{ID: "user-1"}}
	for _, test := range []struct {
		name    string
		limiter APIKeyLimiter
		status  int
	}{{"limited", limiterStub{allowed: false, retry: 12 * time.Second}, http.StatusTooManyRequests}, {"unavailable", limiterStub{err: errors.New("redis down")}, http.StatusServiceUnavailable}} {
		t.Run(test.name, func(t *testing.T) {
			next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
			handler := authContextMiddleware(cfg, repo, repo, matrixCache{}, test.limiter, nil, next)
			request := httptest.NewRequest(http.MethodGet, "/api/v1/wallets", nil)
			request.Header.Set("Authorization", "Bearer mpk_test")
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != test.status {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
			if test.status == http.StatusTooManyRequests && response.Header().Get("Retry-After") != "12" {
				t.Fatalf("retry-after=%q", response.Header().Get("Retry-After"))
			}
		})
	}
}
