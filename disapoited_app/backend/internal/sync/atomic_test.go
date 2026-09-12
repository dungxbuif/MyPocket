package sync_test

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"mypocket/internal/finance"
	"mypocket/internal/portfolio"
	mysync "mypocket/internal/sync"
)

func TestAtomicMutationRollsBackOnReceiptFailure(t *testing.T) {
	conn := migratedSyncPostgres(t)
	ctx := context.Background()
	owner := createSyncUser(t, conn, "atomic@example.com")
	repo := finance.NewRepository(conn)
	wallet := createSyncWallet(t, repo, owner)
	_, err := conn.ExecContext(ctx, `CREATE FUNCTION fail_receipt() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'injected receipt failure'; END $$; CREATE TRIGGER fail_receipt BEFORE INSERT ON sync_mutations FOR EACH ROW EXECUTE FUNCTION fail_receipt();`)
	if err != nil {
		t.Fatal(err)
	}
	service := mysync.NewService(mysync.NewRepository(conn), repo)
	mutation := transactionCreateMutation("receipt-fail", 1, 100)
	mutation.Payload = rawJSON(`{"type":"expense","source_wallet_id":"` + wallet.ID + `","category_id":"` + findSyncSystemCategory(t, conn, "expense_food") + `","amount_vnd":100,"occurred_at":"2026-09-10T00:00:00Z"}`)
	var before int
	if err := conn.QueryRow(`SELECT count(*) FROM sync_changes WHERE user_id=$1`, owner).Scan(&before); err != nil {
		t.Fatal(err)
	}
	if _, err := service.ApplyMutations(ctx, owner, []mysync.Mutation{mutation}); err == nil {
		t.Fatal("expected injected failure")
	}
	assertSyncWalletBalance(t, conn, wallet.ID, 0)
	var rows, changes int
	if err := conn.QueryRow(`SELECT count(*) FROM transactions WHERE user_id=$1`, owner).Scan(&rows); err != nil {
		t.Fatal(err)
	}
	if err := conn.QueryRow(`SELECT count(*) FROM sync_changes WHERE user_id=$1`, owner).Scan(&changes); err != nil {
		t.Fatal(err)
	}
	if rows != 0 || changes != before {
		t.Fatalf("partial commit: transactions=%d changes=%d before=%d", rows, changes, before)
	}
}

func TestChangeFailureRollsBackDirectAndSyncCommands(t *testing.T) {
	conn := migratedSyncPostgres(t)
	ctx := context.Background()
	owner := createSyncUser(t, conn, "feed-failure@example.com")
	repo := finance.NewRepository(conn)
	if _, err := conn.Exec(`CREATE FUNCTION fail_change() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'injected change failure'; END $$; CREATE TRIGGER fail_change BEFORE INSERT ON sync_changes FOR EACH ROW EXECUTE FUNCTION fail_change();`); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.CreateWallet(ctx, owner, finance.CreateWalletInput{Name: "Fail", Type: finance.WalletBasic}); err == nil {
		t.Fatal("expected change failure")
	}
	var n int
	if err := conn.QueryRow(`SELECT count(*) FROM wallets WHERE user_id=$1`, owner).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("direct command escaped rollback: %d", n)
	}
	service := mysync.NewService(mysync.NewRepository(conn), repo)
	mutation := mysync.Mutation{MutationID: "fail-change", DeviceID: "d", Sequence: 1, EntityType: mysync.EntityWallet, EntityID: fixedWalletID, Operation: mysync.OperationCreate, Payload: rawJSON(`{"name":"Fail","type":"basic"}`)}
	if _, err := service.ApplyMutations(ctx, owner, []mysync.Mutation{mutation}); err == nil {
		t.Fatal("expected sync change failure")
	}
	if err := conn.QueryRow(`SELECT count(*) FROM sync_mutations WHERE user_id=$1`, owner).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatal("failure stored success receipt")
	}
	if _, err := conn.Exec(`DROP TRIGGER fail_change ON sync_changes`); err != nil {
		t.Fatal(err)
	}
	result, err := service.ApplyMutations(ctx, owner, []mysync.Mutation{mutation})
	if err != nil || result[0].State != mysync.ResultApplied {
		t.Fatalf("safe retry: %v %v", result, err)
	}
}

