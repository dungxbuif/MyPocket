// Real Chromium E2E: authenticate through the local callback, open the
// floating feedback bubble, and verify screenshot-storage fallback persists
// the text feedback through the real Vite proxy and API.
import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";
import { mkdtemp, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import path from "node:path";
import { spawn } from "node:child_process";

const origin = process.env.TEST_API_ORIGIN ?? "http://127.0.0.1:4173";
const api = `${origin}/api/v1`;
const dbURL = process.env.TEST_DATABASE_URL ?? "postgres://dev:password@127.0.0.1:5432/postgres?sslmode=disable";
let token = "";
let createdTitle = "";

async function request(pathname, method = "GET", body, expected = 200) {
  const response = await fetch(`${api}${pathname}`, {
    method,
    headers: { ...(body ? { "Content-Type": "application/json" } : {}), ...(token ? { Authorization: `Bearer ${token}` } : {}) },
    ...(body ? { body: JSON.stringify(body) } : {}),
  });
  const payload = response.status === 204 ? null : await response.json();
  assert.equal(response.status, expected, `${method} ${pathname}: ${JSON.stringify(payload)}`);
  return payload?.data;
}

const login = await request("/login", "POST", {
  email: process.env.TEST_EMAIL ?? "admin@mypocket.local",
  password: process.env.TEST_PASSWORD ?? "12345678",
});
token = login.token;
const profile = await request("/auth/profile");
const suffix = Date.now().toString(36);
createdTitle = `UI web feedback ${suffix}`;

const chromeProfile = await mkdtemp(path.join(tmpdir(), "mypocket-feedback-browser-"));
const chromePath = process.env.CHROME_BIN ?? "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome";
let chrome;
let socket;
try {
  chrome = spawn(chromePath, [
    "--headless=new", "--no-first-run", "--no-default-browser-check",
    "--disable-background-timer-throttling", "--disable-renderer-backgrounding",
    "--remote-debugging-port=0", `--user-data-dir=${chromeProfile}`, "about:blank",
  ], { stdio: ["ignore", "ignore", "pipe"] });
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
  const send = (method, params = {}, sessionID) => new Promise((resolve, reject) => {
    const id = ++nextID;
    pending.set(id, { resolve, reject });
    socket.send(JSON.stringify({ id, method, params, ...(sessionID ? { sessionId: sessionID } : {}) }));
  });
  const { targetId } = await send("Target.createTarget", { url: "about:blank" });
  const { sessionId } = await send("Target.attachToTarget", { targetId, flatten: true });
  const call = (method, params = {}) => send(method, params, sessionId);
  const evaluate = async expression => {
    const result = await call("Runtime.evaluate", { expression, returnByValue: true, awaitPromise: true });
    if (result.exceptionDetails) throw new Error(JSON.stringify(result.exceptionDetails));
    return result.result.value;
  };
  const wait = ms => new Promise(resolve => setTimeout(resolve, ms));
  const until = async expression => {
    for (let index = 0; index < 160; index += 1) {
      if (await evaluate(expression)) return;
      await wait(50);
    }
    throw new Error(`Timeout: ${expression}\n${await evaluate("document.body.innerText")}`);
  };
  const click = async expression => {
    const clicked = await evaluate(expression);
    assert.equal(clicked, true, `click target must exist: ${expression}`);
    await wait(80);
  };
  const setInput = async (selector, value) => {
    await evaluate(`(()=>{const input=document.querySelector(${JSON.stringify(selector)});if(!input)throw new Error('missing '+${JSON.stringify(selector)});const proto=input instanceof HTMLTextAreaElement?HTMLTextAreaElement.prototype:HTMLInputElement.prototype;Object.getOwnPropertyDescriptor(proto,'value').set.call(input,${JSON.stringify(value)});input.dispatchEvent(new Event('input',{bubbles:true}));})()`);
    await wait(50);
  };

  await call("Emulation.setDeviceMetricsOverride", { width: 390, height: 844, deviceScaleFactor: 1, mobile: true });
  const callback = `${origin}/auth/callback?token=${encodeURIComponent(token)}&user_id=${encodeURIComponent(profile.id)}&email=${encodeURIComponent(profile.email)}&name=${encodeURIComponent(profile.name)}&expires_at=${encodeURIComponent(login.expires_at)}`;
  await call("Page.navigate", { url: callback });
  await until(`document.body.textContent.includes('Giao dịch gần đây') || document.body.textContent.includes('Tổng quan')`);

  await click(`(()=>{const button=document.querySelector('[aria-label="Gửi phản hồi"]');if(!button)return false;button.click();return true;})()`);
  await until(`!!document.querySelector('[role="dialog"][aria-label="Gửi phản hồi"]')`);
  await setInput('input[placeholder="Ví dụ: Số dư chưa cập nhật"]', createdTitle);
  await setInput('textarea[placeholder="Mô tả điều đã xảy ra hoặc điều bạn muốn cải thiện"]', "Ảnh có thể không lưu được nhưng nội dung phản hồi phải còn nguyên.");
  await click(`(()=>{const dialog=document.querySelector('[role="dialog"][aria-label="Gửi phản hồi"]');const button=Array.from(dialog?.querySelectorAll('button') ?? []).find(item=>item.textContent.trim()==='Gửi phản hồi');if(!button)return false;button.click();return true;})()`);
  await until(`document.body.textContent.includes('Đã gửi phản hồi kèm hiện trạng màn hình.')`);

  const rows = await request("/feedback");
  const saved = rows.find(item => item.title === createdTitle);
  assert.ok(saved, "the web feedback bubble must persist its text after screenshot fallback");
  assert.equal(saved.screenshot_available, false, "text-only fallback must not claim a stored screenshot");
  console.log("PASS: real Chromium feedback UI — capture, controlled storage error, JSON fallback and persisted owner row.");
} finally {
  socket?.close();
  if (chrome && chrome.exitCode === null) {
    chrome.kill();
    await new Promise(resolve => chrome.once("exit", resolve));
  }
  await rm(chromeProfile, { recursive: true, force: true });
  if (createdTitle) {
    const escaped = createdTitle.replaceAll("'", "''");
    execFileSync("psql", [dbURL, "-v", "ON_ERROR_STOP=1", "-c", `DELETE FROM feedback WHERE title = '${escaped}';`], { stdio: "ignore" });
  }
  console.log("Cleanup: only the browser feedback row created by this run was removed.");
}
