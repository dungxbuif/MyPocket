package finance_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"mypocket/internal/finance"
)

func TestCreateTransactionInTxRollsBackAllAccountingEffects(t *testing.T) {
	conn := migratedFinancePostgres(t)
	ctx := context.Background()
	owner := createFinanceUser(t, conn, "tx-boundary@example.com")
	repo := finance.NewRepository(conn)
	wallet := createFinanceWallet(t, repo, owner, "Tiền mặt", finance.WalletCash)
	categoryID := findSystemCategory(t, conn, "expense_food")

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin caller transaction: %v", err)
	}
	created, err := finance.CreateTransactionInTx(ctx, tx, owner, finance.CreateTransactionInput{
		IdempotencyKey: "caller-owned-transaction",
		Type:           finance.TransactionExpense,
		SourceWalletID: wallet.ID,
		CategoryID:     categoryID,
		AmountVND:      125000,
		OccurredAt:     time.Date(2026, 9, 10, 2, 0, 0, 0, time.UTC),
		Note:           "Bữa sáng",
	})
	if err != nil {
		_ = tx.Rollback()
		t.Fatalf("create transaction in caller boundary: %v", err)
	}
	if created.ID == "" {
		_ = tx.Rollback()
		t.Fatal("expected transaction identity before rollback")
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback caller transaction: %v", err)
	}

	if got := walletBalanceByID(t, conn, wallet.ID); got != 0 {
		t.Fatalf("wallet effect escaped rollback: got %d", got)
	}
	assertFinanceRowCount(t, conn, `SELECT count(*) FROM transactions WHERE id = $1`, created.ID, 0)
	assertFinanceRowCount(t, conn, `SELECT count(*) FROM finance_idempotency_keys WHERE user_id = $1 AND key = $2`, owner, "caller-owned-transaction", 0)
}

func walletBalanceByID(t *testing.T, conn *sql.DB, walletID string) int64 {
	t.Helper()
	var balance int64
	if err := conn.QueryRowContext(context.Background(), `SELECT balance_vnd FROM wallets WHERE id = $1`, walletID).Scan(&balance); err != nil {
		t.Fatalf("load wallet balance: %v", err)
	}
	return balance
}

func assertFinanceRowCount(t *testing.T, conn *sql.DB, query string, argsAndWant ...any) {
	t.Helper()
	want := argsAndWant[len(argsAndWant)-1].(int)
	args := argsAndWant[:len(argsAndWant)-1]
	var got int
	if err := conn.QueryRowContext(context.Background(), query, args...).Scan(&got); err != nil {
		t.Fatalf("count rows: %v", err)
	}
	if got != want {
		t.Fatalf("expected %d rows, got %d", want, got)
	}
}
