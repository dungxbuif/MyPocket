import assert from "node:assert/strict";
import React from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { createServer } from "vite";

const server = await createServer({ server: { middlewareMode: true, hmr: false }, appType: "custom" });
try {
  const { StatusBadge } = await server.ssrLoadModule("/src/atomic/atoms/StatusBadge.tsx");
  const active = renderToStaticMarkup(React.createElement(StatusBadge, { label: "Đang bật", tone: "success" }));
  assert.match(active, /role="status"/);
  assert.match(active, /Đang bật/);
  assert.match(active, /status-badge/);
  const paused = renderToStaticMarkup(React.createElement(StatusBadge, { label: "Tạm dừng", tone: "muted" }));
  assert.match(paused, /Tạm dừng/);
  const { FinanceAssistantPart } = await server.ssrLoadModule("/src/atomic/molecules/FinanceAssistantPart.tsx");
  const help = renderToStaticMarkup(React.createElement(FinanceAssistantPart, {
    type: "product_help",
    data: { view: { version: "1.1", sources: [{ id: "handbook.budgets.v1.1", title: "Ngân sách", excerpt: "Mở Tài khoản." }] } },
  }));
  assert.match(help, /Hướng dẫn trong ứng dụng/);
  assert.match(help, /handbook\.budgets\.v1\.1/);
} finally {
  await server.close();
}

console.log("Base status badge design contract passed.");
