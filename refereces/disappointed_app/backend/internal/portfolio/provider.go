package portfolio

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type PriceProvider interface {
	Quote(ctx context.Context, providerKey string, providerSymbol string, now time.Time) (ProviderQuote, bool, error)
}

type StaticPriceProvider struct {
	Quotes map[string]ProviderQuote
}

func NewStaticPriceProvider(rawJSON string) (*StaticPriceProvider, error) {
	provider := &StaticPriceProvider{Quotes: map[string]ProviderQuote{}}
	rawJSON = strings.TrimSpace(rawJSON)
	if rawJSON == "" {
		return provider, nil
	}
	var values map[string]struct {
		UnitPriceVND    int64  `json:"unit_price_vnd"`
		PricedAt        string `json:"priced_at"`
		ProviderQuoteID string `json:"provider_quote_id"`
	}
	if err := json.Unmarshal([]byte(rawJSON), &values); err != nil {
		return nil, fmt.Errorf("parse portfolio static prices: %w", err)
	}
	for key, value := range values {
		parts := strings.SplitN(key, ":", 2)
		if len(parts) != 2 || value.UnitPriceVND <= 0 {
			return nil, ErrValidation
		}
		pricedAt := time.Now().UTC()
		if strings.TrimSpace(value.PricedAt) != "" {
			parsed, err := time.Parse(time.RFC3339, value.PricedAt)
			if err != nil {
				return nil, fmt.Errorf("parse portfolio static price time: %w", err)
			}
			pricedAt = parsed.UTC()
		}
		providerKey := strings.TrimSpace(parts[0])
		providerSymbol := strings.TrimSpace(parts[1])
		quoteID := strings.TrimSpace(value.ProviderQuoteID)
		if quoteID == "" {
			quoteID = fmt.Sprintf("%s:%s:%s:%d", providerKey, providerSymbol, pricedAt.Format(time.RFC3339), value.UnitPriceVND)
		}
		provider.Quotes[quoteKey(providerKey, providerSymbol)] = ProviderQuote{
			ProviderKey:     providerKey,
			ProviderSymbol:  providerSymbol,
			UnitPriceVND:    value.UnitPriceVND,
			PricedAt:        pricedAt,
			ProviderQuoteID: quoteID,
		}
	}
	return provider, nil
}

func (p *StaticPriceProvider) Quote(_ context.Context, providerKey string, providerSymbol string, now time.Time) (ProviderQuote, bool, error) {
	if p == nil {
		return ProviderQuote{}, false, nil
	}
	quote, ok := p.Quotes[quoteKey(providerKey, providerSymbol)]
	if !ok {
		return ProviderQuote{}, false, nil
	}
	if quote.PricedAt.IsZero() {
		quote.PricedAt = now.UTC()
	}
	return quote, true, nil
}

func quoteKey(providerKey string, providerSymbol string) string {
	return strings.ToLower(strings.TrimSpace(providerKey)) + ":" + strings.ToUpper(strings.TrimSpace(providerSymbol))
}
