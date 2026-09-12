package worker

import (
	"context"
	"fmt"
	"os"
	"time"

	"mypocket/internal/portfolio"
)

type PortfolioPriceRepository interface {
	AcquireWorkerLease(ctx context.Context, leaseKey string, owner string, ttlSeconds int64) (bool, error)
	ListPriceRefreshCandidates(ctx context.Context, limit int) ([]portfolio.PriceRefreshCandidate, error)
	AddProviderPrice(ctx context.Context, candidate portfolio.PriceRefreshCandidate, quote portfolio.ProviderQuote) (portfolio.Position, error)
}

type PortfolioPriceRunner struct {
	Repo     PortfolioPriceRepository
	Provider portfolio.PriceProvider
	Owner    string
	Now      func() time.Time
	Limit    int
}

func (r PortfolioPriceRunner) RunOnce(ctx context.Context) (int, error) {
	if r.Repo == nil {
		return 0, fmt.Errorf("portfolio price repository is required")
	}
	if r.Provider == nil {
		return 0, nil
	}
	owner := r.Owner
	if owner == "" {
		owner = defaultOwner()
	}
	now := time.Now
	if r.Now != nil {
		now = r.Now
	}
	acquired, err := r.Repo.AcquireWorkerLease(ctx, "portfolio-price-refresh", owner, int64(time.Minute.Seconds()))
	if err != nil {
		return 0, err
	}
	if !acquired {
		return 0, nil
	}
	limit := r.Limit
	if limit <= 0 {
		limit = 100
	}
	candidates, err := r.Repo.ListPriceRefreshCandidates(ctx, limit)
	if err != nil {
		return 0, err
	}
	updated := 0
	for _, candidate := range candidates {
		quote, ok, err := r.Provider.Quote(ctx, candidate.ProviderKey, candidate.ProviderSymbol, now().UTC())
		if err != nil {
			return updated, err
		}
		if !ok {
			continue
		}
		if _, err := r.Repo.AddProviderPrice(ctx, candidate, quote); err != nil {
			return updated, err
		}
		updated++
	}
	return updated, nil
}

func NewPortfolioPriceRunnerFromEnv(repo PortfolioPriceRepository) (PortfolioPriceRunner, error) {
	provider, err := portfolio.NewStaticPriceProvider(os.Getenv("PORTFOLIO_STATIC_PRICES_JSON"))
	if err != nil {
		return PortfolioPriceRunner{}, err
	}
	return PortfolioPriceRunner{Repo: repo, Provider: provider}, nil
}
