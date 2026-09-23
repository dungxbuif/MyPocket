import assert from "node:assert/strict";
import test from "node:test";

import { categoryAppliesToTransaction, signedTransactionAmount } from "../src/services/transactionLogic.ts";
import { totalWalletBalance } from "../src/services/walletLogic.ts";
import { filterLedgerTransactions, ledgerRangeForWallet } from "../src/services/transactionLedger.ts";

const category = { id: "category-1", kind: "expense", name: "Ăn uống", icon_key: "expense_food", is_system: true, wallet_ids: ["wallet-1"] };

test("category applicability matches kind and a restricted wallet", () => {
  assert.equal(categoryAppliesToTransaction(category, "expense", "wallet-1"), true);
  assert.equal(categoryAppliesToTransaction(category, "income", "wallet-1"), false);
  assert.equal(categoryAppliesToTransaction(category, "expense", "wallet-2"), false);
});

test("a category without wallet restrictions applies to every wallet", () => {
  assert.equal(categoryAppliesToTransaction({ ...category, wallet_ids: [] }, "expense", "wallet-2"), true);
});

test("signed transaction amount follows ledger direction", () => {
  assert.equal(signedTransactionAmount({ type: "income", amount: 125000 }), 125000);
  assert.equal(signedTransactionAmount({ type: "expense", amount: 125000 }), -125000);
});

test("goal picker accepts only real savings catalog keys with matching kind and wallet scope", () => {
  const deposit = {kind: "income", system_key: "income_transfer_in", wallet_ids: []};
  assert.equal(categoryAppliesToTransaction(deposit, "income", "goal-1", "goal"), true);
  assert.equal(categoryAppliesToTransaction({...deposit, system_key: "income_interest"}, "income", "goal-1", "goal"), true);
  assert.equal(categoryAppliesToTransaction({...deposit, system_key: "income_salary"}, "income", "goal-1", "goal"), false);
  assert.equal(categoryAppliesToTransaction({...deposit, system_key: null}, "income", "goal-1", "goal"), false);
  assert.equal(categoryAppliesToTransaction({...deposit, wallet_ids: ["other"]}, "income", "goal-1", "goal"), false);
  assert.equal(categoryAppliesToTransaction({...deposit, kind: "expense", system_key: "expense_transfer_out"}, "expense", "goal-1", "goal"), true);
  assert.equal(categoryAppliesToTransaction({...deposit, kind: "expense", system_key: "expense_food"}, "expense", "goal-1", "goal"), false);
  assert.equal(categoryAppliesToTransaction({...deposit, system_key: "income_salary"}, "income", "basic-1", "basic"), true);
});

test("total wallet balance uses current balances and excludes opted-out wallets", () => {
  assert.equal(totalWalletBalance([
    { current_balance: 150000, is_in_total: true },
    { current_balance: -50000, is_in_total: true },
    { current_balance: 900000, is_in_total: false },
  ]), 100000);
});

test("ledger period and wallet scope apply together at account-local day boundaries", () => {
  const rows = [
    { id: "a", wallet_id: "cash", occurred_at: "2026-09-20T17:30:00Z" },
    { id: "b", wallet_id: "cash", occurred_at: "2026-09-27T16:59:00Z" },
    { id: "c", wallet_id: "cash", occurred_at: "2026-09-27T17:00:00Z" },
    { id: "d", wallet_id: "bank", occurred_at: "2026-09-23T12:00:00Z" },
  ];
  const range = { start: "2026-09-21", end: "2026-09-27" };
  assert.deepEqual(filterLedgerTransactions(rows, "cash", range, "Asia/Ho_Chi_Minh").map(row => row.id), ["b", "a"]);
  assert.deepEqual(filterLedgerTransactions(rows, "", range, "Asia/Ho_Chi_Minh").map(row => row.id), ["b", "d", "a"]);
});

test("basic wallet and aggregate use the same selected week, with basic rows scoped to its wallet", () => {
  const rows = [
    { id: "old", wallet_id: "cash", occurred_at: "2026-08-01T12:00:00Z" },
    { id: "current", wallet_id: "cash", occurred_at: "2026-09-24T12:00:00Z" },
    { id: "other", wallet_id: "card", occurred_at: "2026-09-24T12:00:00Z" },
  ];
  const week = { start: "2026-09-21", end: "2026-09-27" };
  assert.deepEqual(filterLedgerTransactions(rows, "cash", ledgerRangeForWallet("basic", week), "Asia/Ho_Chi_Minh").map(row => row.id), ["current"]);
  assert.deepEqual(filterLedgerTransactions(rows, "", ledgerRangeForWallet(undefined, week), "Asia/Ho_Chi_Minh").map(row => row.id), ["current", "other"]);
  assert.deepEqual(filterLedgerTransactions(rows, "card", ledgerRangeForWallet("credit", week), "Asia/Ho_Chi_Minh").map(row => row.id), ["other"]);
});
