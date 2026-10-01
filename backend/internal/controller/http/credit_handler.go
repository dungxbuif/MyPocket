package httpapi

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mypocket/backend/internal/entity"
	categoryrepo "github.com/mypocket/backend/internal/repository"
	transactionrepo "github.com/mypocket/backend/internal/repository"
)

type CreditHandler struct {
	Credits    transactionrepo.CreditRepository
	Wallets    categoryrepo.WalletRepository
	Categories categoryrepo.CategoryRepository
}

type creditEntryInput struct {
	Kind       string  `json:"kind"`
	CategoryID *string `json:"category_id"`
	Amount     int64   `json:"amount"`
	OccurredAt string  `json:"occurred_at"`
	Note       *string `json:"note"`
}

type creditPaymentInput struct {
	SourceWalletID string  `json:"source_wallet_id"`
	Amount         int64   `json:"amount"`
	OccurredAt     string  `json:"occurred_at"`
	Note           *string `json:"note"`
}

func (r *Router) RegisterCreditRoutes(h *CreditHandler) {
	g := r.Engine.Group("/api/v1/credit")
	g.Use(r.AuthMiddleware.RequireAuth)
	g.POST("/wallets/:wallet_id/entries", h.CreateEntry)
	g.POST("/wallets/:wallet_id/payments", h.CreatePayment)
	g.GET("/wallets/:wallet_id/statement", h.Statement)
}

// CreateCreditEntry godoc
// @Summary Record a purchase, refund, fee, or interest entry on a credit wallet
// @Tags Credit
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param wallet_id path string true "Credit wallet ID"
// @Param entry body creditEntryInput true "Credit entry"
// @Success 201 {object} entity.Transaction
// @Failure 400 {object} Problem
// @Failure 401 {object} Problem
// @Failure 404 {object} Problem
// @Router /api/v1/credit/wallets/{wallet_id}/entries [post]
func (h *CreditHandler) CreateEntry(c *gin.Context) {
	owner, ok := walletOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	if h.Credits == nil {
		creditUnavailable(c)
		return
	}
	var input creditEntryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		transactionBadRequest(c, problemDetailInvalidJSON)
		return
	}
	input.Kind = strings.TrimSpace(input.Kind)
	input.CategoryID = normalizeOptional(input.CategoryID)
	if !validCreditEntryKind(input.Kind) || input.Amount <= 0 {
		transactionBadRequest(c, "Loại và số tiền giao dịch tín dụng không hợp lệ.")
		return
	}
	wallet, err := h.Wallets.Find(owner, strings.TrimSpace(c.Param("wallet_id")))
	if err != nil || wallet.Type != entity.WalletTypeCredit {
		creditWalletNotFound(c)
		return
	}
	occurredAt, valid := parseCreditDate(input.OccurredAt)
	if !valid {
		transactionBadRequest(c, transactionDateMessage)
		return
	}
	entry, err := h.Credits.CreateCreditEntry(owner, transactionrepo.CreditEntryInput{WalletID: wallet.ID, Kind: input.Kind, CategoryID: input.CategoryID, Amount: input.Amount, OccurredAt: occurredAt, Note: normalizeOptional(input.Note)})
	if err != nil {
		creditMapError(c, err)
		return
	}
	Created(c, entry)
}

// CreateCreditPayment godoc
// @Summary Record a payment from another wallet to a credit wallet
// @Tags Credit
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param wallet_id path string true "Credit wallet ID"
// @Param payment body creditPaymentInput true "Credit payment"
// @Success 201 {array} entity.Transaction
// @Failure 400 {object} Problem
// @Failure 401 {object} Problem
// @Failure 404 {object} Problem
// @Router /api/v1/credit/wallets/{wallet_id}/payments [post]
func (h *CreditHandler) CreatePayment(c *gin.Context) {
	owner, ok := walletOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	if h.Credits == nil {
		creditUnavailable(c)
		return
	}
	var input creditPaymentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		transactionBadRequest(c, problemDetailInvalidJSON)
		return
	}
	input.SourceWalletID = strings.TrimSpace(input.SourceWalletID)
	if input.SourceWalletID == "" || input.Amount <= 0 {
		transactionBadRequest(c, "Ví nguồn và số tiền thanh toán là bắt buộc.")
		return
	}
	wallet, err := h.Wallets.Find(owner, strings.TrimSpace(c.Param("wallet_id")))
	if err != nil || wallet.Type != entity.WalletTypeCredit {
		creditWalletNotFound(c)
		return
	}
	occurredAt, valid := parseCreditDate(input.OccurredAt)
	if !valid {
		transactionBadRequest(c, transactionDateMessage)
		return
	}
	rows, err := h.Credits.CreateCreditPayment(owner, wallet.ID, input.SourceWalletID, input.Amount, occurredAt, normalizeOptional(input.Note))
	if err != nil {
		creditMapError(c, err)
		return
	}
	Created(c, rows)
}

