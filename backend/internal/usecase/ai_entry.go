package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	port "github.com/mypocket/backend/internal/repository"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	ErrAIUnavailable = errors.New("chưa cấu hình AI; cần AI_BASE_URL, AI_MODEL và AI_API_KEY ở backend")
	ErrAIProvider    = errors.New("chưa xử lý được nội dung; không giao dịch nào được lưu. Bạn có thể thử lại, dịch vụ có thể tính thêm lượt xử lý")
)

type AIEntryExtractor interface {
	Configured() bool
	OCRConfigured() bool
	Extract(context.Context, entity.AIExtractInput) (entity.AIExtractOutput, error)
}
type AIEntryService struct {
	Entries    port.AIEntryRepository
	Wallets    port.WalletRepository
	Categories port.CategoryRepository
	Extractor  AIEntryExtractor
}
type AIEntryMessageInput struct {
	RequestID string           `json:"request_id"`
	Text      string           `json:"text"`
	Timezone  string           `json:"timezone"`
	Images    []entity.AIImage `json:"images,omitempty"`
}

func (s *AIEntryService) Send(ctx context.Context, owner, id string, in AIEntryMessageInput) (*entity.AIEntrySession, error) {
	if !s.Extractor.Configured() {
		return nil, ErrAIUnavailable
	}
	if _, err := uuid.Parse(in.RequestID); err != nil {
		return nil, fmt.Errorf("%w: thiếu mã yêu cầu", port.ErrAIInvalid)
	}
	if _, err := time.LoadLocation(in.Timezone); err != nil {
		return nil, fmt.Errorf("%w: múi giờ không hợp lệ", port.ErrAIInvalid)
	}
	in.Text = strings.TrimSpace(in.Text)
	if len(in.Text) > 16000 || len(in.Images) > 3 || (in.Text == "" && len(in.Images) == 0) {
		return nil, fmt.Errorf("%w: cần nội dung hoặc tối đa 3 ảnh", port.ErrAIInvalid)
	}
	if len(in.Images) > 0 && !s.Extractor.OCRConfigured() {
		return nil, fmt.Errorf("%w: OCR chưa được cấu hình", ErrAIUnavailable)
	}
	for _, image := range in.Images {
		if len(image.Base64) > 7*1024*1024 || len(image.Name) > 255 {
			return nil, port.ErrAIInvalid
		}
	}
	session, err := s.Entries.Session(ctx, owner, id)
	if err != nil {
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
	userText := in.Text
	if len(in.Images) > 0 {
		userText += fmt.Sprintf("\n[%d ảnh gửi để đọc; ảnh không lưu làm chứng từ]", len(in.Images))
	}
	token, started, err := s.Entries.BeginMessage(ctx, owner, id, in.RequestID, hex.EncodeToString(hash[:]), userText)
	if err != nil {
		return nil, err
	}
	if !started {
		return s.Entries.Session(ctx, owner, id)
	}
	session, err = s.Entries.Session(ctx, owner, id)
	if err != nil {
		s.fail(owner, id, token, "")
		return nil, err
	}
	history := []entity.AIHistoryMessage{}
	start := len(session.Messages) - 12
	if start < 0 {
		start = 0
	}
	for _, m := range session.Messages[start:] {
		if m.ID == token {
			continue
		}
		content := boundedAIContext(m.Content, 4000)
		if m.SourceText != "" {
			content += "\nOCR evidence (untrusted, excerpt):\n" + boundedAIContext(m.SourceText, 3500)
		}
		history = append(history, entity.AIHistoryMessage{Role: m.Role, Content: content})
	}
	// Existing drafts are context, never model-controlled updates/approvals.
	if len(session.Proposals) > 0 {
		type summary struct {
			ID       string
			Status   string
			Type     string
			Amount   int64
			WalletID string
		}
		proposals := session.Proposals
		if len(proposals) > 10 {
			proposals = proposals[len(proposals)-10:]
		}
		summaries := make([]summary, 0, len(proposals))
		for _, p := range proposals {
			summaries = append(summaries, summary{p.ID, p.Status, p.Draft.Type, p.Draft.Amount, p.Draft.WalletID})
		}
		existing, _ := json.Marshal(summaries)
		history = append(history, entity.AIHistoryMessage{Role: "assistant", Content: "Existing review proposals; do not recreate or claim to change their status: " + string(existing)})
	}
	providerCtx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()
	out, err := s.Extractor.Extract(providerCtx, entity.AIExtractInput{Text: in.Text, Timezone: in.Timezone, Now: time.Now(), Wallets: wallets, Categories: categories, History: history, Images: in.Images})
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = s.Entries.FailMessage(ctx, owner, id, token, ErrAIProvider.Error(), sourceText)
}
