package agent

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrToolNotFound = errors.New("image tool document not found")
	ErrToolExpired  = errors.New("image tool document expired")
	ErrToolQuota    = errors.New("image tool quota exceeded")
)

type ToolRunStatus string

const (
	ToolQueued     ToolRunStatus = "queued"
	ToolSubmitting ToolRunStatus = "submitting"
	ToolProcessing ToolRunStatus = "processing"
	ToolCompleted  ToolRunStatus = "completed"
	ToolFailed     ToolRunStatus = "failed"
	ToolCancelled  ToolRunStatus = "cancelled"
	ToolExpired    ToolRunStatus = "expired"
)

type ImageInput struct {
	ContentType string
	Bytes       []byte
	Languages   []string
}
type ToolSubmission struct {
	ProviderID string
	Status     ToolRunStatus
	RetryAfter time.Duration
}
type ToolResult struct {
	Status     ToolRunStatus
	Text       string
	Fields     map[string]any
	ExpiresAt  time.Time
	RetryAfter time.Duration
	ErrorCode  string
}
type ImageTool interface {
	Submit(context.Context, ImageInput) (ToolSubmission, error)
	Read(context.Context, string) (ToolResult, error)
}

type ToolRun struct {
	ID                 string          `json:"id"`
	UserID             string          `json:"-"`
	AgentRunID         string          `json:"agent_run_id"`
	ReceiptObjectID    string          `json:"receipt_object_id,omitempty"`
	ProviderDocumentID string          `json:"-"`
	Status             ToolRunStatus   `json:"status"`
	Attempts           int             `json:"attempts"`
	Result             json.RawMessage `json:"result,omitempty"`
	ErrorCode          string          `json:"error_code,omitempty"`
	CreatedAt          time.Time       `json:"created_at"`
	NextAttemptAt      time.Time       `json:"-"`
	LeaseOwner         string          `json:"-"`
	LeaseExpiresAt     time.Time       `json:"-"`
	ObjectKey          string          `json:"-"`
	ContentType        string          `json:"-"`
	ChecksumSHA256     string          `json:"-"`
	SizeBytes          int64           `json:"-"`
}
