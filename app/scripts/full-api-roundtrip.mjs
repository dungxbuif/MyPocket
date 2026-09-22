// Authenticated API contract smoke through the real FE proxy. Test-owned rows are cleaned.
import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";

const origin = process.env.TEST_API_ORIGIN ?? "http://127.0.0.1:4173";
const dbURL = process.env.TEST_DATABASE_URL ?? "postgres://dev:password@127.0.0.1:5432/postgres?sslmode=disable";
let token = "";
const created = { category: "", jar: "", month: "" };

async function request(path, method = "GET", body, expected = 200) {
  const response = await fetch(`${origin}/api/v1${path}`, {
    method,
    headers: { ...(body ? { "Content-Type": "application/json" } : {}), ...(token ? { Authorization: `Bearer ${token}` } : {}) },
    ...(body ? { body: JSON.stringify(body) } : {}),
  });
  const payload = response.status === 204 ? null : await response.json();
  assert.equal(response.status, expected, `${method} ${path}: ${JSON.stringify(payload)}`);
  return payload?.data;
}

function currentMonth(timezone) {
  const parts = new Intl.DateTimeFormat("en-US", { timeZone: timezone, year: "numeric", month: "2-digit" }).formatToParts(new Date());
  const values = Object.fromEntries(parts.map(({ type, value }) => [type, value]));
  return `${values.year}-${values.month}`;
}

function cleanupJar() {
  if (!created.jar) return;
  execFileSync("psql", [dbURL, "-v", "ON_ERROR_STOP=1", "-c", `DELETE FROM jar_month_configs WHERE jar_id = '${created.jar}'; DELETE FROM jars WHERE id = '${created.jar}';`], { stdio: "ignore" });
}

const login = await request("/login", "POST", { email: process.env.TEST_EMAIL ?? "admin@mypocket.local", password: process.env.TEST_PASSWORD ?? "12345678" });
token = login.token;
try {
  const profile = await request("/auth/profile");
  assert.ok(profile.id && profile.email && profile.timezone);
  const month = currentMonth(profile.timezone);
  const home = await request("/home");
  assert.equal(typeof home.greeting, "string");
  assert.equal(typeof home.stats.wallet_count, "number");
  await request("/auth/profile", "PATCH", { timezone: profile.timezone, initialize_only: false });
  await request("/auth/profile", "PATCH", { timezone: "Mars/Phobos", initialize_only: false }, 400);

  const categories = await request("/categories");
  assert.ok(Array.isArray(categories) && categories.length > 0);
  const marker = `Full API probe ${Date.now()}`;
  const category = await request("/categories", "POST", { name: marker, kind: "expense", wallet_ids: [], icon_key: "tag" }, 201);
  created.category = category.id;
  const updatedCategory = await request(`/categories/${category.id}`, "PATCH", { name: `${marker} updated`, kind: "expense", wallet_ids: [], icon_key: "tag" });
  assert.equal(updatedCategory.name, `${marker} updated`);
  await request(`/categories/${category.id}/wallets`, "PATCH", { wallet_ids: [] });

  const jars = await request(`/jars?month=${month}`);
  assert.equal(jars.month, month);
  assert.ok(Array.isArray(jars.items) && Array.isArray(jars.jars));
  const jar = await request("/jars", "POST", { month, name: marker, allocation_mode: "fixed", allocation_amount: 1000 }, 201);
  created.jar = jar.jar_id;
  const listedJars = await request(`/jars?month=${month}`);
  assert.ok(listedJars.items.some((item) => item.jar_id === created.jar));
  await request(`/jars/${created.jar}/months/${month}`, "PUT", { month, name: `${marker} updated`, allocation_mode: "percent", allocation_percent: 10 });
  const cumulative = await request(`/jars/${created.jar}/report?from=${month}&to=${month}`);
  assert.equal(cumulative.jar_id, created.jar);
  await request(`/jars/${created.jar}/months/${month}`, "DELETE", undefined, 204);

  const summary = await request(`/months/${month}`);
  assert.equal(summary.month, month);
  assert.equal(summary.timezone, profile.timezone);
  const note = `${marker} note`;
  await request(`/months/${month}/note`, "PUT", { note }, 204);
  assert.equal((await request(`/months/${month}`)).note, note);
  await request(`/months/${month}/note`, "DELETE", undefined, 204);
  assert.equal((await request(`/months/${month}`)).note, "");

  const capabilities = await request("/ai/entry/capabilities");
  for (const key of ["ai_configured", "ocr_configured", "files_configured"]) assert.equal(typeof capabilities[key], "boolean");
  const docs = await fetch(`${origin}/api/v1/docs/doc.json`);
  assert.equal(docs.status, 200);
  assert.match(await docs.text(), /MyPocket API/);

  await request(`/categories/${created.category}`, "DELETE", undefined, 204);
  created.category = "";
  created.month = month;
  console.log("PASS: full FE proxy API contract — auth/profile/home, category CRUD and wallet scope, jar lifecycle/report, month summary/note, AI capabilities and Swagger.");
} finally {
  if (created.category) await request(`/categories/${created.category}`, "DELETE", undefined, 204).catch(() => {});
  cleanupJar();
  console.log("Cleanup: test category, jar/configuration and month note only; seeded account timezone value was unchanged.");
}
