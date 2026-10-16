package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/mypocket/backend/internal/entity"
	port "github.com/mypocket/backend/internal/repository"
)

type aiEvalFixtures struct {
	Version      int    `json:"version"`
	ProviderMode string `json:"provider_mode"`
	Clock        struct {
		Now      string `json:"now"`
		Timezone string `json:"timezone"`
	} `json:"clock"`
	ReviewPolicy struct {
		ApprovalRequired           bool   `json:"approval_required"`
		ProposalStatusBeforeAction string `json:"proposal_status_before_owner_action"`
		LedgerWritesAllowed        bool   `json:"ledger_writes_allowed"`
	} `json:"review_policy"`
	DuplicateRows struct {
		Drafts   []entity.AIExtractDraft `json:"drafts"`
		Expected struct {
			DraftCount     int     `json:"draft_count"`
			Amounts        []int64 `json:"amounts"`
			MergedQuestion string  `json:"merged_question"`
		} `json:"expected"`
	} `json:"duplicate_rows"`
	MissingDate struct {
		RawOccurredAt string `json:"raw_occurred_at"`
		Expected      struct {
			NormalizedOccurredAt string `json:"normalized_occurred_at"`
			Parsed               bool   `json:"parsed"`
			DateOnly             bool   `json:"date_only"`
			ReviewQuestion       string `json:"review_question_contains"`
		} `json:"expected"`
	} `json:"missing_date"`
	WalletInference struct {
		OwnerID    string                `json:"owner_id"`
		Wallets    []entity.Wallet       `json:"wallets"`
		Categories []entity.Category     `json:"categories"`
		Draft      entity.AIExtractDraft `json:"draft"`
		Expected   struct {
			WalletID        string `json:"wallet_id"`
			ReviewQuestion  string `json:"review_question_contains"`
			OwnerWalletOnly bool   `json:"owner_wallet_only"`
		} `json:"expected"`
	} `json:"wallet_inference"`
	OwnerSafeToolArgs struct {
		OwnerID  string `json:"owner_id"`
		Timezone string `json:"timezone"`
		Accepted struct {
			Tool      string          `json:"tool"`
			Arguments json.RawMessage `json:"arguments"`
		} `json:"accepted"`
		Rejected []struct {
			Tool      string          `json:"tool"`
			Arguments json.RawMessage `json:"arguments"`
		} `json:"rejected"`
		Expected struct {
			ReaderOwner         string `json:"reader_owner"`
			RejectedReaderCalls int    `json:"rejected_reader_calls"`
			NoAutoApproval      bool   `json:"no_auto_approval"`
		} `json:"expected"`
	} `json:"owner_safe_tool_args"`
	ProposalReview struct {
		OwnerID   string          `json:"owner_id"`
		Tool      string          `json:"tool"`
		Arguments json.RawMessage `json:"arguments"`
		Expected  struct {
			Status                       string `json:"status"`
			View                         string `json:"view"`
			RequiresExplicitConfirmation bool   `json:"requires_explicit_confirmation"`
		} `json:"expected"`
	} `json:"proposal_review"`
}

func loadAIEvalFixtures(t *testing.T) aiEvalFixtures {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not locate fixture test source")
	}
	path := filepath.Join(filepath.Dir(filename), "..", "..", "..", "evals", "ai", "fixtures.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read deterministic AI fixture %s: %v", path, err)
	}
	var fixtures aiEvalFixtures
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatalf("decode deterministic AI fixture: %v", err)
	}
	return fixtures
}

