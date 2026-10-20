package agent_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"mypocket/internal/agent"
	"mypocket/internal/finance"
	platformdb "mypocket/internal/platform/db"
)

func TestAgentCompletionCreatesReviewDraftWithoutAccounting(t *testing.T) {
	db := agentDB(t)
	ctx := context.Background()
	user := agentUser(t, db, "agent-owner@example.com")
	other := agentUser(t, db, "agent-other@example.com")
	financeRepo := finance.NewRepository(db)
	wallet, err := financeRepo.CreateWallet(ctx, user, finance.CreateWalletInput{Name: "Cash", Type: finance.WalletBasic})
	if err != nil {
		t.Fatal(err)
	}
	foreign, err := financeRepo.CreateWallet(ctx, other, finance.CreateWalletInput{Name: "Foreign", Type: finance.WalletBasic})
	if err != nil {
		t.Fatal(err)
	}
	var category string
	if err := db.QueryRow(`SELECT id::text FROM categories WHERE system_key='expense_food'`).Scan(&category); err != nil {
		t.Fatal(err)
	}
	repo := agent.NewRepository(db)
	run, err := repo.CreateRun(ctx, user, "once", agent.KindTransactionDraft, "Lunch")
	if err != nil {
		t.Fatal(err)
	}
	run, claimed, err := repo.ClaimDue(ctx, "test", time.Now().UTC(), time.Minute)
	if err != nil || !claimed {
		t.Fatalf("claim=%v err=%v", claimed, err)
	}
	result, err := agent.ParseModelResult(`{"transaction":{"type":"expense","amount_vnd":120000,"source_wallet_id":"`+wallet.ID+`","category_id":"`+category+`","occurred_at":"2026-09-11T00:00:00Z","note":"Lunch"}}`, agent.KindTransactionDraft)
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.Complete(ctx, run, result, map[string]any{"provider": "test"}); err != nil {
		t.Fatal(err)
	}
	loaded, err := repo.GetRun(ctx, user, run.ID)
	if err != nil || loaded.Status != agent.StatusCompleted || len(loaded.DraftIDs) != 1 {
		t.Fatalf("loaded=%#v err=%v", loaded, err)
	}
	var transactions int
	var balance int64
	if err := db.QueryRow(`SELECT count(*) FROM transactions WHERE user_id=$1`, user).Scan(&transactions); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT balance_vnd FROM wallets WHERE id=$1`, wallet.ID).Scan(&balance); err != nil {
		t.Fatal(err)
	}
	if transactions != 0 || balance != 0 {
		t.Fatalf("agent bypassed review boundary transactions=%d balance=%d", transactions, balance)
	}

	second, _ := repo.CreateRun(ctx, user, "foreign", agent.KindTransactionDraft, "bad")
	second, _, _ = repo.ClaimDue(ctx, "test", time.Now().UTC(), time.Minute)
	bad, _ := agent.ParseModelResult(`{"transaction":{"type":"expense","amount_vnd":1,"source_wallet_id":"`+foreign.ID+`","category_id":"`+category+`","occurred_at":"2026-09-11T00:00:00Z","note":""}}`, agent.KindTransactionDraft)
	if err := repo.Complete(ctx, second, bad, nil); !errors.Is(err, agent.ErrValidation) {
		t.Fatalf("expected foreign reference rejection, got %v", err)
	}
}

func TestAgentIntakeRejectsTransfersAndNeverWritesAccounting(t *testing.T) {
	db := agentDB(t)
	ctx := context.Background()
	user := agentUser(t, db, "intake-transfer@example.com")
	financeRepo := finance.NewRepository(db)
	source, err := financeRepo.CreateWallet(ctx, user, finance.CreateWalletInput{Name: "Cash", Type: finance.WalletBasic})
	if err != nil {
		t.Fatal(err)
	}
	destination, err := financeRepo.CreateWallet(ctx, user, finance.CreateWalletInput{Name: "Bank", Type: finance.WalletBasic})
	if err != nil {
		t.Fatal(err)
	}
	_, err = agent.ParseModelResult(`{"transaction":{"type":"transfer","amount_vnd":100000,"source_wallet_id":"`+source.ID+`","destination_wallet_id":"`+destination.ID+`","occurred_at":"2026-09-11T00:00:00Z","note":"move"}}`, agent.KindIntake)
	if !errors.Is(err, agent.ErrValidation) {
		t.Fatalf("expected intake transfer rejection, got %v", err)
	}
	var txCount, draftCount int
	if err := db.QueryRow(`SELECT count(*) FROM transactions WHERE user_id=$1`, user).Scan(&txCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT count(*) FROM transaction_drafts WHERE user_id=$1`, user).Scan(&draftCount); err != nil {
		t.Fatal(err)
	}
	if txCount != 0 || draftCount != 0 {
		t.Fatalf("intake transfer wrote state transactions=%d drafts=%d", txCount, draftCount)
	}
}

