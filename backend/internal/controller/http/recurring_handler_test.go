package httpapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mypocket/backend/internal/entity"
	recurringrepo "github.com/mypocket/backend/internal/repository"
)

type recurringRepositoryStub struct {
	created *entity.RecurringSchedule
	rows    []entity.RecurringSchedule
	row     *entity.RecurringSchedule
	err     error
	runs    int
}

func (s *recurringRepositoryStub) List(string) ([]entity.RecurringSchedule, error) {
	return s.rows, s.err
}
func (s *recurringRepositoryStub) Find(string, string) (*entity.RecurringSchedule, error) {
	return s.row, s.err
}
func (s *recurringRepositoryStub) Create(_ string, schedule *entity.RecurringSchedule) error {
	if s.err == nil {
		s.created = schedule
	}
	return s.err
}
func (s *recurringRepositoryStub) Update(string, string, recurringrepo.RecurringUpdate) (*entity.RecurringSchedule, error) {
	return s.row, s.err
}
func (s *recurringRepositoryStub) Delete(string, string) error           { return s.err }
func (s *recurringRepositoryStub) RunDue(string, time.Time) (int, error) { s.runs++; return 2, s.err }

func recurringTestRouter(handler *RecurringHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set(contextUserIDKey, "owner-1"); c.Next() })
	router.GET("/recurring", handler.List)
	router.POST("/recurring", handler.Create)
	router.POST("/recurring/run-due", handler.RunDue)
	return router
}

func TestCreateRecurringScheduleUsesAccountLocalAnchor(t *testing.T) {
	repo := &recurringRepositoryStub{}
	wallets := &transactionWalletRepositoryStub{wallets: map[string]entity.Wallet{"wallet-1": {ID: "wallet-1", OwnerID: "owner-1", Type: entity.WalletTypeBasic}}}
	users := &transactionUserRepositoryStub{user: entity.User{ID: "owner-1", Timezone: "Asia/Ho_Chi_Minh"}}
	handler := &RecurringHandler{Schedules: repo, Wallets: wallets, Categories: &transactionCategoryRepositoryStub{}, Users: users}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/recurring", bytes.NewBufferString(`{"name":"Tiền nhà","wallet_id":"wallet-1","type":"expense","amount":5000000,"frequency":"monthly","interval":1,"next_run_at":"2026-01-31T09:30:00+07:00"}`))
	request.Header.Set("Content-Type", "application/json")
	recurringTestRouter(handler).ServeHTTP(response, request)
	if response.Code != http.StatusCreated || repo.created == nil {
		t.Fatalf("status=%d body=%s schedule=%#v", response.Code, response.Body.String(), repo.created)
	}
	if repo.created.AnchorDay != 31 || repo.created.AnchorMonth != 1 || repo.created.Active != true {
		t.Fatalf("unexpected recurring anchor: %#v", repo.created)
	}
}

func TestCreateRecurringRejectsCreditWallet(t *testing.T) {
	repo := &recurringRepositoryStub{}
	wallets := &transactionWalletRepositoryStub{wallets: map[string]entity.Wallet{"credit-1": {ID: "credit-1", OwnerID: "owner-1", Type: entity.WalletTypeCredit}}}
	handler := &RecurringHandler{Schedules: repo, Wallets: wallets, Categories: &transactionCategoryRepositoryStub{}, Users: &transactionUserRepositoryStub{user: entity.User{ID: "owner-1", Timezone: "Asia/Ho_Chi_Minh"}}}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/recurring", bytes.NewBufferString(`{"name":"Thẻ","wallet_id":"credit-1","type":"expense","amount":100,"frequency":"monthly"}`))
	request.Header.Set("Content-Type", "application/json")
	recurringTestRouter(handler).ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest || repo.created != nil {
		t.Fatalf("credit schedule should reject: status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestRunDueReturnsCreatedCount(t *testing.T) {
	repo := &recurringRepositoryStub{}
	handler := &RecurringHandler{Schedules: repo}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/recurring/run-due", nil)
	recurringTestRouter(handler).ServeHTTP(response, request)
	if response.Code != http.StatusOK || repo.runs != 1 || !bytes.Contains(response.Body.Bytes(), []byte(`"created":2`)) {
		t.Fatalf("unexpected run response: status=%d body=%s", response.Code, response.Body.String())
	}
}
