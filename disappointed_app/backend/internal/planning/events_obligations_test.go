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
	_, err = planningRepo.UpdateObligation(context.Background(), owner, obligation.ID, planning.UpdateObligationInput{BaseVersion: obligation.Version, Direction: planning.ObligationBorrowed, PrincipalVND: 500000, Counterparty: "Anh Minh", DueOn: "2026-09-30", Note: "Vay sửa nhà"})
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

func TestObligationRepaymentRequiresMatchingTransactionDirection(t *testing.T) {
	conn := migratedPlanningPostgres(t)
	owner := createPlanningUser(t, conn, "obligation-direction@example.com")
	financeRepo := finance.NewRepository(conn)
	planningRepo := planning.NewRepository(conn)
	wallet := createPlanningWallet(t, financeRepo, owner)
	expenseCategoryID := findPlanningSystemCategory(t, conn, "expense_food")
	incomeCategoryID := findPlanningSystemCategory(t, conn, "income_salary")
	occurred := time.Date(2026, 9, 10, 5, 0, 0, 0, time.UTC)
	expense := createPlanningTransaction(t, financeRepo, owner, finance.CreateTransactionInput{IdempotencyKey: "direction-expense", Type: finance.TransactionExpense, SourceWalletID: wallet.ID, CategoryID: expenseCategoryID, AmountVND: 100000, OccurredAt: occurred})
	income := createPlanningTransaction(t, financeRepo, owner, finance.CreateTransactionInput{IdempotencyKey: "direction-income", Type: finance.TransactionIncome, SourceWalletID: wallet.ID, CategoryID: incomeCategoryID, AmountVND: 100000, OccurredAt: occurred})

	borrowed, err := planningRepo.CreateObligation(context.Background(), owner, planning.CreateObligationInput{Direction: planning.ObligationBorrowed, PrincipalVND: 500000, Counterparty: "Anh Minh", DueOn: "2026-09-30"})
	if err != nil {
		t.Fatalf("create borrowed obligation: %v", err)
	}
	lent, err := planningRepo.CreateObligation(context.Background(), owner, planning.CreateObligationInput{Direction: planning.ObligationLent, PrincipalVND: 500000, Counterparty: "Chị Lan", DueOn: "2026-09-30"})
	if err != nil {
		t.Fatalf("create lent obligation: %v", err)
	}

	if err := planningRepo.LinkObligationRepayment(context.Background(), owner, borrowed.ID, income.ID); !errors.Is(err, planning.ErrValidation) {
		t.Fatalf("borrowed repayment must be an expense, got %v", err)
	}
	if err := planningRepo.LinkObligationRepayment(context.Background(), owner, lent.ID, expense.ID); !errors.Is(err, planning.ErrValidation) {
		t.Fatalf("lent repayment must be income, got %v", err)
	}
	if err := planningRepo.LinkObligationRepayment(context.Background(), owner, borrowed.ID, expense.ID); err != nil {
		t.Fatalf("borrowed expense repayment should pass: %v", err)
	}
	if err := planningRepo.LinkObligationRepayment(context.Background(), owner, lent.ID, income.ID); err != nil {
		t.Fatalf("lent income repayment should pass: %v", err)
	}
}

func TestObligationRepaymentCannotCountOneTransactionTwice(t *testing.T) {
	conn := migratedPlanningPostgres(t)
	owner := createPlanningUser(t, conn, "obligation-double-count@example.com")
	financeRepo := finance.NewRepository(conn)
	planningRepo := planning.NewRepository(conn)
	wallet := createPlanningWallet(t, financeRepo, owner)
	categoryID := findPlanningSystemCategory(t, conn, "expense_food")
	repayment := createPlanningTransaction(t, financeRepo, owner, finance.CreateTransactionInput{IdempotencyKey: "one-repayment", Type: finance.TransactionExpense, SourceWalletID: wallet.ID, CategoryID: categoryID, AmountVND: 100000, OccurredAt: time.Date(2026, 9, 10, 5, 0, 0, 0, time.UTC)})
	first, err := planningRepo.CreateObligation(context.Background(), owner, planning.CreateObligationInput{Direction: planning.ObligationBorrowed, PrincipalVND: 500000, Counterparty: "A", DueOn: "2026-09-30"})
	if err != nil {
		t.Fatalf("create first obligation: %v", err)
	}
	second, err := planningRepo.CreateObligation(context.Background(), owner, planning.CreateObligationInput{Direction: planning.ObligationBorrowed, PrincipalVND: 500000, Counterparty: "B", DueOn: "2026-09-30"})
	if err != nil {
		t.Fatalf("create second obligation: %v", err)
	}
	if err := planningRepo.LinkObligationRepayment(context.Background(), owner, first.ID, repayment.ID); err != nil {
		t.Fatalf("link first obligation: %v", err)
	}
	if err := planningRepo.LinkObligationRepayment(context.Background(), owner, second.ID, repayment.ID); !errors.Is(err, planning.ErrValidation) {
		t.Fatalf("same transaction must not repay two obligations, got %v", err)
	}
}
