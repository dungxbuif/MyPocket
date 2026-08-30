import { describe, expect, it } from "vitest";
import { queueCategoryCreate, queueWalletCreate } from "../offline/outbox";
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
});
