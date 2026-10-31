package main

import (
	"context"
	"time"

	"github.com/mypocket/backend/internal/usecase"
)

func runReportAnalysisPass(ctx context.Context, worker *usecase.ReportAnalysisWorker) (int, error) {
	if worker == nil {
		return 0, nil
	}
	return worker.RunPass(ctx, 10)
}

func runReportRetentionPass(ctx context.Context, analyses interface {
	DeleteExpired(context.Context, time.Time, int) (int, error)
}, snapshots interface {
	DeleteExpired(context.Context, time.Time, int) (int, error)
}, now time.Time) (int, error) {
	deletedAnalyses, err := analyses.DeleteExpired(ctx, now, 100)
	if err != nil {
		return deletedAnalyses, err
	}
	deletedSnapshots, err := snapshots.DeleteExpired(ctx, now, 100)
	if err != nil {
		return deletedAnalyses + deletedSnapshots, err
	}
	return deletedAnalyses + deletedSnapshots, nil
}
