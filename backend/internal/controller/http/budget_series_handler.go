package httpapi

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
	"gorm.io/gorm"
	"net/http"
	"strings"
	"time"
)

type BudgetSeriesHandler struct {
	Series repository.BudgetSeriesRepository
	Users  repository.UserRepository
	Now    func() time.Time
}
type budgetSeriesInput struct {
	Name             string  `json:"name"`
	LimitAmount      int64   `json:"limit_amount"`
	WalletID         *string `json:"wallet_id"`
	CategoryID       *string `json:"category_id"`
	CarryoverEnabled bool    `json:"carryover_enabled"`
	Cadence          *string `json:"cadence"`
	StartDate        *string `json:"start_date"`
	Version          int64   `json:"version"`
}
type budgetSeriesList struct {
	Items []entity.BudgetSeries `json:"items"`
}

func (h *BudgetSeriesHandler) now() time.Time {
	if h.Now != nil {
		return h.Now()
	}
	return time.Now()
}

// List godoc
// @Summary List recurring budget series and current/pending revisions
// @Tags Budget series
// @Produce json
// @Security BearerAuth
// @Success 200 {object} budgetSeriesList
// @Failure 401 {object} Problem
// @Router /api/v1/budget-series [get]
func (h *BudgetSeriesHandler) List(c *gin.Context) {
	owner, ok := walletOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	if h.Series == nil {
		budgetSeriesFailure(c, errors.New("series repository unavailable"))
		return
	}
	now := h.now()
	if _, err := h.Series.Ensure(owner, now); err != nil {
		budgetSeriesFailure(c, err)
		return
	}
	rows, err := h.Series.List(owner, now)
	if err != nil {
		budgetSeriesFailure(c, err)
		return
	}
	OK(c, budgetSeriesList{Items: rows})
}

// Create godoc
// @Summary Create a recurring budget series with canonical calendar periods
// @Tags Budget series
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Idempotency-Key header string true "Stable ASCII mutation key"
// @Param series body budgetSeriesInput true "Series configuration"
// @Success 201 {object} entity.BudgetSeries
// @Success 200 {object} entity.BudgetSeries
// @Failure 400 {object} Problem
// @Failure 409 {object} Problem
// @Router /api/v1/budget-series [post]
func (h *BudgetSeriesHandler) Create(c *gin.Context) { h.mutate(c, "create") }

// Update godoc
// @Summary Replace next-period recurring configuration with version protection
// @Tags Budget series
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Series ID"
// @Param Idempotency-Key header string true "Stable ASCII mutation key"
// @Param series body budgetSeriesInput true "Complete configuration and version; cadence/start_date omitted"
// @Success 200 {object} entity.BudgetSeries
// @Failure 400 {object} Problem
// @Failure 404 {object} Problem
// @Failure 409 {object} Problem
// @Router /api/v1/budget-series/{id} [patch]
func (h *BudgetSeriesHandler) Update(c *gin.Context) { h.mutate(c, "update") }

// Stop godoc
// @Summary Stop future generation while preserving existing budget periods
// @Tags Budget series
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Series ID"
// @Param Idempotency-Key header string true "Stable ASCII mutation key"
// @Param series body budgetSeriesInput true "Current version"
// @Success 200 {object} entity.BudgetSeries
// @Failure 400 {object} Problem
// @Failure 404 {object} Problem
// @Failure 409 {object} Problem
// @Router /api/v1/budget-series/{id}/stop [post]
func (h *BudgetSeriesHandler) Stop(c *gin.Context) { h.mutate(c, "stop") }

