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
	"github.com/mypocket/backend/internal/infrastructure/financialsource"
	port "github.com/mypocket/backend/internal/repository"
	"log/slog"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

var (
	ErrAIUnavailable = errors.New("chưa cấu hình AI; cần AI_BASE_URL, AI_MODEL và AI_API_KEY ở backend")
	ErrAIProvider    = errors.New("chưa xử lý được nội dung; không giao dịch nào được lưu. Bạn có thể thử lại, dịch vụ có thể tính thêm lượt xử lý")
	ErrAIStorage     = errors.New("không lưu được chứng từ; OCR và AI chưa được gọi, không phát sinh lượt xử lý")
)

const (
	aiProviderTimeout          = 210 * time.Second
	aiAttachmentStorageTimeout = 45 * time.Second
	aiAttachmentStorageWorkers = 4
	xlsxMIMEType               = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
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

type AttachmentReader interface {
	Read(context.Context, string, int64) ([]byte, error)
}
type AIEntryMessageInput struct {
	RequestID string           `json:"request_id"`
	Text      string           `json:"text"`
	Timezone  string           `json:"timezone"`
	Images    []entity.AIImage `json:"images,omitempty"`
}

// StoreAcceptedAttachments persists private objects before a durable advisor
// task is returned to the client. Processing is intentionally separate: the
// worker later reads these objects and calls ProcessStored.
func (s *AIEntryService) StoreAcceptedAttachments(ctx context.Context, owner, sessionID string, files []entity.AIImage) ([]entity.AIEntryAttachment, error) {
	if len(files) == 0 {
		return []entity.AIEntryAttachment{}, nil
	}
	if s == nil || s.Storage == nil || s.Entries == nil {
		return nil, ErrAIStorage
	}
	prepared, attachments, err := s.storeForOCR(ctx, owner, sessionID, files)
	_ = prepared
	if err != nil {
		return nil, ErrAIStorage
	}
	if err := s.Entries.CreateAttachments(ctx, attachments); err != nil {
		for _, attachment := range attachments {
			_ = s.Storage.Delete(context.WithoutCancel(ctx), attachment.ObjectKey)
		}
		return nil, ErrAIStorage
	}
	return attachments, nil
}

func (s *AIEntryService) Send(ctx context.Context, owner, id string, in AIEntryMessageInput) (*entity.AIEntrySession, error) {
	return s.send(ctx, owner, id, in, nil)
}

// ProcessStored executes the same extraction/normalization/persistence path as
// the legacy endpoint after a durable task has already stored its attachments.
// It never calls storeForOCR, so a worker restart cannot create duplicate
// private objects for the same accepted task.
func (s *AIEntryService) ProcessStored(ctx context.Context, owner, sessionID string, in AIEntryMessageInput) (*entity.AIEntrySession, error) {
	if s == nil || s.Entries == nil {
		return nil, ErrAIStorage
	}
	attachmentsRepo, ok := s.Entries.(port.AIEntryStoredInputRepository)
	if !ok {
		return nil, ErrAIStorage
	}
	attachments, err := attachmentsRepo.ListAttachments(ctx, owner, sessionID)
	if err != nil {
		return nil, err
	}
	images, err := s.loadStoredImages(ctx, attachments)
	if err != nil {
		return nil, err
	}
	in.Images = images
	// The accepted task stores the timezone that was part of the idempotent
	// payload. Do not silently replace it with a later profile edit while a
	// restarted worker is processing the task.
	workerService := *s
	workerService.Users = nil
	return workerService.send(ctx, owner, sessionID, in, attachments)
}

func (s *AIEntryService) send(ctx context.Context, owner, id string, in AIEntryMessageInput, storedAttachments []entity.AIEntryAttachment) (*entity.AIEntrySession, error) {
	slog.Info("AI entry processing started", "process_id", id, "text_bytes", len(in.Text), "image_count", len(in.Images))
	in.Text = strings.TrimSpace(in.Text)
	if err := s.resolveTimezone(ctx, owner, &in); err != nil {
		return nil, err
	}
	if err := s.validate(in); err != nil {
		return nil, err
	}
	structuredRows, err := parseAIEntryWorkbooks(in.Images, in.Timezone)
	if err != nil {
		return nil, fmt.Errorf("%w: tệp XLSX không hợp lệ", port.ErrAIInvalid)
	}
	if storedAttachments != nil {
		// Workbooks are parsed locally into structured rows and are not sent to
		// the OCR/model image lane. This mirrors the legacy upload path after a
		// restart without creating a second private-object upload.
		ocrImages := make([]entity.AIImage, 0, len(in.Images))
		for _, image := range in.Images {
			if image.MIMEType != xlsxMIMEType {
				ocrImages = append(ocrImages, image)
			}
		}
		in.Images = ocrImages
	}
	accountNow := time.Now()
	if location, locationErr := time.LoadLocation(in.Timezone); locationErr == nil {
		accountNow = accountNow.In(location)
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
	activeCategories := make([]entity.Category, 0, len(categories))
	for _, category := range categories {
		if category.IsActive {
			activeCategories = append(activeCategories, category)
		}
	}
	categories = activeCategories
	slog.Info("AI entry catalogs loaded", "process_id", id, "wallet_count", len(wallets), "category_count", len(categories))
	payload, _ := json.Marshal(in)
	hash := sha256.Sum256(payload)
	token, started, err := s.Entries.BeginMessage(ctx, owner, id, in.RequestID, hex.EncodeToString(hash[:]), in.Text)
	if err != nil {
		return nil, err
	}
	slog.Info("AI entry idempotency resolved", "process_id", id, "started", started)
	if !started {
		slog.Info("AI entry replay returned", "process_id", id)
		return s.Entries.Session(ctx, owner, id)
	}
	var ocrAttachments []entity.AIEntryAttachment
	if len(in.Images) > 0 {
		if storedAttachments != nil {
			ocrAttachments = append([]entity.AIEntryAttachment(nil), storedAttachments...)
		} else {
			if s.Storage == nil {
				slog.Warn("AI entry storage unavailable", "process_id", id, "stage", "storage", "code", "storage_not_configured")
				s.fail(owner, id, token, "")
				return nil, ErrAIUnavailable
			}
			// Object storage is a separate network hop from OCR/model execution.
			// Bound it explicitly so a stalled S3-compatible endpoint cannot consume
			// the entire AI request budget before the provider is even called.
			storageCtx, storageCancel := context.WithTimeout(ctx, aiAttachmentStorageTimeout)
			prepared, attachments, err := s.storeForOCR(storageCtx, owner, id, in.Images)
			storageCancel()
			if err != nil {
				slog.Warn("AI entry storage failed", "process_id", id, "stage", "storage", "code", "attachment_put_failed", "file_count", len(in.Images))
				s.failWith(owner, id, token, ErrAIStorage.Error(), "storage_error", "")
				return nil, ErrAIStorage
			}
			if err := s.Entries.CreateAttachments(ctx, attachments); err != nil {
				slog.Warn("AI entry attachment metadata failed", "process_id", id, "stage", "storage", "code", "attachment_metadata_failed", "file_count", len(attachments))
				for _, attachment := range attachments {
					_ = s.Storage.Delete(context.WithoutCancel(ctx), attachment.ObjectKey)
				}
				s.failWith(owner, id, token, ErrAIStorage.Error(), "storage_error", "")
				return nil, ErrAIStorage
			}
			in.Images = prepared
			ocrAttachments = attachments
		}
	}
	providerCtx, cancel := context.WithTimeout(ctx, aiProviderTimeout)
	defer cancel()
	out, err := s.Extractor.Extract(providerCtx, entity.AIExtractInput{Text: in.Text, Instruction: in.Text, Timezone: in.Timezone, Now: accountNow, Wallets: wallets, Categories: categories, Images: in.Images, StructuredRows: structuredRows})
	if len(in.Images) > 0 {
		ocrTexts := alignOCRTexts(in.Images, ocrAttachments, out.AttachmentTexts)
		if updateErr := s.Entries.UpdateAttachmentOCR(ctx, owner, id, ocrTexts, out.OCRComplete); updateErr != nil {
			slog.Warn("AI entry OCR metadata failed", "process_id", id, "stage", "persistence", "code", "ocr_metadata_failed", "attachment_count", len(out.AttachmentTexts))
			s.fail(owner, id, token, "")
			return nil, ErrAIProvider
		}
	}
	if err != nil {
		s.failWithDiagnostic(owner, id, token, err, boundedAIContext(out.SourceText, 64000))
		return nil, ErrAIProvider
	}
	if len(structuredRows) > 0 {
		out.Drafts, err = mergeStructuredAIDrafts(structuredRows, out.Drafts)
		if err != nil {
			s.fail(owner, id, token, "")
			return nil, ErrAIProvider
		}
	}
	if len(out.SourceText) > 200000 {
		s.fail(owner, id, token, "")
		return nil, ErrAIProvider
	}
	invalidDrafts := 0
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
			// Wallet selection is mandatory for reviewable drafts. The model may
			// still omit it despite the contract; use the first supplied wallet
			// as a deterministic fallback and disclose multi-wallet inference.
			if len(wallets) > 0 {
				modelWalletID := d.WalletID
				wallet = wallets[0]
				d.WalletID = wallet.ID
				if modelWalletID != "" || len(wallets) > 1 {
					d.Questions = append(d.Questions, "Ví được suy đoán từ danh mục hiện có; hãy kiểm tra lại trước khi lưu.")
				}
			} else {
				d.WalletID = ""
			}
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
		if category == nil && (d.Type == entity.TransactionTypeIncome || d.Type == entity.TransactionTypeExpense) {
			if fallback := defaultAIEntryCategory(d.AIEntryDraft, wallet, categories); fallback != nil {
				d.CategoryID = &fallback.ID
				category = fallback
				appendAIQuestion(d, "Nhóm được chọn mặc định từ loại giao dịch và ví; hãy kiểm tra lại trước khi lưu.")
			}
		}
		if d.Amount < 0 || d.Amount > 9007199254740991 {
			d.Amount = 0
		}
		if d.SourceID == "" && (d.SourceCurrency == "foreign" || d.SourceCurrency == "unknown" || hasForeignCurrencyCue(d.Note, d.Questions)) {
			d.Amount = 0
			appendAIQuestion(d, "Số tiền không được xác nhận là VND; hãy nhập số VND đã kiểm tra trước khi lưu.")
		}
		normalizedOccurredAt, parsedOccurredAt, dateOnly := normalizeAIEntryOccurredAt(d.OccurredAt, accountNow)
		d.OccurredAt = normalizedOccurredAt
		if !parsedOccurredAt {
			appendAIQuestion(d, "Ngày trên chứng từ chưa rõ; hãy chọn ngày giao dịch trước khi lưu.")
		} else if dateOnly {
			appendAIQuestion(d, "Ngày trên chứng từ không có giờ; đã điền 00:00, hãy kiểm tra lại.")
		}
		if d.Questions == nil {
			d.Questions = []string{}
		}
		if e := ValidateAIEntryDraft(d.AIEntryDraft, wallet, category); e != nil {
			invalidDrafts++
			appendAIQuestion(d, e.Error())
		}
		d.Questions = normalizeAIEntryQuestions(d.Questions, wallet.ID != "")
	}
	out.Reply = aiEntrySummary(out.Drafts, invalidDrafts)
	if err = s.Entries.FinishMessage(ctx, owner, id, token, out); err != nil {
		slog.Warn("AI entry result persistence failed", "process_id", id, "stage", "persistence", "code", "result_persist_failed", "draft_count", len(out.Drafts), "usage_present", out.ModelUsage != nil)
		s.fail(owner, id, token, boundedAIContext(out.SourceText, 64000))
		return nil, err
	}
	slog.Info("AI entry processing completed", "process_id", id, "draft_count", len(out.Drafts), "usage_stored", out.ModelUsage != nil)
	process, err := s.Entries.Session(ctx, owner, id)
	if err != nil {
		return nil, err
	}
	process.Reply = out.Reply
	return process, nil
}

func (s *AIEntryService) loadStoredImages(ctx context.Context, attachments []entity.AIEntryAttachment) ([]entity.AIImage, error) {
	reader, ok := s.Storage.(AttachmentReader)
	if !ok || reader == nil {
		return nil, ErrAIStorage
	}
	images := make([]entity.AIImage, 0, len(attachments))
	for _, attachment := range attachments {
		if attachment.OCRStatus == "deleted" || attachment.OCRStatus == "deleting" || strings.TrimSpace(attachment.ObjectKey) == "" {
			return nil, ErrAIStorage
		}
		data, err := reader.Read(ctx, attachment.ObjectKey, 5*1024*1024)
		if err != nil {
			return nil, ErrAIStorage
		}
		images = append(images, entity.AIImage{Name: attachment.Filename, MIMEType: attachment.MIMEType, Base64: base64.StdEncoding.EncodeToString(data), AttachmentID: attachment.ID})
	}
	return images, nil
}

// Batch returns the authoritative proposal snapshot for an accepted entry
// task. The task/event row only carries a stable reference; clients must not
// treat its item IDs as the draft source of truth.
func (s *AIEntryService) Batch(ctx context.Context, ownerID, batchID string, version int64) (*entity.AdvisorEntryBatchView, error) {
	if s == nil || s.Entries == nil || strings.TrimSpace(ownerID) == "" || strings.TrimSpace(batchID) == "" || version <= 0 {
		return nil, port.ErrAIInvalid
	}
	session, err := s.Entries.Session(ctx, ownerID, batchID)
	if err != nil {
		return nil, err
	}
	if version != 1 {
		return nil, port.ErrAIConflict
	}
	items := make([]entity.AdvisorEntryBatchItem, 0, len(session.Proposals))
	for _, proposal := range session.Proposals {
		status := proposal.Status
		if status == entity.AIProposalApproved {
			status = entity.AIProposalApplied
		}
		item := entity.AdvisorEntryBatchItem{
			ID: proposal.ID, Version: proposal.Version, Status: status,
			Draft: proposal.Draft, Questions: append([]string(nil), proposal.Questions...),
			SupportedEditFields: []string{"type", "amount", "wallet_id", "category_id", "jar_id", "occurred_at", "note", "included_in_reports"},
		}
		if status == entity.AIProposalRejected || status == entity.AIProposalConflict || status == entity.AIProposalExpired {
			item.Error = status
		}
		if proposal.TransactionID != nil {
			item.ResourceID = *proposal.TransactionID
			item.ResourceReceipt = map[string]any{"resource_id": *proposal.TransactionID}
		}
		items = append(items, item)
	}
	return &entity.AdvisorEntryBatchView{BatchID: batchID, Version: version, SessionID: session.ID, Items: items}, nil
}

// HydrateAdvisorMessages replaces only entry_batch reference parts in the
// response copy. Persisted advisor history remains reference-only and can be
// replayed safely while proposal drafts/statuses stay authoritative in the
// AI-entry tables.
func (s *AIEntryService) HydrateAdvisorMessages(ctx context.Context, ownerID string, messages []entity.AdvisorMessage) ([]entity.AdvisorMessage, error) {
	result := append([]entity.AdvisorMessage(nil), messages...)
	for messageIndex := range result {
		result[messageIndex].Parts = append([]entity.AdvisorPart(nil), result[messageIndex].Parts...)
		for partIndex := range result[messageIndex].Parts {
			part := &result[messageIndex].Parts[partIndex]
			if part.Type != "entry_batch" {
				continue
			}
			envelope, ok := part.Data.(map[string]any)
			if ok {
				// Do not mutate the caller's persisted-history value while adding the
				// authoritative proposal snapshot to the response copy.
				copied := make(map[string]any, len(envelope)+1)
				for key, value := range envelope {
					copied[key] = value
				}
				envelope = copied
			}
			if !ok {
				encoded, marshalErr := json.Marshal(part.Data)
				if marshalErr != nil || json.Unmarshal(encoded, &envelope) != nil {
					return nil, port.ErrAIConflict
				}
			}
			batchID, _ := envelope["batch_id"].(string)
			version := int64Value(envelope["version"])
			if batchID == "" || version <= 0 {
				return nil, port.ErrAIConflict
			}
			batch, err := s.Batch(ctx, ownerID, batchID, version)
			if err != nil {
				return nil, err
			}
			envelope["batch"] = batch
			part.Data = envelope
		}
	}
	return result, nil
}

func int64Value(value any) int64 {
	switch number := value.(type) {
	case int64:
		return number
	case int:
		return int64(number)
	case float64:
		return int64(number)
	case json.Number:
		parsed, _ := number.Int64()
		return parsed
	default:
		return 0
	}
}

func defaultAIEntryCategory(d entity.AIEntryDraft, wallet entity.Wallet, categories []entity.Category) *entity.Category {
	preferred := ""
	if wallet.Type == entity.WalletTypeGoal {
		if d.Type == entity.TransactionTypeIncome {
			preferred = "income_transfer_in"
		} else {
			preferred = "expense_transfer_out"
		}
	} else if d.Type == entity.TransactionTypeIncome {
		preferred = "income_other"
	} else {
		preferred = "expense_general"
	}
	applicable := func(category entity.Category) bool {
		if category.Kind != d.Type {
			return false
		}
		if wallet.Type == entity.WalletTypeGoal {
			if category.SystemKey == nil || *category.SystemKey != preferred {
				return false
			}
		}
		if len(category.WalletIDs) == 0 {
			return true
		}
		for _, walletID := range category.WalletIDs {
			if walletID == wallet.ID {
				return true
			}
		}
		return false
	}
	for index := range categories {
		category := &categories[index]
		if category.SystemKey != nil && *category.SystemKey == preferred && applicable(*category) {
			return category
		}
	}
	for index := range categories {
		category := &categories[index]
		if applicable(*category) {
			return category
		}
	}
	return nil
}

func appendAIQuestion(d *entity.AIExtractDraft, question string) {
	for _, existing := range d.Questions {
		if existing == question {
			return
		}
	}
	d.Questions = append(d.Questions, question)
}

var foreignCurrencyCue = regexp.MustCompile(`(?i)(^|[^[:alnum:]])(USD|EUR|GBP|SGD|JPY|CNY|AUD|CAD|HKD|THB|KRW|TWD)([^[:alnum:]]|$)`)
var foreignCurrencyDescription = regexp.MustCompile(`(?i)ngoại tệ|foreign currency|non[- ]?VND`)

func hasForeignCurrencyCue(note string, questions []string) bool {
	if foreignCurrencyCue.MatchString(note) || foreignCurrencyDescription.MatchString(note) {
		return true
	}
	for _, question := range questions {
		if foreignCurrencyCue.MatchString(question) || foreignCurrencyDescription.MatchString(question) {
			return true
		}
	}
	return false
}

func normalizeAIEntryOccurredAt(raw string, accountNow time.Time) (string, bool, bool) {
	raw = strings.TrimSpace(raw)
	if parsed, err := time.Parse(time.RFC3339, raw); err == nil {
		return parsed.Format(time.RFC3339), true, false
	}
	for _, layout := range []string{"2006-01-02", "02/01/2006", "02-01-2006"} {
		if parsed, err := time.ParseInLocation(layout, raw, accountNow.Location()); err == nil {
			return parsed.Format(time.RFC3339), true, true
		}
	}
	return "", false, false
}

func parseAIEntryWorkbooks(files []entity.AIImage, timezone string) ([]entity.AIStructuredRow, error) {
	rows := make([]entity.AIStructuredRow, 0)
	for fileIndex, file := range files {
		if file.MIMEType != xlsxMIMEType {
			continue
		}
		data, err := base64.StdEncoding.Strict().DecodeString(file.Base64)
		if err != nil {
			return nil, err
		}
		document, err := financialsource.ParseWorkbook(file.Name, data, timezone)
		if err != nil {
			return nil, err
		}
		for _, row := range document.Rows {
			rows = append(rows, entity.AIStructuredRow{
				SourceID: fmt.Sprintf("f%d:%s", fileIndex, row.SourceID), OccurredAt: row.OccurredAt.Format(time.RFC3339), Type: row.Type,
				Amount: row.Amount, SourceCurrency: row.SourceCurrency,
				Description: boundedStructuredText(row.Description, 2048), Reference: boundedStructuredText(row.Reference, 512),
			})
		}
	}
	return rows, nil
}

func boundedStructuredText(text string, limit int) string {
	if len(text) <= limit {
		return text
	}
	text = text[:limit]
	for !utf8.ValidString(text) {
		text = text[:len(text)-1]
	}
	return text
}

func mergeStructuredAIDrafts(rows []entity.AIStructuredRow, drafts []entity.AIExtractDraft) ([]entity.AIExtractDraft, error) {
	if len(rows) != len(drafts) {
		return nil, errors.New("structured source set mismatch")
	}
	byID := make(map[string]entity.AIExtractDraft, len(drafts))
	for _, draft := range drafts {
		if draft.SourceID == "" {
			return nil, errors.New("structured source ID is missing")
		}
		if _, exists := byID[draft.SourceID]; exists {
			return nil, errors.New("structured source ID is duplicated")
		}
		byID[draft.SourceID] = draft
	}
	merged := make([]entity.AIExtractDraft, 0, len(rows))
	for _, row := range rows {
		draft, ok := byID[row.SourceID]
		if !ok {
			return nil, errors.New("structured source ID is unknown")
		}
		draft.Type = row.Type
		draft.Amount = row.Amount
		draft.OccurredAt = row.OccurredAt
		draft.SourceCurrency = row.SourceCurrency
		draft.IncludedInReports = row.Type == entity.TransactionTypeIncome || row.Type == entity.TransactionTypeExpense
		merged = append(merged, draft)
	}
	return merged, nil
}

// alignOCRTexts keeps the model-facing OCR order independent from the
// database's attachment ordering. Attachments are persisted in ID order for
// deterministic repository updates, while the model must see the user's
// original upload order so date/account context and overlap boundaries remain
// meaningful. Empty slots preserve ID alignment for failed/unprocessed files;
// the repository must not treat an empty slot as successfully recognized text.
func alignOCRTexts(images []entity.AIImage, attachments []entity.AIEntryAttachment, texts []string) []string {
	if len(attachments) == 0 {
		return append([]string(nil), texts...)
	}
	byAttachmentID := make(map[string]string, len(texts))
	for index, image := range images {
		if index < len(texts) && image.AttachmentID != "" {
			byAttachmentID[image.AttachmentID] = texts[index]
		}
	}
	ordered := make([]string, len(attachments))
	for index, attachment := range attachments {
		ordered[index] = byAttachmentID[attachment.ID]
	}
	return ordered
}

func (s *AIEntryService) storeForOCR(ctx context.Context, owner, process string, files []entity.AIImage) ([]entity.AIImage, []entity.AIEntryAttachment, error) {
	if len(files) == 0 {
		return []entity.AIImage{}, []entity.AIEntryAttachment{}, nil
	}
	prepared := make([]entity.AIImage, 0, len(files))
	attachments := make([]entity.AIEntryAttachment, 0, len(files))
	keys := make([]string, 0, len(files))
	rollback := func() {
		for _, key := range keys {
			_ = s.Storage.Delete(context.WithoutCancel(ctx), key)
		}
	}
	dataByIndex := make([][]byte, len(files))
	for index, file := range files {
		data, err := base64.StdEncoding.Strict().DecodeString(file.Base64)
		if err != nil {
			rollback()
			return nil, nil, port.ErrAIInvalid
		}
		id := uuid.NewString()
		ext := map[string]string{"application/pdf": "pdf", "image/jpeg": "jpg", "image/png": "png", xlsxMIMEType: "xlsx"}[file.MIMEType]
		key, err := s.Storage.Key(owner, process, id, "receipt."+ext)
		if err != nil {
			rollback()
			return nil, nil, err
		}
		keys = append(keys, key)
		dataByIndex[index] = data
		digest := sha256.Sum256(data)
		ocrStatus := "pending"
		if file.MIMEType == xlsxMIMEType {
			ocrStatus = "completed"
		}
		attachments = append(attachments, entity.AIEntryAttachment{ID: id, OwnerID: owner, ProcessID: process, ObjectKey: key, Filename: filepath.Base(file.Name), MIMEType: file.MIMEType, SizeBytes: int64(len(data)), InputOrder: index, SHA256: hex.EncodeToString(digest[:]), OCRStatus: ocrStatus, DeleteAfter: time.Now().Add(24 * time.Hour)})
		// The original remains in MyPocket's private storage for evidence/download.
		// OCR Platform cannot dereference a signed URL from MyPocket's bucket as an
		// app-owned s3:// source, so pass the already validated bytes to the OCR
		// adapter. The adapter's output boundary still sends only OCR text to the LLM.
		if file.MIMEType != xlsxMIMEType {
			prepared = append(prepared, entity.AIImage{Name: file.Name, MIMEType: file.MIMEType, Base64: file.Base64, AttachmentID: id})
		}
	}
	// Upload independently sized attachments with bounded concurrency. The
	// request still has one deadline and rollback remains all-or-nothing, but a
	// twenty-file batch no longer spends its entire budget on serial PUTs.
	uploadCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	workerCount := aiAttachmentStorageWorkers
	if workerCount > len(files) {
		workerCount = len(files)
	}
	jobs := make(chan int)
	var uploadWG sync.WaitGroup
	var uploadMu sync.Mutex
	var uploadErr error
	worker := func() {
		defer uploadWG.Done()
		for index := range jobs {
			if err := s.Storage.Put(uploadCtx, keys[index], files[index].MIMEType, dataByIndex[index]); err != nil {
				uploadMu.Lock()
				if uploadErr == nil {
					uploadErr = err
					cancel()
				}
				uploadMu.Unlock()
			}
		}
	}
	uploadWG.Add(workerCount)
	for index := 0; index < workerCount; index++ {
		go worker()
	}
	go func() {
		defer close(jobs)
		for index := range files {
			select {
			case jobs <- index:
			case <-uploadCtx.Done():
				return
			}
		}
	}()
	uploadWG.Wait()
	uploadMu.Lock()
	err := uploadErr
	uploadMu.Unlock()
	if err != nil {
		rollback()
		return nil, nil, err
	}
	// Keep model input in request order. The repository receives attachments in
	// deterministic ID order, and alignOCRTexts maps OCR results back to that
	// persistence order after extraction.
	sort.Slice(attachments, func(i, j int) bool { return attachments[i].ID < attachments[j].ID })
	return prepared, attachments, nil
}

func (s *AIEntryService) Process(ctx context.Context, owner string, in AIEntryMessageInput) (*entity.AIEntrySession, error) {
	in.Text = strings.TrimSpace(in.Text)
	if err := s.resolveTimezone(ctx, owner, &in); err != nil {
		slog.Warn("AI entry validation failed", "process_id", in.RequestID, "stage", "validation", "code", "timezone_invalid")
		return nil, err
	}
	if err := s.validate(in); err != nil {
		slog.Warn("AI entry validation failed", "process_id", in.RequestID, "stage", "validation", "code", "input_invalid", "file_count", len(in.Images))
		return nil, err
	}
	if err := s.Entries.CreateProcess(ctx, owner, in.RequestID); err != nil {
		slog.Warn("AI entry process initialization failed", "process_id", in.RequestID, "stage", "persistence", "code", "process_create_failed")
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
	if len(in.Text) > 16000 || len(in.Images) > 20 || (in.Text == "" && len(in.Images) == 0) {
		return fmt.Errorf("%w: cần nội dung hoặc tối đa 20 tệp", port.ErrAIInvalid)
	}
	// This slice has no approved contract for combining source-bound workbook
	// rows with source-less OCR drafts. Reject the mixed request before
	// idempotency, storage, OCR, or model work rather than dropping evidence or
	// running OCR once per structured batch.
	if hasStructuredAttachments(in.Images) && hasOCRAttachments(in.Images) {
		return fmt.Errorf("%w: không thể kết hợp XLSX với PDF hoặc ảnh trong cùng yêu cầu", port.ErrAIInvalid)
	}
	if hasOCRAttachments(in.Images) && !s.Extractor.OCRConfigured() {
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

func hasStructuredAttachments(files []entity.AIImage) bool {
	for _, file := range files {
		if file.MIMEType == xlsxMIMEType {
			return true
		}
	}
	return false
}

func hasOCRAttachments(files []entity.AIImage) bool {
	for _, file := range files {
		if file.MIMEType != xlsxMIMEType {
			return true
		}
	}
	return false
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
	s.failWith(owner, id, token, ErrAIProvider.Error(), "provider_error", sourceText)
}
func (s *AIEntryService) failWith(owner, id, token, message, errorCode, sourceText string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = s.Entries.FailMessage(ctx, owner, id, token, message, errorCode, sourceText)
}

type providerDiagnostic interface {
	Diagnostic() (stage, code string, status int)
}

func (s *AIEntryService) failWithDiagnostic(owner, id, token string, providerErr error, sourceText string) {
	stage, code, status := "provider", "provider_error", 0
	var diagnostic providerDiagnostic
	if errors.As(providerErr, &diagnostic) {
		stage, code, status = diagnostic.Diagnostic()
	}
	slog.Warn("AI provider processing failed", "process_id", id, "request_id", id, "stage", stage, "code", code, "provider_status", status)
	s.failWith(owner, id, token, ErrAIProvider.Error(), code, sourceText)
}
