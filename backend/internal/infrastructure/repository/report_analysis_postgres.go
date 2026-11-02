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

type ReportAnalysisPostgresRepository struct{ db *gorm.DB }

func NewReportAnalysisPostgresRepository(db *gorm.DB) *ReportAnalysisPostgresRepository {
	return &ReportAnalysisPostgresRepository{db: db}
}

type reportAnalysisRow struct {
	ID                     string
	OwnerID                string
	SnapshotID             string
	JobKey                 string
	ClientRequestID        string
	PayloadHash            string
	AnalysisFingerprint    string
	Revision               int64
	RuleVersion            string
	PromptVersion          string
	ProviderProfileVersion string
	Status                 string
	ErrorCode              string
	ProviderCalled         bool
	Summary                string
	Observations           []byte
	Suggestions            []byte
	FactRefs               []byte
	Usage                  []byte
	CreatedAt              time.Time
	UpdatedAt              time.Time
	StartedAt              *time.Time
	FinishedAt             *time.Time
	LeaseUntil             *time.Time
	LeaseToken             string
	DeadlineAt             *time.Time
}

func (r *ReportAnalysisPostgresRepository) CreateOrGet(ctx context.Context, analysis *entity.ReportAnalysis) (*entity.ReportAnalysis, bool, error) {
	var result *entity.ReportAnalysis
	replay := false
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtextextended(?, 0))", analysis.OwnerID).Error; err != nil {
			return err
		}
		var existing reportAnalysisRow
		retentionCutoff := analysis.CreatedAt.UTC().Add(-entity.ReportSnapshotRetention)
		if err := tx.WithContext(ctx).Table("report_analyses").Where("owner_id = ? AND client_request_id = ?", analysis.OwnerID, analysis.ClientRequestID).First(&existing).Error; err == nil {
			if existing.Status == entity.ReportAnalysisStatusCompleted && !existing.CreatedAt.After(retentionCutoff) {
				if deleteErr := tx.WithContext(ctx).Table("report_analyses").Where("id = ?", existing.ID).Delete(&entity.ReportAnalysis{}).Error; deleteErr != nil {
					return deleteErr
				}
			} else {
				if existing.PayloadHash != analysis.PayloadHash {
					return contract.ErrAnalysisRequestConflict
				}
				decoded, decodeErr := decodeReportAnalysis(existing)
				result, replay = decoded, true
				return decodeErr
			}
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err := tx.WithContext(ctx).Table("report_analyses").Where("owner_id = ? AND analysis_fingerprint = ? AND (status IN (?, ?) OR (status = ? AND created_at > ?))", analysis.OwnerID, analysis.AnalysisFingerprint, entity.ReportAnalysisStatusQueued, entity.ReportAnalysisStatusRunning, entity.ReportAnalysisStatusCompleted, retentionCutoff).Order("created_at DESC").First(&existing).Error; err == nil {
			decoded, decodeErr := decodeReportAnalysis(existing)
			result, replay = decoded, true
			return decodeErr
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err := tx.WithContext(ctx).Table("report_analyses").Where("owner_id = ? AND analysis_fingerprint = ? AND status = ? AND created_at <= ?", analysis.OwnerID, analysis.AnalysisFingerprint, entity.ReportAnalysisStatusCompleted, retentionCutoff).Delete(&entity.ReportAnalysis{}).Error; err != nil {
			return err
		}
		var active int64
		if err := tx.WithContext(ctx).Table("report_analyses").Where("owner_id = ? AND status IN (?, ?)", analysis.OwnerID, entity.ReportAnalysisStatusQueued, entity.ReportAnalysisStatusRunning).Count(&active).Error; err != nil {
			return err
		}
		if active > 0 {
			return contract.ErrAnalysisActive
		}
		var recent reportAnalysisRow
		if err := tx.WithContext(ctx).Table("report_analyses").Where("owner_id = ? AND created_at > ?", analysis.OwnerID, analysis.CreatedAt.UTC().Add(-entity.ReportAnalysisCooldown)).Order("created_at DESC").First(&recent).Error; err == nil {
			return contract.ErrAnalysisCooldown
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err := insertReportAnalysis(tx, analysis); err != nil {
			var retry reportAnalysisRow
			if queryErr := tx.WithContext(ctx).Table("report_analyses").Where("owner_id = ? AND analysis_fingerprint = ? AND status IN (?, ?, ?)", analysis.OwnerID, analysis.AnalysisFingerprint, entity.ReportAnalysisStatusQueued, entity.ReportAnalysisStatusRunning, entity.ReportAnalysisStatusCompleted).First(&retry).Error; queryErr == nil {
				decoded, decodeErr := decodeReportAnalysis(retry)
				result, replay = decoded, true
				return decodeErr
			}
			return err
		}
		result = analysis
		return nil
	})
	return result, replay, err
}

