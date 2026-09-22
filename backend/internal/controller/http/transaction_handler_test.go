package httpapi

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
)

type transactionRepositoryStub struct {
	created  *entity.Transaction
	existing *entity.Transaction
	updated  *entity.Transaction
}

func (s *transactionRepositoryStub) List(string) ([]entity.Transaction, error) { return nil, nil }
func (s *transactionRepositoryStub) Find(string, string) (*entity.Transaction, error) {
	return s.existing, nil
}
func (s *transactionRepositoryStub) Create(item *entity.Transaction) error {
	s.created = item
	return nil
}
func (s *transactionRepositoryStub) Update(string, string, map[string]any) (*entity.Transaction, error) {
	return s.updated, nil
}
func (s *transactionRepositoryStub) Delete(string, string) error { return nil }

type transactionUserRepositoryStub struct{ user entity.User }

func (s *transactionUserRepositoryStub) FindByEmail(string) (*entity.User, error) { return nil, nil }
func (s *transactionUserRepositoryStub) FindByID(string) (*entity.User, error)    { return &s.user, nil }
func (s *transactionUserRepositoryStub) FindOrCreateGoogleUser(entity.GoogleProfile) (*entity.User, error) {
	return nil, nil
}
func (s *transactionUserRepositoryStub) SetTimezone(string, string, bool) (*entity.User, error) {
	return nil, nil
}
func (s *transactionUserRepositoryStub) Create(*entity.User) error { return nil }

type transactionJarRepositoryStub struct {
	configLookups int
	configErr     error
}

func (s *transactionJarRepositoryStub) ListMonth(string, string, string) (*entity.JarMonthSummary, error) {
	return nil, nil
}
func (s *transactionJarRepositoryStub) Create(string, string, string, string, string, *int64, *int) (*entity.JarMonthConfig, error) {
	return nil, nil
}
func (s *transactionJarRepositoryStub) UpdateMonthConfig(string, string, string, string, string, *int64, *int) (*entity.JarMonthConfig, error) {
	return nil, nil
}
func (s *transactionJarRepositoryStub) RemoveMonthConfig(string, string, string) error { return nil }
func (s *transactionJarRepositoryStub) FindMonthConfig(string, string, string, bool) (*entity.JarMonthConfig, error) {
	s.configLookups++
	return nil, s.configErr
}
func (s *transactionJarRepositoryStub) Cumulative(string, string, string, string, string) (*entity.JarCumulativeSummary, error) {
	return nil, nil
}

type transactionWalletRepositoryStub struct{ wallets map[string]entity.Wallet }

func (s *transactionWalletRepositoryStub) List(string) ([]entity.Wallet, error) { return nil, nil }
func (s *transactionWalletRepositoryStub) Find(ownerID, id string) (*entity.Wallet, error) {
	wallet, ok := s.wallets[id]
	if !ok || wallet.OwnerID != ownerID {
		return nil, errors.New("not found")
	}
	return &wallet, nil
}
func (s *transactionWalletRepositoryStub) Create(*entity.Wallet) error { return nil }
func (s *transactionWalletRepositoryStub) Update(string, string, map[string]any) (*entity.Wallet, error) {
	return nil, errors.New("not found")
}
func (s *transactionWalletRepositoryStub) Delete(string, string) error { return nil }

type transactionCategoryRepositoryStub struct{ categories []entity.Category }

func (s *transactionCategoryRepositoryStub) EnsurePersonalDefaults(string) error { return nil }
func (s *transactionCategoryRepositoryStub) ListVisible(string) ([]entity.Category, error) {
	return s.categories, nil
}
func (s *transactionCategoryRepositoryStub) FindVisible(string, string) (*entity.Category, error) {
	return nil, errors.New("not found")
}
func (s *transactionCategoryRepositoryStub) FindPersonal(string, string) (*entity.Category, error) {
	return nil, errors.New("not found")
}
func (s *transactionCategoryRepositoryStub) Create(string, *entity.Category) error { return nil }
func (s *transactionCategoryRepositoryStub) Update(string, *entity.Category) error { return nil }
func (s *transactionCategoryRepositoryStub) Delete(string, string) error           { return nil }
func (s *transactionCategoryRepositoryStub) ReplaceWallets(string, *entity.Category, []string) error {
	return nil
}
func (s *transactionCategoryRepositoryStub) ValidateWallets(string, []string) error { return nil }

func transactionTestRouter(handler *TransactionHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set(contextUserIDKey, "owner-1"); c.Next() })
	router.POST("/transactions", handler.CreateTransaction)
	router.PATCH("/transactions/:id", handler.UpdateTransaction)
	return router
}

