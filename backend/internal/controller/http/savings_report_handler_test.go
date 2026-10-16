package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/infrastructure/auth"
	"github.com/mypocket/backend/internal/repository"
	"github.com/mypocket/backend/internal/usecase"
)

type savingsHTTPWallets struct {
	repository.WalletRepository
	rows []entity.Wallet
	err  error
}

func (s savingsHTTPWallets) List(owner string) ([]entity.Wallet, error) {
	if owner != "owner" {
		panic("wrong owner")
	}
	return s.rows, s.err
}

type savingsHTTPTransactions struct {
	repository.TransactionRepository
	rows []entity.Transaction
}

func (s savingsHTTPTransactions) List(owner string) ([]entity.Transaction, error) {
	if owner != "owner" {
		panic("wrong owner")
	}
	return s.rows, nil
}

type savingsHTTPUsers struct{ repository.UserRepository }

func (s savingsHTTPUsers) FindByID(owner string) (*entity.User, error) {
	return &entity.User{ID: owner, Timezone: "UTC"}, nil
}

func savingsHTTPFixture() *SavingsReportHandler {
	now := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	return &SavingsReportHandler{Wallets: savingsHTTPWallets{rows: []entity.Wallet{{ID: "11111111-1111-4111-8111-111111111111", OwnerID: "owner", Type: entity.WalletTypeGoal}, {ID: "22222222-2222-4222-8222-222222222222", OwnerID: "owner", Type: entity.WalletTypeBasic}, {ID: "33333333-3333-4333-8333-333333333333", OwnerID: "foreign", Type: entity.WalletTypeGoal}}}, Transactions: savingsHTTPTransactions{rows: []entity.Transaction{{ID: "one", OwnerID: "owner", WalletID: "11111111-1111-4111-8111-111111111111", Type: entity.TransactionTypeIncome, Amount: 100, OccurredAt: now, IncludedInReports: true}, {ID: "foreign", OwnerID: "foreign", WalletID: "11111111-1111-4111-8111-111111111111", Type: entity.TransactionTypeIncome, Amount: 999, OccurredAt: now, IncludedInReports: true}}}, Users: savingsHTTPUsers{}, Now: func() time.Time { return now }}
}

func savingsHTTPRequest(h *SavingsReportHandler, path string, authenticated bool) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	if authenticated {
		r.Use(func(c *gin.Context) { c.Set(contextUserIDKey, "owner") })
	}
	r.GET("/savings", h.GetSavingsReport)
	r.GET("/savings/transactions", h.ListSavingsTransactions)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
	return w
}

func TestSavingsHTTPEnforcesOwnerAndQueryValidation(t *testing.T) {
	for _, tc := range []struct {
		path   string
		auth   bool
		status int
		code   string
	}{
		{"/savings", false, 401, ""},
		{"/savings/transactions?section=goals", false, 401, ""},
		{"/savings?month=2026-13", true, 400, "SAVINGS_INVALID_QUERY"},
		{"/savings?month=2026-11", true, 400, "SAVINGS_FUTURE_MONTH"},
		{"/savings?wallet_id=bad", true, 400, "SAVINGS_INVALID_QUERY"},
		{"/savings?wallet_id=22222222-2222-4222-8222-222222222222", true, 404, "SAVINGS_WALLET_NOT_FOUND"},
		{"/savings?wallet_id=33333333-3333-4333-8333-333333333333", true, 404, "SAVINGS_WALLET_NOT_FOUND"},
		{"/savings/transactions", true, 400, "SAVINGS_INVALID_QUERY"},
		{"/savings/transactions?section=surplus&wallet_id=11111111-1111-4111-8111-111111111111", true, 400, "SAVINGS_INVALID_QUERY"},
		{"/savings/transactions?section=surplus&kind=deposit", true, 400, "SAVINGS_INVALID_QUERY"},
		{"/savings/transactions?section=goals&kind=income", true, 400, "SAVINGS_INVALID_QUERY"},
		{"/savings/transactions?section=goals&limit=0", true, 400, "SAVINGS_INVALID_QUERY"},
		{"/savings/transactions?section=goals&limit=101", true, 400, "SAVINGS_INVALID_QUERY"},
		{"/savings/transactions?section=goals&offset=-1", true, 400, "SAVINGS_INVALID_QUERY"},
		{"/savings/transactions?section=goals&offset=x", true, 400, "SAVINGS_INVALID_QUERY"},
		{"/savings/transactions?section=goals&offset=", true, 400, "SAVINGS_INVALID_QUERY"},
		{"/savings/transactions?section=goals&limit=", true, 400, "SAVINGS_INVALID_QUERY"},
		{"/savings?month=2026-10&month=2026-09", true, 400, "SAVINGS_INVALID_QUERY"},
	} {
		w := savingsHTTPRequest(savingsHTTPFixture(), tc.path, tc.auth)
		if w.Code != tc.status || (tc.code != "" && !strings.Contains(w.Body.String(), `"code":"`+tc.code+`"`)) {
			t.Fatalf("%s: status=%d body=%s", tc.path, w.Code, w.Body.String())
		}
	}
}

