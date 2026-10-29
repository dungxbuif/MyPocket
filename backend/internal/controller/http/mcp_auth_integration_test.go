package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/infrastructure/auth"
	postgresrepo "github.com/mypocket/backend/internal/infrastructure/repository"
	"github.com/mypocket/backend/internal/usecase"
)

// Removing canonical bearer propagation would either deny an authorized owner
// or expose the other owner's rows; literal owner IDs determine expectations.
func TestMCPAuthJWTAPIKeySameOwnerParity(t *testing.T) {
	f := newMCPParityFixture(t)
	// A legacy limited-scope row must retain the same owner authority.
	if err := f.db.Model(&entity.UserAPIKey{}).Where("id = ?", f.key.Key.ID).Update("scopes", `["feedback:read"]`).Error; err != nil {
		t.Fatal(err)
	}
	for _, token := range []string{f.jwt, f.key.Secret} {
		session := f.session(t, token)
		result, response := mcpParityCall(t, session, "wallets.list", map[string]any{}, "")
		requireMCPParityStatus(t, result, response, 200)
		wallets := mcpParityData[[]entity.Wallet](t, response)
		if len(wallets) != 2 {
			t.Fatalf("wallets=%d want=2", len(wallets))
		}
		ownedIDs := map[string]bool{}
		for _, wallet := range wallets {
			if wallet.OwnerID != f.owner || wallet.ID == f.foreign.wallet {
				t.Fatal("MCP exposed foreign owner wallet")
			}
			ownedIDs[wallet.ID] = true
		}
		if !ownedIDs[f.wallet] || !ownedIDs[f.destination] {
			t.Fatal("MCP omitted an owned wallet")
		}
		rest := f.rest(t, "GET", "/api/v1/wallets", nil, token, "")
		if rest.Code != 200 {
			t.Fatalf("REST status=%d", rest.Code)
		}
		var envelope struct {
			Data []entity.Wallet `json:"data"`
		}
		if err := json.Unmarshal(rest.Body.Bytes(), &envelope); err != nil {
			t.Fatal(err)
		}
		requireMCPParityEqualReadback(t, response.Body, rest.Body.Bytes())
	}
}

type mcpParityAdvisorProviderFunc func(context.Context, usecase.AdvisorRequest) (usecase.AdvisorResponse, error)

func (f mcpParityAdvisorProviderFunc) Chat(ctx context.Context, request usecase.AdvisorRequest) (usecase.AdvisorResponse, error) {
	return f(ctx, request)
}

type mcpParityFinanceReaderFunc func(context.Context, string, []entity.NormalizedQuery) (entity.FactBundle, error)

func (f mcpParityFinanceReaderFunc) ReadBundle(ctx context.Context, owner string, queries []entity.NormalizedQuery) (entity.FactBundle, error) {
	return f(ctx, owner, queries)
}

