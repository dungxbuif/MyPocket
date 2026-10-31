package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/infrastructure/marketdata"
	marketrepo "github.com/mypocket/backend/internal/repository"
)

type fakeMarketRepository struct {
	products  []entity.MarketProduct
	claims    int
	completed int
	failed    int
	quotes    []entity.MarketQuote
}

func (f *fakeMarketRepository) ListEnabledMarketProducts() ([]entity.MarketProduct, error) {
	return f.products, nil
}

func (f *fakeMarketRepository) FindMarketProduct(productID string) (*entity.MarketProduct, error) {
	return nil, marketrepo.ErrMarketProductNotFound
}

func (f *fakeMarketRepository) ClaimMarketPoll(providerKey string, slotAt, now time.Time, lease time.Duration) (marketrepo.MarketPollClaim, error) {
	f.claims++
	return marketrepo.MarketPollClaim{Acquired: true, Claim: entity.MarketPollClaim{
		ProviderKey: providerKey, SlotAt: slotAt, LeaseToken: "test-token", LeaseUntil: now.Add(lease), Status: "running",
	}}, nil
}

func (f *fakeMarketRepository) CompleteMarketPoll(claim entity.MarketPollClaim, quotes []entity.MarketQuote, now time.Time) (int, error) {
	f.completed++
	f.quotes = append(f.quotes, quotes...)
	return len(quotes), nil
}

func (f *fakeMarketRepository) FailMarketPoll(claim entity.MarketPollClaim, now time.Time, cause error) error {
	f.failed++
	return nil
}

type fakeMarketProvider struct {
	key    string
	quotes []marketdata.ProviderQuote
	called int
}

func (f *fakeMarketProvider) Key() string      { return f.key }
func (f *fakeMarketProvider) Configured() bool { return true }
func (f *fakeMarketProvider) Status() string   { return "configured" }
func (f *fakeMarketProvider) Fetch(context.Context) ([]marketdata.ProviderQuote, error) {
	f.called++
	return f.quotes, nil
}

func TestMarketPriceServicePollsConfiguredProvidersEveryTenMinuteSlot(t *testing.T) {
	now := time.Date(2026, 11, 1, 9, 17, 31, 0, time.UTC)
	repository := &fakeMarketRepository{products: []entity.MarketProduct{
		{ID: "p-a", ProviderKey: "btmc", ExternalProductID: "a", Unit: "chỉ", Currency: "VND", Enabled: true},
		{ID: "p-b", ProviderKey: "btmh", ExternalProductID: "b", Unit: "chỉ", Currency: "VND", Enabled: true},
	}}
	providerA := &fakeMarketProvider{key: "btmc", quotes: []marketdata.ProviderQuote{{ExternalProductID: "a", Price: 14180000, Unit: "chỉ", Currency: "VND", ObservedAt: now.Add(-time.Minute)}}}
	providerB := &fakeMarketProvider{key: "btmh", quotes: []marketdata.ProviderQuote{{ExternalProductID: "b", Price: 13280000, Unit: "chỉ", Currency: "VND", ObservedAt: now.Add(-time.Minute)}}}

	result, err := NewMarketPriceService(repository, providerA, providerB).PollDue(context.Background(), now)
	if err != nil {
		t.Fatalf("PollDue() error = %v", err)
	}
	wantSlot := time.Date(2026, 11, 1, 9, 10, 0, 0, time.UTC)
	if !result.SlotAt.Equal(wantSlot) {
		t.Fatalf("slot = %s, want %s", result.SlotAt, wantSlot)
	}
	if MarketPollInterval != 10*time.Minute {
		t.Fatalf("MarketPollInterval = %s, want 10m", MarketPollInterval)
	}
	if providerA.called != 1 || providerB.called != 1 || repository.completed != 2 || len(repository.quotes) != 2 {
		t.Fatalf("poll state providers=(%d,%d) completed=%d quotes=%d", providerA.called, providerB.called, repository.completed, len(repository.quotes))
	}
}

