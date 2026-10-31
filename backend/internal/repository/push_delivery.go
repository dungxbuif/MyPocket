package repository

import (
	"context"
	"errors"
	"time"

	"github.com/mypocket/backend/internal/entity"
)

var ErrPushDeliveryLeaseLost = errors.New("push delivery lease lost")

type PushDeliveryJob struct {
	Delivery   entity.PushDelivery
	LeaseUntil time.Time
}

// PushDeliveryRepository owns the durable hand-off from inbox notifications
// to provider delivery. Implementations must fence every terminal/retry update
// with the lease value returned by ClaimPushDeliveries.
type PushDeliveryRepository interface {
	EnqueueNotificationDeliveries(context.Context, time.Time, int) (int, error)
	ClaimPushDeliveries(context.Context, time.Time, int, time.Duration, string) ([]PushDeliveryJob, error)
	MarkPushDeliverySent(context.Context, string, time.Time, time.Time) error
	MarkPushDeliveryRetry(context.Context, string, time.Time, time.Time, string) error
	MarkPushDeliveryRevoked(context.Context, string, time.Time, time.Time) error
	MarkPushDeliveryFailed(context.Context, string, time.Time, time.Time, string) error
}

// PushTokenDeliveryRepository is the internal read/revoke surface used by the
// worker. RevokeIfCurrent prevents a 410 for an old token from revoking a newer
// registration stored on the same device row.
type PushTokenDeliveryRepository interface {
	LoadForDelivery(context.Context, string, string) (entity.PushToken, error)
	RevokeIfCurrent(context.Context, string, string, string, time.Time) (bool, error)
}
