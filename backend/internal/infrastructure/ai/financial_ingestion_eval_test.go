package ai

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/mypocket/backend/internal/config"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/infrastructure/financialsource"
)

func TestSemanticReviewScore(t *testing.T) {
	expenseCategory := "eval-expense"
	drafts := []entity.AIExtractDraft{
		{AIEntryDraft: entity.AIEntryDraft{Type: entity.TransactionTypeExpense, CategoryID: &expenseCategory, Note: "bounded note"}},
		{AIEntryDraft: entity.AIEntryDraft{Type: "unknown", Note: "needs review"}, Questions: []string{"Which transaction type applies?"}},
		{AIEntryDraft: entity.AIEntryDraft{Type: entity.TransactionTypeIncome, Note: "missing category and question"}},
		{AIEntryDraft: entity.AIEntryDraft{Type: entity.TransactionTypeExpense, CategoryID: &expenseCategory, Note: strings.Repeat("x", 257)}},
	}
	categories := []entity.Category{{ID: expenseCategory, Kind: entity.TransactionTypeExpense}}

	passed, total := semanticReviewScore(drafts, categories)
	if passed != 2 || total != 4 {
		t.Fatalf("semantic review score = %d/%d, want 2/4", passed, total)
	}
}

func TestAuditPDFProposalsClassifiesPrivacySafeCauses(t *testing.T) {
	drafts := []entity.AIExtractDraft{
		{AIEntryDraft: entity.AIEntryDraft{Type: entity.TransactionTypeExpense, Amount: 10, OccurredAt: "2026-10-01T00:00:00+07:00", Note: "first"}},
		{AIEntryDraft: entity.AIEntryDraft{Type: entity.TransactionTypeExpense, Amount: 10, OccurredAt: "2026-10-01T00:00:00+07:00", Note: "first"}},
		{AIEntryDraft: entity.AIEntryDraft{Type: entity.TransactionTypeExpense, Amount: 20, OccurredAt: "2026-09-30T00:00:00+07:00", Note: "outside"}},
		{AIEntryDraft: entity.AIEntryDraft{Type: entity.TransactionTypeExpense, Amount: 30, Note: "missing date"}},
		{AIEntryDraft: entity.AIEntryDraft{Type: entity.TransactionTypeExpense, Amount: 40, OccurredAt: "not-a-date", Note: "invalid date"}},
		{AIEntryDraft: entity.AIEntryDraft{Type: "unknown", Amount: 50, OccurredAt: "2026-10-02T00:00:00+07:00", Note: "uncertain"}},
		{AIEntryDraft: entity.AIEntryDraft{Type: entity.TransactionTypeExpense, Amount: 60, OccurredAt: "2026-10-03T00:00:00+07:00", Note: "Statement total"}},
		{AIEntryDraft: entity.AIEntryDraft{Type: entity.TransactionTypeIncome, Amount: 70, OccurredAt: "2026-10-04T00:00:00+07:00", Note: "second"}},
	}

	got := auditPDFProposals(drafts, 2)
	if got.Total != 8 || got.MissingInvalidDates != 2 || got.OutsideOctober != 1 || got.DuplicateFingerprints != 1 || got.UnknownSummaryLike != 2 || got.UnknownType != 1 || got.SummaryLike != 1 {
		t.Fatalf("overlapping PDF audit counts = %+v", got)
	}
	if got.PrimaryInvalidDate != 2 || got.PrimaryOutsideOctober != 1 || got.PrimaryDuplicate != 1 || got.PrimaryUnknownType != 1 || got.PrimarySummaryLike != 1 || got.Eligible != 2 || got.UnresolvedOverrun != 0 || got.SourceShortfall != 0 {
		t.Fatalf("primary PDF audit partition = %+v", got)
	}
}

