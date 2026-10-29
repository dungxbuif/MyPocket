import assert from "node:assert/strict";
import React from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { createServer } from "vite";

const server = await createServer({ server: { middlewareMode: true, hmr: false }, appType: "custom" });
try {
  const { WalletScopeSelector } = await server.ssrLoadModule("/src/atomic/molecules/WalletScopeSelector.tsx");
  const wallets = [
    { id: "cash", name: "Tiền mặt", type: "basic", currency: "VND", opening_balance: 0, current_balance: 125000, is_in_total: true },
    { id: "trip", name: "Du lịch", type: "goal", currency: "VND", opening_balance: 0, current_balance: 500000, is_in_total: true },
  ];
  const aggregate = renderToStaticMarkup(React.createElement(WalletScopeSelector, { wallets, selectedID: "", onSelect() {}, onAdd() {} }));
  assert.match(aggregate, /aria-label="Chọn ví giao dịch"/);
  assert.match(aggregate, />Tổng cộng</);
  assert.match(aggregate, /ChevronDown|svg/);

  const selected = renderToStaticMarkup(React.createElement(WalletScopeSelector, { wallets, selectedID: "cash", onSelect() {}, onAdd() {} }));
  assert.match(selected, />Tiền mặt</);
  assert.doesNotMatch(selected, /<select/);
} finally {
  await server.close();
}

console.log("Wallet scope design contract passed.");
