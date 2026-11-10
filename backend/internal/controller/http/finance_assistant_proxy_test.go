package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
	"github.com/mypocket/backend/internal/usecase"
)

type proxyDelegationRepo struct {
	values  map[string]*entity.AssistantDelegation
	revoked []string
}

func (r *proxyDelegationRepo) Create(_ context.Context, value *entity.AssistantDelegation) error {
	if r.values == nil {
		r.values = map[string]*entity.AssistantDelegation{}
	}
	copy := *value
	r.values[value.ID] = &copy
	return nil
}
func (r *proxyDelegationRepo) FindByLookup(_ context.Context, lookup string) (*entity.AssistantDelegation, error) {
	for _, value := range r.values {
		if value.LookupID == lookup {
			copy := *value
			return &copy, nil
		}
	}
	return nil, repository.ErrNotFound
}
func (r *proxyDelegationRepo) FindByID(_ context.Context, id string) (*entity.AssistantDelegation, error) {
	value := r.values[id]
	if value == nil {
		return nil, repository.ErrNotFound
	}
	copy := *value
	return &copy, nil
}
func (r *proxyDelegationRepo) ListByOwner(context.Context, string) ([]entity.AssistantDelegation, error) {
	return nil, nil
}
func (r *proxyDelegationRepo) Revoke(_ context.Context, owner, id string, at time.Time) error {
	value := r.values[id]
	if value == nil || value.OwnerID != owner {
		return repository.ErrNotFound
	}
	value.RevokedAt = &at
	r.revoked = append(r.revoked, id)
	return nil
}
func (r *proxyDelegationRepo) TouchLastUsed(context.Context, string, time.Time) error { return nil }

func TestFinanceAssistantProxySignsOwnerAndStripsSpoofedHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-finance-assistant-shared-secret-at-least-32-bytes"
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Owner-ID") != "" {
			t.Fatal("spoofed owner header reached assistant")
		}
		encoded := r.Header.Get("Authorization")
		if len(encoded) < 8 {
			t.Fatal("missing assertion")
		}
		claims := &financeAssistantClaims{}
		_, err := jwt.ParseWithClaims(encoded[7:], claims, func(token *jwt.Token) (any, error) { return []byte(secret), nil }, jwt.WithAudience("finance-assistant"), jwt.WithIssuer("mypocket"))
		if err != nil {
			t.Fatalf("verify assertion: %v", err)
		}
		if claims.Subject != "owner-a" || claims.CredentialID != "session-a" {
			t.Fatalf("unexpected claims: %#v", claims)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"enabled": true}})
	}))
	defer upstream.Close()

	proxy := &FinanceAssistantProxy{BaseURL: upstream.URL, Secret: secret, Now: func() time.Time { return time.Unix(1_900_000_000, 0) }}
	engine := gin.New()
	engine.GET(advisorRoute+"/capabilities", func(c *gin.Context) {
		c.Set(contextAdvisorPrincipalKey, usecase.Principal{OwnerID: "owner-a", CredentialID: "session-a", CredentialKind: "session"})
		proxy.Forward(c)
	})
	request := httptest.NewRequest(http.MethodGet, advisorRoute+"/capabilities", nil)
	request.Header.Set("Authorization", "Bearer client-token")
	request.Header.Set("X-Owner-ID", "owner-b")
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestFinanceAssistantProxyUsesAndRevokesEphemeralLedgerDelegation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-finance-assistant-shared-secret-at-least-32-bytes"
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if token := r.Header.Get("X-MyPocket-Delegation"); len(token) < 5 || token[:4] != "mpa_" {
			t.Fatalf("missing scoped delegation: %q", token)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"status":"completed"}}`))
	}))
	defer upstream.Close()
	repo := &proxyDelegationRepo{}
	proxy := &FinanceAssistantProxy{
		BaseURL: upstream.URL, Secret: secret,
		Delegations: usecase.NewAssistantDelegationService(repo),
	}
	engine := gin.New()
	engine.POST(advisorRoute+"/messages", func(c *gin.Context) {
		c.Set(contextAdvisorPrincipalKey, usecase.Principal{OwnerID: "owner-a", CredentialID: "session-a", CredentialKind: "session"})
		proxy.Forward(c)
	})
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, advisorRoute+"/messages", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if len(repo.revoked) != 1 {
		t.Fatalf("ephemeral delegation must be revoked, got %v", repo.revoked)
	}
}

func TestFinanceAssistantProxyWrapsChannelPayloadForNativeEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-finance-assistant-shared-secret-at-least-32-bytes"
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/channels" {
			t.Fatalf("unexpected upstream path %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"adapter-1"}]`))
	}))
	defer upstream.Close()
	proxy := &FinanceAssistantProxy{BaseURL: upstream.URL, Secret: secret}
	engine := gin.New()
	engine.GET(advisorRoute+"/channels", func(c *gin.Context) {
		c.Set(contextAdvisorPrincipalKey, usecase.Principal{OwnerID: "owner-a", CredentialID: "session-a", CredentialKind: "session"})
		proxy.Forward(c)
	})
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, advisorRoute+"/channels", nil))
	if recorder.Code != http.StatusOK || recorder.Body.String() != `{"data":[{"id":"adapter-1"}]}` {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}
