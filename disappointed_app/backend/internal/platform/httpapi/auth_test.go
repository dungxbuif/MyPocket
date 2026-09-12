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
	cfg := authTestConfig()
	cfg.PublicWebURL = "https://pocket.example.test"
	handler := httpapi.NewRouter(cfg, httpapi.Dependencies{IdentityRepository: repo})
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

func TestOAuthFixtureCallbackRejectsEmailOutsideWhitelist(t *testing.T) {
	repo := &authRepoStub{user: identity.User{ID: "user_123", Email: "blocked@example.com", EmailVerified: true}}
	cfg := authTestConfig()
	cfg.AllowedLoginEmails = []string{"allowed@example.com"}
	handler := httpapi.NewRouter(cfg, httpapi.Dependencies{IdentityRepository: repo})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/google/callback?subject=google-sub-1&email=blocked@example.com&email_verified=true", nil)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden, got %d: %s", res.Code, res.Body.String())
	}
	if repo.profile.Email != "" {
		t.Fatalf("repo should not create forbidden user, got profile %#v", repo.profile)
	}
	if !strings.Contains(res.Body.String(), "AUTH_FORBIDDEN") {
		t.Fatalf("expected AUTH_FORBIDDEN body, got %s", res.Body.String())
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

func TestAPIKeysCanBeCreatedListedAndRevoked(t *testing.T) {
	repo := &authRepoStub{user: identity.User{ID: "user_123", Email: "a@example.com", EmailVerified: true}}
	cache := &authCacheStub{}
	cfg := authTestConfig()
	cfg.APIKeyHashSecret = "change-this-development-api-key-hash-secret-32-bytes"
	authValue, csrfValue := signedAuthPair(t, cfg, "user_123")
	handler := httpapi.NewRouter(cfg, httpapi.Dependencies{IdentityRepository: repo, APIKeyRepository: repo, AuthCache: cache})

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/api-keys", strings.NewReader(`{"name":"Agent"}`))
	createReq.AddCookie(&http.Cookie{Name: identity.AuthCookieName, Value: authValue})
	createReq.AddCookie(&http.Cookie{Name: identity.CSRFCookieName, Value: csrfValue})
	createReq.Header.Set(identity.CSRFHeaderName, csrfValue)
	createRes := httptest.NewRecorder()
	handler.ServeHTTP(createRes, createReq)

	if createRes.Code != http.StatusCreated {
		t.Fatalf("expected create 201, got %d: %s", createRes.Code, createRes.Body.String())
	}
	if !strings.Contains(createRes.Body.String(), `"plaintext":"mpk_created"`) {
		t.Fatalf("created key should return plaintext once: %s", createRes.Body.String())
	}
	if cache.setKey == "" || cache.setValue != "user_123" {
		t.Fatalf("created key should be synced to cache, cache=%#v", cache)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/api-keys", nil)
	listReq.AddCookie(&http.Cookie{Name: identity.AuthCookieName, Value: authValue})
	listRes := httptest.NewRecorder()
	handler.ServeHTTP(listRes, listReq)
	if listRes.Code != http.StatusOK || strings.Contains(listRes.Body.String(), `"plaintext"`) || strings.Contains(listRes.Body.String(), "hash_1") {
		t.Fatalf("list should succeed without plaintext, code=%d body=%s", listRes.Code, listRes.Body.String())
	}

	revokeReq := httptest.NewRequest(http.MethodPost, "/api/v1/api-keys/key_1/revoke", nil)
	revokeReq.AddCookie(&http.Cookie{Name: identity.AuthCookieName, Value: authValue})
	revokeReq.AddCookie(&http.Cookie{Name: identity.CSRFCookieName, Value: csrfValue})
	revokeReq.Header.Set(identity.CSRFHeaderName, csrfValue)
	revokeRes := httptest.NewRecorder()
	handler.ServeHTTP(revokeRes, revokeReq)
	if revokeRes.Code != http.StatusOK || repo.revokedKeyID != "key_1" {
		t.Fatalf("revoke failed, code=%d key=%q body=%s", revokeRes.Code, repo.revokedKeyID, revokeRes.Body.String())
	}
	if cache.deletedKey == "" || cache.deletedKey != cache.setKey {
		t.Fatalf("revoked key should be removed from cache, cache=%#v", cache)
	}
}

func TestBearerAPIKeyAuthenticatesCurrentUserWithoutCookie(t *testing.T) {
	repo := &authRepoStub{user: identity.User{ID: "user_123", Email: "agent@example.com", EmailVerified: true}}
	cache := &authCacheStub{}
	cfg := authTestConfig()
	cfg.APIKeyHashSecret = "change-this-development-api-key-hash-secret-32-bytes"
	handler := httpapi.NewRouter(cfg, httpapi.Dependencies{IdentityRepository: repo, APIKeyRepository: repo, AuthCache: cache})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer mpk_test")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected api key auth to work, got %d: %s", res.Code, res.Body.String())
	}
	if repo.authenticatedToken != "mpk_test" || !strings.Contains(res.Body.String(), `"email":"agent@example.com"`) {
		t.Fatalf("api key auth did not resolve user, token=%q body=%s", repo.authenticatedToken, res.Body.String())
	}

	repo.authenticatedToken = ""
	cachedReq := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	cachedReq.Header.Set("Authorization", "Bearer mpk_test")
	cachedRes := httptest.NewRecorder()
	handler.ServeHTTP(cachedRes, cachedReq)
	if cachedRes.Code != http.StatusOK || repo.authenticatedToken != "mpk_test" {
		t.Fatalf("expected Redis hint to retain DB-backed key validation, code=%d token=%q body=%s", cachedRes.Code, repo.authenticatedToken, cachedRes.Body.String())
	}
}

