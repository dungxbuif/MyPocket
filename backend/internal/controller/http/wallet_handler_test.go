package httpapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mypocket/backend/internal/entity"
	walletrepo "github.com/mypocket/backend/internal/repository"
)

type walletRepositoryStub struct {
	created          *entity.Wallet
	deleted          string
	updates          map[string]any
	found            *entity.Wallet
	idempotencyKey   string
	idempotencyCalls int
	idempotencyErr   error
}

// walletRepositoryWithoutMutation models a legacy adapter that cannot safely
// honor a caller-supplied idempotency key. The handler must fail closed rather
// than silently downgrade the write to a non-idempotent create.
type walletRepositoryWithoutMutation struct{}

func (walletRepositoryWithoutMutation) List(string) ([]entity.Wallet, error) { return nil, nil }
func (walletRepositoryWithoutMutation) Find(_, id string) (*entity.Wallet, error) {
	return &entity.Wallet{ID: id, Type: entity.WalletTypeBasic}, nil
}
func (walletRepositoryWithoutMutation) Create(*entity.Wallet) error { return nil }
func (walletRepositoryWithoutMutation) Update(_, id string, _ map[string]any) (*entity.Wallet, error) {
	return &entity.Wallet{ID: id}, nil
}
func (walletRepositoryWithoutMutation) Delete(_, _ string) error { return nil }

func (s *walletRepositoryStub) List(ownerID string) ([]entity.Wallet, error) { return nil, nil }
func (s *walletRepositoryStub) Find(ownerID, id string) (*entity.Wallet, error) {
	if s.found != nil {
		return s.found, nil
	}
	return &entity.Wallet{ID: id, OwnerID: ownerID, Type: entity.WalletTypeBasic}, nil
}
func (s *walletRepositoryStub) Create(wallet *entity.Wallet) error { s.created = wallet; return nil }
func (s *walletRepositoryStub) Update(ownerID, id string, updates map[string]any) (*entity.Wallet, error) {
	s.updates = updates
	return &entity.Wallet{ID: id, OwnerID: ownerID}, nil
}
func (s *walletRepositoryStub) Delete(ownerID, id string) error { s.deleted = id; return nil }
func (s *walletRepositoryStub) CreateIdempotent(wallet *entity.Wallet, key, requestHash string) (*entity.Wallet, bool, error) {
	s.idempotencyCalls++
	s.idempotencyKey = key
	if s.idempotencyErr != nil {
		return nil, false, s.idempotencyErr
	}
	s.created = wallet
	return wallet, false, nil
}

func walletTestRouter(handler *WalletHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set(contextUserIDKey, "owner-1"); c.Next() })
	router.POST("/wallets", handler.CreateWallet)
	router.PATCH("/wallets/:id", handler.UpdateWallet)
	router.DELETE("/wallets/:id", handler.DeleteWallet)
	return router
}