func (h *BudgetSeriesHandler) mutate(c *gin.Context, operation string) {
	owner, ok := walletOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	keys := c.Request.Header.Values("Idempotency-Key")
	key := ""
	if len(keys) == 1 {
		key = strings.TrimSpace(keys[0])
	}
	valid := len(keys) == 1 && len(key) > 0 && len(key) <= 128
	for _, ch := range key {
		if ch < 32 || ch > 126 {
			valid = false
		}
	}
	if !valid {
		budgetBadRequest(c, "Idempotency-Key là bắt buộc và phải là khóa ASCII hợp lệ.")
		return
	}
	var input budgetSeriesInput
	if c.ShouldBindJSON(&input) != nil {
		budgetBadRequest(c, "Dữ liệu chu kỳ không hợp lệ.")
		return
	}
	now := h.now()
	series := entity.BudgetSeries{ID: c.Param("id")}
	if operation != "create" {
		if _, err := uuid.Parse(series.ID); err != nil || input.Version <= 0 {
			budgetBadRequest(c, "Cần mã chu kỳ và phiên bản hợp lệ.")
			return
		}
		if input.Cadence != nil || input.StartDate != nil {
			budgetBadRequest(c, "Không thể thay đổi lịch hoặc ngày bắt đầu chu kỳ.")
			return
		}
	}
	var cfg *entity.BudgetSeriesConfig
	if operation != "stop" {
		input.Name = strings.TrimSpace(input.Name)
		input.WalletID = normalizeOptional(input.WalletID)
		input.CategoryID = normalizeOptional(input.CategoryID)
		if input.Name == "" || len(input.Name) > 200 || input.LimitAmount <= 0 || input.LimitAmount > entity.MaxBudgetMoney {
			budgetBadRequest(c, "Nhập tên và hạn mức nguyên dương hợp lệ.")
			return
		}
		for _, id := range []*string{input.WalletID, input.CategoryID} {
			if id != nil {
				if _, err := uuid.Parse(*id); err != nil {
					budgetBadRequest(c, "Phạm vi ví hoặc nhóm không hợp lệ.")
					return
				}
			}
		}
		cfg = &entity.BudgetSeriesConfig{Name: input.Name, LimitAmount: input.LimitAmount, WalletID: input.WalletID, CategoryID: input.CategoryID, CarryoverEnabled: input.CarryoverEnabled}
	}
	if operation == "create" {
		if input.Cadence == nil || input.StartDate == nil {
			budgetBadRequest(c, "Cần lịch và ngày bắt đầu chu kỳ.")
			return
		}
		start, err := entity.ParseCalendarDate(*input.StartDate)
		if err != nil {
			budgetBadRequest(c, "Ngày bắt đầu không hợp lệ.")
			return
		}
		if _, err = entity.NextBudgetPeriod(start, *input.Cadence); err != nil {
			budgetBadRequest(c, "Ngày bắt đầu phải là đầu tuần, tháng, quý hoặc năm.")
			return
		}
		// Time-dependent validation belongs after the receipt lookup in Mutate,
		// so retries in a later period still replay their original response.
		series.ID = uuid.NewString()
		series.StartDate = start
		series.Cadence = *input.Cadence
	}
	if h.Series == nil {
		budgetSeriesFailure(c, errors.New("series repository unavailable"))
		return
	}
	hash := hashMutationInput(struct {
		ID    string
		Input budgetSeriesInput
	}{c.Param("id"), input})
	result, replayed, err := h.Series.Mutate(owner, &series, cfg, operation, key, hash, input.Version, now)
	if err != nil {
		budgetSeriesFailure(c, err)
		return
	}
	if operation == "create" && !replayed {
		Created(c, result)
	} else {
		OK(c, result)
	}
}

func budgetSeriesFailure(c *gin.Context, err error) {
	if errors.Is(err, repository.ErrMutationConflict) || errors.Is(err, repository.ErrMutationPending) {
		createReceiptFailure(c, err)
		return
	}
	if errors.Is(err, repository.ErrBudgetSeriesVersion) {
		Fail(c, 409, Problem{Code: "BUDGET_SERIES_VERSION_CONFLICT", Title: http.StatusText(409), Detail: "Chu kỳ đã thay đổi. Tải lại trước khi sửa."})
		return
	}
	if errors.Is(err, repository.ErrBudgetSeriesStopped) {
		Fail(c, 409, Problem{Code: "BUDGET_SERIES_STOPPED", Title: http.StatusText(409), Detail: "Chu kỳ đã dừng."})
		return
	}
	if errors.Is(err, gorm.ErrInvalidData) || errors.Is(err, entity.ErrBudgetCadence) {
		budgetBadRequest(c, "Cấu hình chu kỳ không hợp lệ.")
		return
	}
	budgetFailure(c, err)
}
