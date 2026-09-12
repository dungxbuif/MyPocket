import { expect, test, request, type APIRequestContext } from "@playwright/test";

test("two real users cannot read or mutate each other's wallet using cookie or API key", async () => {
  const suffix = crypto.randomUUID();
  const a = await request.newContext({ baseURL: "http://127.0.0.1:18173" });
  const b = await request.newContext({ baseURL: "http://127.0.0.1:18173" });
  async function login(client: APIRequestContext, name: string) {
    const result = await client.get(`/api/v1/auth/google/callback?subject=${name}-${suffix}&email=${name}-${suffix}@example.com&email_verified=true&name=${name}`, { maxRedirects: 0 });
    expect(result.status()).toBe(302);
    const csrf = (await client.storageState()).cookies.find((cookie) => cookie.name === "mypocket_csrf");
    expect(csrf).toBeTruthy();
    return { "X-CSRF-Token": csrf!.value };
  }
  try {
    const headersA = await login(a, "owner-a");
    const headersB = await login(b, "owner-b");
    const created = await a.post("/api/v1/wallets", { headers: headersA, data: { name: "Private A", type: "basic" } });
    expect(created.status()).toBe(201);
    const { wallet } = await created.json();
    expect((await (await b.get("/api/v1/wallets")).json()).wallets ?? []).toEqual([]);
    const denied = await b.post(`/api/v1/wallets/${wallet.id}/archive`, {
      headers: headersB,
      data: { base_version: wallet.version },
    });
    expect(denied.status()).toBe(403);
    const keyResponse = await b.post("/api/v1/api-keys", { headers: headersB, data: { name: "Isolation test" } });
    expect(keyResponse.status()).toBe(201);
    const { key } = await keyResponse.json();
    const bearer = await request.newContext({ baseURL: "http://127.0.0.1:18173", extraHTTPHeaders: { Authorization: `Bearer ${key.plaintext}` } });
    try {
      expect((await bearer.get("/api/v1/me")).status()).toBe(200);
      expect((await bearer.post(`/api/v1/wallets/${wallet.id}/archive`, {
        data: { base_version: wallet.version },
      })).status()).toBe(403);
      expect((await bearer.post("/api/v1/api-keys", { data: { name: "Forbidden key" } })).status()).toBe(403);
      expect((await b.post(`/api/v1/api-keys/${key.id}/revoke`, { headers: headersB })).status()).toBe(200);
      expect((await bearer.get("/api/v1/me")).status()).toBe(401);
    } finally { await bearer.dispose(); }
    expect((await (await a.get("/api/v1/wallets")).json()).wallets).toHaveLength(1);
  } finally { await a.dispose(); await b.dispose(); }
});
