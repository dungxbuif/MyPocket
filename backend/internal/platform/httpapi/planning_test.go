package httpapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"mypocket/internal/identity"
	"mypocket/internal/planning"
	"mypocket/internal/platform/httpapi"
)

func TestBudgetsAPIRequiresAuth(t *testing.T) {
	handler := httpapi.NewRouter(authTestConfig(), httpapi.Dependencies{
		IdentityRepository: &authRepoStub{},
		PlanningRepository: &planningRepoStub{},
	})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/budgets", nil)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", res.Code, res.Body.String())
	}
	if !strings.Contains(res.Body.String(), "AUTH_REQUIRED") {
		t.Fatalf("expected auth error, got %s", res.Body.String())
	}
}

func TestCreateBudgetRequiresCSRF(t *testing.T) {
	handler := httpapi.NewRouter(authTestConfig(), httpapi.Dependencies{
		IdentityRepository: &authRepoStub{user: identity.User{ID: "user_123", Email: "a@example.com", EmailVerified: true}},
		PlanningRepository: &planningRepoStub{},
	})
	req := authenticatedRequest(t, http.MethodPost, "/api/v1/budgets", `{"name":"Ăn uống","period_type":"monthly","amount_vnd":500000}`)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", res.Code, res.Body.String())
	}
	if !strings.Contains(res.Body.String(), "CSRF_REQUIRED") {
		t.Fatalf("expected csrf error, got %s", res.Body.String())
	}
}

func TestBudgetsAPIUsesAuthenticatedUser(t *testing.T) {
	repo := &planningRepoStub{
		progress: []planning.BudgetProgress{{
			Budget:      planning.Budget{ID: "budget-1", UserID: "user_123", Name: "Ăn uống", PeriodType: planning.BudgetMonthly, AmountVND: 500000, AllCategories: true, Version: 1},
			PeriodStart: "2026-08-01",
			PeriodEnd:   "2026-08-31",
			SpentVND:    400000,
			Percent:     80,
			Alert80:     true,
		}},
		created: planning.Budget{ID: "budget-2", UserID: "user_123", Name: "Mua sắm", PeriodType: planning.BudgetWeekly, AmountVND: 300000, Version: 1},
		updated: planning.Budget{ID: "budget-2", UserID: "user_123", Name: "Mua sắm mới", PeriodType: planning.BudgetMonthly, AmountVND: 600000, Version: 2},
	}
	handler := httpapi.NewRouter(authTestConfig(), httpapi.Dependencies{
		IdentityRepository: &authRepoStub{user: identity.User{ID: "user_123", Email: "a@example.com", EmailVerified: true}},
		PlanningRepository: repo,
	})

	listReq := authenticatedRequest(t, http.MethodGet, "/api/v1/budgets", "")
	listRes := httptest.NewRecorder()
	handler.ServeHTTP(listRes, listReq)
	if listRes.Code != http.StatusOK || repo.listUserID != "user_123" {
		t.Fatalf("budget list failed/scoped wrong: code=%d user=%q body=%s", listRes.Code, repo.listUserID, listRes.Body.String())
	}
	if !strings.Contains(listRes.Body.String(), `"percent":80`) {
		t.Fatalf("unexpected list response: %s", listRes.Body.String())
	}

	createReq := authenticatedRequest(t, http.MethodPost, "/api/v1/budgets", `{"name":"Mua sắm","period_type":"weekly","amount_vnd":300000,"category_ids":["cat-1"]}`)
	addCSRF(createReq)
	createRes := httptest.NewRecorder()
	handler.ServeHTTP(createRes, createReq)
	if createRes.Code != http.StatusCreated || repo.createUserID != "user_123" || repo.createInput.CategoryIDs[0] != "cat-1" {
		t.Fatalf("budget create failed/scoped wrong: code=%d user=%q input=%#v body=%s", createRes.Code, repo.createUserID, repo.createInput, createRes.Body.String())
	}

	updateReq := authenticatedRequest(t, http.MethodPatch, "/api/v1/budgets/budget-2", `{"name":"Mua sắm mới","period_type":"monthly","amount_vnd":600000}`)
	addCSRF(updateReq)
	updateRes := httptest.NewRecorder()
	handler.ServeHTTP(updateRes, updateReq)
	if updateRes.Code != http.StatusOK || repo.updateUserID != "user_123" || repo.updateID != "budget-2" || repo.updateInput.PeriodType != planning.BudgetMonthly {
		t.Fatalf("budget update failed/scoped wrong: code=%d user=%q id=%q input=%#v body=%s", updateRes.Code, repo.updateUserID, repo.updateID, repo.updateInput, updateRes.Body.String())
	}

	archiveReq := authenticatedRequest(t, http.MethodPost, "/api/v1/budgets/budget-2/archive", "")
	addCSRF(archiveReq)
	archiveRes := httptest.NewRecorder()
	handler.ServeHTTP(archiveRes, archiveReq)
	if archiveRes.Code != http.StatusOK || repo.archiveUserID != "user_123" || repo.archiveID != "budget-2" {
		t.Fatalf("budget archive failed/scoped wrong: code=%d user=%q id=%q body=%s", archiveRes.Code, repo.archiveUserID, repo.archiveID, archiveRes.Body.String())
	}
}

