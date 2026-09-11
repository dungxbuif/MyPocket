import { describe, expect, it, vi } from "vitest";

import { readOutbox } from "../app/outbox";
import {
  enqueueMutation,
  readPendingReceiptUploads,
  readOfflineSnapshot,
  readServerEpoch,
  queueReceiptUpload,
  saveFinanceMirror,
  saveOfflineConflict,
} from "./db";
import {
  discardLocalConflict,
  editAndRetryTransactionConflict,
  fullResync,
  reconcileServerEpoch,
  keepServerConflict,
  listOpenConflicts,
} from "./conflicts";
import type { OfflineConflict } from "./types";

describe("offline conflict recovery", () => {
  it("keeps the server transaction and removes the conflicting local mutation", async () => {
    await saveFinanceMirror({ transactions: [transaction("tx_1", "Local note", 1)] });
    const conflict = await createTransactionConflict("tx_1", "Local note", "Server note");

    await keepServerConflict(conflict);

    const snapshot = await readOfflineSnapshot();
    expect(snapshot.transactions[0].note).toBe("Server note");
    expect(await listOpenConflicts()).toHaveLength(0);
    expect(await readOutbox()).toHaveLength(0);
  });

  it("discards the local intent without applying it to the mirror", async () => {
    await saveFinanceMirror({ transactions: [transaction("tx_1", "Local note", 1)] });
    const conflict = await createTransactionConflict("tx_1", "Local note", "Server note");

    await discardLocalConflict(conflict);

    const snapshot = await readOfflineSnapshot();
    expect(snapshot.transactions[0].note).toBe("Server note");
    expect(await listOpenConflicts()).toHaveLength(0);
    expect(await readOutbox()).toHaveLength(0);
  });

  it("creates a revised mutation against the server version for edit and retry", async () => {
    const conflict = await createTransactionConflict("tx_1", "Local note", "Server note");

    await editAndRetryTransactionConflict(conflict, { amount_vnd: 55000, note: "Retry note" });

    const outbox = await readOutbox();
    expect(await listOpenConflicts()).toHaveLength(0);
    expect(outbox).toHaveLength(1);
    expect(outbox[0].input).toMatchObject({ note: "Retry note", amount_vnd: 55000 });
  });

  it("full resync replaces authoritative mirrors while preserving unsent mutations", async () => {
    await enqueueMutation({ entity_type: "transaction", entity_id: "tx_pending", operation: "create", base_version: 0, payload: { note: "Pending" } });
    vi.stubGlobal("fetch", vi.fn(async () => new Response(JSON.stringify({
      status: "ok",
      correlation_id: "req_test",
      snapshot: {
        wallets: [{ id: "wallet_1", name: "Ví server", type: "cash", balance_vnd: 1000, include_in_total: true, is_default_ai: true, version: 1 }],
        categories: [{ id: "cat_1", kind: "expense", name: "Ăn uống", is_system: true, version: 1 }],
        transactions: [transaction("tx_server", "Server only", 3)],
        server_epoch: "atomic-sync-v1",
        next_cursor: 12,
      },
    }), { status: 200, headers: { "Content-Type": "application/json" } })));

    await fullResync();

    const snapshot = await readOfflineSnapshot();
    expect(snapshot.wallets[0].name).toBe("Ví server");
    expect(snapshot.transactions.map((item) => item.note)).toEqual(["Server only"]);
    expect(snapshot.cursor).toBe(12);
    expect(await readOutbox()).toHaveLength(1);
  });

  it("performs one post-0012 full resync without deleting pending intent", async () => {
    await enqueueMutation({ entity_type: "transaction", entity_id: "tx_pending", operation: "create", base_version: 0, payload: { note: "Pending" } });
    const pending = await createTransactionConflict("tx_conflict", "Local", "Server");
    await queueReceiptUpload({ transaction_id: "tx_pending", file: new Blob(["receipt"]), filename: "bill.jpg", content_type: "image/jpeg" });
    const fetchMock = vi.fn(async () => new Response(JSON.stringify({
      status: "ok",
      correlation_id: "req_epoch",
      snapshot: {
        wallets: [{ id: "wallet_epoch", name: "Ví server", type: "cash", balance_vnd: 1000, include_in_total: true, is_default_ai: true, version: 1 }],
        categories: [],
        transactions: [],
        assets: [],
        server_epoch: "atomic-sync-v1",
        next_cursor: 21,
      },
    }), { status: 200, headers: { "Content-Type": "application/json" } }));
    vi.stubGlobal("fetch", fetchMock);

    expect(await reconcileServerEpoch("atomic-sync-v1")).toBe(true);
    expect(await reconcileServerEpoch("atomic-sync-v1")).toBe(false);

    const snapshot = await readOfflineSnapshot();
    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(snapshot.wallets.map((wallet) => wallet.id)).toEqual(["wallet_epoch"]);
    expect(snapshot.cursor).toBe(21);
    expect(await readServerEpoch()).toBe("atomic-sync-v1");
    expect(await readOutbox()).toHaveLength(2);
    expect(snapshot.conflicts).toContainEqual(pending);
    expect(await readPendingReceiptUploads()).toHaveLength(1);
  });
});

async function createTransactionConflict(entityID: string, localNote: string, serverNote: string) {
  const mutation = await enqueueMutation({ entity_type: "transaction", entity_id: entityID, operation: "update", base_version: 1, payload: { note: localNote, amount_vnd: 44000 } });
  const conflict: OfflineConflict = {
    conflict_id: mutation.mutation_id,
    mutation_id: mutation.mutation_id,
    entity_type: "transaction",
    entity_id: entityID,
    operation: "update",
    base_version: 1,
    server_version: 2,
    local_payload: { note: localNote, amount_vnd: 44000 },
    server_payload: transaction(entityID, serverNote, 2),
    status: "open",
    created_at: "2026-08-31T00:00:00Z",
  };
  await saveOfflineConflict(conflict);
  return conflict;
}

function transaction(id: string, note: string, version: number) {
  return {
    id,
    type: "expense" as const,
    source_wallet_id: "wallet_1",
    category_id: "cat_1",
    amount_vnd: 44000,
    balance_after_vnd: 56000,
    occurred_at: "2026-08-31T00:00:00Z",
    note,
    with_person: "",
    event_ref: "",
    excluded_from_reports: false,
    version,
  };
}
