import assert from "node:assert/strict";
import React from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { createServer } from "vite";

const server = await createServer({ server: { middlewareMode: true, hmr: false }, appType: "custom" });
try {
  const { WalletDetailPanel } = await server.ssrLoadModule("/src/atomic/organisms/SavingsWalletPanel.tsx");
  const wallet = { id: "goal-1", name: "Sổ tiết kiệm", type: "goal", currency: "VND", opening_balance: 0, current_balance: 10000000, target_amount: 1000000000, target_date: "2026-10-05", is_in_total: true };
  const transactions = [{ id: "txn-1", owner_id: "owner-1", wallet_id: wallet.id, category_id: "category-1", type: "income", amount: 10000000, occurred_at: "2026-09-23T10:00:00Z", included_in_reports: true, created_at: "2026-09-23T10:00:00Z", updated_at: "2026-09-23T10:00:00Z" }];
  const categories = [{ id: "category-1", name: "Tiền chuyển đến", kind: "income", system_key: "income_transfer_in", icon_key: "income_transfer_in", is_system: true, wallet_ids: [] }];
  const html = renderToStaticMarkup(React.createElement(WalletDetailPanel, { wallet, transactions, categories, embedded: true, onBack() {}, onChanged() {} }));
  assert.match(html, /Số dư/);
  assert.match(html, /CẦN THÊM/);
  assert.match(html, /TẤT CẢ CÁC GIAO DỊCH/);
  assert.match(html, /Tiền chuyển đến/);
  assert.doesNotMatch(html, /Chọn ví/);
  assert.doesNotMatch(html, /Theo tuần/);

  const credit = { ...wallet, id: "credit-1", name: "Thẻ tín dụng", type: "credit", credit_limit: 20000000 };
  const creditRows = [{ ...transactions[0], id: "credit-row", wallet_id: credit.id }];
  const creditHtml = renderToStaticMarkup(React.createElement(WalletDetailPanel, { wallet: credit, transactions: creditRows, categories, embedded: true, onBack() {}, onChanged() {} }));
  assert.doesNotMatch(creditHtml, /Thêm giao dịch/);
  assert.match(creditHtml, /disabled=""[^>]*>.*Tiền chuyển đến/s);
} finally {
  await server.close();
}
