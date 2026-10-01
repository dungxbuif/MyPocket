package httpapi

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	categoryrepo "github.com/mypocket/backend/internal/repository"
	recurringrepo "github.com/mypocket/backend/internal/repository"
)

type RecurringHandler struct {
	Schedules  recurringrepo.RecurringRepository
	Wallets    categoryrepo.WalletRepository
	Categories categoryrepo.CategoryRepository
	Users      categoryrepo.UserRepository
}

type recurringInput struct {
	Name       string  `json:"name"`
	WalletID   string  `json:"wallet_id"`
	CategoryID *string `json:"category_id"`
	Type       string  `json:"type"`
	Amount     int64   `json:"amount"`
	Note       *string `json:"note"`
	Frequency  string  `json:"frequency"`
	Interval   int     `json:"interval"`
	NextRunAt  string  `json:"next_run_at"`
	EndsAt     *string `json:"ends_at"`
	Active     *bool   `json:"active"`
}

func (r *Router) RegisterRecurringRoutes(h *RecurringHandler) {
	g := r.Engine.Group("/api/v1/recurring")
	g.Use(r.AuthMiddleware.RequireAuth)
	g.GET("", h.List)
	g.POST("", h.Create)
	g.POST("/run-due", h.RunDue)
	g.PATCH("/:id", h.Update)
	g.DELETE("/:id", h.Delete)
}

// ListRecurring godoc
// @Summary List recurring transaction schedules
// @Tags Recurring
// @Produce json
// @Security BearerAuth
// @Success 200 {array} entity.RecurringSchedule
// @Router /api/v1/recurring [get]
func (h *RecurringHandler) List(c *gin.Context) {
	owner, ok := transactionOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	rows, err := h.Schedules.List(owner)
	if err != nil {
		recurringFailure(c)
		return
	}
	OK(c, rows)
}

// CreateRecurring godoc
// @Summary Create a recurring transaction schedule
// @Tags Recurring
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param schedule body recurringInput true "Recurring schedule"
// @Success 201 {object} entity.RecurringSchedule
// @Failure 400 {object} Problem
// @Router /api/v1/recurring [post]
func (h *RecurringHandler) Create(c *gin.Context) {
	owner, ok := transactionOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	input, valid := h.bind(c, owner, nil)
	if !valid {
		return
	}
	schedule := input
	schedule.ID, schedule.OwnerID = uuid.NewString(), owner
	if err := h.Schedules.Create(owner, &schedule); err != nil {
		h.mapError(c, err)
		return
	}
	Created(c, &schedule)
}

// UpdateRecurring godoc
// @Summary Update a recurring transaction schedule
// @Tags Recurring
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Schedule ID"
// @Param schedule body recurringInput true "Recurring schedule"
// @Success 200 {object} entity.RecurringSchedule
// @Router /api/v1/recurring/{id} [patch]
func (h *RecurringHandler) Update(c *gin.Context) {
	owner, ok := transactionOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	current, err := h.Schedules.Find(owner, c.Param("id"))
	if err != nil {
		recurringNotFound(c)
		return
	}
	input, valid := h.bind(c, owner, current)
	if !valid {
		return
	}
	active := input.Active
	updated, err := h.Schedules.Update(owner, current.ID, recurringrepo.RecurringUpdate{Name: input.Name, Amount: input.Amount, CategoryID: input.CategoryID, Type: input.Type, Note: input.Note, Frequency: input.Frequency, Interval: input.Interval, NextRunAt: input.NextRunAt, EndsAt: input.EndsAt, Active: active})
	if err != nil {
		h.mapError(c, err)
		return
	}
	OK(c, updated)
}

// DeleteRecurring godoc
// @Summary Delete a recurring transaction schedule
// @Tags Recurring
// @Security BearerAuth
// @Param id path string true "Schedule ID"
// @Success 204
// @Router /api/v1/recurring/{id} [delete]
func (h *RecurringHandler) Delete(c *gin.Context) {
	owner, ok := transactionOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	if err := h.Schedules.Delete(owner, c.Param("id")); err != nil {
		recurringNotFound(c)
		return
	}
	NoContent(c)
}

// RunDueRecurring godoc
// @Summary Materialize due recurring transaction occurrences
// @Tags Recurring
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]int
// @Router /api/v1/recurring/run-due [post]
func (h *RecurringHandler) RunDue(c *gin.Context) {
	owner, ok := transactionOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	created, err := h.Schedules.RunDue(owner, time.Now().UTC())
	if err != nil {
		h.mapError(c, err)
		return
	}
	OK(c, map[string]int{"created": created})
}

