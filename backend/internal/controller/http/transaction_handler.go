package httpapi

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	categoryrepo "github.com/mypocket/backend/internal/repository"
	transactionrepo "github.com/mypocket/backend/internal/repository"
	walletrepo "github.com/mypocket/backend/internal/repository"
)

const (
	transactionUnauthorizedMessage = "chưa đăng nhập"
	transactionAmountMessage       = "số tiền phải lớn hơn 0"
	transactionTypeMessage         = "loại giao dịch không hợp lệ"
	transactionWalletMessage       = "ví không tồn tại hoặc không thuộc tài khoản"
	transactionCategoryMessage     = "nhóm không phù hợp với loại giao dịch"
	transactionDateMessage         = "thời điểm giao dịch không hợp lệ"
	transactionNotFoundMessage     = "không tìm thấy giao dịch"
	transactionLoadMessage         = "không đọc được giao dịch"
	transactionSaveMessage         = "không lưu được giao dịch"
)

type TransactionHandler struct {
	Transactions transactionrepo.TransactionRepository
	Wallets      walletrepo.WalletRepository
	Categories   categoryrepo.CategoryRepository
}

type transactionInput struct {
	WalletID          string  `json:"wallet_id"`
	CategoryID        *string `json:"category_id"`
	Type              string  `json:"type"`
	Amount            int64   `json:"amount"`
	OccurredAt        string  `json:"occurred_at"`
	Note              *string `json:"note"`
	IncludedInReports *bool   `json:"included_in_reports"`
}

func NewTransactionHandler(transactions transactionrepo.TransactionRepository, wallets walletrepo.WalletRepository, categories categoryrepo.CategoryRepository) *TransactionHandler {
	return &TransactionHandler{Transactions: transactions, Wallets: wallets, Categories: categories}
}

// ListTransactions godoc
// @Summary List transactions
// @Tags Transactions
// @Produce json
// @Security BearerAuth
// @Success 200 {array} entity.Transaction
// @Failure 401 {object} Problem
// @Router /api/v1/transactions [get]
func (h *TransactionHandler) ListTransactions(c *gin.Context) {
	owner, ok := transactionOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	transactions, err := h.Transactions.List(owner)
	if err != nil {
		Fail(c, http.StatusInternalServerError, Problem{Code: problemCodeTransactionLoadFailed, Title: problemTitleInternalServer, Detail: transactionLoadMessage})
		return
	}
	OK(c, transactions)
}

// CreateTransaction godoc
// @Summary Create an income or expense transaction
// @Tags Transactions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param transaction body transactionInput true "Transaction input"
// @Success 201 {object} entity.Transaction
// @Failure 400 {object} Problem
// @Failure 401 {object} Problem
// @Failure 404 {object} Problem
// @Router /api/v1/transactions [post]
func (h *TransactionHandler) CreateTransaction(c *gin.Context) {
	owner, ok := transactionOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	input, occurredAt, valid := h.bindAndValidate(c, owner)
	if !valid {
		return
	}
	included := true
	if input.IncludedInReports != nil {
		included = *input.IncludedInReports
	}
	transaction := &entity.Transaction{ID: uuid.NewString(), OwnerID: owner, WalletID: input.WalletID, CategoryID: input.CategoryID, Type: input.Type, Amount: input.Amount, OccurredAt: occurredAt, Note: normalizeOptional(input.Note), IncludedInReports: included}
	if err := h.Transactions.Create(transaction); err != nil {
		Fail(c, http.StatusInternalServerError, Problem{Code: problemCodeTransactionSaveFailed, Title: problemTitleInternalServer, Detail: transactionSaveMessage})
		return
	}
	Created(c, transaction)
}

