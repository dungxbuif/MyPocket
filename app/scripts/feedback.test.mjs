import { strict as assert } from "node:assert";
import { createServer } from "vite";

const server = await createServer({ server: { middlewareMode: true }, appType: "custom" });
const previousFetch = globalThis.fetch;
const previousStorage = globalThis.localStorage;
globalThis.localStorage = { getItem: () => null };

try {
  const api = await server.ssrLoadModule("/src/services/feedback.ts");
  const calls = [];
  globalThis.fetch = async (_url, options = {}) => {
    calls.push(options);
    if (calls.length === 1) {
      return new Response(JSON.stringify({
        title: "Bad Request",
        status: 400,
        detail: "Ảnh chụp màn hình không hợp lệ.",
        code: "FEEDBACK_SCREENSHOT_INVALID",
      }), { status: 400, headers: { "content-type": "application/problem+json" } });
    }
    return new Response(JSON.stringify({ data: {
      id: "feedback-1",
      type: "bug",
      title: "Không mất feedback",
      description: "Nội dung vẫn được lưu.",
      status: "open",
      created_at: "2026-10-06T10:00:00Z",
      updated_at: "2026-10-06T10:00:00Z",
    }, meta: null }), { status: 201, headers: { "content-type": "application/json" } });
  };

  const result = await api.createFeedback({
    type: "bug",
    title: "Không mất feedback",
    description: "Nội dung vẫn được lưu.",
    screenshot: new Blob(["png"], { type: "image/png" }),
  });
  assert.equal(result.id, "feedback-1");
  assert.equal(calls.length, 2, "storage error must trigger one text-only retry");
  assert.ok(calls[0].body instanceof FormData, "first request keeps the screenshot");
  assert.equal(calls[1].headers.get("Content-Type"), "application/json");
  assert.deepEqual(JSON.parse(calls[1].body), {
    type: "bug",
    title: "Không mất feedback",
    description: "Nội dung vẫn được lưu.",
  });

  let unrelatedCalls = 0;
  globalThis.fetch = async () => {
    unrelatedCalls += 1;
    return new Response(JSON.stringify({ title: "Service Unavailable", status: 503, detail: "down", code: "FEEDBACK_FAILED" }), {
      status: 503,
      headers: { "content-type": "application/problem+json" },
    });
  };
  await assert.rejects(() => api.createFeedback({ type: "bug", title: "x", description: "y", screenshot: new Blob(["png"], { type: "image/png" }) }), /down/);
  assert.equal(unrelatedCalls, 1, "unrelated feedback errors must not be retried");
  console.log("Feedback contract passed: screenshot storage errors fall back to text-only exactly once.");
} finally {
  globalThis.fetch = previousFetch;
  globalThis.localStorage = previousStorage;
  await server.close();
}
