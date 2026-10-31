package repository

import (
	"context"
	"time"

	"github.com/mypocket/backend/internal/entity"
)

var (
	ErrSnapshotExpired = &RepositoryError{Code: "snapshot_expired"}
)

type RepositoryError struct{ Code string }

func (e *RepositoryError) Error() string { return e.Code }

type ReportSnapshotRepository interface {
	ReadOrCreate(context.Context, string, entity.MonthlyReportScope, time.Time) (*entity.ReportSnapshot, error)
	GetOwned(context.Context, string, string, time.Time) (*entity.ReportSnapshot, error)
	CurrentRevision(context.Context, string) (int64, error)
	DeleteExpired(context.Context, time.Time, int) (int, error)
}
