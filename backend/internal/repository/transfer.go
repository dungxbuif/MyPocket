package repository

import "github.com/mypocket/backend/internal/entity"

// TransferRepository persists both sides of an internal wallet transfer as one
// atomic ledger operation. The two rows remain ordinary income/expense rows so
// existing balance and report calculations stay correct.
type TransferRepository interface {
	CreateTransfer(ownerID string, source, destination *entity.Transaction) error
}
