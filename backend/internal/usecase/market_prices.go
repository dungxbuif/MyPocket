package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/infrastructure/marketdata"
	marketrepo "github.com/mypocket/backend/internal/repository"
)

const (
	MarketPollInterval = 30 * time.Minute
	MarketPollLease    = 2 * time.Minute
	MarketClockSkew    = 5 * time.Minute
)

type PollResult struct {
	SlotAt          time.Time
	Providers       int
	Claimed         int
	Skipped         int
	Failed          int
	QuotesAccepted  int
	QuotesRejected  int
	ProviderStatus  map[string]string
	ProviderLatency map[string]time.Duration
}

type MarketPriceService struct {
	Repository marketrepo.MarketRepository
	Providers  map[string]marketdata.MarketPriceProvider
}

func NewMarketPriceService(repository marketrepo.MarketRepository, providers ...marketdata.MarketPriceProvider) *MarketPriceService {
	registry := make(map[string]marketdata.MarketPriceProvider, len(providers))
	for _, provider := range providers {
		if provider == nil || strings.TrimSpace(provider.Key()) == "" {
			continue
		}
		registry[provider.Key()] = provider
	}
	return &MarketPriceService{Repository: repository, Providers: registry}
}

func (s *MarketPriceService) PollDue(ctx context.Context, now time.Time) (PollResult, error) {
	result := PollResult{ProviderStatus: map[string]string{}, ProviderLatency: map[string]time.Duration{}}
	if s == nil || s.Repository == nil || now.IsZero() {
		return result, marketrepo.ErrMarketProductInvalid
	}
	if ctx == nil {
		ctx = context.Background()
	}
	now = now.UTC()
	result.SlotAt = now.Truncate(MarketPollInterval)
	products, err := s.Repository.ListEnabledMarketProducts()
	if err != nil {
		return result, err
	}
	byProvider := make(map[string][]entity.MarketProduct)
	for _, product := range products {
		key := strings.TrimSpace(product.ProviderKey)
		if key == "" {
			continue
		}
		byProvider[key] = append(byProvider[key], product)
	}
	var firstErr error
	for providerKey, providerProducts := range byProvider {
		result.Providers++
		provider, configured := s.Providers[providerKey]
		if !configured || provider == nil || !provider.Configured() {
			result.Skipped++
			if provider == nil {
				result.ProviderStatus[providerKey] = "unconfigured"
			} else {
				result.ProviderStatus[providerKey] = provider.Status()
			}
			continue
		}
		result.ProviderStatus[providerKey] = provider.Status()
		claim, err := s.Repository.ClaimMarketPoll(providerKey, result.SlotAt, now, MarketPollLease)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		if !claim.Acquired {
			continue
		}
		result.Claimed++
		started := time.Now()
		fetchCtx, cancel := context.WithTimeout(ctx, marketdata.ProviderHTTPTimeout)
		rawQuotes, fetchErr := provider.Fetch(fetchCtx)
		cancel()
		result.ProviderLatency[providerKey] = time.Since(started)
		if fetchErr != nil {
			result.Failed++
			if err := s.Repository.FailMarketPoll(claim.Claim, now, fetchErr); err != nil && firstErr == nil {
				firstErr = err
			}
			if firstErr == nil {
				firstErr = fetchErr
			}
			continue
		}
		quotes, rejected := normalizeProviderQuotes(providerKey, providerProducts, rawQuotes, now)
		result.QuotesRejected += rejected
		accepted, completeErr := s.Repository.CompleteMarketPoll(claim.Claim, quotes, now)
		if completeErr != nil {
			if firstErr == nil {
				firstErr = completeErr
			}
			continue
		}
		result.QuotesAccepted += accepted
	}
	return result, firstErr
}

func normalizeProviderQuotes(providerKey string, products []entity.MarketProduct, raw []marketdata.ProviderQuote, now time.Time) ([]entity.MarketQuote, int) {
	byExternalID := make(map[string]entity.MarketProduct, len(products))
	for _, product := range products {
		byExternalID[product.ExternalProductID] = product
	}
	quotes := make([]entity.MarketQuote, 0, len(raw))
	seen := make(map[string]struct{}, len(raw))
	rejected := 0
	for _, candidate := range raw {
		product, ok := byExternalID[candidate.ExternalProductID]
		if !ok || product.ProviderKey != providerKey || candidate.Price <= 0 || strings.TrimSpace(candidate.Unit) != strings.TrimSpace(product.Unit) || strings.TrimSpace(candidate.Currency) != strings.TrimSpace(product.Currency) || candidate.ObservedAt.IsZero() || candidate.ObservedAt.After(now.Add(MarketClockSkew)) {
			rejected++
			continue
		}
		if _, duplicate := seen[product.ID]; duplicate {
			rejected++
			continue
		}
		seen[product.ID] = struct{}{}
		quotes = append(quotes, entity.MarketQuote{ProductID: product.ID, Price: candidate.Price, ObservedAt: candidate.ObservedAt.UTC(), FetchedAt: now})
	}
	return quotes, rejected
}
