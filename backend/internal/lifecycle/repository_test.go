package lifecycle

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"mypocket/internal/identity"
	platformdb "mypocket/internal/platform/db"
)

func TestRepositoryIsolatesJobsAndReplaysIdempotently(t *testing.T) {
	db := lifecycleDB(t)
	a := lifecycleUser(t, db, "life-a@example.com")
	b := lifecycleUser(t, db, "life-b@example.com")
	repo := NewRepository(db)
	first, err := repo.CreateJob(context.Background(), a, KindExport, "same", ExportRequest{Datasets: []string{"transactions"}})
	if err != nil {
		t.Fatal(err)
	}
	replay, err := repo.CreateJob(context.Background(), a, KindExport, "same", ExportRequest{Datasets: []string{"wallets"}})
	if err != nil {
		t.Fatal(err)
	}
	if replay.ID != first.ID {
		t.Fatalf("idempotent replay created %s after %s", replay.ID, first.ID)
	}
	claimed, err := repo.ClaimDue(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err = repo.Retry(context.Background(), claimed, "TEMPORARY"); err != nil {
		t.Fatal(err)
	}
	retried, err := repo.GetJob(context.Background(), a, first.ID)
	if err != nil {
		t.Fatal(err)
	}
	if retried.Status != StatusQueued || retried.ErrorCode != "TEMPORARY" {
		t.Fatalf("job was not resumable: %+v", retried)
	}
	if _, err = repo.GetJob(context.Background(), b, first.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("cross-user job leak: %v", err)
	}
}

func TestRepositoryConfirmationIsVersionedAndResetPreservesIdentity(t *testing.T) {
	db := lifecycleDB(t)
	user := lifecycleUser(t, db, "life-reset@example.com")
	repo := NewRepository(db)
	if _, err := db.Exec(`INSERT INTO wallets(user_id,name,type) VALUES($1,'Cash','cash')`, user); err != nil {
		t.Fatal(err)
	}
	job, err := repo.CreateJob(context.Background(), user, KindImport, "import", ImportRequest{ObjectKey: "users/" + user + "/imports/a.csv"})
	if err != nil {
		t.Fatal(err)
	}
	claimed, err := repo.ClaimDue(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	preview := ImportPreview{ID: job.ID, Version: claimed.Version + 1, Confirmable: true, ValidRows: []ImportRow{}}
	if err = repo.AwaitConfirmation(context.Background(), claimed, preview); err != nil {
		t.Fatal(err)
	}
	awaiting, err := repo.GetJob(context.Background(), user, job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = repo.ConfirmImport(context.Background(), user, job.ID, awaiting.Version-1); !errors.Is(err, ErrConflict) {
		t.Fatalf("stale confirmation = %v", err)
	}
	if _, err = repo.ConfirmImport(context.Background(), user, job.ID, awaiting.Version); err != nil {
		t.Fatal(err)
	}
	if err = repo.ResetUserData(context.Background(), user); err != nil {
		t.Fatal(err)
	}
	var users, wallets int
	if err = db.QueryRow(`SELECT count(*) FROM users WHERE id=$1`, user).Scan(&users); err != nil {
		t.Fatal(err)
	}
	if err = db.QueryRow(`SELECT count(*) FROM wallets WHERE user_id=$1`, user).Scan(&wallets); err != nil {
		t.Fatal(err)
	}
	if users != 1 || wallets != 0 {
		t.Fatalf("reset users=%d wallets=%d", users, wallets)
	}
	_ = json.Valid(job.Request)
}

func TestDisableUserImmediatelyRejectsSessionAndAPIKeyLookups(t *testing.T) {
	db := lifecycleDB(t)
	user := lifecycleUser(t, db, "life-disabled@example.com")
	repo := NewRepository(db)
	identityRepo := identity.NewRepository(db)
	secret := "test-api-key-hash-secret-at-least-32-bytes"
	key, err := identityRepo.CreateAPIKey(context.Background(), user, "integration", secret)
	if err != nil {
		t.Fatal(err)
	}
	if err = repo.DisableUser(context.Background(), user); err != nil {
		t.Fatal(err)
	}
	if _, err = identityRepo.FindByID(context.Background(), user); !errors.Is(err, identity.ErrUserNotFound) {
		t.Fatalf("disabled session lookup=%v", err)
	}
	if _, _, err = identityRepo.AuthenticateAPIKey(context.Background(), key.Plaintext, secret); !errors.Is(err, identity.ErrUserNotFound) {
		t.Fatalf("disabled api key lookup=%v", err)
	}
}

func lifecycleDB(t *testing.T) *sql.DB {
	t.Helper()
	url := os.Getenv("MYPOCKET_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("MYPOCKET_TEST_DATABASE_URL is not set")
	}
	db, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if _, err = db.Exec(`DROP SCHEMA public CASCADE;CREATE SCHEMA public`); err != nil {
		t.Fatal(err)
	}
	if err = platformdb.Migrate(context.Background(), db, os.DirFS("../../migrations")); err != nil {
		t.Fatal(err)
	}
	return db
}
func lifecycleUser(t *testing.T, db *sql.DB, email string) string {
	t.Helper()
	var id string
	if err := db.QueryRow(`INSERT INTO users(google_subject,email,email_verified)VALUES($1,$2,true)RETURNING id::text`, "sub-"+email, email).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}
