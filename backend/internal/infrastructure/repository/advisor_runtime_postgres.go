package repository

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	contract "github.com/mypocket/backend/internal/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	advisorRuntimeLease = 45 * time.Second
	advisorRuntimeLimit = 100
)

func (r *AdvisorPostgresRepository) LoadAdvisorContextState(ctx context.Context, ownerID, conversationID string, generation int64) (entity.AdvisorContextState, error) {
	var conversation entity.AdvisorConversation
	err := r.db.WithContext(ctx).Where("id = ? AND owner_id = ? AND generation = ?", conversationID, ownerID, generation).First(&conversation).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return entity.AdvisorContextState{}, contract.ErrNotFound
	}
	if err != nil {
		return entity.AdvisorContextState{}, err
	}
	if expiry := conversation.IdleExpiresAt(); expiry != nil && !expiry.After(time.Now().UTC()) {
		return entity.AdvisorContextState{}, contract.ErrAdvisorPurged
	}
	payload, err := json.Marshal(conversation.Summary)
	if err != nil || len(payload) > 4096 {
		return entity.AdvisorContextState{}, entity.ErrAdvisorRuntimePayload
	}
	var state entity.AdvisorContextState
	if err := json.Unmarshal(payload, &state); err != nil {
		return entity.AdvisorContextState{}, entity.ErrAdvisorRuntimePayload
	}
	if state.Version != 1 || state.Generation != generation {
		return entity.AdvisorContextState{Version: 1, Generation: generation}, nil
	}
	return state, nil
}

// ClaimRun atomically changes one queued run to running. Conversation then run
// is the lock order shared by completion/reset/reaper; an already-running or
// terminal run can never be stolen by a second executor.
func (r *AdvisorPostgresRepository) ClaimRun(ctx context.Context, ownerID, runID, token string) (entity.AdvisorRunLease, error) {
	if r == nil || r.db == nil || strings.TrimSpace(ownerID) == "" || strings.TrimSpace(runID) == "" || strings.TrimSpace(token) == "" {
		return entity.AdvisorRunLease{}, contract.ErrAdvisorLeaseLost
	}
	var lease entity.AdvisorRunLease
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var snapshot entity.AdvisorRun
		if err := tx.Where("id = ? AND owner_id = ?", runID, ownerID).First(&snapshot).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return contract.ErrNotFound
			}
			return err
		}
		var conversation entity.AdvisorConversation
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND owner_id = ?", snapshot.ConversationID, ownerID).First(&conversation).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return contract.ErrNotFound
			}
			return err
		}
		now, err := advisorDBNow(tx)
		if err != nil {
			return err
		}
		if err := r.expireIdleConversation(tx, &conversation, now); err != nil {
			return err
		}
		var run entity.AdvisorRun
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND owner_id = ?", runID, ownerID).First(&run).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return contract.ErrNotFound
			}
			return err
		}
		if run.Generation != conversation.Generation || run.Status != entity.AdvisorStatusQueued {
			return contract.ErrAdvisorLeaseLost
		}
		deadline := run.DeadlineAt
		if deadline == nil || !deadline.After(now) {
			if err := tx.Model(&entity.AdvisorRun{}).Where("id = ? AND owner_id = ?", runID, ownerID).Updates(map[string]any{
				"status": entity.AdvisorStatusFailed, "error_code": "deadline_exceeded", "finished_at": now, "lease_token": nil, "lease_until": nil,
			}).Error; err != nil {
				return err
			}
			return contract.ErrAdvisorRuntimeExpired
		}
		until := now.Add(advisorRuntimeLease)
		if until.After(*deadline) {
			until = *deadline
		}
		if err := tx.Model(&entity.AdvisorRun{}).Where("id = ? AND owner_id = ? AND status = ?", runID, ownerID, entity.AdvisorStatusQueued).Updates(map[string]any{
			"status": entity.AdvisorStatusRunning, "lease_token": token, "lease_until": until,
		}).Error; err != nil {
			return err
		}
		lease = entity.AdvisorRunLease{OwnerID: ownerID, RunID: runID, Generation: run.Generation, Token: token, DeadlineAt: *deadline, LeaseUntil: until}
		return nil
	})
	return lease, err
}

