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

const advisorEntryLease = 45 * time.Second

type AdvisorEntryPostgresRepository struct{ db *gorm.DB }

func NewAdvisorEntryPostgresRepository(db *gorm.DB) *AdvisorEntryPostgresRepository {
	return &AdvisorEntryPostgresRepository{db: db}
}

func (r *AdvisorEntryPostgresRepository) CreateTask(ctx context.Context, task entity.AdvisorEntryTask) (*entity.AdvisorEntryTask, error) {
	if r == nil || r.db == nil || task.ID == "" {
		return nil, contract.ErrAdvisorEntryInvalid
	}
	if err := task.Validate(); err != nil {
		return nil, contract.ErrAdvisorEntryInvalid
	}
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now().UTC()
	}
	if task.UpdatedAt.IsZero() {
		task.UpdatedAt = task.CreatedAt
	}
	if task.DeadlineAt.IsZero() {
		task.DeadlineAt = task.CreatedAt.Add(10 * time.Minute)
	}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var conversation entity.AdvisorConversation
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND owner_id = ?", task.ConversationID, task.OwnerID).First(&conversation).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return contract.ErrNotFound
			}
			return err
		}
		if conversation.Generation != task.Generation {
			return contract.ErrAdvisorEntryConflict
		}
		var session entity.AIEntrySession
		if err := tx.Where("id = ? AND owner_id = ?", task.SessionID, task.OwnerID).First(&session).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return contract.ErrNotFound
			}
			return err
		}
		if err := tx.Create(&task).Error; err != nil {
			if isUniqueViolation(err) {
				return contract.ErrAdvisorEntryConflict
			}
			return err
		}
		var attachments []entity.AIEntryAttachment
		if err := tx.Where("owner_id = ? AND session_id = ?", task.OwnerID, task.SessionID).Order("created_at, id").Find(&attachments).Error; err != nil {
			return err
		}
		bindings := make([]entity.AdvisorEntryAttachmentBinding, 0, len(attachments))
		for _, attachment := range attachments {
			bindings = append(bindings, entity.AdvisorEntryAttachmentBinding{TaskID: task.ID, OwnerID: task.OwnerID, AttachmentID: attachment.ID, CreatedAt: task.CreatedAt})
		}
		if len(bindings) > 0 {
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&bindings).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *AdvisorEntryPostgresRepository) GetTask(ctx context.Context, ownerID, taskID string) (*entity.AdvisorEntryTask, error) {
	if strings.TrimSpace(ownerID) == "" || strings.TrimSpace(taskID) == "" {
		return nil, contract.ErrNotFound
	}
	var task entity.AdvisorEntryTask
	err := r.db.WithContext(ctx).Where("id = ? AND owner_id = ?", taskID, ownerID).First(&task).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, contract.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *AdvisorEntryPostgresRepository) GetTaskByRequest(ctx context.Context, ownerID, conversationID string, generation int64, requestID string) (*entity.AdvisorEntryTask, error) {
	if strings.TrimSpace(ownerID) == "" || strings.TrimSpace(conversationID) == "" || generation <= 0 || strings.TrimSpace(requestID) == "" {
		return nil, contract.ErrNotFound
	}
	var task entity.AdvisorEntryTask
	err := r.db.WithContext(ctx).Where("owner_id = ? AND conversation_id = ? AND generation = ? AND client_request_id = ?", ownerID, conversationID, generation, requestID).First(&task).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, contract.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *AdvisorEntryPostgresRepository) ClaimTask(ctx context.Context, ownerID, taskID, token string) (contract.AdvisorEntryTaskLease, error) {
	if strings.TrimSpace(ownerID) == "" || strings.TrimSpace(taskID) == "" || strings.TrimSpace(token) == "" {
		return contract.AdvisorEntryTaskLease{}, contract.ErrAdvisorEntryLeaseLost
	}
	var lease contract.AdvisorEntryTaskLease
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var task entity.AdvisorEntryTask
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND owner_id = ?", taskID, ownerID).First(&task).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return contract.ErrNotFound
			}
			return err
		}
		return r.claimTaskTx(tx, &task, token, &lease)
	})
	return lease, err
}

