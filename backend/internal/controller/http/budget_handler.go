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

type BudgetHandler struct {
	Budgets      repository.BudgetRepository
	Wallets      repository.WalletRepository
	Categories   repository.CategoryRepository
	Transactions repository.TransactionRepository
	Users        repository.UserRepository
	Series       repository.BudgetSeriesRepository
}
type budgetInput struct {
	Name        string  `json:"name"`
	LimitAmount int64   `json:"limit_amount"`
	WalletID    *string `json:"wallet_id"`
	CategoryID  *string `json:"category_id"`
	StartDate   string  `json:"start_date"`
	EndDate     string  `json:"end_date"`
}

// ListBudgets godoc
// @Summary List budgets with ledger-derived progress and active deduplicated totals
// @Tags Budgets
// @Produce json
// @Security BearerAuth
// @Success 200 {object} entity.BudgetSummary
// @Failure 401 {object} Problem
// @Router /api/v1/budgets [get]
func (h *BudgetHandler) ListBudgets(c *gin.Context) {
	owner, ok := walletOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	now := time.Now()
	if h.Series != nil {
		if _, err := h.Series.Ensure(owner, now); err != nil {
			budgetFailure(c, err)
			return
		}
	}
	budgets, err := h.Budgets.List(owner)
	if err != nil {
		budgetFailure(c, err)
		return
	}
	transactions, err := h.Transactions.List(owner)
	if err != nil {
		budgetFailure(c, err)
		return
	}
	categories, err := h.Categories.ListVisible(owner)
	if err != nil {
		budgetFailure(c, err)
		return
	}
	if h.Users == nil {
		budgetFailure(c, errors.New("account timezone repository unavailable"))
		return
	}
	user, err := h.Users.FindByID(owner)
	if err != nil {
		budgetFailure(c, err)
		return
	}
	location, err := time.LoadLocation(user.Timezone)
	if err != nil {
		budgetFailure(c, err)
		return
	}
	summary, err := entity.CalculateBudgetsChecked(budgets, transactions, categories, now, location)
	if err != nil {
		budgetFailure(c, err)
		return
	}
	OK(c, summary)
}

// CreateBudget godoc
// @Summary Create a budget for an explicit interval
// @Tags Budgets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param budget body budgetInput true "Budget input"
// @Param Idempotency-Key header string false "Stable ASCII key reused to replay a confirmed budget write"
// @Success 201 {object} entity.Budget
// @Failure 400 {object} Problem
// @Failure 409 {object} Problem
// @Router /api/v1/budgets [post]
func (h *BudgetHandler) CreateBudget(c *gin.Context) { h.save(c, true) }

// UpdateBudget godoc
// @Summary Replace budget configuration before its interval has ended
// @Tags Budgets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Budget ID"
// @Param budget body budgetInput true "Budget input"
// @Success 200 {object} entity.Budget
// @Failure 400 {object} Problem
// @Failure 404 {object} Problem
// @Failure 409 {object} Problem
// @Router /api/v1/budgets/{id} [patch]
func (h *BudgetHandler) UpdateBudget(c *gin.Context) { h.save(c, false) }

