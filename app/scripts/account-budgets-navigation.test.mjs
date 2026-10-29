import assert from "node:assert/strict";
import test from "node:test";
import { createServer } from "vite";
import React from "react";
import { renderToStaticMarkup } from "react-dom/server";

const server = await createServer({ server: { middlewareMode: true }, appType: "custom" });
try {
  const { BottomNavigation } = await server.ssrLoadModule("/src/atomic/organisms/BottomNavigation.tsx");
  const nav = renderToStaticMarkup(React.createElement(BottomNavigation, { tab: "transactions", onTabChange() {}, onAdd() {} }));
  test("bottom navigation keeps budgets out of primary destinations", () => {
    assert.doesNotMatch(nav, /Ngân sách/);
    assert.match(nav, /Giao dịch/);
    assert.match(nav, /Tài khoản/);
  });

  const fs = await import("node:fs/promises");
  const accountSource = await fs.readFile(new URL("../src/atomic/organisms/AccountPanel.tsx", import.meta.url), "utf8");
  const pageSource = await fs.readFile(new URL("../src/atomic/pages/FinancePrototypePage.tsx", import.meta.url), "utf8");
  const transactionSource = await fs.readFile(new URL("../src/atomic/organisms/TransactionsPanel.tsx", import.meta.url), "utf8");
  test("account owns the canonical budgets route", () => {
    assert.match(accountSource, /account\/budgets/);
    assert.match(pageSource, /account\/budgets/);
  });
  test("transactions consumes the shared screen header", () => {
    assert.match(transactionSource, /ScreenHeader/);
  });
} finally {
  await server.close();
}
