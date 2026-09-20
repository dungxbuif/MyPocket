package httpapi

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mypocket/backend/internal/entity"
)

type transactionRepositoryStub struct{ created *entity.Transaction }

func (s *transactionRepositoryStub) List(string) ([]entity.Transaction, error) { return nil, nil }
func (s *transactionRepositoryStub) Create(item *entity.Transaction) error {
	s.created = item
	return nil
}
func (s *transactionRepositoryStub) Update(string, string, map[string]any) (*entity.Transaction, error) {
	return nil, errors.New("not found")
}
func (s *transactionRepositoryStub) Delete(string, string) error { return nil }

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
	return router
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
