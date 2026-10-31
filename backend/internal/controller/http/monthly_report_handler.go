package httpapi

import (
	"errors"
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