func TestDirectTransferPublishesBothWalletsAndTombstone(t *testing.T) {
	conn := migratedSyncPostgres(t)
	ctx := context.Background()
	owner := createSyncUser(t, conn, "transfer-feed@example.com")
	repo := finance.NewRepository(conn)
	source := createSyncWallet(t, repo, owner)
	dest := createSyncWallet(t, repo, owner)
	value, err := repo.CreateTransaction(ctx, owner, finance.CreateTransactionInput{IdempotencyKey: "transfer", Type: finance.TransactionTransfer, SourceWalletID: source.ID, DestinationWalletID: dest.ID, AmountVND: 100, OccurredAt: fixedTime()})
	if err != nil {
		t.Fatal(err)
	}
	if err = repo.ArchiveTransaction(ctx, owner, value.ID, value.Version); err != nil {
		t.Fatal(err)
	}
	changes, err := mysync.NewRepository(conn).ListChanges(ctx, owner, 2, 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes.Changes) != 6 {
		t.Fatalf("expected 3 create effects plus 3 reversal effects: %d", len(changes.Changes))
	}
	for _, change := range changes.Changes[:2] {
		var wallet finance.Wallet
		if err := json.Unmarshal(change.Payload, &wallet); err != nil {
			t.Fatal(err)
		}
		want := int64(100)
		if wallet.ID == source.ID {
			want = -100
		}
		if wallet.BalanceVND != want {
			t.Fatalf("incorrect wallet payload: %+v", wallet)
		}
	}
	last := changes.Changes[5]
	if last.EntityID != value.ID || last.Operation != mysync.OperationArchive || last.Version != 2 {
		t.Fatalf("missing tombstone: %+v", last)
	}
	assertSyncWalletBalance(t, conn, source.ID, 0)
	assertSyncWalletBalance(t, conn, dest.ID, 0)
}

func TestPortfolioFeedAndResyncUseCanonicalPosition(t *testing.T) {
	conn := migratedSyncPostgres(t)
	ctx := context.Background()
	owner := createSyncUser(t, conn, "asset-feed@example.com")
	repo := portfolio.NewRepository(conn)
	asset, err := repo.CreatePosition(ctx, owner, portfolio.CreatePositionInput{Type: portfolio.AssetGold, Name: "Gold", Unit: "tael", PricingMode: portfolio.PricingManual})
	if err != nil {
		t.Fatal(err)
	}
	_, err = repo.AddTrade(ctx, owner, asset.ID, portfolio.AddTradeInput{Side: portfolio.TradeBuy, Quantity: "2", UnitPriceVND: 100, OccurredAt: time.Now(), BaseVersion: asset.Version})
	if err != nil {
		t.Fatal(err)
	}
	changes, err := mysync.NewRepository(conn).ListChanges(ctx, owner, 0, 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes.Changes) != 2 {
		t.Fatalf("asset feed length %d", len(changes.Changes))
	}
	var payload portfolio.Position
	if err = json.Unmarshal(changes.Changes[1].Payload, &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Trades) != 1 || payload.Version != 2 {
		t.Fatalf("incomplete position payload: %+v", payload)
	}
	snapshot, err := mysync.NewService(mysync.NewRepository(conn), finance.NewRepository(conn), repo).Resync(ctx, owner)
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Assets) != 1 || snapshot.Assets[0].Version != 2 || snapshot.NextCursor != 2 {
		t.Fatalf("snapshot: %+v", snapshot)
	}
}

func TestDirectWalletMutationPublishesChange(t *testing.T) {
	conn := migratedSyncPostgres(t)
	owner := createSyncUser(t, conn, "direct-feed@example.com")
	wallet := createSyncWallet(t, finance.NewRepository(conn), owner)
	changes, err := mysync.NewRepository(conn).ListChanges(context.Background(), owner, 0, 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes.Changes) != 1 || changes.Changes[0].EntityID != wallet.ID {
		t.Fatalf("missing direct wallet change: %#v", changes)
	}
}

func TestConcurrentMutationReplaysExactlyOnce(t *testing.T) {
	conn := migratedSyncPostgres(t)
	ctx := context.Background()
	owner := createSyncUser(t, conn, "concurrent-sync@example.com")
	service := mysync.NewService(mysync.NewRepository(conn), finance.NewRepository(conn))
	mutation := mysync.Mutation{MutationID: "same", DeviceID: "device", Sequence: 1, EntityType: mysync.EntityWallet, EntityID: fixedWalletID, Operation: mysync.OperationCreate, Payload: rawJSON(`{"name":"Cash","type":"basic"}`)}
	var wg sync.WaitGroup
	results := make(chan mysync.ResultState, 2)
	errs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, err := service.ApplyMutations(ctx, owner, []mysync.Mutation{mutation})
			if err != nil {
				errs <- err
				return
			}
			results <- r[0].State
		}()
	}
	wg.Wait()
	close(results)
	close(errs)
	for err := range errs {
		t.Error(err)
	}
	counts := map[mysync.ResultState]int{}
	for state := range results {
		counts[state]++
	}
	if counts[mysync.ResultApplied] != 1 || counts[mysync.ResultReplayed] != 1 {
		t.Fatalf("states: %v", counts)
	}
}

