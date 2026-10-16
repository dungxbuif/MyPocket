package httpapi

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
	"github.com/mypocket/backend/internal/usecase"
)

type SavingsReportHandler struct {
	Wallets      repository.WalletRepository
	Transactions repository.TransactionRepository
	Users        repository.UserRepository
	Now          func() time.Time
}

func (h *SavingsReportHandler) service() *usecase.SavingsReports {
	return &usecase.SavingsReports{Wallets: h.Wallets, Transactions: h.Transactions, Users: h.Users, Now: h.Now}
}

func savingsQuery(c *gin.Context) (month, wallet string, err error) {
	for _, key := range []string{"month", "wallet_id", "section", "kind", "limit", "offset"} {
		if len(c.QueryArray(key)) > 1 {
			return "", "", entity.ErrSavingsInvalidQuery
		}
	}
	month, wallet = c.Query("month"), c.Query("wallet_id")
	if wallet != "" {
		if _, parseErr := uuid.Parse(wallet); parseErr != nil {
			return "", "", entity.ErrSavingsInvalidQuery
		}
	}
	return month, wallet, nil
}

// GetSavingsReport godoc
// @Summary Report monthly goal-wallet activity and separate account-wide surplus
// @Tags Reports
// @Produce json
// @Security BearerAuth
// @Param month query string false "Account-local YYYY-MM; defaults to current month"
// @Param wallet_id query string false "Owned goal wallet; only filters goals"
// @Success 200 {object} entity.SavingsReport
// @Failure 400 {object} Problem
// @Failure 401 {object} Problem
// @Failure 404 {object} Problem
// @Failure 500 {object} Problem
// @Router /api/v1/reports/savings [get]
func (h *SavingsReportHandler) GetSavingsReport(c *gin.Context) {
	owner, ok := walletOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	month, wallet, err := savingsQuery(c)
	if err != nil {
		savingsFailure(c, err)
		return
	}
	report, err := h.service().Report(owner, month, wallet)
	if err != nil {
		savingsFailure(c, err)
		return
	}
	OK(c, report)
}

// ListSavingsTransactions godoc
// @Summary Page transactions using the same savings report scope and predicates
// @Tags Reports
// @Produce json
// @Security BearerAuth
// @Param month query string false "Account-local YYYY-MM"
// @Param wallet_id query string false "Owned goal wallet (goals only)"
// @Param section query string true "goals or surplus"
// @Param kind query string false "all, deposit, withdrawal, adjustment, transfer, income, expense"
// @Param limit query int false "1..100; default 50"
// @Param offset query int false "Nonnegative; default 0"
// @Success 200 {object} entity.SavingsTransactionPage
// @Failure 400 {object} Problem
// @Failure 401 {object} Problem
// @Failure 404 {object} Problem
// @Failure 500 {object} Problem
// @Router /api/v1/reports/savings/transactions [get]
func (h *SavingsReportHandler) ListSavingsTransactions(c *gin.Context) {
	owner, ok := walletOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	month, wallet, err := savingsQuery(c)
	if err != nil {
		savingsFailure(c, err)
		return
	}
	limit, offset := 50, 0
	if raw, present := c.GetQuery("limit"); present {
		limit, err = strconv.Atoi(raw)
		if err != nil {
			savingsFailure(c, entity.ErrSavingsInvalidQuery)
			return
		}
	}
	if raw, present := c.GetQuery("offset"); present {
		offset, err = strconv.Atoi(raw)
		if err != nil {
			savingsFailure(c, entity.ErrSavingsInvalidQuery)
			return
		}
	}
	page, err := h.service().TransactionsPage(owner, month, wallet, c.Query("section"), c.Query("kind"), limit, offset)
	if err != nil {
		savingsFailure(c, err)
		return
	}
	OK(c, page)
}

func savingsFailure(c *gin.Context, err error) {
	status, code, message := http.StatusInternalServerError, "SAVINGS_FAILED", "Không thể tải báo cáo tiết kiệm. Vui lòng thử lại."
	switch {
	case errors.Is(err, entity.ErrSavingsInvalidQuery):
		status, code, message = 400, "SAVINGS_INVALID_QUERY", "Tháng, phạm vi hoặc phân trang không hợp lệ."
	case errors.Is(err, entity.ErrSavingsFutureMonth):
		status, code, message = 400, "SAVINGS_FUTURE_MONTH", "Chưa có báo cáo tiết kiệm cho tháng tương lai."
	case errors.Is(err, entity.ErrSavingsWalletNotFound):
		status, code, message = 404, "SAVINGS_WALLET_NOT_FOUND", "Không tìm thấy ví mục tiêu."
	case errors.Is(err, entity.ErrSavingsOverflow):
		code, message = "SAVINGS_OVERFLOW", "Số tiền trong báo cáo vượt giới hạn tính toán."
	}
	Fail(c, status, Problem{Code: code, Title: http.StatusText(status), Detail: message})
}
