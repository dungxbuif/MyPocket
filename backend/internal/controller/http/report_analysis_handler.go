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

type ReportAnalysisHandler struct {
	Service *usecase.ReportAnalysisService
}

type reportAnalysisBeginInput struct {
	SnapshotID      string `json:"snapshot_id"`
	ClientRequestID string `json:"client_request_id"`
}

func (r *Router) RegisterReportAnalysisRoutes(h *ReportAnalysisHandler) {
	g := r.Engine.Group("/api/v1/reports")
	g.Use(r.AuthMiddleware.RequireAdvisorAuth)
	g.POST("/analyses", h.Begin)
	g.GET("/analyses/latest", h.Latest)
	g.GET("/analyses/:id", h.Get)
	g.POST("/analyses/:id/cancel", h.Cancel)
}

// Latest returns the retained analysis for a report snapshot.
// @Summary Read the latest analysis for a report snapshot
// @Tags Reports
// @Produce json
// @Security BearerAuth
// @Param snapshot_id query string true "Report snapshot ID"
// @Success 200 {object} map[string]any
// @Router /api/v1/reports/analyses/latest [get]
func (h *ReportAnalysisHandler) Latest(c *gin.Context) {
	owner, ok := transactionOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	if h == nil || h.Service == nil {
		reportAnalysisUnavailable(c)
		return
	}
	snapshotID := strings.TrimSpace(c.Query("snapshot_id"))
	if snapshotID == "" || len(snapshotID) > 128 {
		Fail(c, http.StatusBadRequest, Problem{Code: "REPORT_ANALYSIS_BAD_REQUEST", Title: problemTitleBadRequest, Detail: "Bản chụp báo cáo không hợp lệ."})
		return
	}
	latest, err := h.Service.Latest(c.Request.Context(), advisorPrincipal(c, owner), snapshotID)
	if err != nil {
		reportAnalysisMapError(c, err)
		return
	}
	OK(c, latest)
}

// Begin starts or recovers a report analysis using a stable client request ID.
// @Summary Start a report snapshot analysis
// @Tags Reports
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body reportAnalysisBeginInput true "Snapshot and stable client request ID"
// @Success 202 {object} map[string]any
// @Success 200 {object} map[string]any
// @Router /api/v1/reports/analyses [post]
func (h *ReportAnalysisHandler) Begin(c *gin.Context) {
	owner, ok := transactionOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	if h == nil || h.Service == nil {
		reportAnalysisUnavailable(c)
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 8*1024)
	var input reportAnalysisBeginInput
	if c.ShouldBindJSON(&input) != nil || strings.TrimSpace(input.SnapshotID) == "" || strings.TrimSpace(input.ClientRequestID) == "" {
		Fail(c, http.StatusBadRequest, Problem{Code: "REPORT_ANALYSIS_BAD_REQUEST", Title: problemTitleBadRequest, Detail: "Yêu cầu phân tích không hợp lệ."})
		return
	}
	principal := advisorPrincipal(c, owner)
	analysis, replay, err := h.Service.Begin(c.Request.Context(), principal, usecase.ReportAnalysisBeginRequest{SnapshotID: input.SnapshotID, ClientRequestID: input.ClientRequestID})
	if err != nil {
		reportAnalysisMapError(c, err)
		return
	}
	response := map[string]any{"id": analysis.ID, "snapshot_id": analysis.SnapshotID, "status": analysis.Status, "client_request_id": analysis.ClientRequestID, "server_time": analysis.UpdatedAt}
	if replay {
		OK(c, response)
		return
	}
	c.JSON(http.StatusAccepted, Response{Data: response, Meta: nil})
}

// Get reads one owner-bound analysis.
// @Summary Read a report analysis
// @Tags Reports
// @Produce json
// @Security BearerAuth
// @Param id path string true "Analysis ID"
// @Success 200 {object} map[string]any
// @Router /api/v1/reports/analyses/{id} [get]
func (h *ReportAnalysisHandler) Get(c *gin.Context) {
	owner, ok := transactionOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	if h == nil || h.Service == nil {
		reportAnalysisUnavailable(c)
		return
	}
	analysis, err := h.Service.Get(c.Request.Context(), advisorPrincipal(c, owner), c.Param("id"))
	if err != nil {
		reportAnalysisMapError(c, err)
		return
	}
	OK(c, analysis)
}

