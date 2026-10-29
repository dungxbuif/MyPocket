package httpapi

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
)

// A scope/confirmation gate added only to MCP would break these direct writes.
// Readback uses both transports and real repositories, not dispatcher doubles.
func TestMCPDirectOrdinaryWriteFamiliesJWTAndAPIKey(t *testing.T) {
	f := newMCPParityFixture(t)
	for _, credential := range []struct{ name, token string }{{"JWT", f.jwt}, {"APIKey", f.key.Secret}} {
		session := f.session(t, credential.token)
		budgetWallet := f.wallet
		if credential.name == "APIKey" {
			budgetWallet = f.destination
		}
		for _, tc := range []struct {
			tool, listTool, path string
			args                 map[string]any
			listArgs             map[string]any
		}{
			{"wallets.create", "wallets.list", "/api/v1/wallets", map[string]any{"name": "MCP wallet " + credential.name, "type": "basic", "opening_balance": 1000}, map[string]any{}},
			{"categories.create", "categories.list", "/api/v1/categories", map[string]any{"name": "MCP category " + credential.name, "kind": "expense", "icon_key": "tag"}, map[string]any{}},
			{"budgets.create", "budgets.list", "/api/v1/budgets", map[string]any{"name": "MCP budget " + credential.name, "limit_amount": 9000, "wallet_id": budgetWallet, "start_date": "2026-10-01", "end_date": "2026-10-31"}, map[string]any{}},
			{"jars.create", "jars.list", "/api/v1/jars?month=2026-10", map[string]any{"month": "2026-10", "name": "MCP jar " + credential.name, "allocation_mode": "fixed", "allocation_amount": 500}, map[string]any{"month": "2026-10"}},
			{"recurring.create", "recurring.list", "/api/v1/recurring", map[string]any{"name": "MCP recurring " + credential.name, "wallet_id": f.wallet, "category_id": f.category, "type": "expense", "amount": 100, "frequency": "monthly", "interval": 1, "next_run_at": "2026-11-01T03:00:00Z"}, map[string]any{}},
			{"travel.create", "travel.list", "/api/v1/travel/events", map[string]any{"name": "MCP trip " + credential.name, "active": false}, map[string]any{}},
			{"portfolio.asset.create", "portfolio.asset.list", "/api/v1/portfolio/assets", map[string]any{"symbol": "MCP" + credential.name, "name": "MCP asset " + credential.name}, map[string]any{}},
			{"feedback.create", "feedback.list", "/api/v1/feedback", map[string]any{"type": "improvement", "title": "MCP feedback " + credential.name, "description": "Synthetic direct owner write"}, map[string]any{}},
		} {
			t.Run(credential.name+"/"+tc.tool, func(t *testing.T) {
				result, response := mcpParityCall(t, session, tc.tool, tc.args, "")
				requireMCPParityStatus(t, result, response, 201)
				created := mcpParityData[map[string]any](t, response)
				id, _ := created["id"].(string)
				if id == "" {
					id, _ = created["jar_id"].(string)
				}
				if id == "" {
					t.Fatalf("write did not return a persisted ID: %s", response.Body)
				}
				result, read := mcpParityCall(t, session, tc.listTool, tc.listArgs, "")
				requireMCPParityStatus(t, result, read, 200)
				if !bytes.Contains(read.Body, []byte(id)) {
					t.Fatalf("MCP readback missing %s", id)
				}
				rest := f.rest(t, "GET", tc.path, nil, credential.token, "")
				if rest.Code != 200 || !bytes.Contains(rest.Body.Bytes(), []byte(id)) {
					t.Fatalf("REST readback=%d body=%s", rest.Code, rest.Body.String())
				}
				requireMCPParityEqualReadback(t, read.Body, rest.Body.Bytes())
			})
		}
	}
}

