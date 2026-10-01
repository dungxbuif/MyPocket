package httpapi

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	walletrepo "github.com/mypocket/backend/internal/repository"
)

const (
	walletUnauthorizedMessage = "chưa đăng nhập"
	walletNameRequiredMessage = "tên ví không được để trống"
	walletInvalidTypeMessage  = "loại ví không hợp lệ"
	walletTargetAmountMessage = "mục tiêu tiết kiệm phải lớn hơn 0"
	walletCreditLimitMessage  = "hạn mức tín dụng phải lớn hơn 0"
	walletNotFoundMessage     = "không tìm thấy ví"
	walletLoadErrorMessage    = "không đọc được danh sách ví"
	walletSaveErrorMessage    = "không lưu được ví"
)

type WalletHandler struct{ Wallets walletrepo.WalletRepository }

type walletInput struct {
	Name                 string  `json:"name"`
	Type                 string  `json:"type"`
	OpeningBalance       int64   `json:"opening_balance"`
	IsInTotal            *bool   `json:"is_in_total"`
	Description          *string `json:"description"`
	TargetAmount         *int64  `json:"target_amount"`
	CreditLimit          *int64  `json:"credit_limit"`
	LastStatementBalance *int64  `json:"last_statement_balance"`
	StatementDay         *int    `json:"statement_day"`
	PaymentDueDay        *int    `json:"payment_due_day"`
	TargetDate           *string `json:"target_date"`
}

func NewWalletHandler(wallets walletrepo.WalletRepository) *WalletHandler {
	return &WalletHandler{Wallets: wallets}
}

// ListWallets godoc
// @Summary List wallets
// @Tags Wallets
// @Produce json
// @Security BearerAuth
// @Success 200 {array} entity.Wallet
// @Failure 401 {object} Problem
// @Router /api/v1/wallets [get]
func (h *WalletHandler) ListWallets(c *gin.Context) {
	owner, ok := walletOwner(c)
	if !ok {
		Fail(c, http.StatusUnauthorized, Problem{Code: problemCodeAuthRequired, Title: problemTitleUnauthorized, Detail: walletUnauthorizedMessage})
		return
	}
	wallets, err := h.Wallets.List(owner)
	if err != nil {
		Fail(c, http.StatusInternalServerError, Problem{Code: problemCodeWalletLoadFailed, Title: problemTitleInternalServer, Detail: walletLoadErrorMessage})
		return
	}
	OK(c, wallets)
}

// CreateWallet godoc
// @Summary Create a wallet
// @Tags Wallets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param wallet body walletInput true "Wallet input"
// @Success 201 {object} entity.Wallet
// @Failure 400 {object} Problem
// @Failure 401 {object} Problem
// @Router /api/v1/wallets [post]
func (h *WalletHandler) CreateWallet(c *gin.Context) {
	owner, ok := walletOwner(c)
	if !ok {
		Fail(c, http.StatusUnauthorized, Problem{Code: problemCodeAuthRequired, Title: problemTitleUnauthorized, Detail: walletUnauthorizedMessage})
		return
	}
	input, valid := bindWalletInput(c)
	if !valid {
		return
	}
	isInTotal := true
	if input.IsInTotal != nil {
		isInTotal = *input.IsInTotal
	}
	wallet := &entity.Wallet{ID: uuid.NewString(), OwnerID: owner, Name: input.Name, Type: input.Type, Currency: entity.WalletCurrencyVND, OpeningBalance: input.OpeningBalance, IsInTotal: isInTotal, Description: input.Description, TargetAmount: input.TargetAmount, CreditLimit: input.CreditLimit, LastStatementBalance: input.LastStatementBalance, StatementDay: input.StatementDay, PaymentDueDay: input.PaymentDueDay}
	if input.TargetDate != nil && *input.TargetDate != "" {
		parsed, _ := entity.ParseCalendarDate(*input.TargetDate)
		wallet.TargetDate = &parsed
	}
	if err := h.Wallets.Create(wallet); err != nil {
		Fail(c, http.StatusInternalServerError, Problem{Code: problemCodeWalletSaveFailed, Title: problemTitleInternalServer, Detail: walletSaveErrorMessage})
		return
	}
	Created(c, wallet)
}

// UpdateWallet godoc
// @Summary Update a wallet
// @Tags Wallets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Wallet ID"
// @Param wallet body walletInput true "Wallet input"
// @Success 200 {object} entity.Wallet
// @Failure 400 {object} Problem
// @Failure 401 {object} Problem
// @Failure 404 {object} Problem
// @Router /api/v1/wallets/{id} [patch]
func (h *WalletHandler) UpdateWallet(c *gin.Context) {
	owner, ok := walletOwner(c)
	if !ok {
		Fail(c, http.StatusUnauthorized, Problem{Code: problemCodeAuthRequired, Title: problemTitleUnauthorized, Detail: walletUnauthorizedMessage})
		return
	}
	existing, err := h.Wallets.Find(owner, c.Param("id"))
	if err != nil {
		Fail(c, http.StatusNotFound, Problem{Code: problemCodeWalletNotFound, Title: problemTitleNotFound, Detail: walletNotFoundMessage})
		return
	}
	input, valid := bindWalletInput(c, existing.Type)
	if !valid {
		return
	}
	if existing.Type == entity.WalletTypeCredit {
		if input.LastStatementBalance == nil {
			input.LastStatementBalance = existing.LastStatementBalance
		}
		if input.StatementDay == nil {
			input.StatementDay = existing.StatementDay
		}
		if input.PaymentDueDay == nil {
			input.PaymentDueDay = existing.PaymentDueDay
		}
	}
	updates := map[string]any{"name": input.Name, "description": input.Description, "target_amount": input.TargetAmount, "credit_limit": input.CreditLimit, "last_statement_balance": input.LastStatementBalance, "statement_day": input.StatementDay, "payment_due_day": input.PaymentDueDay}
	if input.TargetDate != nil {
		updates["target_date"] = nil
		if *input.TargetDate != "" {
			parsed, _ := entity.ParseCalendarDate(*input.TargetDate)
			updates["target_date"] = parsed
		}
	}
	if input.IsInTotal != nil {
		updates["is_in_total"] = *input.IsInTotal
	}
	wallet, err := h.Wallets.Update(owner, c.Param("id"), updates)
	if err != nil {
		Fail(c, http.StatusNotFound, Problem{Code: problemCodeWalletNotFound, Title: problemTitleNotFound, Detail: walletNotFoundMessage})
		return
	}
	OK(c, wallet)
}

