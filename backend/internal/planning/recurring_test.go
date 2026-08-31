package planning_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"mypocket/internal/finance"
	"mypocket/internal/planning"
)

func TestRecurringScheduleCreatesDeterministicDraftWithoutAccounting(t *testing.T) {
	conn := migratedPlanningPostgres(t)
	owner := createPlanningUser(t, conn, "recurring-owner@example.com")
	financeRepo := finance.NewRepository(conn)
	planningRepo := planning.NewRepository(conn)
	wallet := createPlanningWallet(t, financeRepo, owner)
	categoryID := findPlanningSystemCategory(t, conn, "expense_food")

	schedule, err := planningRepo.CreateRecurringSchedule(context.Background(), owner, planning.CreateRecurringScheduleInput{
		Name:           "Tiền nhà",
		Frequency:      planning.RecurrenceMonthly,
		Timezone:       "Asia/Ho_Chi_Minh",
		StartsAt:       "2026-08-01T09:00:00+07:00",
		Type:           finance.TransactionExpense,
		SourceWalletID: wallet.ID,
		CategoryID:     categoryID,
		AmountVND:      3500000,
		Note:           "Thuê nhà",
	})
	if err != nil {
		t.Fatalf("create schedule: %v", err)
	}
	before := walletBalance(t, conn, wallet.ID)
	processed, err := planningRepo.ProcessDueRecurringSchedules(context.Background(), "worker-a", time.Date(2026, 9, 2, 2, 0, 0, 0, time.UTC), 10)
	if err != nil {
		t.Fatalf("process schedules: %v", err)
	}
	if processed != 2 {
		t.Fatalf("expected two due occurrences, got %d", processed)
	}
	again, err := planningRepo.ProcessDueRecurringSchedules(context.Background(), "worker-b", time.Date(2026, 9, 2, 2, 0, 0, 0, time.UTC), 10)
	if err != nil {
		t.Fatalf("process schedules again: %v", err)
	}
	if again != 0 {
		t.Fatalf("deterministic occurrences should not duplicate, got %d", again)
	}
	drafts, err := planningRepo.ListTransactionDrafts(context.Background(), owner)
	if err != nil {
		t.Fatalf("list drafts: %v", err)
	}
	if len(drafts) != 2 || drafts[0].ScheduleID != schedule.ID || drafts[0].AmountVND != 3500000 || drafts[0].Note != "Thuê nhà" {
		t.Fatalf("unexpected drafts: %#v", drafts)
	}
	after := walletBalance(t, conn, wallet.ID)
	if after != before {
		t.Fatalf("recurring draft changed wallet balance: before=%d after=%d", before, after)
	}
}

func TestWorkerLeaseAllowsSingleActiveOwner(t *testing.T) {
	conn := migratedPlanningPostgres(t)
	planningRepo := planning.NewRepository(conn)
	now := time.Date(2026, 8, 31, 2, 0, 0, 0, time.UTC)

	acquired, err := planningRepo.AcquireWorkerLease(context.Background(), "recurring", "worker-a", time.Minute, now)
	if err != nil {
		t.Fatalf("acquire lease: %v", err)
	}
	if !acquired {
		t.Fatal("first worker should acquire lease")
	}
	acquired, err = planningRepo.AcquireWorkerLease(context.Background(), "recurring", "worker-b", time.Minute, now.Add(30*time.Second))
	if err != nil {
		t.Fatalf("acquire competing lease: %v", err)
	}
	if acquired {
		t.Fatal("second worker should not acquire active lease")
	}
	acquired, err = planningRepo.AcquireWorkerLease(context.Background(), "recurring", "worker-b", time.Minute, now.Add(2*time.Minute))
	if err != nil {
		t.Fatalf("acquire expired lease: %v", err)
	}
	if !acquired {
		t.Fatal("second worker should acquire expired lease")
	}
}

func walletBalance(t *testing.T, conn *sql.DB, walletID string) int64 {
	t.Helper()
	var balance int64
	if err := conn.QueryRowContext(context.Background(), `SELECT balance_vnd FROM wallets WHERE id = $1`, walletID).Scan(&balance); err != nil {
		t.Fatalf("wallet balance: %v", err)
	}
	return balance
}
