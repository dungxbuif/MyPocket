import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import { test } from "node:test";

const manifest = JSON.parse(await readFile(new URL("../public/manifest.webmanifest", import.meta.url), "utf8"));
const index = await readFile(new URL("../index.html", import.meta.url), "utf8");
const main = await readFile(new URL("../src/main.tsx", import.meta.url), "utf8");
const shell = await readFile(new URL("../src/atomic/templates/MobileAppShell.tsx", import.meta.url), "utf8");
const prompt = await readFile(new URL("../src/atomic/organisms/PwaInstallPrompt.tsx", import.meta.url), "utf8").catch(() => "");
const worker = await readFile(new URL("../public/sw.js", import.meta.url), "utf8").catch(() => "");

test("PWA manifest and install surfaces are complete", () => {
  assert.equal(manifest.display, "standalone");
  assert.equal(manifest.start_url, "/");
  assert.ok(Array.isArray(manifest.icons) && manifest.icons.length > 0);
  assert.match(index, /manifest\.webmanifest/);
  assert.match(index, /viewport-fit=cover/);
  assert.match(main, /serviceWorker\.register/);
  assert.match(shell, /PwaInstallPrompt/);
  assert.match(prompt, /beforeinstallprompt/);
  assert.match(prompt, /\.prompt\(\)/);
  assert.match(worker, /addEventListener\(["']fetch["']/);
});
