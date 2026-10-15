package repository

import (
	"encoding/json"
	"errors"

	"github.com/mypocket/backend/internal/entity"
)

const TransactionMutationCompleted = "completed"

var (
	ErrMutationConflict = errors.New("idempotency key payload conflict")
	ErrMutationPending  = errors.New("idempotency mutation is pending")
	ErrMutationInvalid  = errors.New("invalid idempotent create mutation")
)

// TransactionCreateMutationRepository is the durable create boundary for the
// offline client. Callers hash the canonical request, excluding generated IDs.
// Both the financial write and replay envelope commit or roll back together.
type TransactionCreateMutationRepository interface {
	ReplayCreateReceipt(ownerID, operation, idempotencyKey, requestHash string) (json.RawMessage, bool, error)
	CreateIdempotent(transaction *entity.Transaction, idempotencyKey, requestHash string) (*entity.Transaction, bool, error)
	CreateTransferIdempotent(ownerID string, source, destination *entity.Transaction, idempotencyKey, requestHash string) ([]entity.Transaction, bool, error)
}

// TransferMutationRepository adds durable replay protection to transfer
// mutations without widening the ordinary transaction repository contract.
type TransferMutationRepository interface {
	UpdateTransferIdempotent(ownerID, transferID string, updates TransferUpdate, idempotencyKey, requestHash string) ([]entity.Transaction, bool, error)
	DeleteTransferIdempotent(ownerID, transferID, idempotencyKey, requestHash string) (bool, error)
}