type pdfProposalAudit struct {
	Total, MissingInvalidDates, OutsideOctober, DuplicateFingerprints int
	UnknownSummaryLike, UnknownType, SummaryLike                      int
	PrimaryInvalidDate, PrimaryOutsideOctober, PrimaryDuplicate       int
	PrimaryUnknownType, PrimarySummaryLike, Eligible                  int
	UnresolvedOverrun, SourceShortfall                                int
}

func auditPDFProposals(drafts []entity.AIExtractDraft, expectedSources int) pdfProposalAudit {
	type fingerprint struct {
		Date, Type string
		Amount     int64
		NoteHash   [32]byte
	}
	result := pdfProposalAudit{Total: len(drafts)}
	seen := make(map[fingerprint]struct{}, len(drafts))
	location := time.FixedZone("ICT", 7*60*60)
	for _, draft := range drafts {
		unknownType := draft.Type == "" || draft.Type == "unknown"
		summaryLike := isSummaryLike(draft)
		if unknownType {
			result.UnknownType++
		}
		if summaryLike {
			result.SummaryLike++
		}
		if unknownType || summaryLike {
			result.UnknownSummaryLike++
		}
		occurredAt, err := time.Parse(time.RFC3339, strings.TrimSpace(draft.OccurredAt))
		if err != nil {
			result.MissingInvalidDates++
			result.PrimaryInvalidDate++
			continue
		}
		localDate := occurredAt.In(location).Format("2006-01-02")
		outsideOctober := occurredAt.In(location).Year() != 2026 || occurredAt.In(location).Month() != time.October
		if outsideOctober {
			result.OutsideOctober++
		}
		key := fingerprint{
			Date: localDate, Type: draft.Type, Amount: draft.Amount,
			NoteHash: sha256.Sum256([]byte(strings.TrimSpace(draft.Note))),
		}
		_, duplicate := seen[key]
		if duplicate {
			result.DuplicateFingerprints++
		} else {
			seen[key] = struct{}{}
		}
		switch {
		case outsideOctober:
			result.PrimaryOutsideOctober++
		case duplicate:
			result.PrimaryDuplicate++
		case unknownType:
			result.PrimaryUnknownType++
		case summaryLike:
			result.PrimarySummaryLike++
		default:
			result.Eligible++
		}
	}
	if result.Eligible > expectedSources {
		result.UnresolvedOverrun = result.Eligible - expectedSources
	}
	if result.Eligible < expectedSources {
		result.SourceShortfall = expectedSources - result.Eligible
	}
	return result
}

func isSummaryLike(draft entity.AIExtractDraft) bool {
	if draft.Amount <= 0 {
		return true
	}
	note := strings.ToLower(strings.TrimSpace(draft.Note))
	for _, marker := range []string{"summary", "statement total", "total amount", "opening balance", "closing balance", "account balance"} {
		if strings.Contains(note, marker) {
			return true
		}
	}
	return false
}

func semanticReviewScore(drafts []entity.AIExtractDraft, categories []entity.Category) (int, int) {
	kinds := make(map[string]string, len(categories))
	for _, category := range categories {
		kinds[category.ID] = category.Kind
	}
	passed := 0
	for _, draft := range drafts {
		note := strings.TrimSpace(draft.Note)
		valid := note != "" && len(note) <= 256
		for _, question := range draft.Questions {
			valid = valid && strings.TrimSpace(question) != ""
		}
		switch draft.Type {
		case entity.TransactionTypeExpense, entity.TransactionTypeIncome:
			valid = valid && draft.CategoryID != nil && kinds[*draft.CategoryID] == draft.Type
		case "transfer", "unknown":
			valid = valid && draft.CategoryID == nil && len(draft.Questions) > 0
		default:
			valid = false
		}
		if valid {
			passed++
		}
	}
	return passed, len(drafts)
}

