import { describe, expect, it } from "vitest";

import { LEGACY_OUTBOX_KEY } from "./migrations";
import {
  archiveOfflineTransaction,
  enqueueMutation,
  initializeOfflineStore,
  markMutationsSynced,
  readOfflineSnapshot,
  saveFinanceMirror,
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
});
