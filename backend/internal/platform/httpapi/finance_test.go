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

func TestUpdateWalletUsesAuthenticatedUser(t *testing.T) {
	includeInTotal := false
	repo := &financeRepoStub{
		updatedWallet: finance.Wallet{ID: "wallet-2", UserID: "user_123", Name: "Ngân hàng phụ", Type: finance.WalletBank, IncludeInTotal: includeInTotal, Version: 2},
	}
	handler := httpapi.NewRouter(authTestConfig(), httpapi.Dependencies{
		IdentityRepository: &authRepoStub{user: identity.User{ID: "user_123", Email: "a@example.com", EmailVerified: true}},
		FinanceRepository:  repo,
	})
	req := authenticatedRequest(t, http.MethodPatch, "/api/v1/wallets/wallet-2", `{"name":"Ngân hàng phụ","include_in_total":false}`)
	addCSRF(req)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	if repo.updateWalletUserID != "user_123" || repo.updateWalletID != "wallet-2" {
		t.Fatalf("wallet update not scoped: user=%q id=%q", repo.updateWalletUserID, repo.updateWalletID)
	}
	if repo.updateWalletInput.Name != "Ngân hàng phụ" || repo.updateWalletInput.IncludeInTotal == nil || *repo.updateWalletInput.IncludeInTotal {
		t.Fatalf("unexpected wallet update input: %#v", repo.updateWalletInput)
	}
}

func TestWalletCommandsUseAuthenticatedUser(t *testing.T) {
	repo := &financeRepoStub{}
	handler := httpapi.NewRouter(authTestConfig(), httpapi.Dependencies{
		IdentityRepository: &authRepoStub{user: identity.User{ID: "user_123", Email: "a@example.com", EmailVerified: true}},
		FinanceRepository:  repo,
	})

	archiveReq := authenticatedRequest(t, http.MethodPost, "/api/v1/wallets/wallet-2/archive", "")
	addCSRF(archiveReq)
	archiveRes := httptest.NewRecorder()
	handler.ServeHTTP(archiveRes, archiveReq)
	if archiveRes.Code != http.StatusOK {
		t.Fatalf("expected archive 200, got %d: %s", archiveRes.Code, archiveRes.Body.String())
	}

	defaultReq := authenticatedRequest(t, http.MethodPost, "/api/v1/wallets/wallet-2/default-ai", "")
	addCSRF(defaultReq)
	defaultRes := httptest.NewRecorder()
	handler.ServeHTTP(defaultRes, defaultReq)
	if defaultRes.Code != http.StatusOK {
		t.Fatalf("expected default 200, got %d: %s", defaultRes.Code, defaultRes.Body.String())
	}

	if repo.archiveWalletUserID != "user_123" || repo.archiveWalletID != "wallet-2" {
		t.Fatalf("wallet archive not scoped: user=%q id=%q", repo.archiveWalletUserID, repo.archiveWalletID)
	}
	if repo.defaultWalletUserID != "user_123" || repo.defaultWalletID != "wallet-2" {
		t.Fatalf("wallet default not scoped: user=%q id=%q", repo.defaultWalletUserID, repo.defaultWalletID)
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

func TestCreateCategoryUsesAuthenticatedUser(t *testing.T) {
	repo := &financeRepoStub{
		createdCategory: finance.Category{ID: "cat-custom", UserID: "user_123", Kind: finance.CategoryExpense, Name: "Cafe", IsSystem: false},
	}
	handler := httpapi.NewRouter(authTestConfig(), httpapi.Dependencies{
		IdentityRepository: &authRepoStub{user: identity.User{ID: "user_123", Email: "a@example.com", EmailVerified: true}},
		FinanceRepository:  repo,
	})
	req := authenticatedRequest(t, http.MethodPost, "/api/v1/categories", `{"kind":"expense","name":"Cafe"}`)
	addCSRF(req)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", res.Code, res.Body.String())
	}
	if repo.createCategoryUserID != "user_123" {
		t.Fatalf("expected create category scoped to auth user, got %q", repo.createCategoryUserID)
	}
	if repo.createCategoryInput.Kind != finance.CategoryExpense || repo.createCategoryInput.Name != "Cafe" {
		t.Fatalf("unexpected category input: %#v", repo.createCategoryInput)
	}
}

func TestCategoryCommandsUseAuthenticatedUser(t *testing.T) {
	repo := &financeRepoStub{
		updatedCategory: finance.Category{ID: "cat-custom", UserID: "user_123", Kind: finance.CategoryExpense, Name: "Cafe mới"},
	}
	handler := httpapi.NewRouter(authTestConfig(), httpapi.Dependencies{
		IdentityRepository: &authRepoStub{user: identity.User{ID: "user_123", Email: "a@example.com", EmailVerified: true}},
		FinanceRepository:  repo,
	})

	updateReq := authenticatedRequest(t, http.MethodPatch, "/api/v1/categories/cat-custom", `{"name":"Cafe mới"}`)
	addCSRF(updateReq)
	updateRes := httptest.NewRecorder()
	handler.ServeHTTP(updateRes, updateReq)
	if updateRes.Code != http.StatusOK {
		t.Fatalf("expected category update 200, got %d: %s", updateRes.Code, updateRes.Body.String())
	}

	archiveReq := authenticatedRequest(t, http.MethodPost, "/api/v1/categories/cat-custom/archive", "")
	addCSRF(archiveReq)
	archiveRes := httptest.NewRecorder()
	handler.ServeHTTP(archiveRes, archiveReq)
	if archiveRes.Code != http.StatusOK {
		t.Fatalf("expected category archive 200, got %d: %s", archiveRes.Code, archiveRes.Body.String())
	}

	if repo.updateCategoryUserID != "user_123" || repo.updateCategoryID != "cat-custom" || repo.updateCategoryInput.Name != "Cafe mới" {
		t.Fatalf("category update not scoped: user=%q id=%q input=%#v", repo.updateCategoryUserID, repo.updateCategoryID, repo.updateCategoryInput)
	}
	if repo.archiveCategoryUserID != "user_123" || repo.archiveCategoryID != "cat-custom" {
		t.Fatalf("category archive not scoped: user=%q id=%q", repo.archiveCategoryUserID, repo.archiveCategoryID)
	}
}

func TestWalletCategoryActivationUsesAuthenticatedUser(t *testing.T) {
	repo := &financeRepoStub{}
	handler := httpapi.NewRouter(authTestConfig(), httpapi.Dependencies{
		IdentityRepository: &authRepoStub{user: identity.User{ID: "user_123", Email: "a@example.com", EmailVerified: true}},
		FinanceRepository:  repo,
	})
	req := authenticatedRequest(t, http.MethodPut, "/api/v1/wallets/wallet-2/categories/cat-food", `{"active":false}`)
	addCSRF(req)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	if repo.walletCategoryUserID != "user_123" || repo.walletCategoryWalletID != "wallet-2" || repo.walletCategoryCategoryID != "cat-food" || repo.walletCategoryActive {
		t.Fatalf("wallet category activation not scoped: user=%q wallet=%q category=%q active=%t", repo.walletCategoryUserID, repo.walletCategoryWalletID, repo.walletCategoryCategoryID, repo.walletCategoryActive)
	}
}

func TestTransactionsAPIRequiresAuth(t *testing.T) {
	handler := httpapi.NewRouter(authTestConfig(), httpapi.Dependencies{
		IdentityRepository: &authRepoStub{},
		FinanceRepository:  &financeRepoStub{},
	})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/transactions", nil)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", res.Code, res.Body.String())
	}
	if !strings.Contains(res.Body.String(), "AUTH_REQUIRED") {
		t.Fatalf("expected auth error code, got %s", res.Body.String())
	}
}

