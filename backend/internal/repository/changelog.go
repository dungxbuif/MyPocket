package repository

import (
	"context"

	"github.com/mypocket/backend/internal/entity"
)

type ChangelogRepository interface {
	Publish(context.Context, *entity.Changelog, []string) error
	ListPublic(context.Context, int) ([]entity.Changelog, error)
	FindPublic(context.Context, string) (*entity.Changelog, error)
}
