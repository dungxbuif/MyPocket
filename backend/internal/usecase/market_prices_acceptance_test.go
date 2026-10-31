package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/infrastructure/marketdata"
)

func TestMarketPollIsSharedAcrossOwnersForOneProviderSlot(t *testing.T) {
	now := time.Date(2026, 11, 1, 9, 10, 0, 0, time.UTC)
	repository := &fakeMarketRepository{products: []entity.MarketProduct{
		{ID: "p-owner-a", ProviderKey: "btmc", ExternalProductID: "a", Unit: "chỉ", Currency: "VND", Enabled: true},
		{ID: "p-owner-b", ProviderKey: "btmc", ExternalProductID: "b", Unit: "chỉ", Currency: "VND", Enabled: true},
	}}
	provider := &fakeMarketProvider{key: "btmc", quotes: []marketdata.ProviderQuote{
		{ExternalProductID: "a", Price: 14_180_000, Unit: "chỉ", Currency: "VND", ObservedAt: now},
		{ExternalProductID: "b", Price: 14_280_000, Unit: "chỉ", Currency: "VND", ObservedAt: now},
	}}

	result, err := NewMarketPriceService(repository, provider).PollDue(context.Background(), now)
	if err != nil {
		t.Fatal(err)
	}
	if provider.called != 1 || repository.claims != 1 || repository.completed != 1 || len(repository.quotes) != 2 {
		t.Fatalf("provider=%d claims=%d completed=%d quotes=%d; expected one shared provider poll for both holdings", provider.called, repository.claims, repository.completed, len(repository.quotes))
	}
	if result.Providers != 1 || result.Claimed != 1 || result.QuotesAccepted != 2 {
		t.Fatalf("poll result=%+v", result)
	}
}

func TestMarketPollRejectsWrongCurrencyAndNonPositiveQuote(t *testing.T) {
	now := time.Date(2026, 11, 1, 9, 10, 0, 0, time.UTC)
	products := []entity.MarketProduct{{ID: "p-a", ProviderKey: "btmc", ExternalProductID: "a", Unit: "chỉ", Currency: "VND"}}
	quotes, rejected := normalizeProviderQuotes("btmc", products, []marketdata.ProviderQuote{
		{ExternalProductID: "a", Price: 100, Unit: "chỉ", Currency: "USD", ObservedAt: now},
		{ExternalProductID: "a", Price: 0, Unit: "chỉ", Currency: "VND", ObservedAt: now},
		{ExternalProductID: "a", Price: -1, Unit: "chỉ", Currency: "VND", ObservedAt: now},
		{ExternalProductID: "a", Price: 101, Unit: "chỉ", Currency: "VND", ObservedAt: now},
	}, now)
	if len(quotes) != 1 || rejected != 3 {
		t.Fatalf("quotes=%+v rejected=%d, want one accepted VND quote and three rejections", quotes, rejected)
	}
}

func TestMarketPollProviderFailureDoesNotPublishOrEraseLastQuote(t *testing.T) {
	now := time.Date(2026, 11, 1, 9, 10, 0, 0, time.UTC)
	last := entity.MarketQuote{ProductID: "p-a", Price: 14_000_000, ObservedAt: now.Add(-time.Hour), FetchedAt: now.Add(-time.Hour)}
	repository := &fakeMarketRepository{
		products: []entity.MarketProduct{{ID: "p-a", ProviderKey: "btmc", ExternalProductID: "a", Unit: "chỉ", Currency: "VND", Enabled: true}},
		quotes:   []entity.MarketQuote{last},
	}
	provider := &marketProviderError{key: "btmc"}
	result, err := NewMarketPriceService(repository, provider).PollDue(context.Background(), now)
	if !errors.Is(err, errProviderTest) || result.Failed != 1 || repository.failed != 1 {
		t.Fatalf("result=%+v err=%v failed=%d", result, err, repository.failed)
	}
	if len(repository.quotes) != 1 || repository.quotes[0] != last || repository.completed != 0 {
		t.Fatalf("provider failure changed quote state: quotes=%+v completed=%d", repository.quotes, repository.completed)
	}
}
