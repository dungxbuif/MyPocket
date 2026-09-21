package usecase

import (
	"context"
	"errors"
	"github.com/mypocket/backend/internal/entity"
	port "github.com/mypocket/backend/internal/repository"
	"strings"
	"testing"
)

type entryRepoStub struct {
	port.AIEntryRepository
	begins        int
	failed        bool
	finished      bool
	output        entity.AIExtractOutput
	history       []entity.AIEntryMessage
	fresh         string
	failureSource string
}

func (r *entryRepoStub) Session(context.Context, string, string) (*entity.AIEntrySession, error) {
	return &entity.AIEntrySession{ID: "session", Messages: r.history}, nil
}
func (r *entryRepoStub) BeginMessage(context.Context, string, string, string, string, string) (string, bool, error) {
	r.begins++
	if r.fresh != "" {
		r.history = append(r.history, entity.AIEntryMessage{Role: "assistant", Content: r.fresh})
	}
	return "token", true, nil
}

func TestAIEntryUsesHistoryAfterLeaseClaim(t *testing.T) {
	r := &entryRepoStub{fresh: "newly completed proposal"}
	var seen entity.AIExtractInput
	s := &AIEntryService{Entries: r, Wallets: entryWallets{}, Categories: entryCategories{}, Extractor: entryExtractor{configured: true, seen: &seen}}
	if _, err := s.Send(context.Background(), "owner", "session", AIEntryMessageInput{RequestID: "00000000-0000-4000-8000-000000000001", Text: "chọn ví w", Timezone: "Asia/Ho_Chi_Minh"}); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, m := range seen.History {
		if m.Content == r.fresh {
			found = true
		}
	}
	if !found {
		t.Fatal("used snapshot before previous message completed")
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
func (r *entryRepoStub) FailMessage(_ context.Context, _, _, _, _, sourceText string) error {
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
	configured bool
	err        error
	out        entity.AIExtractOutput
	seen       *entity.AIExtractInput
}

func (e entryExtractor) Configured() bool    { return e.configured }
func (e entryExtractor) OCRConfigured() bool { return false }
func (e entryExtractor) Extract(_ context.Context, input entity.AIExtractInput) (entity.AIExtractOutput, error) {
	if e.seen != nil {
		*e.seen = input
	}
	return e.out, e.err
}

func TestAIEntryCanContinueAfterLongOCRHistory(t *testing.T) {
	r := &entryRepoStub{history: []entity.AIEntryMessage{{Role: "user", Content: strings.Repeat("ăn ", 4000)}, {Role: "assistant", Content: "review", SourceText: strings.Repeat("OCR evidence ", 5000)}}}
	var seen entity.AIExtractInput
	s := &AIEntryService{Entries: r, Wallets: entryWallets{}, Categories: entryCategories{}, Extractor: entryExtractor{configured: true, seen: &seen}}
	if _, err := s.Send(context.Background(), "owner", "session", AIEntryMessageInput{RequestID: "00000000-0000-4000-8000-000000000001", Text: "chọn ví w", Timezone: "Asia/Ho_Chi_Minh"}); err != nil {
		t.Fatal(err)
	}
	for _, m := range seen.History {
		if len(m.Content) > 8192 {
			t.Fatalf("history prevents follow-up extraction: %d bytes", len(m.Content))
		}
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
