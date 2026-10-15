package entity

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

const (
	AdvisorIdleTTL       = 30 * time.Minute
	AdvisorRoleUser      = "user"
	AdvisorRoleAssistant = "assistant"

	AdvisorStatusQueued      = "queued"
	AdvisorStatusRunning     = "running"
	AdvisorStatusCompleted   = "completed"
	AdvisorStatusFailed      = "failed"
	AdvisorStatusCancelled   = "cancelled"
	AdvisorStatusInterrupted = "interrupted"
	AdvisorStatusPurged      = "purged"
)

var (
	ErrAdvisorMessageInvalid = errors.New("advisor message is invalid")
	ErrAdvisorRunTransition  = errors.New("advisor run transition is invalid")
)

type AdvisorPart struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	Data any    `json:"data,omitempty"`
}

type AdvisorConversation struct {
	ID                string     `json:"id" gorm:"primaryKey"`
	OwnerID           string     `json:"owner_id" gorm:"not null;uniqueIndex"`
	Generation        int64      `json:"generation" gorm:"not null;default:1"`
	NextMessageSeq    int64      `json:"next_message_seq" gorm:"not null;default:1"`
	Summary           any        `json:"summary" gorm:"serializer:json;type:jsonb"`
	SummaryThroughSeq int64      `json:"summary_through_seq" gorm:"not null;default:0"`
	SummaryVersion    int64      `json:"summary_version" gorm:"not null;default:0"`
	LastUserMessageAt *time.Time `json:"last_user_message_at,omitempty" gorm:"index"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// AdvisorHistoryMeta is additive lifecycle metadata for history reads. The
// message array remains the response data contract; these fields let clients
// suppress stale local state without using a device clock as authority.
type AdvisorHistoryMeta struct {
	ConversationID    string     `json:"conversation_id"`
	Generation        int64      `json:"generation"`
	LastUserMessageAt *time.Time `json:"last_user_message_at,omitempty"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty"`
	ServerTime        time.Time  `json:"server_time"`
}

func (c AdvisorConversation) HistoryMeta(serverTime time.Time) AdvisorHistoryMeta {
	return AdvisorHistoryMeta{
		ConversationID:    c.ID,
		Generation:        c.Generation,
		LastUserMessageAt: c.LastUserMessageAt,
		ExpiresAt:         c.IdleExpiresAt(),
		ServerTime:        serverTime.UTC(),
	}
}

// AdvisorResetRequest is a content-free idempotency ledger for an explicit
// conversation reset. Keeping this separate from advisor runs means a reset
// never masquerades as a finance question or gets replayed through the model.
type AdvisorResetRequest struct {
	OwnerID         string    `json:"owner_id" gorm:"primaryKey"`
	ClientRequestID string    `json:"client_request_id" gorm:"primaryKey"`
	PayloadHash     string    `json:"-" gorm:"not null"`
	ConversationID  string    `json:"conversation_id" gorm:"not null"`
	Generation      int64     `json:"generation" gorm:"not null"`
	CreatedAt       time.Time `json:"created_at"`
}

func (AdvisorResetRequest) TableName() string { return "advisor_reset_requests" }

// AdvisorResetResult is the stable, content-free result returned to a client.
type AdvisorResetResult struct {
	ConversationID string    `json:"conversation_id"`
	Generation     int64     `json:"generation"`
	ServerTime     time.Time `json:"server_time"`
	Replayed       bool      `json:"replayed"`
	// PurgedRunIDs is an in-process coordination detail. It is deliberately
	// omitted from the wire response: clients only need the reset result while
	// the service uses these IDs to interrupt provider work handled locally.
	PurgedRunIDs []string `json:"-"`
}

func (AdvisorConversation) TableName() string { return "advisor_conversations" }

// IdleExpiresAt returns the server-authoritative expiry boundary for the
// conversation. Assistant responses, history reads and polling never move
// this timestamp; only an accepted user message does.
func (c AdvisorConversation) IdleExpiresAt() *time.Time {
	if c.LastUserMessageAt == nil {
		return nil
	}
	expires := c.LastUserMessageAt.UTC().Add(AdvisorIdleTTL)
	return &expires
}

// IsIdleExpired uses an inclusive boundary so exactly 30 minutes without an
// accepted user message starts a fresh conversation generation.
func (c AdvisorConversation) IsIdleExpired(now time.Time) bool {
	expires := c.IdleExpiresAt()
	return expires != nil && !now.UTC().Before(*expires)
}

