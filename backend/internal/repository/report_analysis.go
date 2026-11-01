package repository

import (
	"context"
	"time"

	"github.com/mypocket/backend/internal/entity"
)

var (
	ErrAnalysisRequestConflict = &RepositoryError{Code: "analysis_request_conflict"}
	ErrAnalysisActive          = &RepositoryError{Code: "analysis_active"}
	ErrAnalysisCooldown        = &RepositoryError{Code: "analysis_cooldown"}
	ErrAnalysisLeaseLost       = &RepositoryError{Code: "analysis_lease_lost"}
)

type ReportAnalysisRepository interface {
	CreateOrGet(context.Context, *entity.ReportAnalysis) (*entity.ReportAnalysis, bool, error)
	GetOwned(context.Context, string, string) (*entity.ReportAnalysis, error)
	LatestForScope(context.Context, string, entity.MonthlyReportScope, string, time.Time) (*entity.ReportAnalysis, *entity.ReportAnalysis, error)
	Cancel(context.Context, string, string, time.Time) (*entity.ReportAnalysis, error)
	ClaimNext(context.Context, time.Time, time.Duration) (*entity.ReportAnalysis, error)
	RenewLease(context.Context, string, string, time.Time, time.Duration) error
	RecoverStale(context.Context, time.Time, int) error
	Complete(context.Context, string, string, time.Time, *entity.ReportAnalysis) error
	Fail(context.Context, string, string, time.Time, string) error
	DeleteExpired(context.Context, time.Time, int) (int, error)
}
