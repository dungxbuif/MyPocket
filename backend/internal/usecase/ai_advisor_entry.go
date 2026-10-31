package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
)

type AdvisorEntryInput struct {
	ClientRequestID string
	ConversationID  string
	Generation      int64
	Text            string
	Timezone        string
	Files           []entity.AIImage
}

type AdvisorEntryService struct {
	Tasks   repository.AdvisorEntryTaskRepository
	Entries *AIEntryService
	Now     func() time.Time
}

func (s *AdvisorEntryService) Accept(ctx context.Context, principal Principal, input AdvisorEntryInput) (*entity.AdvisorEntryTask, error) {
	if s == nil || s.Tasks == nil || s.Entries == nil || strings.TrimSpace(principal.OwnerID) == "" {
		return nil, ErrAdvisorProviderUnavailable
	}
	if err := validateAdvisorEntryInput(input); err != nil {
		return nil, err
	}
	input.Text = strings.TrimSpace(input.Text)
	if input.Timezone == "" {
		input.Timezone = principal.Timezone
	}
	if input.Timezone == "" {
		input.Timezone = "Asia/Ho_Chi_Minh"
	}
	payloadHash, err := advisorEntryPayloadHash(input)
	if err != nil {
		return nil, err
	}
	if existing, lookupErr := s.Tasks.GetTaskByRequest(ctx, principal.OwnerID, input.ConversationID, input.Generation, input.ClientRequestID); lookupErr == nil {
		if existing.PayloadHash != payloadHash {
			return nil, repository.ErrAdvisorEntryConflict
		}
		return existing, nil
	} else if !errors.Is(lookupErr, repository.ErrNotFound) {
		return nil, lookupErr
	}

	taskID := uuid.NewString()
	if err := s.Entries.Entries.CreateProcess(ctx, principal.OwnerID, taskID); err != nil {
		return nil, err
	}
	if _, err := s.Entries.StoreAcceptedAttachments(ctx, principal.OwnerID, taskID, input.Files); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if s.Now != nil {
		now = s.Now().UTC()
	}
	task, err := s.Tasks.CreateTask(ctx, entity.AdvisorEntryTask{
		ID: taskID, OwnerID: principal.OwnerID, ConversationID: input.ConversationID,
		Generation: input.Generation, ClientRequestID: input.ClientRequestID,
		PayloadHash: payloadHash, SessionID: taskID, Status: entity.AdvisorEntryTaskAccepted,
		Stage: "accepted", InputText: input.Text, Timezone: input.Timezone,
		CreatedAt: now, UpdatedAt: now, DeadlineAt: now.Add(10 * time.Minute),
	})
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (s *AdvisorEntryService) Get(ctx context.Context, ownerID, taskID string) (*entity.AdvisorEntryTaskView, error) {
	if s == nil || s.Tasks == nil || s.Entries == nil {
		return nil, repository.ErrNotFound
	}
	task, err := s.Tasks.GetTask(ctx, ownerID, taskID)
	if err != nil {
		return nil, err
	}
	view := &entity.AdvisorEntryTaskView{Task: *task, Attachments: []entity.AIEntryAttachment{}, Events: []entity.AdvisorEntryTaskEvent{}}
	if attachmentsRepo, ok := s.Entries.Entries.(repository.AIEntryStoredInputRepository); ok {
		view.Attachments, err = attachmentsRepo.ListAttachments(ctx, ownerID, task.SessionID)
		if err != nil {
			return nil, err
		}
	}
	if eventRepo, ok := s.Tasks.(interface {
		ListEvents(context.Context, string, string, int64, int) ([]entity.AdvisorEntryTaskEvent, error)
	}); ok {
		view.Events, err = eventRepo.ListEvents(ctx, ownerID, task.ID, 0, 100)
		if err != nil {
			return nil, err
		}
	}
	if task.BatchID != "" {
		view.Batch, err = s.Entries.Batch(ctx, ownerID, task.BatchID, task.BatchVersion)
		if err != nil {
			return nil, err
		}
	}
	return view, nil
}

func (s *AdvisorEntryService) Cancel(ctx context.Context, ownerID, taskID string) error {
	if s == nil || s.Tasks == nil {
		return repository.ErrNotFound
	}
	return s.Tasks.CancelTask(ctx, ownerID, taskID)
}

func (s *AdvisorEntryService) ProcessClaimed(ctx context.Context, lease repository.AdvisorEntryTaskLease) error {
	if s == nil || s.Tasks == nil || s.Entries == nil {
		return ErrAdvisorProviderUnavailable
	}
	task, err := s.Tasks.GetTask(ctx, lease.OwnerID, lease.TaskID)
	if err != nil {
		return err
	}
	if task.Generation != lease.Generation || task.ConversationID != lease.ConversationID {
		return repository.ErrAdvisorEntryLeaseLost
	}
	_, err = s.Entries.ProcessStored(ctx, lease.OwnerID, task.SessionID, AIEntryMessageInput{RequestID: task.ClientRequestID, Text: task.InputText, Timezone: task.Timezone})
	if err != nil {
		_ = s.Tasks.FailTask(context.WithoutCancel(ctx), lease, advisorEntryErrorCode(err))
		return err
	}
	session, err := s.Entries.Entries.Session(ctx, lease.OwnerID, task.SessionID)
	if err != nil {
		_ = s.Tasks.FailTask(context.WithoutCancel(ctx), lease, "result_read_failed")
		return err
	}
	itemIDs := make([]string, 0, len(session.Proposals))
	for _, proposal := range session.Proposals {
		itemIDs = append(itemIDs, proposal.ID)
	}
	return s.Tasks.CompleteTask(ctx, lease, entity.AdvisorEntryBatchPart{TaskID: task.ID, BatchID: task.SessionID, Version: 1, ItemIDs: itemIDs, ItemCount: len(itemIDs), Summary: fmt.Sprintf("%d đề xuất cần xem lại", len(itemIDs))})
}

func validateAdvisorEntryInput(input AdvisorEntryInput) error {
	if strings.TrimSpace(input.ClientRequestID) == "" || len(input.ClientRequestID) > 128 || strings.TrimSpace(input.ConversationID) == "" || len(input.ConversationID) > 128 || input.Generation <= 0 {
		return repository.ErrAdvisorEntryInvalid
	}
	if len(input.Text) > 16000 || (strings.TrimSpace(input.Text) == "" && len(input.Files) == 0) || len(input.Files) > 20 {
		return repository.ErrAdvisorEntryInvalid
	}
	if input.Timezone != "" {
		if input.Timezone == "Local" {
			return repository.ErrAdvisorEntryInvalid
		}
		if _, err := time.LoadLocation(input.Timezone); err != nil {
			return repository.ErrAdvisorEntryInvalid
		}
	}
	hasXLSX, hasOCR := false, false
	for _, file := range input.Files {
		if file.Name == "" || len(file.Name) > 255 || len(file.Base64) > 7*1024*1024 {
			return repository.ErrAdvisorEntryInvalid
		}
		data, err := base64.StdEncoding.Strict().DecodeString(file.Base64)
		if err != nil || len(data) == 0 || len(data) > 5*1024*1024 {
			return repository.ErrAdvisorEntryInvalid
		}
		switch file.MIMEType {
		case "image/jpeg", "image/png", "application/pdf":
			hasOCR = true
		case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
			hasXLSX = true
		default:
			return repository.ErrAdvisorEntryInvalid
		}
	}
	if hasXLSX && hasOCR {
		return repository.ErrAdvisorEntryInvalid
	}
	return nil
}

func advisorEntryPayloadHash(input AdvisorEntryInput) (string, error) {
	type fileFingerprint struct {
		Name, MIMEType, SHA256 string
		Size                   int
	}
	fingerprints := make([]fileFingerprint, 0, len(input.Files))
	for _, file := range input.Files {
		data, err := base64.StdEncoding.Strict().DecodeString(file.Base64)
		if err != nil {
			return "", repository.ErrAdvisorEntryInvalid
		}
		digest := sha256.Sum256(data)
		fingerprints = append(fingerprints, fileFingerprint{Name: file.Name, MIMEType: file.MIMEType, SHA256: hex.EncodeToString(digest[:]), Size: len(data)})
	}
	canonical, err := json.Marshal(struct {
		ConversationID string
		Generation     int64
		Text           string
		Timezone       string
		Files          []fileFingerprint
	}{input.ConversationID, input.Generation, strings.TrimSpace(input.Text), input.Timezone, fingerprints})
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(canonical)
	return hex.EncodeToString(digest[:]), nil
}

func advisorEntryErrorCode(err error) string {
	switch {
	case errors.Is(err, ErrAIStorage):
		return "storage_error"
	case errors.Is(err, ErrAIProvider):
		return "provider_error"
	case errors.Is(err, repository.ErrAIInvalid):
		return "input_invalid"
	default:
		return "worker_failed"
	}
}
