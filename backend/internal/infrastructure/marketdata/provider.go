package marketdata

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	ProviderHTTPTimeout = 10 * time.Second
	ProviderMaxBodySize = 1 << 20
)

var (
	ErrProviderNotConfigured = errors.New("market provider is not configured")
	ErrProviderBodyTooLarge  = errors.New("market provider response is too large")
	ErrProviderBadResponse   = errors.New("market provider returned an invalid response")
	ErrProviderPartial       = errors.New("market provider returned partial data")
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

type GoldProviderSettings struct {
	BTMCEndpoint string
	BTMHEndpoint string
}

const (
	DefaultBTMCEndpoint = "https://btmc.vn/"
	DefaultBTMHEndpoint = "https://baotinmanhhai.vn/bang-gia-vang"
)

// DefaultGoldProviders returns backend-owned provider adapters. The endpoint
// values can be overridden by deployment configuration, but are never accepted
// from a user request. Adding a new source means adding another definition to
// this registry rather than changing the poller or catalog identity rules.
func DefaultGoldProviders() []MarketPriceProvider {
	return NewGoldProviders(GoldProviderSettings{
		BTMCEndpoint: defaultString("GOLD_BTMC_ENDPOINT", DefaultBTMCEndpoint),
		BTMHEndpoint: defaultString("GOLD_BTMH_ENDPOINT", DefaultBTMHEndpoint),
	})
}

func NewGoldProviders(settings GoldProviderSettings) []MarketPriceProvider {
	return []MarketPriceProvider{
		NewGoldSourceAProvider(settings.BTMCEndpoint),
		NewGoldSourceBProvider(settings.BTMHEndpoint),
	}
}

func defaultString(key, fallback string) string {
	if value := strings.TrimSpace(getenv(key)); value != "" {
		return value
	}
	return fallback
}

func getenv(key string) string {
	// Kept in this small package boundary so provider configuration stays
	// deployment-owned without coupling marketdata to the application config
	// package. Secrets must be supplied through the endpoint by the deployer and
	// are never logged by the poller.
	return strings.TrimSpace(os.Getenv(key))
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
	Headers     map[string]string
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
	for name, value := range p.Headers {
		request.Header.Set(name, value)
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
		return nil, fmt.Errorf("%w: decode response: %w", ErrProviderBadResponse, err)
	}
	return quotes, nil
}
