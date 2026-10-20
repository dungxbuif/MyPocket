package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	contract "github.com/mypocket/backend/internal/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AdvisorPostgresRepository struct{ db *gorm.DB }

func NewAdvisorPostgresRepository(db *gorm.DB) contract.AdvisorRepository {
	return &AdvisorPostgresRepository{db: db}
}

func (r *AdvisorPostgresRepository) StartRun(ctx context.Context, input contract.AdvisorRunInput) (*entity.AdvisorRun, *entity.AdvisorMessage, bool, error) {
	if r == nil || r.db == nil || strings.TrimSpace(input.OwnerID) == "" || strings.TrimSpace(input.RequestID) == "" || strings.TrimSpace(input.PayloadHash) == "" {
		return nil, nil, false, entity.ErrAdvisorMessageInvalid
	}
	message := &entity.AdvisorMessage{ID: uuid.NewString(), Role: entity.AdvisorRoleUser, PartsVersion: 1, Parts: input.Parts}
	if err := message.Validate(); err != nil {
		return nil, nil, false, err
	}
	var run entity.AdvisorRun
	var replayMessage entity.AdvisorMessage
	replayed := false
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		conversation, err := r.lockConversation(tx, input.OwnerID)
		if err != nil {
			return err
		}
		now := time.Now().UTC()
		if err := r.expireIdleConversation(tx, conversation, now); err != nil {
			return err
		}
		var existing entity.AdvisorRun
		if err := tx.Where("owner_id = ? AND client_request_id = ?", input.OwnerID, input.RequestID).First(&existing).Error; err == nil {
			if existing.PayloadHash != input.PayloadHash {
				return contract.ErrAdvisorRequestConflict
			}
			if existing.Status == entity.AdvisorStatusPurged {
				return contract.ErrAdvisorPurged
			}
			if err := tx.Where("run_id = ? AND role = ?", existing.ID, entity.AdvisorRoleUser).First(&replayMessage).Error; err != nil {
				return err
			}
			run = existing
			replayed = true
			return nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		var active entity.AdvisorRun
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("conversation_id = ? AND status IN ?", conversation.ID, []string{entity.AdvisorStatusQueued, entity.AdvisorStatusRunning}).First(&active).Error; err == nil {
			return contract.ErrAdvisorBusy
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		deadline := now.Add(240 * time.Second)
		run = entity.AdvisorRun{ID: uuid.NewString(), OwnerID: input.OwnerID, ConversationID: conversation.ID, Generation: conversation.Generation, ClientRequestID: input.RequestID, PayloadHash: input.PayloadHash, CredentialKind: input.CredentialKind, CredentialID: input.CredentialID, CredentialExpiresAt: input.ExpiresAt, DeadlineAt: &deadline, Status: entity.AdvisorStatusQueued, ModelUsage: map[string]any{}, CreatedAt: now}
		if err := tx.Create(&run).Error; err != nil {
			return err
		}
		message.ConversationID = conversation.ID
		message.Generation = conversation.Generation
		message.RunID = run.ID
		message.Seq = conversation.NextMessageSeq
		if err := tx.Create(message).Error; err != nil {
			return err
		}
		conversation.NextMessageSeq++
		return tx.Model(&entity.AdvisorConversation{}).Where("id = ?", conversation.ID).Updates(map[string]any{"next_message_seq": conversation.NextMessageSeq, "last_user_message_at": now, "updated_at": now}).Error
	})
	if err != nil {
		return nil, nil, false, err
	}
	if replayed {
		return &run, &replayMessage, true, nil
	}
	return &run, message, false, nil
}