func TestMCPWriteIdempotencyReplayAcrossRESTJWTAndAPIKey(t *testing.T) {
	f := newMCPParityFixture(t)
	key := uuid.NewString()
	args := f.transactionArgs(12500)
	result, first := mcpParityCall(t, f.session(t, f.jwt), "transactions.create", args, key)
	requireMCPParityStatus(t, result, first, 201)
	created := mcpParityData[entity.Transaction](t, first)
	if created.ID == "" || created.OwnerID != f.owner {
		t.Fatal("create did not persist authenticated owner")
	}
	result, replay := mcpParityCall(t, f.session(t, f.key.Secret), "transactions.create", args, key)
	requireMCPParityStatus(t, result, replay, 200)
	if got := mcpParityData[entity.Transaction](t, replay); got.ID != created.ID {
		t.Fatal("cross-credential retry created another transaction")
	}
	var envelope offlineHTTPEnvelope
	if err := json.Unmarshal(replay.Body, &envelope); err != nil {
		t.Fatal(err)
	}
	if !envelope.Meta.Mutation.Replayed || envelope.Meta.Mutation.Key != key {
		t.Fatal("MCP lost canonical receipt metadata")
	}
	rest := f.rest(t, "POST", "/api/v1/transactions", args, f.jwt, key)
	if rest.Code != 200 || !decodeOfflineHTTP(t, rest).Meta.Mutation.Replayed {
		t.Fatalf("REST retry did not share MCP receipt: %d", rest.Code)
	}
	changed := f.transactionArgs(99999)
	result, conflict := mcpParityCall(t, f.session(t, f.key.Secret), "transactions.create", changed, key)
	requireMCPParityStatus(t, result, conflict, 409)
	if response := f.rest(t, "POST", "/api/v1/transactions", changed, f.jwt, key); response.Code != 409 {
		t.Fatalf("REST changed-payload status=%d", response.Code)
	}
	var rows []entity.Transaction
	if err := f.db.Where("owner_id = ?", f.owner).Find(&rows).Error; err != nil || len(rows) != 1 || rows[0].Amount != 12500 {
		t.Fatalf("replay/conflict changed ledger: rows=%+v err=%v", rows, err)
	}
	result, foreign := mcpParityCall(t, f.session(t, f.foreignKey.Secret), "transactions.create", args, key)
	requireMCPParityStatus(t, result, foreign, 404)
	if bytes.Contains(foreign.Body, []byte(created.ID)) {
		t.Fatal("foreign owner saw receipt")
	}
}

func TestMCPWriteDurableCreateReceiptsReplayRESTResults(t *testing.T) {
	f := newMCPParityFixture(t)
	session := f.session(t, f.key.Secret)
	for _, tc := range []struct {
		tool, path string
		args       map[string]any
	}{
		{"wallets.create", "/api/v1/wallets", map[string]any{"name": "Receipt wallet", "type": "basic"}},
		{"categories.create", "/api/v1/categories", map[string]any{"name": "Receipt category", "kind": "expense", "icon_key": "tag"}},
		{"budgets.create", "/api/v1/budgets", map[string]any{"name": "Receipt budget", "limit_amount": 9000, "start_date": "2026-10-01", "end_date": "2026-10-31"}},
		{"jars.create", "/api/v1/jars", map[string]any{"month": "2026-10", "name": "Receipt jar", "allocation_mode": "fixed", "allocation_amount": 500}},
	} {
		t.Run(tc.tool, func(t *testing.T) {
			key := uuid.NewString()
			first := f.rest(t, "POST", tc.path, tc.args, f.jwt, key)
			if first.Code != 201 {
				t.Fatalf("canonical keyed create=%d %s", first.Code, first.Body.String())
			}
			var original struct {
				Data json.RawMessage `json:"data"`
			}
			if err := json.Unmarshal(first.Body.Bytes(), &original); err != nil {
				t.Fatal(err)
			}
			result, replay := mcpParityCall(t, session, tc.tool, tc.args, key)
			requireMCPParityStatus(t, result, replay, 200)
			var actual struct {
				Data json.RawMessage `json:"data"`
			}
			if err := json.Unmarshal(replay.Body, &actual); err != nil {
				t.Fatal(err)
			}
			var expectedData, actualData any
			if err := json.Unmarshal(original.Data, &expectedData); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(actual.Data, &actualData); err != nil {
				t.Fatal(err)
			}
			expectedJSON, _ := json.Marshal(expectedData)
			actualJSON, _ := json.Marshal(actualData)
			if !bytes.Equal(expectedJSON, actualJSON) {
				t.Fatalf("MCP replay replaced original persisted result: want=%s got=%s", expectedJSON, actualJSON)
			}
			changed := map[string]any{}
			for k, v := range tc.args {
				changed[k] = v
			}
			changed["name"] = "Changed payload"
			result, conflict := mcpParityCall(t, session, tc.tool, changed, key)
			requireMCPParityStatus(t, result, conflict, 409)
		})
	}
}

