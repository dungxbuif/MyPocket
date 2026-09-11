package planning_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"mypocket/internal/finance"
	"mypocket/internal/planning"
)

func TestConfirmTransactionDraftPostsAccountingForEachSupportedType(t *testing.T) {
	cases := []struct {
		name              string
		txType            finance.TransactionType
		categoryKey       string
		sourceBefore      int64
		destinationBefore int64
		wantSource        int64
		wantDestination   int64
	}{
		{name: "expense", txType: finance.TransactionExpense, categoryKey: "expense_food", sourceBefore: 900000, wantSource: 775000},
		{name: "income", txType: finance.TransactionIncome, categoryKey: "income_salary", sourceBefore: 900000, wantSource: 1025000},
		{name: "transfer", txType: finance.TransactionTransfer, sourceBefore: 900000, destinationBefore: 300000, wantSource: 775000, wantDestination: 425000},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newDraftDecisionFixture(t)
			setPlanningWalletBalance(t, fixture.conn, fixture.source.ID, tc.sourceBefore)
			setPlanningWalletBalance(t, fixture.conn, fixture.destination.ID, tc.destinationBefore)
			categoryID := ""
			if tc.categoryKey != "" {
				categoryID = findPlanningSystemCategory(t, fixture.conn, tc.categoryKey)
			}
			destinationID := ""
			if tc.txType == finance.TransactionTransfer {
				destinationID = fixture.destination.ID
			}
			draftID := insertTransactionDraft(t, fixture.conn, fixture.owner, "accounting-"+tc.name, tc.txType, fixture.source.ID, destinationID, categoryID, 100000, "Scheduled")

			decision, err := fixture.repo.ConfirmTransactionDraft(context.Background(), fixture.owner, draftID, planning.ConfirmTransactionDraftInput{
				Version:        1,
				AmountVND:      125000,
				Note:           "Accepted",
				IdempotencyKey: "client-accounting-" + tc.name,
			})
			if err != nil {
				t.Fatalf("confirm %s draft: %v", tc.name, err)
			}
			if decision.Transaction == nil || decision.Transaction.ID == "" || decision.Transaction.Type != tc.txType || decision.Transaction.AmountVND != 125000 || decision.Transaction.Note != "Accepted" {
				t.Fatalf("unexpected confirmed transaction: %#v", decision.Transaction)
			}
			if decision.Draft.Status != "confirmed" || decision.Draft.Version != 2 || decision.Draft.AmountVND != 125000 || decision.Draft.Note != "Accepted" || decision.Draft.ConfirmedTransactionID != decision.Transaction.ID {
				t.Fatalf("unexpected terminal draft: %#v", decision.Draft)
			}
			assertPlanningCount(t, fixture.conn, `SELECT count(*) FROM sync_changes WHERE user_id=$1 AND entity_type='transaction' AND entity_id=$2 AND operation='create'`, 1, fixture.owner, decision.Transaction.ID)
			if got := walletBalance(t, fixture.conn, fixture.source.ID); got != tc.wantSource {
				t.Fatalf("source balance: got %d want %d", got, tc.wantSource)
			}
			if got := walletBalance(t, fixture.conn, fixture.destination.ID); got != tc.wantDestination {
				t.Fatalf("destination balance: got %d want %d", got, tc.wantDestination)
			}
			drafts, err := fixture.repo.ListTransactionDrafts(context.Background(), fixture.owner)
			if err != nil || len(drafts) != 1 || drafts[0].ConfirmedTransactionID != decision.Transaction.ID {
				t.Fatalf("list must expose confirmed transaction id: %#v err=%v", drafts, err)
			}
		})
	}
}

