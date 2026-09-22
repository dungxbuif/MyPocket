package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
)

var ErrFinanceQueryInvalid = errors.New("finance query is invalid")

type FinanceQueryService struct {
	Reader repository.FinanceReader
}

func NewFinanceQueryService(reader repository.FinanceReader) *FinanceQueryService {
	return &FinanceQueryService{Reader: reader}
}

func (s *FinanceQueryService) Execute(ctx context.Context, ownerID string, query entity.NormalizedQuery) (entity.FinanceResult, error) {
	bundle, err := s.RefreshBundle(ctx, ownerID, []entity.NormalizedQuery{query})
	if err != nil {
		return entity.FinanceResult{}, err
	}
	if len(bundle.Results) != 1 {
		return entity.FinanceResult{}, ErrFinanceQueryInvalid
	}
	return bundle.Results[0], nil
}

func (s *FinanceQueryService) RefreshBundle(ctx context.Context, ownerID string, queries []entity.NormalizedQuery) (entity.FactBundle, error) {
	if s == nil || s.Reader == nil || strings.TrimSpace(ownerID) == "" || len(queries) == 0 {
		return entity.FactBundle{}, ErrFinanceQueryInvalid
	}
	for _, query := range queries {
		if strings.TrimSpace(query.Key) == "" || strings.TrimSpace(query.Kind) == "" {
			return entity.FactBundle{}, ErrFinanceQueryInvalid
		}
	}
	return s.Reader.ReadBundle(ctx, strings.TrimSpace(ownerID), queries)
}