type AdvisorMessage struct {
	ID             string        `json:"id" gorm:"primaryKey"`
	ConversationID string        `json:"conversation_id" gorm:"not null;index"`
	Generation     int64         `json:"generation" gorm:"not null"`
	Seq            int64         `json:"seq" gorm:"not null"`
	RunID          string        `json:"run_id" gorm:"not null;index"`
	Role           string        `json:"role" gorm:"not null"`
	PartsVersion   int           `json:"parts_version" gorm:"not null;default:1"`
	Parts          []AdvisorPart `json:"parts" gorm:"serializer:json;type:jsonb"`
	CreatedAt      time.Time     `json:"created_at"`
}

func (AdvisorMessage) TableName() string { return "advisor_messages" }

func (m AdvisorMessage) Validate() error {
	if m.Role != AdvisorRoleUser && m.Role != AdvisorRoleAssistant {
		return ErrAdvisorMessageInvalid
	}
	if len(m.Parts) == 0 || len(m.Parts) > 64 {
		return ErrAdvisorMessageInvalid
	}
	totalBytes := 0
	for _, part := range m.Parts {
		if strings.TrimSpace(part.Type) == "" || len(part.Text) > 8192 {
			return ErrAdvisorMessageInvalid
		}
		if part.Data != nil {
			encoded, err := json.Marshal(part.Data)
			if err != nil || len(encoded) > 64*1024 {
				return ErrAdvisorMessageInvalid
			}
			totalBytes += len(encoded)
		}
	}
	if totalBytes > 256*1024 {
		return ErrAdvisorMessageInvalid
	}
	return nil
}

type AdvisorRun struct {
	ID                  string         `json:"id" gorm:"primaryKey"`
	OwnerID             string         `json:"owner_id" gorm:"not null;index"`
	ConversationID      string         `json:"conversation_id" gorm:"not null;index"`
	Generation          int64          `json:"generation" gorm:"not null"`
	ClientRequestID     string         `json:"client_request_id" gorm:"not null"`
	PayloadHash         string         `json:"-" gorm:"not null"`
	CredentialKind      string         `json:"credential_kind" gorm:"not null"`
	CredentialID        string         `json:"-" gorm:"not null"`
	CredentialExpiresAt *time.Time     `json:"-"`
	Status              string         `json:"status" gorm:"not null"`
	LeaseToken          string         `json:"-"`
	LeaseUntil          *time.Time     `json:"-"`
	LastEventSeq        int64          `json:"last_event_seq" gorm:"not null;default:0"`
	ErrorCode           string         `json:"error_code,omitempty"`
	ModelUsage          map[string]any `json:"-" gorm:"serializer:json;type:jsonb"`
	CreatedAt           time.Time      `json:"created_at"`
	FinishedAt          *time.Time     `json:"finished_at,omitempty"`
}

func (AdvisorRun) TableName() string { return "advisor_runs" }

func CanTransitionAdvisorRun(from, to string) bool {
	switch from {
	case AdvisorStatusQueued:
		return to == AdvisorStatusRunning || to == AdvisorStatusCancelled || to == AdvisorStatusPurged
	case AdvisorStatusRunning:
		return to == AdvisorStatusCompleted || to == AdvisorStatusFailed || to == AdvisorStatusCancelled || to == AdvisorStatusInterrupted || to == AdvisorStatusPurged
	default:
		return false
	}
}

type AdvisorEvent struct {
	RunID     string    `json:"run_id" gorm:"primaryKey"`
	Seq       int64     `json:"seq" gorm:"primaryKey"`
	Type      string    `json:"type" gorm:"not null"`
	Payload   any       `json:"payload" gorm:"serializer:json;type:jsonb"`
	CreatedAt time.Time `json:"created_at"`
}

func (AdvisorEvent) TableName() string { return "advisor_events" }

type AdvisorFactBundle struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	OwnerID   string    `json:"owner_id" gorm:"not null;index"`
	RunID     string    `json:"run_id" gorm:"not null;index"`
	Scope     any       `json:"scope" gorm:"serializer:json;type:jsonb"`
	AsOf      time.Time `json:"as_of"`
	Facts     []Fact    `json:"facts" gorm:"serializer:json;type:jsonb"`
	CreatedAt time.Time `json:"created_at"`
}

func (AdvisorFactBundle) TableName() string { return "advisor_fact_bundles" }
