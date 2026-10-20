package ai

import (
	"errors"
	"fmt"
	"strings"
)

// ErrInvalidProviderConfig is intentionally generic. Validation errors may be
// returned to startup diagnostics, so they must never contain an endpoint,
// model name, API key, or other provider credential material.
var ErrInvalidProviderConfig = errors.New("invalid AI provider configuration")

func validateProviderConfig(baseURL, apiKey, model string) error {
	if !validURL(strings.TrimSpace(baseURL)) {
		return fmt.Errorf("%w: endpoint", ErrInvalidProviderConfig)
	}
	if strings.TrimSpace(apiKey) == "" {
		return fmt.Errorf("%w: credential", ErrInvalidProviderConfig)
	}
	if strings.TrimSpace(model) == "" {
		return fmt.Errorf("%w: model", ErrInvalidProviderConfig)
	}
	return nil
}

// Validate checks the single server-side extraction provider. It is a local
// contract guard, not a provider discovery or account-level chooser. Provider
// secrets remain process-only and are never part of a request payload.
func (c Config) Validate() error {
	return validateProviderConfig(c.BaseURL, c.APIKey, c.Model)
}

// Validate checks the single server-side advisor provider. The advisor and
// extraction paths intentionally share the same fail-closed endpoint contract
// while retaining separate credentials/configuration.
func (c AdvisorClientConfig) Validate() error {
	return validateProviderConfig(c.BaseURL, c.APIKey, c.Model)
}