func TestSyncCategoryParentAndDefaultWalletPayload(t *testing.T) {
	conn := migratedSyncPostgres(t)
	ctx := context.Background()
	owner := createSyncUser(t, conn, "sync-shapes@example.com")
	repo := finance.NewRepository(conn)
	wallet := createSyncWallet(t, repo, owner)
	parent, err := repo.CreateCategory(ctx, owner, finance.CreateCategoryInput{Kind: finance.CategoryExpense, Name: "Food"})
	if err != nil {
		t.Fatal(err)
	}
	service := mysync.NewService(mysync.NewRepository(conn), repo)
	mutation := mysync.Mutation{MutationID: "parent-create", DeviceID: "d", Sequence: 1, EntityType: mysync.EntityCategory, EntityID: "00000000-0000-4000-8000-000000008888", Operation: mysync.OperationCreate, Payload: rawJSON(`{"name":"Coffee","kind":"expense","parent_id":"` + parent.ID + `"}`)}
	result, err := service.ApplyMutations(ctx, owner, []mysync.Mutation{mutation})
	if err != nil {
		t.Fatal(err)
	}
	var category finance.Category
	if err = json.Unmarshal(result[0].Payload, &category); err != nil {
		t.Fatal(err)
	}
	if category.ParentID != parent.ID {
		t.Fatalf("parent dropped: %+v", category)
	}
	mutation.MutationID = "parent-clear"
	mutation.Sequence++
	mutation.Operation = mysync.OperationUpdate
	mutation.BaseVersion = category.Version
	mutation.Payload = rawJSON(`{"name":"Coffee","parent_id":null}`)
	result, err = service.ApplyMutations(ctx, owner, []mysync.Mutation{mutation})
	if err != nil {
		t.Fatal(err)
	}
	category = finance.Category{}
	if err = json.Unmarshal(result[0].Payload, &category); err != nil {
		t.Fatal(err)
	}
	if category.ParentID != "" {
		t.Fatal("explicit null did not clear parent")
	}
	mutation = mysync.Mutation{MutationID: "default", DeviceID: "d", Sequence: 3, EntityType: mysync.EntityWallet, EntityID: wallet.ID, Operation: mysync.OperationSetDefaultAI, BaseVersion: wallet.Version, Payload: rawJSON(`{}`)}
	result, err = service.ApplyMutations(ctx, owner, []mysync.Mutation{mutation})
	if err != nil {
		t.Fatal(err)
	}
	var updated finance.Wallet
	if err = json.Unmarshal(result[0].Payload, &updated); err != nil {
		t.Fatal(err)
	}
	if updated.ID != wallet.ID || !updated.IsDefaultAI || updated.Version != 2 {
		t.Fatalf("invalid default payload: %+v", updated)
	}
	mutation.MutationID = "default-noop"
	mutation.Sequence++
	mutation.BaseVersion = 2
	result, err = service.ApplyMutations(ctx, owner, []mysync.Mutation{mutation})
	if err != nil {
		t.Fatal(err)
	}
	if result[0].Version != 2 {
		t.Fatalf("invented version on no-op: %+v", result[0])
	}
}

func TestSyncRejectionThenValidMutationAndHashReuse(t *testing.T) {
	conn := migratedSyncPostgres(t)
	ctx := context.Background()
	owner := createSyncUser(t, conn, "sync-reject@example.com")
	service := mysync.NewService(mysync.NewRepository(conn), finance.NewRepository(conn))
	bad := mysync.Mutation{MutationID: "bad", DeviceID: "d", Sequence: 1, EntityType: mysync.EntityWallet, EntityID: fixedWalletID, Operation: mysync.OperationCreate, Payload: rawJSON(`{"name":"","type":"basic"}`)}
	good := bad
	good.MutationID = "good"
	good.Sequence = 2
	good.Payload = rawJSON(`{"name":"Cash","type":"basic"}`)
	result, err := service.ApplyMutations(ctx, owner, []mysync.Mutation{bad, good})
	if err != nil {
		t.Fatal(err)
	}
	if result[0].State != mysync.ResultRejected || result[1].State != mysync.ResultApplied {
		t.Fatalf("results: %v", result)
	}
	good.Payload = rawJSON(`{"name":"Changed","type":"basic"}`)
	result, err = service.ApplyMutations(ctx, owner, []mysync.Mutation{good})
	if err != nil {
		t.Fatal(err)
	}
	if result[0].State != mysync.ResultRejected {
		t.Fatal("hash reuse accepted")
	}
}
