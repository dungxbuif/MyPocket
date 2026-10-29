package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/infrastructure/auth"
	redisadapter "github.com/mypocket/backend/internal/infrastructure/cache"
	postgresrepo "github.com/mypocket/backend/internal/infrastructure/repository"
	"github.com/mypocket/backend/internal/repository"
	"github.com/mypocket/backend/internal/usecase"
)

// All resources are synthetic, owner scoped, and deleted in reverse dependency
// order. Explicit loopback checks prevent accidentally running against a live URL.
type mcpParityFixture struct {
	offlineHTTPFixture
	foreign             offlineHTTPFixture
	router              *Router
	server              *httptest.Server
	keys                *usecase.UserAPIKeyService
	key, foreignKey     usecase.APIKeyCreated
	jwt, foreignJWT     string
	auth                *usecase.AuthInteractor
	cache               *offlineAuthCache
	tokens              *auth.JWT
	canonicalDispatches atomic.Int64
}

func requireMCPLocalURL(t *testing.T, name string) string {
	t.Helper()
	raw := os.Getenv(name)
	if raw == "" {
		t.Skip("requires disposable local " + name)
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Hostname() != "127.0.0.1" && u.Hostname() != "localhost" && u.Hostname() != "::1") {
		t.Fatal(name + " must identify a disposable loopback service")
	}
	return raw
}

