package db_test

import (
	"context"
	"database/sql"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"testing/fstest"

	_ "github.com/jackc/pgx/v5/stdlib"

	"mypocket/internal/platform/db"
)

func TestMigrateSerializesConcurrentRunners(t *testing.T) {
	ctx := context.Background()
	conn := openTestPostgres(t)
	resetSchema(t, conn)

	migrations := fstest.MapFS{
		"0001_concurrent.sql": {Data: []byte("CREATE TABLE concurrent_migration_marker (id integer PRIMARY KEY); SELECT pg_sleep(0.15);")},
	}
	start := make(chan struct{})
	errs := make(chan error, 2)
	var ready sync.WaitGroup
	ready.Add(2)
	for range 2 {
		go func() {
			ready.Done()
			<-start
			errs <- db.Migrate(ctx, conn, migrations)
		}()
	}
	ready.Wait()
	close(start)

	for range 2 {
		if err := <-errs; err != nil {
			t.Fatalf("concurrent migrate must serialize: %v", err)
		}
	}

	var applied int
	if err := conn.QueryRowContext(ctx, `SELECT count(*) FROM schema_migrations WHERE version = '0001_concurrent.sql'`).Scan(&applied); err != nil {
		t.Fatalf("count migration record: %v", err)
	}
	if applied != 1 {
		t.Fatalf("expected one migration record, got %d", applied)
	}
}

func TestMigration0012UpgradesProductionShapeAndIsIdempotent(t *testing.T) {
	ctx := context.Background()
	conn := openTestPostgres(t)
	resetSchema(t, conn)

	all := loadMigrationMap(t)
	through0011 := fstest.MapFS{}
	for name, file := range all {
		if strings.HasPrefix(name, "0012_") {
			continue
		}
		through0011[name] = file
	}
	if err := db.Migrate(ctx, conn, through0011); err != nil {
		t.Fatalf("migrate through 0011: %v", err)
	}

	var userID, walletID, transactionID, assetID string
	if err := conn.QueryRowContext(ctx, `INSERT INTO users (google_subject, email, email_verified) VALUES ('migration-0012', 'migration-0012@example.com', true) RETURNING id::text`).Scan(&userID); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if err := conn.QueryRowContext(ctx, `INSERT INTO wallets (user_id, name, type, balance_vnd) VALUES ($1, 'Cash', 'cash', 90000) RETURNING id::text`, userID).Scan(&walletID); err != nil {
		t.Fatalf("seed wallet: %v", err)
	}
	if err := conn.QueryRowContext(ctx, `INSERT INTO transactions (user_id, type, source_wallet_id, category_id, amount_vnd, source_delta_vnd, balance_after_vnd, occurred_at) VALUES ($1, 'expense', $2, '00000000-0000-4000-8000-000000000201', 10000, -10000, 90000, '2026-09-11T00:00:00Z') RETURNING id::text`, userID, walletID).Scan(&transactionID); err != nil {
		t.Fatalf("seed transaction: %v", err)
	}
	if err := conn.QueryRowContext(ctx, `INSERT INTO asset_positions (user_id, type, name, unit) VALUES ($1, 'gold', 'Gold', 'chi') RETURNING id::text`, userID).Scan(&assetID); err != nil {
		t.Fatalf("seed asset: %v", err)
	}

	if err := db.Migrate(ctx, conn, all); err != nil {
		t.Fatalf("apply migration 0012: %v", err)
	}
	if err := db.Migrate(ctx, conn, all); err != nil {
		t.Fatalf("repeat migration 0012: %v", err)
	}
	if _, err := conn.ExecContext(ctx, `INSERT INTO sync_changes (user_id, cursor, entity_type, entity_id, operation, version) VALUES ($1, 1, 'asset', $2, 'add_trade', 1)`, userID, assetID); err != nil {
		t.Fatalf("0012 asset change contract not active: %v", err)
	}

	var retained int
	if err := conn.QueryRowContext(ctx, `SELECT count(*) FROM transactions WHERE id = $1 AND user_id = $2`, transactionID, userID).Scan(&retained); err != nil {
		t.Fatalf("verify retained transaction: %v", err)
	}
	if retained != 1 {
		t.Fatalf("production-shaped finance row was not retained")
	}
	var migrationRows int
	if err := conn.QueryRowContext(ctx, `SELECT count(*) FROM schema_migrations WHERE version = '0012_atomic_sync_asset_feed.sql' AND length(checksum) = 64`).Scan(&migrationRows); err != nil {
		t.Fatalf("verify migration checksum: %v", err)
	}
	if migrationRows != 1 {
		t.Fatalf("expected one checksummed 0012 record, got %d", migrationRows)
	}
}

func TestMigrateAppliesFilesOnceAndRejectsChecksumChanges(t *testing.T) {
	ctx := context.Background()
	conn := openTestPostgres(t)
	resetSchema(t, conn)

	first := fstest.MapFS{
		"0001_create_marker.sql": {Data: []byte("CREATE TABLE migration_marker (id integer PRIMARY KEY);")},
	}
	if err := db.Migrate(ctx, conn, first); err != nil {
		t.Fatalf("first migrate: %v", err)
	}
	if err := db.Migrate(ctx, conn, first); err != nil {
		t.Fatalf("second migrate should be idempotent: %v", err)
	}

	changed := fstest.MapFS{
		"0001_create_marker.sql": {Data: []byte("CREATE TABLE migration_marker_changed (id integer PRIMARY KEY);")},
	}
	if err := db.Migrate(ctx, conn, changed); err == nil {
		t.Fatal("expected checksum mismatch error")
	}
}