func TestStaleRedisAPIKeyEntryDoesNotBypassRevocation(t *testing.T) {
	repo := &authRepoStub{user: identity.User{ID: "user_123", Email: "agent@example.com", EmailVerified: true}}
	cache := &authCacheStub{}
	cfg := authTestConfig()
	cfg.APIKeyHashSecret = "change-this-development-api-key-hash-secret-32-bytes"
	handler := httpapi.NewRouter(cfg, httpapi.Dependencies{IdentityRepository: repo, APIKeyRepository: repo, AuthCache: cache})
	first := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	first.Header.Set("Authorization", "Bearer mpk_test")
	firstRes := httptest.NewRecorder()
	handler.ServeHTTP(firstRes, first)
	if firstRes.Code != http.StatusOK {
		t.Fatalf("prime API key cache: %d %s", firstRes.Code, firstRes.Body.String())
	}

	repo.failAPIKeyAuth = true
	revoked := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	revoked.Header.Set("Authorization", "Bearer mpk_test")
	revokedRes := httptest.NewRecorder()
	handler.ServeHTTP(revokedRes, revoked)
	if revokedRes.Code != http.StatusUnauthorized {
		t.Fatalf("stale Redis entry must not authenticate a revoked key, got %d: %s", revokedRes.Code, revokedRes.Body.String())
	}
}

func TestBearerAPIKeyCannotCreateAnotherAPIKey(t *testing.T) {
	repo := &authRepoStub{user: identity.User{ID: "user_123", Email: "agent@example.com", EmailVerified: true}}
	cfg := authTestConfig()
	cfg.APIKeyHashSecret = "change-this-development-api-key-hash-secret-32-bytes"
	handler := httpapi.NewRouter(cfg, httpapi.Dependencies{IdentityRepository: repo, APIKeyRepository: repo, AuthCache: &authCacheStub{}})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/api-keys", strings.NewReader(`{"name":"Nested"}`))
	req.Header.Set("Authorization", "Bearer mpk_test")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusForbidden {
		t.Fatalf("expected bearer key management to be forbidden, got %d: %s", res.Code, res.Body.String())
	}
	if len(repo.keys) != 0 {
		t.Fatalf("api key auth should not create more keys: %#v", repo.keys)
	}
}

