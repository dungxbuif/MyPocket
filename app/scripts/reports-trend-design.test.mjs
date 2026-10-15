import assert from "node:assert/strict";
import React from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { createServer } from "vite";

const server = await createServer({ server: { middlewareMode: true, hmr: false }, appType: "custom" });
try {
  const { BaseTrendChart } = await server.ssrLoadModule("/src/atomic/molecules/BaseCharts.tsx");
  const empty = renderToStaticMarkup(React.createElement(BaseTrendChart, { points: [] }));
  assert.match(empty, /Xu hướng chi tiêu/);
  assert.match(empty, /Chưa có giao dịch được tính vào xu hướng/);

  const chart = renderToStaticMarkup(React.createElement(BaseTrendChart, {
    points: [{ day: "2026-10-01", income: 500000, expense: 120000 }],
  }));
  assert.match(chart, /role="img"/);
  assert.match(chart, /Xu hướng thu chi/);
  assert.match(chart, /polyline/);
} finally {
  await server.close();
}

console.log("Web trend chart design contract passed.");