// Removing any revalidation boundary must permit observable provider/tool work
// or an assistant message after revocation and fail this test. Only the provider
// is synthetic; the orchestrator, tool registry, finance reader, keys and run/
// message/event persistence are the canonical implementations.
func TestMCPAuthAdvisorAsyncRevocationStopsProviderToolAndPersistence(t *testing.T) {
	for _, tc := range []struct {
		boundary                string
		blockAt                 int32
		withTool                bool
		wantProvider, wantReads int32
	}{
		{"before_provider", 2, false, 0, 0},
		{"before_tool", 3, true, 1, 0},
		{"before_next_provider", 4, true, 1, 1},
		{"before_answer_persistence", 3, false, 1, 0},
	} {
		t.Run(tc.boundary, func(t *testing.T) {
			f := newMCPParityFixture(t)
			store := postgresrepo.NewAdvisorPostgresRepository(f.db)
			reader := postgresrepo.NewFinanceQueryPostgresRepository(f.db)
			var providerCalls, reads, validations atomic.Int32
			finance := usecase.NewFinanceQueryService(mcpParityFinanceReaderFunc(func(ctx context.Context, owner string, queries []entity.NormalizedQuery) (entity.FactBundle, error) {
				reads.Add(1)
				return reader.ReadBundle(ctx, owner, queries)
			}))
			provider := mcpParityAdvisorProviderFunc(func(context.Context, usecase.AdvisorRequest) (usecase.AdvisorResponse, error) {
				call := providerCalls.Add(1)
				if tc.withTool && call == 1 {
					return usecase.AdvisorResponse{ToolCalls: []usecase.AdvisorToolCall{{ID: "owned-wallet-read", Name: "get_wallet_balances", Arguments: json.RawMessage(`{}`)}}}, nil
				}
				return usecase.AdvisorResponse{Text: "Synthetic answer that must not be saved after revocation"}, nil
			})
			orchestrator := usecase.NewAdvisorOrchestrator(provider, usecase.NewAdvisorToolRegistry(finance))
			blocked := make(chan usecase.Principal, 1)
			release := make(chan struct{})
			var releaseOnce sync.Once
			unblock := func() { releaseOnce.Do(func() { close(release) }) }
			defer unblock()
			orchestrator.Validator = usecase.AdvisorPrincipalValidatorFunc(func(ctx context.Context, principal usecase.Principal) error {
				if validations.Add(1) == tc.blockAt {
					blocked <- principal
					select {
					case <-release:
					case <-ctx.Done():
						return ctx.Err()
					}
				}
				return f.keys.ValidatePrincipal(ctx, principal)
			})
			service := usecase.NewAdvisorService(store, orchestrator)
			f.router.RegisterAdvisorRoutes(&AdvisorHandler{Service: service, Users: postgresrepo.NewUserPostgresRepository(f.db), Finance: finance})
			result, response := mcpParityCall(t, f.session(t, f.key.Secret), "advisor.messages_async", map[string]any{"text": "Read my wallets", "client_request_id": uuid.NewString()}, "")
			requireMCPParityStatus(t, result, response, 202)
			handle := mcpParityData[advisorAsyncSubmitResponse](t, response)
			if handle.RunID == "" || handle.ConversationID == "" {
				t.Fatal("async MCP did not return persisted run identity")
			}
			// A failed assertion must release and drain the synthetic run before
			// the shared fixture deletes its owner or closes the DB connection.
			t.Cleanup(func() {
				unblock()
				deadline := time.Now().Add(5 * time.Second)
				for time.Now().Before(deadline) {
					run, err := store.GetRun(context.Background(), f.owner, handle.RunID)
					if err != nil || (run.Status != "queued" && run.Status != "running") {
						return
					}
					time.Sleep(time.Millisecond)
				}
				t.Error("async fixture did not terminate before cleanup")
			})
			select {
			case principal := <-blocked:
				if principal.OwnerID != f.owner || principal.CredentialID != f.key.Key.ID || principal.CredentialKind != "user_api_key" {
					t.Fatalf("async principal=%+v", principal)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("async run did not reach the credential revalidation boundary")
			}
			if err := f.keys.Revoke(context.Background(), f.owner, f.key.Key.ID); err != nil {
				t.Fatal(err)
			}
			unblock()
			deadline := time.Now().Add(5 * time.Second)
			var finished *entity.AdvisorRun
			for {
				run, err := store.GetRun(context.Background(), f.owner, handle.RunID)
				if err != nil {
					t.Fatal(err)
				}
				if run.Status != "running" && run.Status != "queued" {
					finished = run
					break
				}
				if time.Now().After(deadline) {
					t.Fatal("revoked async run did not reach terminal state")
				}
				time.Sleep(time.Millisecond)
			}
			if finished.Status != "failed" {
				t.Fatalf("revoked run status=%s want=failed", finished.Status)
			}
			if providerCalls.Load() != tc.wantProvider || reads.Load() != tc.wantReads {
				t.Fatalf("work crossed revoked boundary: provider=%d want=%d reads=%d want=%d", providerCalls.Load(), tc.wantProvider, reads.Load(), tc.wantReads)
			}
			messages, err := store.ListMessages(context.Background(), f.owner, handle.ConversationID, 0, 50)
			if err != nil {
				t.Fatal(err)
			}
			if len(messages) != 1 || messages[0].Role != "user" {
				t.Fatalf("revoked run persisted assistant content: %+v", messages)
			}
			var completionEvents int64
			if err := f.db.Model(&entity.AdvisorEvent{}).Where("run_id = ? AND type = ?", handle.RunID, "run.completed").Count(&completionEvents).Error; err != nil || completionEvents != 0 {
				t.Fatalf("revoked run completion events=%d err=%v", completionEvents, err)
			}
			result, response = mcpParityCall(t, f.session(t, f.jwt), "advisor.run.get", map[string]any{"id": handle.RunID}, "")
			requireMCPParityStatus(t, result, response, 200)
			if run := mcpParityData[entity.AdvisorRun](t, response); run.Status != "failed" || run.CredentialKind != "user_api_key" {
				t.Fatalf("owner readback run=%+v", run)
			}
		})
	}
}

func TestMCPAuthRejectsMalformedExpiredRevokedKeysAndSessions(t *testing.T) {
	f := newMCPParityFixture(t)
	expiry := time.Now().Add(time.Hour)
	expiring, err := f.keys.Create(context.Background(), usecase.APIKeyCreateInput{OwnerID: f.owner, Name: "Expiry proof", ExpiresAt: &expiry})
	if err != nil {
		t.Fatal(err)
	}
	if err := f.db.Model(&entity.UserAPIKey{}).Where("id = ?", expiring.Key.ID).Update("expires_at", time.Now().Add(-time.Hour)).Error; err != nil {
		t.Fatal(err)
	}
	expiredJWT, _, err := auth.NewJWT("mcp-synthetic-test-secret-at-least-thirty-two-bytes", -time.Hour).Generate(f.owner, "", "", "expired")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ name, token, code string }{
		{"missing", "", "AUTH_REQUIRED"}, {"malformed API key", "mpk_invalid", "TOKEN_INVALID"}, {"wrong digest", f.key.Secret[:len(f.key.Secret)-4] + "xxxx", "TOKEN_INVALID"}, {"expired API key", expiring.Secret, "TOKEN_INVALID"}, {"expired JWT", expiredJWT, "TOKEN_INVALID"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, probe, err := f.connectObserved(t, tc.token)
			requireMCPParityAuthRejected(t, probe, nil, err, tc.code)
			if got := f.rest(t, "GET", "/api/v1/sync/capabilities", nil, tc.token, ""); got.Code != 401 {
				t.Fatalf("canonical auth status=%d", got.Code)
			}
		})
	}
	// Reusing an already initialized SDK client must not cache authorization.
	keySession, keyProbe, err := f.connectObserved(t, f.key.Secret)
	if err != nil {
		t.Fatal(err)
	}
	result, authorized := mcpParityCall(t, keySession, "wallets.list", map[string]any{}, "")
	requireMCPParityStatus(t, result, authorized, 200)
	if err := f.keys.Revoke(context.Background(), f.owner, f.key.Key.ID); err != nil {
		t.Fatal(err)
	}
	mcpParityCallAfterRevocation(t, keySession, keyProbe, "TOKEN_INVALID")
	_, freshProbe, err := f.connectObserved(t, f.key.Secret)
	requireMCPParityAuthRejected(t, freshProbe, nil, err, "TOKEN_INVALID")
	jwtSession, jwtProbe, err := f.connectObserved(t, f.jwt)
	if err != nil {
		t.Fatal(err)
	}
	result, authorized = mcpParityCall(t, jwtSession, "wallets.list", map[string]any{}, "")
	requireMCPParityStatus(t, result, authorized, 200)
	claims, err := f.tokens.Verify(f.jwt)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.cache.Delete(f.auth.SessionPrefix + claims.SessionID); err != nil {
		t.Fatal(err)
	}
	mcpParityCallAfterRevocation(t, jwtSession, jwtProbe, "SESSION_EXPIRED")
	short, err := f.auth.Login(context.Background(), usecase.LoginInput{Email: f.owner + "@offline-http.test", Password: "mcp-synthetic-fixture-password"})
	if err != nil {
		t.Fatal(err)
	}
	shortSession, shortProbe, err := f.connectObserved(t, short.Token)
	if err != nil {
		t.Fatal(err)
	}
	result, authorized = mcpParityCall(t, shortSession, "wallets.list", map[string]any{}, "")
	requireMCPParityStatus(t, result, authorized, 200)
	shortClaims, err := f.tokens.Verify(short.Token)
	if err != nil {
		t.Fatal(err)
	}
	sessionKey := f.auth.SessionPrefix + shortClaims.SessionID
	value, found, err := f.cache.Get(sessionKey)
	if err != nil || !found {
		t.Fatal("fixture Redis session missing")
	}
	if err := f.cache.Set(sessionKey, value, time.Millisecond); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(time.Second)
	for {
		_, found, err = f.cache.Get(sessionKey)
		if err != nil {
			t.Fatal(err)
		}
		if !found {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("disposable Redis session did not expire")
		}
		time.Sleep(time.Millisecond)
	}
	mcpParityCallAfterRevocation(t, shortSession, shortProbe, "SESSION_EXPIRED")
	if result, response := mcpParityCall(t, f.session(t, f.foreignKey.Secret), "wallets.list", map[string]any{}, ""); response.Status != 200 || result.IsError {
		t.Fatal("revocation affected another owner")
	}
}

