package httpapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"mypocket/internal/identity"
	"mypocket/internal/platform/httpapi"
	mysync "mypocket/internal/sync"
)

func TestSyncMutationsRequiresCSRF(t *testing.T) {
	handler := httpapi.NewRouter(authTestConfig(), httpapi.Dependencies{
		IdentityRepository: &authRepoStub{user: identity.User{ID: "user_123", Email: "a@example.com", EmailVerified: true}},
		SyncService:        &syncServiceStub{},
	})
	req := authenticatedRequest(t, http.MethodPost, "/api/v1/sync/mutations", `{"mutations":[]}`)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", res.Code, res.Body.String())
	}
	if !strings.Contains(res.Body.String(), "CSRF_REQUIRED") {
		t.Fatalf("expected csrf error, got %s", res.Body.String())
	}
}

func TestSyncMutationsUsesAuthenticatedUser(t *testing.T) {
	service := &syncServiceStub{results: []mysync.MutationResult{{MutationID: "mut_1", EntityType: mysync.EntityWallet, EntityID: "wallet_1", Operation: mysync.OperationCreate, State: mysync.ResultApplied, Version: 1}}}
	handler := httpapi.NewRouter(authTestConfig(), httpapi.Dependencies{
		IdentityRepository: &authRepoStub{user: identity.User{ID: "user_123", Email: "a@example.com", EmailVerified: true}},
		SyncService:        service,
	})
	req := authenticatedRequest(t, http.MethodPost, "/api/v1/sync/mutations", `{"mutations":[{"mutation_id":"mut_1","device_id":"device_1","sequence":1,"entity_type":"wallet","entity_id":"wallet_1","operation":"create","base_version":0,"payload":{"name":"Cash","type":"cash"}}]}`)
	addCSRF(req)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	if service.applyUserID != "user_123" || len(service.appliedMutations) != 1 {
		t.Fatalf("sync mutations not scoped to auth user: user=%q mutations=%#v", service.applyUserID, service.appliedMutations)
	}
	if !strings.Contains(res.Body.String(), `"state":"applied"`) {
		t.Fatalf("unexpected sync response: %s", res.Body.String())
	}
}

func TestSyncChangesAndResyncUseAuthenticatedUser(t *testing.T) {
	service := &syncServiceStub{changes: mysync.ChangesResult{NextCursor: 7, Changes: []mysync.Change{{Cursor: 7, EntityType: mysync.EntityTransaction, EntityID: "tx_1", Operation: mysync.OperationUpdate, Version: 2}}}}
	handler := httpapi.NewRouter(authTestConfig(), httpapi.Dependencies{
		IdentityRepository: &authRepoStub{user: identity.User{ID: "user_123", Email: "a@example.com", EmailVerified: true}},
		SyncService:        service,
	})

	changesReq := authenticatedRequest(t, http.MethodGet, "/api/v1/sync/changes?after=6&limit=25", "")
	changesRes := httptest.NewRecorder()
	handler.ServeHTTP(changesRes, changesReq)
	if changesRes.Code != http.StatusOK {
		t.Fatalf("expected changes 200, got %d: %s", changesRes.Code, changesRes.Body.String())
	}
	if service.changesUserID != "user_123" || service.after != 6 || service.limit != 25 {
		t.Fatalf("changes not scoped/parsed: user=%q after=%d limit=%d", service.changesUserID, service.after, service.limit)
	}

	resyncReq := authenticatedRequest(t, http.MethodPost, "/api/v1/sync/resync", "")
	addCSRF(resyncReq)
	resyncRes := httptest.NewRecorder()
	handler.ServeHTTP(resyncRes, resyncReq)
	if resyncRes.Code != http.StatusOK {
		t.Fatalf("expected resync 200, got %d: %s", resyncRes.Code, resyncRes.Body.String())
	}
	if service.resyncUserID != "user_123" {
		t.Fatalf("resync not scoped to auth user: %q", service.resyncUserID)
	}
}

type syncServiceStub struct {
	applyUserID      string
	appliedMutations []mysync.Mutation
	results          []mysync.MutationResult
	changesUserID    string
	after            int64
	limit            int
	changes          mysync.ChangesResult
	resyncUserID     string
}

func (s *syncServiceStub) ApplyMutations(_ context.Context, userID string, mutations []mysync.Mutation) ([]mysync.MutationResult, error) {
	s.applyUserID = userID
	s.appliedMutations = mutations
	return s.results, nil
}

func (s *syncServiceStub) Changes(_ context.Context, userID string, after int64, limit int) (mysync.ChangesResult, error) {
	s.changesUserID = userID
	s.after = after
	s.limit = limit
	return s.changes, nil
}

func (s *syncServiceStub) Resync(_ context.Context, userID string) (mysync.Snapshot, error) {
	s.resyncUserID = userID
	return mysync.Snapshot{NextCursor: 7}, nil
}