// TestPrivateFinancialIngestionConfiguredProviderEval is an opt-in,
// review-only benchmark. It sends private source content to the configured
// providers but never opens a database, proposal repository or approval path.
// Logs contain aggregate counts and timings only.
func TestPrivateFinancialIngestionConfiguredProviderEval(t *testing.T) {
	if os.Getenv("MYPOCKET_EVAL_PROVIDER_LIVE") != "1" {
		t.Skip("set MYPOCKET_EVAL_PROVIDER_LIVE=1 for the private configured-provider evaluation")
	}
	paths := []string{
		os.Getenv("MYPOCKET_EVAL_TECHCOMBANK_XLSX_A"),
		os.Getenv("MYPOCKET_EVAL_TECHCOMBANK_XLSX_B"),
		os.Getenv("MYPOCKET_EVAL_TECHCOMBANK_PDF"),
	}
	for _, path := range paths {
		if strings.TrimSpace(path) == "" {
			t.Fatal("all three private evaluation paths must be configured")
		}
	}
	if err := config.LoadLocalEnv(filepath.Join("..", "..", "..", ".env.local")); err != nil {
		t.Fatal("cannot load local provider configuration")
	}
	cfg := config.Load()
	client := NewClient(Config{
		BaseURL:    cfg.AIBaseURL,
		APIKey:     cfg.AIAPIKey,
		Model:      cfg.AIModel,
		OCRURL:     cfg.OCRAPIURL,
		OCRKey:     cfg.OCRAPIKey,
		StoreUsage: true,
	})
	if !client.Configured() {
		t.Fatal("configured AI provider is incomplete")
	}
	if !client.OCRConfigured() {
		t.Fatal("configured OCR provider is incomplete")
	}

	wallets := []entity.Wallet{{ID: "eval-wallet", Name: "Evaluation wallet", Type: entity.WalletTypeBasic, Currency: entity.WalletCurrencyVND}}
	categories := []entity.Category{
		{ID: "eval-expense", Name: "Other expense", Kind: entity.TransactionTypeExpense},
		{ID: "eval-income", Name: "Other income", Kind: entity.TransactionTypeIncome},
	}
	now := time.Date(2026, time.October, 27, 12, 0, 0, 0, time.FixedZone("ICT", 7*60*60))
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	structuredRows := make([]entity.AIStructuredRow, 0, 118)
	for index, path := range paths[:2] {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("cannot read private workbook %d", index+1)
		}
		document, err := financialsource.ParseWorkbook(fmt.Sprintf("private-statement-%d.xlsx", index+1), data, "Asia/Ho_Chi_Minh")
		if err != nil {
			t.Fatalf("cannot parse private workbook %d", index+1)
		}
		for _, row := range document.Rows {
			structuredRows = append(structuredRows, entity.AIStructuredRow{
				SourceID: row.SourceID, OccurredAt: row.OccurredAt.Format(time.RFC3339),
				Type: row.Type, Amount: row.Amount, SourceCurrency: row.SourceCurrency,
				Description: row.Description, Reference: row.Reference,
			})
		}
	}
	if len(structuredRows) != 118 {
		t.Fatalf("structured source count = %d, want 118", len(structuredRows))
	}
	structuredStart := time.Now()
	structuredOutput, err := client.Extract(ctx, Input{
		Instruction: "Add concise review semantics for every supplied structured row. Preserve every source ID and do not merge or omit repeated rows.",
		Timezone:    "Asia/Ho_Chi_Minh", Now: now, Wallets: wallets, Categories: categories,
		StructuredRows: structuredRows,
	})
	structuredWall := time.Since(structuredStart)
	if err != nil {
		failPrivateProvider(t, "structured", err, structuredWall)
	}
	if len(structuredOutput.Drafts) != len(structuredRows) {
		t.Fatalf("structured proposal/source count = %d/%d", len(structuredOutput.Drafts), len(structuredRows))
	}
	structuredSemantic, structuredTotal := semanticReviewScore(structuredOutput.Drafts, categories)
	t.Logf("PRIVATE EVAL XLSX model=%s sources=%d proposals=%d financial_exact=%d/%d semantic_review=%d/%d batches=%d provider_latency_ms=%d wall_ms=%d prompt_tokens=%d completion_tokens=%d",
		cfg.AIModel, len(structuredRows), len(structuredOutput.Drafts), len(structuredOutput.Drafts), len(structuredRows),
		structuredSemantic, structuredTotal, modelUsageInt(structuredOutput.ModelUsage, "batch_count"),
		modelUsageInt(structuredOutput.ModelUsage, "latency_ms"), structuredWall.Milliseconds(),
		modelUsageInt(structuredOutput.ModelUsage, "prompt_tokens"), modelUsageInt(structuredOutput.ModelUsage, "completion_tokens"))

	pdfData, err := os.ReadFile(paths[2])
	if err != nil || len(pdfData) < 5 || string(pdfData[:5]) != "%PDF-" {
		t.Fatal("cannot read the private PDF source")
	}
	pdfImage := Image{Name: "private-card-statement.pdf", MIMEType: "application/pdf", Base64: base64.StdEncoding.EncodeToString(pdfData)}
	ocrStart := time.Now()
	ocrText, err := client.ocr(ctx, pdfImage)
	ocrLatency := time.Since(ocrStart)
	if err != nil {
		failPrivateProvider(t, "pdf_ocr", err, ocrLatency)
	}
	modelStart := time.Now()
	pdfOutput, err := client.Extract(ctx, Input{
		Text:        ocrText,
		Instruction: "Extract every October 2026 card event row with its own row-level amount as a separate review proposal. Preserve repeated transaction rows. A missing event date must stay blank with a review question. Exclude September rows, headers, balances, statement or payment totals, opening or closing balances, credit limits, minimum amounts due, and every other aggregate or account-level line. Never turn a summary or balance line into a proposal.",
		Timezone:    "Asia/Ho_Chi_Minh", Now: now, Wallets: wallets, Categories: categories,
	})
	pdfModelWall := time.Since(modelStart)
	if err != nil {
		failPrivateProvider(t, "pdf_model", err, pdfModelWall)
	}
	pdfSemantic, pdfTotal := semanticReviewScore(pdfOutput.Drafts, categories)
	pdfAudit := auditPDFProposals(pdfOutput.Drafts, 31)
	t.Logf("PRIVATE EVAL PDF model=%s october_sources=%d proposals=%d semantic_review=%d/%d ocr_latency_ms=%d provider_latency_ms=%d model_wall_ms=%d end_to_end_ms=%d prompt_tokens=%d completion_tokens=%d",
		cfg.AIModel, 31, len(pdfOutput.Drafts), pdfSemantic, pdfTotal, ocrLatency.Milliseconds(),
		modelUsageInt(pdfOutput.ModelUsage, "latency_ms"), pdfModelWall.Milliseconds(), (ocrLatency + pdfModelWall).Milliseconds(),
		modelUsageInt(pdfOutput.ModelUsage, "prompt_tokens"), modelUsageInt(pdfOutput.ModelUsage, "completion_tokens"))
	t.Logf("PRIVATE EVAL PDF AUDIT total=%d eligible=%d missing_or_invalid_date=%d outside_october=%d duplicate_fingerprint=%d unknown_type=%d summary_like=%d unknown_or_summary=%d primary_invalid=%d primary_outside=%d primary_duplicate=%d primary_unknown=%d primary_summary=%d unresolved_overrun=%d source_shortfall=%d",
		pdfAudit.Total, pdfAudit.Eligible, pdfAudit.MissingInvalidDates, pdfAudit.OutsideOctober,
		pdfAudit.DuplicateFingerprints, pdfAudit.UnknownType, pdfAudit.SummaryLike, pdfAudit.UnknownSummaryLike,
		pdfAudit.PrimaryInvalidDate, pdfAudit.PrimaryOutsideOctober, pdfAudit.PrimaryDuplicate,
		pdfAudit.PrimaryUnknownType, pdfAudit.PrimarySummaryLike,
		pdfAudit.UnresolvedOverrun, pdfAudit.SourceShortfall)
	if len(pdfOutput.Drafts) != 31 || pdfAudit.PrimarySummaryLike != 0 {
		t.Errorf("PDF source-parity gate failed: proposals=%d expected=31 primary_summary=%d", len(pdfOutput.Drafts), pdfAudit.PrimarySummaryLike)
	}
}

