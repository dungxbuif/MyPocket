package repository

import (
	"context"
	"errors"
	"github.com/mypocket/backend/internal/entity"
)

var (
	ErrAIRateLimited = errors.New("đã dùng hết 20 lượt AI trong 24 giờ; vui lòng thử lại sau")
	ErrAIInvalid     = errors.New("dữ liệu đề xuất chưa hợp lệ")
	ErrAIConflict    = errors.New("đề xuất hoặc phiên đã thay đổi; hãy tải lại")
	ErrAIBusy        = errors.New("phiên đang xử lý, vui lòng chờ")
	ErrAISessionFull = errors.New("phiên đã đủ 20 lượt; hãy mở cuộc trò chuyện mới")
)

type AIEntryRepository interface {
	CreateSession(context.Context, string) (*entity.AIEntrySession, error)
	Session(context.Context, string, string) (*entity.AIEntrySession, error)
	LatestSession(context.Context, string) (*entity.AIEntrySession, error)
	BeginMessage(ctx context.Context, owner, id, requestID, hash, text string) (token string, started bool, err error)
	FinishMessage(ctx context.Context, owner, id, token string, output entity.AIExtractOutput) error
	FailMessage(ctx context.Context, owner, id, token, message, sourceText string) error
	EditProposal(ctx context.Context, owner, id string, version int, draft entity.AIEntryDraft) (*entity.AIEntryProposal, error)
	DecideProposal(ctx context.Context, owner, id string, version int, approve bool) (*entity.AIEntryProposal, error)
}