func TestMigrateCreatesUsersWithoutProviderTokens(t *testing.T) {
	ctx := context.Background()
	conn := openTestPostgres(t)
	resetSchema(t, conn)

	migrations := os.DirFS(filepath.Join("..", "..", "..", "migrations"))
	if err := db.Migrate(ctx, conn, migrations); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	columns := loadColumns(t, conn, "users")
	required := []string{
		"id",
		"google_subject",
		"email",
		"email_verified",
		"display_name",
		"avatar_url",
		"created_at",
		"updated_at",
	}
	for _, name := range required {
		if !columns[name] {
			t.Fatalf("users table missing %s column: %#v", name, columns)
		}
	}
	forbidden := []string{"google_access_token", "google_refresh_token", "session_id"}
	for _, name := range forbidden {
		if columns[name] {
			t.Fatalf("users table contains forbidden auth persistence column %s: %#v", name, columns)
		}
	}
}

func TestPhase002FinanceTablesAndSeeds(t *testing.T) {
	ctx := context.Background()
	conn := openTestPostgres(t)
	resetSchema(t, conn)

	migrations := os.DirFS(filepath.Join("..", "..", "..", "migrations"))
	if err := db.Migrate(ctx, conn, migrations); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	assertTableColumns(t, conn, "wallets", []string{
		"id",
		"user_id",
		"name",
		"type",
		"balance_vnd",
		"include_in_total",
		"is_default_ai",
		"archived_at",
		"version",
	})
	assertTableColumns(t, conn, "categories", []string{
		"id",
		"user_id",
		"parent_id",
		"kind",
		"name",
		"system_key",
		"is_system",
		"archived_at",
	})
	assertTableColumns(t, conn, "transactions", []string{
		"id",
		"user_id",
		"type",
		"source_wallet_id",
		"destination_wallet_id",
		"category_id",
		"amount_vnd",
		"source_delta_vnd",
		"destination_delta_vnd",
		"occurred_at",
		"excluded_from_reports",
		"archived_at",
		"version",
	})

	var count int
	err := conn.QueryRowContext(ctx, `
		SELECT count(*)
		FROM categories
		WHERE is_system AND system_key IN ('expense_food', 'income_salary', 'debt_loan')
	`).Scan(&count)
	if err != nil {
		t.Fatalf("count required system categories: %v", err)
	}
	if count != 3 {
		t.Fatalf("expected required Vietnamese system categories, got %d", count)
	}
}

func openTestPostgres(t *testing.T) *sql.DB {
	t.Helper()

	url := os.Getenv("MYPOCKET_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("MYPOCKET_TEST_DATABASE_URL is not set; PostgreSQL integration proof skipped")
	}

	conn, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if err := conn.PingContext(context.Background()); err != nil {
		t.Fatalf("ping postgres: %v", err)
	}
	return conn
}

func resetSchema(t *testing.T, conn *sql.DB) {
	t.Helper()

	if _, err := conn.ExecContext(context.Background(), `DROP SCHEMA public CASCADE; CREATE SCHEMA public;`); err != nil {
		t.Fatalf("reset schema: %v", err)
	}
}

func loadColumns(t *testing.T, conn *sql.DB, table string) map[string]bool {
	t.Helper()

	rows, err := conn.QueryContext(context.Background(), `
		SELECT column_name
		FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = $1
	`, table)
	if err != nil {
		t.Fatalf("load columns: %v", err)
	}
	defer rows.Close()

	columns := map[string]bool{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("scan column: %v", err)
		}
		columns[name] = true
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("columns rows: %v", err)
	}
	return columns
}

func assertTableColumns(t *testing.T, conn *sql.DB, table string, required []string) {
	t.Helper()

	columns := loadColumns(t, conn, table)
	for _, name := range required {
		if !columns[name] {
			t.Fatalf("%s table missing %s column: %#v", table, name, columns)
		}
	}
}

func loadMigrationMap(t *testing.T) fstest.MapFS {
	t.Helper()
	root := filepath.Join("..", "..", "..", "migrations")
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("read migrations: %v", err)
	}
	result := fstest.MapFS{}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}
		body, err := os.ReadFile(filepath.Join(root, entry.Name()))
		if err != nil {
			t.Fatalf("read migration %s: %v", entry.Name(), err)
		}
		result[entry.Name()] = &fstest.MapFile{Data: body}
	}
	return result
}

func TestMigrationsDirectoryContainsOnlySQLFiles(t *testing.T) {
	migrations := os.DirFS(filepath.Join("..", "..", "..", "migrations"))
	entries, err := fs.ReadDir(migrations, ".")
	if err != nil {
		t.Fatalf("read migrations: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			t.Fatalf("unexpected migration entry %s", entry.Name())
		}
	}
}
