package worker_test

import (
	"bytes"
	"context"
	"database/sql"
	"io"
	"os"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"mypocket/internal/finance"
	"mypocket/internal/lifecycle"
	platformdb "mypocket/internal/platform/db"
	"mypocket/internal/worker"
)

type memoryLifecycleStore struct{ objects map[string][]byte }

func (s *memoryLifecycleStore) PutObject(_ context.Context, key, _ string, body io.Reader, _ int64) error {
	data, err := io.ReadAll(body)
	if err == nil {
		s.objects[key] = data
	}
	return err
}
func (s *memoryLifecycleStore) GetObject(_ context.Context, key string, max int64) (io.ReadCloser, int64, error) {
	data := s.objects[key]
	return io.NopCloser(bytes.NewReader(data)), int64(len(data)), nil
}
func (s *memoryLifecycleStore) DeleteObject(_ context.Context, key string) error {
	delete(s.objects, key)
	return nil
}
func (s *memoryLifecycleStore) DeletePrefix(_ context.Context, prefix string) error {
	for key := range s.objects {
		if strings.HasPrefix(key, prefix) {
			delete(s.objects, key)
		}
	}
	return nil
}

func TestLifecycleRunnerPreviewsThenAppliesImportExactlyOnce(t *testing.T) {
	db := workerLifecycleDB(t)
	ctx := context.Background()
	repo := lifecycle.NewRepository(db)
	financeRepo := finance.NewRepository(db)
	user := workerLifecycleUser(t, db)
	wallet, err := financeRepo.CreateWallet(ctx, user, finance.CreateWalletInput{Name: "Cash", Type: finance.WalletBasic})
	if err != nil {
		t.Fatal(err)
	}
	var category string
	if err = db.QueryRow(`SELECT id::text FROM categories WHERE system_key='expense_food'`).Scan(&category); err != nil {
		t.Fatal(err)
	}
	key := "users/" + user + "/imports/input.csv"
	csv := "occurred_at,type,amount_vnd,source_wallet_id,destination_wallet_id,category_id,note,excluded_from_reports\n2026-09-11T10:00:00+07:00,expense,12000," + wallet.ID + ",," + category + ",Cafe,false\n"
	store := &memoryLifecycleStore{objects: map[string][]byte{key: []byte(csv)}}
	job, err := repo.CreateJob(ctx, user, lifecycle.KindImport, "once", lifecycle.ImportRequest{ObjectKey: key})
	if err != nil {
		t.Fatal(err)
	}
	runner := worker.LifecycleRunner{Repo: repo, Store: store}
	if n, err := runner.RunOnce(ctx); err != nil || n != 1 {
		t.Fatalf("preview run = %d %v", n, err)
	}
	job, err = repo.GetJob(ctx, user, job.ID)
	if err != nil || job.Status != lifecycle.StatusAwaitingConfirmation {
		t.Fatalf("preview job: %+v %v", job, err)
	}
	if _, err = repo.ConfirmImport(ctx, user, job.ID, job.Version); err != nil {
		t.Fatal(err)
	}
	if n, err := runner.RunOnce(ctx); err != nil || n != 1 {
		t.Fatalf("apply run = %d %v", n, err)
	}
	if n, err := runner.RunOnce(ctx); err != nil || n != 0 {
		t.Fatalf("replay run = %d %v", n, err)
	}
	var count int
	var balance int64
	if err = db.QueryRow(`SELECT count(*) FROM transactions WHERE user_id=$1`, user).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if err = db.QueryRow(`SELECT balance_vnd FROM wallets WHERE id=$1`, wallet.ID).Scan(&balance); err != nil {
		t.Fatal(err)
	}
	if count != 1 || balance != -12000 {
		t.Fatalf("count=%d balance=%d", count, balance)
	}
}

func TestLifecycleRunnerCreatesPrivateDeterministicExport(t *testing.T) {
	db := workerLifecycleDB(t)
	ctx := context.Background()
	repo := lifecycle.NewRepository(db)
	user := workerLifecycleUser(t, db)
	store := &memoryLifecycleStore{objects: map[string][]byte{}}
	job, err := repo.CreateJob(ctx, user, lifecycle.KindExport, "export", lifecycle.ExportRequest{Datasets: []string{"wallets"}})
	if err != nil {
		t.Fatal(err)
	}
	runner := worker.LifecycleRunner{Repo: repo, Store: store}
	if _, err = runner.RunOnce(ctx); err != nil {
		t.Fatal(err)
	}
	job, err = repo.GetJob(ctx, user, job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if job.Status != lifecycle.StatusCompleted || job.ResultObjectKey != "users/"+user+"/exports/"+job.ID+".csv" || len(store.objects[job.ResultObjectKey]) == 0 {
		t.Fatalf("bad export: %+v", job)
	}
}

func workerLifecycleDB(t *testing.T) *sql.DB {
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
func workerLifecycleUser(t *testing.T, db *sql.DB) string {
	t.Helper()
	var id string
	if err := db.QueryRow(`INSERT INTO users(google_subject,email,email_verified)VALUES('worker-life','worker-life@example.com',true)RETURNING id::text`).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}
