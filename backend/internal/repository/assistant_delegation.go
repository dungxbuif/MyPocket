package repository

import (
	"context"
	"time"

	"github.com/mypocket/backend/internal/entity"
)

type AssistantDelegationRepository interface {
	Create(context.Context, *entity.AssistantDelegation) error
	FindByLookup(context.Context, string) (*entity.AssistantDelegation, error)
	FindByID(context.Context, string) (*entity.AssistantDelegation, error)
	ListByOwner(context.Context, string) ([]entity.AssistantDelegation, error)
	Revoke(context.Context, string, string, time.Time) error
	TouchLastUsed(context.Context, string, time.Time) error
}