// UpdateTransaction godoc
// @Summary Update an income or expense transaction
// @Tags Transactions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Transaction ID"
// @Param transaction body transactionInput true "Transaction input"
// @Success 200 {object} entity.Transaction
// @Failure 400 {object} Problem
// @Failure 401 {object} Problem
// @Failure 404 {object} Problem
// @Router /api/v1/transactions/{id} [patch]
func (h *TransactionHandler) UpdateTransaction(c *gin.Context) {
	owner, ok := transactionOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	input, occurredAt, valid := h.bindAndValidate(c, owner)
	if !valid {
		return
	}
	updates := map[string]any{"wallet_id": input.WalletID, "category_id": input.CategoryID, "type": input.Type, "amount": input.Amount, "occurred_at": occurredAt, "note": normalizeOptional(input.Note)}
	if input.IncludedInReports != nil {
		updates["included_in_reports"] = *input.IncludedInReports
	}
	transaction, err := h.Transactions.Update(owner, c.Param("id"), updates)
	if err != nil {
		Fail(c, http.StatusNotFound, Problem{Code: problemCodeTransactionNotFound, Title: problemTitleNotFound, Detail: transactionNotFoundMessage})
		return
	}
	OK(c, transaction)
}

// DeleteTransaction godoc
// @Summary Delete a transaction
// @Tags Transactions
// @Security BearerAuth
// @Param id path string true "Transaction ID"
// @Success 204
// @Failure 401 {object} Problem
// @Failure 404 {object} Problem
// @Router /api/v1/transactions/{id} [delete]
func (h *TransactionHandler) DeleteTransaction(c *gin.Context) {
	owner, ok := transactionOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	if err := h.Transactions.Delete(owner, c.Param("id")); err != nil {
		Fail(c, http.StatusNotFound, Problem{Code: problemCodeTransactionNotFound, Title: problemTitleNotFound, Detail: transactionNotFoundMessage})
		return
	}
	NoContent(c)
}

func (h *TransactionHandler) bindAndValidate(c *gin.Context, owner string) (transactionInput, time.Time, bool) {
	var input transactionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		Fail(c, http.StatusBadRequest, Problem{Code: problemCodeBadRequest, Title: problemTitleBadRequest, Detail: problemDetailInvalidJSON})
		return transactionInput{}, time.Time{}, false
	}
	input.WalletID, input.Type = strings.TrimSpace(input.WalletID), strings.TrimSpace(input.Type)
	if input.Amount <= 0 {
		transactionBadRequest(c, transactionAmountMessage)
		return transactionInput{}, time.Time{}, false
	}
	if input.Type != entity.TransactionTypeIncome && input.Type != entity.TransactionTypeExpense {
		transactionBadRequest(c, transactionTypeMessage)
		return transactionInput{}, time.Time{}, false
	}
	if _, err := h.Wallets.Find(owner, input.WalletID); err != nil {
		Fail(c, http.StatusNotFound, Problem{Code: problemCodeTransactionWalletNotFound, Title: problemTitleNotFound, Detail: transactionWalletMessage})
		return transactionInput{}, time.Time{}, false
	}
	if input.CategoryID != nil && !h.categoryValid(owner, *input.CategoryID, input.Type) {
		transactionBadRequest(c, transactionCategoryMessage)
		return transactionInput{}, time.Time{}, false
	}
	occurredAt := time.Now().UTC()
	if strings.TrimSpace(input.OccurredAt) != "" {
		parsed, err := time.Parse(time.RFC3339, input.OccurredAt)
		if err != nil {
			transactionBadRequest(c, transactionDateMessage)
			return transactionInput{}, time.Time{}, false
		}
		occurredAt = parsed.UTC()
	}
	return input, occurredAt, true
}

func (h *TransactionHandler) categoryValid(owner, categoryID, kind string) bool {
	categories, err := h.Categories.ListVisible(owner)
	if err != nil {
		return false
	}
	for _, category := range categories {
		if category.ID == categoryID && category.Kind == kind {
			return true
		}
	}
	return false
}

func transactionOwner(c *gin.Context) (string, bool) { return walletOwner(c) }
func transactionUnauthorized(c *gin.Context) {
	Fail(c, http.StatusUnauthorized, Problem{Code: problemCodeAuthRequired, Title: problemTitleUnauthorized, Detail: transactionUnauthorizedMessage})
}
func transactionBadRequest(c *gin.Context, detail string) {
	Fail(c, http.StatusBadRequest, Problem{Code: problemCodeBadRequest, Title: problemTitleBadRequest, Detail: detail})
}
func normalizeOptional(value *string) *string {
	if value == nil {
		return nil
	}
	normalized := strings.TrimSpace(*value)
	if normalized == "" {
		return nil
	}
	return &normalized
}