func TestConfirmTransactionDraftReplayPrecedesVersionCheckAndHasNoNewEffects(t *testing.T) {
	fixture := newDraftDecisionFixture(t)
	setPlanningWalletBalance(t, fixture.conn, fixture.source.ID, 500000)
	categoryID := findPlanningSystemCategory(t, fixture.conn, "expense_food")
	draftID := insertTransactionDraft(t, fixture.conn, fixture.owner, "replay", finance.TransactionExpense, fixture.source.ID, "", categoryID, 100000, "Scheduled")
	input := planning.ConfirmTransactionDraftInput{Version: 1, AmountVND: 125000, Note: "Accepted", IdempotencyKey: "first-key"}

	first, err := fixture.repo.ConfirmTransactionDraft(context.Background(), fixture.owner, draftID, input)
	if err != nil {
		t.Fatalf("first confirm: %v", err)
	}
	replay, err := fixture.repo.ConfirmTransactionDraft(context.Background(), fixture.owner, draftID, planning.ConfirmTransactionDraftInput{Version: 1, AmountVND: 125000, Note: "Accepted", IdempotencyKey: "different-retry-key"})
	if err != nil {
		t.Fatalf("terminal replay with stale version: %v", err)
	}
	if replay.Transaction == nil || first.Transaction == nil || replay.Transaction.ID != first.Transaction.ID || replay.Draft.Version != 2 {
		t.Fatalf("replay changed terminal identity/version: first=%#v replay=%#v", first, replay)
	}
	if got := walletBalance(t, fixture.conn, fixture.source.ID); got != 375000 {
		t.Fatalf("replay changed balance twice: %d", got)
	}
	assertPlanningCount(t, fixture.conn, `SELECT count(*) FROM transactions WHERE user_id = $1`, 1, fixture.owner)
	assertPlanningCount(t, fixture.conn, `SELECT count(*) FROM finance_idempotency_keys WHERE user_id = $1`, 1, fixture.owner)
}

func TestConcurrentTransactionDraftConfirmationsReturnOneAccountingResult(t *testing.T) {
	fixture := newDraftDecisionFixture(t)
	setPlanningWalletBalance(t, fixture.conn, fixture.source.ID, 500000)
	categoryID := findPlanningSystemCategory(t, fixture.conn, "expense_food")
	draftID := insertTransactionDraft(t, fixture.conn, fixture.owner, "concurrent-confirm", finance.TransactionExpense, fixture.source.ID, "", categoryID, 100000, "Scheduled")

	type result struct {
		decision planning.TransactionDraftDecision
		err      error
	}
	start := make(chan struct{})
	results := make(chan result, 2)
	for _, key := range []string{"concurrent-key-a", "concurrent-key-b"} {
		key := key
		go func() {
			<-start
			decision, err := fixture.repo.ConfirmTransactionDraft(context.Background(), fixture.owner, draftID, planning.ConfirmTransactionDraftInput{Version: 1, AmountVND: 125000, Note: "Accepted", IdempotencyKey: key})
			results <- result{decision: decision, err: err}
		}()
	}
	close(start)
	first := <-results
	second := <-results
	if first.err != nil || second.err != nil {
		t.Fatalf("concurrent confirmations must both resolve to success: first=%v second=%v", first.err, second.err)
	}
	if first.decision.Transaction == nil || second.decision.Transaction == nil || first.decision.Transaction.ID != second.decision.Transaction.ID {
		t.Fatalf("concurrent confirmations returned different accounting identities: first=%#v second=%#v", first.decision, second.decision)
	}
	if got := walletBalance(t, fixture.conn, fixture.source.ID); got != 375000 {
		t.Fatalf("concurrent confirmation applied accounting more than once: %d", got)
	}
	assertPlanningCount(t, fixture.conn, `SELECT count(*) FROM transactions WHERE user_id = $1`, 1, fixture.owner)
	assertPlanningCount(t, fixture.conn, `SELECT count(*) FROM finance_idempotency_keys WHERE user_id = $1`, 1, fixture.owner)
}

