import assert from "node:assert/strict";
import React from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { createServer } from "vite";

const server = await createServer({ server: { middlewareMode: true, hmr: false }, appType: "custom" });
try {
  const { LedgerPeriodSelector } = await server.ssrLoadModule("/src/atomic/molecules/LedgerPeriodSelector.tsx");
  const week = renderToStaticMarkup(React.createElement(LedgerPeriodSelector, {
    mode: "week", range: { start: "2026-09-21", end: "2026-09-27" }, weekOffset: 0,
    onModeChange() {}, onPreviousWeek() {}, onNextWeek() {}, onEditCustom() {},
  }));
  assert.match(week, /21\/09\/2026/);
  assert.match(week, /27\/09\/2026/);
  assert.match(week, /Tuần sau/);
  assert.match(week, /disabled=""/);
  assert.match(week, /Theo tuần/);
  assert.match(week, /Tùy chọn/);
  assert.doesNotMatch(week, /Tất cả/);

  const { TransactionsPanel } = await server.ssrLoadModule("/src/atomic/organisms/TransactionsPanel.tsx");
  const initialLedger = renderToStaticMarkup(React.createElement(TransactionsPanel, { onChanged() {} }));
  assert.match(initialLedger, /aria-selected="true"[^>]*>Theo tuần/);
  assert.doesNotMatch(initialLedger, /Tất cả<\/button>/);
} finally {
  await server.close();
}
