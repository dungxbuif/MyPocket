package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/infrastructure/auth"
	redisadapter "github.com/mypocket/backend/internal/infrastructure/cache"
	postgresrepo "github.com/mypocket/backend/internal/infrastructure/repository"
	"github.com/mypocket/backend/internal/usecase"
)

// Network Redis, not an in-memory substitute. The wrapper only namespaces keys
// and tracks exact cleanup targets; it preserves real TTL/revocation behavior.
type offlineAuthCache struct {
	inner  *redisadapter.Redis
	prefix string
	mu     sync.Mutex
	keys   map[string]struct{}
}

func (c *offlineAuthCache) Set(key, value string, ttl time.Duration) error {
	key = c.prefix + key
	c.mu.Lock()
	c.keys[key] = struct{}{}
	c.mu.Unlock()
	return c.inner.Set(key, value, ttl)
}
func (c *offlineAuthCache) Get(key string) (string, bool, error) { return c.inner.Get(c.prefix + key) }
func (c *offlineAuthCache) Delete(key string) error              { return c.inner.Delete(c.prefix + key) }
func (c *offlineAuthCache) GetAndDelete(key string) (string, bool, error) {
	return c.inner.GetAndDelete(c.prefix + key)
}

func TestOfflineCreateHTTPWithRealJWTSessionRedisAndAPIKeyBoundary(t *testing.T) {
	redisURL := os.Getenv("TEST_REDIS_URL")
	if redisURL == "" {
		t.Skip("requires dedicated TEST_REDIS_URL")
	}
	f := newOfflineHTTPFixture(t)
	other := newOfflineHTTPFixture(t)
	network, err := redisadapter.NewRedis(redisURL)
	if err != nil {
		t.Fatal("cannot connect to disposable Redis")
	}
	cache := &offlineAuthCache{inner: network, prefix: "test:offline-auth:" + uuid.NewString() + ":", keys: map[string]struct{}{}}
	t.Cleanup(func() {
		cache.mu.Lock()
		defer cache.mu.Unlock()
		for key := range cache.keys {
			if err := network.Delete(key); err != nil {
				t.Error("fixture Redis cleanup failed")
			}
			if _, exists, err := network.Get(key); err != nil || exists {
				t.Error("fixture Redis key remains")
			}
		}
		if err := network.Close(); err != nil {
			t.Error(err)
		}
	})
	password := auth.NewPasswordService()
	hash, err := password.Hash("offline-fixture-password")
	if err != nil {
		t.Fatal(err)
	}
	for _, owner := range []string{f.owner, other.owner} {
		if err := f.db.Model(&entity.User{}).Where("id = ?", owner).Update("password_hash", hash).Error; err != nil {
			t.Fatal(err)
		}
	}
	const syntheticJWTSecret = "synthetic-offline-auth-secret-at-least-32-bytes"
	jwt := auth.NewJWT(syntheticJWTSecret, time.Hour)
	service := usecase.NewAuthInteractor(postgresrepo.NewUserPostgresRepository(f.db), password, jwt, cache, nil)
	login := func(owner string) *usecase.LoginOutput {
		t.Helper()
		result, err := service.Login(context.Background(), usecase.LoginInput{Email: owner + "@offline-http.test", Password: "offline-fixture-password"})
		if err != nil {
			t.Fatal("fixture login failed")
		}
		return result
	}
	firstSession := login(f.owner)
	otherSession := login(other.owner)
	keys := usecase.NewUserAPIKeyService(postgresrepo.NewUserAPIKeyPostgresRepository(f.db))
	t.Cleanup(func() {
		if err := f.db.Where("owner_id IN ?", []string{f.owner, other.owner}).Delete(&entity.UserAPIKey{}).Error; err != nil {
			t.Error(err)
		}
	})
	middleware := NewAuthMiddleware(jwt, func(sessionID string) (string, error) { return service.VerifySession(context.Background(), sessionID) })
	middleware.APIKeys = keys
	handler := NewTransactionHandler(postgresrepo.NewTransactionPostgresRepository(f.db), postgresrepo.NewWalletPostgresRepository(f.db), postgresrepo.NewCategoryPostgresRepository(f.db))
	router := NewRouter(NewAuthHandler(service, false, nil), nil, nil, nil, nil, handler, middleware, nil)
	capabilityStatus := func(token string) int {
		request := httptest.NewRequest(http.MethodGet, "/api/v1/sync/capabilities", nil)
		request.Header.Set("Authorization", "Bearer "+token)
		response := httptest.NewRecorder()
		router.Engine.ServeHTTP(response, request)
		return response.Code
	}
	if status := capabilityStatus(firstSession.Token); status != 200 {
		t.Fatalf("valid session capability status%d", status)
	}
	intent := uuid.NewString()
	call := func(token, body string) *httptest.ResponseRecorder {
		request := httptest.NewRequest(http.MethodPost, "/api/v1/transactions", bytes.NewBufferString(body))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Idempotency-Key", intent)
		if token != "" {
			request.Header.Set("Authorization", "Bearer "+token)
		}
		response := httptest.NewRecorder()
		router.Engine.ServeHTTP(response, request)
		return response
	}
	body := f.body(12500)
	first := call(firstSession.Token, body)
	if first.Code != 201 {
		t.Fatalf("JWT create failed:%d", first.Code)
	}
	var initial entity.Transaction
	if err := json.Unmarshal(decodeOfflineHTTP(t, first).Data, &initial); err != nil {
		t.Fatal(err)
	}
	if initial.OwnerID != f.owner {
		t.Fatal("JWT session owner was replaced by body owner")
	}
	replayed := call(firstSession.Token, body)
	if replayed.Code != 200 || !decodeOfflineHTTP(t, replayed).Meta.Mutation.Replayed {
		t.Fatalf("JWT replay failed:%d", replayed.Code)
	}
	foreign := call(otherSession.Token, body)
	if foreign.Code != 404 {
		t.Fatalf("foreign JWT replay status%d", foreign.Code)
	}
	claims, err := jwt.Verify(firstSession.Token)
	if err != nil {
		t.Fatal("fixture JWT invalid")
	}
	expired, _, err := auth.NewJWT(syntheticJWTSecret, -time.Hour).Generate(f.owner, "", "", claims.SessionID)
	if err != nil {
		t.Fatal(err)
	}
	for _, token := range []string{"", "invalid-jwt", expired} {
		if response := call(token, body); response.Code != 401 {
			t.Fatalf("invalid/expired JWT reached receipt:%d", response.Code)
		}
	}
	key, err := keys.Create(context.Background(), usecase.APIKeyCreateInput{OwnerID: f.owner, Name: "Offline read boundary", Scopes: []string{entity.APIKeyScopeFinanceRead}})
	if err != nil {
		t.Fatal(err)
	}
	principal, err := keys.Authenticate(context.Background(), key.Secret, []string{entity.APIKeyScopeFinanceRead})
	if err != nil || principal.OwnerID != f.owner {
		t.Fatal("real read API key failed authentication")
	}
	if principal, err := keys.Authenticate(context.Background(), key.Secret, []string{entity.APIKeyScopeAdvisorChat}); err != nil || principal.OwnerID != f.owner {
		t.Fatal("legacy read key must have full owner access")
	}
	if response := call(key.Secret, body); response.Code != 200 || !decodeOfflineHTTP(t, response).Meta.Mutation.Replayed {
		t.Fatalf("full owner API key cannot replay its own financial receipt:%d", response.Code)
	}
	if err := keys.Revoke(context.Background(), f.owner, key.Key.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := keys.Authenticate(context.Background(), key.Secret, []string{entity.APIKeyScopeFinanceRead}); !errors.Is(err, usecase.ErrAPIKeyInvalid) {
		t.Fatal("revoked key remained valid")
	}
	if response := call(key.Secret, body); response.Code != 401 {
		t.Fatal("revoked key reached financial receipt")
	}
	expiry := time.Now().Add(time.Hour)
	expiring, err := keys.Create(context.Background(), usecase.APIKeyCreateInput{OwnerID: f.owner, Name: "Expiry boundary", Scopes: []string{entity.APIKeyScopeFinanceRead}, ExpiresAt: &expiry})
	if err != nil {
		t.Fatal(err)
	}
	if err := f.db.Model(&entity.UserAPIKey{}).Where("id = ? AND owner_id = ?", expiring.Key.ID, f.owner).Update("expires_at", time.Now().Add(-time.Hour)).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := keys.Authenticate(context.Background(), expiring.Secret, []string{entity.APIKeyScopeFinanceRead}); !errors.Is(err, usecase.ErrAPIKeyInvalid) {
		t.Fatal("expired key remained valid")
	}
	if capabilityStatus(expiring.Secret) != 401 {
		t.Fatal("expired key reached authenticated capability")
	}
	shortSession := login(f.owner)
	shortClaims, err := jwt.Verify(shortSession.Token)
	if err != nil {
		t.Fatal("fixture short session JWT invalid")
	}
	sessionKey := service.SessionPrefix + shortClaims.SessionID
	stored, found, err := cache.Get(sessionKey)
	if err != nil || !found {
		t.Fatal("real Redis session was not stored")
	}
	if err := cache.Set(sessionKey, stored, time.Millisecond); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(time.Second)
	for {
		_, found, err = cache.Get(sessionKey)
		if err != nil {
			t.Fatal(err)
		}
		if !found {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("Redis did not expire fixture session")
		}
		time.Sleep(time.Millisecond)
	}
	if response := call(shortSession.Token, body); response.Code != 401 {
		t.Fatalf("expired Redis session reached receipt:%d", response.Code)
	}
	if err := service.Logout(context.Background(), f.owner, claims.SessionID, firstSession.RefreshToken); err != nil {
		t.Fatal("fixture logout failed")
	}
	if response := call(firstSession.Token, body); response.Code != 401 {
		t.Fatalf("revoked session replayed receipt:%d", response.Code)
	}
	if status := capabilityStatus(firstSession.Token); status != 401 {
		t.Fatalf("revoked session read capability:%d", status)
	}
	var count int64
	if err := f.db.Model(&entity.Transaction{}).Where("owner_id = ?", f.owner).Count(&count).Error; err != nil || count != 1 {
		t.Fatalf("auth-denied requests changed ledger:%d err%v", count, err)
	}
	if err := f.db.Model(&entity.TransactionMutation{}).Where("owner_id = ?", f.owner).Count(&count).Error; err != nil || count != 1 {
		t.Fatalf("auth-denied requests changed receipts:%d err%v", count, err)
	}
}
