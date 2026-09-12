package worker_test

import (
	"context"
	"testing"
	"time"

	"mypocket/internal/portfolio"
	"mypocket/internal/worker"
)

func TestPortfolioPriceRunnerRefreshesConfiguredAutomaticAssets(t *testing.T) {
	now := time.Date(2026, 8, 31, 10, 0, 0, 0, time.UTC)
	repo := &portfolioPriceRepoStub{
		candidates: []portfolio.PriceRefreshCandidate{{
			UserID:         "user_1",
			AssetID:        "asset_1",
			ProviderKey:    "static",
			ProviderSymbol: "SJC",
			BaseVersion:    3,
		}},
	}
	provider := &portfolio.StaticPriceProvider{Quotes: map[string]portfolio.ProviderQuote{
		"static:SJC": {
			ProviderKey:     "static",
			ProviderSymbol:  "SJC",
			UnitPriceVND:    75000000,
			PricedAt:        now,
			ProviderQuoteID: "static:SJC:2026-08-31T10:00:00Z",
		},
	}}
	runner := worker.PortfolioPriceRunner{Repo: repo, Provider: provider, Now: func() time.Time { return now }}

	refreshed, err := runner.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("run portfolio price refresh: %v", err)
	}
	if refreshed != 1 || len(repo.quotes) != 1 {
		t.Fatalf("expected one refreshed quote, refreshed=%d stored=%d", refreshed, len(repo.quotes))
	}
	if repo.quotes[0].UnitPriceVND != 75000000 || repo.quotes[0].ProviderQuoteID == "" {
		t.Fatalf("unexpected quote: %#v", repo.quotes[0])
	}
}

type portfolioPriceRepoStub struct {
	candidates []portfolio.PriceRefreshCandidate
	quotes     []portfolio.ProviderQuote
}

func (r *portfolioPriceRepoStub) AcquireWorkerLease(context.Context, string, string, int64) (bool, error) {
	return true, nil
}

func (r *portfolioPriceRepoStub) ListPriceRefreshCandidates(context.Context, int) ([]portfolio.PriceRefreshCandidate, error) {
	return r.candidates, nil
}

func (r *portfolioPriceRepoStub) AddProviderPrice(_ context.Context, _ portfolio.PriceRefreshCandidate, quote portfolio.ProviderQuote) (portfolio.Position, error) {
	r.quotes = append(r.quotes, quote)
	return portfolio.Position{}, nil
}
