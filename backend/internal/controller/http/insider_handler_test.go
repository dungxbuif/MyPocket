package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mypocket/backend/internal/entity"
)

type insiderReaderStub struct {
	owner, month, wallet, category string
}

func (s *insiderReaderStub) ReadInsider(ownerID, month, walletID, categoryID string) (*entity.InsiderSummary, error) {
	s.owner, s.month, s.wallet, s.category = ownerID, month, walletID, categoryID
	return &entity.InsiderSummary{Month: month, Timezone: "Asia/Ho_Chi_Minh"}, nil
}

func TestMoneyInsiderKeepsOwnerAndFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &insiderReaderStub{}
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set(contextUserIDKey, "owner-1"); c.Next() })
	router.GET("/reports/insider", (&InsiderHandler{Reader: stub}).Get)
	record := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/reports/insider?month=2026-09&wallet_id=wallet-1&category_id=cat-1", nil)
	router.ServeHTTP(record, request)
	if record.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", record.Code, record.Body.String())
	}
	if stub.owner != "owner-1" || stub.month != "2026-09" || stub.wallet != "wallet-1" || stub.category != "cat-1" {
		t.Fatalf("unexpected reader scope: %#v", stub)
	}
}

func TestMoneyInsiderRejectsInvalidMonth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set(contextUserIDKey, "owner-1"); c.Next() })
	router.GET("/reports/insider", (&InsiderHandler{Reader: &insiderReaderStub{}}).Get)
	record := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/reports/insider?month=not-a-month", nil)
	router.ServeHTTP(record, request)
	if record.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", record.Code)
	}
}
