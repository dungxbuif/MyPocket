package marketdata

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	ProviderHTTPTimeout = 10 * time.Second
	ProviderMaxBodySize = 1 << 20
)

var (
	ErrProviderNotConfigured   = errors.New("market provider is not configured")
	ErrProviderBodyTooLarge    = errors.New("market provider response is too large")
	ErrProviderBadResponse     = errors.New("market provider returned an invalid response")
	ErrWaitingExternalContract = errors.New("waiting_external_contract")
)

// ProviderQuote is the normalized provider boundary. Adapters must map their
// source-specific response into this shape without converting units or using
// fetch time in place of the source timestamp.
type ProviderQuote struct {
	ExternalProductID string
	Price             int64
	Unit              string
	Currency          string
	ObservedAt        time.Time
}

type MarketPriceProvider interface {
	Key() string
	Configured() bool
	Status() string
	Fetch(context.Context) ([]ProviderQuote, error)
}

// DefaultGoldProviders is deliberately disabled until each real source's
// endpoint, response contract, authentication and product mapping is supplied.
func DefaultGoldProviders() []MarketPriceProvider {
	return []MarketPriceProvider{NewGoldSourceAProvider(), NewGoldSourceBProvider()}
}

func ConfiguredProviderKeys(providers []MarketPriceProvider) map[string]struct{} {
	keys := make(map[string]struct{}, len(providers))
	for _, provider := range providers {
		if provider != nil && provider.Configured() && strings.TrimSpace(provider.Key()) != "" {
			keys[provider.Key()] = struct{}{}
		}
	}
	return keys
}

// HTTPJSONProvider supplies the bounded transport shared by real adapters.
// Endpoint and decoder are backend-owned configuration; they are never read
// from an authenticated user request.
type HTTPJSONProvider struct {
	ProviderKey string
	Endpoint    string
	Client      *http.Client
	Decode      func([]byte) ([]ProviderQuote, error)
}

func NewHTTPJSONProvider(providerKey, endpoint string, client *http.Client, decode func([]byte) ([]ProviderQuote, error)) (*HTTPJSONProvider, error) {
	providerKey = strings.TrimSpace(providerKey)
	endpoint = strings.TrimSpace(endpoint)
	if providerKey == "" || endpoint == "" || decode == nil {
		return nil, ErrProviderNotConfigured
	}
	parsed, err := url.Parse(endpoint)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return nil, fmt.Errorf("%w: invalid backend endpoint", ErrProviderBadResponse)
	}
	if client == nil {
		client = &http.Client{Timeout: ProviderHTTPTimeout}
	}
	return &HTTPJSONProvider{ProviderKey: providerKey, Endpoint: endpoint, Client: client, Decode: decode}, nil
}

func (p *HTTPJSONProvider) Key() string { return p.ProviderKey }

func (p *HTTPJSONProvider) Configured() bool {
	return p != nil && strings.TrimSpace(p.ProviderKey) != "" && strings.TrimSpace(p.Endpoint) != "" && p.Decode != nil && p.Client != nil
}

func (p *HTTPJSONProvider) Status() string {
	if !p.Configured() {
		return "unconfigured"
	}
	return "configured"
}

func (p *HTTPJSONProvider) Fetch(ctx context.Context) ([]ProviderQuote, error) {
	if !p.Configured() {
		return nil, ErrProviderNotConfigured
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, p.Endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: request construction", ErrProviderBadResponse)
	}
	response, err := p.Client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("%w: status %d", ErrProviderBadResponse, response.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, ProviderMaxBodySize+1))
	if err != nil {
		return nil, fmt.Errorf("%w: read response", ErrProviderBadResponse)
	}
	if len(body) > ProviderMaxBodySize {
		return nil, ErrProviderBodyTooLarge
	}
	quotes, err := p.Decode(body)
	if err != nil {
		return nil, fmt.Errorf("%w: decode response", ErrProviderBadResponse)
	}
	return quotes, nil
}