// HeartbeatRun is a run-only fenced update. It deliberately does not acquire
// the conversation lock, preventing the heartbeat path from creating a lock
// inversion with completion/reset.
func (r *AdvisorPostgresRepository) HeartbeatRun(ctx context.Context, lease entity.AdvisorRunLease) error {
	if r == nil || r.db == nil || strings.TrimSpace(lease.OwnerID) == "" || strings.TrimSpace(lease.RunID) == "" || strings.TrimSpace(lease.Token) == "" || lease.Generation <= 0 {
		return contract.ErrAdvisorLeaseLost
	}
	result := r.db.WithContext(ctx).Model(&entity.AdvisorRun{}).
		Where("id = ? AND owner_id = ? AND generation = ? AND lease_token = ? AND status = ? AND lease_until > NOW() AND deadline_at > NOW()", lease.RunID, lease.OwnerID, lease.Generation, lease.Token, entity.AdvisorStatusRunning).
		Updates(map[string]any{"lease_until": gorm.Expr("LEAST(NOW() + interval '45 seconds', deadline_at)")})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return contract.ErrAdvisorLeaseLost
	}
	return nil
}

// CompleteRun commits assistant message, optional evidence, completion event,
// usage, compact context and terminal status in one transaction. A completed
// run is a durable receipt: retrying with the same lease returns it without
// inserting another assistant message.
func (r *AdvisorPostgresRepository) CompleteRun(ctx context.Context, lease entity.AdvisorRunLease, completion entity.AdvisorCompletion) (entity.AdvisorCompletionResult, error) {
	if r == nil || r.db == nil || strings.TrimSpace(lease.OwnerID) == "" || strings.TrimSpace(lease.RunID) == "" || strings.TrimSpace(lease.Token) == "" || lease.Generation <= 0 {
		return entity.AdvisorCompletionResult{}, contract.ErrAdvisorLeaseLost
	}
	message := &entity.AdvisorMessage{ID: uuid.NewString(), Role: entity.AdvisorRoleAssistant, PartsVersion: 1, Parts: completion.Parts}
	if err := message.Validate(); err != nil {
		return entity.AdvisorCompletionResult{}, err
	}
	if err := entity.ValidateAdvisorContextState(completion.ContextState); err != nil {
		return entity.AdvisorCompletionResult{}, err
	}
	usage := completion.Usage
	if usage == nil {
		usage = map[string]any{}
	}
	var result entity.AdvisorCompletionResult
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var snapshot entity.AdvisorRun
		if err := tx.Where("id = ? AND owner_id = ?", lease.RunID, lease.OwnerID).First(&snapshot).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return contract.ErrNotFound
			}
			return err
		}
		var conversation entity.AdvisorConversation
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND owner_id = ?", snapshot.ConversationID, lease.OwnerID).First(&conversation).Error; err != nil {
			return err
		}
		now, err := advisorDBNow(tx)
		if err != nil {
			return err
		}
		if err := r.expireIdleConversation(tx, &conversation, now); err != nil {
			return err
		}
		var run entity.AdvisorRun
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND owner_id = ?", lease.RunID, lease.OwnerID).First(&run).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return contract.ErrNotFound
			}
			return err
		}
		if run.Status == entity.AdvisorStatusCompleted {
			var existing entity.AdvisorMessage
			if err := tx.Where("run_id = ? AND role = ?", lease.RunID, entity.AdvisorRoleAssistant).First(&existing).Error; err != nil {
				return contract.ErrAdvisorLeaseLost
			}
			result = entity.AdvisorCompletionResult{Message: &existing, Run: &run, Replayed: true}
			return nil
		}
		if run.Generation != lease.Generation || run.Generation != conversation.Generation || run.LeaseToken != lease.Token || run.Status != entity.AdvisorStatusRunning || run.LeaseUntil == nil || !run.LeaseUntil.After(now) || run.DeadlineAt == nil || !run.DeadlineAt.After(now) {
			return contract.ErrAdvisorLeaseLost
		}
		message.ConversationID = conversation.ID
		message.Generation = run.Generation
		message.RunID = run.ID
		message.Seq = conversation.NextMessageSeq
		if err := tx.Create(message).Error; err != nil {
			return err
		}
		if completion.FactBundle != nil {
			bundle := *completion.FactBundle
			if bundle.ID == "" {
				bundle.ID = uuid.NewString()
			}
			bundle.OwnerID, bundle.RunID = lease.OwnerID, lease.RunID
			if bundle.Scope == nil {
				bundle.Scope = map[string]any{}
			}
			if bundle.Facts == nil {
				bundle.Facts = []entity.Fact{}
			}
			bundle.CreatedAt = now
			if err := tx.Create(&bundle).Error; err != nil {
				return err
			}
		}
		event := entity.AdvisorEvent{RunID: lease.RunID, Seq: run.LastEventSeq + 1, Type: "run.completed", Payload: map[string]any{"message_id": message.ID}, CreatedAt: now}
		if err := tx.Create(&event).Error; err != nil {
			return err
		}
		updates := map[string]any{"status": entity.AdvisorStatusCompleted, "finished_at": now, "lease_token": nil, "lease_until": nil, "last_event_seq": event.Seq, "model_usage": usage}
		if err := tx.Model(&entity.AdvisorRun{}).Where("id = ? AND owner_id = ?", lease.RunID, lease.OwnerID).Updates(updates).Error; err != nil {
			return err
		}
		conversation.NextMessageSeq++
		conversation.UpdatedAt = now
		conversationUpdates := map[string]any{"next_message_seq": conversation.NextMessageSeq, "updated_at": now}
		if completion.ContextState != nil {
			completion.ContextState["through_seq"] = message.Seq
			conversation.Summary = completion.ContextState
			conversation.SummaryVersion++
			conversation.SummaryThroughSeq = message.Seq
			conversationUpdates["summary"] = completion.ContextState
			conversationUpdates["summary_version"] = conversation.SummaryVersion
			conversationUpdates["summary_through_seq"] = conversation.SummaryThroughSeq
		}
		if err := tx.Model(&entity.AdvisorConversation{}).Where("id = ? AND owner_id = ?", conversation.ID, lease.OwnerID).Updates(conversationUpdates).Error; err != nil {
			return err
		}
		run.Status = entity.AdvisorStatusCompleted
		run.FinishedAt = &now
		run.LeaseToken, run.LeaseUntil = "", nil
		run.LastEventSeq = event.Seq
		run.ModelUsage = usage
		result = entity.AdvisorCompletionResult{Message: message, Run: &run}
		return nil
	})
	return result, err
}

