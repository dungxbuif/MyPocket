package repository

import (
	"context"
	"errors"
	"github.com/mypocket/backend/internal/entity"
	"time"
)

var (
	ErrAIInvalid  = errors.New("dữ liệu đề xuất chưa hợp lệ")
	ErrAIConflict = errors.New("kết quả hoặc đề xuất đã thay đổi; hãy tải lại")
	ErrAIBusy     = errors.New("yêu cầu đang xử lý, vui lòng chờ")
)

type AIEntryRepository interface {
	CreateSession(context.Context, string) (*entity.AIEntrySession, error)
	CreateProcess(context.Context, string, string) error
	CreateAttachments(context.Context, []entity.AIEntryAttachment) error
	UpdateAttachmentOCR(context.Context, string, string, []string, bool) error
	AttachmentForTransaction(ctx context.Context, owner, transactionID, attachmentID string) (*entity.AIEntryAttachment, error)
	Session(context.Context, string, string) (*entity.AIEntrySession, error)
	LatestSession(context.Context, string) (*entity.AIEntrySession, error)
	BeginMessage(ctx context.Context, owner, id, requestID, hash, text string) (token string, started bool, err error)
	FinishMessage(ctx context.Context, owner, id, token string, output entity.AIExtractOutput) error
	FailMessage(ctx context.Context, owner, id, token, message, errorCode, sourceText string) error
	EditProposal(ctx context.Context, owner, id string, version int, draft entity.AIEntryDraft) (*entity.AIEntryProposal, error)
	DecideProposal(ctx context.Context, owner, id string, version int, approve bool) (*entity.AIEntryProposal, error)
}

type AIEntryAttachmentCleanupRepository interface {
	ClaimExpiredAttachments(ctx context.Context, before, staleClaim time.Time, limit int) ([]entity.AIEntryAttachment, error)
	SetAttachmentDeleteStatus(ctx context.Context, id, status string) error
}
