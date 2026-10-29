package usecase

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mypocket/backend/internal/entity"
	port "github.com/mypocket/backend/internal/repository"
	"image"
	"image/png"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"
	"unicode/utf8"
)

func TestAIEntrySessionSerializesSafeErrorCode(t *testing.T) {
	data, err := json.Marshal(entity.AIEntrySession{ID: "session", Error: "OCR không đọc được chứng từ.", ErrorCode: "ocr_failed", Reply: "Chọn ví để tiếp tục."})
	if err != nil {
		t.Fatal(err)
	}
	if string(data) == "" || !strings.Contains(string(data), `"error_code":"ocr_failed"`) {
		t.Fatalf("error code missing from session JSON: %s", data)
	}
	if !strings.Contains(string(data), `"reply":"Chọn ví để tiếp tục."`) {
		t.Fatalf("provider reply missing from session JSON: %s", data)
	}
}

type entryRepoStub struct {
	port.AIEntryRepository
	begins        int
	failed        bool
	finished      bool
	output        entity.AIExtractOutput
	failureSource string
	attachments   []entity.AIEntryAttachment
}

func (r *entryRepoStub) Session(context.Context, string, string) (*entity.AIEntrySession, error) {
	return &entity.AIEntrySession{ID: "session", Messages: []entity.AIEntryMessage{}}, nil
}
func (*entryRepoStub) CreateProcess(context.Context, string, string) error { return nil }
func (r *entryRepoStub) BeginMessage(context.Context, string, string, string, string, string) (string, bool, error) {
	r.begins++
	return "token", true, nil
}
func (r *entryRepoStub) CreateAttachments(_ context.Context, attachments []entity.AIEntryAttachment) error {
	r.attachments = append(r.attachments, attachments...)
	return nil
}
func (r *entryRepoStub) UpdateAttachmentOCR(_ context.Context, owner, process string, texts []string, completed bool) error {
	for i := range r.attachments {
		if i < len(texts) && strings.TrimSpace(texts[i]) != "" {
			r.attachments[i].OCRText = texts[i]
			r.attachments[i].OCRStatus = "completed"
		} else {
			r.attachments[i].OCRStatus = "failed"
		}
	}
	return nil
}

func TestAIEntryExtractsOnlySubmittedTextForOneShot(t *testing.T) {
	r := &entryRepoStub{}
	var seen entity.AIExtractInput
	s := &AIEntryService{Entries: r, Wallets: entryWallets{}, Categories: entryCategories{}, Extractor: entryExtractor{configured: true, seen: &seen}}
	if _, err := s.Send(context.Background(), "owner", "session", AIEntryMessageInput{RequestID: "00000000-0000-4000-8000-000000000001", Text: "chọn ví w", Timezone: "Asia/Ho_Chi_Minh"}); err != nil {
		t.Fatal(err)
	}
	if seen.Text != "chọn ví w" {
		t.Fatalf("extractor received content outside the one-shot input: %q", seen.Text)
	}
}

func TestAIEntryLeavesMissingDateUnapprovable(t *testing.T) {
	r := &entryRepoStub{}
	s := &AIEntryService{Entries: r, Wallets: entryWallets{}, Categories: entryCategories{}, Extractor: entryExtractor{configured: true, out: entity.AIExtractOutput{Drafts: []entity.AIExtractDraft{{AIEntryDraft: entity.AIEntryDraft{Type: "expense", Amount: 45000, WalletID: "w"}}}}}}
	if _, err := s.Send(context.Background(), "owner", "session", AIEntryMessageInput{RequestID: "00000000-0000-0000-0000-000000000011", Text: "cà phê", Timezone: "Asia/Ho_Chi_Minh"}); err != nil {
		t.Fatal(err)
	}
	if len(r.output.Drafts) != 1 {
		t.Fatalf("expected one draft, got %d", len(r.output.Drafts))
	}
	if r.output.Drafts[0].OccurredAt != "" {
		t.Fatalf("missing date became an approvable instant: %q", r.output.Drafts[0].OccurredAt)
	}
	if !strings.Contains(strings.Join(r.output.Drafts[0].Questions, " "), "chọn ngày") {
		t.Fatalf("missing date needs a correction question: %#v", r.output.Drafts[0].Questions)
	}
	if ValidateAIEntryDraft(r.output.Drafts[0].AIEntryDraft, entity.Wallet{ID: "w", Type: "basic"}, nil) == nil {
		t.Fatal("missing date became approvable")
	}
}

