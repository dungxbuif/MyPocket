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
	created             *entity.Transaction
	existing            *entity.Transaction
	updated             *entity.Transaction
	transferSource      *entity.Transaction
	transferDestination *entity.Transaction
	transferUpdated     []entity.Transaction
	transferDeleted     string
	transferErr         error
	mutationCalls       int
	mutationReplay      bool
	mutationErr         error
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
func (s *transactionRepositoryStub) CreateTransfer(_ string, source, destination *entity.Transaction) error {
	s.transferSource, s.transferDestination = source, destination
	return nil
}
func (s *transactionRepositoryStub) UpdateTransfer(_ string, _ string, _ repository.TransferUpdate) ([]entity.Transaction, error) {
	if s.transferErr != nil {
		return nil, s.transferErr
	}
	return s.transferUpdated, nil
}
func (s *transactionRepositoryStub) DeleteTransfer(_ string, transferID string) error {
	if s.transferErr != nil {
		return s.transferErr
	}
	s.transferDeleted = transferID
	return nil
}
func (s *transactionRepositoryStub) UpdateTransferIdempotent(_ string, _ string, _ repository.TransferUpdate, _, _ string) ([]entity.Transaction, bool, error) {
	s.mutationCalls++
	if s.mutationErr != nil {
		return nil, false, s.mutationErr
	}
	return s.transferUpdated, s.mutationReplay, nil
}
func (s *transactionRepositoryStub) DeleteTransferIdempotent(_ string, transferID, _, _ string) (bool, error) {
	s.mutationCalls++
	if s.mutationErr != nil {
		return false, s.mutationErr
	}
	s.transferDeleted = transferID
	return s.mutationReplay, nil
}

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
	router.POST("/transactions/transfer", handler.CreateTransfer)
	router.PATCH("/transactions/transfer/:transfer_id", handler.UpdateTransfer)
	router.DELETE("/transactions/transfer/:transfer_id", handler.DeleteTransfer)
	router.PATCH("/transactions/:id", handler.UpdateTransaction)
	router.DELETE("/transactions/:id", handler.DeleteTransaction)
	return router
}