func (r *AdvisorEntryPostgresRepository) ClaimNextTask(ctx context.Context, token string) (contract.AdvisorEntryTaskLease, error) {
	if strings.TrimSpace(token) == "" {
		return contract.AdvisorEntryTaskLease{}, contract.ErrAdvisorEntryLeaseLost
	}
	var lease contract.AdvisorEntryTaskLease
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var task entity.AdvisorEntryTask
		err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).Where("status = ?", entity.AdvisorEntryTaskAccepted).Order("created_at ASC, id ASC").First(&task).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return contract.ErrNotFound
		}
		if err != nil {
			return err
		}
		return r.claimTaskTx(tx, &task, token, &lease)
	})
	return lease, err
}

func (r *AdvisorEntryPostgresRepository) claimTaskTx(tx *gorm.DB, task *entity.AdvisorEntryTask, token string, lease *contract.AdvisorEntryTaskLease) error {
	if task == nil || task.Status != entity.AdvisorEntryTaskAccepted || task.Generation <= 0 {
		return contract.ErrAdvisorEntryLeaseLost
	}
	var conversation entity.AdvisorConversation
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND owner_id = ?", task.ConversationID, task.OwnerID).First(&conversation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return contract.ErrNotFound
		}
		return err
	}
	now, err := advisorDBNow(tx)
	if err != nil {
		return err
	}
	if conversation.Generation != task.Generation || !task.DeadlineAt.After(now) {
		return contract.ErrAdvisorEntryLeaseLost
	}
	until := now.Add(advisorEntryLease)
	if until.After(task.DeadlineAt) {
		until = task.DeadlineAt
	}
	if err := tx.Model(&entity.AdvisorEntryTask{}).Where("id = ? AND owner_id = ? AND generation = ? AND status = ?", task.ID, task.OwnerID, task.Generation, entity.AdvisorEntryTaskAccepted).Updates(map[string]any{"status": entity.AdvisorEntryTaskProcessing, "lease_token": token, "lease_until": until, "updated_at": now}).Error; err != nil {
		return err
	}
	lease.OwnerID = task.OwnerID
	lease.ConversationID = task.ConversationID
	lease.Generation = task.Generation
	lease.TaskID = task.ID
	lease.Token = token
	lease.LeaseUntil = until
	return nil
}

func (r *AdvisorEntryPostgresRepository) HeartbeatTask(ctx context.Context, lease contract.AdvisorEntryTaskLease) error {
	if !validEntryLease(lease) {
		return contract.ErrAdvisorEntryLeaseLost
	}
	result := r.db.WithContext(ctx).Model(&entity.AdvisorEntryTask{}).Where("id = ? AND owner_id = ? AND conversation_id = ? AND generation = ? AND lease_token = ? AND status = ? AND lease_until > NOW() AND deadline_at > NOW()", lease.TaskID, lease.OwnerID, lease.ConversationID, lease.Generation, lease.Token, entity.AdvisorEntryTaskProcessing).Updates(map[string]any{"lease_until": gorm.Expr("LEAST(NOW() + interval '45 seconds', deadline_at)"), "updated_at": gorm.Expr("NOW()")})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return contract.ErrAdvisorEntryLeaseLost
	}
	return nil
}

func (r *AdvisorEntryPostgresRepository) CompleteTask(ctx context.Context, lease contract.AdvisorEntryTaskLease, batch entity.AdvisorEntryBatchPart) error {
	if !validEntryLease(lease) || strings.TrimSpace(batch.BatchID) == "" || batch.Version <= 0 {
		return contract.ErrAdvisorEntryLeaseLost
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		task, conversation, now, err := r.lockLeasedTask(tx, lease)
		if err != nil {
			return err
		}
		if err := tx.Model(&entity.AdvisorEntryTask{}).Where("id = ? AND owner_id = ?", task.ID, task.OwnerID).Updates(map[string]any{"status": entity.AdvisorEntryTaskReviewReady, "stage": "review_ready", "batch_id": batch.BatchID, "batch_version": batch.Version, "lease_token": nil, "lease_until": nil, "error_code": nil, "completed_at": now, "updated_at": now}).Error; err != nil {
			return err
		}
		if err := r.persistBatchReferenceTx(tx, task, conversation, now, batch); err != nil {
			return err
		}
		return r.appendEventTx(tx, task, conversation, now, "entry.review_ready", map[string]any{"task_id": task.ID, "batch_id": batch.BatchID, "version": batch.Version, "item_ids": batch.ItemIDs, "item_count": batch.ItemCount, "summary": batch.Summary})
	})
}

