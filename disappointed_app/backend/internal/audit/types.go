package audit

import (
	"encoding/json"
	"time"
)

type Outcome string
type Severity string
type Source string

const (
	OutcomeSuccess  Outcome = "success"
	OutcomeFailure  Outcome = "failure"
	OutcomeDenied   Outcome = "denied"
	OutcomeConflict Outcome = "conflict"
	OutcomeReplayed Outcome = "replayed"

	SeverityInfo     Severity = "info"
	SeverityWarn     Severity = "warn"
	SeverityError    Severity = "error"
	SeveritySecurity Severity = "security"

	SourceWeb    Source = "web"
	SourcePWA    Source = "pwa"
	SourceAPI    Source = "api"
	SourceWorker Source = "worker"
	SourceSync   Source = "sync"
	SourceAuth   Source = "auth"
	SourceSystem Source = "system"
)

type Event struct {
	ID             string          `json:"id"`
	OccurredAt     time.Time       `json:"occurred_at"`
	CorrelationID  string          `json:"correlation_id"`
	ActorUserID    string          `json:"actor_user_id,omitempty"`
	ActorEmailHash string          `json:"actor_email_hash,omitempty"`
	Action         string          `json:"action"`
	EntityType     string          `json:"entity_type"`
	EntityID       string          `json:"entity_id"`
	Outcome        Outcome         `json:"outcome"`
	Severity       Severity        `json:"severity"`
	Source         Source          `json:"source"`
	ErrorCode      string          `json:"error_code,omitempty"`
	RequestMethod  string          `json:"request_method,omitempty"`
	RequestPath    string          `json:"request_path,omitempty"`
	IPHash         string          `json:"ip_hash,omitempty"`
	UserAgentHash  string          `json:"user_agent_hash,omitempty"`
	Metadata       json.RawMessage `json:"metadata_json,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}

type Query struct {
	CorrelationID string
	Action        string
	Severity      Severity
	From          time.Time
	To            time.Time
	Limit         int
	Before        time.Time
}
