package planning_test

import (
	"context"
	"database/sql"
	"errors"
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

func TestRecurringScheduleRejectsFinanceFieldsThatWouldCreateUnconfirmableDrafts(t *testing.T) {
	conn := migratedPlanningPostgres(t)
	owner := createPlanningUser(t, conn, "recurring-validation@example.com")
	financeRepo := finance.NewRepository(conn)
	planningRepo := planning.NewRepository(conn)
	source := createPlanningWallet(t, financeRepo, owner)
	destination, err := financeRepo.CreateWallet(context.Background(), owner, finance.CreateWalletInput{Name: "Ngân hàng", Type: finance.WalletBasic})
	if err != nil {
		t.Fatalf("create destination wallet: %v", err)
	}
	expenseCategoryID := findPlanningSystemCategory(t, conn, "expense_food")
	incomeCategoryID := findPlanningSystemCategory(t, conn, "income_salary")
	base := planning.CreateRecurringScheduleInput{
		Name: "Lịch kiểm chứng", Frequency: planning.RecurrenceMonthly, Timezone: "Asia/Ho_Chi_Minh",
		StartsAt: "2026-09-01T09:00:00+07:00", SourceWalletID: source.ID, AmountVND: 100_000,
	}
	tests := []struct {
		name  string
		input planning.CreateRecurringScheduleInput
	}{
		{name: "transfer with category", input: mergeRecurringInput(base, finance.TransactionTransfer, destination.ID, expenseCategoryID)},
		{name: "transfer to the same wallet", input: mergeRecurringInput(base, finance.TransactionTransfer, source.ID, "")},
		{name: "expense with destination wallet", input: mergeRecurringInput(base, finance.TransactionExpense, destination.ID, expenseCategoryID)},
		{name: "income with expense category", input: mergeRecurringInput(base, finance.TransactionIncome, "", expenseCategoryID)},
		{name: "expense with income category", input: mergeRecurringInput(base, finance.TransactionExpense, "", incomeCategoryID)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := planningRepo.CreateRecurringSchedule(context.Background(), owner, tt.input); !errors.Is(err, planning.ErrValidation) {
				t.Fatalf("create schedule error = %v, want validation", err)
			}
		})
	}
}

