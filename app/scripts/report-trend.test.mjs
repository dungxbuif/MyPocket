import assert from "node:assert/strict";
import { createServer } from "vite";

const server = await createServer({ server: { middlewareMode: true, hmr: false }, appType: "custom" });
try {
  const { buildReportTrend } = await server.ssrLoadModule("/src/services/reportTrend.ts");
  const rows = [
    { type: "expense", amount: 120000, occurred_at: "2026-10-01T00:30:00+07:00", included_in_reports: true },
    { type: "income", amount: 500000, occurred_at: "2026-09-30T17:30:00Z", included_in_reports: true },
    { type: "expense", amount: 900000, occurred_at: "2026-10-01T03:00:00+07:00", included_in_reports: false },
    { type: "adjustment", amount: 700000, occurred_at: "2026-10-02T10:00:00+07:00", included_in_reports: true },
  ];
  const points = buildReportTrend(rows, "2026-10", "Asia/Ho_Chi_Minh");
  assert.deepEqual(points, [{ day: "2026-10-01", income: 500000, expense: 120000 }]);
  assert.deepEqual(buildReportTrend([], "2026-10", "Asia/Ho_Chi_Minh"), []);
} finally {
  await server.close();
}

console.log("Web report trend contract passed.");
