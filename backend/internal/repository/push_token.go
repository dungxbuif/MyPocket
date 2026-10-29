package repository

import (
	"context"
	"errors"

	"github.com/mypocket/backend/internal/entity"
)

var ErrPushTokenConflict = errors.New("push token belongs to another owner")

type PushTokenRepository interface {
	Save(context.Context, *entity.PushToken) error
	Revoke(context.Context, string, string) error
}