func TestTransactionDraftTerminalConflictsAndRejectReplay(t *testing.T) {
	t.Run("confirmed draft rejects changed confirmation and rejection", func(t *testing.T) {
		fixture := newDraftDecisionFixture(t)
		categoryID := findPlanningSystemCategory(t, fixture.conn, "expense_food")
		draftID := insertTransactionDraft(t, fixture.conn, fixture.owner, "confirmed-conflict", finance.TransactionExpense, fixture.source.ID, "", categoryID, 100000, "Scheduled")
		_, err := fixture.repo.ConfirmTransactionDraft(context.Background(), fixture.owner, draftID, planning.ConfirmTransactionDraftInput{Version: 1, AmountVND: 100000, Note: "Accepted", IdempotencyKey: "confirm-key"})
		if err != nil {
			t.Fatalf("first confirm: %v", err)
		}
		if _, err := fixture.repo.ConfirmTransactionDraft(context.Background(), fixture.owner, draftID, planning.ConfirmTransactionDraftInput{Version: 1, AmountVND: 100001, Note: "Accepted", IdempotencyKey: "confirm-key"}); !errors.Is(err, planning.ErrDraftAlreadyResolved) {
			t.Fatalf("changed confirm must conflict: %v", err)
		}
		if _, err := fixture.repo.RejectTransactionDraft(context.Background(), fixture.owner, draftID, planning.RejectTransactionDraftInput{Version: 1}); !errors.Is(err, planning.ErrDraftAlreadyResolved) {
			t.Fatalf("reject after confirm must conflict: %v", err)
		}
	})

	t.Run("rejected draft replays before stale version and blocks confirmation", func(t *testing.T) {
		fixture := newDraftDecisionFixture(t)
		categoryID := findPlanningSystemCategory(t, fixture.conn, "expense_food")
		draftID := insertTransactionDraft(t, fixture.conn, fixture.owner, "rejected-replay", finance.TransactionExpense, fixture.source.ID, "", categoryID, 100000, "Scheduled")
		first, err := fixture.repo.RejectTransactionDraft(context.Background(), fixture.owner, draftID, planning.RejectTransactionDraftInput{Version: 1})
		if err != nil {
			t.Fatalf("first reject: %v", err)
		}
		replay, err := fixture.repo.RejectTransactionDraft(context.Background(), fixture.owner, draftID, planning.RejectTransactionDraftInput{Version: 1})
		if err != nil || replay.Draft.ID != first.Draft.ID || replay.Draft.Version != 2 || replay.Draft.Status != "rejected" || replay.Transaction != nil {
			t.Fatalf("unexpected reject replay: first=%#v replay=%#v err=%v", first, replay, err)
		}
		if _, err := fixture.repo.ConfirmTransactionDraft(context.Background(), fixture.owner, draftID, planning.ConfirmTransactionDraftInput{Version: 1, AmountVND: 100000, Note: "Scheduled", IdempotencyKey: "late-confirm"}); !errors.Is(err, planning.ErrDraftAlreadyResolved) {
			t.Fatalf("confirm after reject must conflict: %v", err)
		}
		assertPlanningCount(t, fixture.conn, `SELECT count(*) FROM transactions WHERE user_id = $1`, 0, fixture.owner)
	})
}

func TestTransactionDraftDecisionsEnforceVersionAndOwnership(t *testing.T) {
	fixture := newDraftDecisionFixture(t)
	categoryID := findPlanningSystemCategory(t, fixture.conn, "expense_food")
	draftID := insertTransactionDraft(t, fixture.conn, fixture.owner, "guards", finance.TransactionExpense, fixture.source.ID, "", categoryID, 100000, "Scheduled")

	if _, err := fixture.repo.ConfirmTransactionDraft(context.Background(), fixture.owner, draftID, planning.ConfirmTransactionDraftInput{Version: 2, AmountVND: 100000, Note: "Scheduled", IdempotencyKey: "stale"}); !errors.Is(err, planning.ErrDraftVersionConflict) {
		t.Fatalf("expected version conflict, got %v", err)
	}
	if _, err := fixture.repo.ConfirmTransactionDraft(context.Background(), fixture.other, draftID, planning.ConfirmTransactionDraftInput{Version: 1, AmountVND: 100000, Note: "Scheduled", IdempotencyKey: "foreign"}); !errors.Is(err, planning.ErrForbidden) {
		t.Fatalf("expected owner-hiding forbidden, got %v", err)
	}
	var status string
	var version int64
	if err := fixture.conn.QueryRowContext(context.Background(), `SELECT status, version FROM transaction_drafts WHERE id = $1`, draftID).Scan(&status, &version); err != nil {
		t.Fatalf("load guarded draft: %v", err)
	}
	if status != "pending" || version != 1 || walletBalance(t, fixture.conn, fixture.source.ID) != 0 {
		t.Fatalf("guard failures changed state: status=%s version=%d balance=%d", status, version, walletBalance(t, fixture.conn, fixture.source.ID))
	}
}