// ResetConversation clears only advisor content and rotates the generation.
// The reset ledger is written in the same transaction so an uncertain client
// retry is replayed without deleting a newer question.
func (r *AdvisorPostgresRepository) ResetConversation(ctx context.Context, ownerID, requestID, payloadHash string) (*entity.AdvisorResetResult, error) {
	if r == nil || r.db == nil || strings.TrimSpace(ownerID) == "" || strings.TrimSpace(requestID) == "" || strings.TrimSpace(payloadHash) == "" {
		return nil, entity.ErrAdvisorMessageInvalid
	}
	var result entity.AdvisorResetResult
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing entity.AdvisorResetRequest
		if err := tx.Where("owner_id = ? AND client_request_id = ?", ownerID, requestID).First(&existing).Error; err == nil {
			if existing.PayloadHash != payloadHash {
				return contract.ErrAdvisorRequestConflict
			}
			result = entity.AdvisorResetResult{ConversationID: existing.ConversationID, Generation: existing.Generation, ServerTime: existing.CreatedAt.UTC(), Replayed: true}
			return nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		conversation, err := r.lockConversation(tx, ownerID)
		if err != nil {
			return err
		}
		// A concurrent reset with the same key may have committed while this
		// transaction waited for the conversation lock; replay it instead of
		// rotating the generation twice.
		if err := tx.Where("owner_id = ? AND client_request_id = ?", ownerID, requestID).First(&existing).Error; err == nil {
			if existing.PayloadHash != payloadHash {
				return contract.ErrAdvisorRequestConflict
			}
			result = entity.AdvisorResetResult{ConversationID: existing.ConversationID, Generation: existing.Generation, ServerTime: existing.CreatedAt.UTC(), Replayed: true}
			return nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		now := time.Now().UTC()
		var activeRunIDs []string
		if err := tx.Model(&entity.AdvisorRun{}).
			Where("conversation_id = ? AND status IN ?", conversation.ID, []string{entity.AdvisorStatusQueued, entity.AdvisorStatusRunning}).
			Pluck("id", &activeRunIDs).Error; err != nil {
			return err
		}
		if err := r.resetConversation(tx, conversation, now); err != nil {
			return err
		}
		request := entity.AdvisorResetRequest{
			OwnerID: ownerID, ClientRequestID: requestID, PayloadHash: payloadHash,
			ConversationID: conversation.ID, Generation: conversation.Generation, CreatedAt: now,
		}
		if err := tx.Create(&request).Error; err != nil {
			return err
		}
		result = entity.AdvisorResetResult{ConversationID: conversation.ID, Generation: conversation.Generation, ServerTime: now, Replayed: false, PurgedRunIDs: activeRunIDs}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *AdvisorPostgresRepository) lockConversation(tx *gorm.DB, ownerID string) (*entity.AdvisorConversation, error) {
	var conversation entity.AdvisorConversation
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("owner_id = ?", ownerID).First(&conversation).Error
	if err == nil {
		return &conversation, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	now := time.Now().UTC()
	conversation = entity.AdvisorConversation{ID: uuid.NewString(), OwnerID: ownerID, Generation: 1, NextMessageSeq: 1, Summary: map[string]any{}, CreatedAt: now, UpdatedAt: now}
	if err := tx.Create(&conversation).Error; err != nil {
		if !isUniqueViolation(err) {
			return nil, err
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("owner_id = ?", ownerID).First(&conversation).Error; err != nil {
			return nil, err
		}
	}
	return &conversation, nil
}

func (r *AdvisorPostgresRepository) MarkRunStatus(ctx context.Context, ownerID, id, status string, now time.Time) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var run entity.AdvisorRun
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND owner_id = ?", id, ownerID).First(&run).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return contract.ErrNotFound
			}
			return err
		}
		if !entity.CanTransitionAdvisorRun(run.Status, status) {
			return entity.ErrAdvisorRunTransition
		}
		updates := map[string]any{"status": status}
		if status == entity.AdvisorStatusCompleted || status == entity.AdvisorStatusFailed || status == entity.AdvisorStatusCancelled || status == entity.AdvisorStatusInterrupted || status == entity.AdvisorStatusPurged {
			updates["finished_at"] = now.UTC()
		}
		return tx.Model(&entity.AdvisorRun{}).Where("id = ? AND owner_id = ?", id, ownerID).Updates(updates).Error
	})
}

func (r *AdvisorPostgresRepository) StoreRunUsage(ctx context.Context, ownerID, id string, usage map[string]any) error {
	if usage == nil {
		return nil
	}
	return r.db.WithContext(ctx).Model(&entity.AdvisorRun{}).Where("id = ? AND owner_id = ?", id, ownerID).Update("model_usage", usage).Error
}

func (r *AdvisorPostgresRepository) GetRun(ctx context.Context, ownerID, id string) (*entity.AdvisorRun, error) {
	var run entity.AdvisorRun
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ? AND owner_id = ?", id, ownerID).First(&run).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return contract.ErrNotFound
			}
			return err
		}
		var conversation entity.AdvisorConversation
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND owner_id = ?", run.ConversationID, ownerID).First(&conversation).Error; err != nil {
			return err
		}
		if err := r.expireIdleConversation(tx, &conversation, time.Now().UTC()); err != nil {
			return err
		}
		return tx.Where("id = ? AND owner_id = ?", id, ownerID).First(&run).Error
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, contract.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &run, nil
}

