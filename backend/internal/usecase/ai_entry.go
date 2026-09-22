package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	port "github.com/mypocket/backend/internal/repository"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	ErrAIUnavailable = errors.New("chưa cấu hình AI; cần AI_BASE_URL, AI_MODEL và AI_API_KEY ở backend")
	ErrAIProvider    = errors.New("chưa xử lý được nội dung; không giao dịch nào được lưu. Bạn có thể thử lại, dịch vụ có thể tính thêm lượt xử lý")
	ErrAIStorage     = errors.New("không lưu được chứng từ; OCR và AI chưa được gọi, không phát sinh lượt xử lý")
)

type AIEntryExtractor interface {
	Configured() bool
	OCRConfigured() bool
	ValidateFiles([]entity.AIImage) error
	Extract(context.Context, entity.AIExtractInput) (entity.AIExtractOutput, error)
}
type AIEntryService struct {
	Entries    port.AIEntryRepository
	Wallets    port.WalletRepository
	Categories port.CategoryRepository
	Extractor  AIEntryExtractor
	Storage    AttachmentStorage
	Users      port.UserRepository
}

type AttachmentStorage interface {
	Key(owner, process, attachment, filename string) (string, error)
	Put(context.Context, string, string, []byte) error
	Delete(context.Context, string) error
	SignedGet(context.Context, string) (string, error)
}
type AIEntryMessageInput struct {
	RequestID string           `json:"request_id"`
	Text      string           `json:"text"`
	Timezone  string           `json:"timezone"`
	Images    []entity.AIImage `json:"images,omitempty"`
}

func (s *AIEntryService) Send(ctx context.Context, owner, id string, in AIEntryMessageInput) (*entity.AIEntrySession, error) {
	in.Text = strings.TrimSpace(in.Text)
	if err := s.resolveTimezone(ctx, owner, &in); err != nil {
		return nil, err
	}
	if err := s.validate(in); err != nil {
		return nil, err
	}
	if _, err := s.Entries.Session(ctx, owner, id); err != nil {
		return nil, err
	}
	wallets, err := s.Wallets.List(owner)
	if err != nil {
		return nil, err
	}
	categories, err := s.Categories.ListVisible(owner)
	if err != nil {
		return nil, err
	}
	payload, _ := json.Marshal(in)
	hash := sha256.Sum256(payload)
	token, started, err := s.Entries.BeginMessage(ctx, owner, id, in.RequestID, hex.EncodeToString(hash[:]), in.Text)
	if err != nil {
		return nil, err
	}
	if !started {
		return s.Entries.Session(ctx, owner, id)
	}
	if len(in.Images) > 0 {
		if s.Storage == nil {
			s.fail(owner, id, token, "")
			return nil, ErrAIUnavailable
		}
		prepared, attachments, err := s.storeForOCR(ctx, owner, id, in.Images)
		if err != nil {
			s.failWith(owner, id, token, ErrAIStorage.Error(), "")
			return nil, ErrAIStorage
		}
		if err := s.Entries.CreateAttachments(ctx, attachments); err != nil {
			for _, attachment := range attachments {
				_ = s.Storage.Delete(context.WithoutCancel(ctx), attachment.ObjectKey)
			}
			s.failWith(owner, id, token, ErrAIStorage.Error(), "")
			return nil, ErrAIStorage
		}
		in.Images = prepared
	}
	providerCtx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()
	out, err := s.Extractor.Extract(providerCtx, entity.AIExtractInput{Text: in.Text, Timezone: in.Timezone, Now: time.Now(), Wallets: wallets, Categories: categories, Images: in.Images})
	if len(in.Images) > 0 {
		if updateErr := s.Entries.UpdateAttachmentOCR(ctx, owner, id, out.AttachmentTexts, out.OCRComplete); updateErr != nil {
			s.fail(owner, id, token, "")
			return nil, ErrAIProvider
		}
	}
	if err != nil {
		s.fail(owner, id, token, boundedAIContext(out.SourceText, 64000))
		return nil, ErrAIProvider
	}
	if len(out.Drafts) > 30 || len(out.Reply) > 16000 || len(out.SourceText) > 200000 {
		s.fail(owner, id, token, "")
		return nil, ErrAIProvider
	}
	for i := range out.Drafts {
		d := &out.Drafts[i]
		var wallet entity.Wallet
		var category *entity.Category
		for _, w := range wallets {
			if w.ID == d.WalletID {
				wallet = w
				break
			}
		}
		if wallet.ID == "" {
			d.WalletID = ""
		}
		if d.CategoryID != nil {
			for j := range categories {
				if categories[j].ID == *d.CategoryID {
					category = &categories[j]
					break
				}
			}
			if category == nil {
				d.CategoryID = nil
			}
		}
		if d.Type != "income" && d.Type != "expense" && d.Type != "transfer" {
			d.Type = "unknown"
		}
		if d.Amount < 0 || d.Amount > 9007199254740991 {
			d.Amount = 0
		}
		if _, e := time.Parse(time.RFC3339, d.OccurredAt); e != nil {
			d.OccurredAt = ""
		}
		if d.Questions == nil {
			d.Questions = []string{}
		}
		if e := ValidateAIEntryDraft(d.AIEntryDraft, wallet, category); e != nil {
			d.Questions = append(d.Questions, e.Error())
		}
	}
	if err = s.Entries.FinishMessage(ctx, owner, id, token, out); err != nil {
		s.fail(owner, id, token, boundedAIContext(out.SourceText, 64000))
		return nil, err
	}
	return s.Entries.Session(ctx, owner, id)
}

