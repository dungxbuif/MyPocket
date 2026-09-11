import { describe, expect, it } from "vitest";

import { LEGACY_OUTBOX_KEY } from "./migrations";
import {
  archiveOfflineTransaction,
  enqueueMutation,
  initializeOfflineStore,
  initializeOfflineStoreForUser,
  listOpenConflicts,
  markMutationsSynced,
  readOfflineSnapshot,
  saveFinanceMirror,
  saveOfflineConflict,
  upsertOfflineTransaction,
} from "./db";

describe("offline IndexedDB store", () => {
  it("migrates legacy localStorage queued transactions into the IndexedDB outbox", async () => {
    localStorage.setItem(LEGACY_OUTBOX_KEY, JSON.stringify([
      { id: "legacy_1", created_at: "2026-08-31T00:00:00Z", input: { type: "expense", source_wallet_id: "wallet_1", amount_vnd: 42000, occurred_at: "2026-08-31T00:00:00Z", note: "Legacy" } },
    ]));

    const snapshot = await initializeOfflineStore();

    expect(snapshot.mode).toBe("ready");
    expect(snapshot.outbox).toHaveLength(1);
    expect(snapshot.outbox[0]).toMatchObject({ entity_type: "transaction", entity_id: "legacy_1", operation: "create" });
    expect(snapshot.outbox[0].payload.note).toBe("Legacy");
    expect(localStorage.getItem(LEGACY_OUTBOX_KEY)).toBeNull();
  });

  it("quarantines invalid legacy localStorage outbox entries and continues sequence after migration", async () => {
    localStorage.setItem(LEGACY_OUTBOX_KEY, JSON.stringify([
      { id: "legacy_1", created_at: "2026-08-31T00:00:00Z", input: { type: "expense", source_wallet_id: "wallet_1", amount_vnd: 42000, occurred_at: "2026-08-31T00:00:00Z" } },
      { id: "bad_1", input: { amount_vnd: "42000" } },
    ]));

    const migrated = await initializeOfflineStore();
    const next = await enqueueMutation({ entity_type: "transaction", entity_id: "tx_2", operation: "create", base_version: 0, payload: {} });

    expect(migrated.outbox[0].sequence).toBe(1);
    expect(next.sequence).toBe(2);
    expect(localStorage.getItem("mypocket.transaction-outbox.quarantine.v1")).toContain("bad_1");
  });

  it("keeps mirrored finance data available after transaction archive tombstones", async () => {
    await saveFinanceMirror({
      wallets: [{ id: "wallet_1", name: "Cash", type: "cash", balance_vnd: 100000, include_in_total: true, is_default_ai: true, version: 2 }],
      categories: [{ id: "cat_1", kind: "expense", name: "Food", is_system: false, version: 1 }],
      transactions: [{ id: "tx_1", type: "expense", source_wallet_id: "wallet_1", amount_vnd: 10000, balance_after_vnd: 90000, occurred_at: "2026-08-31T00:00:00Z", note: "Lunch", with_person: "", event_ref: "", excluded_from_reports: false, version: 3 }],
    });

    await archiveOfflineTransaction("tx_1", 3);
    const snapshot = await readOfflineSnapshot();

    expect(snapshot.wallets[0].name).toBe("Cash");
    expect(snapshot.categories[0].name).toBe("Food");
    expect(snapshot.transactions).toEqual([]);
    expect(snapshot.tombstones[0]).toMatchObject({ entity_type: "transaction", entity_id: "tx_1", version: 3 });
  });

  it("isolates snapshots and preserves pending work when switching users", async () => {
    await initializeOfflineStoreForUser("user-a");
    await saveFinanceMirror({
      userID: "user-a",
      wallets: [{ id: "wallet-a", name: "A", type: "cash", balance_vnd: 100000, include_in_total: true, is_default_ai: true, version: 1 }],
    });
    await enqueueMutation({ entity_type: "wallet", entity_id: "pending-a", operation: "create", base_version: 0, payload: { name: "Pending A" } });

    const snapshot = await initializeOfflineStoreForUser("user-b");

    expect(snapshot.mode).toBe("ready");
    expect(snapshot.wallets).toEqual([]);
    expect(snapshot.outbox).toEqual([]);
    const restored = await initializeOfflineStoreForUser("user-a");
    expect(restored.wallets[0]?.id).toBe("wallet-a");
    expect(restored.outbox).toHaveLength(1);
  });

  it("keeps outbox sequence monotonic after synced mutations are removed", async () => {
    const first = await enqueueMutation({ entity_type: "transaction", entity_id: "tx_1", operation: "create", base_version: 0, payload: {} });
    await markMutationsSynced([first.mutation_id]);
    const second = await enqueueMutation({ entity_type: "transaction", entity_id: "tx_2", operation: "create", base_version: 0, payload: {} });

    expect(second.sequence).toBeGreaterThan(first.sequence);
  });

  it("stores optimistic transaction updates with the caller base version", async () => {
    await upsertOfflineTransaction({ id: "tx_1", type: "expense", source_wallet_id: "wallet_1", amount_vnd: 10000, balance_after_vnd: 90000, occurred_at: "2026-08-31T00:00:00Z", note: "Lunch", with_person: "", event_ref: "", excluded_from_reports: false, version: 4 });
    await enqueueMutation({ entity_type: "transaction", entity_id: "tx_1", operation: "update", base_version: 4, payload: { note: "Dinner" } });

    const snapshot = await readOfflineSnapshot();

    expect(snapshot.outbox[0]).toMatchObject({ entity_id: "tx_1", base_version: 4 });
  });

  it("stores open conflicts without removing recoverable outbox mutations", async () => {
    const mutation = await enqueueMutation({ entity_type: "transaction", entity_id: "tx_1", operation: "update", base_version: 1, payload: { note: "Offline" } });
    await saveOfflineConflict({
      conflict_id: mutation.mutation_id,
      mutation_id: mutation.mutation_id,
      entity_type: "transaction",
      entity_id: "tx_1",
      operation: "update",
      base_version: 1,
      server_version: 2,
      local_payload: { note: "Offline" },
      server_payload: { id: "tx_1", note: "Server", version: 2 },
      status: "open",
      created_at: "2026-08-31T00:00:00Z",
    });

    const conflicts = await listOpenConflicts();
    const snapshot = await readOfflineSnapshot();

    expect(conflicts).toHaveLength(1);
    expect(conflicts[0]).toMatchObject({ mutation_id: mutation.mutation_id, server_version: 2 });
    expect(snapshot.outbox).toHaveLength(1);
  });
});