// GetRunStatus is intentionally a short, non-locking owner-scoped probe. It
// lets another API process observe cancellation without taking the
// conversation/run locks used by lifecycle writes.
func (r *AdvisorPostgresRepository) GetRunStatus(ctx context.Context, ownerID, id string) (string, error) {
	if r == nil || r.db == nil || strings.TrimSpace(ownerID) == "" || strings.TrimSpace(id) == "" {
		return "", contract.ErrNotFound
	}
	var run entity.AdvisorRun
	if err := r.db.WithContext(ctx).Select("status").Where("id = ? AND owner_id = ?", id, ownerID).First(&run).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", contract.ErrNotFound
		}
		return "", err
	}
	return run.Status, nil
}

func (r *AdvisorPostgresRepository) GetRunByRequest(ctx context.Context, ownerID, requestID string) (*entity.AdvisorRun, error) {
	var run entity.AdvisorRun
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("owner_id = ? AND client_request_id = ?", ownerID, requestID).First(&run).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return contract.ErrNotFound
			}
			return err
		}
		var conversation entity.AdvisorConversation
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND owner_id = ?", run.ConversationID, ownerID).First(&conversation).Error; err != nil {
			return err
		}
		if err := r.expireIdleConversation(tx, &conversation, time.Now().UTC()); err != nil {
			return err
		}
		return tx.Where("id = ? AND owner_id = ?", run.ID, ownerID).First(&run).Error
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, contract.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &run, nil
}

func (r *AdvisorPostgresRepository) ListMessages(ctx context.Context, ownerID, conversationID string, beforeSeq int64, limit int) ([]entity.AdvisorMessage, error) {
	if limit <= 0 || limit > 50 {
		limit = 50
	}
	var messages []entity.AdvisorMessage
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var conversation entity.AdvisorConversation
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND owner_id = ?", conversationID, ownerID).First(&conversation).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
		if err := r.expireIdleConversation(tx, &conversation, time.Now().UTC()); err != nil {
			return err
		}
		query := tx.Where("conversation_id = ? AND generation = ?", conversationID, conversation.Generation)
		if beforeSeq > 0 {
			query = query.Where("seq < ?", beforeSeq)
		}
		return query.Order("seq DESC").Limit(limit).Find(&messages).Error
	})
	if err != nil {
		return nil, err
	}
	return messages, nil
}

// AdvisorHistoryMeta reads the same owner-scoped conversation boundary as
// ListMessages. Expiry is enforced before the metadata is returned so a read
// can never extend or resurrect an idle generation.
func (r *AdvisorPostgresRepository) AdvisorHistoryMeta(ctx context.Context, ownerID, conversationID string) (entity.AdvisorHistoryMeta, error) {
	if r == nil || r.db == nil || strings.TrimSpace(ownerID) == "" || strings.TrimSpace(conversationID) == "" {
		return entity.AdvisorHistoryMeta{}, contract.ErrNotFound
	}
	var conversation entity.AdvisorConversation
	var activeRun *entity.AdvisorRun
	var serverTime = time.Now().UTC()
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND owner_id = ?", conversationID, ownerID).First(&conversation).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return contract.ErrNotFound
			}
			return err
		}
		if err := r.expireIdleConversation(tx, &conversation, serverTime); err != nil {
			return err
		}
		var candidate entity.AdvisorRun
		if err := tx.Where("owner_id = ? AND conversation_id = ? AND status IN ?", ownerID, conversationID, []string{entity.AdvisorStatusQueued, entity.AdvisorStatusRunning}).Order("created_at DESC").First(&candidate).Error; err == nil {
			activeRun = &candidate
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		return nil
	})
	if err != nil {
		return entity.AdvisorHistoryMeta{}, err
	}
	meta := conversation.HistoryMeta(serverTime)
	if activeRun != nil {
		deadline := activeRun.DeadlineAt
		if deadline == nil {
			value := activeRun.CreatedAt.Add(240 * time.Second)
			deadline = &value
		}
		meta.ActiveRun = &entity.AdvisorActiveRun{ID: activeRun.ID, Status: activeRun.Status, CreatedAt: activeRun.CreatedAt, DeadlineAt: deadline}
	}
	return meta, nil
}