func TestEventsAPIUsesAuthenticatedUser(t *testing.T) {
	repo := &planningRepoStub{
		events:       []planning.EventSummary{{ID: "event-1", UserID: "user_123", Name: "Đà Lạt", StartsOn: "2026-08-31", EndsOn: "2026-09-02", TotalVND: 125000, TransactionCount: 1, Version: 1}},
		createdEvent: planning.EventSummary{ID: "event-2", UserID: "user_123", Name: "Huế", StartsOn: "2026-09-10", EndsOn: "2026-09-10", Version: 1},
		updatedEvent: planning.EventSummary{ID: "event-2", UserID: "user_123", Name: "Huế mới", StartsOn: "2026-09-10", EndsOn: "2026-09-12", Version: 2},
	}
	handler := httpapi.NewRouter(authTestConfig(), httpapi.Dependencies{
		IdentityRepository: &authRepoStub{user: identity.User{ID: "user_123", Email: "a@example.com", EmailVerified: true}},
		PlanningRepository: repo,
	})

	listReq := authenticatedRequest(t, http.MethodGet, "/api/v1/events", "")
	listRes := httptest.NewRecorder()
	handler.ServeHTTP(listRes, listReq)
	if listRes.Code != http.StatusOK || repo.eventListUserID != "user_123" || !strings.Contains(listRes.Body.String(), `"transaction_count":1`) {
		t.Fatalf("event list failed/scoped wrong: code=%d user=%q body=%s", listRes.Code, repo.eventListUserID, listRes.Body.String())
	}

	createReq := authenticatedRequest(t, http.MethodPost, "/api/v1/events", `{"name":"Huế","starts_on":"2026-09-10","ends_on":"2026-09-10","note":"Trip"}`)
	addCSRF(createReq)
	createRes := httptest.NewRecorder()
	handler.ServeHTTP(createRes, createReq)
	if createRes.Code != http.StatusCreated || repo.eventCreateUserID != "user_123" || repo.eventCreateInput.Name != "Huế" {
		t.Fatalf("event create failed/scoped wrong: code=%d user=%q input=%#v body=%s", createRes.Code, repo.eventCreateUserID, repo.eventCreateInput, createRes.Body.String())
	}

	linkReq := authenticatedRequest(t, http.MethodPost, "/api/v1/events/event-2/transactions/tx-1", "")
	addCSRF(linkReq)
	linkRes := httptest.NewRecorder()
	handler.ServeHTTP(linkRes, linkReq)
	if linkRes.Code != http.StatusOK || repo.eventLinkID != "event-2" || repo.eventLinkTxID != "tx-1" {
		t.Fatalf("event link failed/scoped wrong: code=%d event=%q tx=%q body=%s", linkRes.Code, repo.eventLinkID, repo.eventLinkTxID, linkRes.Body.String())
	}
}