func (r *AdvisorEntryPostgresRepository) FailTask(ctx context.Context, lease contract.AdvisorEntryTaskLease, errorCode string) error {
	if !validEntryLease(lease) {
		return contract.ErrAdvisorEntryLeaseLost
	}
	if strings.TrimSpace(errorCode) == "" {
		errorCode = "worker_failed"
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		task, conversation, now, err := r.lockLeasedTask(tx, lease)
		if err != nil {
			return err
		}
		if err := tx.Model(&entity.AdvisorEntryTask{}).Where("id = ? AND owner_id = ?", task.ID, task.OwnerID).Updates(map[string]any{"status": entity.AdvisorEntryTaskFailed, "stage": "failed", "error_code": errorCode, "lease_token": nil, "lease_until": nil, "completed_at": now, "updated_at": now}).Error; err != nil {
			return err
		}
		return r.appendEventTx(tx, task, conversation, now, "entry.failed", map[string]any{"task_id": task.ID, "error_code": errorCode})
	})
}

func (r *AdvisorEntryPostgresRepository) CancelTask(ctx context.Context, ownerID, taskID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var task entity.AdvisorEntryTask
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND owner_id = ?", taskID, ownerID).First(&task).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return contract.ErrNotFound
			}
			return err
		}
		if task.Status == entity.AdvisorEntryTaskReviewReady || task.Status == entity.AdvisorEntryTaskFailed || task.Status == entity.AdvisorEntryTaskCancelled {
			return nil
		}
		var conversation entity.AdvisorConversation
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND owner_id = ?", task.ConversationID, ownerID).First(&conversation).Error; err != nil {
			return err
		}
		now, err := advisorDBNow(tx)
		if err != nil {
			return err
		}
		if err := tx.Model(&entity.AdvisorEntryTask{}).Where("id = ? AND owner_id = ? AND status IN ?", task.ID, ownerID, []string{entity.AdvisorEntryTaskAccepted, entity.AdvisorEntryTaskProcessing}).Updates(map[string]any{"status": entity.AdvisorEntryTaskCancelled, "stage": "cancelled", "lease_token": nil, "lease_until": nil, "completed_at": now, "updated_at": now}).Error; err != nil {
			return err
		}
		task.Status = entity.AdvisorEntryTaskCancelled
		return r.appendEventTx(tx, &task, &conversation, now, "entry.cancelled", map[string]any{"task_id": task.ID})
	})
}

func (r *AdvisorEntryPostgresRepository) AppendEvent(ctx context.Context, lease contract.AdvisorEntryTaskLease, eventType string, payload any) (*entity.AdvisorEntryTaskEvent, error) {
	if !validEntryLease(lease) || strings.TrimSpace(eventType) == "" {
		return nil, contract.ErrAdvisorEntryLeaseLost
	}
	encoded, err := json.Marshal(payload)
	if err != nil || len(encoded) > 64*1024 {
		return nil, contract.ErrAdvisorEntryInvalid
	}
	var event entity.AdvisorEntryTaskEvent
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		task, conversation, now, err := r.lockLeasedTask(tx, lease)
		if err != nil {
			return err
		}
		return r.appendEventTxWithResult(tx, task, conversation, now, eventType, payload, &event)
	})
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *AdvisorEntryPostgresRepository) ListEvents(ctx context.Context, ownerID, taskID string, afterSeq int64, limit int) ([]entity.AdvisorEntryTaskEvent, error) {
	if strings.TrimSpace(ownerID) == "" || strings.TrimSpace(taskID) == "" || afterSeq < 0 {
		return nil, contract.ErrNotFound
	}
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	var events []entity.AdvisorEntryTaskEvent
	err := r.db.WithContext(ctx).Where("task_id = ? AND owner_id = ? AND seq > ?", taskID, ownerID, afterSeq).Order("seq ASC").Limit(limit).Find(&events).Error
	return events, err
}

