package httpapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
)

type BudgetReportHandler struct {
	Budgets      repository.BudgetRepository
	Transactions repository.TransactionRepository
	Categories   repository.CategoryRepository
	Users        repository.UserRepository
	Now          func() time.Time
}

func (r *Router) RegisterBudgetReportRoutes(h *BudgetReportHandler) {
	routes := r.Engine.Group("/api/v1/budgets")
	routes.Use(r.AuthMiddleware.RequireAuth)
	routes.GET("/:id/report", h.GetBudgetReport)
	routes.GET("/:id/transactions", h.ListBudgetTransactions)
}

// GetBudgetReport godoc
// @Summary Read deterministic budget forecast and daily spend
// @Tags Budgets
// @Produce json
// @Security BearerAuth
// @Param id path string true "Owned budget UUID"
// @Success 200 {object} entity.BudgetReport
// @Failure 400 {object} Problem
// @Failure 401 {object} Problem
// @Failure 404 {object} Problem
// @Router /api/v1/budgets/{id}/report [get]
func (h *BudgetReportHandler) GetBudgetReport(c *gin.Context) {
	budget, rows, categories, now, location, ok := h.load(c)
	if !ok {
		return
	}
	report, err := entity.CalculateBudgetReport(budget, rows, categories, now, location)
	if err != nil {
		budgetFailure(c, err)
		return
	}
	OK(c, report)
}

type BudgetTransactionPage struct {
	BudgetID string               `json:"budget_id"`
	Timezone string               `json:"timezone"`
	AsOf     time.Time            `json:"as_of"`
	Items    []entity.Transaction `json:"items"`
	Total    int                  `json:"total"`
	Limit    int                  `json:"limit"`
	Offset   int                  `json:"offset"`
}

// ListBudgetTransactions godoc
// @Summary List transactions matching exactly the budget report scope
// @Tags Budgets
// @Produce json
// @Security BearerAuth
// @Param id path string true "Owned budget UUID"
// @Param limit query int false "Page size 1..100, default 50"
// @Param offset query int false "Offset, default 0"
// @Success 200 {object} BudgetTransactionPage
// @Failure 400 {object} Problem
// @Failure 401 {object} Problem
// @Failure 404 {object} Problem
// @Router /api/v1/budgets/{id}/transactions [get]
func (h *BudgetReportHandler) ListBudgetTransactions(c *gin.Context) {
	limit, offset, ok := budgetReportPagination(c)
	if !ok {
		return
	}
	budget, rows, categories, now, location, ok := h.load(c)
	if !ok {
		return
	}
	matching := entity.MatchBudgetTransactions(budget, rows, categories, now, location)
	begin := min(offset, len(matching))
	end := begin + min(limit, len(matching)-begin)
	OK(c, BudgetTransactionPage{BudgetID: budget.ID, Timezone: location.String(), AsOf: now.UTC(), Items: matching[begin:end], Total: len(matching), Limit: limit, Offset: offset})
}

func budgetReportPagination(c *gin.Context) (int, int, bool) {
	limit, offset := 50, 0
	var err error
	if raw, exists := c.GetQuery("limit"); exists {
		limit, err = strconv.Atoi(raw)
		if err != nil || limit < 1 || limit > 100 {
			budgetBadRequest(c, "Kích thước trang phải từ 1 đến 100.")
			return 0, 0, false
		}
	}
	if raw, exists := c.GetQuery("offset"); exists {
		offset, err = strconv.Atoi(raw)
		if err != nil || offset < 0 {
			budgetBadRequest(c, "Vị trí trang không hợp lệ.")
			return 0, 0, false
		}
	}
	return limit, offset, true
}

func (h *BudgetReportHandler) load(c *gin.Context) (entity.Budget, []entity.Transaction, []entity.Category, time.Time, *time.Location, bool) {
	var empty entity.Budget
	fail := func() (entity.Budget, []entity.Transaction, []entity.Category, time.Time, *time.Location, bool) {
		return empty, nil, nil, time.Time{}, nil, false
	}
	owner, ok := walletOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return fail()
	}
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		budgetBadRequest(c, "Ngân sách không hợp lệ.")
		return fail()
	}
	if h == nil || h.Budgets == nil || h.Users == nil || h.Transactions == nil || h.Categories == nil {
		Fail(c, http.StatusServiceUnavailable, Problem{Code: "BUDGET_REPORT_UNAVAILABLE", Title: "Service Unavailable", Detail: "Chưa thể tải báo cáo ngân sách."})
		return fail()
	}
	budgets, err := h.Budgets.List(owner)
	if err != nil {
		budgetFailure(c, err)
		return fail()
	}
	found := false
	for _, b := range budgets {
		if b.ID == id && b.OwnerID == owner {
			empty = b
			found = true
			break
		}
	}
	if !found {
		Fail(c, http.StatusNotFound, Problem{Code: "BUDGET_NOT_FOUND", Title: problemTitleNotFound, Detail: "Không tìm thấy ngân sách."})
		return fail()
	}
	user, err := h.Users.FindByID(owner)
	if err != nil {
		budgetFailure(c, err)
		return fail()
	}
	if user == nil || user.ID != owner {
		Fail(c, http.StatusNotFound, Problem{Code: "BUDGET_NOT_FOUND", Title: problemTitleNotFound, Detail: "Không tìm thấy tài khoản ngân sách."})
		return fail()
	}
	location, err := time.LoadLocation(user.Timezone)
	if err != nil {
		budgetFailure(c, err)
		return fail()
	}
	rows, err := h.Transactions.List(owner)
	if err != nil {
		budgetFailure(c, err)
		return fail()
	}
	categories, err := h.Categories.ListVisible(owner)
	if err != nil {
		budgetFailure(c, err)
		return fail()
	}
	now := time.Now()
	if h.Now != nil {
		now = h.Now()
	}
	calculated, err := entity.CalculateBudgetsChecked(budgets, rows, categories, now, location)
	if err != nil {
		budgetFailure(c, err)
		return fail()
	}
	for _, b := range calculated.Items {
		if b.ID == id {
			empty = b
			break
		}
	}
	return empty, rows, categories, now, location, true
}