func TestSavingsHTTPReturnsReportAndPagedOwnerLedger(t *testing.T) {
	h := savingsHTTPFixture()
	w := savingsHTTPRequest(h, "/savings", true)
	var report struct {
		Data entity.SavingsReport `json:"data"`
	}
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &report) != nil || report.Data.Goals.ClosingBalance != 100 || report.Data.Surplus.Net != 100 {
		t.Fatalf("report status=%d body=%s", w.Code, w.Body.String())
	}
	w = savingsHTTPRequest(h, "/savings/transactions?section=goals&limit=1&offset=1", true)
	var page struct {
		Data entity.SavingsTransactionPage `json:"data"`
	}
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &page) != nil || page.Data.Total != 1 || len(page.Data.Items) != 0 || page.Data.Limit != 1 || page.Data.Offset != 1 {
		t.Fatalf("page status=%d body=%s", w.Code, w.Body.String())
	}
	h.Wallets = savingsHTTPWallets{err: errors.New("db unavailable")}
	w = savingsHTTPRequest(h, "/savings", true)
	if w.Code != 500 || strings.Contains(w.Body.String(), "db unavailable") || !strings.Contains(w.Body.String(), "SAVINGS_FAILED") {
		t.Fatalf("failure status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestSavingsHTTPAcceptsJWTAndAPIKeyThroughOwnerMiddleware(t *testing.T) {
	h := savingsHTTPFixture()
	keys := usecase.NewUserAPIKeyService(&middlewareAPIKeyRepo{})
	key, err := keys.Create(context.Background(), usecase.APIKeyCreateInput{OwnerID: "owner", Name: "Savings report", Scopes: []string{entity.APIKeyScopeFeedbackRead}})
	if err != nil {
		t.Fatal(err)
	}
	tokens := auth.NewJWT("savings-test-jwt-secret-at-least-32-bytes", time.Hour)
	jwt, _, err := tokens.Generate("owner", "owner@example.test", "Owner", "session")
	if err != nil {
		t.Fatal(err)
	}
	middleware := &AuthMiddleware{TokenSvc: tokens, APIKeys: keys, VerifySession: func(session string) (string, error) {
		if session != "session" {
			return "", usecase.ErrInvalidCredentials
		}
		return "owner", nil
	}}
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/savings", middleware.RequireAuth, h.GetSavingsReport)
	router.GET("/savings/transactions", middleware.RequireAuth, h.ListSavingsTransactions)
	for _, token := range []string{jwt, key.Secret} {
		for _, tc := range []struct {
			path   string
			status int
		}{{"/savings", 200}, {"/savings/transactions?section=surplus", 200}, {"/savings?wallet_id=33333333-3333-4333-8333-333333333333", 404}} {
			request := httptest.NewRequest("GET", tc.path, nil)
			request.Header.Set("Authorization", "Bearer "+token)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != tc.status {
				t.Fatalf("%s: status=%d body=%s", tc.path, response.Code, response.Body.String())
			}
		}
	}
	request := httptest.NewRequest("GET", "/savings", nil)
	request.Header.Set("Authorization", "Bearer mpk_invalid.invalid")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != 401 {
		t.Fatalf("invalid key status=%d body=%s", response.Code, response.Body.String())
	}
}
