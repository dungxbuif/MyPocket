package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	contract "github.com/mypocket/backend/internal/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ReportSnapshotPostgresRepository struct {
	db      *gorm.DB
	monthly *MonthlyReportPostgresRepository
}

func NewReportSnapshotPostgresRepository(db *gorm.DB) *ReportSnapshotPostgresRepository {
	return &ReportSnapshotPostgresRepository{db: db, monthly: NewMonthlyReportPostgresRepository(db)}
}

type reportSnapshotRow struct {
	ID           string
	OwnerID      string
	Month        string
	WalletID     string
	Granularity  string
	Timezone     string
	AsOfDay      string
	DataRevision int64
	RuleVersion  string
	AsOf         time.Time
	ExpiresAt    time.Time
	CreatedAt    time.Time
	Payload      []byte
}

func (r *ReportSnapshotPostgresRepository) ReadOrCreate(ctx context.Context, ownerID string, scope entity.MonthlyReportScope, now time.Time) (*entity.ReportSnapshot, error) {
	var snapshot *entity.ReportSnapshot
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`INSERT INTO report_revisions(owner_id, revision, updated_at) SELECT id, 1, now() FROM "user" WHERE id = ? ON CONFLICT (owner_id) DO NOTHING`, ownerID).Error; err != nil {
			return err
		}
		revision, err := r.monthly.revision(ctx, tx, ownerID, true)
		if err != nil {
			return err
		}
		queryScope, err := r.monthly.resolveScope(ctx, tx, ownerID, scope, now)
		if err != nil {
			return err
		}
		asOfDay := now.In(queryScope.Location).Format("2006-01-02")
		if !queryScope.IsCurrentMonth {
			asOfDay = queryScope.MonthNext.In(queryScope.Location).AddDate(0, 0, -1).Format("2006-01-02")
		}
		var row reportSnapshotRow
		query := tx.WithContext(ctx).Table("report_snapshots").Where("owner_id = ? AND month = ? AND wallet_id = ? AND granularity = ? AND timezone = ? AND as_of_day = ? AND data_revision = ? AND rule_version = ? AND expires_at > ?", ownerID, queryScope.Scope.Month, queryScope.Scope.WalletID, queryScope.Scope.Granularity, queryScope.Location.String(), asOfDay, revision, entity.MonthlyReportRuleVersion, now.UTC()).Order("created_at DESC").First(&row)
		if query.Error == nil {
			snapshot, err = decodeReportSnapshot(row)
			return err
		}
		if !errors.Is(query.Error, gorm.ErrRecordNotFound) {
			return query.Error
		}
		report, err := r.monthly.readMonthlyInTx(ctx, tx, ownerID, scope, now, revision)
		if err != nil {
			return err
		}
		id := uuid.NewString()
		report.SnapshotID = id
		report.DataRevision = revision
		payload, err := json.Marshal(report)
		if err != nil {
			return err
		}
		expiresAt := now.UTC().Add(entity.ReportSnapshotRetention)
		if queryScope.IsCurrentMonth {
			nextLocalDay := time.Date(now.In(queryScope.Location).Year(), now.In(queryScope.Location).Month(), now.In(queryScope.Location).Day()+1, 0, 0, 0, 0, queryScope.Location).UTC()
			if nextLocalDay.Before(expiresAt) {
				expiresAt = nextLocalDay
			}
		}
		row = reportSnapshotRow{ID: id, OwnerID: ownerID, Month: queryScope.Scope.Month, WalletID: queryScope.Scope.WalletID, Granularity: queryScope.Scope.Granularity, Timezone: queryScope.Location.String(), AsOfDay: asOfDay, DataRevision: revision, RuleVersion: entity.MonthlyReportRuleVersion, AsOf: now.UTC(), ExpiresAt: expiresAt, CreatedAt: now.UTC(), Payload: payload}
		if err := tx.WithContext(ctx).Exec(`INSERT INTO report_snapshots(id, owner_id, month, wallet_id, granularity, timezone, as_of_day, data_revision, rule_version, as_of, expires_at, created_at, payload) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, row.ID, row.OwnerID, row.Month, row.WalletID, row.Granularity, row.Timezone, row.AsOfDay, row.DataRevision, row.RuleVersion, row.AsOf, row.ExpiresAt, row.CreatedAt, row.Payload).Error; err != nil {
			return err
		}
		snapshot, err = decodeReportSnapshot(row)
		return err
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	return snapshot, err
}

func (r *ReportSnapshotPostgresRepository) GetOwned(ctx context.Context, ownerID, snapshotID string, now time.Time) (*entity.ReportSnapshot, error) {
	snapshot, err := r.GetRetainedOwned(ctx, ownerID, snapshotID, now)
	if err != nil {
		return nil, err
	}
	if !snapshot.ExpiresAt.After(now.UTC()) {
		return nil, contract.ErrSnapshotExpired
	}
	return snapshot, nil
}

func (r *ReportSnapshotPostgresRepository) GetRetainedOwned(ctx context.Context, ownerID, snapshotID string, now time.Time) (*entity.ReportSnapshot, error) {
	var row reportSnapshotRow
	err := r.db.WithContext(ctx).Table("report_snapshots").Where("id = ? AND owner_id = ?", snapshotID, ownerID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, contract.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if !row.CreatedAt.Add(entity.ReportSnapshotRetention).After(now.UTC()) {
		return nil, contract.ErrSnapshotExpired
	}
	return decodeReportSnapshot(row)
}

func (r *ReportSnapshotPostgresRepository) CurrentRevision(ctx context.Context, ownerID string) (int64, error) {
	return r.monthly.revision(ctx, r.db, ownerID, false)
}

func (r *ReportSnapshotPostgresRepository) DeleteExpired(ctx context.Context, now time.Time, limit int) (int, error) {
	if limit <= 0 {
		return 0, nil
	}
	var ids []string
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		retentionCutoff := now.UTC().Add(-entity.ReportSnapshotRetention)
		query := tx.WithContext(ctx).Table("report_snapshots AS s").Select("s.id").Where("s.created_at <= ?", retentionCutoff).Where("NOT EXISTS (SELECT 1 FROM report_analyses a WHERE a.snapshot_id = s.id AND (a.status IN (?, ?) OR a.created_at > ?))", entity.ReportAnalysisStatusQueued, entity.ReportAnalysisStatusRunning, retentionCutoff).Order("s.created_at ASC").Limit(limit).Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).Scan(&ids)
		if query.Error != nil {
			return query.Error
		}
		if len(ids) == 0 {
			return nil
		}
		if err := tx.Exec("DELETE FROM report_analyses WHERE snapshot_id IN ?", ids).Error; err != nil {
			return err
		}
		return tx.Exec("DELETE FROM report_snapshots WHERE id IN ?", ids).Error
	}, &sql.TxOptions{})
	return len(ids), err
}

func decodeReportSnapshot(row reportSnapshotRow) (*entity.ReportSnapshot, error) {
	var report entity.MonthlyReport
	if err := json.Unmarshal(row.Payload, &report); err != nil {
		return nil, err
	}
	return &entity.ReportSnapshot{ID: row.ID, OwnerID: row.OwnerID, Scope: entity.MonthlyReportScope{Month: row.Month, WalletID: row.WalletID, Granularity: row.Granularity}, Timezone: row.Timezone, AsOfDay: row.AsOfDay, DataRevision: row.DataRevision, RuleVersion: row.RuleVersion, AsOf: row.AsOf, ExpiresAt: row.ExpiresAt, CreatedAt: row.CreatedAt, Payload: row.Payload, Report: report}, nil
}