func TestAIEvalFixtures(t *testing.T) {
	fixtures := loadAIEvalFixtures(t)
	if fixtures.Version != 1 || fixtures.ProviderMode != "synthetic-only" {
		t.Fatalf("fixture pack must remain versioned and provider-free: version=%d mode=%q", fixtures.Version, fixtures.ProviderMode)
	}
	if fixtures.ReviewPolicy.ApprovalRequired != true || fixtures.ReviewPolicy.LedgerWritesAllowed {
		t.Fatalf("fixture review policy weakened: %+v", fixtures.ReviewPolicy)
	}
	if fixtures.ReviewPolicy.ProposalStatusBeforeAction != entity.AIProposalPending {
		t.Fatalf("fixture proposals must remain pending before owner action: %q", fixtures.ReviewPolicy.ProposalStatusBeforeAction)
	}

	t.Run("duplicate rows retain distinct purchases and merge exact overlap", func(t *testing.T) {
		got := mergeExactAIDrafts(fixtures.DuplicateRows.Drafts)
		if len(got) != fixtures.DuplicateRows.Expected.DraftCount {
			t.Fatalf("draft count=%d want=%d: %+v", len(got), fixtures.DuplicateRows.Expected.DraftCount, got)
		}
		amounts := make([]int64, 0, len(got))
		for _, draft := range got {
			amounts = append(amounts, draft.Amount)
		}
		sort.Slice(amounts, func(i, j int) bool { return amounts[i] < amounts[j] })
		if !equalInt64Slices(amounts, fixtures.DuplicateRows.Expected.Amounts) {
			t.Fatalf("distinct amounts were lost: got=%v want=%v", amounts, fixtures.DuplicateRows.Expected.Amounts)
		}
		foundQuestion := false
		for _, draft := range got {
			for _, question := range draft.Questions {
				foundQuestion = foundQuestion || question == fixtures.DuplicateRows.Expected.MergedQuestion
			}
		}
		if !foundQuestion {
			t.Fatalf("merged overlap review question missing: %+v", got)
		}
	})

	t.Run("missing date is fixed clock and disclosed", func(t *testing.T) {
		now, err := time.Parse(time.RFC3339, fixtures.Clock.Now)
		if err != nil {
			t.Fatal(err)
		}
		normalized, parsed, dateOnly := normalizeAIEntryOccurredAt(fixtures.MissingDate.RawOccurredAt, now)
		if normalized != fixtures.MissingDate.Expected.NormalizedOccurredAt || parsed != fixtures.MissingDate.Expected.Parsed || dateOnly != fixtures.MissingDate.Expected.DateOnly {
			t.Fatalf("missing date normalization=%q parsed=%t date_only=%t; want %q/%t/%t", normalized, parsed, dateOnly, fixtures.MissingDate.Expected.NormalizedOccurredAt, fixtures.MissingDate.Expected.Parsed, fixtures.MissingDate.Expected.DateOnly)
		}
		draft := entity.AIExtractDraft{}
		appendAIQuestion(&draft, "Ngày trên chứng từ chưa rõ; đã tự điền ngày hiện tại, hãy kiểm tra lại.")
		if !strings.Contains(strings.Join(draft.Questions, " "), fixtures.MissingDate.Expected.ReviewQuestion) {
			t.Fatalf("missing-date review question not disclosed: %v", draft.Questions)
		}
	})

	t.Run("wallet inference stays owner-scoped and reviewable", func(t *testing.T) {
		repo := &evalEntryRepoStub{}
		service := &AIEntryService{
			Entries:    repo,
			Wallets:    evalWalletRepository{wallets: fixtures.WalletInference.Wallets},
			Categories: evalCategoryRepository{categories: fixtures.WalletInference.Categories},
			Extractor:  entryExtractor{configured: true, out: entity.AIExtractOutput{Drafts: []entity.AIExtractDraft{fixtures.WalletInference.Draft}}},
		}
		if _, err := service.Send(context.Background(), fixtures.WalletInference.OwnerID, "session-wallet-inference", AIEntryMessageInput{RequestID: "00000000-0000-0000-0000-000000000101", Text: "ăn trưa", Timezone: fixtures.Clock.Timezone}); err != nil {
			t.Fatal(err)
		}
		if len(repo.output.Drafts) != 1 || repo.output.Drafts[0].WalletID != fixtures.WalletInference.Expected.WalletID {
			t.Fatalf("wallet inference=%+v want=%q", repo.output.Drafts, fixtures.WalletInference.Expected.WalletID)
		}
		if !strings.Contains(strings.Join(repo.output.Drafts[0].Questions, " "), fixtures.WalletInference.Expected.ReviewQuestion) {
			t.Fatalf("inferred wallet must disclose review question: %v", repo.output.Drafts[0].Questions)
		}
		for _, wallet := range fixtures.WalletInference.Wallets {
			if wallet.ID == repo.output.Drafts[0].WalletID && wallet.OwnerID != fixtures.WalletInference.OwnerID && fixtures.WalletInference.Expected.OwnerWalletOnly {
				t.Fatalf("inferred foreign wallet: %+v", wallet)
			}
		}
		if repo.decideCalls != 0 || len(repo.proposalStatuses) != 1 || repo.proposalStatuses[0] != entity.AIProposalPending {
			t.Fatalf("wallet inference must not auto-approve: decide_calls=%d statuses=%v", repo.decideCalls, repo.proposalStatuses)
		}
	})

	t.Run("advisor tool arguments use authenticated owner and reject overrides", func(t *testing.T) {
		fixture := fixtures.OwnerSafeToolArgs
		reader := &financeReaderStub{bundle: entity.FactBundle{Results: []entity.FinanceResult{{QueryKey: "q1", Status: "ok", ViewKind: "finance_summary", View: json.RawMessage(`{"expense":150}`)}}}}
		registry := NewAdvisorToolRegistry(NewFinanceQueryService(reader))
		tool, ok := registry.Lookup(fixture.Accepted.Tool)
		if !ok {
			t.Fatalf("accepted fixture tool missing: %s", fixture.Accepted.Tool)
		}
		if _, err := tool.Execute(context.Background(), Principal{OwnerID: fixture.OwnerID, Timezone: fixture.Timezone}, fixture.Accepted.Arguments); err != nil {
			t.Fatalf("accepted fixture args rejected: %v", err)
		}
		if reader.owner != fixture.Expected.ReaderOwner || len(reader.queries) != 1 {
			t.Fatalf("reader owner/queries=%q/%d want %q/1", reader.owner, len(reader.queries), fixture.Expected.ReaderOwner)
		}
		query := reader.queries[0]
		if query.Filter.Dates.From != "2026-09-01" || query.Filter.Dates.To != "2026-09-30" {
			t.Fatalf("accepted date range=%+v", query.Filter.Dates)
		}
		for _, rejected := range fixture.Rejected {
			before := len(reader.queries)
			tool, exists := registry.Lookup(rejected.Tool)
			if !exists {
				if rejected.Tool != "execute_sql" {
					t.Fatalf("unexpected missing fixture tool: %s", rejected.Tool)
				}
				continue
			}
			if _, err := tool.Execute(context.Background(), Principal{OwnerID: fixture.OwnerID, Timezone: fixture.Timezone}, rejected.Arguments); err == nil {
				t.Fatalf("owner/unsupported override accepted by %s", rejected.Tool)
			}
			if len(reader.queries) != before {
				t.Fatalf("rejected args reached reader: tool=%s queries=%+v", rejected.Tool, reader.queries)
			}
		}
		if len(reader.queries) != 1+fixture.Expected.RejectedReaderCalls {
			t.Fatalf("rejected arguments changed reader call count: got=%d want=%d", len(reader.queries), 1+fixture.Expected.RejectedReaderCalls)
		}
		if fixture.Expected.NoAutoApproval && !fixtures.ReviewPolicy.ApprovalRequired {
			t.Fatal("fixture contradicts no-auto-approval review policy")
		}
	})

	t.Run("proposal tool remains pending until explicit confirmation", func(t *testing.T) {
		fixture := fixtures.ProposalReview
		tool, ok := NewAdvisorToolRegistry(NewFinanceQueryService(&financeReaderStub{})).Lookup(fixture.Tool)
		if !ok {
			t.Fatalf("proposal tool missing: %s", fixture.Tool)
		}
		result, err := tool.Execute(context.Background(), Principal{OwnerID: fixture.OwnerID, Timezone: fixtures.Clock.Timezone}, fixture.Arguments)
		if err != nil {
			t.Fatal(err)
		}
		if result.Status != fixture.Expected.Status || result.ViewKind != fixture.Expected.View {
			t.Fatalf("proposal status/view=%q/%q want %q/%q", result.Status, result.ViewKind, fixture.Expected.Status, fixture.Expected.View)
		}
		if result.Status == entity.AIProposalApproved || !fixture.Expected.RequiresExplicitConfirmation {
			t.Fatalf("proposal bypassed explicit confirmation: %+v", result)
		}
	})
}

