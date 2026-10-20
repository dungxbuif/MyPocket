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
	"github.com/mypocket/backend/internal/usecase"
)

type versionedTransactionStub struct {
	transactionRepositoryStub
	updateVersionCalled bool
	deleteVersionCalled bool
}

func (s *versionedTransactionStub) UpdateWithVersion(string, string, int64, map[string]any) (*entity.Transaction, error) {
	s.updateVersionCalled = true
	return nil, repository.ErrTransactionVersionConflict
}

func (s *versionedTransactionStub) DeleteWithVersion(string, string, int64) error {
	s.deleteVersionCalled = true
	return repository.ErrTransactionVersionConflict
}

func TestUpdateTransactionReturnsConflictWhenExpectedVersionIsStale(t *testing.T) {
	categoryID := "category"
	transactions := &versionedTransactionStub{transactionRepositoryStub: transactionRepositoryStub{existing: &entity.Transaction{
		ID: "transaction", OwnerID: "owner-1", WalletID: "wallet", CategoryID: &categoryID,
		Type: entity.TransactionTypeExpense, Amount: 100, OccurredAt: time.Now().UTC(), Version: 2,
	}}}
	wallets := &transactionWalletRepositoryStub{wallets: map[string]entity.Wallet{"wallet": {ID: "wallet", OwnerID: "owner-1", Type: entity.WalletTypeBasic}}}
	categories := &transactionCategoryRepositoryStub{categories: []entity.Category{{ID: categoryID, Kind: entity.TransactionTypeExpense, IsActive: true}}}
	router := transactionTestRouter(NewTransactionHandler(transactions, wallets, categories))
	req := httptest.NewRequest(http.MethodPatch, "/transactions/transaction", bytes.NewBufferString(`{"wallet_id":"wallet","category_id":"category","type":"expense","amount":100,"expected_version":1}`))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != http.StatusConflict || !transactions.updateVersionCalled {
		t.Fatalf("status=%d body=%s versioned=%v", res.Code, res.Body.String(), transactions.updateVersionCalled)
	}
}

func TestDeleteTransactionReturnsConflictWhenExpectedVersionIsStale(t *testing.T) {
	categoryID := "category"
	transactions := &versionedTransactionStub{transactionRepositoryStub: transactionRepositoryStub{existing: &entity.Transaction{
		ID: "transaction", OwnerID: "owner-1", WalletID: "wallet", CategoryID: &categoryID,
		Type: entity.TransactionTypeExpense, Amount: 100, OccurredAt: time.Now().UTC(), Version: 2,
	}}}
	router := transactionTestRouter(NewTransactionHandler(transactions, &transactionWalletRepositoryStub{}, &transactionCategoryRepositoryStub{}))
	req := httptest.NewRequest(http.MethodDelete, "/transactions/transaction", bytes.NewBufferString(`{"expected_version":1}`))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != http.StatusConflict || !transactions.deleteVersionCalled {
		t.Fatalf("status=%d body=%s versioned=%v", res.Code, res.Body.String(), transactions.deleteVersionCalled)
	}
}

type versionedCategoryStub struct {
	*categoryRepositoryStubForVersion
	called bool
}

type categoryRepositoryStubForVersion struct{ category entity.Category }

func (s *categoryRepositoryStubForVersion) EnsurePersonalDefaults(string) error { return nil }
func (s *categoryRepositoryStubForVersion) ListVisible(string) ([]entity.Category, error) {
	return []entity.Category{s.category}, nil
}
func (s *categoryRepositoryStubForVersion) FindVisible(_, id string) (*entity.Category, error) {
	if id == s.category.ID {
		return &s.category, nil
	}
	return nil, errors.New("not found")
}
func (s *categoryRepositoryStubForVersion) FindPersonal(_, id string) (*entity.Category, error) {
	if id == s.category.ID {
		return &s.category, nil
	}
	return nil, errors.New("not found")
}
func (s *categoryRepositoryStubForVersion) SetActive(string, string, bool) (*entity.Category, error) {
	return nil, errors.New("not found")
}
func (s *categoryRepositoryStubForVersion) Create(string, *entity.Category) error { return nil }
func (s *categoryRepositoryStubForVersion) Update(string, *entity.Category) error { return nil }
func (s *categoryRepositoryStubForVersion) Delete(string, string) error           { return nil }
func (s *categoryRepositoryStubForVersion) ReplaceWallets(string, *entity.Category, []string) error {
	return nil
}
func (s *categoryRepositoryStubForVersion) ValidateWallets(string, []string) error { return nil }
func (s *versionedCategoryStub) UpdateWithVersion(string, string, int64, *entity.Category, []string) error {
	s.called = true
	return repository.ErrCategoryVersionConflict
}

func TestCategoryUpdateReturnsConflictWhenExpectedVersionIsStale(t *testing.T) {
	base := &categoryRepositoryStubForVersion{category: entity.Category{ID: "category", OwnerID: ptr("owner-1"), Name: "Cafe", Kind: usecase.CategoryKindExpense, IsActive: true, Version: 2}}
	repositoryStub := &versionedCategoryStub{categoryRepositoryStubForVersion: base}
	router := categoryTestRouterForVersion(usecase.NewCategoryInteractor(repositoryStub))
	req := httptest.NewRequest(http.MethodPatch, "/categories/category", bytes.NewBufferString(`{"name":"Cafe 2","kind":"expense","expected_version":1}`))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != http.StatusConflict || !repositoryStub.called {
		t.Fatalf("status=%d body=%s versioned=%v", res.Code, res.Body.String(), repositoryStub.called)
	}
}

func categoryTestRouterForVersion(handler *usecase.CategoryInteractor) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set(contextUserIDKey, "owner-1"); c.Next() })
	router.PATCH("/categories/:id", NewCategoryHandler(handler).UpdateCategory)
	return router
}

func ptr(value string) *string { return &value }