func TestCreateTransactionRequiresCSRF(t *testing.T) {
	handler := httpapi.NewRouter(authTestConfig(), httpapi.Dependencies{
		IdentityRepository: &authRepoStub{user: identity.User{ID: "user_123", Email: "a@example.com", EmailVerified: true}},
		FinanceRepository:  &financeRepoStub{},
	})
	req := authenticatedRequest(t, http.MethodPost, "/api/v1/transactions", `{"type":"income"}`)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", res.Code, res.Body.String())
	}
	if !strings.Contains(res.Body.String(), "CSRF_REQUIRED") {
		t.Fatalf("expected csrf error code, got %s", res.Body.String())
	}
}

func TestCreateTransactionUsesAuthenticatedUserAndIdempotencyKey(t *testing.T) {
	occurredAt := time.Date(2026, 8, 30, 10, 0, 0, 0, time.UTC)
	repo := &financeRepoStub{
		createdTransaction: finance.Transaction{
			ID:              "tx-1",
			UserID:          "user_123",
			Type:            finance.TransactionIncome,
			SourceWalletID:  "wallet-1",
			CategoryID:      "cat-income",
			AmountVND:       500000,
			BalanceAfterVND: 500000,
			OccurredAt:      occurredAt,
			Version:         1,
		},
	}
	handler := httpapi.NewRouter(authTestConfig(), httpapi.Dependencies{
		IdentityRepository: &authRepoStub{user: identity.User{ID: "user_123", Email: "a@example.com", EmailVerified: true}},
		FinanceRepository:  repo,
	})
	req := authenticatedRequest(t, http.MethodPost, "/api/v1/transactions", `{"type":"income","source_wallet_id":"wallet-1","category_id":"cat-income","amount_vnd":500000,"occurred_at":"2026-08-30T10:00:00Z","note":"Lương"}`)
	addCSRF(req)
	req.Header.Set("Idempotency-Key", "idem-tx-1")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", res.Code, res.Body.String())
	}
	if repo.createTransactionUserID != "user_123" || repo.createTransactionInput.IdempotencyKey != "idem-tx-1" {
		t.Fatalf("transaction create not scoped/idempotent: user=%q input=%#v", repo.createTransactionUserID, repo.createTransactionInput)
	}
	if repo.createTransactionInput.OccurredAt.IsZero() || repo.createTransactionInput.AmountVND != 500000 {
		t.Fatalf("transaction input not mapped: %#v", repo.createTransactionInput)
	}
	if !strings.Contains(res.Body.String(), `"id":"tx-1"`) || !strings.Contains(res.Body.String(), `"balance_after_vnd":500000`) {
		t.Fatalf("unexpected transaction response: %s", res.Body.String())
	}
}

