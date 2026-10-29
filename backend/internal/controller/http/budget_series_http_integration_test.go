package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/infrastructure/auth"
	postgresrepo "github.com/mypocket/backend/internal/infrastructure/repository"
	"github.com/mypocket/backend/internal/usecase"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestBudgetSeriesRealHTTPJWTAPIKeyIsolationAndConcurrentList(t *testing.T) {
	f, foreign := newOfflineHTTPFixture(t), newOfflineHTTPFixture(t)
	keys := usecase.NewUserAPIKeyService(postgresrepo.NewUserAPIKeyPostgresRepository(f.db))
	key, err := keys.Create(context.Background(), usecase.APIKeyCreateInput{OwnerID: f.owner, Name: "recurrence"})
	if err != nil {
		t.Fatal(err)
	}
	otherKey, err := keys.Create(context.Background(), usecase.APIKeyCreateInput{OwnerID: foreign.owner, Name: "foreign"})
	if err != nil {
		t.Fatal(err)
	}
	tokens := auth.NewJWT("budget-series-secret-at-least-thirty-two-bytes", time.Hour)
	jwt, _, err := tokens.Generate(f.owner, f.owner+"@test.invalid", "Owner", "session")
	if err != nil {
		t.Fatal(err)
	}
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	router := &Router{Engine: engine, AuthMiddleware: &AuthMiddleware{APIKeys: keys, TokenSvc: tokens, VerifySession: func(string) (string, error) { return f.owner, nil }}}
	series := postgresrepo.NewBudgetSeriesPostgresRepository(f.db)
	users := postgresrepo.NewUserPostgresRepository(f.db)
	handler := &BudgetSeriesHandler{Series: series, Users: users, Now: func() time.Time { return time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC) }}
	router.RegisterBudgetSeriesRoutes(handler)
	router.RegisterBudgetRoutes(&BudgetHandler{Series: series, Users: users, Budgets: postgresrepo.NewBudgetPostgresRepository(f.db), Wallets: postgresrepo.NewWalletPostgresRepository(f.db), Categories: postgresrepo.NewCategoryPostgresRepository(f.db), Transactions: postgresrepo.NewTransactionPostgresRepository(f.db)})
	call := func(method, path, body, token, intent string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		if intent != "" {
			req.Header.Set("Idempotency-Key", intent)
		}
		res := httptest.NewRecorder()
		engine.ServeHTTP(res, req)
		return res
	}
	body := fmt.Sprintf(`{"name":"Recurring","limit_amount":1000,"cadence":"monthly","start_date":"2026-10-01","wallet_id":%q,"category_id":%q}`, f.wallet, f.category)
	created := call("POST", "/api/v1/budget-series", body, jwt, "create")
	if created.Code != 201 {
		t.Fatalf("create %d %s", created.Code, created.Body.String())
	}
	var envelope struct {
		Data entity.BudgetSeries `json:"data"`
	}
	if err = json.Unmarshal(created.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if replay := call("POST", "/api/v1/budget-series", body, key.Secret, "create"); replay.Code != 200 {
		t.Fatalf("key replay %d %s", replay.Code, replay.Body.String())
	}
	instances, err := postgresrepo.NewBudgetPostgresRepository(f.db).List(f.owner)
	if err != nil || len(instances) != 1 {
		t.Fatalf("instances %+v %v", instances, err)
	}
	patch := fmt.Sprintf(`{"name":"direct edit","limit_amount":500,"wallet_id":%q,"category_id":%q,"start_date":"2026-10-01","end_date":"2026-10-31"}`, f.wallet, f.category)
	for _, tc := range []struct{ method, body string }{{"PATCH", patch}, {"DELETE", ""}} {
		if res := call(tc.method, "/api/v1/budgets/"+instances[0].ID, tc.body, key.Secret, ""); res.Code != 409 || !bytes.Contains(res.Body.Bytes(), []byte("BUDGET_SERIES_MANAGED")) {
			t.Fatalf("generated %s %d %s", tc.method, res.Code, res.Body.String())
		}
	}
	for _, token := range []string{jwt, key.Secret} {
		res := call("GET", "/api/v1/budget-series", "", token, "")
		if res.Code != 200 || !bytes.Contains(res.Body.Bytes(), []byte(envelope.Data.ID)) {
			t.Fatalf("owned list %d %s", res.Code, res.Body.String())
		}
	}
	if res := call("GET", "/api/v1/budget-series", "", otherKey.Secret, ""); res.Code != 200 || bytes.Contains(res.Body.Bytes(), []byte(envelope.Data.ID)) {
		t.Fatalf("foreign read %d %s", res.Code, res.Body.String())
	}
	if res := call("POST", "/api/v1/budget-series/"+envelope.Data.ID+"/stop", `{"version":1}`, otherKey.Secret, "stop"); res.Code != 404 {
		t.Fatalf("foreign stop %d %s", res.Code, res.Body.String())
	}
	if res := call("GET", "/api/v1/budget-series", "", "", ""); res.Code != 401 {
		t.Fatalf("unauthorized %d", res.Code)
	}
	handler.Now = func() time.Time { return time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC) }
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			path := "/api/v1/budget-series"
			if i%2 == 0 {
				path = "/api/v1/budgets"
			}
			if res := call("GET", path, "", key.Secret, ""); res.Code != 200 {
				t.Errorf("concurrent list %d %s", res.Code, res.Body.String())
			}
		}(i)
	}
	wg.Wait()
	var count int64
	if err = f.db.Model(&entity.Budget{}).Where("owner_id = ?", f.owner).Count(&count).Error; err != nil || count != 2 {
		t.Fatalf("duplicate periods %d %v", count, err)
	}
	handler.Now = func() time.Time { return time.Date(2027, 2, 16, 12, 0, 0, 0, time.UTC) }
	if res := call("POST", "/api/v1/budget-series", body, key.Secret, "create"); res.Code != 200 {
		t.Fatalf("late replay %d %s", res.Code, res.Body.String())
	}
}