func TestAIEntryNormalizesDateOnlyOutputWithoutReplacingItWithToday(t *testing.T) {
	r := &entryRepoStub{}
	s := &AIEntryService{Entries: r, Wallets: entryWallets{}, Categories: entryCategories{}, Extractor: entryExtractor{configured: true, out: entity.AIExtractOutput{Drafts: []entity.AIExtractDraft{{AIEntryDraft: entity.AIEntryDraft{Type: "expense", Amount: 45000, WalletID: "w", OccurredAt: "20/09/2026"}}}}}}
	if _, err := s.Send(context.Background(), "owner", "session", AIEntryMessageInput{RequestID: "00000000-0000-0000-0000-000000000012", Text: "cà phê", Timezone: "Asia/Ho_Chi_Minh"}); err != nil {
		t.Fatal(err)
	}
	parsed, err := time.Parse(time.RFC3339, r.output.Drafts[0].OccurredAt)
	if err != nil || parsed.In(mustLocation(t, "Asia/Ho_Chi_Minh")).Format("2006-01-02") != "2026-09-20" {
		t.Fatalf("date-only output was not normalized: %q", r.output.Drafts[0].OccurredAt)
	}
	if !strings.Contains(strings.Join(r.output.Drafts[0].Questions, " "), "không có giờ") {
		t.Fatalf("date-only normalization should be disclosed for review: %#v", r.output.Drafts[0].Questions)
	}
}

func TestAIEntryPreservesIdenticalStatementRows(t *testing.T) {
	r := &entryRepoStub{}
	draft := entity.AIEntryDraft{Type: "expense", Amount: 45000, WalletID: "w", OccurredAt: "2026-09-20T08:00:00+07:00", Note: "Highlands"}
	s := &AIEntryService{Entries: r, Wallets: entryWallets{}, Categories: entryCategories{}, Extractor: entryExtractor{configured: true, out: entity.AIExtractOutput{Drafts: []entity.AIExtractDraft{{AIEntryDraft: draft}, {AIEntryDraft: draft, Questions: []string{"Bản sao từ ảnh chồng lấn."}}}}}}
	if _, err := s.Send(context.Background(), "owner", "session", AIEntryMessageInput{RequestID: "00000000-0000-0000-0000-000000000013", Text: "cà phê", Timezone: "Asia/Ho_Chi_Minh"}); err != nil {
		t.Fatal(err)
	}
	if len(r.output.Drafts) != 2 {
		t.Fatalf("separate identical rows were collapsed: count=%d", len(r.output.Drafts))
	}
}

func TestAIEntryHoldsForeignCurrencyAmountForReview(t *testing.T) {
	r := &entryRepoStub{}
	s := &AIEntryService{Entries: r, Wallets: entryWallets{}, Categories: entryCategories{}, Extractor: entryExtractor{configured: true, out: entity.AIExtractOutput{Drafts: []entity.AIExtractDraft{{AIEntryDraft: entity.AIEntryDraft{Type: "expense", Amount: 222, WalletID: "w", OccurredAt: "2026-10-01T00:00:00+07:00", Note: "Ticket 222.22 USD"}, Questions: []string{"USD amount requires conversion"}}}}}}
	if _, err := s.Send(context.Background(), "owner", "session", AIEntryMessageInput{RequestID: "00000000-0000-0000-0000-000000000014", Text: "review statement", Timezone: "Asia/Ho_Chi_Minh"}); err != nil {
		t.Fatal(err)
	}
	d := r.output.Drafts[0]
	if d.Amount != 0 || !strings.Contains(strings.Join(d.Questions, " "), "VND") {
		t.Fatalf("foreign amount was proposed as VND: amount=%d questions=%d", d.Amount, len(d.Questions))
	}
	if ValidateAIEntryDraft(d.AIEntryDraft, entity.Wallet{ID: "w", Type: "basic"}, nil) == nil {
		t.Fatal("foreign amount became approvable")
	}
}

func TestAIEntrySourceCurrencyControlsAmountWithoutNoteCue(t *testing.T) {
	for _, tc := range []struct {
		currency string
		want     int64
	}{
		{currency: "VND", want: 45000},
		{currency: "foreign", want: 0},
		{currency: "unknown", want: 0},
	} {
		t.Run(tc.currency, func(t *testing.T) {
			r := &entryRepoStub{}
			s := &AIEntryService{Entries: r, Wallets: entryWallets{}, Categories: entryCategories{}, Extractor: entryExtractor{configured: true, out: entity.AIExtractOutput{Drafts: []entity.AIExtractDraft{{AIEntryDraft: entity.AIEntryDraft{Type: "expense", Amount: 45000, WalletID: "w", OccurredAt: "2026-10-01T00:00:00+07:00", Note: "Ticket"}, SourceCurrency: tc.currency}}}}}
			if _, err := s.Send(context.Background(), "owner", "session", AIEntryMessageInput{RequestID: "00000000-0000-0000-0000-000000000016", Text: "review", Timezone: "Asia/Ho_Chi_Minh"}); err != nil {
				t.Fatal(err)
			}
			if got := r.output.Drafts[0].Amount; got != tc.want {
				t.Fatalf("amount=%d want=%d", got, tc.want)
			}
		})
	}
}

