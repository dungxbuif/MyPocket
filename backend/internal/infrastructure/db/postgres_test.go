package db

import (
	"bytes"
	"context"
	"errors"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func capturedPostgres(t *testing.T) (*gorm.DB, *bytes.Buffer) {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires disposable PostgreSQL TEST_DATABASE_URL")
	}
	var output bytes.Buffer
	previous := logger.Default
	logger.Default = logger.New(log.New(&output, "", 0), logger.Config{LogLevel: logger.Info, Colorful: false})
	t.Cleanup(func() { logger.Default = previous })
	database, err := NewPostgres(dsn)
	if err != nil {
		t.Fatal("cannot connect to disposable PostgreSQL")
	}
	sqlDB, err := database.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			t.Error(err)
		}
	})
	return database, &output
}

func TestPostgresSQLLogOmitsBoundFinancialValuesButRetainsQueryMetadata(t *testing.T) {
	database, output := capturedPostgres(t)
	const privateNote = "synthetic-private-financial-note-DO-NOT-LOG"
	const amount = int64(987654321234)
	var result struct {
		Note   string
		Amount int64
	}
	if err := database.Raw("SELECT ?::text AS note, ?::bigint AS amount", privateNote, amount).Scan(&result).Error; err != nil {
		t.Fatal(err)
	}
	if result.Note != privateNote || result.Amount != amount {
		t.Fatal("privacy logging changed query result")
	}
	logs := output.String()
	if strings.Contains(logs, privateNote) || strings.Contains(logs, "987654321234") {
		t.Fatal("SQL log exposed bound financial values")
	}
	if !strings.Contains(logs, "SELECT") || !strings.Contains(logs, "rows:1") {
		t.Fatalf("lost query metadata: %q", logs)
	}
}

func TestPostgresSQLLogDoesNotEchoDriverErrorValueAfterLogModeChange(t *testing.T) {
	database, output := capturedPostgres(t)
	const secret = "synthetic-secret-DO-NOT-LOG"
	var result int
	err := database.Session(&gorm.Session{Logger: database.Logger.LogMode(logger.Info)}).
		Raw("SELECT ?::integer", secret).Scan(&result).Error
	if err == nil {
		t.Fatal("expected PostgreSQL cast rejection")
	}
	// The caller still gets the actual error; only diagnostics are sanitized.
	if !strings.Contains(err.Error(), secret) {
		t.Fatal("application error unexpectedly changed")
	}
	logs := output.String()
	if strings.Contains(logs, secret) {
		t.Fatal("SQL diagnostic echoed private driver error value")
	}
	if !strings.Contains(logs, "SQLSTATE 22P02") || !strings.Contains(logs, "SELECT") {
		t.Fatalf("lost safe error class/query: %q", logs)
	}
}

func TestPostgresSQLLogRedactsNonScanQueriesAndGenericEvents(t *testing.T) {
	database, output := capturedPostgres(t)
	const secret = "synthetic-event-note-DO-NOT-LOG"
	var returned string
	if err := database.Raw("SELECT ?::text", secret).Row().Scan(&returned); err != nil {
		t.Fatal(err)
	}
	if returned != secret {
		t.Fatal("redaction changed Row query result")
	}
	ctx := context.Background()
	database.Logger.Info(ctx, "private info %s", secret)
	database.Logger.Warn(ctx, "private warning %s", secret)
	database.Logger.Error(ctx, "private error %s", secret)
	database.Logger.Trace(ctx, time.Now(), func() (string, int64) {
		return "UNTRUSTED_" + secret, 0
	}, errors.New(secret))
	logs := output.String()
	if strings.Contains(logs, secret) || strings.Contains(logs, "UNTRUSTED_") {
		t.Fatal("database logging path exposed private arguments/query/error")
	}
	if !strings.Contains(logs, "SELECT (statement redacted)") || !strings.Contains(logs, "database query failed") {
		t.Fatalf("lost safe diagnostics: %q", logs)
	}
}

func TestPostgresSQLLogPreservesRecordNotFoundWithoutPrivateLookupValue(t *testing.T) {
	database, output := capturedPostgres(t)
	const privateID = "synthetic-private-lookup-DO-NOT-LOG"
	var row struct{ ID string }
	err := database.Table("user").Where("id = ?", privateID).Take(&row).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("record-not-found contract changed: %v", err)
	}
	logs := output.String()
	if strings.Contains(logs, privateID) {
		t.Fatal("SQL log exposed lookup identity")
	}
	if !strings.Contains(logs, "record not found") || !strings.Contains(logs, "rows:0") {
		t.Fatalf("lost safe not-found metadata: %q", logs)
	}
}
