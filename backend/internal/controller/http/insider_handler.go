package httpapi

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
)

type InsiderHandler struct {
	Reader repository.InsiderReader
	Users  repository.UserRepository
}

func (r *Router) RegisterInsiderRoutes(h *InsiderHandler) {
	g := r.Engine.Group("/api/v1/reports")
	// Reports are read-only finance data, so normal sessions and owner API keys
	// with finance:read can consume the same owner-scoped contract.
	g.Use(r.AuthMiddleware.RequireAdvisorAuth)
	g.GET("/insider", h.Get)
}

// GetMoneyInsider godoc
// @Summary Read the deterministic Money Insider report
// @Tags Reports
// @Produce json
// @Security BearerAuth
// @Param month query string false "Calendar month YYYY-MM; defaults to account-local current month"
// @Param wallet_id query string false "Optional owner wallet filter"
// @Param category_id query string false "Optional owner category filter"
// @Success 200 {object} entity.InsiderSummary
// @Failure 400 {object} Problem
// @Failure 401 {object} Problem
// @Failure 404 {object} Problem
// @Router /api/v1/reports/insider [get]
func (h *InsiderHandler) Get(c *gin.Context) {
	owner, ok := transactionOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	if h == nil || h.Reader == nil {
		insiderUnavailable(c)
		return
	}
	month := strings.TrimSpace(c.Query("month"))
	if month == "" {
		month = time.Now().UTC().Format("2006-01")
		if h.Users != nil {
			if user, err := h.Users.FindByID(owner); err == nil {
				if location, loadErr := time.LoadLocation(user.Timezone); loadErr == nil {
					month = time.Now().In(location).Format("2006-01")
				}
			}
		}
	}
	if _, err := entity.ParseMonth(month); err != nil {
		insiderBadRequest(c, "Tháng phải có định dạng YYYY-MM hợp lệ.")
		return
	}
	walletID, categoryID := strings.TrimSpace(c.Query("wallet_id")), strings.TrimSpace(c.Query("category_id"))
	if len(walletID) > 128 || len(categoryID) > 128 {
		insiderBadRequest(c, "Bộ lọc báo cáo không hợp lệ.")
		return
	}
	result, err := h.Reader.ReadInsider(owner, month, walletID, categoryID)
	if err != nil {
		insiderMapError(c, err)
		return
	}
	OK(c, result)
}

func insiderMapError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, entity.ErrFinanceDateInvalid), errors.Is(err, entity.ErrFinanceRangeInvalid):
		insiderBadRequest(c, "Tháng báo cáo không hợp lệ.")
	case errors.Is(err, repository.ErrNotFound):
		Fail(c, http.StatusNotFound, Problem{Code: "INSIDER_NOT_FOUND", Title: problemTitleNotFound, Detail: "Không tìm thấy tài khoản báo cáo."})
	default:
		Fail(c, http.StatusInternalServerError, Problem{Code: "INSIDER_FAILED", Title: problemTitleInternalServer, Detail: "Không thể tải Money Insider."})
	}
}

func insiderBadRequest(c *gin.Context, detail string) {
	Fail(c, http.StatusBadRequest, Problem{Code: "INSIDER_BAD_REQUEST", Title: problemTitleBadRequest, Detail: detail})
}

func insiderUnavailable(c *gin.Context) {
	Fail(c, http.StatusNotImplemented, Problem{Code: "INSIDER_UNAVAILABLE", Title: problemTitleInternalServer, Detail: "Money Insider chưa được cấu hình."})
}