func TestMarketPriceServiceRejectsWrongUnitAndKeepsProviderFailureSafe(t *testing.T) {
	now := time.Date(2026, 11, 1, 9, 10, 0, 0, time.UTC)
	repository := &fakeMarketRepository{products: []entity.MarketProduct{{ID: "p-a", ProviderKey: "btmc", ExternalProductID: "a", Unit: "chỉ", Currency: "VND", Enabled: true}}}
	provider := &fakeMarketProvider{key: "btmc", quotes: []marketdata.ProviderQuote{{ExternalProductID: "a", Price: 14180000, Unit: "lượng", Currency: "VND", ObservedAt: now}}}
	result, err := NewMarketPriceService(repository, provider).PollDue(context.Background(), now)
	if err != nil {
		t.Fatalf("PollDue() error = %v", err)
	}
	if result.QuotesAccepted != 0 || result.QuotesRejected != 1 || len(repository.quotes) != 0 {
		t.Fatalf("wrong-unit result = %#v quotes=%d", result, len(repository.quotes))
	}

	failing := &marketProviderError{key: "btmc"}
	repository = &fakeMarketRepository{products: []entity.MarketProduct{{ID: "p-a", ProviderKey: "btmc", ExternalProductID: "a", Unit: "chỉ", Currency: "VND", Enabled: true}}}
	result, err = NewMarketPriceService(repository, failing).PollDue(context.Background(), now)
	if !errors.Is(err, errProviderTest) || result.Failed != 1 || repository.failed != 1 {
		t.Fatalf("failure result=%#v err=%v failed=%d", result, err, repository.failed)
	}
}

func TestNormalizeProviderQuotesRejectsFutureDuplicateAndUnknownProducts(t *testing.T) {
	now := time.Date(2026, 11, 1, 9, 10, 0, 0, time.UTC)
	products := []entity.MarketProduct{{ID: "p-a", ProviderKey: "btmc", ExternalProductID: "a", Unit: "chỉ", Currency: "VND"}}
	quotes, rejected := normalizeProviderQuotes("btmc", products, []marketdata.ProviderQuote{
		{ExternalProductID: "a", Price: 100, Unit: "chỉ", Currency: "VND", ObservedAt: now},
		{ExternalProductID: "a", Price: 101, Unit: "chỉ", Currency: "VND", ObservedAt: now.Add(time.Minute)},
		{ExternalProductID: "missing", Price: 100, Unit: "chỉ", Currency: "VND", ObservedAt: now},
		{ExternalProductID: "a", Price: 100, Unit: "chỉ", Currency: "VND", ObservedAt: now.Add(MarketClockSkew + time.Second)},
	}, now)
	if len(quotes) != 1 || rejected != 3 {
		t.Fatalf("normalized quotes=%+v rejected=%d", quotes, rejected)
	}
	if !quotes[0].ObservedAt.Equal(now) || !quotes[0].FetchedAt.Equal(now) {
		t.Fatalf("quote timestamps=%+v", quotes[0])
	}
}

func TestMarketPriceServiceSkipsUnconfiguredProviderWithoutClaim(t *testing.T) {
	now := time.Date(2026, 11, 1, 9, 10, 0, 0, time.UTC)
	repository := &fakeMarketRepository{products: []entity.MarketProduct{{ID: "p-a", ProviderKey: "btmc", ExternalProductID: "a", Unit: "chỉ", Currency: "VND", Enabled: true}}}
	result, err := NewMarketPriceService(repository, unconfiguredMarketProvider{key: "btmc"}).PollDue(context.Background(), now)
	if err != nil {
		t.Fatalf("PollDue() error = %v", err)
	}
	if result.Skipped != 1 || result.Claimed != 0 || repository.claims != 0 || result.ProviderStatus["btmc"] != "unconfigured" {
		t.Fatalf("unconfigured result=%+v claims=%d", result, repository.claims)
	}
}

type unconfiguredMarketProvider struct{ key string }

func (p unconfiguredMarketProvider) Key() string      { return p.key }
func (p unconfiguredMarketProvider) Configured() bool { return false }
func (p unconfiguredMarketProvider) Status() string   { return "unconfigured" }
func (p unconfiguredMarketProvider) Fetch(context.Context) ([]marketdata.ProviderQuote, error) {
	return nil, marketdata.ErrProviderNotConfigured
}

var errProviderTest = errors.New("provider test failure")

type marketProviderError struct{ key string }

func (f *marketProviderError) Key() string      { return f.key }
func (f *marketProviderError) Configured() bool { return true }
func (f *marketProviderError) Status() string   { return "configured" }
func (f *marketProviderError) Fetch(context.Context) ([]marketdata.ProviderQuote, error) {
	return nil, errProviderTest
}
