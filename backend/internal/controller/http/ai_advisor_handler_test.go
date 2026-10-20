package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
	"github.com/mypocket/backend/internal/usecase"
)

type advisorActionStore struct {
	messages []entity.AdvisorMessage
}

func (s *advisorActionStore) StartRun(context.Context, repository.AdvisorRunInput) (*entity.AdvisorRun, *entity.AdvisorMessage, bool, error) {
	return nil, nil, false, repository.ErrAdvisorBusy
}
func (s *advisorActionStore) MarkRunStatus(context.Context, string, string, string, time.Time) error {
	return nil
}
func (s *advisorActionStore) GetRun(context.Context, string, string) (*entity.AdvisorRun, error) {
	return nil, repository.ErrNotFound
}
func (s *advisorActionStore) GetRunByRequest(context.Context, string, string) (*entity.AdvisorRun, error) {
	return nil, repository.ErrNotFound
}
func (s *advisorActionStore) ListMessages(_ context.Context, ownerID, conversationID string, _ int64, _ int) ([]entity.AdvisorMessage, error) {
	for _, message := range s.messages {
		if message.ConversationID == conversationID && ownerID == "owner-1" {
			return []entity.AdvisorMessage{message}, nil
		}
	}
	return nil, repository.ErrNotFound
}
func (s *advisorActionStore) AppendEvent(context.Context, string, string, string, any) (*entity.AdvisorEvent, error) {
	return nil, nil
}
func (s *advisorActionStore) CreateAssistantMessage(context.Context, string, string, int64, []entity.AdvisorPart) (*entity.AdvisorMessage, error) {
	return nil, nil
}

type advisorOverviewUsers struct{}

func (advisorOverviewUsers) FindByEmail(string) (*entity.User, error) {
	return nil, repository.ErrNotFound
}
func (advisorOverviewUsers) FindByID(string) (*entity.User, error) {
	return &entity.User{ID: "owner-1", Timezone: "Asia/Ho_Chi_Minh"}, nil
}
func (advisorOverviewUsers) FindOrCreateGoogleUser(entity.GoogleProfile) (*entity.User, error) {
	return nil, repository.ErrNotFound
}
func (advisorOverviewUsers) SetTimezone(string, string, bool) (*entity.User, error) {
	return nil, repository.ErrNotFound
}
func (advisorOverviewUsers) Create(*entity.User) error { return nil }

type advisorOverviewReader struct{}

func (advisorOverviewReader) ReadBundle(context.Context, string, []entity.NormalizedQuery) (entity.FactBundle, error) {
	return entity.FactBundle{Results: []entity.FinanceResult{{Status: "ok", ViewKind: "finance_summary", View: []byte(`{"income":1000,"expense":150,"net":850,"count":3}`), Source: entity.SourceScope{Timezone: "Asia/Ho_Chi_Minh"}}}}, nil
}

type advisorJarRepositoryStub struct {
	created        *entity.JarMonthConfig
	updated        *entity.JarMonthConfig
	idempotencyKey string
}

type advisorBudgetRepositoryStub struct {
	created        *entity.Budget
	updated        *entity.Budget
	idempotencyKey string
}

type advisorTransactionRepositoryStub struct {
	existing *entity.Transaction
	updated  map[string]any
}

func (s *advisorTransactionRepositoryStub) List(string) ([]entity.Transaction, error) {
	return nil, nil
}
func (s *advisorTransactionRepositoryStub) Find(ownerID, id string) (*entity.Transaction, error) {
	if s.existing == nil || s.existing.OwnerID != ownerID || s.existing.ID != id {
		return nil, repository.ErrNotFound
	}
	return s.existing, nil
}
func (s *advisorTransactionRepositoryStub) Create(*entity.Transaction) error           { return nil }
func (s *advisorTransactionRepositoryStub) CreateAdjustment(*entity.Transaction) error { return nil }
func (s *advisorTransactionRepositoryStub) BulkDelete(string, []string) error          { return nil }
func (s *advisorTransactionRepositoryStub) Update(ownerID, id string, updates map[string]any) (*entity.Transaction, error) {
	if _, err := s.Find(ownerID, id); err != nil {
		return nil, err
	}
	s.updated = updates
	return s.existing, nil
}
func (s *advisorTransactionRepositoryStub) Delete(string, string) error { return nil }

