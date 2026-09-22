import assert from "node:assert/strict";
import test from "node:test";

import { jarAssignmentNeedsSelection, jarUsagePercent, monthKeyForInstant, monthLabel, monthOptions, resolveMonthKey, shiftMonthKey } from "../src/services/monthJarLogic.ts";

test("month route defaults to the account-local month and keeps a valid selected month", () => {
  const now = new Date("2026-09-30T17:30:00.000Z");
  assert.equal(resolveMonthKey(undefined, "Asia/Ho_Chi_Minh", now), "2026-10");
  assert.equal(resolveMonthKey("2025-02", "Asia/Ho_Chi_Minh", now), "2025-02");
  assert.equal(resolveMonthKey("2026-13", "Asia/Ho_Chi_Minh", now), "2026-10");
});

test("month navigation crosses year and leap-month boundaries", () => {
  assert.equal(shiftMonthKey("2026-12", 1), "2027-01");
  assert.equal(shiftMonthKey("2024-03", -1), "2024-02");
});

test("month picker options are ordered and include historical selections", () => {
  assert.deepEqual(monthOptions("2026-01", 2, 1), ["2025-11", "2025-12", "2026-01", "2026-02"]);
});

test("jar choices follow the transaction's account-local month, not its UTC month", () => {
  const instant = "2026-09-30T17:30:00.000Z";
  assert.equal(monthKeyForInstant(instant, "Asia/Ho_Chi_Minh"), "2026-10");
  assert.equal(monthKeyForInstant(instant, "America/Los_Angeles"), "2026-09");
});

test("a removed jar can remain on an unchanged historical timestamp but not move to another date", () => {
  const oldDate = "2026-08-31T17:00:00.000Z";
  const removed = [{ jar_id: "jar-food", name: "Ăn uống", active: false }];
  assert.equal(jarAssignmentNeedsSelection("jar-food", removed, oldDate, oldDate), false);
  assert.equal(jarAssignmentNeedsSelection("jar-food", removed, oldDate, "2026-09-01T17:00:00.000Z"), true);
  assert.equal(jarAssignmentNeedsSelection("jar-food", [{ ...removed[0], active: true }], oldDate, "2026-09-01T17:00:00.000Z"), false);
});

test("month labels are calendar-based and jar progress exposes overspend", () => {
  assert.match(monthLabel("2026-02"), /tháng 2/i);
  assert.equal(jarUsagePercent(150, 100), 150);
  assert.equal(jarUsagePercent(150, 0), null);
});
