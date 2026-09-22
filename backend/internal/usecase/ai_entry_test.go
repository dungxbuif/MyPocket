package usecase

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/mypocket/backend/internal/entity"
	port "github.com/mypocket/backend/internal/repository"
	"strings"
	"testing"
	"time"
)

func TestAIEntrySessionSerializesSafeErrorCode(t *testing.T) {
	data, err := json.Marshal(entity.AIEntrySession{ID: "session", Error: "OCR không đọc được chứng từ.", ErrorCode: "ocr_failed"})
	if err != nil {
		t.Fatal(err)
	}
	if string(data) == "" || !strings.Contains(string(data), `"error_code":"ocr_failed"`) {
		t.Fatalf("error code missing from session JSON: %s", data)
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
		if i < len(texts) {
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

func TestAIEntryStoresOriginalPrivatelyAndPassesOnlyReadURLToOCR(t *testing.T) {
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
	if len(seen.Images) != 1 || seen.Images[0].Base64 != "" || seen.Images[0].SourceURL != storage.signedURL {
		t.Fatalf("OCR input must carry only a short-lived private read URL: %+v", seen.Images)
	}
}

type attachmentStoreStub struct {
	signedURL, contentType string
	data                   []byte
	putErr                 error
	deleteErr              error
	deleted                int
}

func (s *attachmentStoreStub) Key(string, string, string, string) (string, error) {
	return "private/env/key.pdf", nil
}
func (s *attachmentStoreStub) Put(_ context.Context, _ string, contentType string, data []byte) error {
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

func TestAIEntryUnknownWalletCannotBePrefilledAsOwned(t *testing.T) {
	r := &entryRepoStub{}
	s := &AIEntryService{Entries: r, Wallets: entryWallets{}, Categories: entryCategories{}, Extractor: entryExtractor{configured: true, out: entity.AIExtractOutput{Reply: "review", Drafts: []entity.AIExtractDraft{{AIEntryDraft: entity.AIEntryDraft{Type: "expense", Amount: 35000, WalletID: "foreign"}}}}}}
	_, err := s.Send(context.Background(), "owner", "session", AIEntryMessageInput{RequestID: "00000000-0000-4000-8000-000000000001", Text: "ăn 35k", Timezone: "Asia/Ho_Chi_Minh"})
	if err != nil {
		t.Fatal(err)
	}
	if !r.finished || r.output.Drafts[0].WalletID != "" || len(r.output.Drafts[0].Questions) == 0 {
		t.Fatalf("foreign wallet not cleared: %+v", r.output)
	}
}