func TestMCPWriteStaleCategoryCASDoesNotMutate(t *testing.T) {
	f := newMCPParityFixture(t)
	session := f.session(t, f.key.Secret)
	args := map[string]any{"id": f.category, "name": "Current name", "kind": "expense", "icon_key": "tag", "expected_version": 1}
	result, response := mcpParityCall(t, session, "categories.update", args, "")
	requireMCPParityStatus(t, result, response, 200)
	updated := mcpParityData[entity.Category](t, response)
	if updated.Version != 2 || updated.Name != "Current name" {
		t.Fatalf("category update=%+v", updated)
	}
	args["name"] = "Stale name"
	result, response = mcpParityCall(t, session, "categories.update", args, "")
	requireMCPParityStatus(t, result, response, 409)
	var stored entity.Category
	if err := f.db.First(&stored, "id = ? AND owner_id = ?", f.category, f.owner).Error; err != nil || stored.Name != "Current name" || stored.Version != 2 {
		t.Fatalf("stale category mutation: %+v err=%v", stored, err)
	}
}

// This proves observable pair state and validation rejection without partial
// rows. Storage-failure rollback between SQL statements belongs to repository
// coverage; no unstable failure-injection seam is invented at the MCP boundary.
func TestMCPWriteTransferPairAtomicityAndReplay(t *testing.T) {
	f := newMCPParityFixture(t)
	session := f.session(t, f.key.Secret)
	args := map[string]any{"source_wallet_id": f.wallet, "destination_wallet_id": f.destination, "amount": 700, "occurred_at": "2026-10-20T03:00:00Z"}
	key := uuid.NewString()
	result, first := mcpParityCall(t, session, "transactions.transfer.create", args, key)
	requireMCPParityStatus(t, result, first, 201)
	pair := mcpParityData[[]entity.Transaction](t, first)
	if len(pair) != 2 || pair[0].TransferID == nil || pair[1].TransferID == nil || *pair[0].TransferID != *pair[1].TransferID {
		t.Fatalf("not one transfer pair: %+v", pair)
	}
	for _, row := range pair {
		if row.Amount != 700 || row.IncludedInReports || row.OwnerID != f.owner {
			t.Fatalf("invalid transfer row: %+v", row)
		}
	}
	result, replay := mcpParityCall(t, f.session(t, f.jwt), "transactions.transfer.create", args, key)
	requireMCPParityStatus(t, result, replay, 200)
	replayed := mcpParityData[[]entity.Transaction](t, replay)
	if len(replayed) != 2 || replayed[0].ID != pair[0].ID || replayed[1].ID != pair[1].ID {
		t.Fatal("transfer retry replaced or duplicated pair")
	}
	invalid := map[string]any{"source_wallet_id": f.wallet, "destination_wallet_id": f.foreign.wallet, "amount": 700, "occurred_at": "2026-10-20T03:00:00Z"}
	result, denied := mcpParityCall(t, session, "transactions.transfer.create", invalid, uuid.NewString())
	requireMCPParityStatus(t, result, denied, 404)
	var count int64
	if err := f.db.Model(&entity.Transaction{}).Where("owner_id = ?", f.owner).Count(&count).Error; err != nil || count != 2 {
		t.Fatalf("invalid transfer left partial rows: count=%d err=%v", count, err)
	}
	transferID := *pair[0].TransferID
	updateArgs := map[string]any{"transfer_id": transferID, "amount": 800, "occurred_at": "2026-10-20T03:00:00Z"}
	result, missingKey := mcpParityCall(t, session, "transactions.transfer.update", updateArgs, "")
	requireMCPParityStatus(t, result, missingKey, 400)
	updateKey := uuid.NewString()
	result, updated := mcpParityCall(t, session, "transactions.transfer.update", updateArgs, updateKey)
	requireMCPParityStatus(t, result, updated, 200)
	result, updateReplay := mcpParityCall(t, session, "transactions.transfer.update", updateArgs, updateKey)
	requireMCPParityStatus(t, result, updateReplay, 200)
	requireMCPParityEqualReadback(t, updated.Body, updateReplay.Body)
	restUpdateReplay := f.rest(t, "PATCH", "/api/v1/transactions/transfer/"+transferID, updateArgs, f.jwt, updateKey)
	if restUpdateReplay.Code != 200 || restUpdateReplay.Header().Get("Idempotent-Replayed") != "true" {
		t.Fatalf("REST transfer update replay=%d %s", restUpdateReplay.Code, restUpdateReplay.Body.String())
	}
	requireMCPParityEqualReadback(t, updated.Body, restUpdateReplay.Body.Bytes())
	changedUpdate := map[string]any{"transfer_id": transferID, "amount": 900, "occurred_at": "2026-10-20T03:00:00Z"}
	result, updateConflict := mcpParityCall(t, session, "transactions.transfer.update", changedUpdate, updateKey)
	requireMCPParityStatus(t, result, updateConflict, 409)
	var stored []entity.Transaction
	if err := f.db.Where("owner_id = ? AND transfer_id = ?", f.owner, transferID).Find(&stored).Error; err != nil || len(stored) != 2 {
		t.Fatal("updated transfer pair missing")
	}
	for _, row := range stored {
		if row.Amount != 800 {
			t.Fatal("transfer update changed only one side")
		}
	}
	deleteKey := uuid.NewString()
	result, deleted := mcpParityCall(t, session, "transactions.transfer.delete", map[string]any{"transfer_id": transferID}, deleteKey)
	requireMCPParityStatus(t, result, deleted, 204)
	result, deleteReplay := mcpParityCall(t, session, "transactions.transfer.delete", map[string]any{"transfer_id": transferID}, deleteKey)
	requireMCPParityStatus(t, result, deleteReplay, 204)
	restDeleteReplay := f.rest(t, "DELETE", "/api/v1/transactions/transfer/"+transferID, nil, f.jwt, deleteKey)
	if restDeleteReplay.Code != 204 || restDeleteReplay.Header().Get("Idempotent-Replayed") != "true" {
		t.Fatalf("REST transfer delete replay=%d %s", restDeleteReplay.Code, restDeleteReplay.Body.String())
	}
	// Delete has no body; changing the transfer ID changes its canonical hash.
	result, deleteConflict := mcpParityCall(t, session, "transactions.transfer.delete", map[string]any{"transfer_id": uuid.NewString()}, deleteKey)
	requireMCPParityStatus(t, result, deleteConflict, 409)
	if err := f.db.Model(&entity.Transaction{}).Where("owner_id = ? AND transfer_id = ?", f.owner, transferID).Count(&count).Error; err != nil || count != 0 {
		t.Fatal("transfer delete left a half pair")
	}
}

