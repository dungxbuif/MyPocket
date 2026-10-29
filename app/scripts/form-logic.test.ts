import { strict as assert } from "node:assert";
import { test } from "node:test";
import { monthDays, shiftDay } from "../src/services/formDates.ts";
import { calculateAmount } from "../src/services/amountCalculator.ts";
test("calendar uses Monday grid and preserves time across leap/year boundaries", () => {
  assert.equal(monthDays(2026, 7)[0], "2026-07-27");
  assert.equal(monthDays(2026, 7).length, 42);
  assert.equal(shiftDay("2024-02-28T23:10", 1), "2024-02-29T23:10");
  assert.equal(shiftDay("2026-12-31", 1), "2027-01-01");
});
test("VND calculator accepts arithmetic and rejects fractions/unsafe input", () => {
  assert.equal(calculateAmount("35000+12000×2"), "59000");
  assert.equal(calculateAmount("100000÷4"), "25000");
  for (const input of ["10÷0", "5÷2", "1-2", "1+", "alert(1)", "9007199254740992"]) assert.equal(calculateAmount(input), null, input);
});
