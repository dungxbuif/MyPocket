package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/usecase"
)

type reportBudgetRepo struct {
	budgetRepoStub
	items []entity.Budget
	owner string
}

func (r *reportBudgetRepo) List(owner string) ([]entity.Budget, error) {
	r.owner = owner
	return r.items, nil
}

type reportTransactionsRepo struct {
	transactionRepositoryStub
	items []entity.Transaction
	owner string
}

func (r *reportTransactionsRepo) List(owner string) ([]entity.Transaction, error) {
	r.owner = owner
	return r.items, nil
}

func TestBudgetReportRoutesAPIKeyOwnerScopeAndDrilldown(t *testing.T) {
	gin.SetMode(gin.TestMode)
	owner, id := "owner-1", uuid.NewString()
	from, _ := entity.ParseCalendarDate("2026-09-01")
	to, _ := entity.ParseCalendarDate("2026-09-30")
	now := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	budgets := &reportBudgetRepo{items: []entity.Budget{{ID: id, OwnerID: owner, LimitAmount: 4000000, StartDate: from, EndDate: to}}}
	rows := &reportTransactionsRepo{items: []entity.Transaction{
		{ID: "b", OwnerID: owner, Type: "expense", Amount: 100000, IncludedInReports: true, OccurredAt: now},
		{ID: "a", OwnerID: owner, Type: "expense", Amount: 200000, IncludedInReports: true, OccurredAt: now.Add(-time.Hour)},
		{ID: "foreign", OwnerID: "other", Type: "expense", Amount: 500000, IncludedInReports: true, OccurredAt: now},
	}}
	h := &BudgetReportHandler{Budgets: budgets, Transactions: rows, Categories: &transactionCategoryRepositoryStub{}, Users: &transactionUserRepositoryStub{user: entity.User{ID: owner, Timezone: "UTC"}}, Now: func() time.Time { return now }}
	keyService := usecase.NewUserAPIKeyService(&middlewareAPIKeyRepo{})
	key, err := keyService.Create(context.Background(), usecase.APIKeyCreateInput{OwnerID: owner, Name: "budget test"})
	if err != nil {
		t.Fatal(err)
	}
	engine := gin.New()
	router := &Router{Engine: engine, AuthMiddleware: &AuthMiddleware{APIKeys: keyService}}
	router.RegisterBudgetReportRoutes(h)
	request := func(path, token string) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		res := httptest.NewRecorder()
		engine.ServeHTTP(res, req)
		return res
	}
	res := request("/api/v1/budgets/"+id+"/report", key.Secret)
	var report struct {
		Data entity.BudgetReport `json:"data"`
	}
	if err = json.Unmarshal(res.Body.Bytes(), &report); err != nil || res.Code != 200 || report.Data.Spent != 300000 || *report.Data.Forecast.ProjectedSpent != 3000000 || budgets.owner != owner || rows.owner != owner {
		t.Fatalf("report scope/total mismatch: %d %s %v", res.Code, res.Body.String(), err)
	}
	res = request("/api/v1/budgets/"+id+"/transactions?limit=1&offset=1", key.Secret)
	var page struct {
		Data BudgetTransactionPage `json:"data"`
	}
	if err = json.Unmarshal(res.Body.Bytes(), &page); err != nil || res.Code != 200 || page.Data.Total != 2 || len(page.Data.Items) != 1 || page.Data.Items[0].ID != "a" {
		t.Fatalf("drilldown mismatch: %d %s %v", res.Code, res.Body.String(), err)
	}
	for _, tc := range []struct {
		path, token string
		code        int
	}{
		{"/api/v1/budgets/" + id + "/report", "", 401},
		{"/api/v1/budgets/" + id + "/report", "mpk_invalid.invalid", 401},
		{"/api/v1/budgets/" + uuid.NewString() + "/report", key.Secret, 404},
		{"/api/v1/budgets/" + id + "/transactions?offset=-1", key.Secret, 400},
		{"/api/v1/budgets/" + id + "/transactions?limit=101", key.Secret, 400},
		{"/api/v1/budgets/" + id + "/transactions?offset=9223372036854775807", key.Secret, 200},
	} {
		if res = request(tc.path, tc.token); res.Code != tc.code {
			t.Fatalf("path %s status=%d want%d: %s", tc.path, res.Code, tc.code, res.Body.String())
		}
	}
	// Even a broken repository returning another owner's budget must fail closed.
	budgets.items[0].OwnerID = "foreign"
	if res = request("/api/v1/budgets/"+id+"/report", key.Secret); res.Code != 404 {
		t.Fatalf("foreign budget escaped: %d %s", res.Code, res.Body.String())
	}
}
