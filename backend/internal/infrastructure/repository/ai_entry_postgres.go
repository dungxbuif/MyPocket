package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	port "github.com/mypocket/backend/internal/repository"
	"github.com/mypocket/backend/internal/usecase"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"strings"
	"time"
)

type AIEntryPostgresRepository struct{ db *gorm.DB }

func NewAIEntryPostgresRepository(db *gorm.DB) *AIEntryPostgresRepository {
	return &AIEntryPostgresRepository{db: db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)})}
}

func (r *AIEntryPostgresRepository) CreateSession(ctx context.Context, owner string) (*entity.AIEntrySession, error) {
	s := entity.AIEntrySession{ID: uuid.NewString(), OwnerID: owner, Messages: []entity.AIEntryMessage{}, Proposals: []entity.AIEntryProposal{}}
	err := r.db.WithContext(ctx).Create(&s).Error
	return &s, err
}
func (r *AIEntryPostgresRepository) Session(ctx context.Context, owner, id string) (*entity.AIEntrySession, error) {
	var s entity.AIEntrySession
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Serialize snapshots with message completion so UI never sees half a result.
		if err := tx.Clauses(clause.Locking{Strength: "SHARE"}).Where("id = ? AND owner_id = ?", id, owner).First(&s).Error; err != nil {
			return err
		}
		s.Messages = []entity.AIEntryMessage{}
		s.Proposals = []entity.AIEntryProposal{}
		if err := tx.Where("session_id = ?", id).Order("created_at, id").Find(&s.Messages).Error; err != nil {
			return err
		}
		return tx.Where("session_id = ? AND owner_id = ?", id, owner).Order("created_at, id").Find(&s.Proposals).Error
	})
	if s.ProcessingUntil != nil {
		s.Processing = time.Now().Before(*s.ProcessingUntil)
		if !s.Processing && s.Error == "" {
			s.Error = "Lần xử lý trước đã gián đoạn. Bạn có thể gửi lại; dịch vụ có thể tính thêm lượt xử lý."
		}
	}
	return &s, err
}
func (r *AIEntryPostgresRepository) LatestSession(ctx context.Context, owner string) (*entity.AIEntrySession, error) {
	var s entity.AIEntrySession
	err := r.db.WithContext(ctx).Where("owner_id = ?", owner).Order("updated_at DESC, id DESC").First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return r.Session(ctx, owner, s.ID)
}
func (r *AIEntryPostgresRepository) BeginMessage(ctx context.Context, owner, id, requestID, hash, text string) (token string, started bool, err error) {
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Serialize the owner budget across separate sessions, without blocking FK reads.
		var account entity.User
		if e := tx.Select("id").Clauses(clause.Locking{Strength: "NO KEY UPDATE"}).Where("id = ?", owner).First(&account).Error; e != nil {
			return e
		}
		var s entity.AIEntrySession
		if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND owner_id = ?", id, owner).First(&s).Error; e != nil {
			return e
		}
		var prior entity.AIEntryRequest
		e := tx.Where("session_id = ? AND request_id = ?", id, requestID).First(&prior).Error
		if e == nil {
			if prior.Hash != hash {
				return port.ErrAIConflict
			}
			return nil
		}
		if !errors.Is(e, gorm.ErrRecordNotFound) {
			return e
		}
		if s.ProcessingUntil != nil && time.Now().Before(*s.ProcessingUntil) {
			return port.ErrAIBusy
		}
		var daily int64
		if e = tx.Model(&entity.AIEntryRequest{}).Joins("JOIN ai_entry_sessions ON ai_entry_sessions.id = ai_entry_requests.session_id").Where("ai_entry_sessions.owner_id = ? AND ai_entry_requests.created_at > ?", owner, time.Now().Add(-24*time.Hour)).Count(&daily).Error; e != nil {
			return e
		}
		if daily >= 20 {
			return port.ErrAIRateLimited
		}
		var count int64
		if e = tx.Model(&entity.AIEntryMessage{}).Where("session_id = ? AND role = ?", id, "user").Count(&count).Error; e != nil {
			return e
		}
		if count >= 20 {
			return port.ErrAISessionFull
		}
		token = uuid.NewString()
		until := time.Now().Add(2 * time.Minute)
		request := entity.AIEntryRequest{SessionID: id, RequestID: requestID, Hash: hash, Token: token}
		if e = tx.Create(&request).Error; e != nil {
			return e
		}
		if e = tx.Create(&entity.AIEntryMessage{ID: token, SessionID: id, Role: "user", Content: text}).Error; e != nil {
			return e
		}
		if e = tx.Model(&s).Updates(map[string]any{"request_token": token, "processing_until": until, "error": ""}).Error; e != nil {
			return e
		}
		started = true
		return nil
	})
	return
}
func (r *AIEntryPostgresRepository) FinishMessage(ctx context.Context, owner, id, token string, out entity.AIExtractOutput) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var s entity.AIEntrySession
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND owner_id = ?", id, owner).First(&s).Error; err != nil {
			return err
		}
		if s.RequestToken != token || s.ProcessingUntil == nil || !time.Now().Before(*s.ProcessingUntil) {
			return port.ErrAIConflict
		}
		if len(out.Drafts) > 30 {
			return fmt.Errorf("%w: quá nhiều đề xuất", port.ErrAIInvalid)
		}
		for _, d := range out.Drafts {
			if d.Questions == nil {
				d.Questions = []string{}
			}
			p := entity.AIEntryProposal{ID: uuid.NewString(), SessionID: id, OwnerID: owner, Version: 1, Status: entity.AIProposalPending, Draft: d.AIEntryDraft, Questions: d.Questions}
			if err := tx.Create(&p).Error; err != nil {
				return err
			}
		}
		if err := tx.Create(&entity.AIEntryMessage{ID: uuid.NewString(), SessionID: id, Role: "assistant", Content: out.Reply, SourceText: out.SourceText}).Error; err != nil {
			return err
		}
		return tx.Model(&s).Updates(map[string]any{"request_token": "", "processing_until": nil, "error": ""}).Error
	})
}
func (r *AIEntryPostgresRepository) FailMessage(ctx context.Context, owner, id, token, message, sourceText string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var s entity.AIEntrySession
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND owner_id = ? AND request_token = ?", id, owner, token).First(&s).Error; err != nil {
			return err
		}
		if len(sourceText) > 65536 {
			return port.ErrAIInvalid
		}
		if sourceText != "" {
			if err := tx.Create(&entity.AIEntryMessage{ID: uuid.NewString(), SessionID: id, Role: "assistant", Content: "Đã đọc được văn bản, nhưng AI chưa tạo được đề xuất. Bạn có thể tiếp tục với văn bản đã đọc.", SourceText: sourceText}).Error; err != nil {
				return err
			}
		}
		return tx.Model(&s).Updates(map[string]any{"request_token": "", "processing_until": nil, "error": message}).Error
	})
}
func (r *AIEntryPostgresRepository) EditProposal(ctx context.Context, owner, id string, version int, draft entity.AIEntryDraft) (*entity.AIEntryProposal, error) {
	var p entity.AIEntryProposal
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND owner_id = ?", id, owner).First(&p).Error; err != nil {
			return err
		}
		if p.Version != version || p.Status != entity.AIProposalPending {
			return port.ErrAIConflict
		}
		if len(draft.Note) > 8000 || len(draft.WalletID) > 100 || len(draft.OccurredAt) > 100 || draft.Amount < 0 || draft.Amount > 9007199254740991 {
			return port.ErrAIInvalid
		}
		if draft.Type != "income" && draft.Type != "expense" && draft.Type != "transfer" && draft.Type != "unknown" {
			return port.ErrAIInvalid
		}
		if draft.CategoryID != nil && strings.TrimSpace(*draft.CategoryID) == "" {
			draft.CategoryID = nil
		}
		p.Draft = draft
		p.Version++
		p.Questions = []string{}
		return tx.Select("draft", "version", "questions", "updated_at").Save(&p).Error
	})
	return &p, err
}
func (r *AIEntryPostgresRepository) DecideProposal(ctx context.Context, owner, id string, version int, approve bool) (*entity.AIEntryProposal, error) {
	var p entity.AIEntryProposal
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND owner_id = ?", id, owner).First(&p).Error; err != nil {
			return err
		}
		if (approve && p.Status == entity.AIProposalApproved) || (!approve && p.Status == entity.AIProposalRejected) {
			return nil
		}
		if p.Status != entity.AIProposalPending || p.Version != version {
			return port.ErrAIConflict
		}
		if approve {
			var w entity.Wallet
			if err := tx.Clauses(clause.Locking{Strength: "SHARE"}).Where("id = ? AND owner_id = ?", p.Draft.WalletID, owner).First(&w).Error; err != nil {
				return fmt.Errorf("%w: ví không còn hợp lệ", port.ErrAIInvalid)
			}
			var category *entity.Category
			if p.Draft.CategoryID != nil && *p.Draft.CategoryID != "" {
				category = &entity.Category{}
				if err := tx.Clauses(clause.Locking{Strength: "SHARE"}).Where("id = ? AND (owner_id = ? OR is_system = true)", *p.Draft.CategoryID, owner).First(category).Error; err != nil {
					return fmt.Errorf("%w: nhóm không còn hợp lệ", port.ErrAIInvalid)
				}
				if err := tx.Model(&entity.CategoryWallet{}).Joins("JOIN wallets ON wallets.id = category_wallets.wallet_id").Where("category_id = ? AND wallets.owner_id = ?", category.ID, owner).Pluck("wallet_id", &category.WalletIDs).Error; err != nil {
					return err
				}
			}
			if err := usecase.ValidateAIEntryDraft(p.Draft, w, category); err != nil {
				return fmt.Errorf("%w: %s", port.ErrAIInvalid, err)
			}
			occurred, _ := time.Parse(time.RFC3339, p.Draft.OccurredAt)
			note := strings.TrimSpace(p.Draft.Note)
			row := entity.Transaction{ID: uuid.NewString(), OwnerID: owner, WalletID: w.ID, CategoryID: p.Draft.CategoryID, Type: p.Draft.Type, Amount: p.Draft.Amount, OccurredAt: occurred.UTC(), Note: &note, IncludedInReports: p.Draft.IncludedInReports}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
			// GORM default:true must not override an explicit false from the review.
			if !p.Draft.IncludedInReports {
				if err := tx.Model(&row).Update("included_in_reports", false).Error; err != nil {
					return err
				}
			}
			p.TransactionID = &row.ID
			p.Status = entity.AIProposalApproved
		} else {
			p.Status = entity.AIProposalRejected
		}
		p.Version++
		return tx.Select("status", "version", "transaction_id", "updated_at").Save(&p).Error
	})
	return &p, err
}