func newMCPParityFixture(t *testing.T) *mcpParityFixture {
	t.Helper()
	requireMCPLocalURL(t, "TEST_DATABASE_URL")
	redisURL := requireMCPLocalURL(t, "TEST_REDIS_URL")
	f := &mcpParityFixture{offlineHTTPFixture: newOfflineHTTPFixture(t), foreign: newOfflineHTTPFixture(t)}
	// Cleanup runs before the underlying offline fixtures remove wallets/users.
	t.Cleanup(func() {
		owners := []string{f.owner, f.foreign.owner}
		for _, model := range []any{&entity.Transaction{}, &entity.TransactionMutation{}, &entity.Budget{}, &entity.JarMonthConfig{}, &entity.JarMonth{}, &entity.Jar{}, &entity.RecurringSchedule{}, &entity.PortfolioTrade{}, &entity.PortfolioAsset{}, &entity.UserAPIKey{}} {
			if err := f.db.Where("owner_id IN ?", owners).Delete(model).Error; err != nil {
				t.Error(err)
			}
		}
		if err := f.db.Where("user_id IN ?", owners).Delete(&entity.Feedback{}).Error; err != nil {
			t.Error(err)
		}
		if err := f.db.Exec("DELETE FROM category_wallets WHERE category_id IN (SELECT id FROM categories WHERE owner_id IN ?)", owners).Error; err != nil {
			t.Error(err)
		}
		if err := f.db.Where("owner_id IN ?", owners).Delete(&entity.Category{}).Error; err != nil {
			t.Error(err)
		}
	})
	network, err := redisadapter.NewRedis(redisURL)
	if err != nil {
		t.Fatal("cannot connect to disposable Redis")
	}
	f.cache = &offlineAuthCache{inner: network, prefix: "test:mcp-parity:" + uuid.NewString() + ":", keys: map[string]struct{}{}}
	t.Cleanup(func() {
		f.cache.mu.Lock()
		defer f.cache.mu.Unlock()
		for key := range f.cache.keys {
			if err := network.Delete(key); err != nil {
				t.Error(err)
			}
		}
		if err := network.Close(); err != nil {
			t.Error(err)
		}
	})
	passwords := auth.NewPasswordService()
	hash, err := passwords.Hash("mcp-synthetic-fixture-password")
	if err != nil {
		t.Fatal(err)
	}
	for _, owner := range []string{f.owner, f.foreign.owner} {
		if err := f.db.Model(&entity.User{}).Where("id = ?", owner).Update("password_hash", hash).Error; err != nil {
			t.Fatal(err)
		}
	}
	f.tokens = auth.NewJWT("mcp-synthetic-test-secret-at-least-thirty-two-bytes", time.Hour)
	f.auth = usecase.NewAuthInteractor(postgresrepo.NewUserPostgresRepository(f.db), passwords, f.tokens, f.cache, nil)
	for _, pair := range []struct {
		owner string
		token *string
	}{{f.owner, &f.jwt}, {f.foreign.owner, &f.foreignJWT}} {
		login, err := f.auth.Login(context.Background(), usecase.LoginInput{Email: pair.owner + "@offline-http.test", Password: "mcp-synthetic-fixture-password"})
		if err != nil {
			t.Fatal("fixture login failed")
		}
		*pair.token = login.Token
	}
	f.keys = usecase.NewUserAPIKeyService(postgresrepo.NewUserAPIKeyPostgresRepository(f.db))
	f.key, err = f.keys.Create(context.Background(), usecase.APIKeyCreateInput{OwnerID: f.owner, Name: "MCP synthetic owner"})
	if err != nil {
		t.Fatal(err)
	}
	f.foreignKey, err = f.keys.Create(context.Background(), usecase.APIKeyCreateInput{OwnerID: f.foreign.owner, Name: "MCP synthetic other"})
	if err != nil {
		t.Fatal(err)
	}
	middleware := NewAuthMiddleware(f.tokens, func(id string) (string, error) { return f.auth.VerifySession(context.Background(), id) })
	middleware.APIKeys = f.keys
	wallets, transactions, categories, users := postgresrepo.NewWalletPostgresRepository(f.db), postgresrepo.NewTransactionPostgresRepository(f.db), postgresrepo.NewCategoryPostgresRepository(f.db), postgresrepo.NewUserPostgresRepository(f.db)
	txHandler := NewTransactionHandler(transactions, wallets, categories)
	txHandler.Users, txHandler.Jars, txHandler.Travel = users, postgresrepo.NewJarPostgresRepository(f.db), postgresrepo.NewTravelPostgresRepository(f.db)
	gin.SetMode(gin.TestMode)
	f.router = NewRouter(NewAuthHandler(f.auth, false, nil), nil, nil, NewCategoryHandler(usecase.NewCategoryInteractor(categories)), NewWalletHandler(wallets), txHandler, middleware, nil)
	f.router.RegisterAPIKeyRoutes(&APIKeyHandler{Service: f.keys})
	f.router.RegisterBudgetRoutes(&BudgetHandler{Budgets: postgresrepo.NewBudgetPostgresRepository(f.db), Wallets: wallets, Categories: categories, Transactions: transactions, Users: users})
	f.router.RegisterJarRoutes(&JarHandler{Jars: postgresrepo.NewJarPostgresRepository(f.db), Users: users})
	f.router.RegisterRecurringRoutes(&RecurringHandler{Schedules: postgresrepo.NewRecurringPostgresRepository(f.db), Wallets: wallets, Categories: categories, Users: users})
	f.router.RegisterTravelRoutes(&TravelHandler{Events: postgresrepo.NewTravelPostgresRepository(f.db)})
	f.router.RegisterPortfolioRoutes(&PortfolioHandler{Portfolio: postgresrepo.NewPortfolioPostgresRepository(f.db)})
	f.router.RegisterCreditRoutes(&CreditHandler{Credits: transactions.(repository.CreditRepository), Wallets: wallets, Categories: categories})
	feedback := &FeedbackHandler{Service: usecase.NewFeedbackService(postgresrepo.NewFeedbackPostgresRepository(f.db), nil, nil)}
	// Mount only user feedback routes: service-token routes are outside this fixture.
	group := f.router.Engine.Group("/api/v1/feedback", middleware.RequireAuth)
	group.GET("", feedback.List)
	group.POST("", feedback.Create)
	group.GET("/:id", feedback.Get)
	group.PATCH("/:id/resolution", feedback.SetResolution)
	dispatcher := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.canonicalDispatches.Add(1)
		f.router.Engine.ServeHTTP(w, r)
	})
	handler := NewMCPHandler(DefaultExternalOperationCatalog(), dispatcher, MCPHandlerOptions{})
	f.router.Engine.Any("/api/v1/mcp", middleware.RequireAuth, gin.WrapH(handler))
	f.server = httptest.NewServer(f.router.Engine)
	t.Cleanup(f.server.Close)
	return f
}