func (r *AdvisorPostgresRepository) AppendEvent(ctx context.Context, ownerID, runID, eventType string, payload any) (*entity.AdvisorEvent, error) {
	var event entity.AdvisorEvent
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var run entity.AdvisorRun
		if err := tx.Where("id = ? AND owner_id = ?", runID, ownerID).First(&run).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return contract.ErrNotFound
			}
			return err
		}
		var conversation entity.AdvisorConversation
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND owner_id = ?", run.ConversationID, ownerID).First(&conversation).Error; err != nil {
			return err
		}
		if err := r.expireIdleConversation(tx, &conversation, time.Now().UTC()); err != nil {
			return err
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND owner_id = ?", runID, ownerID).First(&run).Error; err != nil {
			return err
		}
		if run.Generation != conversation.Generation {
			return contract.ErrAdvisorLeaseLost
		}
		if run.Status != entity.AdvisorStatusRunning && run.Status != entity.AdvisorStatusQueued {
			return contract.ErrAdvisorLeaseLost
		}
		seq := run.LastEventSeq + 1
		event = entity.AdvisorEvent{RunID: runID, Seq: seq, Type: eventType, Payload: payload, CreatedAt: time.Now().UTC()}
		if err := tx.Create(&event).Error; err != nil {
			return err
		}
		return tx.Model(&entity.AdvisorRun{}).Where("id = ?", runID).Update("last_event_seq", seq).Error
	})
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// ListEvents returns a bounded, owner-scoped replay window. The run lookup is
// deliberately separate from the event query so a foreign run is
// indistinguishable from a missing run to callers.
func (r *AdvisorPostgresRepository) ListEvents(ctx context.Context, ownerID, runID string, afterSeq, limit int) ([]entity.AdvisorEvent, error) {
	if r == nil || r.db == nil || strings.TrimSpace(ownerID) == "" || strings.TrimSpace(runID) == "" || afterSeq < 0 {
		return nil, contract.ErrNotFound
	}
	var run entity.AdvisorRun
	if err := r.db.WithContext(ctx).Where("id = ? AND owner_id = ?", runID, ownerID).First(&run).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, contract.ErrNotFound
		}
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	var events []entity.AdvisorEvent
	if err := r.db.WithContext(ctx).Where("run_id = ? AND seq > ?", runID, afterSeq).Order("seq ASC").Limit(limit).Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

func (r *AdvisorPostgresRepository) CreateAssistantMessage(ctx context.Context, ownerID, runID string, generation int64, parts []entity.AdvisorPart) (*entity.AdvisorMessage, error) {
	message := &entity.AdvisorMessage{ID: uuid.NewString(), Role: entity.AdvisorRoleAssistant, PartsVersion: 1, Parts: parts}
	if err := message.Validate(); err != nil {
		return nil, err
	}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var run entity.AdvisorRun
		if err := tx.Where("id = ? AND owner_id = ?", runID, ownerID).First(&run).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return contract.ErrNotFound
			}
			return err
		}
		if run.Generation != generation {
			return contract.ErrAdvisorLeaseLost
		}
		if run.Status != entity.AdvisorStatusQueued && run.Status != entity.AdvisorStatusRunning {
			return contract.ErrAdvisorLeaseLost
		}
		var conversation entity.AdvisorConversation
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND owner_id = ?", run.ConversationID, ownerID).First(&conversation).Error; err != nil {
			return err
		}
		if err := r.expireIdleConversation(tx, &conversation, time.Now().UTC()); err != nil {
			return err
		}
		// Match StartRun's conversation -> run lock order. Cancellation may have
		// committed while this writer waited for the conversation lock.
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND owner_id = ?", runID, ownerID).First(&run).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return contract.ErrNotFound
			}
			return err
		}
		if run.Generation != generation || run.Generation != conversation.Generation || (run.Status != entity.AdvisorStatusQueued && run.Status != entity.AdvisorStatusRunning) {
			return contract.ErrAdvisorLeaseLost
		}
		message.ConversationID = conversation.ID
		message.Generation = generation
		message.RunID = runID
		message.Seq = conversation.NextMessageSeq
		if err := tx.Create(message).Error; err != nil {
			return err
		}
		return tx.Model(&entity.AdvisorConversation{}).Where("id = ?", conversation.ID).Updates(map[string]any{"next_message_seq": conversation.NextMessageSeq + 1, "updated_at": time.Now().UTC()}).Error
	})
	if err != nil {
		return nil, err
	}
	return message, nil
}

