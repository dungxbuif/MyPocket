// Package push contains the provider boundary for server push delivery.
// The transport is injected so retry/revocation semantics stay deterministic
// in tests and Personal Team builds can keep using the inbox with push off.
package push

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

const maxAttempts = 5

var (
	ErrInvalidConfig = errors.New("invalid APNs configuration")
	ErrTransport     = errors.New("APNs transport failure")
)

// Config is server-side only. PrivateKey must be loaded from a secret source
// and must never be included in logs, API responses or persisted jobs.
type Config struct {
	Enabled     bool
	TeamID      string
	KeyID       string
	Topic       string
	PrivateKey  string
	Environment string
}

func (c Config) Validate() error {
	if !c.Enabled {
		return nil
	}
	if strings.TrimSpace(c.TeamID) == "" || strings.TrimSpace(c.KeyID) == "" || strings.TrimSpace(c.Topic) == "" || strings.TrimSpace(c.PrivateKey) == "" {
		return ErrInvalidConfig
	}
	switch strings.ToLower(strings.TrimSpace(c.Environment)) {
	case "sandbox", "production":
		return nil
	default:
		return ErrInvalidConfig
	}
}

// Request contains provider delivery data. DeviceToken is never returned in
// a DeliveryResult and must be redacted before structured logging.
type Request struct {
	NotificationID string
	TokenID        string
	DeviceToken    string
	CollapseID     string
	Payload        []byte
}

type Response struct {
	StatusCode int
	RetryAfter time.Duration
}

type Transport interface {
	Send(context.Context, Request) (Response, error)
}

type Attempt struct{ Number int }

type State string

const (
	StateSent        State = "sent"
	StateRetry       State = "retry"
	StateRevokeToken State = "revoke_token"
	StateFailed      State = "failed"
)

type DeliveryResult struct {
	State       State
	RetryAt     *time.Time
	RevokeToken bool
}

// Deliver classifies one provider response. It does not sleep or send again;
// a durable queue owns the next-attempt schedule and can safely resume after a
// worker restart.
func Deliver(ctx context.Context, transport Transport, request Request, attempt Attempt, now time.Time) (DeliveryResult, error) {
	if transport == nil || strings.TrimSpace(request.NotificationID) == "" || strings.TrimSpace(request.TokenID) == "" || strings.TrimSpace(request.DeviceToken) == "" {
		return DeliveryResult{State: StateFailed}, ErrTransport
	}
	if attempt.Number < 1 {
		attempt.Number = 1
	}
	response, err := transport.Send(ctx, request)
	if err != nil {
		return retryResult(attempt.Number, now, 0), fmt.Errorf("%w: %v", ErrTransport, err)
	}
	switch {
	case response.StatusCode >= 200 && response.StatusCode < 300:
		return DeliveryResult{State: StateSent}, nil
	case response.StatusCode == 410:
		// Revoke only this registration; a newer token for the same device stays.
		return DeliveryResult{State: StateRevokeToken, RevokeToken: true}, nil
	case response.StatusCode == 429 || response.StatusCode >= 500:
		return retryResult(attempt.Number, now, response.RetryAfter), nil
	default:
		// 400/403 include malformed payload and invalid credentials/topic. They
		// are operator-visible permanent failures, not retryable user jobs.
		return DeliveryResult{State: StateFailed}, nil
	}
}

func retryResult(attempt int, now time.Time, retryAfter time.Duration) DeliveryResult {
	if attempt >= maxAttempts {
		return DeliveryResult{State: StateFailed}
	}
	if retryAfter <= 0 {
		retryAfter = time.Second << (attempt - 1)
		if retryAfter > 5*time.Minute {
			retryAfter = 5 * time.Minute
		}
	}
	next := now.Add(retryAfter)
	return DeliveryResult{State: StateRetry, RetryAt: &next}
}