func TestUpdateWalletKeepsExistingType(t *testing.T) {
	stub := &walletRepositoryStub{}
	router := walletTestRouter(NewWalletHandler(stub))
	request := httptest.NewRequest(http.MethodPatch, "/wallets/wallet-1", bytes.NewBufferString(`{"name":"Quỹ mới","type":"goal","opening_balance":100000,"target_amount":500000}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if _, exists := stub.updates["type"]; exists {
		t.Fatalf("wallet type must be immutable: %#v", stub.updates)
	}
}

func TestUpdateWalletDoesNotOverwriteOpeningBalance(t *testing.T) {
	stub := &walletRepositoryStub{}
	router := walletTestRouter(NewWalletHandler(stub))
	request := httptest.NewRequest(http.MethodPatch, "/wallets/wallet-1", bytes.NewBufferString(`{"name":"Tiền mặt","type":"basic","opening_balance":999999}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if _, exists := stub.updates["opening_balance"]; exists {
		t.Fatalf("opening balance must be immutable during metadata update: %#v", stub.updates)
	}
}

func TestCreateWalletScopesOwnerAndSeedsVND(t *testing.T) {
	stub := &walletRepositoryStub{}
	router := walletTestRouter(NewWalletHandler(stub))
	request := httptest.NewRequest(http.MethodPost, "/wallets", bytes.NewBufferString(`{"name":"Tiền mặt","type":"basic","opening_balance":250000}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if stub.created == nil || stub.created.OwnerID != "owner-1" || stub.created.Currency != entity.WalletCurrencyVND {
		t.Fatalf("unexpected wallet: %#v", stub.created)
	}
}

func TestCreateWalletWithIdempotencyKeyUsesDurableMutationAdapter(t *testing.T) {
	stub := &walletRepositoryStub{}
	router := walletTestRouter(NewWalletHandler(stub))
	request := httptest.NewRequest(http.MethodPost, "/wallets", bytes.NewBufferString(`{"name":"Tiền mặt","type":"basic","opening_balance":250000}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "wallet-confirm-1")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusCreated || stub.idempotencyCalls != 1 || stub.idempotencyKey != "wallet-confirm-1" {
		t.Fatalf("status=%d calls=%d key=%q body=%s", response.Code, stub.idempotencyCalls, stub.idempotencyKey, response.Body.String())
	}
}

func TestCreateWalletWithIdempotencyKeyFailsClosedWithoutDurableAdapter(t *testing.T) {
	router := walletTestRouter(NewWalletHandler(walletRepositoryWithoutMutation{}))
	request := httptest.NewRequest(http.MethodPost, "/wallets", bytes.NewBufferString(`{"name":"Tiền mặt","type":"basic"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "wallet-confirm-legacy")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestCreateWalletRejectsNonASCIIIdempotencyKey(t *testing.T) {
	router := walletTestRouter(NewWalletHandler(&walletRepositoryStub{}))
	request := httptest.NewRequest(http.MethodPost, "/wallets", bytes.NewBufferString(`{"name":"Tiền mặt","type":"basic"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "khóa-wallet")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestCreateWalletRejectsMultipleIdempotencyKeys(t *testing.T) {
	router := walletTestRouter(NewWalletHandler(&walletRepositoryStub{}))
	request := httptest.NewRequest(http.MethodPost, "/wallets", bytes.NewBufferString(`{"name":"Tiền mặt","type":"basic"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Add("Idempotency-Key", "wallet-one")
	request.Header.Add("Idempotency-Key", "wallet-two")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestCreateWalletMapsIdempotencyConflict(t *testing.T) {
	stub := &walletRepositoryStub{idempotencyErr: walletrepo.ErrMutationConflict}
	router := walletTestRouter(NewWalletHandler(stub))
	request := httptest.NewRequest(http.MethodPost, "/wallets", bytes.NewBufferString(`{"name":"Tiền mặt","type":"basic"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "wallet-confirm-1")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusConflict {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestCreateWalletRejectsUnknownType(t *testing.T) {
	router := walletTestRouter(NewWalletHandler(&walletRepositoryStub{}))
	request := httptest.NewRequest(http.MethodPost, "/wallets", bytes.NewBufferString(`{"name":"Ví lạ","type":"shared"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}

func TestCreateGoalWalletRequiresPositiveTarget(t *testing.T) {
	router := walletTestRouter(NewWalletHandler(&walletRepositoryStub{}))
	request := httptest.NewRequest(http.MethodPost, "/wallets", bytes.NewBufferString(`{"name":"Quỹ dự phòng","type":"goal"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}

func TestCreateGoalWalletPersistsNotificationPreference(t *testing.T) {
	stub := &walletRepositoryStub{}
	router := walletTestRouter(NewWalletHandler(stub))
	request := httptest.NewRequest(http.MethodPost, "/wallets", bytes.NewBufferString(`{"name":"Quỹ dự phòng","type":"goal","target_amount":1000000,"goal_notifications_enabled":true}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusCreated || stub.created == nil || !stub.created.GoalNotificationsEnabled {
		t.Fatalf("goal notification preference not persisted: status=%d body=%s wallet=%#v", response.Code, response.Body.String(), stub.created)
	}
}

func TestCreateBasicWalletIgnoresGoalNotificationPreference(t *testing.T) {
	stub := &walletRepositoryStub{}
	router := walletTestRouter(NewWalletHandler(stub))
	request := httptest.NewRequest(http.MethodPost, "/wallets", bytes.NewBufferString(`{"name":"Tiền mặt","type":"basic","goal_notifications_enabled":true}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusCreated || stub.created == nil || stub.created.GoalNotificationsEnabled {
		t.Fatalf("non-goal wallet must not enable goal notifications: status=%d body=%s wallet=%#v", response.Code, response.Body.String(), stub.created)
	}
}

func TestCreateCreditWalletRequiresPositiveLimit(t *testing.T) {
	router := walletTestRouter(NewWalletHandler(&walletRepositoryStub{}))
	request := httptest.NewRequest(http.MethodPost, "/wallets", bytes.NewBufferString(`{"name":"Thẻ tín dụng","type":"credit","credit_limit":0}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}

func TestCreateCreditWalletPersistsStatementCycle(t *testing.T) {
	stub := &walletRepositoryStub{}
	router := walletTestRouter(NewWalletHandler(stub))
	request := httptest.NewRequest(http.MethodPost, "/wallets", bytes.NewBufferString(`{"name":"Thẻ tín dụng","type":"credit","credit_limit":1000000,"last_statement_balance":250000,"statement_day":15,"payment_due_day":5}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusCreated || stub.created == nil {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if stub.created.LastStatementBalance == nil || *stub.created.LastStatementBalance != 250000 || stub.created.StatementDay == nil || *stub.created.StatementDay != 15 || stub.created.PaymentDueDay == nil || *stub.created.PaymentDueDay != 5 {
		t.Fatalf("cycle fields not persisted: %#v", stub.created)
	}
}

func TestCreateCreditWalletRejectsInvalidStatementDay(t *testing.T) {
	router := walletTestRouter(NewWalletHandler(&walletRepositoryStub{}))
	request := httptest.NewRequest(http.MethodPost, "/wallets", bytes.NewBufferString(`{"name":"Thẻ tín dụng","type":"credit","credit_limit":1000000,"statement_day":32}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}

func TestDeleteWalletUsesOwnerScopedRepository(t *testing.T) {
	stub := &walletRepositoryStub{}
	router := walletTestRouter(NewWalletHandler(stub))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodDelete, "/wallets/wallet-1", nil))
	if response.Code != http.StatusNoContent || stub.deleted != "wallet-1" {
		t.Fatalf("status = %d, deleted = %q", response.Code, stub.deleted)
	}
}

func TestGoalTargetDate(t *testing.T) {
	for _, tc := range []struct {
		date   string
		status int
	}{{"2026-12-31", 201}, {"2026-02-30", 400}, {"bad", 400}} {
		stub := &walletRepositoryStub{}
		r := walletTestRouter(NewWalletHandler(stub))
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/wallets", bytes.NewBufferString(`{"name":"Goal","type":"goal","target_amount":1000,"target_date":"`+tc.date+`"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != tc.status {
			t.Fatalf("%s: status %d", tc.date, w.Code)
		}
		if tc.status == 201 && (stub.created.TargetDate == nil || stub.created.TargetDate.Format("2006-01-02") != tc.date) {
			t.Fatal("target date not persisted")
		}
	}
}

func TestGoalTargetDateUpdateOmitAndClear(t *testing.T) {
	for _, tc := range []struct {
		suffix  string
		changed bool
	}{{"", false}, {`,"target_date":""`, true}} {
		stub := &walletRepositoryStub{found: &entity.Wallet{ID: "goal", Type: entity.WalletTypeGoal}}
		r := walletTestRouter(NewWalletHandler(stub))
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPatch, "/wallets/goal", bytes.NewBufferString(`{"name":"Goal","target_amount":1000`+tc.suffix+`}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("status %d", w.Code)
		}
		value, present := stub.updates["target_date"]
		if present != tc.changed || value != nil {
			t.Fatalf("unexpected date update %#v", stub.updates)
		}
	}
}
