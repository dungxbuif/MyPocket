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
		var existing entity.AdvisorRun
		if err := tx.Where("owner_id = ? AND client_request_id = ?", input.OwnerID, input.RequestID).First(&existing).Error; err == nil {
			if existing.PayloadHash != input.PayloadHash {
				return contract.ErrAdvisorRequestConflict
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
		now := time.Now().UTC()
		run = entity.AdvisorRun{ID: uuid.NewString(), OwnerID: input.OwnerID, ConversationID: conversation.ID, Generation: conversation.Generation, ClientRequestID: input.RequestID, PayloadHash: input.PayloadHash, CredentialKind: input.CredentialKind, CredentialID: input.CredentialID, CredentialExpiresAt: input.ExpiresAt, Status: entity.AdvisorStatusQueued, ModelUsage: map[string]any{}, CreatedAt: now}
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
		return tx.Model(&entity.AdvisorConversation{}).Where("id = ?", conversation.ID).Updates(map[string]any{"next_message_seq": conversation.NextMessageSeq, "updated_at": now}).Error
	})
	if err != nil {
		return nil, nil, false, err
	}
	if replayed {
		return &run, &replayMessage, true, nil
	}
	return &run, message, false, nil
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
	err := r.db.WithContext(ctx).Where("id = ? AND owner_id = ?", id, ownerID).First(&run).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, contract.ErrNotFound
	}
	return &run, err
}

func (r *AdvisorPostgresRepository) GetRunByRequest(ctx context.Context, ownerID, requestID string) (*entity.AdvisorRun, error) {
	var run entity.AdvisorRun
	err := r.db.WithContext(ctx).Where("owner_id = ? AND client_request_id = ?", ownerID, requestID).First(&run).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, contract.ErrNotFound
	}
	return &run, err
}

func (r *AdvisorPostgresRepository) ListMessages(ctx context.Context, ownerID, conversationID string, beforeSeq int64, limit int) ([]entity.AdvisorMessage, error) {
	if limit <= 0 || limit > 50 {
		limit = 50
	}
	query := r.db.WithContext(ctx).Where("conversation_id = ? AND EXISTS (SELECT 1 FROM advisor_conversations WHERE id = advisor_messages.conversation_id AND owner_id = ?)", conversationID, ownerID)
	if beforeSeq > 0 {
		query = query.Where("seq < ?", beforeSeq)
	}
	var messages []entity.AdvisorMessage
	if err := query.Order("seq DESC").Limit(limit).Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *AdvisorPostgresRepository) AppendEvent(ctx context.Context, ownerID, runID, eventType string, payload any) (*entity.AdvisorEvent, error) {
	var event entity.AdvisorEvent
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var run entity.AdvisorRun
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND owner_id = ?", runID, ownerID).First(&run).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return contract.ErrNotFound
			}
			return err
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

func isUniqueViolation(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "duplicate key") || strings.Contains(strings.ToLower(err.Error()), "unique constraint")
}
