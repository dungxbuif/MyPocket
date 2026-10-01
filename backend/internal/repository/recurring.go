package repository

import (
	"errors"
	"time"

	"github.com/mypocket/backend/internal/entity"
)

var (
	ErrRecurringInvalid         = errors.New("invalid recurring schedule")
	ErrRecurringNotFound        = errors.New("recurring schedule not found")
	ErrRecurringWalletInvalid   = errors.New("invalid recurring wallet")
	ErrRecurringCategoryInvalid = errors.New("invalid recurring category")
)

type RecurringUpdate struct {
	Name        string
	Amount      int64
	CategoryID  *string
	Type        string
	Note        *string
	Frequency   string
	Interval    int
	NextRunAt   time.Time
	EndsAt      *time.Time
	Active      bool
	AnchorDay   int
	AnchorMonth int
}

type RecurringRepository interface {
	List(ownerID string) ([]entity.RecurringSchedule, error)
	Find(ownerID, id string) (*entity.RecurringSchedule, error)
	Create(ownerID string, schedule *entity.RecurringSchedule) error
	Update(ownerID, id string, updates RecurringUpdate) (*entity.RecurringSchedule, error)
	Delete(ownerID, id string) error
	RunDue(ownerID string, now time.Time) (int, error)
}
