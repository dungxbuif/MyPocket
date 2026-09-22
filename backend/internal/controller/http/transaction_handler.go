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
	transactionrepo "github.com/mypocket/backend/internal/repository"
	walletrepo "github.com/mypocket/backend/internal/repository"
)

const (
	transactionUnauthorizedMessage = "chưa đăng nhập"
	transactionAmountMessage       = "số tiền phải lớn hơn 0"
	transactionTypeMessage         = "loại giao dịch không hợp lệ"
	transactionWalletMessage       = "ví không tồn tại hoặc không thuộc tài khoản"
	transactionCreditWalletMessage = "ví tín dụng cần luồng giao dịch tín dụng riêng"
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
	Users        categoryrepo.UserRepository
	Jars         categoryrepo.JarRepository
	Transfers    transactionrepo.TransferRepository
}

type transactionInput struct {
	WalletID          string  `json:"wallet_id"`
	CategoryID        *string `json:"category_id"`
	JarID             *string `json:"jar_id"`
	Type              string  `json:"type"`
	Amount            int64   `json:"amount"`
	OccurredAt        string  `json:"occurred_at"`
	Note              *string `json:"note"`
	IncludedInReports *bool   `json:"included_in_reports"`
}

func NewTransactionHandler(transactions transactionrepo.TransactionRepository, wallets walletrepo.WalletRepository, categories categoryrepo.CategoryRepository) *TransactionHandler {
	handler := &TransactionHandler{Transactions: transactions, Wallets: wallets, Categories: categories}
	if transfers, ok := transactions.(transactionrepo.TransferRepository); ok {
		handler.Transfers = transfers
	}
	return handler
}

type transferInput struct {
	SourceWalletID      string  `json:"source_wallet_id"`
	DestinationWalletID string  `json:"destination_wallet_id"`
	Amount              int64   `json:"amount"`
	OccurredAt          string  `json:"occurred_at"`
	Note                *string `json:"note"`
}

