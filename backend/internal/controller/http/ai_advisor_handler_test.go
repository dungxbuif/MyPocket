package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
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
		"GET /api/v1/ai/advisor/conversation/messages",
		"POST /api/v1/ai/advisor/conversation/reset",
		"GET /api/v1/ai/advisor/runs/:id",
		"POST /api/v1/ai/advisor/runs/:id/cancel",
		"POST /api/v1/ai/advisor/actions/:action_id/confirm",
	} {
		if !routes[expected] {
			t.Fatalf("missing route %s", expected)
		}
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