func TestObligationsAPIUsesAuthenticatedUser(t *testing.T) {
	repo := &planningRepoStub{
		obligations:       []planning.ObligationSummary{{ID: "debt-1", UserID: "user_123", Direction: planning.ObligationBorrowed, PrincipalVND: 1000000, Counterparty: "Anh Minh", DueOn: "2026-09-30", RepaidVND: 600000, RemainingVND: 400000, Version: 1}},
		createdObligation: planning.ObligationSummary{ID: "debt-2", UserID: "user_123", Direction: planning.ObligationLent, PrincipalVND: 300000, Counterparty: "Chị Lan", DueOn: "2026-09-15", RemainingVND: 300000, Version: 1},
	}
	handler := httpapi.NewRouter(authTestConfig(), httpapi.Dependencies{
		IdentityRepository: &authRepoStub{user: identity.User{ID: "user_123", Email: "a@example.com", EmailVerified: true}},
		PlanningRepository: repo,
	})

	listReq := authenticatedRequest(t, http.MethodGet, "/api/v1/obligations", "")
	listRes := httptest.NewRecorder()
	handler.ServeHTTP(listRes, listReq)
	if listRes.Code != http.StatusOK || repo.obligationListUserID != "user_123" || !strings.Contains(listRes.Body.String(), `"remaining_vnd":400000`) {
		t.Fatalf("obligation list failed/scoped wrong: code=%d user=%q body=%s", listRes.Code, repo.obligationListUserID, listRes.Body.String())
	}

	createReq := authenticatedRequest(t, http.MethodPost, "/api/v1/obligations", `{"direction":"lent","principal_vnd":300000,"counterparty":"Chị Lan","due_on":"2026-09-15","note":"Ứng trước"}`)
	addCSRF(createReq)
	createRes := httptest.NewRecorder()
	handler.ServeHTTP(createRes, createReq)
	if createRes.Code != http.StatusCreated || repo.obligationCreateUserID != "user_123" || repo.obligationCreateInput.Direction != planning.ObligationLent {
		t.Fatalf("obligation create failed/scoped wrong: code=%d user=%q input=%#v body=%s", createRes.Code, repo.obligationCreateUserID, repo.obligationCreateInput, createRes.Body.String())
	}

	linkReq := authenticatedRequest(t, http.MethodPost, "/api/v1/obligations/debt-2/repayments/tx-1", "")
	addCSRF(linkReq)
	linkRes := httptest.NewRecorder()
	handler.ServeHTTP(linkRes, linkReq)
	if linkRes.Code != http.StatusOK || repo.obligationLinkID != "debt-2" || repo.obligationLinkTxID != "tx-1" {
		t.Fatalf("obligation link failed/scoped wrong: code=%d obligation=%q tx=%q body=%s", linkRes.Code, repo.obligationLinkID, repo.obligationLinkTxID, linkRes.Body.String())
	}
}