func advisorJarLinkedTransactionView(action, actionID, transactionID, categoryID string) map[string]any {
	return map[string]any{
		"action_id": actionID, "action": action, "conversation_id": "conversation-1", "generation": int64(4),
		"transaction_id": transactionID, "category_id": categoryID, "wallet_id": "22222222-2222-4222-8222-222222222222",
		"type": entity.TransactionTypeExpense, "amount": int64(100), "occurred_at": "2026-09-10T10:00:00Z", "expected_version": int64(1),
	}
}

func advisorJarLinkedTransactionDependencies(transactionID, categoryID, jarID string) (*advisorTransactionRepositoryStub, *transactionWalletRepositoryStub, *transactionCategoryRepositoryStub, *transactionUserRepositoryStub, *transactionJarRepositoryStub) {
	return &advisorTransactionRepositoryStub{existing: &entity.Transaction{
			ID: transactionID, OwnerID: "owner-1", WalletID: "22222222-2222-4222-8222-222222222222", CategoryID: &categoryID,
			JarID: &jarID, Type: entity.TransactionTypeExpense, Amount: 100, OccurredAt: time.Date(2026, 9, 10, 10, 0, 0, 0, time.UTC), Version: 1,
		}}, &transactionWalletRepositoryStub{wallets: map[string]entity.Wallet{
			"22222222-2222-4222-8222-222222222222": {ID: "22222222-2222-4222-8222-222222222222", OwnerID: "owner-1", Type: entity.WalletTypeBasic},
		}}, &transactionCategoryRepositoryStub{categories: []entity.Category{
			{ID: categoryID, Kind: entity.TransactionTypeExpense, IsActive: true},
		}}, &transactionUserRepositoryStub{user: entity.User{ID: "owner-1", Timezone: "Asia/Ho_Chi_Minh"}}, &transactionJarRepositoryStub{}
}

func (s *advisorBudgetRepositoryStub) List(string) ([]entity.Budget, error) { return nil, nil }
func (s *advisorBudgetRepositoryStub) Save(owner string, budget *entity.Budget, creating bool) error {
	budget.OwnerID = owner
	if creating {
		s.created = budget
	} else {
		s.updated = budget
	}
	return nil
}
func (s *advisorBudgetRepositoryStub) Delete(string, string) error { return nil }
func (s *advisorBudgetRepositoryStub) CreateIdempotent(owner string, budget *entity.Budget, key, _ string) (*entity.Budget, bool, error) {
	s.idempotencyKey = key
	if s.created != nil {
		return s.created, true, nil
	}
	if err := s.Save(owner, budget, true); err != nil {
		return nil, false, err
	}
	return s.created, false, nil
}

func (s *advisorJarRepositoryStub) ListMonth(string, string, string) (*entity.JarMonthSummary, error) {
	return &entity.JarMonthSummary{}, nil
}
func (s *advisorJarRepositoryStub) Create(owner, month, _ string, name, mode string, amount *int64, percent *int) (*entity.JarMonthConfig, error) {
	s.created = &entity.JarMonthConfig{OwnerID: owner, Name: name, AllocationMode: mode, AllocationAmount: amount, AllocationPercentBPS: percent}
	return s.created, nil
}
func (s *advisorJarRepositoryStub) UpdateMonthConfig(owner, jarID, month, name, mode string, amount *int64, percent *int) (*entity.JarMonthConfig, error) {
	parsedMonth, err := entity.ParseCalendarDate(month + "-01")
	if err != nil {
		return nil, err
	}
	s.updated = &entity.JarMonthConfig{OwnerID: owner, JarID: jarID, Month: parsedMonth, Name: name, AllocationMode: mode, AllocationAmount: amount, AllocationPercentBPS: percent, Active: true}
	return s.updated, nil
}
func (s *advisorJarRepositoryStub) RemoveMonthConfig(string, string, string) error { return nil }
func (s *advisorJarRepositoryStub) FindMonthConfig(string, string, string, bool) (*entity.JarMonthConfig, error) {
	return nil, repository.ErrNotFound
}
func (s *advisorJarRepositoryStub) Cumulative(string, string, string, string, string) (*entity.JarCumulativeSummary, error) {
	return nil, repository.ErrNotFound
}
func (s *advisorJarRepositoryStub) CreateIdempotent(owner, month, timezone, name, mode string, amount *int64, percent *int, key, _ string) (*entity.JarMonthConfig, bool, error) {
	s.idempotencyKey = key
	if s.created != nil {
		return s.created, true, nil
	}
	created, err := s.Create(owner, month, timezone, name, mode, amount, percent)
	return created, false, err
}