func TestAIEntryHoldsForeignAmountWhenQuestionSaysNgoaiTe(t *testing.T) {
	r := &entryRepoStub{}
	s := &AIEntryService{Entries: r, Wallets: entryWallets{}, Categories: entryCategories{}, Extractor: entryExtractor{configured: true, out: entity.AIExtractOutput{Drafts: []entity.AIExtractDraft{{AIEntryDraft: entity.AIEntryDraft{Type: "expense", Amount: 222, WalletID: "w", OccurredAt: "2026-10-01T00:00:00+07:00", Note: "Ticket 222.22"}, Questions: []string{"Giá trị ngoại tệ cần xác minh"}, SourceCurrency: "VND"}}}}}
	if _, err := s.Send(context.Background(), "owner", "session", AIEntryMessageInput{RequestID: "00000000-0000-0000-0000-000000000017", Text: "review", Timezone: "Asia/Ho_Chi_Minh"}); err != nil {
		t.Fatal(err)
	}
	if got := r.output.Drafts[0].Amount; got != 0 {
		t.Fatalf("foreign amount with explicit review cue was proposed as %d VND", got)
	}
}

func mustLocation(t *testing.T, name string) *time.Location {
	t.Helper()
	location, err := time.LoadLocation(name)
	if err != nil {
		t.Fatal(err)
	}
	return location
}

func TestAIEntryKeepsOCRSourceWhenModelFails(t *testing.T) {
	r := &entryRepoStub{}
	s := &AIEntryService{Entries: r, Wallets: entryWallets{}, Categories: entryCategories{}, Extractor: entryExtractor{configured: true, err: errors.New("model unavailable"), out: entity.AIExtractOutput{SourceText: "OCR total 35000"}}}
	_, err := s.Send(context.Background(), "owner", "session", AIEntryMessageInput{RequestID: "00000000-0000-4000-8000-000000000001", Text: "bill", Timezone: "Asia/Ho_Chi_Minh"})
	if !errors.Is(err, ErrAIProvider) || r.failureSource != "OCR total 35000" {
		t.Fatalf("OCR evidence lost on model failure: %q %v", r.failureSource, err)
	}
}
func (r *entryRepoStub) FailMessage(_ context.Context, _, _, _, _, _, sourceText string) error {
	r.failed = true
	r.failureSource = sourceText
	return nil
}
func (r *entryRepoStub) FinishMessage(_ context.Context, _, _, _ string, o entity.AIExtractOutput) error {
	r.finished = true
	r.output = o
	return nil
}

type entryWallets struct{ port.WalletRepository }

func (entryWallets) List(string) ([]entity.Wallet, error) {
	return []entity.Wallet{{ID: "w", OwnerID: "owner", Type: "basic"}}, nil
}

type entryCategories struct{ port.CategoryRepository }

func (entryCategories) ListVisible(string) ([]entity.Category, error) {
	return []entity.Category{}, nil
}

type entryExtractor struct {
	configured    bool
	ocrConfigured bool
	err           error
	validationErr error
	out           entity.AIExtractOutput
	seen          *entity.AIExtractInput
}

func (e entryExtractor) Configured() bool                     { return e.configured }
func (e entryExtractor) OCRConfigured() bool                  { return e.ocrConfigured }
func (e entryExtractor) ValidateFiles([]entity.AIImage) error { return e.validationErr }
func (e entryExtractor) Extract(_ context.Context, input entity.AIExtractInput) (entity.AIExtractOutput, error) {
	if e.seen != nil {
		*e.seen = input
	}
	if e.out.Drafts == nil && len(input.StructuredRows) > 0 && e.err == nil {
		e.out.Drafts = make([]entity.AIExtractDraft, len(input.StructuredRows))
		for i, row := range input.StructuredRows {
			e.out.Drafts[i] = entity.AIExtractDraft{SourceID: row.SourceID, AIEntryDraft: entity.AIEntryDraft{WalletID: "w"}}
		}
	}
	return e.out, e.err
}

func TestAIEntryInvalidFileIsClientErrorBeforeProviderCharge(t *testing.T) {
	r := &entryRepoStub{}
	s := &AIEntryService{Entries: r, Wallets: entryWallets{}, Categories: entryCategories{}, Extractor: entryExtractor{configured: true, ocrConfigured: true, validationErr: errors.New("invalid PDF signature")}}
	_, err := s.Send(context.Background(), "owner", "session", AIEntryMessageInput{RequestID: "00000000-0000-4000-8000-000000000001", Text: "receipt", Timezone: "UTC", Images: []entity.AIImage{{Name: "receipt.pdf", MIMEType: "application/pdf", Base64: "JVBERi0="}}})
	if !errors.Is(err, port.ErrAIInvalid) || r.begins != 0 {
		t.Fatalf("invalid file should fail before request creation/provider work: begins=%d err=%v", r.begins, err)
	}
}