func (r *AdvisorEntryPostgresRepository) ReapExpiredTasks(ctx context.Context, ownerID string, limit int) (int, error) {
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	query := r.db.WithContext(ctx).Table("advisor_entry_tasks").Select("id, owner_id").Where("status IN ? AND (deadline_at <= NOW() OR (status = ? AND lease_until <= NOW()))", []string{entity.AdvisorEntryTaskAccepted, entity.AdvisorEntryTaskProcessing}, entity.AdvisorEntryTaskProcessing)
	if strings.TrimSpace(ownerID) != "" {
		query = query.Where("owner_id = ?", ownerID)
	}
	var candidates []struct{ ID, OwnerID string }
	if err := query.Order("created_at ASC, id ASC").Limit(limit).Find(&candidates).Error; err != nil {
		return 0, err
	}
	count := 0
	for _, candidate := range candidates {
		updated, err := r.reapOne(ctx, candidate.OwnerID, candidate.ID)
		if err != nil {
			return count, err
		}
		if updated {
			count++
		}
	}
	return count, nil
}

func (r *AdvisorEntryPostgresRepository) reapOne(ctx context.Context, ownerID, taskID string) (bool, error) {
	updated := false
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var task entity.AdvisorEntryTask
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND owner_id = ?", taskID, ownerID).First(&task).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
		if task.Status != entity.AdvisorEntryTaskAccepted && task.Status != entity.AdvisorEntryTaskProcessing {
			return nil
		}
		var conversation entity.AdvisorConversation
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND owner_id = ?", task.ConversationID, ownerID).First(&conversation).Error; err != nil {
			return err
		}
		now, err := advisorDBNow(tx)
		if err != nil {
			return err
		}
		if task.DeadlineAt.After(now) && (task.Status != entity.AdvisorEntryTaskProcessing || task.LeaseUntil != nil && task.LeaseUntil.After(now)) {
			return nil
		}
		code := "executor_lost"
		if !task.DeadlineAt.After(now) {
			code = "deadline_exceeded"
		}
		if err := tx.Model(&entity.AdvisorEntryTask{}).Where("id = ? AND owner_id = ? AND status IN ?", task.ID, ownerID, []string{entity.AdvisorEntryTaskAccepted, entity.AdvisorEntryTaskProcessing}).Updates(map[string]any{"status": entity.AdvisorEntryTaskFailed, "stage": "failed", "error_code": code, "lease_token": nil, "lease_until": nil, "completed_at": now, "updated_at": now}).Error; err != nil {
			return err
		}
		task.Status = entity.AdvisorEntryTaskFailed
		if err := r.appendEventTx(tx, &task, &conversation, now, "entry.failed", map[string]any{"task_id": task.ID, "error_code": code}); err != nil {
			return err
		}
		updated = true
		return nil
	})
	return updated, err
}

func (r *AdvisorEntryPostgresRepository) lockLeasedTask(tx *gorm.DB, lease contract.AdvisorEntryTaskLease) (*entity.AdvisorEntryTask, *entity.AdvisorConversation, time.Time, error) {
	var task entity.AdvisorEntryTask
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND owner_id = ?", lease.TaskID, lease.OwnerID).First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, time.Time{}, contract.ErrNotFound
		}
		return nil, nil, time.Time{}, err
	}
	var conversation entity.AdvisorConversation
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND owner_id = ?", lease.ConversationID, lease.OwnerID).First(&conversation).Error; err != nil {
		return nil, nil, time.Time{}, err
	}
	now, err := advisorDBNow(tx)
	if err != nil {
		return nil, nil, time.Time{}, err
	}
	if task.ConversationID != lease.ConversationID || task.Generation != lease.Generation || conversation.Generation != lease.Generation || task.LeaseToken != lease.Token || task.Status != entity.AdvisorEntryTaskProcessing || task.LeaseUntil == nil || !task.LeaseUntil.After(now) || !task.DeadlineAt.After(now) {
		return nil, nil, time.Time{}, contract.ErrAdvisorEntryLeaseLost
	}
	return &task, &conversation, now, nil
}