func TestAdvisorRoutesExposeConversationSurface(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := &Router{Engine: gin.New(), AuthMiddleware: &AuthMiddleware{}}
	router.RegisterAdvisorRoutes(&AdvisorHandler{})
	routes := map[string]bool{}
	for _, route := range router.Engine.Routes() {
		routes[route.Method+" "+route.Path] = true
	}
	for _, expected := range []string{
		"GET /api/v1/ai/advisor/capabilities",
		"GET /api/v1/ai/advisor/overview",
		"POST /api/v1/ai/advisor/messages",
		"POST /api/v1/ai/advisor/messages/async",
		"GET /api/v1/ai/advisor/conversation/messages",
		"POST /api/v1/ai/advisor/conversation/reset",
		"GET /api/v1/ai/advisor/runs/:id",
		"GET /api/v1/ai/advisor/runs/:id/events",
		"POST /api/v1/ai/advisor/runs/:id/cancel",
		"POST /api/v1/ai/advisor/actions/:action_id/confirm",
	} {
		if !routes[expected] {
			t.Fatalf("missing route %s", expected)
		}
	}
}

func TestAdvisorCapabilitiesAdvertiseBoundedWriteActionContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/capabilities", (&AdvisorHandler{}).Capabilities)
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/capabilities", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var envelope struct {
		Data struct {
			Enabled              bool     `json:"enabled"`
			Write                bool     `json:"write"`
			DirectWrites         bool     `json:"direct_writes"`
			ProposalActions      bool     `json:"proposal_actions"`
			ConfirmationRequired bool     `json:"confirmation_required"`
			RunTimeoutSeconds    int      `json:"run_timeout_seconds"`
			GroundingVersion     int      `json:"grounding_version"`
			Tools                []string `json:"tools"`
			Actions              []string `json:"actions"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Data.Enabled || envelope.Data.Write {
		t.Fatalf("unconfigured advisor must remain disabled and non-mutating: %+v", envelope.Data)
	}
	if envelope.Data.DirectWrites || !envelope.Data.ProposalActions || !envelope.Data.ConfirmationRequired {
		t.Fatalf("advisor proposal contract drifted: %+v", envelope.Data)
	}
	if envelope.Data.RunTimeoutSeconds != 240 || envelope.Data.GroundingVersion != 1 {
		t.Fatalf("advisor lifecycle metadata drifted: %+v", envelope.Data)
	}
	wantTools := []string{"search_transactions", "get_transaction", "get_finance_summary", "compare_spending_periods", "get_wallet_balances", "get_budget_progress", "get_goal_progress", "get_jar_progress", "get_product_help", "propose_create_wallet", "propose_create_jar", "propose_update_jar", "propose_create_budget", "propose_update_budget", "propose_create_transaction", "propose_update_transaction", "propose_delete_transaction", "propose_transfer_wallet", "propose_create_category", "propose_update_category", "propose_recategorize_transaction"}
	wantActions := []string{"create_wallet", "create_jar", "update_jar", "create_budget", "update_budget", "create_transaction", "update_transaction", "delete_transaction", "transfer_wallet", "create_category", "update_category", "recategorize_transaction"}
	if !reflect.DeepEqual(envelope.Data.Tools, wantTools) {
		t.Fatalf("capability tool set drifted: got=%v want=%v", envelope.Data.Tools, wantTools)
	}
	if !reflect.DeepEqual(envelope.Data.Actions, wantActions) {
		t.Fatalf("capability action set drifted: got=%v want=%v", envelope.Data.Actions, wantActions)
	}
}

type advisorEventStoreStub struct {
	advisorActionStore
	run    *entity.AdvisorRun
	events []entity.AdvisorEvent
}

func (s *advisorEventStoreStub) GetRun(_ context.Context, ownerID, runID string) (*entity.AdvisorRun, error) {
	if s.run == nil || ownerID != s.run.OwnerID || runID != s.run.ID {
		return nil, repository.ErrNotFound
	}
	return s.run, nil
}

func (s *advisorEventStoreStub) ListEvents(_ context.Context, ownerID, runID string, afterSeq int64, _ int) ([]entity.AdvisorEvent, error) {
	if s.run == nil || ownerID != s.run.OwnerID || runID != s.run.ID {
		return nil, repository.ErrNotFound
	}
	result := make([]entity.AdvisorEvent, 0, len(s.events))
	for _, event := range s.events {
		if event.Seq > afterSeq {
			result = append(result, event)
		}
	}
	return result, nil
}

func TestAdvisorRunEventsReplaysOwnerScopedSSE(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &advisorEventStoreStub{
		run:    &entity.AdvisorRun{ID: "run-1", OwnerID: "owner-1", Status: entity.AdvisorStatusCompleted},
		events: []entity.AdvisorEvent{{RunID: "run-1", Seq: 1, Type: "run.completed", Payload: map[string]any{"text": "ok"}}},
	}
	engine := gin.New()
	engine.Use(func(c *gin.Context) { c.Set(contextUserIDKey, "owner-1"); c.Next() })
	engine.GET("/runs/:id/events", (&AdvisorHandler{Service: &usecase.AdvisorService{Store: store}}).RunEvents)
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/runs/run-1/events?after_seq=0", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if got := recorder.Header().Get("Content-Type"); got != "text/event-stream; charset=utf-8" {
		t.Fatalf("content type=%q", got)
	}
	body := recorder.Body.String()
	if !strings.Contains(body, "id: 1\nevent: run.completed\ndata:") || !strings.Contains(body, `"run_id":"run-1"`) {
		t.Fatalf("unexpected SSE body: %s", body)
	}
}

func TestAdvisorConfirmWalletActionUsesPersistedProposalAndMutationKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &advisorActionStore{messages: []entity.AdvisorMessage{{
		ConversationID: "conversation-1", Generation: 4, Role: entity.AdvisorRoleAssistant,
		Parts: []entity.AdvisorPart{{Type: "wallet_confirmation", Data: map[string]any{"view": map[string]any{
			"action_id": "action-1", "action": "create_wallet", "conversation_id": "conversation-1", "generation": int64(4),
			"name": "Du lịch", "type": "basic", "opening_balance": int64(0), "is_in_total": true,
		}}}},
	}}}
	wallets := &walletRepositoryStub{}
	service := &usecase.AdvisorService{Store: store}
	engine := gin.New()
	engine.Use(func(c *gin.Context) { c.Set(contextUserIDKey, "owner-1"); c.Next() })
	engine.POST("/actions/:action_id/confirm", (&AdvisorHandler{Service: service, Wallets: wallets}).ConfirmAction)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/actions/action-1/confirm", strings.NewReader(`{"conversation_id":"conversation-1","generation":4}`))
	request.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if wallets.idempotencyKey != "action-1" || wallets.created == nil || wallets.created.Name != "Du lịch" {
		t.Fatalf("dispatcher did not use persisted proposal: key=%q wallet=%+v", wallets.idempotencyKey, wallets.created)
	}
}

func TestAdvisorConfirmJarActionUsesPersistedProposalAndMutationKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &advisorActionStore{messages: []entity.AdvisorMessage{{
		ConversationID: "conversation-1", Generation: 4, Role: entity.AdvisorRoleAssistant,
		Parts: []entity.AdvisorPart{{Type: "jar_confirmation", Data: map[string]any{"view": map[string]any{
			"action_id": "jar-action-1", "action": "create_jar", "conversation_id": "conversation-1", "generation": int64(4),
			"month": "2026-09", "name": "Ăn uống", "allocation_mode": "fixed", "allocation_amount": int64(400000),
		}}}},
	}}}
	service := &usecase.AdvisorService{Store: store}
	jars := &advisorJarRepositoryStub{}
	engine := gin.New()
	engine.Use(func(c *gin.Context) { c.Set(contextUserIDKey, "owner-1"); c.Next() })
	engine.POST("/actions/:action_id/confirm", (&AdvisorHandler{Service: service, Users: advisorOverviewUsers{}, Jars: jars}).ConfirmAction)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/actions/jar-action-1/confirm", strings.NewReader(`{"conversation_id":"conversation-1","generation":4}`))
	request.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if jars.idempotencyKey != "jar-action-1" || jars.created == nil || jars.created.Name != "Ăn uống" || jars.created.AllocationAmount == nil || *jars.created.AllocationAmount != 400000 {
		t.Fatalf("dispatcher did not use persisted jar proposal: key=%q jar=%+v", jars.idempotencyKey, jars.created)
	}
	replay := httptest.NewRecorder()
	replayRequest := httptest.NewRequest(http.MethodPost, "/actions/jar-action-1/confirm", strings.NewReader(`{"conversation_id":"conversation-1","generation":4}`))
	replayRequest.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(replay, replayRequest)
	if replay.Code != http.StatusOK {
		t.Fatalf("same jar action must replay without creating a second jar: status=%d body=%s", replay.Code, replay.Body.String())
	}
}

func TestAdvisorConfirmJarUpdateUsesPersistedProposal(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &advisorActionStore{messages: []entity.AdvisorMessage{{
		ConversationID: "conversation-1", Generation: 4, Role: entity.AdvisorRoleAssistant,
		Parts: []entity.AdvisorPart{{Type: "jar_confirmation", Data: map[string]any{"view": map[string]any{
			"action_id": "jar-update-1", "action": "update_jar", "jar_id": "11111111-1111-4111-8111-111111111111", "conversation_id": "conversation-1", "generation": int64(4),
			"month": "2026-09", "name": "Ăn uống mới", "allocation_mode": "percent", "allocation_percent": float64(20),
		}}}},
	}}}
	service := &usecase.AdvisorService{Store: store}
	jars := &advisorJarRepositoryStub{}
	engine := gin.New()
	engine.Use(func(c *gin.Context) { c.Set(contextUserIDKey, "owner-1"); c.Next() })
	engine.POST("/actions/:action_id/confirm", (&AdvisorHandler{Service: service, Users: advisorOverviewUsers{}, Jars: jars}).ConfirmAction)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/actions/jar-update-1/confirm", strings.NewReader(`{"conversation_id":"conversation-1","generation":4}`))
	request.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK || jars.updated == nil || jars.updated.JarID != "11111111-1111-4111-8111-111111111111" || jars.updated.Name != "Ăn uống mới" || jars.updated.AllocationPercentBPS == nil || *jars.updated.AllocationPercentBPS != 2000 {
		t.Fatalf("dispatcher did not use persisted jar update: status=%d body=%s jar=%+v", recorder.Code, recorder.Body.String(), jars.updated)
	}
}

func TestAdvisorConfirmBudgetActionUsesPersistedProposalAndMutationKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &advisorActionStore{messages: []entity.AdvisorMessage{{
		ConversationID: "conversation-1", Generation: 4, Role: entity.AdvisorRoleAssistant,
		Parts: []entity.AdvisorPart{{Type: "budget_confirmation", Data: map[string]any{"view": map[string]any{
			"action_id": "budget-action-1", "action": "create_budget", "conversation_id": "conversation-1", "generation": int64(4),
			"name": "Ăn uống", "limit_amount": int64(4000000), "start_date": "2026-09-01", "end_date": "2026-09-30",
		}}}},
	}}}
	service := &usecase.AdvisorService{Store: store}
	budgets := &advisorBudgetRepositoryStub{}
	engine := gin.New()
	engine.Use(func(c *gin.Context) { c.Set(contextUserIDKey, "owner-1"); c.Next() })
	engine.POST("/actions/:action_id/confirm", (&AdvisorHandler{Service: service, Budgets: budgets}).ConfirmAction)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/actions/budget-action-1/confirm", strings.NewReader(`{"conversation_id":"conversation-1","generation":4}`))
	request.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusCreated || budgets.idempotencyKey != "budget-action-1" || budgets.created == nil || budgets.created.Name != "Ăn uống" || budgets.created.LimitAmount != 4000000 {
		t.Fatalf("dispatcher did not use persisted budget proposal: status=%d body=%s budget=%+v", recorder.Code, recorder.Body.String(), budgets.created)
	}
	replay := httptest.NewRecorder()
	replayRequest := httptest.NewRequest(http.MethodPost, "/actions/budget-action-1/confirm", strings.NewReader(`{"conversation_id":"conversation-1","generation":4}`))
	replayRequest.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(replay, replayRequest)
	if replay.Code != http.StatusOK {
		t.Fatalf("same budget action must replay without creating a second budget: status=%d body=%s", replay.Code, replay.Body.String())
	}
}

func TestAdvisorConfirmBudgetUpdateUsesPersistedProposal(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &advisorActionStore{messages: []entity.AdvisorMessage{{
		ConversationID: "conversation-1", Generation: 4, Role: entity.AdvisorRoleAssistant,
		Parts: []entity.AdvisorPart{{Type: "budget_confirmation", Data: map[string]any{"view": map[string]any{
			"action_id": "budget-update-1", "action": "update_budget", "budget_id": "11111111-1111-4111-8111-111111111111", "conversation_id": "conversation-1", "generation": int64(4),
			"name": "Ăn uống mới", "limit_amount": int64(5000000), "start_date": "2026-09-01", "end_date": "2026-09-30",
		}}}},
	}}}
	service := &usecase.AdvisorService{Store: store}
	budgets := &advisorBudgetRepositoryStub{}
	engine := gin.New()
	engine.Use(func(c *gin.Context) { c.Set(contextUserIDKey, "owner-1"); c.Next() })
	engine.POST("/actions/:action_id/confirm", (&AdvisorHandler{Service: service, Budgets: budgets, Users: advisorOverviewUsers{}}).ConfirmAction)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/actions/budget-update-1/confirm", strings.NewReader(`{"conversation_id":"conversation-1","generation":4}`))
	request.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK || budgets.updated == nil || budgets.updated.ID != "11111111-1111-4111-8111-111111111111" || budgets.updated.Name != "Ăn uống mới" || budgets.updated.LimitAmount != 5000000 {
		t.Fatalf("dispatcher did not use persisted budget update: status=%d body=%s budget=%+v", recorder.Code, recorder.Body.String(), budgets.updated)
	}
}

func TestAdvisorConfirmTransactionUpdatePreservesExistingJar(t *testing.T) {
	gin.SetMode(gin.TestMode)
	transactionID := "11111111-1111-4111-8111-111111111111"
	categoryID := "33333333-3333-4333-8333-333333333333"
	jarID := "44444444-4444-4444-8444-444444444444"
	transactions, wallets, categories, users, jars := advisorJarLinkedTransactionDependencies(transactionID, categoryID, jarID)
	store := &advisorActionStore{messages: []entity.AdvisorMessage{{
		ConversationID: "conversation-1", Generation: 4, Role: entity.AdvisorRoleAssistant,
		Parts: []entity.AdvisorPart{{Type: "transaction_confirmation", Data: map[string]any{"view": advisorJarLinkedTransactionView("update_transaction", "transaction-update-jar-1", transactionID, categoryID)}}},
	}}}
	engine := gin.New()
	engine.Use(func(c *gin.Context) { c.Set(contextUserIDKey, "owner-1"); c.Next() })
	engine.POST("/actions/:action_id/confirm", (&AdvisorHandler{
		Service: &usecase.AdvisorService{Store: store}, Transactions: transactions, Wallets: wallets,
		Categories: categories, Users: users, Jars: jars,
	}).ConfirmAction)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/actions/transaction-update-jar-1/confirm", strings.NewReader(`{"conversation_id":"conversation-1","generation":4}`))
	request.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	gotJar, ok := transactions.updated["jar_id"].(*string)
	if !ok || gotJar == nil || *gotJar != jarID {
		t.Fatalf("AI update cleared existing jar: updates=%#v", transactions.updated)
	}
}

func TestAdvisorConfirmRecategorizeJarLinkedTransactionUsesJarValidationDependencies(t *testing.T) {
	gin.SetMode(gin.TestMode)
	transactionID := "11111111-1111-4111-8111-111111111111"
	oldCategoryID := "33333333-3333-4333-8333-333333333333"
	newCategoryID := "55555555-5555-4555-8555-555555555555"
	jarID := "44444444-4444-4444-8444-444444444444"
	transactions, wallets, categories, users, jars := advisorJarLinkedTransactionDependencies(transactionID, oldCategoryID, jarID)
	categories.categories = append(categories.categories, entity.Category{ID: newCategoryID, Kind: entity.TransactionTypeExpense, IsActive: true})
	store := &advisorActionStore{messages: []entity.AdvisorMessage{{
		ConversationID: "conversation-1", Generation: 4, Role: entity.AdvisorRoleAssistant,
		Parts: []entity.AdvisorPart{{Type: "transaction_confirmation", Data: map[string]any{"view": map[string]any{
			"action_id": "transaction-recat-jar-1", "action": "recategorize_transaction", "conversation_id": "conversation-1", "generation": int64(4),
			"transaction_id": transactionID, "category_id": newCategoryID, "expected_version": int64(1),
		}}}},
	}}}
	engine := gin.New()
	engine.Use(func(c *gin.Context) { c.Set(contextUserIDKey, "owner-1"); c.Next() })
	engine.POST("/actions/:action_id/confirm", (&AdvisorHandler{
		Service: &usecase.AdvisorService{Store: store}, Transactions: transactions, Wallets: wallets,
		Categories: categories, Users: users, Jars: jars,
	}).ConfirmAction)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/actions/transaction-recat-jar-1/confirm", strings.NewReader(`{"conversation_id":"conversation-1","generation":4}`))
	request.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	gotJar, ok := transactions.updated["jar_id"].(*string)
	if !ok || gotJar == nil || *gotJar != jarID {
		t.Fatalf("recategorization did not preserve jar: updates=%#v", transactions.updated)
	}
}

func TestAdvisorConfirmWalletActionRejectsStaleGenerationAndClientFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &advisorActionStore{messages: []entity.AdvisorMessage{{
		ConversationID: "conversation-1", Generation: 4, Role: entity.AdvisorRoleAssistant,
		Parts: []entity.AdvisorPart{{Type: "wallet_confirmation", Data: map[string]any{"view": map[string]any{
			"action_id": "action-1", "action": "create_wallet", "conversation_id": "conversation-1", "generation": int64(4),
			"name": "Du lịch", "type": "basic", "opening_balance": int64(0), "is_in_total": true,
		}}}},
	}}}
	service := &usecase.AdvisorService{Store: store}
	engine := gin.New()
	engine.Use(func(c *gin.Context) { c.Set(contextUserIDKey, "owner-1"); c.Next() })
	engine.POST("/actions/:action_id/confirm", (&AdvisorHandler{Service: service, Wallets: &walletRepositoryStub{}}).ConfirmAction)

	stale := httptest.NewRecorder()
	staleRequest := httptest.NewRequest(http.MethodPost, "/actions/action-1/confirm", strings.NewReader(`{"conversation_id":"conversation-1","generation":3}`))
	staleRequest.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(stale, staleRequest)
	if stale.Code != http.StatusConflict || !strings.Contains(stale.Body.String(), "advisor_action_stale") {
		t.Fatalf("stale action status=%d body=%s", stale.Code, stale.Body.String())
	}

	tampered := httptest.NewRecorder()
	tamperedRequest := httptest.NewRequest(http.MethodPost, "/actions/action-1/confirm", strings.NewReader(`{"conversation_id":"conversation-1","generation":4,"name":"Attacker"}`))
	tamperedRequest.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(tampered, tamperedRequest)
	if tampered.Code != http.StatusBadRequest {
		t.Fatalf("tampered action status=%d body=%s", tampered.Code, tampered.Body.String())
	}
}

func TestAdvisorConfirmVersionedActionRejectsProposalWithoutExpectedVersion(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &advisorActionStore{messages: []entity.AdvisorMessage{{
		ConversationID: "conversation-1", Generation: 4, Role: entity.AdvisorRoleAssistant,
		Parts: []entity.AdvisorPart{{Type: "transaction_confirmation", Data: map[string]any{"view": map[string]any{
			"action_id": "transaction-update-1", "action": "update_transaction", "conversation_id": "conversation-1", "generation": int64(4),
			"transaction_id": "11111111-1111-4111-8111-111111111111", "wallet_id": "22222222-2222-4222-8222-222222222222", "type": "expense", "amount": int64(100),
		}}}},
	}}}
	service := &usecase.AdvisorService{Store: store}
	engine := gin.New()
	engine.Use(func(c *gin.Context) { c.Set(contextUserIDKey, "owner-1"); c.Next() })
	engine.POST("/actions/:action_id/confirm", (&AdvisorHandler{Service: service}).ConfirmAction)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/actions/transaction-update-1/confirm", strings.NewReader(`{"conversation_id":"conversation-1","generation":4}`))
	request.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusConflict || !strings.Contains(recorder.Body.String(), "advisor_action_stale") {
		t.Fatalf("missing expected version must be rejected before dispatch: status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestAdvisorOverviewUsesFinanceReaderWithoutProvider(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(func(c *gin.Context) { c.Set(contextUserIDKey, "owner-1"); c.Next() })
	service := usecase.NewFinanceQueryService(advisorOverviewReader{})
	engine.GET("/overview", (&AdvisorHandler{Users: advisorOverviewUsers{}, Finance: service}).Overview)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/overview?month=2026-09", nil)
	engine.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"expense":150`) {
		t.Fatalf("unexpected overview response: status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestAdvisorFailureMapsPurgedHistoryToGone(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/advisor-error", func(c *gin.Context) {
		advisorFailure(c, repository.ErrAdvisorPurged)
	})
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/advisor-error", nil))
	if recorder.Code != http.StatusGone {
		t.Fatalf("purged history status=%d, want410; body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "advisor_history_expired") {
		t.Fatalf("purged history code missing: %s", recorder.Body.String())
	}
	if errors.Is(repository.ErrAdvisorPurged, repository.ErrAdvisorRequestConflict) {
		t.Fatal("purged and conflict errors must remain distinct")
	}
}

func TestAdvisorFailureMapsProviderOutageToServiceUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/advisor-provider-error", func(c *gin.Context) {
		advisorFailure(c, usecase.ErrAdvisorProviderUnavailable)
	})
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/advisor-provider-error", nil))
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("provider outage status=%d, want503; body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "advisor_unavailable") {
		t.Fatalf("provider outage code missing: %s", recorder.Body.String())
	}
}

func TestAdvisorResetRequiresIdempotencyKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(func(c *gin.Context) { c.Set(contextUserIDKey, "owner-1"); c.Next() })
	engine.POST("/reset", (&AdvisorHandler{}).ResetConversation)
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/reset", nil))
	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), "advisor_invalid") {
		t.Fatalf("missing reset idempotency key: status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}
