package db

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
)

func Migrate(ctx context.Context, conn *sql.DB, migrations fs.FS) error {
	locked, err := conn.Conn(ctx)
	if err != nil {
		return fmt.Errorf("reserve migration connection: %w", err)
	}
	defer locked.Close()
	if _, err := locked.ExecContext(ctx, `SELECT pg_advisory_lock(7424374638811478388)`); err != nil {
		return fmt.Errorf("lock migrations: %w", err)
	}
	defer func() {
		// Closing the dedicated connection also releases the session lock. The
		// explicit unlock keeps a healthy pooled connection immediately reusable.
		_, _ = locked.ExecContext(context.Background(), `SELECT pg_advisory_unlock(7424374638811478388)`)
	}()

	entries, err := fs.ReadDir(migrations, ".")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}
	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			return fmt.Errorf("invalid migration entry %s: directories are not supported", entry.Name())
		}
		if filepath.Ext(entry.Name()) == ".sql" {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)

	if err := ensureMigrationTable(ctx, locked); err != nil {
		return err
	}
	for _, name := range files {
		body, err := fs.ReadFile(migrations, name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		sum := checksum(body)
		if err := applyMigration(ctx, locked, name, sum, string(body)); err != nil {
			return err
		}
	}
	return nil
}

type migrationExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

type migrationBeginner interface {
	BeginTx(context.Context, *sql.TxOptions) (*sql.Tx, error)
}

func ensureMigrationTable(ctx context.Context, conn migrationExecutor) error {
	_, err := conn.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version text PRIMARY KEY,
			checksum text NOT NULL,
			applied_at timestamptz NOT NULL DEFAULT now()
		)
	`)
	if err != nil {
		return fmt.Errorf("ensure schema_migrations: %w", err)
	}
	return nil
}

func applyMigration(ctx context.Context, conn migrationBeginner, version string, checksum string, sqlText string) error {
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", version, err)
	}
	defer func() { _ = tx.Rollback() }()

	var existing string
	err = tx.QueryRowContext(ctx, `SELECT checksum FROM schema_migrations WHERE version = $1`, version).Scan(&existing)
	switch {
	case err == nil:
		if existing != checksum {
			return fmt.Errorf("migration %s checksum mismatch", version)
		}
		return tx.Commit()
	case err != sql.ErrNoRows:
		return fmt.Errorf("check migration %s: %w", version, err)
	}

	if _, err := tx.ExecContext(ctx, sqlText); err != nil {
		return fmt.Errorf("apply migration %s: %w", version, err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (version, checksum) VALUES ($1, $2)`, version, checksum); err != nil {
		return fmt.Errorf("record migration %s: %w", version, err)
	}
	return tx.Commit()
}

func checksum(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}