func (r *AdvisorEntryPostgresRepository) appendEventTx(tx *gorm.DB, task *entity.AdvisorEntryTask, conversation *entity.AdvisorConversation, now time.Time, eventType string, payload any) error {
	return r.appendEventTxWithResult(tx, task, conversation, now, eventType, payload, nil)
}

func (r *AdvisorEntryPostgresRepository) appendEventTxWithResult(tx *gorm.DB, task *entity.AdvisorEntryTask, _ *entity.AdvisorConversation, now time.Time, eventType string, payload any, result *entity.AdvisorEntryTaskEvent) error {
	encoded, err := json.Marshal(payload)
	if err != nil || len(encoded) > 64*1024 {
		return contract.ErrAdvisorEntryInvalid
	}
	seq := task.LastEventSeq + 1
	event := entity.AdvisorEntryTaskEvent{TaskID: task.ID, OwnerID: task.OwnerID, Seq: seq, Type: eventType, Payload: payload, CreatedAt: now}
	if err := tx.Create(&event).Error; err != nil {
		return err
	}
	if err := tx.Model(&entity.AdvisorEntryTask{}).Where("id = ? AND owner_id = ?", task.ID, task.OwnerID).Update("last_event_seq", seq).Error; err != nil {
		return err
	}
	if result != nil {
		*result = event
	}
	return nil
}

// persistBatchReferenceTx makes completed entry work visible in the advisor
// conversation without duplicating proposal drafts. A synthetic completed run
// satisfies the existing advisor_messages foreign key; the entry batch itself
// remains authoritative in ai_entry_proposals and is hydrated on read.
func (r *AdvisorEntryPostgresRepository) persistBatchReferenceTx(tx *gorm.DB, task *entity.AdvisorEntryTask, conversation *entity.AdvisorConversation, now time.Time, batch entity.AdvisorEntryBatchPart) error {
	deadline := now
	run := entity.AdvisorRun{
		ID: task.ID, OwnerID: task.OwnerID, ConversationID: conversation.ID, Generation: task.Generation,
		ClientRequestID: "entry:" + task.ID, PayloadHash: task.PayloadHash, CredentialKind: "entry_task",
		CredentialID: "", DeadlineAt: &deadline, Status: entity.AdvisorStatusCompleted, CreatedAt: now, FinishedAt: &now,
		ModelUsage: map[string]any{},
	}
	if err := tx.Create(&run).Error; err != nil {
		if !isUniqueViolation(err) {
			return err
		}
	}
	message := entity.AdvisorMessage{
		ID: uuid.NewString(), ConversationID: conversation.ID, Generation: task.Generation,
		Seq: conversation.NextMessageSeq, RunID: task.ID, Role: entity.AdvisorRoleAssistant,
		PartsVersion: 1, Parts: []entity.AdvisorPart{{Type: "entry_batch", Data: map[string]any{
			"task_id": task.ID, "batch_id": batch.BatchID, "version": batch.Version,
			"item_ids": batch.ItemIDs, "item_count": batch.ItemCount, "summary": batch.Summary,
		}}}, CreatedAt: now,
	}
	if err := message.Validate(); err != nil {
		return err
	}
	if err := tx.Create(&message).Error; err != nil {
		if !isUniqueViolation(err) {
			return err
		}
		return nil
	}
	conversation.NextMessageSeq++
	return tx.Model(&entity.AdvisorConversation{}).Where("id = ? AND owner_id = ?", conversation.ID, conversation.OwnerID).Updates(map[string]any{"next_message_seq": conversation.NextMessageSeq, "updated_at": now}).Error
}

func validEntryLease(lease contract.AdvisorEntryTaskLease) bool {
	return strings.TrimSpace(lease.OwnerID) != "" && strings.TrimSpace(lease.ConversationID) != "" && lease.Generation > 0 && strings.TrimSpace(lease.TaskID) != "" && strings.TrimSpace(lease.Token) != ""
}

var _ contract.AdvisorEntryTaskRepository = (*AdvisorEntryPostgresRepository)(nil)
