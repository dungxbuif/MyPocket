import assert from "node:assert/strict";
import test from "node:test";

import { categoryMatchRank } from "../src/services/categorySearch.ts";

test("search ranks a matching root before a child-only match", () => {
  assert.equal(categoryMatchRank("Ăn uống", ["Highlands"], "ăn"), 0);
  assert.equal(categoryMatchRank("Khác", ["Highlands"], "high"), 1);
  assert.equal(categoryMatchRank("Khác", ["Mua sắm"], "high"), 2);
});