func TestInvalidBearerDoesNotFallBackToCookieSession(t *testing.T) {
	repo := &authRepoStub{user: identity.User{ID: "user_123", Email: "owner@example.com", EmailVerified: true}, failAPIKeyAuth: true}
	cfg := authTestConfig()
	cfg.APIKeyHashSecret = "change-this-development-api-key-hash-secret-32-bytes"
	authValue, _ := signedAuthPair(t, cfg, "user_123")
	handler := httpapi.NewRouter(cfg, httpapi.Dependencies{IdentityRepository: repo, APIKeyRepository: repo, AuthCache: &authCacheStub{}})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer mpk_revoked")
	req.AddCookie(&http.Cookie{Name: identity.AuthCookieName, Value: authValue})
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("invalid bearer must not inherit cookie identity, got %d: %s", res.Code, res.Body.String())
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
	if got := res.Header().Get("Access-Control-Allow-Methods"); !strings.Contains(got, "PUT") {
		t.Fatalf("expected PUT to be allowed for wallet/category activation, got %q", got)
	}
}

type authRepoStub struct {
	user               identity.User
	profile            identity.GoogleProfile
	keys               []identity.APIKey
	revokedKeyID       string
	authenticatedToken string
	failAPIKeyAuth     bool
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

func (r *authRepoStub) CreateAPIKey(_ context.Context, userID string, name string, _ string) (identity.CreatedAPIKey, error) {
	key := identity.APIKey{ID: "key_1", UserID: userID, Name: name, KeyPrefix: "mpk_created", KeyHash: "hash_1", CreatedAt: time.Now()}
	r.keys = append(r.keys, key)
	return identity.CreatedAPIKey{APIKey: key, Plaintext: "mpk_created"}, nil
}

func (r *authRepoStub) ListAPIKeys(_ context.Context, userID string) ([]identity.APIKey, error) {
	var keys []identity.APIKey
	for _, key := range r.keys {
		if key.UserID == userID {
			keys = append(keys, key)
		}
	}
	return keys, nil
}

func (r *authRepoStub) RevokeAPIKey(_ context.Context, userID string, keyID string) (identity.APIKey, error) {
	if userID != r.user.ID {
		return identity.APIKey{}, identity.ErrUserNotFound
	}
	r.revokedKeyID = keyID
	return identity.APIKey{ID: keyID, UserID: userID, KeyHash: "hash_1"}, nil
}

func (r *authRepoStub) AuthenticateAPIKey(_ context.Context, plaintext string, _ string) (identity.User, identity.APIKey, error) {
	if r.failAPIKeyAuth {
		return identity.User{}, identity.APIKey{}, identity.ErrUserNotFound
	}
	r.authenticatedToken = plaintext
	if plaintext != "mpk_test" {
		return identity.User{}, identity.APIKey{}, identity.ErrUserNotFound
	}
	return r.user, identity.APIKey{ID: "key_1", UserID: r.user.ID, KeyPrefix: "mpk_test"}, nil
}

type authCacheStub struct {
	values     map[string]string
	setKey     string
	setValue   string
	deletedKey string
}

func (c *authCacheStub) Get(_ context.Context, key string) (string, bool, error) {
	if c.values == nil {
		return "", false, nil
	}
	value, ok := c.values[key]
	return value, ok, nil
}

func (c *authCacheStub) Set(_ context.Context, key string, value string, _ time.Duration) error {
	if c.values == nil {
		c.values = map[string]string{}
	}
	c.setKey = key
	c.setValue = value
	c.values[key] = value
	return nil
}

func (c *authCacheStub) Delete(_ context.Context, key string) error {
	c.deletedKey = key
	delete(c.values, key)
	return nil
}

func signedAuthPair(t *testing.T, cfg config.Config, userID string) (string, string) {
	t.Helper()
	signer := identity.NewCookieSigner([]byte(cfg.CookieSecret))
	value, err := signer.Sign(userID, time.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	return value, "csrf-token"
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