func (h *RecurringHandler) bind(c *gin.Context, owner string, existing *entity.RecurringSchedule) (entity.RecurringSchedule, bool) {
	var input recurringInput
	if c.ShouldBindJSON(&input) != nil {
		transactionBadRequest(c, problemDetailInvalidJSON)
		return entity.RecurringSchedule{}, false
	}
	input.Name, input.WalletID, input.Type, input.Frequency = strings.TrimSpace(input.Name), strings.TrimSpace(input.WalletID), strings.TrimSpace(input.Type), strings.TrimSpace(input.Frequency)
	if existing != nil && input.WalletID == "" {
		input.WalletID = existing.WalletID
	}
	if input.Active == nil {
		value := true
		if existing != nil {
			value = existing.Active
		}
		input.Active = &value
	}
	if input.Interval <= 0 {
		if existing != nil {
			input.Interval = existing.Interval
		} else {
			input.Interval = 1
		}
	}
	if input.Name == "" || input.Amount <= 0 || (input.Type != entity.TransactionTypeIncome && input.Type != entity.TransactionTypeExpense) || (input.Frequency != "daily" && input.Frequency != "weekly" && input.Frequency != "monthly" && input.Frequency != "yearly") {
		transactionBadRequest(c, "Lịch định kỳ chưa hợp lệ.")
		return entity.RecurringSchedule{}, false
	}
	wallet, err := h.Wallets.Find(owner, input.WalletID)
	if err != nil {
		Fail(c, http.StatusNotFound, Problem{Code: problemCodeTransactionWalletNotFound, Title: problemTitleNotFound, Detail: transactionWalletMessage})
		return entity.RecurringSchedule{}, false
	}
	if wallet.Type == entity.WalletTypeCredit {
		transactionBadRequest(c, transactionCreditWalletMessage)
		return entity.RecurringSchedule{}, false
	}
	input.CategoryID = normalizeOptional(input.CategoryID)
	if input.CategoryID != nil && !categoryValidForRecurring(h.Categories, owner, *input.CategoryID, input.Type) {
		transactionBadRequest(c, transactionCategoryMessage)
		return entity.RecurringSchedule{}, false
	}
	user, err := h.Users.FindByID(owner)
	if err != nil {
		recurringFailure(c)
		return entity.RecurringSchedule{}, false
	}
	location, err := time.LoadLocation(user.Timezone)
	if err != nil || user.Timezone == "Local" {
		transactionBadRequest(c, "Múi giờ tài khoản không hợp lệ.")
		return entity.RecurringSchedule{}, false
	}
	next := time.Now().In(location)
	if strings.TrimSpace(input.NextRunAt) != "" {
		parsed, parseErr := time.Parse(time.RFC3339, input.NextRunAt)
		if parseErr != nil {
			transactionBadRequest(c, transactionDateMessage)
			return entity.RecurringSchedule{}, false
		}
		next = parsed.In(location)
	}
	ends := (*time.Time)(nil)
	if input.EndsAt != nil && strings.TrimSpace(*input.EndsAt) != "" {
		parsed, parseErr := time.Parse(time.RFC3339, *input.EndsAt)
		if parseErr != nil {
			transactionBadRequest(c, transactionDateMessage)
			return entity.RecurringSchedule{}, false
		}
		ends = &parsed
	}
	if ends != nil && ends.Before(next) {
		transactionBadRequest(c, "Ngày kết thúc phải sau lần chạy kế tiếp.")
		return entity.RecurringSchedule{}, false
	}
	return entity.RecurringSchedule{Name: input.Name, OwnerID: owner, WalletID: wallet.ID, CategoryID: input.CategoryID, Type: input.Type, Amount: input.Amount, Note: normalizeOptional(input.Note), Frequency: input.Frequency, Interval: input.Interval, NextRunAt: next.UTC(), EndsAt: ends, AnchorDay: next.Day(), AnchorMonth: int(next.Month()), Active: *input.Active}, true
}

func categoryValidForRecurring(categories categoryrepo.CategoryRepository, owner, id, kind string) bool {
	row, err := categories.FindVisible(owner, id)
	return err == nil && row != nil && row.Kind == kind
}

func (h *RecurringHandler) mapError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, recurringrepo.ErrRecurringNotFound):
		recurringNotFound(c)
	case errors.Is(err, recurringrepo.ErrRecurringWalletInvalid), errors.Is(err, recurringrepo.ErrRecurringCategoryInvalid), errors.Is(err, recurringrepo.ErrRecurringInvalid):
		transactionBadRequest(c, "Lịch định kỳ chưa hợp lệ.")
	default:
		recurringFailure(c)
	}
}
func recurringNotFound(c *gin.Context) {
	Fail(c, http.StatusNotFound, Problem{Code: problemCodeTransactionNotFound, Title: problemTitleNotFound, Detail: "không tìm thấy lịch định kỳ"})
}
func recurringFailure(c *gin.Context) {
	Fail(c, http.StatusInternalServerError, Problem{Code: problemCodeTransactionSaveFailed, Title: problemTitleInternalServer, Detail: "không xử lý được lịch định kỳ"})
}