func TestTransactionsAPIMapsListFilters(t *testing.T) {
	from := "2026-08-01T00:00:00Z"
	to := "2026-08-31T23:59:59Z"
	excluded := true
	repo := &financeRepoStub{
		transactions: []finance.Transaction{{ID: "tx-filtered", UserID: "user_123", Type: finance.TransactionExpense, SourceWalletID: "wallet-1", CategoryID: "cat-food", AmountVND: 45000, OccurredAt: time.Date(2026, 8, 30, 10, 0, 0, 0, time.UTC), Version: 1}},
	}
	handler := httpapi.NewRouter(authTestConfig(), httpapi.Dependencies{
		IdentityRepository: &authRepoStub{user: identity.User{ID: "user_123", Email: "a@example.com", EmailVerified: true}},
		FinanceRepository:  repo,
	})
	req := authenticatedRequest(t, http.MethodGet, "/api/v1/transactions?wallet_id=wallet-1&category_id=cat-food&type=expense&date_from="+from+"&date_to="+to+"&q=cafe&excluded_from_reports=true", "")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	if repo.listTransactionsUserID != "user_123" {
		t.Fatalf("expected list scoped to auth user, got %q", repo.listTransactionsUserID)
	}
	got := repo.listTransactionFilters
	if got.WalletID != "wallet-1" || got.CategoryID != "cat-food" || got.Type != finance.TransactionExpense || got.Query != "cafe" {
		t.Fatalf("unexpected list filters: %#v", got)
	}
	if got.DateFrom == nil || got.DateTo == nil || got.ExcludedFromReports == nil || *got.ExcludedFromReports != excluded {
		t.Fatalf("missing date/excluded filters: %#v", got)
	}
	if !strings.Contains(res.Body.String(), `"transactions":[`) || !strings.Contains(res.Body.String(), `"id":"tx-filtered"`) {
		t.Fatalf("unexpected transactions response: %s", res.Body.String())
	}
}

func TestUpdateTransactionUsesAuthenticatedUser(t *testing.T) {
	repo := &financeRepoStub{
		updatedTransaction: finance.Transaction{ID: "tx-1", UserID: "user_123", Type: finance.TransactionExpense, SourceWalletID: "wallet-1", CategoryID: "cat-food", AmountVND: 125000, BalanceAfterVND: -125000, OccurredAt: time.Date(2026, 8, 30, 11, 0, 0, 0, time.UTC), Version: 2},
	}
	handler := httpapi.NewRouter(authTestConfig(), httpapi.Dependencies{
		IdentityRepository: &authRepoStub{user: identity.User{ID: "user_123", Email: "a@example.com", EmailVerified: true}},
		FinanceRepository:  repo,
	})
	req := authenticatedRequest(t, http.MethodPatch, "/api/v1/transactions/tx-1", `{"type":"expense","source_wallet_id":"wallet-1","category_id":"cat-food","amount_vnd":125000,"occurred_at":"2026-08-30T11:00:00Z"}`)
	addCSRF(req)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	if repo.updateTransactionUserID != "user_123" || repo.updateTransactionID != "tx-1" {
		t.Fatalf("transaction update not scoped: user=%q id=%q", repo.updateTransactionUserID, repo.updateTransactionID)
	}
	if repo.updateTransactionInput.Type != finance.TransactionExpense || repo.updateTransactionInput.AmountVND != 125000 {
		t.Fatalf("unexpected update input: %#v", repo.updateTransactionInput)
	}
}