// CreditStatement godoc
// @Summary Read a credit wallet statement and available credit
// @Tags Credit
// @Produce json
// @Security BearerAuth
// @Param wallet_id path string true "Credit wallet ID"
// @Param from query string false "Inclusive date YYYY-MM-DD"
// @Param to query string false "Exclusive date YYYY-MM-DD"
// @Success 200 {object} entity.CreditStatement
// @Failure 400 {object} Problem
// @Failure 401 {object} Problem
// @Failure 404 {object} Problem
// @Router /api/v1/credit/wallets/{wallet_id}/statement [get]
func (h *CreditHandler) Statement(c *gin.Context) {
	owner, ok := walletOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	if h.Credits == nil {
		creditUnavailable(c)
		return
	}
	from, to, valid := parseCreditRange(c.Query("from"), c.Query("to"))
	if !valid {
		transactionBadRequest(c, "Khoảng ngày phải có định dạng YYYY-MM-DD hợp lệ.")
		return
	}
	statement, err := h.Credits.ListCreditStatement(owner, strings.TrimSpace(c.Param("wallet_id")), from, to)
	if err != nil {
		creditMapError(c, err)
		return
	}
	OK(c, statement)
}

func validCreditEntryKind(kind string) bool {
	return kind == entity.CreditKindPurchase || kind == entity.CreditKindRefund || kind == entity.CreditKindFee || kind == entity.CreditKindInterest
}

func parseCreditDate(raw string) (time.Time, bool) {
	if strings.TrimSpace(raw) == "" {
		return time.Now().UTC(), true
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	return parsed.UTC(), err == nil
}

func parseCreditRange(fromRaw, toRaw string) (*time.Time, *time.Time, bool) {
	var from, to *time.Time
	if strings.TrimSpace(fromRaw) != "" {
		parsed, err := time.Parse("2006-01-02", fromRaw)
		if err != nil {
			return nil, nil, false
		}
		utc := parsed.UTC()
		from = &utc
	}
	if strings.TrimSpace(toRaw) != "" {
		parsed, err := time.Parse("2006-01-02", toRaw)
		if err != nil {
			return nil, nil, false
		}
		utc := parsed.UTC().AddDate(0, 0, 1)
		to = &utc
	}
	if from != nil && to != nil && !from.Before(*to) {
		return nil, nil, false
	}
	return from, to, true
}

func creditMapError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, transactionrepo.ErrCreditWalletInvalid):
		creditWalletNotFound(c)
	case errors.Is(err, transactionrepo.ErrCreditInvalid), errors.Is(err, transactionrepo.ErrCreditPaymentInvalid):
		transactionBadRequest(c, "Thao tác tín dụng không hợp lệ.")
	default:
		Fail(c, http.StatusInternalServerError, Problem{Code: problemCodeTransactionSaveFailed, Title: problemTitleInternalServer, Detail: "Không xử lý được sổ tín dụng."})
	}
}

func creditWalletNotFound(c *gin.Context) {
	Fail(c, http.StatusNotFound, Problem{Code: problemCodeTransactionWalletNotFound, Title: problemTitleNotFound, Detail: transactionWalletMessage})
}

func creditUnavailable(c *gin.Context) {
	Fail(c, http.StatusNotImplemented, Problem{Code: problemCodeTransactionSaveFailed, Title: problemTitleInternalServer, Detail: "Luồng tín dụng chưa được cấu hình."})
}
