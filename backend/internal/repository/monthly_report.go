package repository

import (
	"context"
	"time"

	"github.com/mypocket/backend/internal/entity"
)

type MonthlyReportReader interface {
	ReadMonthly(context.Context, string, entity.MonthlyReportScope, time.Time) (entity.MonthlyReport, error)
}
