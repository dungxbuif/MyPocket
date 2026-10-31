package marketdata

import "context"

const GoldSourceAKey = "gold_source_a"

// GoldSourceAProvider is intentionally a disabled shell. The owner has not
// supplied a real endpoint, schema, auth method, or stable product mapping.
type GoldSourceAProvider struct{}

func NewGoldSourceAProvider() *GoldSourceAProvider { return &GoldSourceAProvider{} }

func (*GoldSourceAProvider) Key() string      { return GoldSourceAKey }
func (*GoldSourceAProvider) Configured() bool { return false }
func (*GoldSourceAProvider) Status() string   { return "waiting_external_contract" }
func (*GoldSourceAProvider) Fetch(context.Context) ([]ProviderQuote, error) {
	return nil, ErrWaitingExternalContract
}