func (r *ReportAnalysisPostgresRepository) LatestForScope(ctx context.Context, ownerID string, scope entity.MonthlyReportScope, timezone string, now time.Time) (*entity.ReportAnalysis, *entity.ReportAnalysis, error) {
	read := func(statuses []string) (*entity.ReportAnalysis, error) {
		var row reportAnalysisRow
		query := r.db.WithContext(ctx).Table("report_analyses AS a").Select("a.*").Joins("JOIN report_snapshots s ON s.id = a.snapshot_id").Where("a.owner_id = ? AND s.month = ? AND s.wallet_id = ? AND s.timezone = ? AND a.created_at > ? AND a.status IN ?", ownerID, scope.Month, scope.WalletID, timezone, now.UTC().Add(-entity.ReportSnapshotRetention), statuses).Order("a.created_at DESC, a.id DESC").First(&row)
		if errors.Is(query.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		if query.Error != nil {
			return nil, query.Error
		}
		return decodeReportAnalysis(row)
	}
	completed, err := read([]string{entity.ReportAnalysisStatusCompleted})
	if err != nil {
		return nil, nil, err
	}
	active, err := read([]string{entity.ReportAnalysisStatusQueued, entity.ReportAnalysisStatusRunning})
	if err != nil {
		return nil, nil, err
	}
	return completed, active, nil
}

func (r *ReportAnalysisPostgresRepository) GetOwned(ctx context.Context, ownerID, analysisID string) (*entity.ReportAnalysis, error) {
	var row reportAnalysisRow
	err := r.db.WithContext(ctx).Table("report_analyses").Where("owner_id = ? AND id = ?", ownerID, analysisID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, contract.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return decodeReportAnalysis(row)
}

func (r *ReportAnalysisPostgresRepository) Cancel(ctx context.Context, ownerID, analysisID string, now time.Time) (*entity.ReportAnalysis, error) {
	result := r.db.WithContext(ctx).Table("report_analyses").Where("owner_id = ? AND id = ? AND status IN (?, ?)", ownerID, analysisID, entity.ReportAnalysisStatusQueued, entity.ReportAnalysisStatusRunning).Updates(map[string]any{"status": entity.ReportAnalysisStatusCancelled, "error_code": "cancelled_by_user", "updated_at": now.UTC(), "finished_at": now.UTC()})
	if result.Error != nil {
		return nil, result.Error
	}
	return r.GetOwned(ctx, ownerID, analysisID)
}

func (r *ReportAnalysisPostgresRepository) ClaimNext(ctx context.Context, now time.Time, lease time.Duration) (*entity.ReportAnalysis, error) {
	var result *entity.ReportAnalysis
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row reportAnalysisRow
		query := tx.WithContext(ctx).Table("report_analyses").Where("status = ? AND (deadline_at IS NULL OR deadline_at > ?)", entity.ReportAnalysisStatusQueued, now.UTC()).Order("created_at ASC, id ASC").Limit(1).Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).First(&row)
		if errors.Is(query.Error, gorm.ErrRecordNotFound) {
			return contract.ErrNotFound
		}
		if query.Error != nil {
			return query.Error
		}
		token := uuid.NewString()
		leaseUntil := now.UTC().Add(lease)
		deadline := now.UTC().Add(entity.ReportAnalysisDeadline)
		if err := tx.WithContext(ctx).Table("report_analyses").Where("id = ? AND status = ?", row.ID, entity.ReportAnalysisStatusQueued).Updates(map[string]any{"status": entity.ReportAnalysisStatusRunning, "lease_token": token, "lease_until": leaseUntil, "deadline_at": deadline, "started_at": now.UTC(), "updated_at": now.UTC()}).Error; err != nil {
			return err
		}
		row.Status, row.LeaseToken, row.LeaseUntil, row.DeadlineAt, row.StartedAt = entity.ReportAnalysisStatusRunning, token, &leaseUntil, &deadline, &now
		var err error
		result, err = decodeReportAnalysis(row)
		return err
	})
	return result, err
}

