import { strict as assert } from "node:assert";
import { createServer } from "vite";

const server = await createServer({ server: { middlewareMode: true }, appType: "custom" });
const originalFetch = globalThis.fetch;
const originalStorage = globalThis.localStorage;
const calls = [];
globalThis.localStorage = { getItem: () => null, setItem() {}, removeItem() {} };
globalThis.fetch = async (url, options) => {
  calls.push({ url: String(url), options });
  return new Response(JSON.stringify({ data: {
    token: "access-token",
    expires_at: "2026-10-07T00:00:00Z",
    refresh_token: "refresh-token",
    refresh_expires_at: "2026-11-06T00:00:00Z",
    user: { id: "user-1", name: "Test User", email: "user@example.test", timezone: "Asia/Ho_Chi_Minh", timezone_confirmed: true, created_at: "2026-01-01T00:00:00Z" },
  } }), { status: 200, headers: { "content-type": "application/json" } });
};

try {
  const { loginWithPassword } = await server.ssrLoadModule("/src/services/auth.ts");
  const output = await loginWithPassword("user@example.test", "not-a-real-password");
  assert.equal(calls.length, 1);
  assert.equal(calls[0].url, "/api/v1/login");
  assert.equal(calls[0].options.method, "POST");
  assert.deepEqual(JSON.parse(calls[0].options.body), { email: "user@example.test", password: "not-a-real-password" });
  assert.equal(output.user.email, "user@example.test");
  assert.equal(output.token, "access-token");
  console.log("Password login contract passed: POST /api/v1/login and LoginOutput parsing.");
} finally {
  globalThis.fetch = originalFetch;
  globalThis.localStorage = originalStorage;
  await server.close();
}
