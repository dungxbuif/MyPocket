package httpapi

import (
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
	"gorm.io/gorm"
)

type MonthHandler struct {
	Users        repository.UserRepository
	Transactions repository.TransactionRepository
	Categories   repository.CategoryRepository
	Jars         repository.JarRepository
	Notes        repository.MonthNoteRepository
}

type monthNoteInput struct {
	Note string `json:"note"`
}

func (r *Router) RegisterMonthRoutes(h *MonthHandler) {
	g := r.Engine.Group("/api/v1/months")
	g.Use(r.AuthMiddleware.RequireAuth)
	g.GET("/:month", h.Get)
	g.PUT("/:month/note", h.SaveNote)
	g.DELETE("/:month/note", h.DeleteNote)
}

// GetMonth godoc
// @Summary Get live account-local month figures, jars, and independent note
// @Tags Months
// @Produce json
// @Security BearerAuth
// @Param month path string true "Calendar month YYYY-MM"
// @Success 200 {object} entity.MonthSummary
// @Router /api/v1/months/{month} [get]
func (h *MonthHandler) Get(c *gin.Context) {
	owner, ok := walletOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	month, err := entity.ParseMonth(c.Param("month"))
	if err != nil {
		monthFail(c, repository.ErrJarInvalid)
		return
	}
	user, err := h.Users.FindByID(owner)
	if err != nil {
		monthFail(c, err)
		return
	}
	location, err := time.LoadLocation(user.Timezone)
	if err != nil || user.Timezone == "Local" {
		monthFail(c, repository.ErrJarInvalid)
		return
	}
	start, next, err := entity.MonthRangeUTC(month, location)
	if err != nil {
		monthFail(c, err)
		return
	}
	transactions, err := h.Transactions.List(owner)
	if err != nil {
		monthFail(c, err)
		return
	}
	categories, err := h.Categories.ListVisible(owner)
	if err != nil {
		monthFail(c, err)
		return
	}
	categoryNames := make(map[string]string, len(categories))
	categoryKeys := make(map[string]string, len(categories))
	for _, category := range categories {
		categoryNames[category.ID] = category.Name
		if category.SystemKey != nil {
			categoryKeys[category.ID] = *category.SystemKey
		}
	}
	now := time.Now()
	monthKey := month.Format("2006-01")
	byCategory := map[string]*entity.MonthCategoryTotal{}
	result := entity.MonthSummary{Month: monthKey, Timezone: user.Timezone, StartAt: start, NextStartAt: next, IsCurrent: now.In(location).Format("2006-01") == monthKey, IsComplete: entity.IsMonthComplete(month, now, location), Categories: []entity.MonthCategoryTotal{}, CalculatedAt: now.UTC()}
	for _, transaction := range transactions {
		if transaction.OccurredAt.Before(start) || !transaction.OccurredAt.Before(next) || !transaction.IncludedInReports {
			continue
		}
		if transaction.Type == entity.TransactionTypeIncome && categoryKeys[valueOrEmpty(transaction.CategoryID)] == "income_transfer_in" {
			continue
		}
		if transaction.Type == entity.TransactionTypeExpense && categoryKeys[valueOrEmpty(transaction.CategoryID)] == "expense_transfer_out" {
			continue
		}
		if transaction.Type != entity.TransactionTypeIncome && transaction.Type != entity.TransactionTypeExpense {
			continue
		}
		result.TransactionCount++
		if transaction.Type == entity.TransactionTypeIncome {
			result.Income += transaction.Amount
		} else {
			result.Expense += transaction.Amount
		}
		categoryID := valueOrEmpty(transaction.CategoryID)
		key := transaction.Type + ":" + categoryID
		row := byCategory[key]
		if row == nil {
			label := categoryNames[categoryID]
			if label == "" {
				if transaction.Type == entity.TransactionTypeIncome {
					label = "Khoản thu chưa phân nhóm"
				} else {
					label = "Khoản chi chưa phân nhóm"
				}
			}
			row = &entity.MonthCategoryTotal{CategoryID: transaction.CategoryID, Name: label, Type: transaction.Type}
			byCategory[key] = row
		}
		row.Amount += transaction.Amount
		row.Count++
	}
	result.Net = result.Income - result.Expense
	for _, row := range byCategory {
		result.Categories = append(result.Categories, *row)
	}
	sort.Slice(result.Categories, func(i, j int) bool { return result.Categories[i].Amount > result.Categories[j].Amount })
	note, err := h.Notes.Find(owner, monthKey)
	if err != nil {
		monthFail(c, err)
		return
	}
	result.Note = note
	jarSummary, err := h.Jars.ListMonth(owner, monthKey, user.Timezone)
	if err != nil {
		monthFail(c, err)
		return
	}
	result.JarSummary = jarSummary
	OK(c, result)
}

// SaveMonthNote godoc
// @Summary Save the user's independent note for an account month
// @Tags Months
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param month path string true "Calendar month YYYY-MM"
// @Param note body monthNoteInput true "User note"
// @Success 204
// @Router /api/v1/months/{month}/note [put]
func (h *MonthHandler) SaveNote(c *gin.Context) {
	owner, ok := walletOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	month := c.Param("month")
	if _, err := entity.ParseMonth(month); err != nil {
		monthFail(c, repository.ErrJarInvalid)
		return
	}
	var input monthNoteInput
	if c.ShouldBindJSON(&input) != nil || len(input.Note) > 5000 {
		monthFail(c, repository.ErrJarInvalid)
		return
	}
	if err := h.Notes.Save(owner, month, strings.TrimSpace(input.Note)); err != nil {
		monthFail(c, err)
		return
	}
	NoContent(c)
}

// DeleteMonthNote godoc
// @Summary Delete the user's note for one account month
// @Tags Months
// @Security BearerAuth
// @Param month path string true "Calendar month YYYY-MM"
// @Success 204
// @Router /api/v1/months/{month}/note [delete]
func (h *MonthHandler) DeleteNote(c *gin.Context) {
	owner, ok := walletOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	month := c.Param("month")
	if _, err := entity.ParseMonth(month); err != nil {
		monthFail(c, repository.ErrJarInvalid)
		return
	}
	if err := h.Notes.Save(owner, month, ""); err != nil {
		monthFail(c, err)
		return
	}
	NoContent(c)
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func monthFail(c *gin.Context, err error) {
	status, code, title, message := http.StatusInternalServerError, "MONTH_FAILED", problemTitleInternalServer, "Không thể tải tổng kết tháng. Vui lòng thử lại."
	switch {
	case errors.Is(err, repository.ErrJarInvalid):
		status, code, title, message = http.StatusBadRequest, problemCodeBadRequest, problemTitleBadRequest, "Tháng phải có định dạng YYYY-MM hợp lệ."
	case errors.Is(err, gorm.ErrRecordNotFound), errors.Is(err, repository.ErrNotFound):
		status, code, title, message = http.StatusNotFound, "MONTH_NOT_FOUND", problemTitleNotFound, "Không tìm thấy tài khoản hoặc dữ liệu tháng."
	}
	Fail(c, status, Problem{Code: code, Title: title, Detail: message})
}
