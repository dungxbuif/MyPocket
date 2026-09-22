package repository

import (
	"context"

	"github.com/mypocket/backend/internal/entity"
)

type FinanceReader interface {
	ReadBundle(context.Context, string, []entity.NormalizedQuery) (entity.FactBundle, error)
}
