import assert from "node:assert/strict";
import test from "node:test";

import { mergeAdvisorMessages } from "../src/services/advisorHistory.ts";

type AdvisorMessage = { id: string; conversation_id: string; seq: number; role: "user" | "assistant"; parts: { type: string; text?: string }[]; created_at: string };

const message = (id: string, seq: number): AdvisorMessage => ({
  id,
  conversation_id: "conversation-1",
  seq,
  role: seq % 2 ? "user" : "assistant",
  parts: [{ type: "text", text: id }],
  created_at: `2026-09-22T00:0${seq}:00Z`,
});

test("merges older advisor pages without duplicates and keeps chronological order", () => {
  const merged = mergeAdvisorMessages([message("m3", 3), message("m4", 4)], [message("m1", 1), message("m2", 2), message("m3", 3)]);
  assert.deepEqual(merged.map(item => item.id), ["m1", "m2", "m3", "m4"]);
});
