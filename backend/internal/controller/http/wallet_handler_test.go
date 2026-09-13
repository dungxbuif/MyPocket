package httpapi

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mypocket/backend/internal/entity"
)

type walletRepositoryStub struct {
	created *entity.Wallet
	deleted string
}

func (s *walletRepositoryStub) List(ownerID string) ([]entity.Wallet, error) { return nil, nil }
func (s *walletRepositoryStub) Find(ownerID, id string) (*entity.Wallet, error) {
	return nil, errors.New("not found")
}
func (s *walletRepositoryStub) Create(wallet *entity.Wallet) error { s.created = wallet; return nil }
func (s *walletRepositoryStub) Update(ownerID, id string, updates map[string]any) (*entity.Wallet, error) {
	return nil, errors.New("not found")
}
func (s *walletRepositoryStub) Delete(ownerID, id string) error { s.deleted = id; return nil }

func walletTestRouter(handler *WalletHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set(contextUserIDKey, "owner-1"); c.Next() })
	router.POST("/wallets", handler.CreateWallet)
	router.DELETE("/wallets/:id", handler.DeleteWallet)
	return router
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

func TestDeleteWalletUsesOwnerScopedRepository(t *testing.T) {
	stub := &walletRepositoryStub{}
	router := walletTestRouter(NewWalletHandler(stub))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodDelete, "/wallets/wallet-1", nil))
	if response.Code != http.StatusNoContent || stub.deleted != "wallet-1" {
		t.Fatalf("status = %d, deleted = %q", response.Code, stub.deleted)
	}
}
