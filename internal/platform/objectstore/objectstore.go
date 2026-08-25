package objectstore

import "context"

type Store interface {
	PutSmokeObject(ctx context.Context) error
	DeleteSmokeObject(ctx context.Context) error
}