func TestUpdatingUnchangedJarAssignmentSurvivesTimezoneRebucketing(t *testing.T) {
	occurredAt := time.Date(2026, time.February, 1, 0, 30, 0, 0, time.UTC)
	jarID := "8ec3f4f6-c002-4e59-a16e-ae5ea6237a70"
	existing := &entity.Transaction{ID: "transaction-1", OwnerID: "owner-1", WalletID: "wallet-1", Type: entity.TransactionTypeExpense, Amount: 1200, OccurredAt: occurredAt, JarID: &jarID}
	transactions := &transactionRepositoryStub{existing: existing, updated: existing}
	users := &transactionUserRepositoryStub{user: entity.User{ID: "owner-1", Timezone: "America/Los_Angeles"}}
	jars := &transactionJarRepositoryStub{configErr: repository.ErrJarInvalid}
	wallets := &transactionWalletRepositoryStub{wallets: map[string]entity.Wallet{
		"wallet-1": {ID: "wallet-1", OwnerID: "owner-1", Type: entity.WalletTypeBasic},
	}}
	handler := NewTransactionHandler(transactions, wallets, &transactionCategoryRepositoryStub{})
	handler.Users, handler.Jars = users, jars
	router := transactionTestRouter(handler)
	body := fmt.Sprintf(`{"wallet_id":"wallet-1","type":"expense","amount":1200,"occurred_at":%q,"jar_id":%q,"note":"updated note"}`, occurredAt.Format(time.RFC3339), jarID)
	request := httptest.NewRequest(http.MethodPatch, "/transactions/transaction-1", bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("unchanged historical jar association should remain editable after timezone rebucketing; status=%d body=%s", response.Code, response.Body.String())
	}
	if jars.configLookups != 0 {
		t.Fatalf("an unchanged persisted association should not require a configuration in the newly derived local month, got %d lookups", jars.configLookups)
	}
}

func TestCreateTransactionRejectsCategoryRestrictedToAnotherWallet(t *testing.T) {
	transactions := &transactionRepositoryStub{}
	wallets := &transactionWalletRepositoryStub{wallets: map[string]entity.Wallet{
		"wallet-1": {ID: "wallet-1", OwnerID: "owner-1"},
	}}
	categories := &transactionCategoryRepositoryStub{categories: []entity.Category{
		{ID: "category-1", Kind: entity.TransactionTypeExpense, WalletIDs: []string{"wallet-2"}},
	}}
	router := transactionTestRouter(NewTransactionHandler(transactions, wallets, categories))
	request := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewBufferString(`{"wallet_id":"wallet-1","category_id":"category-1","type":"expense","amount":50000}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if transactions.created != nil {
		t.Fatalf("restricted category was persisted: %#v", transactions.created)
	}
}

func TestCreateTransactionAllowsCategoryWithoutWalletRestriction(t *testing.T) {
	transactions := &transactionRepositoryStub{}
	wallets := &transactionWalletRepositoryStub{wallets: map[string]entity.Wallet{
		"wallet-1": {ID: "wallet-1", OwnerID: "owner-1"},
	}}
	categories := &transactionCategoryRepositoryStub{categories: []entity.Category{
		{ID: "category-1", Kind: entity.TransactionTypeIncome, WalletIDs: []string{}},
	}}
	router := transactionTestRouter(NewTransactionHandler(transactions, wallets, categories))
	request := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewBufferString(`{"wallet_id":"wallet-1","category_id":"category-1","type":"income","amount":50000}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated || transactions.created == nil {
		t.Fatalf("status = %d, transaction = %#v, body = %s", response.Code, transactions.created, response.Body.String())
	}
}

func TestCreateTransactionRejectsCreditWalletUntilCreditLedgerExists(t *testing.T) {
	transactions := &transactionRepositoryStub{}
	wallets := &transactionWalletRepositoryStub{wallets: map[string]entity.Wallet{
		"credit-1": {ID: "credit-1", OwnerID: "owner-1", Type: entity.WalletTypeCredit},
	}}
	router := transactionTestRouter(NewTransactionHandler(transactions, wallets, &transactionCategoryRepositoryStub{}))
	request := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewBufferString(`{"wallet_id":"credit-1","type":"expense","amount":50000}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if transactions.created != nil {
		t.Fatalf("credit transaction was persisted: %#v", transactions.created)
	}
}

func TestGoalTransactionCategoryRules(t *testing.T) {
	for _, tc := range []struct {
		key, kind  string
		restricted bool
		want       int
	}{
		{"income_transfer_in", "income", false, 201},
		{"income_interest", "income", false, 201},
		{"expense_transfer_out", "expense", false, 201},
		{"income_salary", "income", false, 400},
		{"expense_food", "expense", false, 400},
		{"income_interest", "expense", false, 400},
		{"income_interest", "income", true, 400},
	} {
		t.Run(tc.key+tc.kind+fmt.Sprint(tc.restricted), func(t *testing.T) {
			transactions := &transactionRepositoryStub{}
			wallets := &transactionWalletRepositoryStub{wallets: map[string]entity.Wallet{"goal": {ID: "goal", OwnerID: "owner-1", Type: entity.WalletTypeGoal}}}
			category := entity.Category{ID: "category", Kind: tc.kind, SystemKey: &tc.key, IsSystem: true}
			if tc.restricted {
				category.WalletIDs = []string{"other"}
			}
			router := transactionTestRouter(NewTransactionHandler(transactions, wallets, &transactionCategoryRepositoryStub{categories: []entity.Category{category}}))
			request := httptest.NewRequest("POST", "/transactions", bytes.NewBufferString(fmt.Sprintf(`{"wallet_id":"goal","category_id":"category","type":%q,"amount":10000}`, tc.kind)))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != tc.want {
				t.Fatalf("status=%d want=%d body=%s", response.Code, tc.want, response.Body.String())
			}
			if tc.want == 201 && (transactions.created == nil || transactions.created.WalletID != "goal" || !transactions.created.IncludedInReports) {
				t.Fatal("external entry must persist in selected wallet with reports enabled")
			}
			if tc.want == 400 && transactions.created != nil {
				t.Fatal("invalid category persisted")
			}
		})
	}
}
