package repository

import (
	"context"
	"github.com/mypocket/backend/internal/entity"
)

type DeviceCredentialRepository interface {
	Save(context.Context, *entity.DeviceCredential) error
	FindByID(context.Context, string) (*entity.DeviceCredential, error)
	Revoke(ctx context.Context, ownerID, id string) error
}