func TestMCPAuthLogoutRevokesCurrentAPIKey(t *testing.T) {
	f := newMCPParityFixture(t)
	session := f.session(t, f.key.Secret)
	result, response := mcpParityCall(t, session, "account.logout", map[string]any{}, "")
	requireMCPParityStatus(t, result, response, 204)
	if _, err := f.keys.Authenticate(context.Background(), f.key.Secret, nil); err == nil {
		t.Fatal("MCP logout did not revoke canonical API key")
	}
	if response := f.rest(t, "GET", "/api/v1/wallets", nil, f.key.Secret, ""); response.Code != 401 {
		t.Fatalf("logged out key REST status=%d", response.Code)
	}
	if response := f.rest(t, "GET", "/api/v1/wallets", nil, f.jwt, ""); response.Code != 200 {
		t.Fatal("API-key logout revoked independent JWT")
	}
}

func TestMCPOwnerForeignResourcePathsRemainNotFound(t *testing.T) {
	f := newMCPParityFixture(t)
	start, _ := entity.ParseCalendarDate("2026-10-01")
	end, _ := entity.ParseCalendarDate("2026-10-31")
	budget := entity.Budget{ID: uuid.NewString(), OwnerID: f.foreign.owner, Name: "Foreign budget", LimitAmount: 1000, StartDate: start, EndDate: end}
	feedback := entity.Feedback{ID: uuid.NewString(), UserID: f.foreign.owner, Type: "bug", Title: "Foreign feedback", Description: "Synthetic fixture", Status: "open"}
	transaction := entity.Transaction{ID: uuid.NewString(), OwnerID: f.foreign.owner, WalletID: f.foreign.wallet, CategoryID: &f.foreign.category, Type: "expense", Amount: 100, OccurredAt: time.Now()}
	for _, row := range []any{&budget, &feedback, &transaction} {
		if err := f.db.Create(row).Error; err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct {
		name, method, path string
		args               map[string]any
	}{
		{"wallets.update", "PATCH", "/api/v1/wallets/" + f.foreign.wallet, map[string]any{"id": f.foreign.wallet, "name": "Stolen", "type": "basic"}},
		{"categories.update", "PATCH", "/api/v1/categories/" + f.foreign.category, map[string]any{"id": f.foreign.category, "name": "Stolen", "kind": "expense", "icon_key": "tag", "expected_version": 1}},
		{"transactions.update", "PATCH", "/api/v1/transactions/" + transaction.ID, map[string]any{"id": transaction.ID, "wallet_id": f.wallet, "category_id": f.category, "type": "expense", "amount": 200, "expected_version": 1}},
		{"budgets.update", "PATCH", "/api/v1/budgets/" + budget.ID, map[string]any{"id": budget.ID, "name": "Stolen", "limit_amount": 200, "start_date": "2026-10-01", "end_date": "2026-10-31"}},
		{"feedback.get", "GET", "/api/v1/feedback/" + feedback.ID, map[string]any{"id": feedback.ID}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, token := range []string{f.jwt, f.key.Secret} {
				result, response := mcpParityCall(t, f.session(t, token), tc.name, tc.args, "")
				requireMCPParityStatus(t, result, response, 404)
				rest := f.rest(t, tc.method, tc.path, tc.args, token, "")
				if rest.Code != 404 {
					t.Fatalf("canonical foreign resource status=%d", rest.Code)
				}
			}
		})
	}
	var stored entity.Transaction
	if err := f.db.First(&stored, "id = ?", transaction.ID).Error; err != nil || stored.Amount != 100 || stored.OwnerID != f.foreign.owner {
		t.Fatal("foreign transaction was mutated")
	}
}

