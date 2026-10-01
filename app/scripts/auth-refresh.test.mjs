import assert from "node:assert/strict";
import { createServer } from "vite";

const server = await createServer({ server: { middlewareMode: true }, appType: "custom" });
const previousFetch = globalThis.fetch;
const previousStorage = globalThis.localStorage;
const session = {
  token: "expired-access",
  refreshToken: "mpr_refresh",
  refreshExpiresAt: "2026-10-30T00:00:00Z",
  expiresAt: "2026-09-30T00:00:00Z",
  user: { id: "u1" },
};
const storage = new Map([["mypocket.auth.session.v1", JSON.stringify(session)]]);
globalThis.localStorage = { getItem: key => storage.get(key) ?? null, setItem: (key, value) => storage.set(key, value), removeItem: key => storage.delete(key) };
let calls = [];
globalThis.fetch = async (url, options = {}) => {
  calls.push({ url: String(url), options });
  if (String(url).endsWith("/api/v1/auth/refresh")) {
    return new Response(JSON.stringify({ data: { token: "fresh-access", expires_at: "2026-10-01T00:00:00Z", refresh_token: "mpr_rotated", refresh_expires_at: "2026-10-31T00:00:00Z" } }), { status: 200, headers: { "content-type": "application/json" } });
  }
  if (calls.length === 1) return new Response(JSON.stringify({ detail: "expired" }), { status: 401, headers: { "content-type": "application/problem+json" } });
  return new Response(JSON.stringify({ data: { ok: true } }), { status: 200, headers: { "content-type": "application/json" } });
};

try {
  const { apiRequest } = await server.ssrLoadModule("/src/services/api.ts");
  const result = await apiRequest("/api/v1/protected", {}, "expired-access");
  assert.deepEqual(result, { ok: true });
  assert.equal(calls.length, 3, "one failed request, one refresh and one retry");
  assert.match(calls[2].options.headers.get("Authorization"), /fresh-access/);
  assert.equal(JSON.parse(storage.get("mypocket.auth.session.v1")).refreshToken, "mpr_rotated");
  console.log("Auth refresh contract passed: one-shot refresh, rotation persistence and retry.");
} finally {
  globalThis.fetch = previousFetch;
  globalThis.localStorage = previousStorage;
  await server.close();
}
