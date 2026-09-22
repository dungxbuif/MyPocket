import assert from "node:assert/strict";
import test from "node:test";

import { accountMonthKey, instantFromLocalDateTime, localDateTimeAt, dateKeyAt } from "../src/services/accountTime.ts";

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