func TestRecurringSchedulesAPIUsesAuthenticatedUser(t *testing.T) {
	now := time.Date(2026, 8, 31, 2, 0, 0, 0, time.UTC)
	repo := &planningRepoStub{
		schedules:       []planning.RecurringSchedule{{ID: "schedule-1", UserID: "user_123", Name: "Tiền nhà", Frequency: planning.RecurrenceMonthly, Timezone: "Asia/Ho_Chi_Minh", StartsAt: now, NextOccursAt: now, Version: 1}},
		createdSchedule: planning.RecurringSchedule{ID: "schedule-2", UserID: "user_123", Name: "Internet", Frequency: planning.RecurrenceMonthly, Timezone: "Asia/Ho_Chi_Minh", StartsAt: now, NextOccursAt: now, Version: 1},
		drafts:          []planning.TransactionDraft{{ID: "draft-1", UserID: "user_123", ScheduleID: "schedule-1", OccurrenceKey: "recurring:schedule-1:2026-08-31T02:00:00Z", AmountVND: 250000, OccurredAt: now, Status: "pending", Version: 1}},
	}
	handler := httpapi.NewRouter(authTestConfig(), httpapi.Dependencies{
		IdentityRepository: &authRepoStub{user: identity.User{ID: "user_123", Email: "a@example.com", EmailVerified: true}},
		PlanningRepository: repo,
	})

	listReq := authenticatedRequest(t, http.MethodGet, "/api/v1/recurring-schedules", "")
	listRes := httptest.NewRecorder()
	handler.ServeHTTP(listRes, listReq)
	if listRes.Code != http.StatusOK || repo.scheduleListUserID != "user_123" || !strings.Contains(listRes.Body.String(), "Tiền nhà") {
		t.Fatalf("schedule list failed/scoped wrong: code=%d user=%q body=%s", listRes.Code, repo.scheduleListUserID, listRes.Body.String())
	}

	createReq := authenticatedRequest(t, http.MethodPost, "/api/v1/recurring-schedules", `{"name":"Internet","frequency":"monthly","timezone":"Asia/Ho_Chi_Minh","starts_at":"2026-08-31T09:00:00+07:00","type":"expense","source_wallet_id":"wallet-1","category_id":"cat-1","amount_vnd":250000,"note":"Wifi"}`)
	addCSRF(createReq)
	createRes := httptest.NewRecorder()
	handler.ServeHTTP(createRes, createReq)
	if createRes.Code != http.StatusCreated || repo.scheduleCreateUserID != "user_123" || repo.scheduleCreateInput.AmountVND != 250000 {
		t.Fatalf("schedule create failed/scoped wrong: code=%d user=%q input=%#v body=%s", createRes.Code, repo.scheduleCreateUserID, repo.scheduleCreateInput, createRes.Body.String())
	}

	draftReq := authenticatedRequest(t, http.MethodGet, "/api/v1/transaction-drafts", "")
	draftRes := httptest.NewRecorder()
	handler.ServeHTTP(draftRes, draftReq)
	if draftRes.Code != http.StatusOK || repo.draftListUserID != "user_123" || !strings.Contains(draftRes.Body.String(), `"status":"pending"`) {
		t.Fatalf("draft list failed/scoped wrong: code=%d user=%q body=%s", draftRes.Code, repo.draftListUserID, draftRes.Body.String())
	}
}

type planningRepoStub struct {
	progress                []planning.BudgetProgress
	created                 planning.Budget
	updated                 planning.Budget
	events                  []planning.EventSummary
	createdEvent            planning.EventSummary
	updatedEvent            planning.EventSummary
	obligations             []planning.ObligationSummary
	createdObligation       planning.ObligationSummary
	updatedObligation       planning.ObligationSummary
	listUserID              string
	createUserID            string
	createInput             planning.CreateBudgetInput
	updateUserID            string
	updateID                string
	updateInput             planning.UpdateBudgetInput
	archiveUserID           string
	archiveID               string
	eventListUserID         string
	eventCreateUserID       string
	eventCreateInput        planning.CreateEventInput
	eventUpdateUserID       string
	eventUpdateID           string
	eventUpdateInput        planning.UpdateEventInput
	eventArchiveUserID      string
	eventArchiveID          string
	eventLinkUserID         string
	eventLinkID             string
	eventLinkTxID           string
	obligationListUserID    string
	obligationCreateUserID  string
	obligationCreateInput   planning.CreateObligationInput
	obligationUpdateUserID  string
	obligationUpdateID      string
	obligationUpdateInput   planning.UpdateObligationInput
	obligationArchiveUserID string
	obligationArchiveID     string
	obligationLinkUserID    string
	obligationLinkID        string
	obligationLinkTxID      string
	schedules               []planning.RecurringSchedule
	createdSchedule         planning.RecurringSchedule
	scheduleListUserID      string
	scheduleCreateUserID    string
	scheduleCreateInput     planning.CreateRecurringScheduleInput
	scheduleArchiveUserID   string
	scheduleArchiveID       string
	drafts                  []planning.TransactionDraft
	draftListUserID         string
}

func (s *planningRepoStub) ListBudgetProgress(_ context.Context, userID string, _ time.Time) ([]planning.BudgetProgress, error) {
	s.listUserID = userID
	return s.progress, nil
}