// FinishRun records an executor-fenced failure/interrupt without allowing a
// stale worker to terminally overwrite a newer owner action.
func (r *AdvisorPostgresRepository) FinishRun(ctx context.Context, lease entity.AdvisorRunLease, status, errorCode string, usage map[string]any) error {
	if status != entity.AdvisorStatusFailed && status != entity.AdvisorStatusInterrupted {
		return contract.ErrAdvisorLeaseLost
	}
	if r == nil || r.db == nil || strings.TrimSpace(lease.OwnerID) == "" || strings.TrimSpace(lease.RunID) == "" || strings.TrimSpace(lease.Token) == "" {
		return contract.ErrAdvisorLeaseLost
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var run entity.AdvisorRun
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND owner_id = ?", lease.RunID, lease.OwnerID).First(&run).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return contract.ErrNotFound
			}
			return err
		}
		if run.Generation != lease.Generation || run.LeaseToken != lease.Token || run.Status != entity.AdvisorStatusRunning {
			return contract.ErrAdvisorLeaseLost
		}
		now, err := advisorDBNow(tx)
		if err != nil {
			return err
		}
		event := entity.AdvisorEvent{RunID: run.ID, Seq: run.LastEventSeq + 1, Type: "run." + status, Payload: map[string]any{"error_code": errorCode}, CreatedAt: now}
		if err := tx.Create(&event).Error; err != nil {
			return err
		}
		if usage == nil {
			usage = map[string]any{}
		}
		return tx.Model(&entity.AdvisorRun{}).Where("id = ? AND owner_id = ?", run.ID, lease.OwnerID).Updates(map[string]any{"status": status, "error_code": errorCode, "finished_at": now, "lease_token": nil, "lease_until": nil, "last_event_seq": event.Seq, "model_usage": usage}).Error
	})
}

