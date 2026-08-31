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

type planningRepoStub struct {
	progress      []planning.BudgetProgress
	created       planning.Budget
	updated       planning.Budget
	listUserID    string
	createUserID  string
	createInput   planning.CreateBudgetInput
	updateUserID  string
	updateID      string
	updateInput   planning.UpdateBudgetInput
	archiveUserID string
	archiveID     string
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