func TestAgentIntakeCreatesMultipleReviewDrafts(t *testing.T) {
	db := agentDB(t)
	ctx := context.Background()
	user := agentUser(t, db, "intake-multiple@example.com")
	financeRepo := finance.NewRepository(db)
	wallet, err := financeRepo.CreateWallet(ctx, user, finance.CreateWalletInput{Name: "Cash", Type: finance.WalletBasic})
	if err != nil {
		t.Fatal(err)
	}
	var category string
	if err := db.QueryRow(`SELECT id::text FROM categories WHERE system_key='expense_food'`).Scan(&category); err != nil {
		t.Fatal(err)
	}
	repo := agent.NewRepository(db)
	run, err := repo.CreateRun(ctx, user, "multi-intake", agent.KindIntake, "Ăn trưa 80k, taxi 120k")
	if err != nil {
		t.Fatal(err)
	}
	run, _, _ = repo.ClaimDue(ctx, "test", time.Now().UTC(), time.Minute)
	result, err := agent.ParseModelResult(`{"drafts":[{"type":"expense","amount_vnd":80000,"source_wallet_id":"`+wallet.ID+`","category_id":"`+category+`","occurred_at":"2026-09-11T00:00:00Z","note":"Ăn trưa"},{"type":"expense","amount_vnd":120000,"source_wallet_id":"`+wallet.ID+`","category_id":"`+category+`","occurred_at":"2026-09-11T01:00:00Z","note":"Taxi"}]}`, agent.KindIntake)
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.Complete(ctx, run, result, map[string]any{"provider": "test"}); err != nil {
		t.Fatal(err)
	}
	loaded, err := repo.GetRun(ctx, user, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Status != agent.StatusCompleted || len(loaded.DraftIDs) != 2 {
		t.Fatalf("loaded=%#v", loaded)
	}
	var txCount int
	if err := db.QueryRow(`SELECT count(*) FROM transactions WHERE user_id=$1`, user).Scan(&txCount); err != nil {
		t.Fatal(err)
	}
	if txCount != 0 {
		t.Fatalf("intake bypassed review boundary transactions=%d", txCount)
	}
}

func TestAgentIntakePersistsSessionMessagesAndDraftActionCard(t *testing.T) {
	db := agentDB(t)
	ctx := context.Background()
	user := agentUser(t, db, "intake-history@example.com")
	financeRepo := finance.NewRepository(db)
	wallet, err := financeRepo.CreateWallet(ctx, user, finance.CreateWalletInput{Name: "Cash", Type: finance.WalletBasic})
	if err != nil {
		t.Fatal(err)
	}
	var category string
	if err := db.QueryRow(`SELECT id::text FROM categories WHERE system_key='expense_food'`).Scan(&category); err != nil {
		t.Fatal(err)
	}
	repo := agent.NewRepository(db)
	run, session, userMessage, err := repo.CreateSessionRun(ctx, user, "", "intake-once", agent.KindIntake, "Ăn trưa 80k", "")
	if err != nil {
		t.Fatal(err)
	}
	if session.Kind != agent.KindIntake || userMessage.Role != agent.MessageRoleUser || userMessage.Text != "Ăn trưa 80k" || userMessage.RunID != run.ID {
		t.Fatalf("session=%#v userMessage=%#v run=%#v", session, userMessage, run)
	}
	run, _, _ = repo.ClaimDue(ctx, "test", time.Now().UTC(), time.Minute)
	result, err := agent.ParseModelResult(`{"drafts":[{"type":"expense","amount_vnd":80000,"source_wallet_id":"`+wallet.ID+`","category_id":"`+category+`","occurred_at":"2026-09-11T00:00:00Z","note":"Ăn trưa"}]}`, agent.KindIntake)
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.Complete(ctx, run, result, map[string]any{"provider": "test"}); err != nil {
		t.Fatal(err)
	}
	history, err := repo.GetSession(ctx, user, session.ID, 20)
	if err != nil {
		t.Fatal(err)
	}
	if history.Session.ID != session.ID || len(history.Messages) != 2 {
		t.Fatalf("history=%#v", history)
	}
	if history.Messages[0].Role != agent.MessageRoleUser || history.Messages[0].Text != "Ăn trưa 80k" {
		t.Fatalf("user message not persisted: %#v", history.Messages[0])
	}
	assistantMessage := history.Messages[1]
	if assistantMessage.Role != agent.MessageRoleAssistant || assistantMessage.RunID != run.ID {
		t.Fatalf("assistant message not linked to run: %#v", assistantMessage)
	}
	if assistantMessage.Action == nil || assistantMessage.Action.Type != agent.ActionDraftsCreated || len(assistantMessage.Action.DraftIDs) != 1 {
		t.Fatalf("assistant action card missing draft_ids: %#v", assistantMessage.Action)
	}
}

func TestAgentAdvisorCompletionNeverCreatesDrafts(t *testing.T) {
	db := agentDB(t)
	ctx := context.Background()
	user := agentUser(t, db, "advisor-readonly@example.com")
	repo := agent.NewRepository(db)
	run, err := repo.CreateRun(ctx, user, "advisor", agent.KindAdvisor, "Tháng này thế nào?")
	if err != nil {
		t.Fatal(err)
	}
	run, _, _ = repo.ClaimDue(ctx, "test", time.Now().UTC(), time.Minute)
	result, err := agent.ParseModelResult(`{"answer":"Tháng này chi 0 VND."}`, agent.KindAdvisor)
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.Complete(ctx, run, result, map[string]any{"provider": "test"}); err != nil {
		t.Fatal(err)
	}
	loaded, err := repo.GetRun(ctx, user, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Status != agent.StatusCompleted || len(loaded.DraftIDs) != 0 {
		t.Fatalf("advisor wrote drafts or failed: %#v", loaded)
	}
}

func TestAgentWaitsForOwnedOCRToolAndExposesOnlyThatResult(t *testing.T) {
	db := agentDB(t)
	ctx := context.Background()
	user := agentUser(t, db, "ocr-owner@example.com")
	other := agentUser(t, db, "ocr-other@example.com")
	financeRepo := finance.NewRepository(db)
	receipt, err := financeRepo.CreateReceiptObject(ctx, user, finance.CreateReceiptObjectInput{ObjectKey: "users/" + user + "/receipts/a.jpg", ContentType: "image/jpeg", SizeBytes: 5, ChecksumSHA256: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", OriginalFilename: "a.jpg"})
	if err != nil {
		t.Fatal(err)
	}
	foreign, err := financeRepo.CreateReceiptObject(ctx, other, finance.CreateReceiptObjectInput{ObjectKey: "users/" + other + "/receipts/b.jpg", ContentType: "image/jpeg", SizeBytes: 5, ChecksumSHA256: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", OriginalFilename: "b.jpg"})
	if err != nil {
		t.Fatal(err)
	}
	repo := agent.NewRepository(db)
	if _, err = repo.CreateRunWithTool(ctx, user, "foreign-image", agent.KindAnalysis, "read", foreign.ID); !errors.Is(err, agent.ErrNotFound) {
		t.Fatalf("expected foreign image hiding, got %v", err)
	}
	run, err := repo.CreateRunWithTool(ctx, user, "image", agent.KindAnalysis, "read", receipt.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, claimed, err := repo.ClaimDue(ctx, "agent", time.Now().UTC(), time.Minute); err != nil || claimed {
		t.Fatalf("agent must wait for OCR claimed=%v err=%v", claimed, err)
	}
	tool, claimed, err := repo.ClaimToolDue(ctx, "ocr", time.Now().UTC(), time.Minute)
	if err != nil || !claimed || tool.AgentRunID != run.ID || tool.ObjectKey != receipt.ObjectKey {
		t.Fatalf("tool=%#v claimed=%v err=%v", tool, claimed, err)
	}
	if err = repo.MarkToolSubmitted(ctx, tool, agent.ToolSubmission{ProviderID: "doc", Status: agent.ToolProcessing}, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	tool, claimed, err = repo.ClaimToolDue(ctx, "ocr", time.Now().UTC().Add(time.Second), time.Minute)
	if err != nil || !claimed {
		t.Fatalf("poll claim=%v err=%v", claimed, err)
	}
	if err = repo.CompleteTool(ctx, tool, agent.ToolResult{Status: agent.ToolCompleted, Text: "Total 120000", Fields: map[string]any{"total": 120000}}); err != nil {
		t.Fatal(err)
	}
	claimedRun, claimed, err := repo.ClaimDue(ctx, "agent", time.Now().UTC(), time.Minute)
	if err != nil || !claimed || claimedRun.ID != run.ID {
		t.Fatalf("run=%#v claimed=%v err=%v", claimedRun, claimed, err)
	}
	contextJSON, err := repo.Context(ctx, user, run.ID)
	if err != nil || !strings.Contains(contextJSON, "Total 120000") {
		t.Fatalf("context=%s err=%v", contextJSON, err)
	}
	var count int
	if err = db.QueryRow(`SELECT count(*) FROM transactions WHERE user_id=$1`, user).Scan(&count); err != nil || count != 0 {
		t.Fatalf("OCR created accounting rows count=%d err=%v", count, err)
	}
}

func agentDB(t *testing.T) *sql.DB {
	url := os.Getenv("MYPOCKET_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("MYPOCKET_TEST_DATABASE_URL not set")
	}
	db, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if _, err = db.Exec(`DROP SCHEMA public CASCADE; CREATE SCHEMA public`); err != nil {
		t.Fatal(err)
	}
	if err = platformdb.Migrate(context.Background(), db, os.DirFS("../../migrations")); err != nil {
		t.Fatal(err)
	}
	return db
}

func agentUser(t *testing.T, db *sql.DB, email string) string {
	var id string
	if err := db.QueryRow(`INSERT INTO users (google_subject,email,email_verified) VALUES ($1,$2,true) RETURNING id::text`, "subject-"+email, email).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}