func TestAIEntryStoresOriginalPrivatelyAndPassesValidatedBytesToOCR(t *testing.T) {
	r := &entryRepoStub{}
	var seen entity.AIExtractInput
	storage := &attachmentStoreStub{signedURL: "https://private-storage.invalid/read?signature=temporary"}
	s := &AIEntryService{Entries: r, Wallets: entryWallets{}, Categories: entryCategories{}, Extractor: entryExtractor{configured: true, ocrConfigured: true, seen: &seen, out: entity.AIExtractOutput{AttachmentTexts: []string{"OCR amount 42000"}}}, Storage: storage}
	fileBytes := []byte("%PDF-1.7 private receipt %%EOF")
	_, err := s.Process(context.Background(), "owner", AIEntryMessageInput{RequestID: "00000000-0000-4000-8000-000000000001", Text: "receipt", Timezone: "UTC", Images: []entity.AIImage{{Name: "receipt.pdf", MIMEType: "application/pdf", Base64: base64.StdEncoding.EncodeToString(fileBytes)}}})
	if err != nil {
		t.Fatal(err)
	}
	if string(storage.data) != string(fileBytes) || storage.contentType != "application/pdf" || len(r.attachments) != 1 {
		t.Fatalf("original was not retained privately: bytes=%d type=%s metadata=%d", len(storage.data), storage.contentType, len(r.attachments))
	}
	if len(seen.Images) != 1 || seen.Images[0].Base64 != base64.StdEncoding.EncodeToString(fileBytes) || seen.Images[0].SourceURL != "" {
		t.Fatalf("OCR input must carry validated bytes and no MyPocket storage URL: %+v", seen.Images)
	}
}

