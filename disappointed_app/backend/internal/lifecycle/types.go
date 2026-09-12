package lifecycle

import (
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("lifecycle job not found")
	ErrConflict = errors.New("lifecycle job version conflict")
	ErrInvalid  = errors.New("invalid lifecycle request")
)

type Kind string
type Status string

const (
	KindImport Kind = "import"
	KindExport Kind = "export"
	KindReset  Kind = "reset"
	KindDelete Kind = "delete"

	StatusQueued               Status = "queued"
	StatusRunning              Status = "running"
	StatusAwaitingConfirmation Status = "awaiting_confirmation"
	StatusCompleted            Status = "completed"
	StatusFailed               Status = "failed"
)

type Job struct {
	ID              string          `json:"id"`
	UserID          string          `json:"user_id,omitempty"`
	Kind            Kind            `json:"kind"`
	Status          Status          `json:"status"`
	Request         json.RawMessage `json:"request,omitempty"`
	Result          json.RawMessage `json:"result,omitempty"`
	ResultObjectKey string          `json:"-"`
	Attempts        int             `json:"attempts"`
	Version         int64           `json:"version"`
	ErrorCode       string          `json:"error_code,omitempty"`
	NextAttemptAt   time.Time       `json:"next_attempt_at"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type ExportRequest struct {
	From       *time.Time `json:"from,omitempty"`
	To         *time.Time `json:"to,omitempty"`
	Datasets   []string   `json:"datasets"`
	SnapshotAt time.Time  `json:"snapshot_at,omitempty"`
}

type ImportRequest struct {
	ObjectKey string `json:"object_key"`
}

type ImportRow struct {
	Row                 int       `json:"row"`
	OccurredAt          time.Time `json:"occurred_at"`
	Type                string    `json:"type"`
	AmountVND           int64     `json:"amount_vnd"`
	SourceWalletID      string    `json:"source_wallet_id"`
	DestinationWalletID string    `json:"destination_wallet_id,omitempty"`
	CategoryID          string    `json:"category_id,omitempty"`
	Note                string    `json:"note,omitempty"`
	ExcludedFromReports bool      `json:"excluded_from_reports"`
}

type ImportRowError struct {
	Row     int    `json:"row"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ImportPreview struct {
	ID          string           `json:"id"`
	Version     int64            `json:"version"`
	ValidRows   []ImportRow      `json:"valid_rows"`
	Errors      []ImportRowError `json:"errors"`
	Confirmable bool             `json:"confirmable"`
}

type DestructiveRequest struct {
	Confirmation string `json:"confirmation"`
	PreviewToken string `json:"preview_token"`
}

type DestructivePreview struct {
	Token          string         `json:"preview_token"`
	AffectedCounts map[string]int `json:"affected_counts"`
	ExpiresAt      time.Time      `json:"expires_at"`
}
