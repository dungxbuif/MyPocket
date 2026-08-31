package planning_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"mypocket/internal/finance"
	"mypocket/internal/planning"
)

func TestEventTotalsUseLinkedIncludedOwnedTransactions(t *testing.T) {
	conn := migratedPlanningPostgres(t)
	owner := createPlanningUser(t, conn, "event-owner@example.com")
	other := createPlanningUser(t, conn, "event-other@example.com")
	financeRepo := finance.NewRepository(conn)
	planningRepo := planning.NewRepository(conn)
	wallet := createPlanningWallet(t, financeRepo, owner)
	otherWallet := createPlanningWallet(t, financeRepo, other)
	categoryID := findPlanningSystemCategory(t, conn, "expense_food")
	occurred := time.Date(2026, 8, 31, 2, 0, 0, 0, time.UTC)

	counted := createPlanningTransaction(t, financeRepo, owner, finance.CreateTransactionInput{IdempotencyKey: "event-counted", Type: finance.TransactionExpense, SourceWalletID: wallet.ID, CategoryID: categoryID, AmountVND: 125000, OccurredAt: occurred, Note: "Counted event"})
	excluded := createPlanningTransaction(t, financeRepo, owner, finance.CreateTransactionInput{IdempotencyKey: "event-excluded", Type: finance.TransactionExpense, SourceWalletID: wallet.ID, CategoryID: categoryID, AmountVND: 50000, OccurredAt: occurred, Note: "Excluded event", ExcludedFromReports: true})
	otherTx := createPlanningTransaction(t, financeRepo, other, finance.CreateTransactionInput{IdempotencyKey: "event-other", Type: finance.TransactionExpense, SourceWalletID: otherWallet.ID, CategoryID: categoryID, AmountVND: 75000, OccurredAt: occurred, Note: "Other event"})
	event, err := planningRepo.CreateEvent(context.Background(), owner, planning.CreateEventInput{Name: "Đà Lạt", StartsOn: "2026-08-31", EndsOn: "2026-09-02", Note: "Trip"})
	if err != nil {
		t.Fatalf("create event: %v", err)
	}

	if err := planningRepo.LinkEventTransaction(context.Background(), owner, event.ID, counted.ID); err != nil {
		t.Fatalf("link counted: %v", err)
	}
	if err := planningRepo.LinkEventTransaction(context.Background(), owner, event.ID, excluded.ID); err != nil {
		t.Fatalf("link excluded: %v", err)
	}
	if err := planningRepo.LinkEventTransaction(context.Background(), owner, event.ID, otherTx.ID); !errors.Is(err, planning.ErrForbidden) {
		t.Fatalf("expected other transaction rejection, got %v", err)
	}
	events, err := planningRepo.ListEvents(context.Background(), owner)
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	if len(events) != 1 || events[0].TotalVND != 125000 || events[0].TransactionCount != 1 {
		t.Fatalf("event total should count only linked included owner transactions: %#v", events)
	}
}

func TestObligationRepaymentsCannotOverpayPrincipal(t *testing.T) {
	conn := migratedPlanningPostgres(t)
	owner := createPlanningUser(t, conn, "obligation-owner@example.com")
	financeRepo := finance.NewRepository(conn)
	planningRepo := planning.NewRepository(conn)
	wallet := createPlanningWallet(t, financeRepo, owner)
	categoryID := findPlanningSystemCategory(t, conn, "expense_food")
	occurred := time.Date(2026, 8, 31, 3, 0, 0, 0, time.UTC)
	first := createPlanningTransaction(t, financeRepo, owner, finance.CreateTransactionInput{IdempotencyKey: "repayment-1", Type: finance.TransactionExpense, SourceWalletID: wallet.ID, CategoryID: categoryID, AmountVND: 600000, OccurredAt: occurred, Note: "Trả nợ 1"})
	second := createPlanningTransaction(t, financeRepo, owner, finance.CreateTransactionInput{IdempotencyKey: "repayment-2", Type: finance.TransactionExpense, SourceWalletID: wallet.ID, CategoryID: categoryID, AmountVND: 500000, OccurredAt: occurred, Note: "Trả nợ 2"})
	obligation, err := planningRepo.CreateObligation(context.Background(), owner, planning.CreateObligationInput{Direction: planning.ObligationBorrowed, PrincipalVND: 1000000, Counterparty: "Anh Minh", DueOn: "2026-09-30", Note: "Vay sửa nhà"})
	if err != nil {
		t.Fatalf("create obligation: %v", err)
	}

	if err := planningRepo.LinkObligationRepayment(context.Background(), owner, obligation.ID, first.ID); err != nil {
		t.Fatalf("link first repayment: %v", err)
	}
	if err := planningRepo.LinkObligationRepayment(context.Background(), owner, obligation.ID, first.ID); err != nil {
		t.Fatalf("relink same repayment should be idempotent: %v", err)
	}
	if err := planningRepo.LinkObligationRepayment(context.Background(), owner, obligation.ID, second.ID); !errors.Is(err, planning.ErrValidation) {
		t.Fatalf("expected overpayment validation, got %v", err)
	}
	_, err = planningRepo.UpdateObligation(context.Background(), owner, obligation.ID, planning.UpdateObligationInput{Direction: planning.ObligationBorrowed, PrincipalVND: 500000, Counterparty: "Anh Minh", DueOn: "2026-09-30", Note: "Vay sửa nhà"})
	if !errors.Is(err, planning.ErrValidation) {
		t.Fatalf("expected lower-principal validation, got %v", err)
	}
	obligations, err := planningRepo.ListObligations(context.Background(), owner)
	if err != nil {
		t.Fatalf("list obligations: %v", err)
	}
	if len(obligations) != 1 || obligations[0].RepaidVND != 600000 || obligations[0].RemainingVND != 400000 {
		t.Fatalf("wrong obligation remaining amount: %#v", obligations)
	}
}
