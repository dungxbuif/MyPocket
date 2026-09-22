package repository

import (
	"context"
	"errors"
	"time"

	"github.com/mypocket/backend/internal/entity"
)

var (
	ErrFeedbackInvalid  = errors.New("invalid feedback")
	ErrFeedbackConflict = errors.New("feedback state conflict")
)

type FeedbackRepository interface {
	Create(context.Context, *entity.Feedback) error
	ListByOwner(context.Context, string) ([]entity.Feedback, error)
	FindByOwner(context.Context, string, string) (*entity.Feedback, error)
	ListForAgent(context.Context, string, int) ([]entity.Feedback, error)
	UpdateStatus(context.Context, string, string, time.Time) (*entity.Feedback, error)
}
