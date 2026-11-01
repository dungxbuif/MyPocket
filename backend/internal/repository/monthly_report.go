package repository

import (
	"context"
	"time"

	"github.com/mypocket/backend/internal/entity"
)

type MonthlyReportReader interface {
	ReadMonthly(context.Context, string, entity.MonthlyReportScope, time.Time) (entity.MonthlyReport, error)
}

// MonthlyReportTransactionReader is optional so existing report readers remain
// source compatible while native clients request bounded drill-down pages.
type MonthlyReportTransactionReader interface {
	ReadMonthlyTransactions(context.Context, string, entity.MonthlyReportScope, string, string, int, int, time.Time) (entity.SavingsTransactionPage, error)
}
