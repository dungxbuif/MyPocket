package httpapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mypocket/backend/internal/entity"
)

type walletRepositoryStub struct {
	created *entity.Wallet
	deleted string
	updates map[string]any
	found   *entity.Wallet
}

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