func TestRecurringScheduleCanBeUpdatedPausedResumedAndEnded(t *testing.T) {
	conn := migratedPlanningPostgres(t)
	owner := createPlanningUser(t, conn, "recurring-controls@example.com")
	financeRepo := finance.NewRepository(conn)
	planningRepo := planning.NewRepository(conn)
	wallet := createPlanningWallet(t, financeRepo, owner)
	categoryID := findPlanningSystemCategory(t, conn, "expense_food")

	schedule, err := planningRepo.CreateRecurringSchedule(context.Background(), owner, planning.CreateRecurringScheduleInput{
		Name: "Ăn sáng", Frequency: planning.RecurrenceDaily, Timezone: "Asia/Ho_Chi_Minh",
		StartsAt: "2026-09-01T07:00:00+07:00", Type: finance.TransactionExpense,
		SourceWalletID: wallet.ID, CategoryID: categoryID, AmountVND: 50000, Note: "Cũ",
	})
	if err != nil {
		t.Fatalf("create recurring schedule: %v", err)
	}
	updated, err := planningRepo.UpdateRecurringSchedule(context.Background(), owner, schedule.ID, planning.UpdateRecurringScheduleInput{
		BaseVersion: schedule.Version,
		Name:        "Ăn sáng updated", Frequency: planning.RecurrenceDaily, Timezone: "Asia/Ho_Chi_Minh",
		StartsAt: "2026-09-01T07:00:00+07:00", EndsAt: "2026-09-01T07:00:00+07:00",
		Type: finance.TransactionExpense, SourceWalletID: wallet.ID, CategoryID: categoryID, AmountVND: 75000, Note: "Mới",
	})
	if err != nil {
		t.Fatalf("update recurring schedule: %v", err)
	}
	if updated.AmountVND != 75000 || updated.Note != "Mới" || updated.EndsAt == nil || updated.Version != 2 {
		t.Fatalf("unexpected updated schedule: %#v", updated)
	}
	paused, err := planningRepo.PauseRecurringSchedule(context.Background(), owner, schedule.ID, updated.Version)
	if err != nil {
		t.Fatalf("pause recurring schedule: %v", err)
	}
	if paused.PausedAt == nil || paused.Version != 3 {
		t.Fatalf("paused schedule should expose paused_at/version: %#v", paused)
	}
	processed, err := planningRepo.ProcessDueRecurringSchedules(context.Background(), "worker-paused", time.Date(2026, 9, 2, 2, 0, 0, 0, time.UTC), 10)
	if err != nil {
		t.Fatalf("process paused schedule: %v", err)
	}
	if processed != 0 {
		t.Fatalf("paused schedule must not create occurrences, got %d", processed)
	}
	resumed, err := planningRepo.ResumeRecurringSchedule(context.Background(), owner, schedule.ID, paused.Version, time.Date(2026, 8, 31, 2, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("resume recurring schedule: %v", err)
	}
	if resumed.PausedAt != nil || resumed.Version != 4 {
		t.Fatalf("resumed schedule should clear paused_at/version: %#v", resumed)
	}
	processed, err = planningRepo.ProcessDueRecurringSchedules(context.Background(), "worker-ended", time.Date(2026, 9, 3, 2, 0, 0, 0, time.UTC), 10)
	if err != nil {
		t.Fatalf("process ended schedule: %v", err)
	}
	if processed != 1 {
		t.Fatalf("ends_at should allow first occurrence only, got %d", processed)
	}
	drafts, err := planningRepo.ListTransactionDrafts(context.Background(), owner)
	if err != nil {
		t.Fatalf("list drafts: %v", err)
	}
	if len(drafts) != 1 || drafts[0].AmountVND != 75000 || drafts[0].Note != "Mới" {
		t.Fatalf("updated resumed schedule should create one updated draft: %#v", drafts)
	}
}

func TestRecurringScheduleAutoPostCreatesConfirmedTransactionOnce(t *testing.T) {
	conn := migratedPlanningPostgres(t)
	owner := createPlanningUser(t, conn, "recurring-autopost@example.com")
	financeRepo := finance.NewRepository(conn)
	planningRepo := planning.NewRepository(conn)
	wallet := createPlanningWallet(t, financeRepo, owner)
	setPlanningWalletBalance(t, conn, wallet.ID, 500000)
	categoryID := findPlanningSystemCategory(t, conn, "expense_food")

	schedule, err := planningRepo.CreateRecurringSchedule(context.Background(), owner, planning.CreateRecurringScheduleInput{
		Name: "Auto", Frequency: planning.RecurrenceDaily, Timezone: "Asia/Ho_Chi_Minh",
		StartsAt: "2026-09-01T07:00:00+07:00", EndsAt: "2026-09-01T07:00:00+07:00",
		PostingMode: planning.RecurringPostingAutoPost,
		Type:        finance.TransactionExpense, SourceWalletID: wallet.ID, CategoryID: categoryID, AmountVND: 100000, Note: "Auto-post",
	})
	if err != nil {
		t.Fatalf("create auto-post schedule: %v", err)
	}
	if schedule.PostingMode != planning.RecurringPostingAutoPost {
		t.Fatalf("unexpected posting mode: %#v", schedule)
	}
	processed, err := planningRepo.ProcessDueRecurringSchedules(context.Background(), "worker-auto", time.Date(2026, 9, 3, 2, 0, 0, 0, time.UTC), 10)
	if err != nil {
		t.Fatalf("process auto-post schedule: %v", err)
	}
	if processed != 1 {
		t.Fatalf("auto-post ended schedule should process once, got %d", processed)
	}
	again, err := planningRepo.ProcessDueRecurringSchedules(context.Background(), "worker-auto-again", time.Date(2026, 9, 3, 2, 0, 0, 0, time.UTC), 10)
	if err != nil {
		t.Fatalf("process auto-post schedule again: %v", err)
	}
	if again != 0 {
		t.Fatalf("auto-post occurrence should be idempotent and ended, got %d", again)
	}
	assertPlanningCount(t, conn, `SELECT count(*) FROM transaction_drafts WHERE user_id = $1`, 0, owner)
	assertPlanningCount(t, conn, `SELECT count(*) FROM transactions WHERE user_id = $1 AND note = 'Auto-post'`, 1, owner)
	if got := walletBalance(t, conn, wallet.ID); got != 400000 {
		t.Fatalf("auto-post should apply accounting once, got balance %d", got)
	}
}

func TestRecurringCatchUpIsBoundedByWorkerLimit(t *testing.T) {
	conn := migratedPlanningPostgres(t)
	owner := createPlanningUser(t, conn, "recurring-bounded@example.com")
	financeRepo := finance.NewRepository(conn)
	planningRepo := planning.NewRepository(conn)
	wallet := createPlanningWallet(t, financeRepo, owner)
	categoryID := findPlanningSystemCategory(t, conn, "expense_food")

	_, err := planningRepo.CreateRecurringSchedule(context.Background(), owner, planning.CreateRecurringScheduleInput{
		Name: "Daily", Frequency: planning.RecurrenceDaily, Timezone: "Asia/Ho_Chi_Minh",
		StartsAt: "2026-01-01T09:00:00+07:00", Type: finance.TransactionExpense,
		SourceWalletID: wallet.ID, CategoryID: categoryID, AmountVND: 10000,
	})
	if err != nil {
		t.Fatalf("create recurring schedule: %v", err)
	}

	processed, err := planningRepo.ProcessDueRecurringSchedules(context.Background(), "worker-a", time.Date(2026, 9, 10, 2, 0, 0, 0, time.UTC), 2)
	if err != nil {
		t.Fatalf("bounded catch-up: %v", err)
	}
	if processed != 2 {
		t.Fatalf("expected exactly two occurrence attempts, got %d", processed)
	}
	var drafts int
	if err := conn.QueryRowContext(context.Background(), `SELECT count(*) FROM transaction_drafts WHERE user_id = $1`, owner).Scan(&drafts); err != nil {
		t.Fatalf("count drafts: %v", err)
	}
	if drafts != 2 {
		t.Fatalf("worker limit must bound inserted drafts, got %d", drafts)
	}
}

func mergeRecurringInput(base planning.CreateRecurringScheduleInput, txType finance.TransactionType, destinationWalletID string, categoryID string) planning.CreateRecurringScheduleInput {
	base.Type = txType
	base.DestinationWalletID = destinationWalletID
	base.CategoryID = categoryID
	return base
}

func walletBalance(t *testing.T, conn *sql.DB, walletID string) int64 {
	t.Helper()
	var balance int64
	if err := conn.QueryRowContext(context.Background(), `SELECT balance_vnd FROM wallets WHERE id = $1`, walletID).Scan(&balance); err != nil {
		t.Fatalf("wallet balance: %v", err)
	}
	return balance
}