// CreateTransfer godoc
// @Summary Create an atomic internal wallet transfer
// @Tags Transactions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param transfer body transferInput true "Internal transfer input"
// @Success 201 {array} entity.Transaction
// @Failure 400 {object} Problem
// @Failure 401 {object} Problem
// @Failure 404 {object} Problem
// @Failure 501 {object} Problem
// @Router /api/v1/transactions/transfer [post]
func (h *TransactionHandler) CreateTransfer(c *gin.Context) {
	owner, ok := transactionOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	if h.Transfers == nil {
		Fail(c, http.StatusNotImplemented, Problem{Code: problemCodeTransactionSaveFailed, Title: problemTitleInternalServer, Detail: "luồng chuyển ví chưa được cấu hình"})
		return
	}
	var input transferInput
	if err := c.ShouldBindJSON(&input); err != nil {
		Fail(c, http.StatusBadRequest, Problem{Code: problemCodeBadRequest, Title: problemTitleBadRequest, Detail: problemDetailInvalidJSON})
		return
	}
	input.SourceWalletID, input.DestinationWalletID = strings.TrimSpace(input.SourceWalletID), strings.TrimSpace(input.DestinationWalletID)
	if input.Amount <= 0 {
		transactionBadRequest(c, transactionAmountMessage)
		return
	}
	if input.SourceWalletID == "" || input.DestinationWalletID == "" || input.SourceWalletID == input.DestinationWalletID {
		transactionBadRequest(c, "Ví chuyển đi và ví nhận phải khác nhau.")
		return
	}
	sourceWallet, sourceErr := h.Wallets.Find(owner, input.SourceWalletID)
	destinationWallet, destinationErr := h.Wallets.Find(owner, input.DestinationWalletID)
	if sourceErr != nil || destinationErr != nil {
		Fail(c, http.StatusNotFound, Problem{Code: problemCodeTransactionWalletNotFound, Title: problemTitleNotFound, Detail: transactionWalletMessage})
		return
	}
	if sourceWallet.Type == entity.WalletTypeCredit || destinationWallet.Type == entity.WalletTypeCredit {
		transactionBadRequest(c, transactionCreditWalletMessage)
		return
	}
	occurredAt := time.Now().UTC()
	if strings.TrimSpace(input.OccurredAt) != "" {
		parsed, err := time.Parse(time.RFC3339, input.OccurredAt)
		if err != nil {
			transactionBadRequest(c, transactionDateMessage)
			return
		}
		occurredAt = parsed.UTC()
	}
	var sourceCategoryID, destinationCategoryID string
	categories, err := h.Categories.ListVisible(owner)
	if err != nil {
		Fail(c, http.StatusInternalServerError, Problem{Code: problemCodeTransactionSaveFailed, Title: problemTitleInternalServer, Detail: transactionSaveMessage})
		return
	}
	for _, category := range categories {
		if category.SystemKey == nil {
			continue
		}
		switch *category.SystemKey {
		case "expense_transfer_out":
			sourceCategoryID = category.ID
		case "income_transfer_in":
			destinationCategoryID = category.ID
		}
	}
	if sourceCategoryID == "" || destinationCategoryID == "" {
		Fail(c, http.StatusInternalServerError, Problem{Code: problemCodeTransactionSaveFailed, Title: problemTitleInternalServer, Detail: "thiếu nhóm hệ thống cho chuyển ví"})
		return
	}
	transferID := uuid.NewString()
	source := &entity.Transaction{ID: uuid.NewString(), OwnerID: owner, WalletID: sourceWallet.ID, CategoryID: &sourceCategoryID, Type: entity.TransactionTypeExpense, Amount: input.Amount, OccurredAt: occurredAt, Note: normalizeOptional(input.Note), IncludedInReports: false, TransferID: &transferID}
	destination := &entity.Transaction{ID: uuid.NewString(), OwnerID: owner, WalletID: destinationWallet.ID, CategoryID: &destinationCategoryID, Type: entity.TransactionTypeIncome, Amount: input.Amount, OccurredAt: occurredAt, Note: normalizeOptional(input.Note), IncludedInReports: false, TransferID: &transferID}
	if err := h.Transfers.CreateTransfer(owner, source, destination); err != nil {
		if errors.Is(err, transactionrepo.ErrTransferWalletInvalid) || errors.Is(err, transactionrepo.ErrTransferInvalid) {
			transactionBadRequest(c, "Chuyển ví không hợp lệ.")
			return
		}
		Fail(c, http.StatusInternalServerError, Problem{Code: problemCodeTransactionSaveFailed, Title: problemTitleInternalServer, Detail: transactionSaveMessage})
		return
	}
	Created(c, []*entity.Transaction{source, destination})
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
	input, occurredAt, valid := h.bindAndValidate(c, owner, nil)
	if !valid {
		return
	}
	included := true
	if input.IncludedInReports != nil {
		included = *input.IncludedInReports
	}
	transaction := &entity.Transaction{ID: uuid.NewString(), OwnerID: owner, WalletID: input.WalletID, CategoryID: input.CategoryID, JarID: input.JarID, Type: input.Type, Amount: input.Amount, OccurredAt: occurredAt, Note: normalizeOptional(input.Note), IncludedInReports: included}
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
	existing, err := h.Transactions.Find(owner, c.Param("id"))
	if err != nil {
		Fail(c, http.StatusNotFound, Problem{Code: problemCodeTransactionNotFound, Title: problemTitleNotFound, Detail: transactionNotFoundMessage})
		return
	}
	input, occurredAt, valid := h.bindAndValidate(c, owner, existing)
	if !valid {
		return
	}
	updates := map[string]any{"wallet_id": input.WalletID, "category_id": input.CategoryID, "jar_id": input.JarID, "type": input.Type, "amount": input.Amount, "occurred_at": occurredAt, "note": normalizeOptional(input.Note)}
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

func (h *TransactionHandler) bindAndValidate(c *gin.Context, owner string, existing *entity.Transaction) (transactionInput, time.Time, bool) {
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
	wallet, err := h.Wallets.Find(owner, input.WalletID)
	if err != nil {
		Fail(c, http.StatusNotFound, Problem{Code: problemCodeTransactionWalletNotFound, Title: problemTitleNotFound, Detail: transactionWalletMessage})
		return transactionInput{}, time.Time{}, false
	}
	if wallet.Type == entity.WalletTypeCredit {
		transactionBadRequest(c, transactionCreditWalletMessage)
		return transactionInput{}, time.Time{}, false
	}
	input.CategoryID = normalizeOptional(input.CategoryID)
	if input.CategoryID != nil && !h.categoryValid(owner, *input.CategoryID, input.Type, input.WalletID, wallet.Type) {
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
	input.JarID = normalizeOptional(input.JarID)
	if input.JarID != nil {
		if _, err := uuid.Parse(*input.JarID); err != nil || input.Type != entity.TransactionTypeExpense || h.Users == nil || h.Jars == nil {
			transactionBadRequest(c, "Chỉ có thể gắn một hũ đang dùng cho khoản chi thường.")
			return transactionInput{}, time.Time{}, false
		}
		if input.CategoryID != nil {
			category, categoryErr := h.Categories.FindVisible(owner, *input.CategoryID)
			if categoryErr != nil || (category.SystemKey != nil && *category.SystemKey == "expense_transfer_out") {
				transactionBadRequest(c, "Không thể gắn hũ cho giao dịch chuyển nội bộ.")
				return transactionInput{}, time.Time{}, false
			}
		}
		preserveExistingJar := existing != nil && existing.JarID != nil && *existing.JarID == *input.JarID && existing.OccurredAt.Equal(occurredAt)
		if preserveExistingJar {
			return input, occurredAt, true
		}
		user, userErr := h.Users.FindByID(owner)
		if userErr != nil {
			transactionBadRequest(c, "Không xác định được múi giờ tài khoản.")
			return transactionInput{}, time.Time{}, false
		}
		location, zoneErr := time.LoadLocation(user.Timezone)
		if zoneErr != nil || user.Timezone == "Local" {
			transactionBadRequest(c, "Không xác định được múi giờ tài khoản.")
			return transactionInput{}, time.Time{}, false
		}
		month := occurredAt.In(location).Format("2006-01")
		if _, jarErr := h.Jars.FindMonthConfig(owner, *input.JarID, month, false); jarErr != nil {
			transactionBadRequest(c, "Hũ không có trong cấu hình tháng của giao dịch.")
			return transactionInput{}, time.Time{}, false
		}
	}
	return input, occurredAt, true
}

func (h *TransactionHandler) categoryValid(owner, categoryID, kind, walletID, walletType string) bool {
	categories, err := h.Categories.ListVisible(owner)
	if err != nil {
		return false
	}
	for _, category := range categories {
		if category.ID != categoryID || category.Kind != kind {
			continue
		}
		if walletType == entity.WalletTypeGoal {
			if category.SystemKey == nil {
				return false
			}
			key := *category.SystemKey
			allowed := kind == entity.TransactionTypeIncome && (key == "income_transfer_in" || key == "income_interest") || kind == entity.TransactionTypeExpense && key == "expense_transfer_out"
			if !allowed {
				return false
			}
		}
		if len(category.WalletIDs) == 0 {
			return true
		}
		for _, applicableWalletID := range category.WalletIDs {
			if applicableWalletID == walletID {
				return true
			}
		}
		return false
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
