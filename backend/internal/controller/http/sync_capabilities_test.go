package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mypocket/backend/internal/infrastructure/auth"
	postgresrepo "github.com/mypocket/backend/internal/infrastructure/repository"
)

func TestSyncCapabilitiesIsNotCacheable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	handler := &TransactionHandler{}
	engine.GET("/sync/capabilities", handler.SyncCapabilities)
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/sync/capabilities", nil))
	if got := response.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("offline capability must not be cached: Cache-Control=%q", got)
	}
}

func TestSyncCapabilitiesAdvertisesOnlyConfiguredDurableCreateContract(t *testing.T) {
	f := newOfflineHTTPFixture(t)
	jwt := auth.NewJWT("synthetic-capability-secret-at-least32bytes", time.Hour)
	token, _, err := jwt.Generate(f.owner, "", "", "capability-session")
	if err != nil {
		t.Fatal(err)
	}
	middleware := NewAuthMiddleware(jwt, func(_ string) (string, error) { return f.owner, nil })
	for _, tc := range []struct {
		name    string
		handler *TransactionHandler
		enabled bool
	}{
		{"durable", NewTransactionHandler(postgresrepo.NewTransactionPostgresRepository(f.db), postgresrepo.NewWalletPostgresRepository(f.db), postgresrepo.NewCategoryPostgresRepository(f.db)), true},
		{"legacy", NewTransactionHandler(&transactionRepositoryStub{}, &transactionWalletRepositoryStub{}, &transactionCategoryRepositoryStub{}), false},
		{"unconfigured", nil, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := NewRouter(NewAuthHandler(nil, false, nil), nil, nil, nil, nil, tc.handler, middleware, nil)
			request := httptest.NewRequest(http.MethodGet, "/api/v1/sync/capabilities", nil)
			request.Header.Set("Authorization", "Bearer "+token)
			response := httptest.NewRecorder()
			r.Engine.ServeHTTP(response, request)
			if response.Code != 200 {
				t.Fatalf("missing capability endpoint:%d", response.Code)
			}
			if got := response.Header().Get("Cache-Control"); got != "no-store" {
				t.Fatalf("offline capability must not be cached: Cache-Control=%q", got)
			}
			var envelope struct {
				Data struct {
					Version         int      `json:"contract_version"`
					Enabled         bool     `json:"enabled"`
					Operations      []string `json:"operations"`
					ReceiptMetadata string   `json:"receipt_metadata"`
					KeyRequired     bool     `json:"idempotency_key_required_for_offline"`
					Legacy          bool     `json:"legacy_unkeyed_online"`
					CreateContext   string   `json:"transaction_create_context"`
				} `json:"data"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
				t.Fatal(err)
			}
			cap := envelope.Data
			if cap.Version != 1 || cap.Enabled != tc.enabled || !cap.KeyRequired || !cap.Legacy || cap.ReceiptMetadata != "/meta/mutation" || cap.Operations == nil {
				t.Fatal("capability contract mismatch")
			}
			if tc.enabled {
				if cap.CreateContext != "explicit_travel_v1" {
					t.Fatal("durable writer cannot freeze entry-time travel context")
				}
				if len(cap.Operations) != 2 || cap.Operations[0] != "transaction.create" || cap.Operations[1] != "transfer.create" {
					t.Fatal("unexpected offline operations")
				}
			} else if len(cap.Operations) != 0 {
				t.Fatal("unconfigured writer advertised operations")
			}
			if !tc.enabled && cap.CreateContext != "" {
				t.Fatal("disabled writer advertised frozen context")
			}
			unauth := httptest.NewRecorder()
			r.Engine.ServeHTTP(unauth, httptest.NewRequest(http.MethodGet, "/api/v1/sync/capabilities", nil))
			if unauth.Code != 401 {
				t.Fatal("capability bypassed session authentication")
			}
		})
	}
}