func (s *planningRepoStub) CreateBudget(_ context.Context, userID string, input planning.CreateBudgetInput) (planning.Budget, error) {
	s.createUserID = userID
	s.createInput = input
	return s.created, nil
}

func (s *planningRepoStub) UpdateBudget(_ context.Context, userID string, budgetID string, input planning.UpdateBudgetInput) (planning.Budget, error) {
	s.updateUserID = userID
	s.updateID = budgetID
	s.updateInput = input
	return s.updated, nil
}

func (s *planningRepoStub) ArchiveBudget(_ context.Context, userID string, budgetID string) error {
	s.archiveUserID = userID
	s.archiveID = budgetID
	return nil
}

func (s *planningRepoStub) ListEvents(_ context.Context, userID string) ([]planning.EventSummary, error) {
	s.eventListUserID = userID
	return s.events, nil
}

func (s *planningRepoStub) CreateEvent(_ context.Context, userID string, input planning.CreateEventInput) (planning.EventSummary, error) {
	s.eventCreateUserID = userID
	s.eventCreateInput = input
	return s.createdEvent, nil
}

func (s *planningRepoStub) UpdateEvent(_ context.Context, userID string, eventID string, input planning.UpdateEventInput) (planning.EventSummary, error) {
	s.eventUpdateUserID = userID
	s.eventUpdateID = eventID
	s.eventUpdateInput = input
	return s.updatedEvent, nil
}

func (s *planningRepoStub) ArchiveEvent(_ context.Context, userID string, eventID string) error {
	s.eventArchiveUserID = userID
	s.eventArchiveID = eventID
	return nil
}

func (s *planningRepoStub) LinkEventTransaction(_ context.Context, userID string, eventID string, transactionID string) error {
	s.eventLinkUserID = userID
	s.eventLinkID = eventID
	s.eventLinkTxID = transactionID
	return nil
}

func (s *planningRepoStub) ListObligations(_ context.Context, userID string) ([]planning.ObligationSummary, error) {
	s.obligationListUserID = userID
	return s.obligations, nil
}

func (s *planningRepoStub) CreateObligation(_ context.Context, userID string, input planning.CreateObligationInput) (planning.ObligationSummary, error) {
	s.obligationCreateUserID = userID
	s.obligationCreateInput = input
	return s.createdObligation, nil
}

func (s *planningRepoStub) UpdateObligation(_ context.Context, userID string, obligationID string, input planning.UpdateObligationInput) (planning.ObligationSummary, error) {
	s.obligationUpdateUserID = userID
	s.obligationUpdateID = obligationID
	s.obligationUpdateInput = input
	return s.updatedObligation, nil
}

func (s *planningRepoStub) ArchiveObligation(_ context.Context, userID string, obligationID string) error {
	s.obligationArchiveUserID = userID
	s.obligationArchiveID = obligationID
	return nil
}

func (s *planningRepoStub) LinkObligationRepayment(_ context.Context, userID string, obligationID string, transactionID string) error {
	s.obligationLinkUserID = userID
	s.obligationLinkID = obligationID
	s.obligationLinkTxID = transactionID
	return nil
}

func (s *planningRepoStub) ListRecurringSchedules(_ context.Context, userID string) ([]planning.RecurringSchedule, error) {
	s.scheduleListUserID = userID
	return s.schedules, nil
}

func (s *planningRepoStub) CreateRecurringSchedule(_ context.Context, userID string, input planning.CreateRecurringScheduleInput) (planning.RecurringSchedule, error) {
	s.scheduleCreateUserID = userID
	s.scheduleCreateInput = input
	return s.createdSchedule, nil
}

func (s *planningRepoStub) ArchiveRecurringSchedule(_ context.Context, userID string, scheduleID string) error {
	s.scheduleArchiveUserID = userID
	s.scheduleArchiveID = scheduleID
	return nil
}

func (s *planningRepoStub) ListTransactionDrafts(_ context.Context, userID string) ([]planning.TransactionDraft, error) {
	s.draftListUserID = userID
	return s.drafts, nil
}