func failPrivateProvider(t *testing.T, stage string, err error, elapsed time.Duration) {
	t.Helper()
	var diagnostic interface{ Diagnostic() (string, string, int) }
	if errors.As(err, &diagnostic) {
		providerStage, code, status := diagnostic.Diagnostic()
		t.Fatalf("%s provider evaluation failed: provider_stage=%s code=%s status=%d elapsed_ms=%d", stage, providerStage, code, status, elapsed.Milliseconds())
	}
	t.Fatalf("%s provider evaluation failed: error_type=%T elapsed_ms=%d", stage, err, elapsed.Milliseconds())
}

// TestPrivateTechcombankWorkbookTruthSet is opt-in because the owner-provided
// statements are private and must never be copied into repository fixtures.
// It records aggregate proof only; descriptions and references are not logged.
func TestPrivateTechcombankWorkbookTruthSet(t *testing.T) {
	paths := []string{
		os.Getenv("MYPOCKET_EVAL_TECHCOMBANK_XLSX_A"),
		os.Getenv("MYPOCKET_EVAL_TECHCOMBANK_XLSX_B"),
	}
	if paths[0] == "" || paths[1] == "" {
		t.Skip("set both MYPOCKET_EVAL_TECHCOMBANK_XLSX_A and MYPOCKET_EVAL_TECHCOMBANK_XLSX_B")
	}

	type aggregate struct {
		rows, debits, credits   int
		debitTotal, creditTotal int64
		digest                  string
	}
	want := map[int]aggregate{
		66: {rows: 66, debits: 46, credits: 20, debitTotal: 60_054_891, creditTotal: 62_159_320, digest: "080f1ab8050f46ed639869a2a5ed834fbc2ad713e1bda7240a0716a239e97392"},
		52: {rows: 52, debits: 47, credits: 5, debitTotal: 19_220_273, creditTotal: 9_551_000, digest: "4c226c386c698f9435f4f240a452167a71f913290a53a17507cf79efbf06e4ef"},
	}

	for index, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read private workbook %d", index+1)
		}
		document, err := financialsource.ParseWorkbook("private-statement.xlsx", data, "Asia/Ho_Chi_Minh")
		if err != nil {
			t.Fatalf("parse private workbook %d", index+1)
		}
		got := aggregate{rows: len(document.Rows)}
		seen := make(map[string]struct{}, len(document.Rows))
		manifest := make([]string, 0, len(document.Rows))
		for _, row := range document.Rows {
			if _, exists := seen[row.SourceID]; exists {
				t.Fatalf("private workbook %d produced duplicate source identity", index+1)
			}
			seen[row.SourceID] = struct{}{}
			switch row.Type {
			case "expense":
				got.debits++
				got.debitTotal += row.Amount
			case "income":
				got.credits++
				got.creditTotal += row.Amount
			default:
				t.Fatalf("private workbook %d produced unsupported direction", index+1)
			}
			manifest = append(manifest, fmt.Sprintf("%s|%s|%d|%s|%s|%s", row.OccurredAt.Format("2006-01-02"), row.Type, row.Amount, row.SourceCurrency, row.Reference, row.SourceID))
		}
		sort.Strings(manifest)
		digest := sha256.Sum256([]byte(strings.Join(manifest, "\n")))
		got.digest = hex.EncodeToString(digest[:])
		expected, exists := want[got.rows]
		if !exists || got != expected {
			t.Fatalf("private workbook %d aggregate fingerprint = %+v", index+1, got)
		}
	}
}