func (s *AIEntryService) storeForOCR(ctx context.Context, owner, process string, files []entity.AIImage) ([]entity.AIImage, []entity.AIEntryAttachment, error) {
	prepared := make([]entity.AIImage, 0, len(files))
	attachments := make([]entity.AIEntryAttachment, 0, len(files))
	keys := make([]string, 0, len(files))
	rollback := func() {
		for _, key := range keys {
			_ = s.Storage.Delete(context.WithoutCancel(ctx), key)
		}
	}
	for _, file := range files {
		data, err := base64.StdEncoding.Strict().DecodeString(file.Base64)
		if err != nil {
			rollback()
			return nil, nil, port.ErrAIInvalid
		}
		id := uuid.NewString()
		ext := map[string]string{"application/pdf": "pdf", "image/jpeg": "jpg", "image/png": "png"}[file.MIMEType]
		key, err := s.Storage.Key(owner, process, id, "receipt."+ext)
		if err != nil {
			rollback()
			return nil, nil, err
		}
		keys = append(keys, key)
		if err = s.Storage.Put(ctx, key, file.MIMEType, data); err != nil {
			rollback()
			return nil, nil, err
		}
		readURL, err := s.Storage.SignedGet(ctx, key)
		if err != nil {
			rollback()
			return nil, nil, err
		}
		digest := sha256.Sum256(data)
		attachments = append(attachments, entity.AIEntryAttachment{ID: id, OwnerID: owner, ProcessID: process, ObjectKey: key, Filename: filepath.Base(file.Name), MIMEType: file.MIMEType, SizeBytes: int64(len(data)), SHA256: hex.EncodeToString(digest[:]), OCRStatus: "pending", DeleteAfter: time.Now().Add(24 * time.Hour)})
		prepared = append(prepared, entity.AIImage{Name: file.Name, MIMEType: file.MIMEType, SourceURL: readURL, AttachmentID: id})
	}
	sort.Slice(attachments, func(i, j int) bool { return attachments[i].ID < attachments[j].ID })
	sort.Slice(prepared, func(i, j int) bool { return prepared[i].AttachmentID < prepared[j].AttachmentID })
	return prepared, attachments, nil
}

func (s *AIEntryService) Process(ctx context.Context, owner string, in AIEntryMessageInput) (*entity.AIEntrySession, error) {
	in.Text = strings.TrimSpace(in.Text)
	if err := s.resolveTimezone(ctx, owner, &in); err != nil {
		return nil, err
	}
	if err := s.validate(in); err != nil {
		return nil, err
	}
	if err := s.Entries.CreateProcess(ctx, owner, in.RequestID); err != nil {
		return nil, err
	}
	return s.Send(ctx, owner, in.RequestID, in)
}

func (s *AIEntryService) resolveTimezone(ctx context.Context, owner string, input *AIEntryMessageInput) error {
	if s.Users != nil {
		user, err := s.Users.FindByID(owner)
		if err != nil {
			return err
		}
		input.Timezone = user.Timezone
	}
	if strings.TrimSpace(input.Timezone) == "" {
		input.Timezone = "Asia/Ho_Chi_Minh"
	}
	if input.Timezone == "Local" {
		return fmt.Errorf("%w: múi giờ không hợp lệ", port.ErrAIInvalid)
	}
	if _, err := time.LoadLocation(input.Timezone); err != nil {
		return fmt.Errorf("%w: múi giờ không hợp lệ", port.ErrAIInvalid)
	}
	return nil
}

func (s *AIEntryService) validate(in AIEntryMessageInput) error {
	if !s.Extractor.Configured() {
		return ErrAIUnavailable
	}
	if _, err := uuid.Parse(in.RequestID); err != nil {
		return fmt.Errorf("%w: thiếu mã yêu cầu", port.ErrAIInvalid)
	}
	if _, err := time.LoadLocation(in.Timezone); err != nil {
		return fmt.Errorf("%w: múi giờ không hợp lệ", port.ErrAIInvalid)
	}
	if len(in.Text) > 16000 || len(in.Images) > 3 || (in.Text == "" && len(in.Images) == 0) {
		return fmt.Errorf("%w: cần nội dung hoặc tối đa 3 tệp", port.ErrAIInvalid)
	}
	if len(in.Images) > 0 && !s.Extractor.OCRConfigured() {
		return fmt.Errorf("%w: OCR chưa được cấu hình", ErrAIUnavailable)
	}
	if len(in.Images) > 0 && s.Extractor.ValidateFiles(in.Images) != nil {
		return fmt.Errorf("%w: tệp không hợp lệ", port.ErrAIInvalid)
	}
	for _, image := range in.Images {
		if len(image.Base64) > 7*1024*1024 || len(image.Name) > 255 {
			return port.ErrAIInvalid
		}
	}
	if len(in.Images) > 0 && s.Storage == nil {
		return fmt.Errorf("%w: lưu trữ chứng từ chưa được cấu hình", ErrAIUnavailable)
	}
	return nil
}

func boundedAIContext(text string, limit int) string {
	if len(text) <= limit {
		return text
	}
	text = text[:limit]
	for !utf8.ValidString(text) {
		text = text[:len(text)-1]
	}
	return text + "\n[excerpt; earlier content omitted]"
}
func (s *AIEntryService) fail(owner, id, token, sourceText string) {
	s.failWith(owner, id, token, ErrAIProvider.Error(), sourceText)
}
func (s *AIEntryService) failWith(owner, id, token, message, sourceText string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = s.Entries.FailMessage(ctx, owner, id, token, message, sourceText)
}