func TestMCPWriteStaleTransactionCASDoesNotMutate(t *testing.T) {
	f := newMCPParityFixture(t)
	session := f.session(t, f.key.Secret)
	result, first := mcpParityCall(t, session, "transactions.create", f.transactionArgs(100), "")
	requireMCPParityStatus(t, result, first, 201)
	row := mcpParityData[entity.Transaction](t, first)
	if row.Version != 1 {
		t.Fatalf("initial version=%d want=1", row.Version)
	}
	args := f.transactionArgs(200)
	delete(args, "auto_link_travel")
	args["id"], args["expected_version"] = row.ID, int64(1)
	result, updated := mcpParityCall(t, session, "transactions.update", args, "")
	requireMCPParityStatus(t, result, updated, 200)
	if got := mcpParityData[entity.Transaction](t, updated); got.Amount != 200 || got.Version != 2 {
		t.Fatalf("CAS result=%+v", got)
	}
	args["amount"] = int64(300)
	result, stale := mcpParityCall(t, session, "transactions.update", args, "")
	requireMCPParityStatus(t, result, stale, 409)
	result, staleDelete := mcpParityCall(t, session, "transactions.delete", map[string]any{"id": row.ID, "expected_version": 1}, "")
	requireMCPParityStatus(t, result, staleDelete, 409)
	var stored entity.Transaction
	if err := f.db.First(&stored, "id = ? AND owner_id = ?", row.ID, f.owner).Error; err != nil || stored.Amount != 200 || stored.Version != 2 {
		t.Fatalf("stale write mutated state: %+v err=%v", stored, err)
	}
	rest := f.rest(t, "GET", "/api/v1/transactions", nil, f.jwt, "")
	if rest.Code != 200 || !bytes.Contains(rest.Body.Bytes(), []byte(row.ID)) {
		t.Fatalf("REST lost CAS row: %d", rest.Code)
	}
}

