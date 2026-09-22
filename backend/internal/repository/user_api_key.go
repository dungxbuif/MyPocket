package repository

import (
	"context"
	"errors"
	"time"

	"github.com/mypocket/backend/internal/entity"
)

var (
	ErrAPIKeyNotFound = ErrNotFound
	ErrAPIKeyRevoked  = errors.New("api key revoked or invalid")
)

type UserAPIKeyRepository interface {
	Create(context.Context, *entity.UserAPIKey) error
	FindByLookup(context.Context, string) (*entity.UserAPIKey, error)
	FindByID(context.Context, string) (*entity.UserAPIKey, error)
	ListByOwner(context.Context, string) ([]entity.UserAPIKey, error)
	Revoke(context.Context, string, string, time.Time) error
	TouchLastUsed(context.Context, string, time.Time) error
}
