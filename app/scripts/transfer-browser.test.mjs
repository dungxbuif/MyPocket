// Real Chromium E2E: authenticates through the local callback, opens the
// production-shaped Quick Add UI, submits a transfer, and verifies API state.
import assert from "node:assert/strict";
import { spawn } from "node:child_process";
import { mkdtemp, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import path from "node:path";

const origin = process.env.TEST_API_ORIGIN ?? "http://127.0.0.1:4173";
const api = `${origin}/api/v1`;
let token = "";
const createdWallets = [];
const createdTransactions = [];

async function request(pathname, method = "GET", body, expected = 200) {
  const response = await fetch(`${api}${pathname}`, {
    method,
    headers: { "Content-Type": "application/json", ...(token ? { Authorization: `Bearer ${token}` } : {}) },
    ...(body ? { body: JSON.stringify(body) } : {}),
  });
  const payload = response.status === 204 ? null : await response.json();
  assert.equal(response.status, expected, `${method} ${pathname}: ${JSON.stringify(payload)}`);
  return payload?.data;
}

const login = await request("/login", "POST", { email: process.env.TEST_EMAIL ?? "admin@mypocket.local", password: process.env.TEST_PASSWORD ?? "12345678" });
token = login.token;
const suffix = Date.now().toString(36);

const profile = await request("/auth/profile");
const source = await request("/wallets", "POST", { name: `UI E2E nguồn ${suffix}`, type: "basic", currency: "VND", opening_balance: 100000, is_in_total: false }, 201);
const destination = await request("/wallets", "POST", { name: `UI E2E đích ${suffix}`, type: "basic", currency: "VND", opening_balance: 50000, is_in_total: false }, 201);
createdWallets.push(source.id, destination.id);

const chromeProfile = await mkdtemp(path.join(tmpdir(), "mypocket-transfer-browser-"));
const chromePath = process.env.CHROME_BIN ?? "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome";
let chrome;
let socket;
try {
  chrome = spawn(chromePath, ["--headless=new", "--no-first-run", "--no-default-browser-check", "--disable-background-timer-throttling", "--disable-renderer-backgrounding", "--remote-debugging-port=0", `--user-data-dir=${chromeProfile}`, "about:blank"], { stdio: ["ignore", "ignore", "pipe"] });
  const endpoint = await new Promise((resolve, reject) => {
    let output = "";
    const timeout = setTimeout(() => reject(new Error("Chrome startup timeout")), 15000);
    chrome.once("error", reject);
    chrome.stderr.on("data", chunk => {
      output += chunk;
      const match = output.match(/DevTools listening on (ws:\/\/[^\s]+)/);
      if (match) { clearTimeout(timeout); resolve(match[1]); }
    });
  });
  socket = new WebSocket(endpoint);
  await new Promise(resolve => socket.addEventListener("open", resolve, { once: true }));
  let nextID = 0;
  const pending = new Map();
  socket.addEventListener("message", event => {
    const message = JSON.parse(event.data);
    if (!pending.has(message.id)) return;
    const item = pending.get(message.id);
    pending.delete(message.id);
    if (message.error) item.reject(new Error(message.error.message)); else item.resolve(message.result);
  });
  const send = (method, params = {}, sessionId) => new Promise((resolve, reject) => {
    const id = ++nextID;
    pending.set(id, { resolve, reject });
    socket.send(JSON.stringify({ id, method, params, ...(sessionId ? { sessionId } : {}) }));
  });
  const { targetId } = await send("Target.createTarget", { url: "about:blank" });
  const { sessionId } = await send("Target.attachToTarget", { targetId, flatten: true });
  const call = (method, params = {}) => send(method, params, sessionId);
  const evaluate = async expression => {
    try {
      const result = await call("Runtime.evaluate", { expression, returnByValue: true, awaitPromise: true });
      if (result.exceptionDetails) throw new Error(JSON.stringify(result.exceptionDetails));
      return result.result.value;
    } catch (error) {
      throw new Error(`evaluate failed: ${expression}\n${error instanceof Error ? error.message : String(error)}`);
    }
  };
  const wait = ms => new Promise(resolve => setTimeout(resolve, ms));
  const until = async expression => {
    for (let index = 0; index < 160; index += 1) {
      if (await evaluate(expression)) return;
      await wait(50);
    }
    throw new Error(`Timeout: ${expression}\n${await evaluate("document.body.innerText")}`);
  };
  const clickText = async text => {
    const clicked = await evaluate(`(()=>{const button=Array.from(document.querySelectorAll('button')).find(item=>item.textContent.trim()===${JSON.stringify(text)});if(!button)return false;button.click();return true;})()`);
    assert.equal(clicked, true, `button ${text} must exist in the real UI`);
    await wait(80);
  };
  const setInput = async (selector, value) => {
    await evaluate(`(()=>{const input=document.querySelector(${JSON.stringify(selector)});if(!input)throw new Error('missing '+${JSON.stringify(selector)});const proto=input instanceof HTMLSelectElement?HTMLSelectElement.prototype:HTMLInputElement.prototype;Object.getOwnPropertyDescriptor(proto,'value').set.call(input,${JSON.stringify(value)});input.dispatchEvent(new Event('input',{bubbles:true}));input.dispatchEvent(new Event('change',{bubbles:true}));})()`);
    await wait(50);
  };

  await call("Emulation.setDeviceMetricsOverride", { width: 390, height: 844, deviceScaleFactor: 1, mobile: true });
  const callback = `${origin}/auth/callback?token=${encodeURIComponent(token)}&user_id=${encodeURIComponent(profile.id)}&email=${encodeURIComponent(profile.email)}&name=${encodeURIComponent(profile.name)}&expires_at=${encodeURIComponent(login.expires_at)}`;
  await call("Page.navigate", { url: callback });
  await until(`document.body.textContent.includes('Giao dịch gần đây') || document.body.textContent.includes('Tổng quan')`);
  await call("Page.navigate", { url: `${origin}/transactions` });
  await until(`document.body.textContent.includes('Giao dịch') && !!document.querySelector('[aria-label="Tùy chọn giao dịch"]')`);
  await evaluate(`document.querySelector('[aria-label="Tùy chọn giao dịch"]').click()`);
  await clickText("Chuyển tiền đến ví khác");
  await until(`!!document.querySelector('[role="dialog"][aria-label="Thêm giao dịch"]') && document.body.textContent.includes('Chuyển ví')`);
  await clickText("Chuyển ví");
  await until(`!!document.querySelector('[aria-label="Ví chuyển đi"]') && !!document.querySelector('[aria-label="Ví nhận"]')`);
  assert.ok(await evaluate(`document.body.textContent.includes(${JSON.stringify(source.name)})`), `source wallet must be rendered in the transfer UI: ${await evaluate("document.body.innerText")}`);
  assert.ok(await evaluate(`document.body.textContent.includes(${JSON.stringify(destination.name)})`), `destination wallet must be rendered in the transfer UI: ${await evaluate("document.body.innerText")}`);
  await evaluate(`Array.from(document.querySelectorAll('[aria-label="Ví chuyển đi"] option')).find(option=>option.textContent===${JSON.stringify(source.name)}).selected=true`);
  await evaluate(`document.querySelector('[aria-label="Ví chuyển đi"]').dispatchEvent(new Event('change',{bubbles:true}))`);
  await evaluate(`Array.from(document.querySelectorAll('[aria-label="Ví nhận"] option')).find(option=>option.textContent===${JSON.stringify(destination.name)}).selected=true`);
  await evaluate(`document.querySelector('[aria-label="Ví nhận"]').dispatchEvent(new Event('change',{bubbles:true}))`);
  await setInput('[aria-label="Số tiền"]', "25000");
  await setInput('[aria-label="Ghi chú chuyển ví"]', "UI E2E transfer");
  await clickText("Lưu");
  await until(`!document.querySelector('[role="dialog"][aria-label="Thêm giao dịch"]')`);
  await until(`document.body.textContent.includes('UI E2E transfer')`);
  const transactions = await request("/transactions");
  const rows = transactions.filter(row => row.note === "UI E2E transfer");
  assert.equal(rows.length, 2, "real UI save must persist two rows");
  assert.equal(rows[0].transfer_id, rows[1].transfer_id);
  createdTransactions.push(...rows.map(row => row.id));
  const wallets = await request("/wallets");
  assert.equal(wallets.find(row => row.id === source.id).current_balance, 75000);
  assert.equal(wallets.find(row => row.id === destination.id).current_balance, 75000);
  const screenshot = await call("Page.captureScreenshot", { format: "png" });
  await writeFile(process.env.TRANSFER_E2E_SCREENSHOT ?? "/tmp/mypocket-transfer-e2e.png", Buffer.from(screenshot.data, "base64"));
  console.log("PASS: real Chromium UI — transfer mode rendered, submitted, refreshed, and matched persisted API state.");
} finally {
  socket?.close();
  if (chrome && chrome.exitCode === null) {
    chrome.kill();
    await new Promise(resolve => chrome.once("exit", resolve));
  }
  await rm(chromeProfile, { recursive: true, force: true });
  for (const id of createdTransactions) await request(`/transactions/${id}`, "DELETE", undefined, 204);
  for (const id of createdWallets) await request(`/wallets/${id}`, "DELETE", undefined, 204);
  console.log("Cleanup: only UI E2E rows and wallets created by this run were removed.");
}
