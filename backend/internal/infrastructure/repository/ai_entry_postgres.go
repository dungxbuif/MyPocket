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
func (r *AIEntryPostgresRepository) CreateProcess(ctx context.Context, owner, requestID string) error {
	process := entity.AIEntrySession{ID: requestID, OwnerID: owner, Proposals: []entity.AIEntryProposal{}}
	if err := r.db.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, DoNothing: true}).Create(&process).Error; err != nil {
		return err
	}
	var existing entity.AIEntrySession
	return r.db.WithContext(ctx).Where("id = ? AND owner_id = ?", requestID, owner).First(&existing).Error
}
func (r *AIEntryPostgresRepository) CreateAttachments(ctx context.Context, attachments []entity.AIEntryAttachment) error {
	if len(attachments) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&attachments).Error
}
func (r *AIEntryPostgresRepository) UpdateAttachmentOCR(ctx context.Context, owner, process string, texts []string, completed bool) error {
	var attachments []entity.AIEntryAttachment
	if err := r.db.WithContext(ctx).Where("owner_id = ? AND session_id = ?", owner, process).Order("id").Find(&attachments).Error; err != nil {
		return err
	}
	if len(attachments) == 0 || len(texts) > len(attachments) || (completed && len(attachments) != len(texts)) {
		return port.ErrAIConflict
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, attachment := range attachments {
			text := ""
			status := "failed"
			if i < len(texts) {
				text = texts[i]
				status = "completed"
			}
			if len(text) > 65536 {
				return port.ErrAIInvalid
			}
			if err := tx.Model(&attachment).Updates(map[string]any{"ocr_status": status, "ocr_text": text}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
func (r *AIEntryPostgresRepository) AttachmentForTransaction(ctx context.Context, owner, transactionID, attachmentID string) (*entity.AIEntryAttachment, error) {
	var attachment entity.AIEntryAttachment
	err := r.db.WithContext(ctx).Joins("JOIN transaction_attachment_links ON transaction_attachment_links.attachment_id = transaction_attachments.id").Joins("JOIN transactions ON transactions.id = transaction_attachment_links.transaction_id").Where("transactions.owner_id = ? AND transactions.id = ? AND transaction_attachments.id = ? AND transaction_attachments.owner_id = ?", owner, transactionID, attachmentID, owner).First(&attachment).Error
	return &attachment, err
}
func (r *AIEntryPostgresRepository) ClaimExpiredAttachments(ctx context.Context, before, staleClaim time.Time, limit int) ([]entity.AIEntryAttachment, error) {
	if limit <= 0 || limit > 500 {
		return nil, port.ErrAIInvalid
	}
	var claimed []entity.AIEntryAttachment
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var rows []entity.AIEntryAttachment
		query := tx.Clauses(clause.Locking{Strength: "UPDATE", Table: clause.Table{Name: "transaction_attachments"}, Options: "SKIP LOCKED"}).Where("NOT EXISTS (SELECT 1 FROM transaction_attachment_links WHERE transaction_attachment_links.attachment_id = transaction_attachments.id) AND transaction_attachments.delete_after <= ? AND (transaction_attachments.ocr_status IN ? OR (transaction_attachments.ocr_status = 'deleting' AND transaction_attachments.updated_at <= ?))", before, []string{"pending", "completed", "failed", "delete_failed"}, staleClaim).Order("transaction_attachments.delete_after").Limit(limit)
		if err := query.Find(&rows).Error; err != nil {
			return err
		}
		for _, row := range rows {
			if err := tx.Model(&row).Update("ocr_status", "deleting").Error; err != nil {
				return err
			}
			row.OCRStatus = "deleting"
			claimed = append(claimed, row)
		}
		return nil
	})
	return claimed, err
}
func (r *AIEntryPostgresRepository) SetAttachmentDeleteStatus(ctx context.Context, id, status string) error {
	if status != "deleted" && status != "delete_failed" {
		return port.ErrAIInvalid
	}
	result := r.db.WithContext(ctx).Model(&entity.AIEntryAttachment{}).Where("id = ? AND ocr_status = 'deleting'", id).Update("ocr_status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return port.ErrAIConflict
	}
	return nil
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
		// One-shot processes deliberately do not load or expose conversation history.
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
		// Serialize provider work start across separate sessions, without blocking FK reads.
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
		token = uuid.NewString()
		until := time.Now().Add(2 * time.Minute)
		request := entity.AIEntryRequest{SessionID: id, RequestID: requestID, Hash: hash, Token: token}
		if e = tx.Create(&request).Error; e != nil {
			return e
		}
		_ = text // request content is not retained as conversation history.
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
		_ = sourceText // OCR evidence is not stored in a conversation message.
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
			if approve && p.Status == entity.AIProposalApproved && p.TransactionID != nil {
				return tx.Model(&entity.AIEntryAttachmentLink{}).Joins("JOIN transaction_attachments ON transaction_attachments.id = transaction_attachment_links.attachment_id").Where("transaction_attachment_links.transaction_id = ? AND transaction_attachments.owner_id = ?", *p.TransactionID, owner).Order("transaction_attachment_links.attachment_id").Pluck("transaction_attachment_links.attachment_id", &p.AttachmentIDs).Error
			}
			return nil
		}
		if p.Status != entity.AIProposalPending || p.Version != version {
			return port.ErrAIConflict
		}
		if approve {
			var deleting int64
			if err := tx.Model(&entity.AIEntryAttachment{}).Where("session_id = ? AND owner_id = ? AND ocr_status IN ?", p.SessionID, owner, []string{"deleting", "deleted"}).Count(&deleting).Error; err != nil {
				return err
			}
			if deleting > 0 {
				return port.ErrAIConflict
			}
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
			if p.Draft.JarID != nil {
				if p.Draft.Type != entity.TransactionTypeExpense || (category != nil && category.SystemKey != nil && *category.SystemKey == "expense_transfer_out") {
					return fmt.Errorf("%w: chỉ khoản chi thường mới được gắn hũ", port.ErrAIInvalid)
				}
				if err := validateJarAssignment(tx, owner, p.Draft.JarID, occurred.UTC(), nil); err != nil {
					return fmt.Errorf("%w: hũ không thuộc cấu hình tháng của giao dịch", port.ErrAIInvalid)
				}
			}
			note := strings.TrimSpace(p.Draft.Note)
			row := entity.Transaction{ID: uuid.NewString(), OwnerID: owner, WalletID: w.ID, CategoryID: p.Draft.CategoryID, JarID: p.Draft.JarID, Type: p.Draft.Type, Amount: p.Draft.Amount, OccurredAt: occurred.UTC(), Note: &note, IncludedInReports: p.Draft.IncludedInReports}
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
			var attachments []entity.AIEntryAttachment
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("session_id = ? AND owner_id = ?", p.SessionID, owner).Find(&attachments).Error; err != nil {
				return err
			}
			links := make([]entity.AIEntryAttachmentLink, 0, len(attachments))
			for _, attachment := range attachments {
				if attachment.OCRStatus == "deleting" || attachment.OCRStatus == "deleted" {
					return port.ErrAIConflict
				}
				links = append(links, entity.AIEntryAttachmentLink{TransactionID: row.ID, AttachmentID: attachment.ID})
				p.AttachmentIDs = append(p.AttachmentIDs, attachment.ID)
			}
			if len(links) > 0 {
				if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&links).Error; err != nil {
					return err
				}
			}
		} else {
			p.Status = entity.AIProposalRejected
		}
		p.Version++
		return tx.Select("status", "version", "transaction_id", "updated_at").Save(&p).Error
	})
	return &p, err
}