func TestConfirmTransactionDraftTreatsInvalidStoredFinanceReferencesAsValidation(t *testing.T) {
	cases := []string{"foreign wallet", "archived wallet", "foreign category", "archived category", "inactive category", "wrong-kind category"}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			fixture := newDraftDecisionFixture(t)
			financeRepo := finance.NewRepository(fixture.conn)
			sourceWalletID := fixture.source.ID
			categoryID := findPlanningSystemCategory(t, fixture.conn, "expense_food")

			switch name {
			case "foreign wallet":
				sourceWalletID = createPlanningWallet(t, financeRepo, fixture.other).ID
			case "foreign category":
				category, err := financeRepo.CreateCategory(context.Background(), fixture.other, finance.CreateCategoryInput{Kind: finance.CategoryExpense, Name: "Other owner"})
				if err != nil {
					t.Fatalf("create foreign category: %v", err)
				}
				categoryID = category.ID
			case "archived category":
				category, err := financeRepo.CreateCategory(context.Background(), fixture.owner, finance.CreateCategoryInput{Kind: finance.CategoryExpense, Name: "Archived"})
				if err != nil {
					t.Fatalf("create archived category: %v", err)
				}
				categoryID = category.ID
			case "wrong-kind category":
				categoryID = findPlanningSystemCategory(t, fixture.conn, "income_salary")
			}

			draftID := insertTransactionDraft(t, fixture.conn, fixture.owner, "invalid-ref-"+name, finance.TransactionExpense, sourceWalletID, "", categoryID, 100000, "Scheduled")
			switch name {
			case "archived wallet":
				if err := financeRepo.ArchiveWallet(context.Background(), fixture.owner, fixture.source.ID, fixture.source.Version); err != nil {
					t.Fatalf("archive source wallet: %v", err)
				}
			case "archived category":
				if _, err := fixture.conn.ExecContext(context.Background(), `UPDATE categories SET archived_at = now() WHERE id = $1`, categoryID); err != nil {
					t.Fatalf("archive category fixture: %v", err)
				}
			case "inactive category":
				if err := financeRepo.SetWalletCategoryActive(context.Background(), fixture.owner, fixture.source.ID, categoryID, false); err != nil {
					t.Fatalf("deactivate category: %v", err)
				}
			}

			_, err := fixture.repo.ConfirmTransactionDraft(context.Background(), fixture.owner, draftID, planning.ConfirmTransactionDraftInput{Version: 1, AmountVND: 100000, Note: "Scheduled", IdempotencyKey: "invalid-ref-key"})
			if !errors.Is(err, planning.ErrValidation) {
				t.Fatalf("invalid stored finance reference must be validation, got %v", err)
			}
			assertPlanningCount(t, fixture.conn, `SELECT count(*) FROM transactions WHERE user_id = $1`, 0, fixture.owner)
			assertPlanningCount(t, fixture.conn, `SELECT count(*) FROM finance_idempotency_keys WHERE user_id = $1`, 0, fixture.owner)
			var status string
			var version int64
			if scanErr := fixture.conn.QueryRowContext(context.Background(), `SELECT status, version FROM transaction_drafts WHERE id = $1`, draftID).Scan(&status, &version); scanErr != nil {
				t.Fatalf("load pending draft: %v", scanErr)
			}
			if status != "pending" || version != 1 {
				t.Fatalf("invalid reference changed draft: status=%s version=%d", status, version)
			}
			if got := walletBalance(t, fixture.conn, sourceWalletID); got != 0 {
				t.Fatalf("invalid reference changed wallet balance: %d", got)
			}
		})
	}
}

