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
