package agent

import (
	"context"
	"encoding/json"
	"time"
)

type Kind string

const (
	KindTransactionDraft Kind = "transaction_draft"
	KindAnalysis         Kind = "analysis"
	KindIntake           Kind = "intake"
	KindAdvisor          Kind = "advisor"
)

type Status string

const (
	StatusQueued     Status = "queued"
	StatusProcessing Status = "processing"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
)

type Run struct {
	ID             string    `json:"id"`
	UserID         string    `json:"-"`
	SessionID      string    `json:"session_id,omitempty"`
	Kind           Kind      `json:"kind"`
	Status         Status    `json:"status"`
	RequestText    string    `json:"request_text"`
	ResponseText   string    `json:"response_text,omitempty"`
	ErrorCode      string    `json:"error_code,omitempty"`
	Attempts       int       `json:"attempts"`
	DraftIDs       []string  `json:"draft_ids"`
	ToolRuns       []ToolRun `json:"tool_runs,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	NextAttemptAt  time.Time `json:"-"`
	LeaseOwner     string    `json:"-"`
	LeaseExpiresAt time.Time `json:"-"`
}

type MessageRole string

const (
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
	MessageRoleSystem    MessageRole = "system"
)

const ActionDraftsCreated = "drafts_created"

type ActionCard struct {
	Type     string   `json:"type"`
	DraftIDs []string `json:"draft_ids,omitempty"`
}

type Session struct {
	ID             string    `json:"id"`
	UserID         string    `json:"-"`
	Kind           Kind      `json:"kind"`
	Title          string    `json:"title"`
	ContextSummary string    `json:"context_summary,omitempty"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Message struct {
	ID        string      `json:"id"`
	UserID    string      `json:"-"`
	SessionID string      `json:"session_id"`
	RunID     string      `json:"run_id,omitempty"`
	Role      MessageRole `json:"role"`
	Text      string      `json:"text"`
	Action    *ActionCard `json:"action,omitempty"`
	CreatedAt time.Time   `json:"created_at"`
}

type SessionHistory struct {
	Session  Session   `json:"session"`
	Messages []Message `json:"messages"`
}

type ModelRequest struct {
	System string
	Input  string
	Schema json.RawMessage
}

type ModelResponse struct {
	Text    string
	Refusal string
}

type Model interface {
	Generate(context.Context, ModelRequest) (ModelResponse, error)
}

type ProposedTransaction struct {
	Type                string  `json:"type"`
	AmountVND           int64   `json:"amount_vnd"`
	SourceWalletID      string  `json:"source_wallet_id"`
	DestinationWalletID *string `json:"destination_wallet_id,omitempty"`
	CategoryID          *string `json:"category_id,omitempty"`
	OccurredAt          string  `json:"occurred_at"`
	Note                string  `json:"note"`
}

type ModelResult struct {
	Answer      string                `json:"answer,omitempty"`
	Transaction *ProposedTransaction  `json:"transaction,omitempty"`
	Drafts      []ProposedTransaction `json:"drafts,omitempty"`
}