// Cancel cancels one owner-bound analysis.
// @Summary Cancel a report analysis
// @Tags Reports
// @Produce json
// @Security BearerAuth
// @Param id path string true "Analysis ID"
// @Success 200 {object} map[string]any
// @Router /api/v1/reports/analyses/{id}/cancel [post]
func (h *ReportAnalysisHandler) Cancel(c *gin.Context) {
	owner, ok := transactionOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	if h == nil || h.Service == nil {
		reportAnalysisUnavailable(c)
		return
	}
	analysis, err := h.Service.Cancel(c.Request.Context(), advisorPrincipal(c, owner), c.Param("id"))
	if err != nil {
		reportAnalysisMapError(c, err)
		return
	}
	OK(c, analysis)
}

func reportAnalysisMapError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, entity.ErrMonthlyReportScopeInvalid):
		Fail(c, http.StatusBadRequest, Problem{Code: "REPORT_ANALYSIS_BAD_REQUEST", Title: problemTitleBadRequest, Detail: "Yêu cầu phân tích không hợp lệ."})
	case errors.Is(err, usecase.ErrReportAnalysisProviderUnavailable):
		Fail(c, http.StatusServiceUnavailable, Problem{Code: "REPORT_ANALYSIS_PROVIDER_UNAVAILABLE", Title: problemTitleInternalServer, Detail: "Nhà cung cấp phân tích hiện chưa khả dụng."})
	case errors.Is(err, usecase.ErrReportAnalysisBusy):
		Fail(c, http.StatusConflict, Problem{Code: "REPORT_ANALYSIS_BUSY", Title: "Conflict", Detail: "Đang có một phân tích khác đang xử lý."})
	case errors.Is(err, usecase.ErrReportAnalysisCooldown):
		c.Header("Retry-After", "60")
		Fail(c, http.StatusTooManyRequests, Problem{Code: "REPORT_ANALYSIS_COOLDOWN", Title: "Too Many Requests", Detail: "Vui lòng thử lại sau thời gian chờ phân tích."})
	case errors.Is(err, usecase.ErrReportAnalysisRequestConflict):
		Fail(c, http.StatusConflict, Problem{Code: "REPORT_ANALYSIS_CONFLICT", Title: "Conflict", Detail: "Yêu cầu phân tích bị trùng với yêu cầu khác."})
	case errors.Is(err, entity.ErrReportSnapshotStale):
		Fail(c, http.StatusConflict, Problem{Code: "REPORT_SNAPSHOT_STALE", Title: "Conflict", Detail: "Cách tính báo cáo đã thay đổi. Hãy tải lại báo cáo trước khi phân tích."})
	case errors.Is(err, repository.ErrSnapshotExpired):
		Fail(c, http.StatusGone, Problem{Code: "REPORT_SNAPSHOT_EXPIRED", Title: problemTitleNotFound, Detail: "Bản chụp báo cáo đã hết hạn."})
	case errors.Is(err, repository.ErrNotFound):
		Fail(c, http.StatusNotFound, Problem{Code: "REPORT_ANALYSIS_NOT_FOUND", Title: problemTitleNotFound, Detail: "Không tìm thấy phân tích."})
	default:
		Fail(c, http.StatusInternalServerError, Problem{Code: "REPORT_ANALYSIS_FAILED", Title: problemTitleInternalServer, Detail: "Không thể xử lý phân tích báo cáo."})
	}
}

func reportAnalysisUnavailable(c *gin.Context) {
	Fail(c, http.StatusNotImplemented, Problem{Code: "REPORT_ANALYSIS_UNAVAILABLE", Title: problemTitleInternalServer, Detail: "Phân tích báo cáo chưa được cấu hình."})
}
