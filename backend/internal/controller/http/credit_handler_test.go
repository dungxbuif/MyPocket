package httpapi

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
)

type creditRepositoryStub struct {
	entry      *entity.Transaction
	rows       []entity.Transaction
	status     entity.CreditStatement
	entryErr   error
	paymentErr error
}

type creditCategoryRepositoryStub struct {
	categories []entity.Category
}

func (s *creditCategoryRepositoryStub) EnsurePersonalDefaults(string) error { return nil }
func (s *creditCategoryRepositoryStub) ListVisible(string) ([]entity.Category, error) {
	return s.categories, nil
}
func (s *creditCategoryRepositoryStub) FindVisible(string, string) (*entity.Category, error) {
	return nil, errors.New("not implemented")
}
func (s *creditCategoryRepositoryStub) FindPersonal(string, string) (*entity.Category, error) {
	return nil, errors.New("not implemented")
}
func (s *creditCategoryRepositoryStub) Create(string, *entity.Category) error {
	return errors.New("not implemented")
}
func (s *creditCategoryRepositoryStub) Update(string, *entity.Category) error {
	return errors.New("not implemented")
}
func (s *creditCategoryRepositoryStub) Delete(string, string) error {
	return errors.New("not implemented")
}
func (s *creditCategoryRepositoryStub) ReplaceWallets(string, *entity.Category, []string) error {
	return errors.New("not implemented")
}
func (s *creditCategoryRepositoryStub) ValidateWallets(string, []string) error {
	return errors.New("not implemented")
}

func (s *creditRepositoryStub) CreateCreditEntry(_ string, input repository.CreditEntryInput) (*entity.Transaction, error) {
	if s.entryErr != nil {
		return nil, s.entryErr
	}
	kind := input.Kind
	typeValue := entity.TransactionTypeExpense
	if kind == entity.CreditKindRefund {
		typeValue = entity.TransactionTypeIncome
	}
	s.entry = &entity.Transaction{ID: "entry-1", WalletID: input.WalletID, Type: typeValue, Amount: input.Amount, CreditKind: &kind}
	return s.entry, nil
}
func (s *creditRepositoryStub) CreateCreditPayment(_ string, _, _ string, _ int64, _ time.Time, _ *string) ([]entity.Transaction, error) {
	if s.paymentErr != nil {
		return nil, s.paymentErr
	}
	return s.rows, nil
}
func (s *creditRepositoryStub) ListCreditStatement(_ string, _ string, _, _ *time.Time) (entity.CreditStatement, error) {
	return s.status, nil
}

func creditTestRouter(handler *CreditHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set(contextUserIDKey, "owner-1"); c.Next() })
	router.POST("/credit/wallets/:wallet_id/entries", handler.CreateEntry)
	router.POST("/credit/wallets/:wallet_id/payments", handler.CreatePayment)
	router.GET("/credit/wallets/:wallet_id/statement", handler.Statement)
	return router
}

func TestCreateCreditEntryUsesDedicatedWalletFlow(t *testing.T) {
	repo := &creditRepositoryStub{}
	wallets := &transactionWalletRepositoryStub{wallets: map[string]entity.Wallet{"credit-1": {ID: "credit-1", OwnerID: "owner-1", Type: entity.WalletTypeCredit, CreditLimit: int64Ptr(10000000)}}}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/credit/wallets/credit-1/entries", bytes.NewBufferString(`{"kind":"purchase","amount":125000}`))
	request.Header.Set("Content-Type", "application/json")
	creditTestRouter(&CreditHandler{Credits: repo, Wallets: wallets}).ServeHTTP(response, request)
	if response.Code != http.StatusCreated || repo.entry == nil || repo.entry.Type != entity.TransactionTypeExpense {
		t.Fatalf("unexpected credit entry: status=%d body=%s entry=%#v", response.Code, response.Body.String(), repo.entry)
	}
}

func TestCreateCreditPaymentRequiresCreditWallet(t *testing.T) {
	repo := &creditRepositoryStub{}
	wallets := &transactionWalletRepositoryStub{wallets: map[string]entity.Wallet{"basic-1": {ID: "basic-1", OwnerID: "owner-1", Type: entity.WalletTypeBasic}}}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/credit/wallets/basic-1/payments", bytes.NewBufferString(`{"source_wallet_id":"basic-1","amount":100}`))
	request.Header.Set("Content-Type", "application/json")
	creditTestRouter(&CreditHandler{Credits: repo, Wallets: wallets}).ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("expected credit wallet rejection: status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestCreateCreditEntryRejectsInapplicableCategory(t *testing.T) {
	repo := &creditRepositoryStub{}
	wallets := &transactionWalletRepositoryStub{wallets: map[string]entity.Wallet{"credit-1": {ID: "credit-1", OwnerID: "owner-1", Type: entity.WalletTypeCredit, CreditLimit: int64Ptr(10000000)}}}
	categories := &creditCategoryRepositoryStub{categories: []entity.Category{{ID: "food", Kind: entity.TransactionTypeExpense, WalletIDs: []string{"other-wallet"}}}}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/credit/wallets/credit-1/entries", bytes.NewBufferString(`{"kind":"purchase","category_id":"food","amount":125000}`))
	request.Header.Set("Content-Type", "application/json")
	creditTestRouter(&CreditHandler{Credits: repo, Wallets: wallets, Categories: categories}).ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest || repo.entry != nil {
		t.Fatalf("expected category scope rejection: status=%d body=%s entry=%#v", response.Code, response.Body.String(), repo.entry)
	}
}

func TestCreditRangeRejectsReversedDates(t *testing.T) {
	repo := &creditRepositoryStub{}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/credit/wallets/credit-1/statement?from=2026-10-02&to=2026-10-01", nil)
	creditTestRouter(&CreditHandler{Credits: repo}).ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid range: status=%d", response.Code)
	}
}

func int64Ptr(value int64) *int64 { return &value }