func TestAIEvalFixturesContainNoProviderSecrets(t *testing.T) {
	fixtures := loadAIEvalFixtures(t)
	data, err := json.Marshal(fixtures)
	if err != nil {
		t.Fatal(err)
	}
	lower := strings.ToLower(string(data))
	for _, forbidden := range []string{"api_key", "bearer ", "sk-", "secret", "authorization"} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("fixture contains forbidden provider credential marker %q", forbidden)
		}
	}
}

type evalEntryRepoStub struct {
	entryRepoStub
	proposalStatuses []string
	decideCalls      int
}

func (r *evalEntryRepoStub) FinishMessage(_ context.Context, _, _, _ string, output entity.AIExtractOutput) error {
	r.finished = true
	r.output = output
	for range output.Drafts {
		r.proposalStatuses = append(r.proposalStatuses, entity.AIProposalPending)
	}
	return nil
}

func (r *evalEntryRepoStub) DecideProposal(context.Context, string, string, int, bool) (*entity.AIEntryProposal, error) {
	r.decideCalls++
	return nil, errors.New("proposal decision is not part of extraction")
}

type evalWalletRepository struct {
	port.WalletRepository
	wallets []entity.Wallet
}

func (r evalWalletRepository) List(string) ([]entity.Wallet, error) {
	return append([]entity.Wallet(nil), r.wallets...), nil
}

type evalCategoryRepository struct {
	port.CategoryRepository
	categories []entity.Category
}

func (r evalCategoryRepository) ListVisible(string) ([]entity.Category, error) {
	return append([]entity.Category(nil), r.categories...), nil
}

func equalInt64Slices(left, right []int64) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
