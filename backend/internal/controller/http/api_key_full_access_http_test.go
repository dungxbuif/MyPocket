package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	postgresrepo "github.com/mypocket/backend/internal/infrastructure/repository"
	"github.com/mypocket/backend/internal/usecase"
)

func TestFullAPIKeyFinancialWritesManagementAndOwnerIsolation(t *testing.T) {
	f, other := newOfflineHTTPFixture(t), newOfflineHTTPFixture(t)
	keys := usecase.NewUserAPIKeyService(postgresrepo.NewUserAPIKeyPostgresRepository(f.db))
	key, err := keys.Create(context.Background(), usecase.APIKeyCreateInput{OwnerID: f.owner, Name: "Owner automation"})
	if err != nil {
		t.Fatal(err)
	}
	// Simulate a persisted pre-change limited key, not a freshly normalized one.
	if err := f.db.Model(&entity.UserAPIKey{}).Where("id=?", key.Key.ID).Update("scopes", `["feedback:read"]`).Error; err != nil {
		t.Fatal(err)
	}
	otherKey, err := keys.Create(context.Background(), usecase.APIKeyCreateInput{OwnerID: other.owner, Name: "Other owner"})
	if err != nil {
		t.Fatal(err)
	}
	middleware := &AuthMiddleware{APIKeys: keys}
	transactions := NewTransactionHandler(postgresrepo.NewTransactionPostgresRepository(f.db), postgresrepo.NewWalletPostgresRepository(f.db), postgresrepo.NewCategoryPostgresRepository(f.db))
	router := NewRouter(NewAuthHandler(nil, false, nil), nil, nil, nil, NewWalletHandler(postgresrepo.NewWalletPostgresRepository(f.db)), transactions, middleware, nil)
	router.RegisterAPIKeyRoutes(&APIKeyHandler{Service: keys})
	call := func(method, path, body, token, intent string) *httptest.ResponseRecorder {
		t.Helper()
		request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", "Bearer "+token)
		if intent != "" {
			request.Header.Set("Idempotency-Key", intent)
		}
		response := httptest.NewRecorder()
		router.Engine.ServeHTTP(response, request)
		return response
	}
	intent := uuid.NewString()
	created := call("POST", "/api/v1/transactions", f.body(12500), key.Secret, intent)
	if created.Code != 201 {
		t.Fatalf("key create:%d", created.Code)
	}
	var row entity.Transaction
	if err := json.Unmarshal(decodeOfflineHTTP(t, created).Data, &row); err != nil {
		t.Fatal(err)
	}
	if row.OwnerID != f.owner {
		t.Fatal("body selected another owner")
	}
	if replay := call("POST", "/api/v1/transactions", f.body(12500), key.Secret, intent); replay.Code != 200 || !decodeOfflineHTTP(t, replay).Meta.Mutation.Replayed {
		t.Fatal("key receipt replay failed")
	}
	if denied := call("POST", "/api/v1/transactions", f.body(12500), otherKey.Secret, intent); denied.Code != 404 {
		t.Fatal("cross-owner receipt/wallet accepted")
	}
	if list := call("GET", "/api/v1/transactions", "", otherKey.Secret, ""); list.Code != 200 || bytes.Contains(list.Body.Bytes(), []byte(row.ID)) {
		t.Fatal("key read leaked foreign ledger")
	}
	pairBody := fmt.Sprintf(`{"source_wallet_id":%q,"destination_wallet_id":%q,"amount":100,"occurred_at":"2026-09-01T03:00:00Z"}`, f.wallet, f.destination)
	pairKey := uuid.NewString()
	for index := 0; index < 2; index++ {
		response := call("POST", "/api/v1/transactions/transfer", pairBody, key.Secret, pairKey)
		want := 201
		if index == 1 {
			want = 200
		}
		if response.Code != want {
			t.Fatalf("key transfer:%d", response.Code)
		}
	}
	var rows []entity.Transaction
	if err := f.db.Where("owner_id=?", f.owner).Find(&rows).Error; err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3 {
		t.Fatalf("ledger rows=%d", len(rows))
	}
	var transferRows int
	for _, stored := range rows {
		if stored.TransferID != nil {
			transferRows++
			if stored.IncludedInReports {
				t.Fatal("transfer included in reports")
			}
		}
	}
	if transferRows != 2 {
		t.Fatal("atomic pair missing")
	}
	child := call("POST", "/api/v1/api-keys", `{"name":"Child integration"}`, key.Secret, "")
	if child.Code != 201 {
		t.Fatalf("key cannot create another own key:%d", child.Code)
	}
	var envelope struct {
		Data struct {
			ID     string `json:"id"`
			Secret string `json:"secret"`
		}
	}
	if err := json.Unmarshal(child.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Data.ID == "" || envelope.Data.Secret == "" {
		t.Fatal("one-time key response missing")
	}
	metadata := call("GET", "/api/v1/api-keys", "", key.Secret, "")
	if metadata.Code != 200 || bytes.Contains(metadata.Body.Bytes(), []byte(otherKey.Key.ID)) || bytes.Contains(metadata.Body.Bytes(), []byte("secret")) || bytes.Contains(metadata.Body.Bytes(), []byte("lookup_id")) {
		t.Fatal("key metadata leaked another owner or credential secrets")
	}
	var listed struct{ Data []entity.UserAPIKey }
	if err := json.Unmarshal(metadata.Body.Bytes(), &listed); err != nil {
		t.Fatal(err)
	}
	if len(listed.Data) != 2 {
		t.Fatal("own key metadata missing")
	}
	for _, item := range listed.Data {
		if item.OwnerID != f.owner || item.Scopes == nil || len(item.Scopes) != 0 {
			t.Fatal("metadata advertised obsolete scope permissions")
		}
	}
	var persisted entity.UserAPIKey
	if err := f.db.Where("id=?", key.Key.ID).First(&persisted).Error; err != nil || len(persisted.Scopes) != 1 || persisted.Scopes[0] != "feedback:read" {
		t.Fatal("policy rollout must not rewrite existing key rows")
	}
	if foreign := call("DELETE", "/api/v1/api-keys/"+envelope.Data.ID, "", otherKey.Secret, ""); foreign.Code != 404 {
		t.Fatalf("foreign key must get not-found without revoking credential:%d", foreign.Code)
	}
	if removed := call("DELETE", "/api/v1/api-keys/"+envelope.Data.ID, "", key.Secret, ""); removed.Code != 204 {
		t.Fatalf("owner key revoke:%d", removed.Code)
	}
	if denied := call("GET", "/api/v1/wallets", "", envelope.Data.Secret, ""); denied.Code != 401 {
		t.Fatal("revoked child key accepted")
	}
	logout := call("POST", "/api/v1/auth/logout", `{}`, key.Secret, "")
	if logout.Code != 204 {
		t.Fatalf("key logout must revoke itself:%d", logout.Code)
	}
	if denied := call("GET", "/api/v1/sync/capabilities", "", key.Secret, ""); denied.Code != 401 {
		t.Fatal("logged-out key still active")
	}
	if own := call("GET", "/api/v1/wallets", "", otherKey.Secret, ""); own.Code != 200 {
		t.Fatal("self logout affected another owner")
	}
}
