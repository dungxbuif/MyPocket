package httpapi

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
	"github.com/mypocket/backend/internal/usecase"
)

type MonthlyReportHandler struct {
	Service *usecase.MonthlyReportService
}

func (r *Router) RegisterMonthlyReportRoutes(h *MonthlyReportHandler) {
	g := r.Engine.Group("/api/v1/reports")
	g.Use(r.AuthMiddleware.RequireAdvisorAuth)
	g.GET("/monthly", h.Get)
	g.GET("/monthly/transactions", h.ListTransactions)
}

// GetMonthlyReport godoc
// @Summary Read the deterministic monthly report read model
// @Tags Reports
// @Produce json
// @Security BearerAuth
// @Param month query string false "Calendar month YYYY-MM"
// @Param wallet_id query string false "Optional owner wallet filter"
// @Param granularity query string false "Series granularity: week or day"
// @Success 200 {object} entity.MonthlyReport
// @Router /api/v1/reports/monthly [get]
func (h *MonthlyReportHandler) Get(c *gin.Context) {
	owner, ok := transactionOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	principal, _ := advisorPrincipalFromContext(c)
	principal.OwnerID = owner
	scope := entity.MonthlyReportScope{Month: strings.TrimSpace(c.Query("month")), WalletID: strings.TrimSpace(c.Query("wallet_id")), Granularity: strings.TrimSpace(c.Query("granularity"))}
	if len(scope.Month) > 16 || len(scope.WalletID) > 128 || len(scope.Granularity) > 16 {
		Fail(c, http.StatusBadRequest, Problem{Code: "REPORT_BAD_REQUEST", Title: problemTitleBadRequest, Detail: "Bộ lọc báo cáo không hợp lệ."})
		return
	}
	if h == nil || h.Service == nil {
		Fail(c, http.StatusNotImplemented, Problem{Code: "REPORT_UNAVAILABLE", Title: problemTitleInternalServer, Detail: "Báo cáo chưa được cấu hình."})
		return
	}
	report, err := h.Service.Read(c.Request.Context(), principal, scope)
	if err != nil {
		monthlyReportMapError(c, err)
		return
	}
	OK(c, report)
}

func monthlyReportMapError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, entity.ErrMonthlyReportScopeInvalid), errors.Is(err, entity.ErrMonthlyReportFuture):
		Fail(c, http.StatusBadRequest, Problem{Code: "REPORT_BAD_REQUEST", Title: problemTitleBadRequest, Detail: "Tháng hoặc phạm vi báo cáo không hợp lệ."})
	case errors.Is(err, repository.ErrNotFound):
		Fail(c, http.StatusNotFound, Problem{Code: "REPORT_NOT_FOUND", Title: problemTitleNotFound, Detail: "Không tìm thấy phạm vi báo cáo."})
	case errors.Is(err, entity.ErrMonthlyReportOverflow):
		Fail(c, http.StatusUnprocessableEntity, Problem{Code: "REPORT_OVERFLOW", Title: problemTitleBadRequest, Detail: "Số liệu báo cáo vượt giới hạn an toàn."})
	default:
		Fail(c, http.StatusInternalServerError, Problem{Code: "REPORT_FAILED", Title: problemTitleInternalServer, Detail: "Không thể tải báo cáo."})
	}
}

// ListTransactions returns a bounded ledger page only after the user drills
// into a monthly chart/category. The report landing page never fetches it.
// @Summary List a bounded monthly report drill-down
// @Tags Reports
// @Produce json
// @Security BearerAuth
// @Param month query string true "Calendar month YYYY-MM"
// @Param wallet_id query string false "Optional owner wallet filter"
// @Param category_id query string false "Optional category filter"
// @Param kind query string false "all, income, or expense"
// @Param limit query int false "Page size up to 100"
// @Param offset query int false "Page offset"
// @Success 200 {object} entity.SavingsTransactionPage
// @Router /api/v1/reports/monthly/transactions [get]
func (h *MonthlyReportHandler) ListTransactions(c *gin.Context) {
	owner, ok := transactionOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	principal, _ := advisorPrincipalFromContext(c)
	principal.OwnerID = owner
	limit, offset, ok := reportPagination(c)
	if !ok {
		return
	}
	scope := entity.MonthlyReportScope{Month: strings.TrimSpace(c.Query("month")), WalletID: strings.TrimSpace(c.Query("wallet_id")), Granularity: "day"}
	categoryID, kind := strings.TrimSpace(c.Query("category_id")), strings.TrimSpace(c.Query("kind"))
	if h == nil || h.Service == nil || len(categoryID) > 128 || len(kind) > 16 {
		monthlyReportMapError(c, entity.ErrMonthlyReportScopeInvalid)
		return
	}
	page, err := h.Service.Transactions(c.Request.Context(), principal, scope, categoryID, kind, limit, offset)
	if err != nil {
		monthlyReportMapError(c, err)
		return
	}
	OK(c, page)
}

func reportPagination(c *gin.Context) (int, int, bool) {
	limit, offset := 50, 0
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		if _, err := fmt.Sscanf(raw, "%d", &limit); err != nil {
			monthlyReportMapError(c, entity.ErrMonthlyReportScopeInvalid)
			return 0, 0, false
		}
	}
	if raw := strings.TrimSpace(c.Query("offset")); raw != "" {
		if _, err := fmt.Sscanf(raw, "%d", &offset); err != nil {
			monthlyReportMapError(c, entity.ErrMonthlyReportScopeInvalid)
			return 0, 0, false
		}
	}
	if limit < 1 || limit > 100 || offset < 0 {
		monthlyReportMapError(c, entity.ErrMonthlyReportScopeInvalid)
		return 0, 0, false
	}
	return limit, offset, true
}
