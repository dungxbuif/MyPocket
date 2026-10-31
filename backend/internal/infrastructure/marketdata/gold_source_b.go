package marketdata

import "context"

const GoldSourceBKey = "gold_source_b"

// GoldSourceBProvider is intentionally a disabled shell. It must not invent
// a URL or parse a guessed response before the external contract is recorded.
type GoldSourceBProvider struct{}

func NewGoldSourceBProvider() *GoldSourceBProvider { return &GoldSourceBProvider{} }

func (*GoldSourceBProvider) Key() string      { return GoldSourceBKey }
func (*GoldSourceBProvider) Configured() bool { return false }
func (*GoldSourceBProvider) Status() string   { return "waiting_external_contract" }
func (*GoldSourceBProvider) Fetch(context.Context) ([]ProviderQuote, error) {
	return nil, ErrWaitingExternalContract
}
