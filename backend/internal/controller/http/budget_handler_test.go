package httpapi

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	"net/http/httptest"
	"testing"
)

type budgetRepoStub struct{ saved *entity.Budget }

func (r *budgetRepoStub) List(string) ([]entity.Budget, error) { return []entity.Budget{}, nil }
func (r *budgetRepoStub) Save(owner string, b *entity.Budget, creating bool) error {
	r.saved = b
	return nil
}

func TestCreateBudgetRejectsArchivedCategory(t *testing.T) {
	categoryID := uuid.NewString()
	repo := &budgetRepoStub{}
	h := &BudgetHandler{Budgets: repo, Wallets: &transactionWalletRepositoryStub{}, Categories: &transactionCategoryRepositoryStub{categories: []entity.Category{{ID: categoryID, Kind: entity.TransactionTypeExpense, IsActive: false}}}, Transactions: &transactionRepositoryStub{}}
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set(contextUserIDKey, "owner-1") })
	router.POST("/budgets", h.CreateBudget)
	body := `{"name":"Food","limit_amount":1000,"category_id":"` + categoryID + `","start_date":"2026-09-01","end_date":"2026-09-30"}`
	req := httptest.NewRequest("POST", "/budgets", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != 400 || !bytes.Contains(res.Body.Bytes(), []byte(`"code":"category_inactive"`)) || repo.saved != nil {
		t.Fatalf("archived category must not be assigned to a new budget: status=%d body=%s saved=%#v", res.Code, res.Body.String(), repo.saved)
	}
}
func (r *budgetRepoStub) Delete(string, string) error { return nil }

func TestBudgetInputValidation(t *testing.T) {
	for _, tc := range []struct {
		body string
		want int
	}{
		{`{"name":"Food","limit_amount":1000,"start_date":"2026-09-01","end_date":"2026-09-30"}`, 201},
		{`{"name":"Food","limit_amount":0,"start_date":"2026-09-01","end_date":"2026-09-30"}`, 400},
		{`{"name":"Food","limit_amount":1000,"wallet_id":"foreign","start_date":"2026-09-01","end_date":"2026-09-30"}`, 400},
		{`{"name":"Food","limit_amount":1000,"category_id":"income","start_date":"2026-09-01","end_date":"2026-09-30"}`, 400},
		{`{"name":"Food","limit_amount":1000,"start_date":"2026-10-01","end_date":"2026-09-30"}`, 400},
	} {
		repo := &budgetRepoStub{}
		h := &BudgetHandler{Budgets: repo, Wallets: &transactionWalletRepositoryStub{}, Categories: &transactionCategoryRepositoryStub{}, Transactions: &transactionRepositoryStub{}}
		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.Use(func(c *gin.Context) { c.Set(contextUserIDKey, "owner-1") })
		router.POST("/budgets", h.CreateBudget)
		req := httptest.NewRequest("POST", "/budgets", bytes.NewBufferString(tc.body))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		if res.Code != tc.want {
			t.Fatalf("status=%d want=%d body=%s", res.Code, tc.want, res.Body.String())
		}
		if tc.want == 400 && repo.saved != nil {
			t.Fatal("invalid configuration persisted")
		}
		if tc.want == 201 && (repo.saved == nil || repo.saved.OwnerID != "owner-1") {
			t.Fatal("missing owner scope")
		}
	}
}

func TestCreateBudgetPersistsOneExplicitCalendarInterval(t *testing.T) {
	repo := &budgetRepoStub{}
	h := &BudgetHandler{Budgets: repo, Wallets: &transactionWalletRepositoryStub{}, Categories: &transactionCategoryRepositoryStub{}, Transactions: &transactionRepositoryStub{}}
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set(contextUserIDKey, "owner-1") })
	router.POST("/budgets", h.CreateBudget)
	req := httptest.NewRequest("POST", "/budgets", bytes.NewBufferString(`{"name":"Tháng 9","limit_amount":10000,"start_date":"2026-09-01","end_date":"2026-09-30"}`))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != 201 || repo.saved == nil {
		t.Fatalf("status=%d body=%s saved=%#v", res.Code, res.Body.String(), repo.saved)
	}
	if repo.saved.StartDate.Format("2006-01-02") != "2026-09-01" || repo.saved.EndDate.Format("2006-01-02") != "2026-09-30" {
		t.Fatalf("explicit calendar interval changed: start=%s end=%s", repo.saved.StartDate, repo.saved.EndDate)
	}
}
