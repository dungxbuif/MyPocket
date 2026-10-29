package repository

import "github.com/mypocket/backend/internal/entity"

type WalletRepository interface {
	List(ownerID string) ([]entity.Wallet, error)
	Find(ownerID, id string) (*entity.Wallet, error)
	Create(wallet *entity.Wallet) error
	Update(ownerID, id string, updates map[string]any) (*entity.Wallet, error)
	Delete(ownerID, id string) error
}

// WalletCreateMutationRepository is an optional durable replay boundary for
// confirmed wallet writes. Legacy adapters can omit it and keep ordinary
// online behavior; PostgreSQL implements it using transaction_mutations.
type WalletCreateMutationRepository interface {
	CreateIdempotent(wallet *entity.Wallet, idempotencyKey, requestHash string) (*entity.Wallet, bool, error)
}
