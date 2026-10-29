import { describe, expect, it } from "vitest";
import { mergePendingTransactions } from "./pendingTransactions";
import type { Transaction } from "./finance";

const confirmed: Transaction = { id: "tx-1", type: "expense", source_wallet_id: "wallet-1", amount_vnd: 123000, balance_after_vnd: -123000, occurred_at: "2026-09-10T12:00:00+07:00", note: "confirmed", with_person: "", event_ref: "", excluded_from_reports: false, version: 1 };
const pending = { id: "mutation-1", entity_id: "tx-1", entity_type: "transaction" as const, operation: "create" as const, input: { ...confirmed, note: "pending" }, created_at: "2026-09-10" };

describe("pending transaction reconciliation", () => {
 it("does not duplicate a server-committed create whose receipt was lost", () => {
  expect(mergePendingTransactions([confirmed], [pending])).toEqual([confirmed]);
 });
 it("uses entity ID for an uncommitted create and excludes wallet commands", () => {
  const rows = mergePendingTransactions([], [pending, { ...pending, entity_type: "wallet", entity_id: "wallet-2" }]);
  expect(rows).toHaveLength(1);
  expect(rows[0].id).toBe("tx-1");
 });
 it("overlays an edit once and hides an archived transaction", () => {
  const edited = mergePendingTransactions([confirmed], [{ ...pending, operation: "update" }]);
  expect(edited).toHaveLength(1);
  expect(edited[0]).toMatchObject({id: "tx-1", note: "pending"});
  expect(mergePendingTransactions([confirmed], [{...pending, operation: "archive", input: {}}])).toEqual([]);
 });
});
