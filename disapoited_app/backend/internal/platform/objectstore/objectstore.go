package objectstore

import (
	"context"
	"time"
)

type Store interface {
	PutSmokeObject(ctx context.Context) error
	DeleteSmokeObject(ctx context.Context) error
	PresignPut(ctx context.Context, key string, contentType string, expires time.Duration) (string, error)
	PresignGet(ctx context.Context, key string, expires time.Duration) (string, error)
	DeleteObject(ctx context.Context, key string) error
}
