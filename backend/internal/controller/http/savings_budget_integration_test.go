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
	repo "github.com/mypocket/backend/internal/infrastructure/repository"
	"github.com/mypocket/backend/internal/usecase"
)

func TestSavingsBudgetReportsRealLedgerAndOwnerKey(t *testing.T) {
	f := newOfflineHTTPFixture(t)
	now := time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC)
	goal, incomeCategory, transfer, budgetID := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	target := int64(20000000)
	if err := f.db.Create(&entity.Wallet{ID: goal, OwnerID: f.owner, Name: "House", Type: entity.WalletTypeGoal, Currency: "VND", OpeningBalance: 100000, TargetAmount: &target, CreatedAt: now.AddDate(0, -1, 0)}).Error; err != nil {
		t.Fatal(err)
	}
	if err := f.db.Create(&entity.Category{ID: incomeCategory, OwnerID: &f.owner, Name: "Salary", Kind: "income", IsActive: true}).Error; err != nil {
		t.Fatal(err)
	}
	rows := []entity.Transaction{
		{ID: uuid.NewString(), OwnerID: f.owner, WalletID: f.wallet, CategoryID: &incomeCategory, Type: "income", Amount: 20000000, IncludedInReports: true, OccurredAt: now},
		{ID: uuid.NewString(), OwnerID: f.owner, WalletID: f.wallet, CategoryID: &f.category, Type: "expense", Amount: 12000000, IncludedInReports: true, OccurredAt: now},
		{ID: uuid.NewString(), OwnerID: f.owner, WalletID: f.wallet, CategoryID: &f.category, Type: "expense", Amount: 5000000, TransferID: &transfer, OccurredAt: now},
		{ID: uuid.NewString(), OwnerID: f.owner, WalletID: goal, CategoryID: &incomeCategory, Type: "income", Amount: 5000000, TransferID: &transfer, OccurredAt: now},
	}
	for _, row := range rows {
		if err := f.db.Create(&row).Error; err != nil {
			t.Fatal(err)
		}
		if row.TransferID != nil {
			if err := f.db.Model(&entity.Transaction{}).Where("id = ?", row.ID).Update("included_in_reports", false).Error; err != nil {
				t.Fatal(err)
			}
		}
	}
	start, _ := entity.ParseCalendarDate("2026-09-01")
	end, _ := entity.ParseCalendarDate("2026-09-30")
	if err := f.db.Create(&entity.Budget{ID: budgetID, OwnerID: f.owner, Name: "Food", WalletID: &f.wallet, CategoryID: &f.category, LimitAmount: 20000000, StartDate: start, EndDate: end}).Error; err != nil {
		t.Fatal(err)
	}
	wallets, txs, cats, users := repo.NewWalletPostgresRepository(f.db), repo.NewTransactionPostgresRepository(f.db), repo.NewCategoryPostgresRepository(f.db), repo.NewUserPostgresRepository(f.db)
	keys := usecase.NewUserAPIKeyService(repo.NewUserAPIKeyPostgresRepository(f.db))
	key, err := keys.Create(context.Background(), usecase.APIKeyCreateInput{OwnerID: f.owner, Name: "Synthetic reports UAT"})
	if err != nil {
		t.Fatal(err)
	}
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	middleware := &AuthMiddleware{APIKeys: keys}
	router := &Router{Engine: engine, AuthMiddleware: middleware}
	router.RegisterBudgetReportRoutes(&BudgetReportHandler{Budgets: repo.NewBudgetPostgresRepository(f.db), Transactions: txs, Categories: cats, Users: users, Now: func() time.Time { return now }})
	savings := &SavingsReportHandler{Wallets: wallets, Transactions: txs, Users: users, Now: func() time.Time { return now }}
	engine.GET("/api/v1/reports/savings", middleware.RequireAuth, savings.GetSavingsReport)
	engine.GET("/api/v1/reports/savings/transactions", middleware.RequireAuth, savings.ListSavingsTransactions)
	get := func(path string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("Authorization", "Bearer "+key.Secret)
		res := httptest.NewRecorder()
		engine.ServeHTTP(res, req)
		return res
	}
	res := get("/api/v1/reports/savings?month=2026-09")
	var savingsResult struct {
		Data entity.SavingsReport `json:"data"`
	}
	if err = json.Unmarshal(res.Body.Bytes(), &savingsResult); err != nil || res.Code != 200 {
		t.Fatalf("savings status=%d err=%v body=%s", res.Code, err, res.Body.String())
	}
	if savingsResult.Data.Surplus.Net != 8000000 || savingsResult.Data.Goals.ClosingBalance != 5100000 || savingsResult.Data.Goals.InternalTransfers != 5000000 {
		t.Fatalf("ledger not reconciled: %+v", savingsResult.Data)
	}
	res = get("/api/v1/reports/savings/transactions?month=2026-09&section=surplus")
	var page struct {
		Data entity.SavingsTransactionPage `json:"data"`
	}
	if err = json.Unmarshal(res.Body.Bytes(), &page); err != nil || res.Code != 200 || page.Data.Total != 2 {
		t.Fatalf("surplus page: %d %s", res.Code, res.Body.String())
	}
	res = get("/api/v1/budgets/" + budgetID + "/report")
	var forecast struct {
		Data entity.BudgetReport `json:"data"`
	}
	if err = json.Unmarshal(res.Body.Bytes(), &forecast); err != nil || res.Code != 200 || forecast.Data.Spent != 12000000 || *forecast.Data.Forecast.ProjectedSpent != 24000000 {
		t.Fatalf("budget ledger mismatch: %d %s err=%v", res.Code, res.Body.String(), err)
	}
	res = get("/api/v1/budgets/" + budgetID + "/transactions")
	var budgetPage struct {
		Data BudgetTransactionPage `json:"data"`
	}
	if err = json.Unmarshal(res.Body.Bytes(), &budgetPage); err != nil || res.Code != 200 || budgetPage.Data.Total != 1 {
		t.Fatalf("budget page: %d %s", res.Code, res.Body.String())
	}
	if res = get("/api/v1/reports/savings?month=2026-09&wallet_id=" + f.wallet); res.Code != 404 {
		t.Fatalf("non-goal selected: %d", res.Code)
	}
	// Clean exact owned configuration before shared fixture removes its account.
	t.Cleanup(func() { f.db.Where("id = ? AND owner_id = ?", budgetID, f.owner).Delete(&entity.Budget{}) })
}