func (r *ReportAnalysisPostgresRepository) RenewLease(ctx context.Context, analysisID, leaseToken string, now time.Time, lease time.Duration) error {
	result := r.db.WithContext(ctx).Table("report_analyses").Where("id = ? AND status = ? AND lease_token = ? AND (deadline_at IS NULL OR deadline_at > ?)", analysisID, entity.ReportAnalysisStatusRunning, leaseToken, now.UTC()).Updates(map[string]any{"lease_until": now.UTC().Add(lease), "updated_at": now.UTC()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return contract.ErrAnalysisLeaseLost
	}
	return nil
}

func (r *ReportAnalysisPostgresRepository) RecoverStale(ctx context.Context, now time.Time, limit int) error {
	if limit <= 0 {
		return nil
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var ids []string
		query := tx.WithContext(ctx).Table("report_analyses").Select("id").Where("status = ? AND (lease_until <= ? OR deadline_at <= ?)", entity.ReportAnalysisStatusRunning, now.UTC(), now.UTC()).Order("updated_at ASC").Limit(limit).Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).Scan(&ids)
		if query.Error != nil || len(ids) == 0 {
			return query.Error
		}
		return tx.WithContext(ctx).Table("report_analyses").Where("id IN ?", ids).Updates(map[string]any{"status": entity.ReportAnalysisStatusFailed, "error_code": "worker_interrupted", "updated_at": now.UTC(), "finished_at": now.UTC(), "lease_token": "", "lease_until": nil}).Error
	}, &sql.TxOptions{})
}

func (r *ReportAnalysisPostgresRepository) Complete(ctx context.Context, analysisID, leaseToken string, now time.Time, analysis *entity.ReportAnalysis) error {
	updates, err := analysisPayload(analysis)
	if err != nil {
		return err
	}
	updates["status"], updates["updated_at"], updates["finished_at"], updates["lease_token"], updates["lease_until"] = entity.ReportAnalysisStatusCompleted, now.UTC(), now.UTC(), "", nil
	result := r.db.WithContext(ctx).Table("report_analyses").Where("id = ? AND status = ? AND lease_token = ?", analysisID, entity.ReportAnalysisStatusRunning, leaseToken).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return contract.ErrAnalysisLeaseLost
	}
	return nil
}

func (r *ReportAnalysisPostgresRepository) Fail(ctx context.Context, analysisID, leaseToken string, now time.Time, code string) error {
	result := r.db.WithContext(ctx).Table("report_analyses").Where("id = ? AND status = ? AND lease_token = ?", analysisID, entity.ReportAnalysisStatusRunning, leaseToken).Updates(map[string]any{"status": entity.ReportAnalysisStatusFailed, "error_code": code, "updated_at": now.UTC(), "finished_at": now.UTC(), "lease_token": "", "lease_until": nil})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return contract.ErrAnalysisLeaseLost
	}
	return nil
}

