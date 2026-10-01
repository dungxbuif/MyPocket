package repository

import (
	"errors"

	"github.com/mypocket/backend/internal/entity"
)

const TransactionMutationCompleted = "completed"

var (
	ErrMutationConflict = errors.New("idempotency key payload conflict")
	ErrMutationPending  = errors.New("idempotency mutation is pending")
)

// TransferMutationRepository adds durable replay protection to transfer
// mutations without widening the ordinary transaction repository contract.
type TransferMutationRepository interface {
	UpdateTransferIdempotent(ownerID, transferID string, updates TransferUpdate, idempotencyKey, requestHash string) ([]entity.Transaction, bool, error)
	DeleteTransferIdempotent(ownerID, transferID, idempotencyKey, requestHash string) (bool, error)
}
