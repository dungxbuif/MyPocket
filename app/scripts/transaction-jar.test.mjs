import assert from "node:assert/strict";
import React from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { createServer } from "vite";

const server = await createServer({ server: { middlewareMode: true, hmr: false }, appType: "custom" });
try {
  const { TransactionFields } = await server.ssrLoadModule("/src/atomic/molecules/TransactionFields.tsx");
  const render = (state, categories = []) => renderToStaticMarkup(React.createElement(TransactionFields, {
    state,
    onChange() {},
    wallets: [{ id: "wallet-1", name: "Tiền mặt", type: "basic", currency: "VND", current_balance: 0 }],
    categories,
    jars: [{ jar_id: "jar-food", name: "Ăn uống", active: true }],
  }));
  const base = { type: "expense", amount: "45000", walletID: "wallet-1", categoryID: "", jarID: "jar-food", occurredAt: "2026-09-22T12:00", note: "", includedInReports: true };
  const expense = render(base);
  assert.match(expense, /Hũ \(tùy chọn\)/);
  assert.match(expense, /Không gắn hũ/);
  assert.match(expense, /Ăn uống/);

  const income = render({ ...base, type: "income" });
  assert.doesNotMatch(income, /Chọn hũ cho giao dịch/);

  const transferOut = render({ ...base, categoryID: "transfer-out" }, [{ id: "transfer-out", name: "Chuyển đi", kind: "expense", system_key: "expense_transfer_out", wallet_ids: [] }]);
  assert.doesNotMatch(transferOut, /Chọn hũ cho giao dịch/);
} finally {
  await server.close();
}
