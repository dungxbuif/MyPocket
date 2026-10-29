import assert from "node:assert/strict";
import { createServer } from "vite";

const server = await createServer({ server: { middlewareMode: true, hmr: false }, appType: "custom" });
try {
  const { buildWeekReportTotals } = await server.ssrLoadModule("/src/services/weekReport.ts");
  const rows = [
    { type: "expense", amount: 120000, occurred_at: "2026-10-11T17:00:00Z", included_in_reports: true },
    { type: "income", amount: 500000, occurred_at: "2026-10-11T17:00:00Z", included_in_reports: true },
    { type: "expense", amount: 1000000, occurred_at: "2026-10-11T16:59:59Z", included_in_reports: true },
    { type: "income", amount: 200000, occurred_at: "2026-10-18T16:59:59Z", included_in_reports: true },
    { type: "expense", amount: 2000000, occurred_at: "2026-10-18T17:00:00Z", included_in_reports: true },
    { type: "expense", amount: 900000, occurred_at: "2026-10-13T03:00:00+07:00", included_in_reports: false },
    { type: "transfer", amount: 700000, occurred_at: "2026-10-14T10:00:00+07:00", included_in_reports: true },
    { type: "adjustment", amount: 600000, occurred_at: "2026-10-15T10:00:00+07:00", included_in_reports: true },
  ];

  assert.deepEqual(
    buildWeekReportTotals(rows, "Asia/Ho_Chi_Minh", new Date("2026-10-11T18:00:00Z")),
    { income: 700000, expense: 120000 },
  );
} finally {
  await server.close();
}

console.log("Web overview week report contract passed.");