func TestConfirmTransactionDraftScopesRawClientKeyPerDraft(t *testing.T) {
	fixture := newDraftDecisionFixture(t)
	categoryID := findPlanningSystemCategory(t, fixture.conn, "expense_food")
	firstID := insertTransactionDraft(t, fixture.conn, fixture.owner, "key-scope-a", finance.TransactionExpense, fixture.source.ID, "", categoryID, 100000, "Same")
	secondID := insertTransactionDraft(t, fixture.conn, fixture.owner, "key-scope-b", finance.TransactionExpense, fixture.source.ID, "", categoryID, 100000, "Same")
	input := planning.ConfirmTransactionDraftInput{Version: 1, AmountVND: 100000, Note: "Same", IdempotencyKey: "shared-client-key"}

	first, err := fixture.repo.ConfirmTransactionDraft(context.Background(), fixture.owner, firstID, input)
	if err != nil {
		t.Fatalf("confirm first draft: %v", err)
	}
	second, err := fixture.repo.ConfirmTransactionDraft(context.Background(), fixture.owner, secondID, input)
	if err != nil {
		t.Fatalf("confirm second draft with same client key: %v", err)
	}
	if first.Transaction == nil || second.Transaction == nil || first.Transaction.ID == second.Transaction.ID {
		t.Fatalf("draft-scoped confirmations linked same transaction: first=%#v second=%#v", first, second)
	}
	assertPlanningCount(t, fixture.conn, `SELECT count(*) FROM finance_idempotency_keys WHERE user_id = $1 AND key <> $2 AND length(key) = 64`, 2, fixture.owner, "shared-client-key")
	assertPlanningCount(t, fixture.conn, `SELECT count(*) FROM finance_idempotency_keys WHERE user_id = $1 AND key = $2`, 0, fixture.owner, "shared-client-key")
}

func TestConfirmTransactionDraftRollsBackAccountingWhenFinalizationFails(t *testing.T) {
	fixture := newDraftDecisionFixture(t)
	setPlanningWalletBalance(t, fixture.conn, fixture.source.ID, 500000)
	categoryID := findPlanningSystemCategory(t, fixture.conn, "expense_food")
	draftID := insertTransactionDraft(t, fixture.conn, fixture.owner, "rollback", finance.TransactionExpense, fixture.source.ID, "", categoryID, 100000, "Scheduled")
	if _, err := fixture.conn.ExecContext(context.Background(), `
		CREATE FUNCTION fail_draft_confirmation() RETURNS trigger LANGUAGE plpgsql AS $$
		BEGIN
			RAISE EXCEPTION 'injected draft finalization failure';
		END;
		$$;
		CREATE TRIGGER fail_draft_confirmation
		BEFORE UPDATE ON transaction_drafts
		FOR EACH ROW WHEN (NEW.status = 'confirmed')
		EXECUTE FUNCTION fail_draft_confirmation();
	`); err != nil {
		t.Fatalf("install failure injection: %v", err)
	}

	_, err := fixture.repo.ConfirmTransactionDraft(context.Background(), fixture.owner, draftID, planning.ConfirmTransactionDraftInput{Version: 1, AmountVND: 125000, Note: "Accepted", IdempotencyKey: "rollback-key"})
	if err == nil {
		t.Fatal("expected injected finalization failure")
	}
	if got := walletBalance(t, fixture.conn, fixture.source.ID); got != 500000 {
		t.Fatalf("wallet effect escaped rollback: %d", got)
	}
	assertPlanningCount(t, fixture.conn, `SELECT count(*) FROM transactions WHERE user_id = $1`, 0, fixture.owner)
	assertPlanningCount(t, fixture.conn, `SELECT count(*) FROM finance_idempotency_keys WHERE user_id = $1`, 0, fixture.owner)
	var status string
	assertPlanningCount(t, fixture.conn, `SELECT count(*) FROM sync_changes WHERE user_id=$1 AND operation <> 'create'`, 0, fixture.owner)
	assertPlanningCount(t, fixture.conn, `SELECT count(*) FROM sync_changes WHERE user_id=$1 AND entity_type='transaction'`, 0, fixture.owner)
	var version int64
	var confirmed sql.NullString
	if scanErr := fixture.conn.QueryRowContext(context.Background(), `SELECT status, version, confirmed_transaction_id::text FROM transaction_drafts WHERE id = $1`, draftID).Scan(&status, &version, &confirmed); scanErr != nil {
		t.Fatalf("load rolled-back draft: %v", scanErr)
	}
	if status != "pending" || version != 1 || confirmed.Valid {
		t.Fatalf("draft finalization escaped rollback: status=%s version=%d confirmed=%#v", status, version, confirmed)
	}
}

