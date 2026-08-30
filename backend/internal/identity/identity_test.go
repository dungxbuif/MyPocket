package identity_test

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"mypocket/internal/identity"
	"mypocket/internal/platform/db"
)

func TestSignedCookieRoundTrip(t *testing.T) {
	signer := identity.NewCookieSigner([]byte("01234567890123456789012345678901"))
	value, err := signer.Sign("user_123", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	claims, err := signer.Verify(value)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != "user_123" {
		t.Fatalf("wrong user id: %s", claims.UserID)
	}
}

func TestSignedCookieRejectsTampering(t *testing.T) {
	signer := identity.NewCookieSigner([]byte("01234567890123456789012345678901"))
	value, err := signer.Sign("user_123", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	_, err = signer.Verify(value + "tampered")
	if err == nil {
		t.Fatal("expected tampered cookie to be rejected")
	}
}

func TestCSRFMissingHeaderRejected(t *testing.T) {
	handler := identity.RequireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil))

	if res.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", res.Code)
	}
	if !strings.Contains(res.Body.String(), "CSRF_REQUIRED") {
		t.Fatalf("expected stable CSRF error code, got %s", res.Body.String())
	}
}

func TestOAuthFixtureProvisionsUserWithoutProviderTokens(t *testing.T) {
	conn := migratedTestPostgres(t)
	repo := identity.NewRepository(conn)
	profile := identity.GoogleProfile{
		Subject:       "google-sub-1",
		Email:         "a@example.com",
		EmailVerified: true,
		DisplayName:   "A",
		AvatarURL:     "https://example.com/a.png",
	}

	user, err := repo.FindOrCreateGoogleUser(context.Background(), profile)
	if err != nil {
		t.Fatal(err)
	}
	if user.Email != "a@example.com" {
		t.Fatalf("wrong email: %s", user.Email)
	}
	if user.GoogleSubject != "google-sub-1" {
		t.Fatalf("wrong google subject: %s", user.GoogleSubject)
	}
	assertNoProviderTokenColumns(t, conn)
}

func TestRequireOwnerRejectsOtherUsersObject(t *testing.T) {
	conn := migratedTestPostgres(t)
	repo := identity.NewRepository(conn)
	userA, err := repo.FindOrCreateGoogleUser(context.Background(), identity.GoogleProfile{
		Subject:       "google-sub-a",
		Email:         "a@example.com",
		EmailVerified: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	userB, err := repo.FindOrCreateGoogleUser(context.Background(), identity.GoogleProfile{
		Subject:       "google-sub-b",
		Email:         "b@example.com",
		EmailVerified: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	objectID := createOwnedHarnessObject(t, conn, userA.ID)

	err = identity.RequireOwner(context.Background(), conn, userB.ID, objectID)

	if !errors.Is(err, identity.ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func migratedTestPostgres(t *testing.T) *sql.DB {
	t.Helper()

	url := os.Getenv("MYPOCKET_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("MYPOCKET_TEST_DATABASE_URL is not set; identity repository integration proof skipped")
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

func assertNoProviderTokenColumns(t *testing.T, conn *sql.DB) {
	t.Helper()

	rows, err := conn.QueryContext(context.Background(), `
		SELECT column_name
		FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = 'users'
	`)
	if err != nil {
		t.Fatalf("load columns: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("scan column: %v", err)
		}
		if name == "google_access_token" || name == "google_refresh_token" || name == "session_id" {
			t.Fatalf("users table contains forbidden auth persistence column %s", name)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("columns rows: %v", err)
	}
}

func createOwnedHarnessObject(t *testing.T, conn *sql.DB, userID string) string {
	t.Helper()

	if _, err := conn.ExecContext(context.Background(), `
		CREATE TABLE ownership_harness_objects (
			id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id uuid NOT NULL REFERENCES users(id)
		)
	`); err != nil {
		t.Fatalf("create ownership harness: %v", err)
	}

	var objectID string
	if err := conn.QueryRowContext(context.Background(), `
		INSERT INTO ownership_harness_objects (user_id)
		VALUES ($1)
		RETURNING id::text
	`, userID).Scan(&objectID); err != nil {
		t.Fatalf("insert ownership harness object: %v", err)
	}
	return objectID
}