func TestArchiveTransactionUsesAuthenticatedUser(t *testing.T) {
	repo := &financeRepoStub{}
	handler := httpapi.NewRouter(authTestConfig(), httpapi.Dependencies{
		IdentityRepository: &authRepoStub{user: identity.User{ID: "user_123", Email: "a@example.com", EmailVerified: true}},
		FinanceRepository:  repo,
	})
	req := authenticatedRequest(t, http.MethodPost, "/api/v1/transactions/tx-1/archive", "")
	addCSRF(req)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	if repo.archiveTransactionUserID != "user_123" || repo.archiveTransactionID != "tx-1" {
		t.Fatalf("transaction archive not scoped: user=%q id=%q", repo.archiveTransactionUserID, repo.archiveTransactionID)
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
	wallets                  []finance.Wallet
	categories               []finance.Category
	transactions             []finance.Transaction
	createdWallet            finance.Wallet
	updatedWallet            finance.Wallet
	createdTransaction       finance.Transaction
	createdCategory          finance.Category
	updatedCategory          finance.Category
	updatedTransaction       finance.Transaction
	listWalletsUserID        string
	createWalletUserID       string
	createWalletInput        finance.CreateWalletInput
	updateWalletUserID       string
	updateWalletID           string
	updateWalletInput        finance.UpdateWalletInput
	archiveWalletUserID      string
	archiveWalletID          string
	defaultWalletUserID      string
	defaultWalletID          string
	listCategoriesUserID     string
	createCategoryUserID     string
	createCategoryInput      finance.CreateCategoryInput
	updateCategoryUserID     string
	updateCategoryID         string
	updateCategoryInput      finance.UpdateCategoryInput
	archiveCategoryUserID    string
	archiveCategoryID        string
	walletCategoryUserID     string
	walletCategoryWalletID   string
	walletCategoryCategoryID string
	walletCategoryActive     bool
	listTransactionsUserID   string
	listTransactionFilters   finance.TransactionFilters
	createTransactionUserID  string
	createTransactionInput   finance.CreateTransactionInput
	updateTransactionUserID  string
	updateTransactionID      string
	updateTransactionInput   finance.UpdateTransactionInput
	archiveTransactionUserID string
	archiveTransactionID     string
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

func (r *financeRepoStub) UpdateWallet(_ context.Context, userID string, walletID string, input finance.UpdateWalletInput) (finance.Wallet, error) {
	r.updateWalletUserID = userID
	r.updateWalletID = walletID
	r.updateWalletInput = input
	return r.updatedWallet, nil
}

func (r *financeRepoStub) ArchiveWallet(_ context.Context, userID string, walletID string) error {
	r.archiveWalletUserID = userID
	r.archiveWalletID = walletID
	return nil
}

func (r *financeRepoStub) SetDefaultAIWallet(_ context.Context, userID string, walletID string) error {
	r.defaultWalletUserID = userID
	r.defaultWalletID = walletID
	return nil
}

func (r *financeRepoStub) ListCategories(_ context.Context, userID string) ([]finance.Category, error) {
	r.listCategoriesUserID = userID
	return r.categories, nil
}

func (r *financeRepoStub) CreateCategory(_ context.Context, userID string, input finance.CreateCategoryInput) (finance.Category, error) {
	r.createCategoryUserID = userID
	r.createCategoryInput = input
	return r.createdCategory, nil
}

func (r *financeRepoStub) UpdateCategory(_ context.Context, userID string, categoryID string, input finance.UpdateCategoryInput) (finance.Category, error) {
	r.updateCategoryUserID = userID
	r.updateCategoryID = categoryID
	r.updateCategoryInput = input
	return r.updatedCategory, nil
}

func (r *financeRepoStub) ArchiveCategory(_ context.Context, userID string, categoryID string) error {
	r.archiveCategoryUserID = userID
	r.archiveCategoryID = categoryID
	return nil
}

func (r *financeRepoStub) SetWalletCategoryActive(_ context.Context, userID string, walletID string, categoryID string, active bool) error {
	r.walletCategoryUserID = userID
	r.walletCategoryWalletID = walletID
	r.walletCategoryCategoryID = categoryID
	r.walletCategoryActive = active
	return nil
}

func (r *financeRepoStub) ListTransactions(_ context.Context, userID string, filters finance.TransactionFilters) ([]finance.Transaction, error) {
	r.listTransactionsUserID = userID
	r.listTransactionFilters = filters
	return r.transactions, nil
}

func (r *financeRepoStub) CreateTransaction(_ context.Context, userID string, input finance.CreateTransactionInput) (finance.Transaction, error) {
	r.createTransactionUserID = userID
	r.createTransactionInput = input
	return r.createdTransaction, nil
}

func (r *financeRepoStub) UpdateTransaction(_ context.Context, userID string, transactionID string, input finance.UpdateTransactionInput) (finance.Transaction, error) {
	r.updateTransactionUserID = userID
	r.updateTransactionID = transactionID
	r.updateTransactionInput = input
	return r.updatedTransaction, nil
}

func (r *financeRepoStub) ArchiveTransaction(_ context.Context, userID string, transactionID string) error {
	r.archiveTransactionUserID = userID
	r.archiveTransactionID = transactionID
	return nil
}
