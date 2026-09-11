package agent_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
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
	wallet, err := financeRepo.CreateWallet(ctx, user, finance.CreateWalletInput{Name: "Cash", Type: finance.WalletCash})
	if err != nil {
		t.Fatal(err)
	}
	foreign, err := financeRepo.CreateWallet(ctx, other, finance.CreateWalletInput{Name: "Foreign", Type: finance.WalletCash})
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
