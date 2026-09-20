package httpapi

import (
	"bytes"
	"github.com/gin-gonic/gin"
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
func (r *budgetRepoStub) Delete(string, string) error { return nil }

func TestBudgetInputValidation(t *testing.T) {
	for _, tc := range []struct {
		body string
		want int
	}{
		{`{"name":"Food","limit_amount":1000,"start_at":"2026-09-01T00:00:00Z","end_at":"2026-10-01T00:00:00Z"}`, 201},
		{`{"name":"Food","limit_amount":0,"start_at":"2026-09-01T00:00:00Z","end_at":"2026-10-01T00:00:00Z"}`, 400},
		{`{"name":"Food","limit_amount":1000,"wallet_id":"foreign","start_at":"2026-09-01T00:00:00Z","end_at":"2026-10-01T00:00:00Z"}`, 400},
		{`{"name":"Food","limit_amount":1000,"category_id":"income","start_at":"2026-09-01T00:00:00Z","end_at":"2026-10-01T00:00:00Z"}`, 400},
		{`{"name":"Food","limit_amount":1000,"start_at":"2026-10-01T00:00:00Z","end_at":"2026-09-01T00:00:00Z"}`, 400},
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