func (h *BudgetHandler) save(c *gin.Context, creating bool) {
	owner, ok := walletOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	var input budgetInput
	if c.ShouldBindJSON(&input) != nil {
		budgetBadRequest(c, "Dữ liệu ngân sách không hợp lệ.")
		return
	}
	start, e1 := entity.ParseCalendarDate(input.StartDate)
	end, e2 := entity.ParseCalendarDate(input.EndDate)
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" || len(input.Name) > 200 || input.LimitAmount <= 0 || input.LimitAmount > 9007199254740991 || e1 != nil || e2 != nil || end.Time.Before(start.Time) {
		budgetBadRequest(c, "Nhập tên, hạn mức nguyên dương và khoảng ngày hợp lệ.")
		return
	}
	input.WalletID = normalizeOptional(input.WalletID)
	input.CategoryID = normalizeOptional(input.CategoryID)
	id := c.Param("id")
	if creating {
		id = uuid.NewString()
	} else if _, err := uuid.Parse(id); err != nil {
		budgetBadRequest(c, "Ngân sách không hợp lệ.")
		return
	}
	if input.WalletID != nil {
		if _, err := uuid.Parse(*input.WalletID); err != nil {
			budgetBadRequest(c, "Ví không hợp lệ.")
			return
		}
		wallet, err := h.Wallets.Find(owner, *input.WalletID)
		if err != nil || wallet.Type == entity.WalletTypeCredit {
			budgetBadRequest(c, "Ví không tồn tại hoặc chưa hỗ trợ ngân sách.")
			return
		}
	}
	if input.CategoryID != nil {
		if _, err := uuid.Parse(*input.CategoryID); err != nil {
			budgetBadRequest(c, "Nhóm không hợp lệ.")
			return
		}
		category, err := h.Categories.FindVisible(owner, *input.CategoryID)
		if err != nil || category.Kind != entity.TransactionTypeExpense {
			budgetBadRequest(c, "Chọn nhóm chi thuộc tài khoản.")
			return
		}
		if !category.IsActive {
			keepHistoricalCategory := false
			if !creating {
				budgets, listErr := h.Budgets.List(owner)
				if listErr != nil {
					budgetFailure(c, listErr)
					return
				}
				for _, current := range budgets {
					if current.ID == id && current.CategoryID != nil && *current.CategoryID == category.ID {
						keepHistoricalCategory = true
						break
					}
				}
			}
			if !keepHistoricalCategory {
				categoryInactive(c)
				return
			}
		}
	}
	budget := entity.Budget{ID: id, OwnerID: owner, Name: input.Name, LimitAmount: input.LimitAmount, WalletID: input.WalletID, CategoryID: input.CategoryID, StartDate: start, EndDate: end}
	if creating {
		keyHeaders := c.Request.Header.Values("Idempotency-Key")
		if len(keyHeaders) > 1 {
			Fail(c, http.StatusBadRequest, Problem{Code: problemCodeBadRequest, Title: problemTitleBadRequest, Detail: "Idempotency-Key không hợp lệ."})
			return
		}
		idempotencyKey := ""
		if len(keyHeaders) == 1 {
			idempotencyKey = strings.TrimSpace(keyHeaders[0])
		}
		validKey := len(idempotencyKey) <= 128
		for _, character := range idempotencyKey {
			if character < 32 || character > 126 {
				validKey = false
			}
		}
		if !validKey {
			Fail(c, http.StatusBadRequest, Problem{Code: problemCodeBadRequest, Title: problemTitleBadRequest, Detail: "Idempotency-Key không hợp lệ."})
			return
		}
		if idempotencyKey != "" {
			mutator, ok := h.Budgets.(repository.BudgetCreateMutationRepository)
			if !ok {
				Fail(c, http.StatusServiceUnavailable, Problem{Code: "mutation_unavailable", Title: problemTitleInternalServer, Detail: "Chưa thể xử lý yêu cầu có khóa chống trùng."})
				return
			}
			persisted, replayed, createErr := mutator.CreateIdempotent(owner, &budget, idempotencyKey, hashMutationInput(input))
			if createErr != nil {
				if !createReceiptFailure(c, createErr) {
					budgetFailure(c, createErr)
				}
				return
			}
			if replayed {
				OK(c, persisted)
				return
			}
			Created(c, persisted)
			return
		}
	}
	if err := h.Budgets.Save(owner, &budget, creating); err != nil {
		budgetFailure(c, err)
		return
	}
	if creating {
		Created(c, budget)
	} else {
		OK(c, budget)
	}
}

// DeleteBudget godoc
// @Summary Delete budget configuration without changing transactions
// @Tags Budgets
// @Security BearerAuth
// @Param id path string true "Budget ID"
// @Success 204
// @Failure 404 {object} Problem
// @Router /api/v1/budgets/{id} [delete]
func (h *BudgetHandler) DeleteBudget(c *gin.Context) {
	owner, ok := walletOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	if _, err := uuid.Parse(c.Param("id")); err != nil {
		budgetBadRequest(c, "Ngân sách không hợp lệ.")
		return
	}
	if err := h.Budgets.Delete(owner, c.Param("id")); err != nil {
		budgetFailure(c, err)
		return
	}
	NoContent(c)
}
func budgetBadRequest(c *gin.Context, message string) {
	Fail(c, 400, Problem{Code: problemCodeBadRequest, Title: problemTitleBadRequest, Detail: message})
}
func budgetFailure(c *gin.Context, err error) {
	if errors.Is(err, repository.ErrCategoryInactive) {
		categoryInactive(c)
		return
	}
	status, code, message := http.StatusInternalServerError, "BUDGET_FAILED", "Không thể xử lý ngân sách. Vui lòng thử lại."
	if errors.Is(err, repository.ErrBudgetSeriesManaged) {
		status = 409
		code = "BUDGET_SERIES_MANAGED"
		message = "Ngân sách thuộc chu kỳ. Sửa hoặc dừng chu kỳ để thay đổi kỳ tiếp theo."
	}
	if errors.Is(err, entity.ErrBudgetMoneyOverflow) {
		status = 422
		code = "BUDGET_AMOUNT_OVERFLOW"
		message = "Tổng ngân sách vượt giới hạn số tiền được hỗ trợ."
	}
	if errors.Is(err, repository.ErrBudgetConflict) {
		status = 409
		code = "BUDGET_OVERLAP"
		message = "Đã có ngân sách cùng ví, nhóm và khoảng ngày chồng lấn."
	}
	if errors.Is(err, repository.ErrBudgetEnded) {
		status = 409
		code = "BUDGET_ENDED"
		message = "Ngân sách đã kết thúc, chỉ có thể xem hoặc xóa."
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		status = 404
		code = "BUDGET_NOT_FOUND"
		message = "Không tìm thấy ngân sách."
	}
	Fail(c, status, Problem{Code: code, Title: http.StatusText(status), Detail: message})
}
