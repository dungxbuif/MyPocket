package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mypocket/backend/internal/entity"
	transactionrepo "github.com/mypocket/backend/internal/repository"
)

type createReceiptContext struct {
	repository           transactionrepo.TransactionCreateMutationRepository
	operation, key, hash string
}

type CreateMutationMetadata struct {
	Operation      string `json:"operation"`
	IdempotencyKey string `json:"idempotency_key"`
	Status         string `json:"status"`
	Replayed       bool   `json:"replayed"`
}

type CreateMutationMeta struct {
	Mutation CreateMutationMetadata `json:"mutation"`
}
type CreateTransactionResponse struct {
	Data entity.Transaction  `json:"data"`
	Meta *CreateMutationMeta `json:"meta"`
}
type CreateTransferResponse struct {
	Data []entity.Transaction `json:"data"`
	Meta *CreateMutationMeta  `json:"meta"`
}

func (receipt *createReceiptContext) respond(c *gin.Context, data any, replayed bool) {
	status := http.StatusCreated
	if replayed {
		status = http.StatusOK
	}
	c.JSON(status, Response{Data: data, Meta: CreateMutationMeta{Mutation: CreateMutationMetadata{
		Operation: receipt.operation, IdempotencyKey: receipt.key, Status: "applied", Replayed: replayed,
	}}})
}

func (h *TransactionHandler) prepareCreateReceipt(c *gin.Context, owner, operation string, input any) (*createReceiptContext, bool) {
	keys := c.Request.Header.Values("Idempotency-Key")
	if len(keys) == 0 {
		return nil, false
	} // Transitional legacy online path.
	valid := len(keys) == 1
	key := ""
	if valid {
		key = keys[0]
		valid = strings.TrimSpace(key) != "" && len(key) <= 128
	}
	for _, character := range key {
		if character < 32 || character > 126 {
			valid = false
		}
	}
	if !valid {
		Fail(c, http.StatusBadRequest, Problem{Code: "INVALID_IDEMPOTENCY_KEY", Title: problemTitleBadRequest, Detail: "Idempotency-Key phải là một giá trị ASCII, dài từ 1 đến 128 byte."})
		return nil, true
	}
	mutator, ok := h.Transactions.(transactionrepo.TransactionCreateMutationRepository)
	if !ok {
		Fail(c, http.StatusServiceUnavailable, Problem{Code: "mutation_unavailable", Title: problemTitleInternalServer, Detail: "Chưa thể xử lý yêu cầu có khóa chống trùng."})
		return nil, true
	}
	receipt := &createReceiptContext{repository: mutator, operation: operation, key: key, hash: hashMutationInput(input)}
	result, replayed, err := mutator.ReplayCreateReceipt(owner, operation, key, receipt.hash)
	if err != nil {
		if !createReceiptFailure(c, err) {
			Fail(c, http.StatusInternalServerError, Problem{Code: problemCodeTransactionSaveFailed, Title: problemTitleInternalServer, Detail: transactionSaveMessage})
		}
		return nil, true
	}
	if replayed {
		receipt.respond(c, result, true)
		return nil, true
	}
	return receipt, false
}

func createReceiptFailure(c *gin.Context, err error) bool {
	switch {
	case errors.Is(err, transactionrepo.ErrMutationConflict):
		Fail(c, http.StatusConflict, Problem{Code: "IDEMPOTENCY_KEY_REUSED", Title: problemTitleBadRequest, Detail: "Idempotency-Key đã được dùng cho nội dung khác."})
	case errors.Is(err, transactionrepo.ErrMutationPending):
		Fail(c, http.StatusServiceUnavailable, Problem{Code: "mutation_in_progress", Title: problemTitleInternalServer, Detail: "Yêu cầu chưa có kết quả hoàn tất; hãy dùng lại cùng khóa để kiểm tra."})
	case errors.Is(err, transactionrepo.ErrMutationInvalid):
		Fail(c, http.StatusBadRequest, Problem{Code: "mutation_invalid", Title: problemTitleBadRequest, Detail: "Yêu cầu chống trùng không hợp lệ."})
	default:
		return false
	}
	return true
}
