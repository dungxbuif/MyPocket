package sync_test

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"mypocket/internal/finance"
	"mypocket/internal/platform/db"
	mysync "mypocket/internal/sync"
)

func TestRepositoryIntegrationAppliesIdempotentMutationAndListsUserChanges(t *testing.T) {
	conn := migratedSyncPostgres(t)
	financeRepo := finance.NewRepository(conn)
	store := mysync.NewRepository(conn)
	service := mysync.NewService(store, financeRepo)
	userA := createSyncUser(t, conn, "sync-a@example.com")
	userB := createSyncUser(t, conn, "sync-b@example.com")
	wallet := createSyncWallet(t, financeRepo, userA)
	categoryID := findSyncSystemCategory(t, conn, "expense_food")
	mutation := mysync.Mutation{
		MutationID:  "00000000-0000-4000-8000-000000000901",
		DeviceID:    "device_a",
		Sequence:    1,
		EntityType:  mysync.EntityTransaction,
		EntityID:    "00000000-0000-4000-8000-000000000301",
		Operation:   mysync.OperationCreate,
		BaseVersion: 0,
		Payload:     rawJSON(`{"type":"expense","source_wallet_id":"` + wallet.ID + `","category_id":"` + categoryID + `","amount_vnd":42000,"occurred_at":"2026-08-31T00:00:00Z","note":"Sync cafe"}`),
	}

	first, err := service.ApplyMutations(context.Background(), userA, []mysync.Mutation{mutation})
	if err != nil {
		t.Fatalf("apply first: %v", err)
	}
	second, err := service.ApplyMutations(context.Background(), userA, []mysync.Mutation{mutation})
	if err != nil {
		t.Fatalf("apply replay: %v", err)
	}
	changesA, err := service.Changes(context.Background(), userA, 0, 100)
	if err != nil {
		t.Fatalf("list changes A: %v", err)
	}
	changesB, err := service.Changes(context.Background(), userB, 0, 100)
	if err != nil {
		t.Fatalf("list changes B: %v", err)
	}

	if first[0].State != mysync.ResultApplied || second[0].State != mysync.ResultReplayed {
		t.Fatalf("unexpected mutation states first=%s second=%s", first[0].State, second[0].State)
	}
	if len(changesA.Changes) != 3 || changesA.Changes[0].EntityID != wallet.ID || changesA.Changes[1].EntityID != wallet.ID || changesA.Changes[1].Version != 2 || changesA.Changes[2].EntityID != mutation.EntityID || changesA.NextCursor != 3 {
		t.Fatalf("unexpected user A changes: %#v", changesA)
	}
	if len(changesB.Changes) != 0 {
		t.Fatalf("change feed leaked to user B: %#v", changesB)
	}
	assertSyncWalletBalance(t, conn, wallet.ID, -42000)
}

func TestRepositoryIntegrationReturnsConflictForStaleTransactionVersion(t *testing.T) {
	conn := migratedSyncPostgres(t)
	financeRepo := finance.NewRepository(conn)
	service := mysync.NewService(mysync.NewRepository(conn), financeRepo)
	userID := createSyncUser(t, conn, "sync-conflict@example.com")
	wallet := createSyncWallet(t, financeRepo, userID)
	categoryID := findSyncSystemCategory(t, conn, "expense_food")
	created, err := financeRepo.CreateTransaction(context.Background(), userID, finance.CreateTransactionInput{
		IdempotencyKey: "direct-create",
		Type:           finance.TransactionExpense,
		SourceWalletID: wallet.ID,
		CategoryID:     categoryID,
		AmountVND:      10000,
		OccurredAt:     fixedTime(),
	})
	if err != nil {
		t.Fatalf("create transaction: %v", err)
	}
	if _, err := financeRepo.UpdateTransaction(context.Background(), userID, created.ID, finance.UpdateTransactionInput{
		BaseVersion:    created.Version,
		Type:           finance.TransactionExpense,
		SourceWalletID: wallet.ID,
		CategoryID:     categoryID,
		AmountVND:      12000,
		OccurredAt:     fixedTime(),
		Note:           "server update",
	}); err != nil {
		t.Fatalf("direct update: %v", err)
	}

	result, err := service.ApplyMutations(context.Background(), userID, []mysync.Mutation{{
		MutationID:  "00000000-0000-4000-8000-000000000902",
		DeviceID:    "device_a",
		Sequence:    1,
		EntityType:  mysync.EntityTransaction,
		EntityID:    created.ID,
		Operation:   mysync.OperationUpdate,
		BaseVersion: created.Version,
		Payload:     rawJSON(`{"type":"expense","source_wallet_id":"` + wallet.ID + `","category_id":"` + categoryID + `","amount_vnd":13000,"occurred_at":"2026-08-31T00:00:00Z","note":"local stale"}`),
	}})
	if err != nil {
		t.Fatalf("apply stale update: %v", err)
	}

	if result[0].State != mysync.ResultConflict || result[0].Conflict == nil {
		t.Fatalf("expected conflict, got %#v", result[0])
	}
	if result[0].Conflict.ServerVersion <= result[0].Conflict.BaseVersion {
		t.Fatalf("expected newer server version, got %#v", result[0].Conflict)
	}
}

