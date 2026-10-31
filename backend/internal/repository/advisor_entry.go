package repository

import (
	"context"
	"time"

	"github.com/mypocket/backend/internal/entity"
)

// AdvisorEntryTaskLease is private executor state. Every write using it must
// match owner, conversation, generation, task and token against current rows.
type AdvisorEntryTaskLease struct {
	OwnerID        string
	ConversationID string
	Generation     int64
	TaskID         string
	Token          string
	LeaseUntil     time.Time
}

// AdvisorEntryTaskRepository owns durable entry lifecycle storage. All IDs are
// scoped by owner, and missing/foreign rows have the same not-found result.
type AdvisorEntryTaskRepository interface {
	CreateTask(context.Context, entity.AdvisorEntryTask) (*entity.AdvisorEntryTask, error)
	GetTask(context.Context, string, string) (*entity.AdvisorEntryTask, error)
	GetTaskByRequest(context.Context, string, string, int64, string) (*entity.AdvisorEntryTask, error)
	ClaimTask(context.Context, string, string, string) (AdvisorEntryTaskLease, error)
	HeartbeatTask(context.Context, AdvisorEntryTaskLease) error
	CompleteTask(context.Context, AdvisorEntryTaskLease, entity.AdvisorEntryBatchPart) error
	FailTask(context.Context, AdvisorEntryTaskLease, string) error
	CancelTask(context.Context, string, string) error
	AppendEvent(context.Context, string, string, string, any) (*entity.AdvisorEntryTaskEvent, error)
	ListEvents(context.Context, string, string, int64, int) ([]entity.AdvisorEntryTaskEvent, error)
	ReapExpiredTasks(context.Context, int) (int, error)
}
