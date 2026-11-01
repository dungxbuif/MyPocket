package usecase

import (
	"context"
	"time"

	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
)

type MonthlyReportService struct {
	Reader    repository.MonthlyReportReader
	Snapshots repository.ReportSnapshotRepository
	Now       func() time.Time
}

func NewMonthlyReportService(reader repository.MonthlyReportReader, snapshots repository.ReportSnapshotRepository) *MonthlyReportService {
	return &MonthlyReportService{Reader: reader, Snapshots: snapshots, Now: time.Now}
}

func (s *MonthlyReportService) Read(ctx context.Context, principal Principal, scope entity.MonthlyReportScope) (*entity.MonthlyReport, error) {
	now := time.Now().UTC()
	if s != nil && s.Now != nil {
		now = s.Now().UTC()
	}
	if s == nil || s.Reader == nil {
		return nil, repository.ErrNotFound
	}
	if s.Snapshots != nil {
		snapshot, err := s.Snapshots.ReadOrCreate(ctx, principal.OwnerID, scope, now)
		if err != nil {
			return nil, err
		}
		return &snapshot.Report, nil
	}
	report, err := s.Reader.ReadMonthly(ctx, principal.OwnerID, scope, now)
	if err != nil {
		return nil, err
	}
	report.Insights = entity.BuildReportInsights(report)
	return &report, nil
}

func (s *MonthlyReportService) Snapshot(ctx context.Context, principal Principal, snapshotID string) (*entity.ReportSnapshot, error) {
	if s == nil || s.Snapshots == nil {
		return nil, repository.ErrNotFound
	}
	now := time.Now().UTC()
	if s.Now != nil {
		now = s.Now().UTC()
	}
	return s.Snapshots.GetOwned(ctx, principal.OwnerID, snapshotID, now)
}

func (s *MonthlyReportService) Transactions(ctx context.Context, principal Principal, scope entity.MonthlyReportScope, categoryID, kind string, limit, offset int) (entity.SavingsTransactionPage, error) {
	if s == nil {
		return entity.SavingsTransactionPage{}, repository.ErrNotFound
	}
	reader, ok := s.Reader.(repository.MonthlyReportTransactionReader)
	if !ok {
		return entity.SavingsTransactionPage{}, repository.ErrNotFound
	}
	now := time.Now().UTC()
	if s.Now != nil {
		now = s.Now().UTC()
	}
	return reader.ReadMonthlyTransactions(ctx, principal.OwnerID, scope, categoryID, kind, limit, offset, now)
}