type mcpParityLedgerSnapshot struct {
	wallets      []entity.Wallet
	transactions []entity.Transaction
}

func (f *mcpParityFixture) ledgerSnapshot(t *testing.T) mcpParityLedgerSnapshot {
	t.Helper()
	var snapshot mcpParityLedgerSnapshot
	owners := []string{f.owner, f.foreign.owner}
	if err := f.db.Where("owner_id IN ?", owners).Order("id ASC").Find(&snapshot.wallets).Error; err != nil {
		t.Fatal(err)
	}
	if err := f.db.Where("owner_id IN ?", owners).Order("id ASC").Find(&snapshot.transactions).Error; err != nil {
		t.Fatal(err)
	}
	return snapshot
}

type mcpParityBearerTransport struct {
	token string
	base  http.RoundTripper
	probe *mcpParityHTTPProbe
}

// Observe the same HTTP response the SDK consumes. The SDK intentionally
// summarizes HTTP auth failures as transport errors rather than retaining our
// Problem object, so an arbitrary SDK error alone is insufficient evidence.
type mcpParityHTTPProbe struct {
	mu     sync.Mutex
	status int
	code   string
}

func (p *mcpParityHTTPProbe) reset() { p.mu.Lock(); defer p.mu.Unlock(); p.status, p.code = 0, "" }
func (p *mcpParityHTTPProbe) snapshot() (int, string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.status, p.code
}

func (r mcpParityBearerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.Header = req.Header.Clone()
	if r.token != "" {
		clone.Header.Set("Authorization", "Bearer "+r.token)
	}
	response, err := r.base.RoundTrip(clone)
	if err == nil && r.probe != nil {
		code := ""
		if response.StatusCode >= 400 && response.Body != nil {
			original := response.Body
			prefix, readErr := io.ReadAll(io.LimitReader(original, 8192))
			response.Body = struct {
				io.Reader
				io.Closer
			}{io.MultiReader(bytes.NewReader(prefix), original), original}
			if readErr == nil {
				var problem struct {
					Code string `json:"code"`
				}
				if json.Unmarshal(prefix, &problem) == nil {
					code = problem.Code
				}
			}
		}
		r.probe.mu.Lock()
		r.probe.status, r.probe.code = response.StatusCode, code
		r.probe.mu.Unlock()
	}
	return response, err
}

func (f *mcpParityFixture) connect(t *testing.T, token string) (*mcp.ClientSession, error) {
	t.Helper()
	session, _, err := f.connectObserved(t, token)
	return session, err
}

func (f *mcpParityFixture) connectObserved(t *testing.T, token string) (*mcp.ClientSession, *mcpParityHTTPProbe, error) {
	t.Helper()
	probe := &mcpParityHTTPProbe{}
	client := mcp.NewClient(&mcp.Implementation{Name: "mypocket-parity-tests", Version: "1"}, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	session, err := client.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: f.server.URL + "/api/v1/mcp", HTTPClient: &http.Client{Transport: mcpParityBearerTransport{token: token, base: http.DefaultTransport, probe: probe}}, DisableStandaloneSSE: true, MaxRetries: -1}, nil)
	if err == nil {
		t.Cleanup(func() { _ = session.Close() })
	}
	return session, probe, err
}

