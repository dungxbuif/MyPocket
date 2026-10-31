package main

import (
	"context"
	"time"

	"github.com/mypocket/backend/internal/usecase"
)

// runMarketPricePass is the isolated worker seam for the shared quote poller.
// The common worker main intentionally does not wire it until the parent
// branch reconciles its lifecycle and provider configuration.
func runMarketPricePass(ctx context.Context, service *usecase.MarketPriceService, now time.Time) (usecase.PollResult, error) {
	if service == nil {
		return usecase.PollResult{}, nil
	}
	return service.PollDue(ctx, now)
}