// ReapExpiredRuns terminally closes queued/running runs whose deadline or
// executor lease elapsed. Candidate selection is non-locking; each candidate
// is rechecked under conversation -> run locks to avoid lock inversion.
func (r *AdvisorPostgresRepository) ReapExpiredRuns(ctx context.Context, limit int) (int, error) {
	if r == nil || r.db == nil {
		return 0, contract.ErrNotFound
	}
	if limit <= 0 || limit > advisorRuntimeLimit {
		limit = advisorRuntimeLimit
	}
	type candidate struct{ ID, OwnerID, ConversationID string }
	var candidates []candidate
	if err := r.db.WithContext(ctx).Table("advisor_runs").Select("id, owner_id, conversation_id").Where("status IN ? AND (deadline_at <= NOW() OR (status = ? AND lease_until <= NOW()))", []string{entity.AdvisorStatusQueued, entity.AdvisorStatusRunning}, entity.AdvisorStatusRunning).Order("created_at ASC").Limit(limit).Find(&candidates).Error; err != nil {
		return 0, err
	}
	count := 0
	for _, candidate := range candidates {
		updated, err := r.reapOne(ctx, candidate)
		if err != nil {
			return count, err
		}
		if updated {
			count++
		}
	}
	return count, nil
}

func (r *AdvisorPostgresRepository) reapOne(ctx context.Context, candidate struct{ ID, OwnerID, ConversationID string }) (bool, error) {
	updated := false
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var conversation entity.AdvisorConversation
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND owner_id = ?", candidate.ConversationID, candidate.OwnerID).First(&conversation).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
		var run entity.AdvisorRun
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND owner_id = ?", candidate.ID, candidate.OwnerID).First(&run).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
		now, err := advisorDBNow(tx)
		if err != nil {
			return err
		}
		if run.Status != entity.AdvisorStatusQueued && run.Status != entity.AdvisorStatusRunning {
			return nil
		}
		status, code := entity.AdvisorStatusInterrupted, "executor_lost"
		if run.DeadlineAt == nil || !run.DeadlineAt.After(now) {
			status, code = entity.AdvisorStatusFailed, "deadline_exceeded"
		} else if run.Status == entity.AdvisorStatusQueued {
			return nil
		} else if run.LeaseUntil != nil && run.LeaseUntil.After(now) {
			return nil
		}
		event := entity.AdvisorEvent{RunID: run.ID, Seq: run.LastEventSeq + 1, Type: "run." + status, Payload: map[string]any{"error_code": code}, CreatedAt: now}
		if err := tx.Create(&event).Error; err != nil {
			return err
		}
		if err := tx.Model(&entity.AdvisorRun{}).Where("id = ? AND owner_id = ? AND status IN ?", run.ID, candidate.OwnerID, []string{entity.AdvisorStatusQueued, entity.AdvisorStatusRunning}).Updates(map[string]any{"status": status, "error_code": code, "finished_at": now, "lease_token": nil, "lease_until": nil, "last_event_seq": event.Seq}).Error; err != nil {
			return err
		}
		updated = true
		return nil
	})
	return updated, err
}

func advisorDBNow(tx *gorm.DB) (time.Time, error) {
	var now time.Time
	if err := tx.Raw("SELECT NOW()").Scan(&now).Error; err != nil {
		return time.Time{}, err
	}
	return now.UTC(), nil
}
