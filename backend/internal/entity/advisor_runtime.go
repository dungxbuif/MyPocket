package entity

import (
	"encoding/json"
	"errors"
	"time"
)

// AdvisorRuntimeLease is the opaque executor lease used to fence provider
// work. Token values must never be sent to clients or written to logs.
type AdvisorRunLease struct {
	OwnerID    string
	RunID      string
	Generation int64
	Token      string
	DeadlineAt time.Time
	LeaseUntil time.Time
}

// AdvisorCompletion is the complete set of durable output written by one
// executor. The repository commits all fields atomically or none of them.
type AdvisorCompletion struct {
	Parts        []AdvisorPart
	Usage        map[string]any
	FactBundle   *AdvisorFactBundle
	ContextState map[string]any
}

// AdvisorCompletionResult is safe to return after an uncertain commit. When
// Replayed is true, the existing assistant message is returned and no new
// rows were inserted.
type AdvisorCompletionResult struct {
	Message  *AdvisorMessage
	Run      *AdvisorRun
	Replayed bool
}

var ErrAdvisorRuntimePayload = errors.New("advisor runtime payload is invalid")

type AdvisorRuntimePolicy struct {
	RunDeadline    time.Duration
	LeaseDuration  time.Duration
	HeartbeatEvery time.Duration
	ReaperEvery    time.Duration
	DBOperation    time.Duration
}

func DefaultAdvisorRuntimePolicy() AdvisorRuntimePolicy {
	return AdvisorRuntimePolicy{RunDeadline: 240 * time.Second, LeaseDuration: 45 * time.Second, HeartbeatEvery: 10 * time.Second, ReaperEvery: 15 * time.Second, DBOperation: 2 * time.Second}
}

const advisorContextStateMaxBytes = 4 * 1024

// ValidateAdvisorContextState enforces the bounded, JSON-safe context state
// contract. It intentionally does not inspect financial values; the state is
// only a set of references and filters, never a source of financial truth.
func ValidateAdvisorContextState(state map[string]any) error {
	if state == nil {
		return nil
	}
	encoded, err := json.Marshal(state)
	if err != nil || len(encoded) > advisorContextStateMaxBytes {
		return ErrAdvisorRuntimePayload
	}
	return nil
}
