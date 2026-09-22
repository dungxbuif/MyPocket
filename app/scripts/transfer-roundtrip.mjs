// Real integrated proof: Vite proxy -> Gin -> PostgreSQL. Cleans only IDs
// created by this run and verifies both persisted state and report exclusion.
import assert from "node:assert/strict";

const origin = process.env.TEST_API_ORIGIN ?? "http://127.0.0.1:4173";
let token = "";

async function request(path, method = "GET", body, expected = 200) {
  const response = await fetch(`${origin}/api/v1${path}`, {
    method,
    headers: { "Content-Type": "application/json", ...(token ? { Authorization: `Bearer ${token}` } : {}) },
    ...(body ? { body: JSON.stringify(body) } : {}),
  });
  const payload = response.status === 204 ? null : await response.json();
  assert.equal(response.status, expected, `${method} ${path}: ${JSON.stringify(payload)}`);
  return payload?.data;
}

const login = await request("/login", "POST", {
  email: process.env.TEST_EMAIL ?? "admin@mypocket.local",
  password: process.env.TEST_PASSWORD ?? "12345678",
});
token = login.token;

const createdWallets = [];
const createdTransactions = [];
try {
  const suffix = Date.now().toString(36);
  const source = await request("/wallets", "POST", { name: `E2E nguồn ${suffix}`, type: "basic", currency: "VND", opening_balance: 100000, is_in_total: false }, 201);
  const destination = await request("/wallets", "POST", { name: `E2E đích ${suffix}`, type: "basic", currency: "VND", opening_balance: 50000, is_in_total: false }, 201);
  createdWallets.push(source.id, destination.id);

  const profile = await request("/auth/profile");
  const parts = new Intl.DateTimeFormat("en", { timeZone: profile.timezone, year: "numeric", month: "2-digit" }).formatToParts(new Date());
  const localMonth = `${parts.find(part => part.type === "year").value}-${parts.find(part => part.type === "month").value}`;
  const beforeMonth = await request(`/months/${localMonth}`);
  const transfer = await request("/transactions/transfer", "POST", {
    source_wallet_id: source.id,
    destination_wallet_id: destination.id,
    amount: 25000,
    occurred_at: new Date().toISOString(),
    note: "E2E transfer",
  }, 201);
  assert.equal(transfer.length, 2, "transfer response must contain source and destination rows");
  assert.equal(transfer[0].transfer_id, transfer[1].transfer_id, "paired rows must share transfer_id");
  assert.deepEqual(new Set(transfer.map(row => row.type)), new Set(["income", "expense"]));
  assert.equal(transfer.every(row => row.included_in_reports === false), true, "paired rows must be excluded from reports");
  assert.equal(transfer.every(row => row.jar_id == null), true, "paired rows must not carry a jar");
  createdTransactions.push(...transfer.map(row => row.id));

  const wallets = await request("/wallets");
  assert.equal(wallets.find(row => row.id === source.id).current_balance, 75000);
  assert.equal(wallets.find(row => row.id === destination.id).current_balance, 75000);
  const rows = await request("/transactions");
  assert.equal(rows.filter(row => row.transfer_id === transfer[0].transfer_id).length, 2);
  const afterMonth = await request(`/months/${localMonth}`);
  assert.equal(afterMonth.income, beforeMonth.income, "transfer must not change monthly income");
  assert.equal(afterMonth.expense, beforeMonth.expense, "transfer must not change monthly expense");

  const unauth = await fetch(`${origin}/api/v1/transactions/transfer`, { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ source_wallet_id: source.id, destination_wallet_id: destination.id, amount: 1 }) });
  assert.equal(unauth.status, 401, "transfer endpoint must remain protected");
  console.log("PASS: integrated transfer state — paired rows, balances, report exclusion, jar exclusion and auth guard.");
} finally {
  for (const id of createdTransactions) await request(`/transactions/${id}`, "DELETE", undefined, 204);
  for (const id of createdWallets) await request(`/wallets/${id}`, "DELETE", undefined, 204);
  console.log("Cleanup: only transfer E2E rows and wallets created by this run were removed.");
}
