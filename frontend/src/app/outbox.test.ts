import { afterEach, describe, expect, it } from "vitest";
import { drainOutbox, queueTransaction, readOutbox } from "./outbox";

describe("transaction outbox", () => {
  afterEach(() => localStorage.clear());

  it("stores an optimistic transaction for offline replay", () => {
    const transaction = queueTransaction({ type: "expense", source_wallet_id: "wallet_1", amount_vnd: 42000, occurred_at: "2026-08-30T00:00:00Z", note: "Offline coffee" });
    expect(transaction.version).toBe(0);
    expect(readOutbox()).toHaveLength(1);
    expect(readOutbox()[0].input.note).toBe("Offline coffee");
  });

  it("drains queued mutations in order", async () => {
    queueTransaction({ type: "expense", source_wallet_id: "wallet_1", amount_vnd: 1, occurred_at: "2026-08-30T00:00:00Z" });
    queueTransaction({ type: "expense", source_wallet_id: "wallet_1", amount_vnd: 2, occurred_at: "2026-08-30T00:00:00Z" });
    const amounts: number[] = [];
    expect(await drainOutbox(async (input) => { amounts.push(Number(input.amount_vnd)); })).toBe(2);
    expect(amounts).toEqual([1, 2]);
    expect(readOutbox()).toHaveLength(0);
  });
});
