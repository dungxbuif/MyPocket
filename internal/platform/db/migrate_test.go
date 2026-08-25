package db_test

import (
	"context"
	"database/sql"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	_ "github.com/jackc/pgx/v5/stdlib"

	"mypocket/internal/platform/db"
)

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