func (r *ReportAnalysisPostgresRepository) DeleteExpired(ctx context.Context, now time.Time, limit int) (int, error) {
	if limit <= 0 {
		return 0, nil
	}
	var ids []string
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		query := tx.WithContext(ctx).Table("report_analyses").Select("id").Where("created_at <= ? AND status NOT IN (?, ?)", now.UTC().Add(-entity.ReportSnapshotRetention), entity.ReportAnalysisStatusQueued, entity.ReportAnalysisStatusRunning).Order("created_at ASC").Limit(limit).Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).Scan(&ids)
		if query.Error != nil || len(ids) == 0 {
			return query.Error
		}
		return tx.WithContext(ctx).Table("report_analyses").Where("id IN ?", ids).Delete(&entity.ReportAnalysis{}).Error
	}, &sql.TxOptions{})
	return len(ids), err
}

func insertReportAnalysis(tx *gorm.DB, analysis *entity.ReportAnalysis) error {
	payload, err := analysisPayload(analysis)
	if err != nil {
		return err
	}
	return tx.WithContext(context.Background()).Exec(`INSERT INTO report_analyses(id, owner_id, snapshot_id, job_key, client_request_id, payload_hash, analysis_fingerprint, revision, rule_version, prompt_version, provider_profile_version, status, error_code, provider_called, summary, observations, suggestions, fact_refs, usage, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, analysis.ID, analysis.OwnerID, analysis.SnapshotID, analysis.JobKey, analysis.ClientRequestID, analysis.PayloadHash, analysis.AnalysisFingerprint, analysis.Revision, analysis.RuleVersion, analysis.PromptVersion, analysis.ProviderProfileVersion, analysis.Status, analysis.ErrorCode, analysis.ProviderCalled, analysis.Summary, payload["observations"], payload["suggestions"], payload["fact_refs"], payload["usage"], analysis.CreatedAt.UTC(), analysis.UpdatedAt.UTC()).Error
}

func analysisPayload(analysis *entity.ReportAnalysis) (map[string]any, error) {
	observations, err := json.Marshal(analysis.Observations)
	if err != nil {
		return nil, err
	}
	suggestions, err := json.Marshal(analysis.Suggestions)
	if err != nil {
		return nil, err
	}
	factRefs, err := json.Marshal(analysis.FactRefs)
	if err != nil {
		return nil, err
	}
	usage, err := json.Marshal(analysis.Usage)
	if err != nil {
		return nil, err
	}
	return map[string]any{"observations": observations, "suggestions": suggestions, "fact_refs": factRefs, "usage": usage, "provider_called": analysis.ProviderCalled, "summary": analysis.Summary, "error_code": analysis.ErrorCode}, nil
}

func decodeReportAnalysis(row reportAnalysisRow) (*entity.ReportAnalysis, error) {
	result := &entity.ReportAnalysis{ID: row.ID, OwnerID: row.OwnerID, SnapshotID: row.SnapshotID, JobKey: row.JobKey, ClientRequestID: row.ClientRequestID, PayloadHash: row.PayloadHash, AnalysisFingerprint: row.AnalysisFingerprint, Revision: row.Revision, RuleVersion: row.RuleVersion, PromptVersion: row.PromptVersion, ProviderProfileVersion: row.ProviderProfileVersion, Status: row.Status, ErrorCode: row.ErrorCode, ProviderCalled: row.ProviderCalled, Summary: row.Summary, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt, StartedAt: row.StartedAt, FinishedAt: row.FinishedAt, LeaseUntil: row.LeaseUntil, LeaseToken: row.LeaseToken, DeadlineAt: row.DeadlineAt, RetainedUntil: row.CreatedAt.Add(entity.ReportSnapshotRetention)}
	if len(row.Observations) > 0 {
		if err := json.Unmarshal(row.Observations, &result.Observations); err != nil {
			return nil, err
		}
	}
	if len(row.Suggestions) > 0 {
		if err := json.Unmarshal(row.Suggestions, &result.Suggestions); err != nil {
			return nil, err
		}
	}
	if len(row.FactRefs) > 0 {
		if err := json.Unmarshal(row.FactRefs, &result.FactRefs); err != nil {
			return nil, err
		}
	}
	if len(row.Usage) > 0 {
		if err := json.Unmarshal(row.Usage, &result.Usage); err != nil {
			return nil, err
		}
	}
	return result, nil
}
