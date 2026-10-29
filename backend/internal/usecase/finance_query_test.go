package usecase

import (
	"context"
	"testing"

	"github.com/mypocket/backend/internal/entity"
)

type financeReaderStub struct {
	owner   string
	queries []entity.NormalizedQuery
	bundle  entity.FactBundle
}

func (s *financeReaderStub) ReadBundle(_ context.Context, owner string, queries []entity.NormalizedQuery) (entity.FactBundle, error) {
	s.owner = owner
	s.queries = append([]entity.NormalizedQuery(nil), queries...)
	return s.bundle, nil
}

func TestFinanceQueryServiceExecuteDelegatesOwnerAndOneQuery(t *testing.T) {
	reader := &financeReaderStub{bundle: entity.FactBundle{ID: "bundle-1", Results: []entity.FinanceResult{{QueryKey: "q1"}}}}
	service := NewFinanceQueryService(reader)
	query := entity.NormalizedQuery{Key: "q1", Kind: "get_finance_summary"}
	result, err := service.Execute(context.Background(), "owner-1", query)
	if err != nil {
		t.Fatal(err)
	}
	if reader.owner != "owner-1" || len(reader.queries) != 1 || result.QueryKey != "q1" {
		t.Fatalf("unexpected delegation: owner=%q queries=%+v result=%+v", reader.owner, reader.queries, result)
	}
}

func TestFinanceQueryServiceRejectsEmptyOwnerOrQueries(t *testing.T) {
	service := NewFinanceQueryService(&financeReaderStub{})
	if _, err := service.Execute(context.Background(), "", entity.NormalizedQuery{}); err == nil {
		t.Fatal("empty owner must be rejected")
	}
	if _, err := service.RefreshBundle(context.Background(), "owner", nil); err == nil {
		t.Fatal("empty query list must be rejected")
	}
}
