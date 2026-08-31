import { describe, expect, it, vi } from "vitest";
import { queueCategoryCreate, queueWalletCreate } from "../offline/outbox";
import { readOfflineSnapshot } from "../offline/db";
import { drainOutbox, queueTransaction, readOutbox } from "./outbox";

describe("transaction outbox", () => {
  it("stores an optimistic transaction for offline replay", async () => {
    const transaction = await queueTransaction({ type: "expense", source_wallet_id: "wallet_1", amount_vnd: 42000, occurred_at: "2026-08-31T00:00:00Z", note: "Offline coffee" });
    expect(transaction.version).toBe(0);
    const outbox = await readOutbox();
    expect(outbox).toHaveLength(1);
    expect(outbox[0].input.note).toBe("Offline coffee");
  });

  it("drains queued mutations in order", async () => {
    await queueTransaction({ type: "expense", source_wallet_id: "wallet_1", amount_vnd: 1, occurred_at: "2026-08-31T00:00:00Z" });
    await queueTransaction({ type: "expense", source_wallet_id: "wallet_1", amount_vnd: 2, occurred_at: "2026-08-31T00:00:00Z" });
    const amounts: number[] = [];
    expect(await drainOutbox(async (input) => { amounts.push(Number(input.amount_vnd)); })).toBe(2);
    expect(amounts).toEqual([1, 2]);
    expect(await readOutbox()).toHaveLength(0);
  });

  it("queues wallet and category mutations with durable ordering", async () => {
    const wallet = await queueWalletCreate({ name: "Ví offline", type: "cash" });
    const category = await queueCategoryCreate({ name: "Ăn offline", kind: "expense" });
    const outbox = await readOutbox();

    expect(wallet.name).toBe("Ví offline");
    expect(category.name).toBe("Ăn offline");
    expect(outbox.map((item) => item.input.name)).toEqual(["Ví offline", "Ăn offline"]);
  });

  it("drains queued mutations through the sync API and updates the mirror", async () => {
    await queueWalletCreate({ name: "Ví offline", type: "cash" });
    const outbox = await readOutbox();
    vi.stubGlobal("fetch", vi.fn(async (input: RequestInfo | URL) => {
      expect(new URL(String(input), "http://localhost").pathname).toBe("/api/v1/sync/mutations");
      return new Response(JSON.stringify({
        status: "ok",
        correlation_id: "req_test",
        results: [{
          mutation_id: outbox[0].id,
          entity_type: "wallet",
          entity_id: "wallet_server",
          operation: "create",
          state: "applied",
          version: 1,
          payload: { id: "wallet_server", name: "Ví server", type: "cash", balance_vnd: 0, include_in_total: true, is_default_ai: false, version: 1 },
        }],
      }), { status: 200, headers: { "Content-Type": "application/json" } });
    }));

    expect(await drainOutbox(async () => undefined)).toBe(1);

    const snapshot = await readOfflineSnapshot();
    expect(await readOutbox()).toHaveLength(0);
    expect(snapshot.wallets.some((wallet) => wallet.name === "Ví server")).toBe(true);
  });

  it("stores sync conflicts and keeps the failed mutation pending", async () => {
    const transaction = await queueTransaction({ type: "expense", source_wallet_id: "wallet_1", category_id: "cat_1", amount_vnd: 10000, occurred_at: "2026-08-31T00:00:00Z", note: "Offline edit" });
    const outbox = await readOutbox();
    vi.stubGlobal("fetch", vi.fn(async () => new Response(JSON.stringify({
      status: "ok",
      correlation_id: "req_test",
      results: [{
        mutation_id: outbox[0].id,
        entity_type: "transaction",
        entity_id: transaction.id,
        operation: "create",
        state: "conflict",
        conflict: {
          entity_type: "transaction",
          entity_id: transaction.id,
          operation: "create",
          base_version: 1,
          server_version: 2,
          local_payload: { note: "Offline edit" },
          server_payload: { id: transaction.id, type: "expense", source_wallet_id: "wallet_1", amount_vnd: 10000, balance_after_vnd: 90000, occurred_at: "2026-08-31T00:00:00Z", note: "Server edit", with_person: "", event_ref: "", excluded_from_reports: false, version: 2 },
        },
      }],
    }), { status: 200, headers: { "Content-Type": "application/json" } })));

    expect(await drainOutbox(async () => undefined)).toBe(0);

    const snapshot = await readOfflineSnapshot();
    expect(snapshot.conflicts).toHaveLength(1);
    expect(await readOutbox()).toHaveLength(1);
  });
});
