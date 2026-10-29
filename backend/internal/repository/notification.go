package repository

import (
	"context"
	"github.com/mypocket/backend/internal/entity"
)

type NotificationRepository interface {
	List(context.Context, string) ([]entity.Notification, error)
	MarkRead(context.Context, string, string) error
}
