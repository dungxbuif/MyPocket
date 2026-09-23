import assert from "node:assert/strict";
import test from "node:test";

import { accountMonthKey, instantFromLocalDateTime, localDateTimeAt, dateKeyAt, weekDateRange } from "../src/services/accountTime.ts";

test("calendar grouping follows the account timezone, not the browser timezone", () => {
  const instant = "2026-09-30T17:30:00.000Z";
  assert.equal(dateKeyAt(instant, "Asia/Ho_Chi_Minh"), "2026-10-01");
  assert.equal(accountMonthKey(instant, "Asia/Ho_Chi_Minh"), "2026-10");
  assert.equal(dateKeyAt(instant, "America/Los_Angeles"), "2026-09-30");
});

test("local datetime conversion handles a DST gap and chooses the earlier fold", () => {
  assert.throws(() => instantFromLocalDateTime("2026-03-08T02:30", "America/Los_Angeles"), /does not exist/i);
  assert.equal(instantFromLocalDateTime("2026-11-01T01:30", "America/Los_Angeles"), "2026-11-01T08:30:00.000Z");
});

test("formatting an instant for a datetime field uses account wall time", () => {
  assert.equal(localDateTimeAt("2026-10-01T01:23:00Z", "Asia/Ho_Chi_Minh"), "2026-10-01T08:23");
});

test("ledger weeks start Monday and cross month and year boundaries", () => {
  assert.deepEqual(weekDateRange("2026-09-23", 0), { start: "2026-09-21", end: "2026-09-27" });
  assert.deepEqual(weekDateRange("2026-01-01", 0), { start: "2025-12-29", end: "2026-01-04" });
  assert.deepEqual(weekDateRange("2026-09-23", -1), { start: "2026-09-14", end: "2026-09-20" });
});

test("ledger current week anchors to the account date, not UTC date", () => {
  const today = dateKeyAt("2026-09-20T18:00:00Z", "Asia/Ho_Chi_Minh");
  assert.deepEqual(weekDateRange(today, 0), { start: "2026-09-21", end: "2026-09-27" });
});