func TestCreateTransferPersistsAtomicPairedRows(t *testing.T) {
	transactions := &transactionRepositoryStub{}
	wallets := &transactionWalletRepositoryStub{wallets: map[string]entity.Wallet{
		"wallet-1": {ID: "wallet-1", OwnerID: "owner-1", Type: entity.WalletTypeBasic},
		"wallet-2": {ID: "wallet-2", OwnerID: "owner-1", Type: entity.WalletTypeGoal},
	}}
	categories := &transactionCategoryRepositoryStub{categories: []entity.Category{
		{ID: "cat-out", IsSystem: true, SystemKey: stringPtr("expense_transfer_out")},
		{ID: "cat-in", IsSystem: true, SystemKey: stringPtr("income_transfer_in")},
	}}
	router := transactionTestRouter(NewTransactionHandler(transactions, wallets, categories))
	request := httptest.NewRequest(http.MethodPost, "/transactions/transfer", bytes.NewBufferString(`{"source_wallet_id":"wallet-1","destination_wallet_id":"wallet-2","amount":50000,"occurred_at":"2026-09-22T10:00:00+07:00","note":"tiết kiệm"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if transactions.transferSource == nil || transactions.transferDestination == nil {
		t.Fatal("expected both transfer rows")
	}
	if transactions.transferSource.Type != entity.TransactionTypeExpense || transactions.transferDestination.Type != entity.TransactionTypeIncome {
		t.Fatalf("unexpected transfer directions: %#v %#v", transactions.transferSource, transactions.transferDestination)
	}
	if transactions.transferSource.TransferID == nil || transactions.transferDestination.TransferID == nil || *transactions.transferSource.TransferID != *transactions.transferDestination.TransferID {
		t.Fatal("transfer rows must share a transfer id")
	}
	if transactions.transferSource.IncludedInReports || transactions.transferDestination.IncludedInReports {
		t.Fatal("internal transfer rows must be excluded from reports")
	}
}

func TestCreateTransferRejectsSameWallet(t *testing.T) {
	transactions := &transactionRepositoryStub{}
	wallets := &transactionWalletRepositoryStub{wallets: map[string]entity.Wallet{"wallet-1": {ID: "wallet-1", OwnerID: "owner-1", Type: entity.WalletTypeBasic}}}
	router := transactionTestRouter(NewTransactionHandler(transactions, wallets, &transactionCategoryRepositoryStub{}))
	request := httptest.NewRequest(http.MethodPost, "/transactions/transfer", bytes.NewBufferString(`{"source_wallet_id":"wallet-1","destination_wallet_id":"wallet-1","amount":1}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest || transactions.transferSource != nil {
		t.Fatalf("same-wallet transfer should be rejected: status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestUpdateTransferReturnsBothRows(t *testing.T) {
	when := time.Date(2026, time.September, 22, 3, 0, 0, 0, time.UTC)
	rows := []entity.Transaction{
		{ID: "source", OwnerID: "owner-1", Type: entity.TransactionTypeExpense, Amount: 90000, OccurredAt: when},
		{ID: "destination", OwnerID: "owner-1", Type: entity.TransactionTypeIncome, Amount: 90000, OccurredAt: when},
	}
	transactions := &transactionRepositoryStub{transferUpdated: rows}
	router := transactionTestRouter(NewTransactionHandler(transactions, &transactionWalletRepositoryStub{}, &transactionCategoryRepositoryStub{}))
	request := httptest.NewRequest(http.MethodPatch, "/transactions/transfer/transfer-1", bytes.NewBufferString(`{"amount":90000,"occurred_at":"2026-09-22T10:00:00Z","note":"updated"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "transfer-update-1")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if len(transactions.transferUpdated) != 2 {
		t.Fatalf("expected both transfer rows, got %#v", transactions.transferUpdated)
	}
}

func TestDeleteTransferDeletesBothRowsAtomically(t *testing.T) {
	transactions := &transactionRepositoryStub{}
	router := transactionTestRouter(NewTransactionHandler(transactions, &transactionWalletRepositoryStub{}, &transactionCategoryRepositoryStub{}))
	request := httptest.NewRequest(http.MethodDelete, "/transactions/transfer/transfer-1", nil)
	request.Header.Set("Idempotency-Key", "transfer-delete-1")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if transactions.transferDeleted != "transfer-1" {
		t.Fatalf("expected transfer deletion, got %q", transactions.transferDeleted)
	}
}

func TestTransferMutationRequiresIdempotencyKey(t *testing.T) {
	transactions := &transactionRepositoryStub{transferUpdated: []entity.Transaction{{ID: "source"}, {ID: "destination"}}}
	router := transactionTestRouter(NewTransactionHandler(transactions, &transactionWalletRepositoryStub{}, &transactionCategoryRepositoryStub{}))
	request := httptest.NewRequest(http.MethodPatch, "/transactions/transfer/transfer-1", bytes.NewBufferString(`{"amount":90000,"occurred_at":"2026-09-22T10:00:00Z"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest || transactions.mutationCalls != 0 {
		t.Fatalf("missing idempotency key should reject without mutation: status=%d calls=%d body=%s", response.Code, transactions.mutationCalls, response.Body.String())
	}
}

func TestTransferMutationChangedReplayReturnsConflict(t *testing.T) {
	transactions := &transactionRepositoryStub{mutationErr: repository.ErrMutationConflict}
	router := transactionTestRouter(NewTransactionHandler(transactions, &transactionWalletRepositoryStub{}, &transactionCategoryRepositoryStub{}))
	request := httptest.NewRequest(http.MethodPatch, "/transactions/transfer/transfer-1", bytes.NewBufferString(`{"amount":90001,"occurred_at":"2026-09-22T10:00:00Z"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "transfer-update-1")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusConflict {
		t.Fatalf("changed replay should return conflict: status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestTransferMutationReplayMarksResponse(t *testing.T) {
	transactions := &transactionRepositoryStub{transferUpdated: []entity.Transaction{{ID: "source"}, {ID: "destination"}}, mutationReplay: true}
	router := transactionTestRouter(NewTransactionHandler(transactions, &transactionWalletRepositoryStub{}, &transactionCategoryRepositoryStub{}))
	request := httptest.NewRequest(http.MethodPatch, "/transactions/transfer/transfer-1", bytes.NewBufferString(`{"amount":90000,"occurred_at":"2026-09-22T10:00:00Z"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "transfer-update-1")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK || response.Header().Get("Idempotent-Replayed") != "true" {
		t.Fatalf("replay should return the stored response marker: status=%d replay=%q body=%s", response.Code, response.Header().Get("Idempotent-Replayed"), response.Body.String())
	}
}

func TestSingleRowUpdateRejectsLinkedTransfer(t *testing.T) {
	transferID := "transfer-1"
	existing := &entity.Transaction{ID: "source", OwnerID: "owner-1", WalletID: "wallet-1", Type: entity.TransactionTypeExpense, Amount: 90000, OccurredAt: time.Now().UTC(), TransferID: &transferID}
	transactions := &transactionRepositoryStub{existing: existing, updated: existing}
	wallets := &transactionWalletRepositoryStub{wallets: map[string]entity.Wallet{"wallet-1": {ID: "wallet-1", OwnerID: "owner-1", Type: entity.WalletTypeBasic}}}
	router := transactionTestRouter(NewTransactionHandler(transactions, wallets, &transactionCategoryRepositoryStub{}))
	request := httptest.NewRequest(http.MethodPatch, "/transactions/source", bytes.NewBufferString(`{"wallet_id":"wallet-1","type":"expense","amount":90000}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("linked row should require pair mutation, status=%d body=%s", response.Code, response.Body.String())
	}
}

func stringPtr(value string) *string { return &value }

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