// ExpireIdleBatch purges at most limit owner conversations whose server-side
// idle deadline has elapsed. Row locks are skipped so a busy conversation
// cannot stall other owners; a later worker pass retries it.
func (r *AdvisorPostgresRepository) ExpireIdleBatch(ctx context.Context, now time.Time, limit int) (int, error) {
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	var conversations []entity.AdvisorConversation
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("last_user_message_at IS NOT NULL AND last_user_message_at <= ?", now.UTC().Add(-entity.AdvisorIdleTTL)).
			Order("last_user_message_at ASC").Limit(limit).Find(&conversations).Error; err != nil {
			return err
		}
		for index := range conversations {
			if err := r.expireIdleConversation(tx, &conversations[index], now.UTC()); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return len(conversations), nil
}

// expireIdleConversation rotates the conversation generation and removes only
// advisor content. It intentionally keeps run rows as minimal idempotency and
// usage tombstones so an old request cannot be replayed into the fresh chat.
// The caller must already hold the conversation row lock.
func (r *AdvisorPostgresRepository) expireIdleConversation(tx *gorm.DB, conversation *entity.AdvisorConversation, now time.Time) error {
	if conversation == nil || !conversation.IsIdleExpired(now) {
		return nil
	}
	var runIDs []string
	if err := tx.Model(&entity.AdvisorRun{}).Where("conversation_id = ?", conversation.ID).Pluck("id", &runIDs).Error; err != nil {
		return err
	}
	if len(runIDs) > 0 {
		if err := tx.Model(&entity.AdvisorRun{}).Where("conversation_id = ?", conversation.ID).Updates(map[string]any{
			"status":      entity.AdvisorStatusPurged,
			"finished_at": now,
			"lease_token": nil,
			"lease_until": nil,
		}).Error; err != nil {
			return err
		}
		if err := tx.Where("run_id IN ?", runIDs).Delete(&entity.AdvisorEvent{}).Error; err != nil {
			return err
		}
		if err := tx.Where("run_id IN ?", runIDs).Delete(&entity.AdvisorFactBundle{}).Error; err != nil {
			return err
		}
	}
	if err := tx.Where("conversation_id = ?", conversation.ID).Delete(&entity.AdvisorMessage{}).Error; err != nil {
		return err
	}
	conversation.Generation++
	conversation.LastUserMessageAt = nil
	conversation.Summary = map[string]any{}
	conversation.SummaryThroughSeq = 0
	conversation.SummaryVersion = 0
	conversation.UpdatedAt = now
	return tx.Model(&entity.AdvisorConversation{}).Where("id = ?", conversation.ID).Updates(map[string]any{
		"generation":           conversation.Generation,
		"last_user_message_at": nil,
		"summary":              map[string]any{},
		"summary_through_seq":  0,
		"summary_version":      0,
		"updated_at":           now,
	}).Error
}

// resetConversation is the explicit-reset counterpart to idle expiry. The
// caller must hold the conversation row lock; all advisor runs become content-
// free tombstones and financial tables are never touched.
func (r *AdvisorPostgresRepository) resetConversation(tx *gorm.DB, conversation *entity.AdvisorConversation, now time.Time) error {
	if conversation == nil {
		return entity.ErrAdvisorMessageInvalid
	}
	var runIDs []string
	if err := tx.Model(&entity.AdvisorRun{}).Where("conversation_id = ?", conversation.ID).Pluck("id", &runIDs).Error; err != nil {
		return err
	}
	if len(runIDs) > 0 {
		if err := tx.Model(&entity.AdvisorRun{}).Where("conversation_id = ?", conversation.ID).Updates(map[string]any{
			"status":      entity.AdvisorStatusPurged,
			"finished_at": now,
			"lease_token": nil,
			"lease_until": nil,
		}).Error; err != nil {
			return err
		}
		if err := tx.Where("run_id IN ?", runIDs).Delete(&entity.AdvisorEvent{}).Error; err != nil {
			return err
		}
		if err := tx.Where("run_id IN ?", runIDs).Delete(&entity.AdvisorFactBundle{}).Error; err != nil {
			return err
		}
	}
	if err := tx.Where("conversation_id = ?", conversation.ID).Delete(&entity.AdvisorMessage{}).Error; err != nil {
		return err
	}
	conversation.Generation++
	conversation.LastUserMessageAt = nil
	conversation.Summary = map[string]any{}
	conversation.SummaryThroughSeq = 0
	conversation.SummaryVersion = 0
	conversation.UpdatedAt = now
	return tx.Model(&entity.AdvisorConversation{}).Where("id = ? AND owner_id = ?", conversation.ID, conversation.OwnerID).Updates(map[string]any{
		"generation":           conversation.Generation,
		"last_user_message_at": nil,
		"summary":              map[string]any{},
		"summary_through_seq":  0,
		"summary_version":      0,
		"updated_at":           now,
	}).Error
}

func isUniqueViolation(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "duplicate key") || strings.Contains(strings.ToLower(err.Error()), "unique constraint")
}
