import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

test("AI API proxy allows the bounded OCR/model request to finish", async () => {
  const config = await readFile(new URL("../nginx.conf", import.meta.url), "utf8");
  assert.match(config, /proxy_read_timeout\s+240s;/);
  assert.match(config, /proxy_send_timeout\s+240s;/);
});