func TestAIEntryXLSXIsStoredParsedWithRequestTimezoneAndSkipsOCR(t *testing.T) {
	r := &entryRepoStub{}
	var seen entity.AIExtractInput
	storage := &attachmentStoreStub{}
	s := &AIEntryService{
		Entries: r, Wallets: entryWallets{}, Categories: entryCategories{},
		Extractor: entryExtractor{configured: true, ocrConfigured: false, seen: &seen}, Storage: storage,
	}
	workbook := aiEntryWorkbook(t, true)
	_, err := s.Process(context.Background(), "owner", AIEntryMessageInput{
		RequestID: "00000000-0000-4000-8000-000000000021", Timezone: "Asia/Ho_Chi_Minh",
		Images: []entity.AIImage{{Name: "statement.xlsx", MIMEType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", Base64: base64.StdEncoding.EncodeToString(workbook)}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if storage.contentType != "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" || !bytes.Equal(storage.data, workbook) || len(r.attachments) != 1 {
		t.Fatalf("workbook was not retained privately: type=%q bytes=%d attachments=%d", storage.contentType, len(storage.data), len(r.attachments))
	}
	if r.attachments[0].OCRStatus != "completed" {
		t.Fatalf("non-OCR workbook used unsupported attachment status %q", r.attachments[0].OCRStatus)
	}
	if len(seen.Images) != 0 {
		t.Fatalf("workbook was sent through OCR: %#v", seen.Images)
	}
	if len(seen.StructuredRows) != 2 {
		t.Fatalf("canonical rows=%d want=2", len(seen.StructuredRows))
	}
	if seen.StructuredRows[0].OccurredAt != "2026-10-26T00:00:00+07:00" {
		t.Fatalf("request timezone was not used: %q", seen.StructuredRows[0].OccurredAt)
	}
	if seen.StructuredRows[0].SourceID == seen.StructuredRows[1].SourceID {
		t.Fatalf("identical source rows share an ID: %q", seen.StructuredRows[0].SourceID)
	}
}

func TestAIEntryStructuredRowsOverrideProviderFinancialFields(t *testing.T) {
	workbook := aiEntryWorkbook(t, false)
	file := entity.AIImage{Name: "statement.xlsx", MIMEType: xlsxMIMEType, Base64: base64.StdEncoding.EncodeToString(workbook)}
	parsed, err := parseAIEntryWorkbooks([]entity.AIImage{file}, "Asia/Ho_Chi_Minh")
	if err != nil {
		t.Fatal(err)
	}
	r := &entryRepoStub{}
	s := &AIEntryService{
		Entries: r, Wallets: entryWallets{}, Categories: entryCategories{}, Storage: &attachmentStoreStub{},
		Extractor: entryExtractor{configured: true, out: entity.AIExtractOutput{Drafts: []entity.AIExtractDraft{{
			SourceID:       parsed[0].SourceID,
			AIEntryDraft:   entity.AIEntryDraft{Type: "income", Amount: 1, WalletID: "w", OccurredAt: "2030-01-01T00:00:00Z", Note: "semantic note 50 USD", IncludedInReports: false},
			Questions:      []string{"Foreign currency amount may need conversion"},
			SourceCurrency: "foreign",
		}}}},
	}
	_, err = s.Process(context.Background(), "owner", AIEntryMessageInput{
		RequestID: "00000000-0000-4000-8000-000000000022", Timezone: "Asia/Ho_Chi_Minh",
		Images: []entity.AIImage{file},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.output.Drafts) != 1 {
		t.Fatalf("drafts=%d want=1", len(r.output.Drafts))
	}
	draft := r.output.Drafts[0]
	if draft.Type != "expense" || draft.Amount != 50000 || draft.OccurredAt != "2026-10-26T00:00:00+07:00" || draft.SourceCurrency != "VND" || !draft.IncludedInReports {
		t.Fatalf("provider changed parser-owned fields: %#v", draft)
	}
	if draft.Note != "semantic note 50 USD" || draft.WalletID != "w" {
		t.Fatalf("semantic enrichment was lost: %#v", draft)
	}
}

func TestAIEntryRejectsMixedXLSXAndOCRBeforeStorageOrProvider(t *testing.T) {
	workbook := aiEntryWorkbook(t, true)
	storage := &attachmentStoreStub{}
	var seen entity.AIExtractInput
	r := &entryRepoStub{}
	s := &AIEntryService{
		Entries: r, Wallets: entryWallets{}, Categories: entryCategories{}, Storage: storage,
		Extractor: entryExtractor{configured: true, ocrConfigured: true, seen: &seen},
	}
	_, err := s.Send(context.Background(), "owner", "session", AIEntryMessageInput{
		RequestID: "00000000-0000-4000-8000-000000000023", Timezone: "UTC",
		Images: []entity.AIImage{
			{Name: "statement.xlsx", MIMEType: xlsxMIMEType, Base64: base64.StdEncoding.EncodeToString(workbook)},
			{Name: "receipt.pdf", MIMEType: "application/pdf", Base64: base64.StdEncoding.EncodeToString([]byte("%PDF-1.7 receipt %%EOF"))},
		},
	})
	if !errors.Is(err, port.ErrAIInvalid) {
		t.Fatalf("mixed structured and OCR submission error=%v want invalid input", err)
	}
	if r.begins != 0 || storage.puts != 0 || len(seen.Images) != 0 || len(seen.StructuredRows) != 0 {
		t.Fatalf("mixed submission caused side effects: begins=%d puts=%d seen=%+v", r.begins, storage.puts, seen)
	}
}

func TestAIEntryStructuredRowsRejectSourceIDSetMismatch(t *testing.T) {
	workbook := aiEntryWorkbook(t, true)
	file := entity.AIImage{Name: "statement.xlsx", MIMEType: xlsxMIMEType, Base64: base64.StdEncoding.EncodeToString(workbook)}
	parsed, err := parseAIEntryWorkbooks([]entity.AIImage{file}, "UTC")
	if err != nil {
		t.Fatal(err)
	}
	valid := func(sourceID string) entity.AIExtractDraft {
		return entity.AIExtractDraft{SourceID: sourceID, AIEntryDraft: entity.AIEntryDraft{Type: "expense", Amount: 50000, WalletID: "w", OccurredAt: "2026-10-26T00:00:00Z"}, SourceCurrency: "VND"}
	}
	tests := []struct {
		name   string
		drafts []entity.AIExtractDraft
	}{
		{name: "missing", drafts: []entity.AIExtractDraft{valid(parsed[0].SourceID)}},
		{name: "duplicate", drafts: []entity.AIExtractDraft{valid(parsed[0].SourceID), valid(parsed[0].SourceID)}},
		{name: "unknown", drafts: []entity.AIExtractDraft{valid(parsed[0].SourceID), valid("foreign-source-id")}},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &entryRepoStub{}
			s := &AIEntryService{Entries: r, Wallets: entryWallets{}, Categories: entryCategories{}, Storage: &attachmentStoreStub{}, Extractor: entryExtractor{configured: true, out: entity.AIExtractOutput{Drafts: tt.drafts}}}
			_, err := s.Process(context.Background(), "owner", AIEntryMessageInput{
				RequestID: fmt.Sprintf("00000000-0000-4000-8000-%012d", 30+i), Timezone: "UTC",
				Images: []entity.AIImage{file},
			})
			if !errors.Is(err, ErrAIProvider) || r.finished {
				t.Fatalf("source-ID mismatch did not fail closed: err=%v finished=%v", err, r.finished)
			}
		})
	}
}

func TestAIEntryIdenticalRowsAcrossIdenticallyNamedWorkbooksRemainDistinct(t *testing.T) {
	workbook := aiEntryWorkbook(t, false)
	files := []entity.AIImage{
		{Name: "statement.xlsx", MIMEType: xlsxMIMEType, Base64: base64.StdEncoding.EncodeToString(workbook)},
		{Name: "statement.xlsx", MIMEType: xlsxMIMEType, Base64: base64.StdEncoding.EncodeToString(workbook)},
	}
	rows, err := parseAIEntryWorkbooks(files, "UTC")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[0].SourceID == rows[1].SourceID {
		t.Fatalf("identical workbook rows were not kept distinct: %#v", rows)
	}
}

func TestStructuredRowTextBoundIsUTF8SafeAndDoesNotExceedLimit(t *testing.T) {
	got := boundedStructuredText(strings.Repeat("é", 1500), 2048)
	if len(got) > 2048 || !utf8.ValidString(got) {
		t.Fatalf("bounded structured text has bytes=%d valid_utf8=%v", len(got), utf8.ValidString(got))
	}
}

func aiEntryWorkbook(t *testing.T, duplicate bool) []byte {
	t.Helper()
	rows := `<row r="1"><c r="A1" t="inlineStr"><is><t>Ngay giao dich</t></is></c><c r="B1" t="inlineStr"><is><t>So tham chieu</t></is></c><c r="C1" t="inlineStr"><is><t>Dien giai</t></is></c><c r="D1" t="inlineStr"><is><t>Ghi no</t></is></c><c r="E1" t="inlineStr"><is><t>Ghi co</t></is></c></row>` +
		`<row r="2"><c r="A2" t="inlineStr"><is><t>26/10/2026</t></is></c><c r="B2" t="inlineStr"><is><t>REF</t></is></c><c r="C2" t="inlineStr"><is><t>Coffee</t></is></c><c r="D2" t="inlineStr"><is><t>50,000</t></is></c><c r="E2" t="inlineStr"><is><t></t></is></c></row>`
	if duplicate {
		rows += `<row r="3"><c r="A3" t="inlineStr"><is><t>26/10/2026</t></is></c><c r="B3" t="inlineStr"><is><t>REF</t></is></c><c r="C3" t="inlineStr"><is><t>Coffee</t></is></c><c r="D3" t="inlineStr"><is><t>50,000</t></is></c><c r="E3" t="inlineStr"><is><t></t></is></c></row>`
	}
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	files := map[string]string{
		"[Content_Types].xml":        `<?xml version="1.0"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/><Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/></Types>`,
		"_rels/.rels":                `<?xml version="1.0"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/></Relationships>`,
		"xl/workbook.xml":            `<?xml version="1.0"?><workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets><sheet name="Transactions" sheetId="1" r:id="rId1"/></sheets></workbook>`,
		"xl/_rels/workbook.xml.rels": `<?xml version="1.0"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/></Relationships>`,
		"xl/worksheets/sheet1.xml":   `<?xml version="1.0"?><worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>` + rows + `</sheetData></worksheet>`,
	}
	for name, content := range files {
		writer, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := writer.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestAIEntryKeepsSubmittedImageOrderForOCR(t *testing.T) {
	storage := &orderedAttachmentStoreStub{}
	service := &AIEntryService{Storage: storage}
	files := make([]entity.AIImage, 10)
	want := make([]string, len(files))
	for i := range files {
		files[i] = entity.AIImage{Name: fmt.Sprintf("page-%02d.png", i), MIMEType: "image/png", Base64: base64.StdEncoding.EncodeToString(pngBytes(t))}
		want[i] = files[i].Name
	}
	got, attachments, err := service.storeForOCR(context.Background(), "owner", "session", files)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(want) || len(attachments) != len(want) {
		t.Fatalf("stored image count mismatch: images=%d attachments=%d want=%d", len(got), len(attachments), len(want))
	}
	attachmentNames := make(map[string]bool, len(attachments))
	for _, attachment := range attachments {
		attachmentNames[attachment.Filename] = true
	}
	for i := range want {
		if got[i].Name != want[i] {
			t.Fatalf("submitted image order changed at %d: image=%q want=%q", i, got[i].Name, want[i])
		}
		if !attachmentNames[want[i]] {
			t.Fatalf("stored attachment missing submitted file %q", want[i])
		}
	}
}

func TestAIEntryUploadsAttachmentsWithBoundedConcurrency(t *testing.T) {
	storage := &concurrentAttachmentStoreStub{}
	service := &AIEntryService{Storage: storage}
	files := make([]entity.AIImage, 8)
	for i := range files {
		files[i] = entity.AIImage{Name: fmt.Sprintf("page-%02d.png", i), MIMEType: "image/png", Base64: base64.StdEncoding.EncodeToString(pngBytes(t))}
	}
	if _, _, err := service.storeForOCR(context.Background(), "owner", "session", files); err != nil {
		t.Fatal(err)
	}
	if got := storage.max.Load(); got < 2 || got > int32(aiAttachmentStorageWorkers) {
		t.Fatalf("attachment upload concurrency=%d want between 2 and %d", got, aiAttachmentStorageWorkers)
	}
}

func TestAIEntryAlignsOCRTextToAttachmentPersistenceOrder(t *testing.T) {
	images := []entity.AIImage{{Name: "page-a.png", AttachmentID: "a"}, {Name: "page-z.png", AttachmentID: "z"}}
	attachments := []entity.AIEntryAttachment{{ID: "z"}, {ID: "a"}}
	texts := []string{"OCR page a", "OCR page z"}
	got := alignOCRTexts(images, attachments, texts)
	want := []string{"OCR page z", "OCR page a"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("OCR text was not aligned to attachment persistence order: got=%v want=%v", got, want)
	}
}

func pngBytes(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

type orderedAttachmentStoreStub struct{}

func (*orderedAttachmentStoreStub) Key(_, _, _, filename string) (string, error) {
	return "private/" + filename, nil
}
func (*orderedAttachmentStoreStub) Put(context.Context, string, string, []byte) error { return nil }
func (*orderedAttachmentStoreStub) Delete(context.Context, string) error              { return nil }
func (*orderedAttachmentStoreStub) SignedGet(context.Context, string) (string, error) {
	return "", nil
}

type concurrentAttachmentStoreStub struct {
	current atomic.Int32
	max     atomic.Int32
}

func (*concurrentAttachmentStoreStub) Key(_, _, _, filename string) (string, error) {
	return "private/" + filename, nil
}
func (s *concurrentAttachmentStoreStub) Put(context.Context, string, string, []byte) error {
	current := s.current.Add(1)
	for {
		max := s.max.Load()
		if current <= max || s.max.CompareAndSwap(max, current) {
			break
		}
	}
	time.Sleep(10 * time.Millisecond)
	s.current.Add(-1)
	return nil
}
func (*concurrentAttachmentStoreStub) Delete(context.Context, string) error { return nil }
func (*concurrentAttachmentStoreStub) SignedGet(context.Context, string) (string, error) {
	return "", nil
}

type attachmentStoreStub struct {
	signedURL, contentType string
	data                   []byte
	putErr                 error
	deleteErr              error
	deleted                int
	puts                   int
}

func (s *attachmentStoreStub) Key(string, string, string, string) (string, error) {
	return "private/env/key.pdf", nil
}
func (s *attachmentStoreStub) Put(_ context.Context, _ string, contentType string, data []byte) error {
	s.puts++
	if s.putErr != nil {
		return s.putErr
	}
	s.contentType = contentType
	s.data = append([]byte(nil), data...)
	return nil
}

func TestAIEntryStorageFailureStopsBeforeOCRAndLLM(t *testing.T) {
	r := &entryRepoStub{}
	var seen entity.AIExtractInput
	s := &AIEntryService{Entries: r, Wallets: entryWallets{}, Categories: entryCategories{}, Extractor: entryExtractor{configured: true, ocrConfigured: true, seen: &seen}, Storage: &attachmentStoreStub{putErr: errors.New("storage unavailable")}}
	_, err := s.Process(context.Background(), "owner", AIEntryMessageInput{RequestID: "00000000-0000-4000-8000-000000000002", Text: "receipt", Timezone: "UTC", Images: []entity.AIImage{{Name: "receipt.pdf", MIMEType: "application/pdf", Base64: base64.StdEncoding.EncodeToString([]byte("%PDF-1.7 receipt"))}}})
	if !errors.Is(err, ErrAIStorage) || seen.Text != "" || r.finished {
		t.Fatalf("storage failure must stop before OCR/model and proposals: err=%v seen=%+v", err, seen)
	}
}
func (s *attachmentStoreStub) Delete(context.Context, string) error { s.deleted++; return s.deleteErr }
func (s *attachmentStoreStub) SignedGet(context.Context, string) (string, error) {
	return s.signedURL, nil
}

type attachmentCleanupRepoStub struct {
	items    []entity.AIEntryAttachment
	statuses []string
}

func (r *attachmentCleanupRepoStub) ClaimExpiredAttachments(context.Context, time.Time, time.Time, int) ([]entity.AIEntryAttachment, error) {
	return r.items, nil
}
func (r *attachmentCleanupRepoStub) SetAttachmentDeleteStatus(_ context.Context, _, status string) error {
	r.statuses = append(r.statuses, status)
	return nil
}

func TestAttachmentCleanupDeletesExpiredCandidatesAndRetainsFailureState(t *testing.T) {
	repo := &attachmentCleanupRepoStub{items: []entity.AIEntryAttachment{{ID: "a", ObjectKey: "private/a"}, {ID: "b", ObjectKey: "private/b"}}}
	store := &attachmentStoreStub{deleteErr: errors.New("temporary failure")}
	cleanup := &AIEntryAttachmentCleanup{Entries: repo, Storage: store, Now: func() time.Time { return time.Date(2026, 9, 22, 0, 0, 0, 0, time.UTC) }}
	deleted, err := cleanup.Run(context.Background(), 25)
	if err == nil || deleted != 0 || store.deleted != 2 || len(repo.statuses) != 2 || repo.statuses[0] != "delete_failed" {
		t.Fatalf("cleanup should retain retry state: deleted=%d statuses=%v err=%v", deleted, repo.statuses, err)
	}
}

func TestAIEntryUnavailableAndInvalidRequestsDoNotCreateDrafts(t *testing.T) {
	r := &entryRepoStub{}
	s := &AIEntryService{Entries: r, Wallets: entryWallets{}, Categories: entryCategories{}, Extractor: entryExtractor{}}
	valid := AIEntryMessageInput{RequestID: "00000000-0000-4000-8000-000000000001", Text: "ăn 35k", Timezone: "Asia/Ho_Chi_Minh"}
	if _, err := s.Send(context.Background(), "owner", "session", valid); !errors.Is(err, ErrAIUnavailable) {
		t.Fatalf("configuration: %v", err)
	}
	if r.begins != 0 {
		t.Fatal("unconfigured model accepted job")
	}
	s.Extractor = entryExtractor{configured: true}
	for _, change := range []func(*AIEntryMessageInput){func(i *AIEntryMessageInput) { i.Timezone = "wrong/timezone" }, func(i *AIEntryMessageInput) { i.RequestID = "" }, func(i *AIEntryMessageInput) { i.Text = "" }} {
		i := valid
		change(&i)
		if _, err := s.Send(context.Background(), "owner", "session", i); err == nil {
			t.Fatalf("bad input accepted: %+v", i)
		}
	}
	if r.begins != 0 {
		t.Fatal("bad input created job")
	}
}

func TestAIEntryProviderFailureDoesNotPersistProposals(t *testing.T) {
	r := &entryRepoStub{}
	s := &AIEntryService{Entries: r, Wallets: entryWallets{}, Categories: entryCategories{}, Extractor: entryExtractor{configured: true, err: errors.New("upstream secret raw response")}}
	_, err := s.Send(context.Background(), "owner", "session", AIEntryMessageInput{RequestID: "00000000-0000-4000-8000-000000000001", Text: "ăn 35k", Timezone: "Asia/Ho_Chi_Minh"})
	if !errors.Is(err, ErrAIProvider) || !r.failed || r.finished {
		t.Fatalf("failure not isolated: %v %+v", err, r)
	}
}

func TestAIEntryUnknownWalletFallsBackToSuppliedWallet(t *testing.T) {
	r := &entryRepoStub{}
	s := &AIEntryService{Entries: r, Wallets: entryWallets{}, Categories: entryCategories{}, Extractor: entryExtractor{configured: true, out: entity.AIExtractOutput{Reply: "review", Drafts: []entity.AIExtractDraft{{AIEntryDraft: entity.AIEntryDraft{Type: "expense", Amount: 35000, WalletID: "foreign"}}}}}}
	_, err := s.Send(context.Background(), "owner", "session", AIEntryMessageInput{RequestID: "00000000-0000-4000-8000-000000000001", Text: "ăn 35k", Timezone: "Asia/Ho_Chi_Minh"})
	if err != nil {
		t.Fatal(err)
	}
	if !r.finished || r.output.Drafts[0].WalletID != "w" || len(r.output.Drafts[0].Questions) == 0 {
		t.Fatalf("foreign wallet did not fall back safely: %+v", r.output)
	}
}

func TestAIEntryUsesSoleWalletWhenModelLeavesWalletBlank(t *testing.T) {
	r := &entryRepoStub{}
	s := &AIEntryService{Entries: r, Wallets: entryWallets{}, Categories: entryCategories{}, Extractor: entryExtractor{configured: true, out: entity.AIExtractOutput{Reply: "review", Drafts: []entity.AIExtractDraft{{AIEntryDraft: entity.AIEntryDraft{Type: "expense", Amount: 35000}}}}}}
	if _, err := s.Send(context.Background(), "owner", "session", AIEntryMessageInput{RequestID: "00000000-0000-4000-8000-000000000001", Text: "ăn 35k", Timezone: "Asia/Ho_Chi_Minh"}); err != nil {
		t.Fatal(err)
	}
	if !r.finished || r.output.Drafts[0].WalletID != "w" {
		t.Fatalf("sole wallet fallback missing: %+v", r.output)
	}
}
