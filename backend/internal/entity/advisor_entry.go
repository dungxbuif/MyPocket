package entity

import (
	"errors"
	"strings"
	"time"
)

const (
	AdvisorEntryTaskAccepted    = "accepted"
	AdvisorEntryTaskProcessing  = "processing"
	AdvisorEntryTaskReviewReady = "review_ready"
	AdvisorEntryTaskFailed      = "failed"
	AdvisorEntryTaskCancelled   = "cancelled"

	AdvisorEntryProposalPending    = "pending"
	AdvisorEntryProposalApplying   = "applying"
	AdvisorEntryProposalApplied    = "applied"
	AdvisorEntryProposalRejected   = "rejected"
	AdvisorEntryProposalExpired    = "expired"
	AdvisorEntryProposalConflict   = "conflict"
	AdvisorEntryProposalSuperseded = "superseded"
)

var ErrAdvisorEntryTaskInvalid = errors.New("advisor entry task is invalid")

type AdvisorEntryTask struct {
	ID              string     `json:"id"`
	OwnerID         string     `json:"owner_id"`
	ConversationID  string     `json:"conversation_id"`
	Generation      int64      `json:"generation"`
	ClientRequestID string     `json:"client_request_id"`
	PayloadHash     string     `json:"-"`
	SessionID       string     `json:"session_id"`
	Status          string     `json:"status"`
	Stage           string     `json:"stage,omitempty"`
	BatchID         string     `json:"batch_id,omitempty"`
	BatchVersion    int64      `json:"batch_version,omitempty"`
	ErrorCode       string     `json:"error_code,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
}

func (AdvisorEntryTask) TableName() string { return "advisor_entry_tasks" }

func (task AdvisorEntryTask) Validate() error {
	if strings.TrimSpace(task.ID) == "" || strings.TrimSpace(task.OwnerID) == "" ||
		strings.TrimSpace(task.ConversationID) == "" || task.Generation <= 0 ||
		strings.TrimSpace(task.ClientRequestID) == "" || strings.TrimSpace(task.PayloadHash) == "" ||
		strings.TrimSpace(task.SessionID) == "" || task.BatchVersion < 0 ||
		(task.BatchID != "" && task.BatchVersion == 0) ||
		(task.BatchID == "" && task.BatchVersion != 0) {
		return ErrAdvisorEntryTaskInvalid
	}
	switch task.Status {
	case AdvisorEntryTaskAccepted, AdvisorEntryTaskProcessing, AdvisorEntryTaskReviewReady,
		AdvisorEntryTaskFailed, AdvisorEntryTaskCancelled:
		return nil
	default:
		return ErrAdvisorEntryTaskInvalid
	}
}

func CanTransitionAdvisorEntryTask(from, to string) bool {
	switch from {
	case AdvisorEntryTaskAccepted:
		return to == AdvisorEntryTaskProcessing || to == AdvisorEntryTaskFailed || to == AdvisorEntryTaskCancelled
	case AdvisorEntryTaskProcessing:
		return to == AdvisorEntryTaskReviewReady || to == AdvisorEntryTaskFailed || to == AdvisorEntryTaskCancelled
	default:
		return false
	}
}

// AdvisorEntryBatchPart is a reference-only conversation part. The entry
// service remains the source of truth for proposal drafts and their versions.
type AdvisorEntryBatchPart struct {
	TaskID    string   `json:"task_id"`
	BatchID   string   `json:"batch_id"`
	Version   int64    `json:"version"`
	ItemIDs   []string `json:"item_ids"`
	ItemCount int      `json:"item_count"`
	Summary   string   `json:"summary"`
}

type AdvisorEntryTaskEvent struct {
	TaskID    string    `json:"task_id"`
	OwnerID   string    `json:"owner_id"`
	Seq       int64     `json:"seq"`
	Type      string    `json:"type"`
	Payload   any       `json:"payload"`
	CreatedAt time.Time `json:"created_at"`
}