func TestMCPWriteAdjustmentAndBulkDeleteUseCanonicalLedger(t *testing.T) {
	f := newMCPParityFixture(t)
	session := f.session(t, f.key.Secret)
	result, adjustment := mcpParityCall(t, session, "transactions.adjustment", map[string]any{"wallet_id": f.wallet, "amount": 250, "direction": "increase", "occurred_at": "2026-10-20T03:00:00Z"}, "")
	requireMCPParityStatus(t, result, adjustment, 201)
	adjusted := mcpParityData[entity.Transaction](t, adjustment)
	if adjusted.Type != "adjustment" || adjusted.Amount != 250 || adjusted.IncludedInReports {
		t.Fatalf("adjustment=%+v", adjusted)
	}
	var ids []string
	for _, amount := range []int64{100, 200} {
		result, created := mcpParityCall(t, session, "transactions.create", f.transactionArgs(amount), "")
		requireMCPParityStatus(t, result, created, 201)
		ids = append(ids, mcpParityData[entity.Transaction](t, created).ID)
	}
	result, deleted := mcpParityCall(t, session, "transactions.bulk_delete", map[string]any{"transaction_ids": ids}, "")
	requireMCPParityStatus(t, result, deleted, 204)
	var rows []entity.Transaction
	if err := f.db.Where("owner_id = ?", f.owner).Find(&rows).Error; err != nil || len(rows) != 1 || rows[0].ID != adjusted.ID {
		t.Fatalf("bulk-delete ledger=%+v err=%v", rows, err)
	}
	result, listed := mcpParityCall(t, session, "transactions.list", map[string]any{}, "")
	requireMCPParityStatus(t, result, listed, 200)
	for _, id := range ids {
		if bytes.Contains(listed.Body, []byte(id)) {
			t.Fatalf("deleted transaction %s survived readback", id)
		}
	}
}

func TestMCPWriteCreditEntryAndPaymentPreserveLinkedLedger(t *testing.T) {
	f := newMCPParityFixture(t)
	session := f.session(t, f.key.Secret)
	result, response := mcpParityCall(t, session, "wallets.create", map[string]any{"name": "Synthetic credit", "type": "credit", "credit_limit": 1000000}, "")
	requireMCPParityStatus(t, result, response, 201)
	wallet := mcpParityData[entity.Wallet](t, response)
	if err := f.db.Create(&entity.CategoryWallet{CategoryID: f.category, WalletID: wallet.ID}).Error; err != nil {
		t.Fatal(err)
	}
	result, response = mcpParityCall(t, session, "credit.entry.create", map[string]any{"wallet_id": wallet.ID, "kind": "purchase", "category_id": f.category, "amount": 1000, "occurred_at": "2026-10-20T03:00:00Z"}, "")
	requireMCPParityStatus(t, result, response, 201)
	entry := mcpParityData[entity.Transaction](t, response)
	if entry.CreditKind == nil || *entry.CreditKind != "purchase" || entry.Amount != 1000 || entry.OwnerID != f.owner {
		t.Fatalf("credit purchase=%+v", entry)
	}
	result, response = mcpParityCall(t, session, "credit.payment.create", map[string]any{"wallet_id": wallet.ID, "source_wallet_id": f.wallet, "amount": 600, "occurred_at": "2026-10-20T03:00:00Z"}, "")
	requireMCPParityStatus(t, result, response, 201)
	pair := mcpParityData[[]entity.Transaction](t, response)
	if len(pair) != 2 || pair[0].TransferID == nil || pair[1].TransferID == nil || *pair[0].TransferID != *pair[1].TransferID {
		t.Fatalf("credit payment pair=%+v", pair)
	}
	creditSides, sourceSides := 0, 0
	for _, row := range pair {
		if row.Amount != 600 || row.IncludedInReports || row.OwnerID != f.owner {
			t.Fatalf("payment row=%+v", row)
		}
		if row.WalletID == wallet.ID && row.Type == "income" && row.CreditKind != nil && *row.CreditKind == "payment" && row.CreditPaymentID != nil && *row.CreditPaymentID == *row.TransferID {
			creditSides++
		}
		if row.WalletID == f.wallet && row.Type == "expense" && row.CreditKind == nil && row.CreditPaymentID == nil {
			sourceSides++
		}
	}
	if creditSides != 1 || sourceSides != 1 {
		t.Fatal("payment must contain one credit repayment and one source expense")
	}
	result, response = mcpParityCall(t, session, "transactions.list", map[string]any{}, "")
	requireMCPParityStatus(t, result, response, 200)
	rest := f.rest(t, "GET", "/api/v1/transactions", nil, f.jwt, "")
	for _, id := range []string{entry.ID, pair[0].ID, pair[1].ID} {
		if !bytes.Contains(response.Body, []byte(id)) || rest.Code != 200 || !bytes.Contains(rest.Body.Bytes(), []byte(id)) {
			t.Fatalf("credit ledger %s missing from MCP/REST readback", id)
		}
	}
}