func requireMCPParityAuthRejected(t *testing.T, probe *mcpParityHTTPProbe, result *mcp.CallToolResult, err error, wantCode string) {
	t.Helper()
	if err == nil && (result == nil || !result.IsError) {
		t.Fatal("invalid bearer was accepted by SDK")
	}
	status, code := probe.snapshot()
	if status != http.StatusUnauthorized || code != wantCode {
		t.Fatalf("SDK auth failure HTTP=%d code=%q want=401/%s; SDK error=%v", status, code, wantCode, err)
	}
}

func mcpParityCallAfterRevocation(t *testing.T, session *mcp.ClientSession, probe *mcpParityHTTPProbe, wantCode string) {
	t.Helper()
	probe.reset()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "wallets.list", Arguments: map[string]any{}})
	requireMCPParityAuthRejected(t, probe, result, err, wantCode)
}

func requireMCPParityEqualReadback(t *testing.T, mcpBody, restBody []byte) {
	t.Helper()
	var left, right struct {
		Data any `json:"data"`
	}
	if err := json.Unmarshal(mcpBody, &left); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(restBody, &right); err != nil {
		t.Fatal(err)
	}
	l, err := json.Marshal(left.Data)
	if err != nil {
		t.Fatal(err)
	}
	r, err := json.Marshal(right.Data)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(l, r) {
		t.Fatalf("MCP/REST readback differs: MCP=%s REST=%s", l, r)
	}
}

func (f *mcpParityFixture) session(t *testing.T, token string) *mcp.ClientSession {
	t.Helper()
	session, err := f.connect(t, token)
	if err != nil {
		t.Fatalf("SDK initialize failed: %v", err)
	}
	return session
}

type mcpParityResponse struct {
	Status    int             `json:"status"`
	Body      json.RawMessage `json:"body"`
	RequestID string          `json:"request_id"`
}

func mcpParityCall(t *testing.T, session *mcp.ClientSession, name string, args map[string]any, key string) (*mcp.CallToolResult, mcpParityResponse) {
	t.Helper()
	params := &mcp.CallToolParams{Name: name, Arguments: args}
	if key != "" {
		params.Meta = mcp.Meta{"idempotency_key": key}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, err := session.CallTool(ctx, params)
	if err != nil {
		t.Fatalf("%s SDK call: %v", name, err)
	}
	raw, err := json.Marshal(result.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}
	var response mcpParityResponse
	if err := json.Unmarshal(raw, &response); err != nil {
		t.Fatalf("%s structured response: %v", name, err)
	}
	return result, response
}

func requireMCPParityStatus(t *testing.T, result *mcp.CallToolResult, response mcpParityResponse, want int) {
	t.Helper()
	if response.Status != want {
		t.Fatalf("MCP HTTP status=%d want=%d body=%s content=%+v", response.Status, want, response.Body, result.Content)
	}
	if result.IsError != (want >= 400) {
		t.Fatalf("IsError=%v for status=%d", result.IsError, want)
	}
}

func mcpParityData[T any](t *testing.T, response mcpParityResponse) T {
	t.Helper()
	var envelope struct {
		Data T `json:"data"`
	}
	if err := json.Unmarshal(response.Body, &envelope); err != nil {
		t.Fatal(err)
	}
	return envelope.Data
}

func (f *mcpParityFixture) rest(t *testing.T, method, path string, body any, token, key string) *httptest.ResponseRecorder {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(method, path, bytes.NewReader(raw))
	request.Header.Set("Content-Type", "application/json")
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	if key != "" {
		request.Header.Set("Idempotency-Key", key)
	}
	response := httptest.NewRecorder()
	f.router.Engine.ServeHTTP(response, request)
	return response
}

func (f *mcpParityFixture) transactionArgs(amount int64) map[string]any {
	return map[string]any{"wallet_id": f.wallet, "category_id": f.category, "type": "expense", "amount": amount, "occurred_at": "2026-10-20T03:00:00Z", "auto_link_travel": false}
}