// DeleteWallet godoc
// @Summary Permanently delete a wallet
// @Tags Wallets
// @Security BearerAuth
// @Param id path string true "Wallet ID"
// @Success 204
// @Failure 401 {object} Problem
// @Failure 404 {object} Problem
// @Router /api/v1/wallets/{id} [delete]
func (h *WalletHandler) DeleteWallet(c *gin.Context) {
	owner, ok := walletOwner(c)
	if !ok {
		Fail(c, http.StatusUnauthorized, Problem{Code: problemCodeAuthRequired, Title: problemTitleUnauthorized, Detail: walletUnauthorizedMessage})
		return
	}
	if err := h.Wallets.Delete(owner, c.Param("id")); err != nil {
		Fail(c, http.StatusNotFound, Problem{Code: problemCodeWalletNotFound, Title: problemTitleNotFound, Detail: walletNotFoundMessage})
		return
	}
	NoContent(c)
}

func walletOwner(c *gin.Context) (string, bool) {
	value, ok := c.Get(contextUserIDKey)
	ownerID, valid := value.(string)
	return ownerID, ok && valid && strings.TrimSpace(ownerID) != ""
}

func bindWalletInput(c *gin.Context, existingType ...string) (walletInput, bool) {
	var input walletInput
	if err := c.ShouldBindJSON(&input); err != nil {
		Fail(c, http.StatusBadRequest, Problem{Code: problemCodeBadRequest, Title: problemTitleBadRequest, Detail: problemDetailInvalidJSON})
		return walletInput{}, false
	}
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		Fail(c, http.StatusBadRequest, Problem{Code: problemCodeBadRequest, Title: problemTitleBadRequest, Detail: walletNameRequiredMessage})
		return walletInput{}, false
	}
	if len(existingType) > 0 {
		input.Type = existingType[0]
	} else if input.Type == "" {
		input.Type = entity.WalletTypeBasic
	}
	if input.Type != entity.WalletTypeBasic && input.Type != entity.WalletTypeGoal && input.Type != entity.WalletTypeCredit {
		Fail(c, http.StatusBadRequest, Problem{Code: problemCodeBadRequest, Title: problemTitleBadRequest, Detail: walletInvalidTypeMessage})
		return walletInput{}, false
	}
	switch input.Type {
	case entity.WalletTypeBasic:
		input.TargetAmount, input.CreditLimit, input.LastStatementBalance, input.StatementDay, input.PaymentDueDay = nil, nil, nil, nil, nil
	case entity.WalletTypeGoal:
		input.CreditLimit, input.LastStatementBalance, input.StatementDay, input.PaymentDueDay = nil, nil, nil, nil
		if input.TargetAmount == nil || *input.TargetAmount <= 0 {
			Fail(c, http.StatusBadRequest, Problem{Code: problemCodeBadRequest, Title: problemTitleBadRequest, Detail: walletTargetAmountMessage})
			return walletInput{}, false
		}
	case entity.WalletTypeCredit:
		input.TargetAmount = nil
		if input.CreditLimit == nil || *input.CreditLimit <= 0 {
			Fail(c, http.StatusBadRequest, Problem{Code: problemCodeBadRequest, Title: problemTitleBadRequest, Detail: walletCreditLimitMessage})
			return walletInput{}, false
		}
		if input.LastStatementBalance != nil && *input.LastStatementBalance < 0 {
			Fail(c, http.StatusBadRequest, Problem{Code: problemCodeBadRequest, Title: problemTitleBadRequest, Detail: "dư nợ sao kê không được âm"})
			return walletInput{}, false
		}
		if input.StatementDay != nil && (*input.StatementDay < 1 || *input.StatementDay > 31) {
			Fail(c, http.StatusBadRequest, Problem{Code: problemCodeBadRequest, Title: problemTitleBadRequest, Detail: "ngày sao kê phải từ 1 đến 31"})
			return walletInput{}, false
		}
		if input.PaymentDueDay != nil && (*input.PaymentDueDay < 1 || *input.PaymentDueDay > 31) {
			Fail(c, http.StatusBadRequest, Problem{Code: problemCodeBadRequest, Title: problemTitleBadRequest, Detail: "ngày đến hạn phải từ 1 đến 31"})
			return walletInput{}, false
		}
	}
	if input.Type != entity.WalletTypeGoal {
		input.TargetDate = nil
	}
	if input.TargetDate != nil && *input.TargetDate != "" {
		if _, err := entity.ParseCalendarDate(*input.TargetDate); err != nil {
			Fail(c, http.StatusBadRequest, Problem{Code: problemCodeBadRequest, Title: problemTitleBadRequest, Detail: "ngày mục tiêu phải có định dạng YYYY-MM-DD hợp lệ"})
			return walletInput{}, false
		}
	}
	return input, true
}
