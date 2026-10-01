package httpapi

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
)

type travelRepositoryStub struct {
	events   map[string]entity.TravelEvent
	activeID string
	linked   *string
}

func (s *travelRepositoryStub) List(string) ([]entity.TravelEvent, error) {
	rows := make([]entity.TravelEvent, 0, len(s.events))
	for _, row := range s.events {
		rows = append(rows, row)
	}
	return rows, nil
}
func (s *travelRepositoryStub) Find(_ string, id string) (*entity.TravelEvent, error) {
	row, ok := s.events[id]
	if !ok {
		return nil, repository.ErrTravelNotFound
	}
	return &row, nil
}
func (s *travelRepositoryStub) Active(_ string) (*entity.TravelEvent, error) {
	if s.activeID == "" {
		return nil, nil
	}
	row, ok := s.events[s.activeID]
	if !ok {
		return nil, repository.ErrTravelNotFound
	}
	return &row, nil
}
func (s *travelRepositoryStub) Create(_ string, event *entity.TravelEvent) error {
	s.events[event.ID] = *event
	return nil
}
func (s *travelRepositoryStub) Update(_ string, id string, updates repository.TravelUpdate) (*entity.TravelEvent, error) {
	row, ok := s.events[id]
	if !ok {
		return nil, repository.ErrTravelNotFound
	}
	row.Name, row.Context, row.StartsOn, row.EndsOn = updates.Name, updates.Context, updates.StartsOn, updates.EndsOn
	s.events[id] = row
	return &row, nil
}
func (s *travelRepositoryStub) Delete(_ string, id string) error {
	if _, ok := s.events[id]; !ok {
		return repository.ErrTravelNotFound
	}
	delete(s.events, id)
	if s.activeID == id {
		s.activeID = ""
	}
	return nil
}
func (s *travelRepositoryStub) SetActive(_ string, id string, active bool) (*entity.TravelEvent, error) {
	row, ok := s.events[id]
	if !ok {
		return nil, repository.ErrTravelNotFound
	}
	for key, item := range s.events {
		item.Active = false
		s.events[key] = item
	}
	row.Active = active
	s.events[id] = row
	if active {
		s.activeID = id
	} else {
		s.activeID = ""
	}
	return &row, nil
}
func (s *travelRepositoryStub) LinkTransaction(_ string, _ string, eventID *string) (*entity.Transaction, error) {
	s.linked = eventID
	return &entity.Transaction{TravelEventID: eventID}, nil
}

func travelTestRouter(handler *TravelHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set(contextUserIDKey, "owner-1"); c.Next() })
	router.POST("/travel/events", handler.Create)
	router.POST("/travel/events/:id/activate", handler.Activate)
	router.PATCH("/travel/events/:id", handler.Update)
	return router
}

func TestTravelCreateParsesDateAndPreservesInactiveDefault(t *testing.T) {
	repo := &travelRepositoryStub{events: map[string]entity.TravelEvent{}}
	request := httptest.NewRequest(http.MethodPost, "/travel/events", bytes.NewBufferString(`{"name":"Đà Lạt","starts_on":"2026-11-01","ends_on":"2026-11-05"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	travelTestRouter(&TravelHandler{Events: repo}).ServeHTTP(response, request)
	if response.Code != http.StatusCreated || len(repo.events) != 1 {
		t.Fatalf("unexpected create: status=%d body=%s events=%#v", response.Code, response.Body.String(), repo.events)
	}
	for _, event := range repo.events {
		if event.Active || event.StartsOn == nil || event.StartsOn.String() != "2026-11-01" {
			t.Fatalf("unexpected event: %#v", event)
		}
	}
}

func TestTravelActivateSwitchesActiveEvent(t *testing.T) {
	repo := &travelRepositoryStub{events: map[string]entity.TravelEvent{
		"old": {ID: "old", OwnerID: "owner-1", Name: "Cũ", Active: true},
		"new": {ID: "new", OwnerID: "owner-1", Name: "Mới"},
	}, activeID: "old"}
	request := httptest.NewRequest(http.MethodPost, "/travel/events/new/activate", nil)
	response := httptest.NewRecorder()
	travelTestRouter(&TravelHandler{Events: repo}).ServeHTTP(response, request)
	if response.Code != http.StatusOK || repo.activeID != "new" || repo.events["old"].Active || !repo.events["new"].Active {
		t.Fatalf("active invariant failed: status=%d body=%s events=%#v", response.Code, response.Body.String(), repo.events)
	}
}

func TestCreateTransactionAutoAttachesActiveTravelEvent(t *testing.T) {
	transactions := &transactionRepositoryStub{}
	travel := &travelRepositoryStub{events: map[string]entity.TravelEvent{"trip": {ID: "trip", OwnerID: "owner-1", Name: "Chuyến", Active: true}}, activeID: "trip"}
	wallets := &transactionWalletRepositoryStub{wallets: map[string]entity.Wallet{"wallet": {ID: "wallet", OwnerID: "owner-1", Type: entity.WalletTypeBasic}}}
	request := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewBufferString(`{"wallet_id":"wallet","type":"expense","amount":1000,"occurred_at":"2026-10-01T10:00:00+07:00"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router := transactionTestRouter(NewTransactionHandler(transactions, wallets, &transactionCategoryRepositoryStub{}))
	transactionsHandler := NewTransactionHandler(transactions, wallets, &transactionCategoryRepositoryStub{})
	transactionsHandler.Travel = travel
	router = transactionTestRouter(transactionsHandler)
	router.ServeHTTP(response, request)
	if response.Code != http.StatusCreated || transactions.created == nil || transactions.created.TravelEventID == nil || *transactions.created.TravelEventID != "trip" {
		t.Fatalf("active travel event was not attached: status=%d body=%s row=%#v", response.Code, response.Body.String(), transactions.created)
	}
}

func TestTravelLinkStubKeepsClearSemantics(t *testing.T) {
	repo := &travelRepositoryStub{events: map[string]entity.TravelEvent{}}
	if _, err := repo.Find("owner-1", "missing"); !errors.Is(err, repository.ErrTravelNotFound) {
		t.Fatalf("expected missing event error, got %v", err)
	}
	var cleared *string
	if _, err := repo.LinkTransaction("owner-1", "tx", cleared); err != nil || repo.linked != nil {
		t.Fatalf("expected clear link, err=%v linked=%v", err, repo.linked)
	}
}