func TestMCPOwnerRejectsBodyQueryAndPathInjectionBeforeWrite(t *testing.T) {
	f := newMCPParityFixture(t)
	baseline := entity.Transaction{ID: uuid.NewString(), OwnerID: f.owner, WalletID: f.wallet, CategoryID: &f.category, Type: "expense", Amount: 375, OccurredAt: time.Date(2026, 10, 20, 3, 0, 0, 0, time.UTC)}
	if err := f.db.Create(&baseline).Error; err != nil {
		t.Fatal(err)
	}
	session, probe, err := f.connectObserved(t, f.key.Secret)
	if err != nil {
		t.Fatal(err)
	}
	// Positive control proves this observer sees real canonical dispatches,
	// while initialize itself never invokes a business route.
	if f.canonicalDispatches.Load() != 0 {
		t.Fatal("initialize dispatched a business operation")
	}
	result, authorized := mcpParityCall(t, session, "wallets.list", map[string]any{}, "")
	requireMCPParityStatus(t, result, authorized, 200)
	if f.canonicalDispatches.Load() != 1 {
		t.Fatal("dispatch observer missed the authorized control call")
	}
	before := f.ledgerSnapshot(t)
	for _, tc := range []struct {
		name, tool string
		args       map[string]any
	}{
		{"body owner", "wallets.create", map[string]any{"name": "Injected", "type": "basic", "owner_id": f.foreign.owner}},
		{"body user", "wallets.create", map[string]any{"name": "Injected", "type": "basic", "user_id": f.foreign.owner}},
		{"query owner", "wallets.list", map[string]any{"owner_id": f.foreign.owner}},
		{"nested query owner", "transactions.list", map[string]any{"query": map[string]any{"owner_id": f.foreign.owner}}},
		{"path traversal", "wallets.update", map[string]any{"id": f.wallet + "/../../wallets/" + f.foreign.wallet, "name": "Injected", "type": "basic"}},
		{"body retry key", "wallets.create", map[string]any{"name": "Injected", "type": "basic", "idempotency_key": "hidden-body-key"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dispatches := f.canonicalDispatches.Load()
			probe.reset()
			// mcpParityCall creates a fresh bounded context and rejects every
			// SDK/transport error; only the declared adapter Problem may pass.
			result, response := mcpParityCall(t, session, tc.tool, tc.args, "")
			requireMCPParityStatus(t, result, response, 400)
			transportStatus, _ := probe.snapshot()
			if transportStatus != 200 {
				t.Fatalf("SDK transport status=%d want=200", transportStatus)
			}
			var problem struct {
				Status int    `json:"status"`
				Code   string `json:"code"`
			}
			if err := json.Unmarshal(response.Body, &problem); err != nil {
				t.Fatal(err)
			}
			if problem.Status != 400 || problem.Code != "MCP_INVALID_ARGUMENTS" {
				t.Fatalf("adapter Problem=%+v want=400/MCP_INVALID_ARGUMENTS", problem)
			}
			if f.canonicalDispatches.Load() != dispatches {
				t.Fatal("rejected injection invoked a canonical REST operation")
			}
			if bytes.Contains(response.Body, []byte(f.foreign.wallet)) || strings.Contains(string(response.Body), "HTTP destination") {
				t.Fatal("selector rejection leaked foreign data")
			}
			after := f.ledgerSnapshot(t)
			if !reflect.DeepEqual(before.wallets, after.wallets) {
				t.Fatalf("injection changed pre-existing wallet fields: before=%+v after=%+v", before.wallets, after.wallets)
			}
			if !reflect.DeepEqual(before.transactions, after.transactions) {
				t.Fatalf("injection changed ledger state: before=%+v after=%+v", before.transactions, after.transactions)
			}
		})
	}
}
