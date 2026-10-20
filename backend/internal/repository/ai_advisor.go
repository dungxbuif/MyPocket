package repository

import (
	"context"
	"errors"
	"time"

	"github.com/mypocket/backend/internal/entity"
)

var (
	ErrAdvisorRequestConflict = errors.New("advisor request conflict")
	ErrAdvisorBusy            = errors.New("advisor conversation busy")
	ErrAdvisorPurged          = errors.New("advisor history purged")
	ErrAdvisorLeaseLost       = errors.New("advisor lease lost")
	ErrAdvisorRuntimeExpired  = errors.New("advisor runtime expired")
	ErrAdvisorEventsExpired   = errors.New("advisor events expired")
)

type AdvisorRunInput struct {
	OwnerID        string
	CredentialKind string
	CredentialID   string
	ExpiresAt      *time.Time
	RequestID      string
	PayloadHash    string
	Parts          []entity.AdvisorPart
}

type AdvisorRepository interface {
	StartRun(context.Context, AdvisorRunInput) (*entity.AdvisorRun, *entity.AdvisorMessage, bool, error)
	MarkRunStatus(context.Context, string, string, string, time.Time) error
	GetRun(context.Context, string, string) (*entity.AdvisorRun, error)
	GetRunByRequest(context.Context, string, string) (*entity.AdvisorRun, error)
	ListMessages(context.Context, string, string, int64, int) ([]entity.AdvisorMessage, error)
	AppendEvent(context.Context, string, string, string, any) (*entity.AdvisorEvent, error)
	CreateAssistantMessage(context.Context, string, string, int64, []entity.AdvisorPart) (*entity.AdvisorMessage, error)
}

// AdvisorRuntimeRepository fences executor writes and makes answer
// persistence atomic. Implementations must use database time and owner/run/
// generation/token matching; lease tokens are never client-visible.
type AdvisorRuntimeRepository interface {
	ClaimRun(context.Context, string, string, string) (entity.AdvisorRunLease, error)
	HeartbeatRun(context.Context, entity.AdvisorRunLease) error
	CompleteRun(context.Context, entity.AdvisorRunLease, entity.AdvisorCompletion) (entity.AdvisorCompletionResult, error)
	FinishRun(context.Context, entity.AdvisorRunLease, string, string, map[string]any) error
	ReapExpiredRuns(context.Context, int) (int, error)
}

// AdvisorContextRepository reads only the current owner's current generation.
// Reference validation remains the finance service's responsibility.
type AdvisorContextRepository interface {
	LoadAdvisorContextState(context.Context, string, string, int64) (entity.AdvisorContextState, error)
}

// AdvisorRunStatusRepository is an optional low-cost status probe used by a
// worker that may be executing in a different API process than the caller of
// the cancel endpoint. Implementations must scope the lookup by ownerID and
// return ErrNotFound for foreign or missing runs.
type AdvisorRunStatusRepository interface {
	GetRunStatus(context.Context, string, string) (string, error)
}

// AdvisorEventRepository is optional so adapters that only support polling can
// keep the core contract while PostgreSQL-backed deployments expose replayable
// SSE events.
type AdvisorEventRepository interface {
	ListEvents(context.Context, string, string, int64, int) ([]entity.AdvisorEvent, error)
}

// AdvisorExpiryRepository is implemented by the PostgreSQL store used by the
// worker. Keeping it optional preserves the existing AdvisorRepository contract
// for deterministic service tests and non-worker adapters.
type AdvisorExpiryRepository interface {
	ExpireIdleBatch(context.Context, time.Time, int) (int, error)
}

// AdvisorHistoryMetadataRepository is optional so memory/test adapters can
// keep returning the legacy history envelope while the PostgreSQL adapter
// exposes server-authoritative lifecycle metadata.
type AdvisorHistoryMetadataRepository interface {
	AdvisorHistoryMeta(context.Context, string, string) (entity.AdvisorHistoryMeta, error)
}

// AdvisorLifecycleRepository contains explicit reset operations separately
// from the core advisor contract so deterministic provider/unit adapters do not
// need to implement database lifecycle storage.
type AdvisorLifecycleRepository interface {
	ResetConversation(context.Context, string, string, string) (*entity.AdvisorResetResult, error)
}