func TestRepositoryIntegrationResyncReturnsBoundedAuthoritativeSnapshot(t *testing.T) {
	conn := migratedSyncPostgres(t)
	financeRepo := finance.NewRepository(conn)
	service := mysync.NewService(mysync.NewRepository(conn), financeRepo)
	userID := createSyncUser(t, conn, "sync-resync@example.com")
	wallet := createSyncWallet(t, financeRepo, userID)

	snapshot, err := service.Resync(context.Background(), userID)
	if err != nil {
		t.Fatalf("resync: %v", err)
	}

	if len(snapshot.Wallets) != 1 || snapshot.Wallets[0].ID != wallet.ID {
		t.Fatalf("unexpected snapshot wallets: %#v", snapshot.Wallets)
	}
	if len(snapshot.Categories) == 0 {
		t.Fatalf("expected system categories in snapshot")
	}
	if snapshot.ServerEpoch != mysync.ServerEpoch {
		t.Fatalf("expected server epoch %q, got %q", mysync.ServerEpoch, snapshot.ServerEpoch)
	}
}

func migratedSyncPostgres(t *testing.T) *sql.DB {
	t.Helper()

	url := os.Getenv("MYPOCKET_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("MYPOCKET_TEST_DATABASE_URL is not set; sync repository integration proof skipped")
	}
	conn, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if _, err := conn.ExecContext(context.Background(), `DROP SCHEMA public CASCADE; CREATE SCHEMA public;`); err != nil {
		t.Fatalf("reset schema: %v", err)
	}
	if err := db.Migrate(context.Background(), conn, os.DirFS("../../migrations")); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return conn
}

func createSyncUser(t *testing.T, conn *sql.DB, email string) string {
	t.Helper()

	var userID string
	err := conn.QueryRowContext(context.Background(), `
		INSERT INTO users (google_subject, email, email_verified)
		VALUES ($1, $2, true)
		RETURNING id::text
	`, "subject-"+email, email).Scan(&userID)
	if err != nil {
		t.Fatalf("create sync user: %v", err)
	}
	return userID
}

func createSyncWallet(t *testing.T, repo *finance.Repository, userID string) finance.Wallet {
	t.Helper()

	wallet, err := repo.CreateWallet(context.Background(), userID, finance.CreateWalletInput{Name: "Tiền mặt", Type: finance.WalletCash})
	if err != nil {
		t.Fatalf("create sync wallet: %v", err)
	}
	return wallet
}

func findSyncSystemCategory(t *testing.T, conn *sql.DB, systemKey string) string {
	t.Helper()

	var categoryID string
	err := conn.QueryRowContext(context.Background(), `
		SELECT id::text
		FROM categories
		WHERE system_key = $1
	`, systemKey).Scan(&categoryID)
	if err != nil {
		t.Fatalf("find system category %s: %v", systemKey, err)
	}
	return categoryID
}

func assertSyncWalletBalance(t *testing.T, conn *sql.DB, walletID string, want int64) {
	t.Helper()

	var balance int64
	if err := conn.QueryRowContext(context.Background(), `SELECT balance_vnd FROM wallets WHERE id = $1`, walletID).Scan(&balance); err != nil {
		t.Fatalf("load wallet balance: %v", err)
	}
	if balance != want {
		t.Fatalf("expected balance %d, got %d", want, balance)
	}
}