type draftDecisionFixture struct {
	conn        *sql.DB
	repo        *planning.Repository
	owner       string
	other       string
	source      finance.Wallet
	destination finance.Wallet
}

func newDraftDecisionFixture(t *testing.T) draftDecisionFixture {
	t.Helper()
	conn := migratedPlanningPostgres(t)
	owner := createPlanningUser(t, conn, fmt.Sprintf("draft-owner-%d@example.com", time.Now().UnixNano()))
	other := createPlanningUser(t, conn, fmt.Sprintf("draft-other-%d@example.com", time.Now().UnixNano()))
	financeRepo := finance.NewRepository(conn)
	return draftDecisionFixture{
		conn:        conn,
		repo:        planning.NewRepository(conn),
		owner:       owner,
		other:       other,
		source:      createPlanningWallet(t, financeRepo, owner),
		destination: createPlanningWallet(t, financeRepo, owner),
	}
}

func insertTransactionDraft(t *testing.T, conn *sql.DB, userID, occurrenceKey string, txType finance.TransactionType, sourceWalletID, destinationWalletID, categoryID string, amountVND int64, note string) string {
	t.Helper()
	var id string
	if err := conn.QueryRowContext(context.Background(), `
		INSERT INTO transaction_drafts (user_id, occurrence_key, transaction_type, source_wallet_id, destination_wallet_id, category_id, amount_vnd, occurred_at, note)
		VALUES ($1, $2, $3, $4, NULLIF($5, '')::uuid, NULLIF($6, '')::uuid, $7, $8, $9)
		RETURNING id::text
	`, userID, occurrenceKey, string(txType), sourceWalletID, destinationWalletID, categoryID, amountVND, time.Date(2026, 9, 10, 2, 0, 0, 0, time.UTC), note).Scan(&id); err != nil {
		t.Fatalf("insert transaction draft: %v", err)
	}
	return id
}

func setPlanningWalletBalance(t *testing.T, conn *sql.DB, walletID string, amountVND int64) {
	t.Helper()
	if _, err := conn.ExecContext(context.Background(), `UPDATE wallets SET balance_vnd = $2 WHERE id = $1`, walletID, amountVND); err != nil {
		t.Fatalf("set wallet balance: %v", err)
	}
}

func assertPlanningCount(t *testing.T, conn *sql.DB, query string, want int, args ...any) {
	t.Helper()
	var got int
	if err := conn.QueryRowContext(context.Background(), query, args...).Scan(&got); err != nil {
		t.Fatalf("count rows: %v", err)
	}
	if got != want {
		t.Fatalf("expected %d rows, got %d", want, got)
	}
}
