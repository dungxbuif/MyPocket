package httpapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"mypocket/internal/finance"
	"mypocket/internal/identity"
	"mypocket/internal/platform/httpapi"
)

func TestWalletsAPIRequiresAuth(t *testing.T) {
	handler := httpapi.NewRouter(authTestConfig(), httpapi.Dependencies{
		IdentityRepository: &authRepoStub{},
		FinanceRepository:  &financeRepoStub{},
	})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/wallets", nil)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", res.Code, res.Body.String())
	}
	if !strings.Contains(res.Body.String(), "AUTH_REQUIRED") {
		t.Fatalf("expected auth error code, got %s", res.Body.String())
	}
}

func TestWalletsAPIListsAuthenticatedUsersWallets(t *testing.T) {
	repo := &financeRepoStub{
		wallets: []finance.Wallet{{ID: "wallet-1", UserID: "user_123", Name: "Tiền mặt", Type: finance.WalletCash, BalanceVND: 120000, IncludeInTotal: true, Version: 1}},
	}
	handler := httpapi.NewRouter(authTestConfig(), httpapi.Dependencies{
		IdentityRepository: &authRepoStub{user: identity.User{ID: "user_123", Email: "a@example.com", EmailVerified: true}},
		FinanceRepository:  repo,
	})
	req := authenticatedRequest(t, http.MethodGet, "/api/v1/wallets", "")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	if repo.listWalletsUserID != "user_123" {
		t.Fatalf("expected repo scoped to auth user, got %q", repo.listWalletsUserID)
	}
	body := res.Body.String()
	if !strings.Contains(body, `"name":"Tiền mặt"`) || !strings.Contains(body, `"balance_vnd":120000`) {
		t.Fatalf("unexpected wallet response: %s", body)
	}
}

func TestCreateWalletRequiresCSRF(t *testing.T) {
	handler := httpapi.NewRouter(authTestConfig(), httpapi.Dependencies{
		IdentityRepository: &authRepoStub{user: identity.User{ID: "user_123", Email: "a@example.com", EmailVerified: true}},
		FinanceRepository:  &financeRepoStub{},
	})
	req := authenticatedRequest(t, http.MethodPost, "/api/v1/wallets", `{"name":"Tiền mặt","type":"cash"}`)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", res.Code, res.Body.String())
	}
	if !strings.Contains(res.Body.String(), "CSRF_REQUIRED") {
		t.Fatalf("expected csrf error code, got %s", res.Body.String())
	}
}

func TestCreateWalletUsesAuthenticatedUser(t *testing.T) {
	repo := &financeRepoStub{
		createdWallet: finance.Wallet{ID: "wallet-2", UserID: "user_123", Name: "Ngân hàng", Type: finance.WalletBank, IncludeInTotal: true, Version: 1},
	}
	handler := httpapi.NewRouter(authTestConfig(), httpapi.Dependencies{
		IdentityRepository: &authRepoStub{user: identity.User{ID: "user_123", Email: "a@example.com", EmailVerified: true}},
		FinanceRepository:  repo,
	})
	req := authenticatedRequest(t, http.MethodPost, "/api/v1/wallets", `{"name":"Ngân hàng","type":"bank"}`)
	addCSRF(req)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", res.Code, res.Body.String())
	}
	if repo.createWalletUserID != "user_123" {
		t.Fatalf("expected create scoped to auth user, got %q", repo.createWalletUserID)
	}
	if repo.createWalletInput.Name != "Ngân hàng" || repo.createWalletInput.Type != finance.WalletBank {
		t.Fatalf("unexpected wallet input: %#v", repo.createWalletInput)
	}
}

func TestCategoriesAPIListsSystemAndOwnedCategories(t *testing.T) {
	repo := &financeRepoStub{
		categories: []finance.Category{{ID: "cat-1", Name: "Ăn uống", Kind: finance.CategoryExpense, SystemKey: "expense_food", IsSystem: true}},
	}
	handler := httpapi.NewRouter(authTestConfig(), httpapi.Dependencies{
		IdentityRepository: &authRepoStub{user: identity.User{ID: "user_123", Email: "a@example.com", EmailVerified: true}},
		FinanceRepository:  repo,
	})
	req := authenticatedRequest(t, http.MethodGet, "/api/v1/categories", "")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	if repo.listCategoriesUserID != "user_123" {
		t.Fatalf("expected categories scoped to auth user, got %q", repo.listCategoriesUserID)
	}
	var body map[string]any
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !strings.Contains(res.Body.String(), `"system_key":"expense_food"`) {
		t.Fatalf("unexpected categories response: %s", res.Body.String())
	}
}

func authenticatedRequest(t *testing.T, method string, path string, body string) *http.Request {
	t.Helper()

	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	signer := identity.NewCookieSigner([]byte(authTestConfig().CookieSecret))
	value, err := signer.Sign("user_123", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	req.AddCookie(&http.Cookie{Name: identity.AuthCookieName, Value: value})
	return req
}

func addCSRF(req *http.Request) {
	req.AddCookie(&http.Cookie{Name: identity.CSRFCookieName, Value: "csrf-token"})
	req.Header.Set(identity.CSRFHeaderName, "csrf-token")
}

type financeRepoStub struct {
	wallets              []finance.Wallet
	categories           []finance.Category
	createdWallet        finance.Wallet
	listWalletsUserID    string
	createWalletUserID   string
	createWalletInput    finance.CreateWalletInput
	listCategoriesUserID string
}

func (r *financeRepoStub) ListWallets(_ context.Context, userID string) ([]finance.Wallet, error) {
	r.listWalletsUserID = userID
	return r.wallets, nil
}

func (r *financeRepoStub) CreateWallet(_ context.Context, userID string, input finance.CreateWalletInput) (finance.Wallet, error) {
	r.createWalletUserID = userID
	r.createWalletInput = input
	return r.createdWallet, nil
}

func (r *financeRepoStub) ListCategories(_ context.Context, userID string) ([]finance.Category, error) {
	r.listCategoriesUserID = userID
	return r.categories, nil
}
