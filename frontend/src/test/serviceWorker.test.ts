import { readFileSync } from "node:fs";
import { runInNewContext } from "node:vm";
import { expect, it, vi } from "vitest";

it("does not intercept private docs, API data or another origin", () => {
  const listeners: Record<string, (event: unknown) => void> = {};
  runInNewContext(readFileSync("public/sw.js", "utf8"), {
    self: { location: { origin: "https://pocket.test" }, addEventListener: (name: string, listener: (event: unknown) => void) => { listeners[name] = listener; } },
    URL, fetch: vi.fn().mockResolvedValue(new Response()),
  });
  for (const url of ["https://pocket.test/docs/", "https://pocket.test/api/v1/me", "https://other.test/private"]) {
    const respondWith = vi.fn();
    listeners.fetch({ request: { url, method: "GET" }, respondWith });
    expect(respondWith, url).not.toHaveBeenCalled();
  }
});

it("serves the installed shell without a network request for known-offline navigation", async () => {
  const listeners: Record<string, (event: unknown) => void> = {};
  const fetch = vi.fn();
  const cached = new Response("installed shell");
  runInNewContext(readFileSync("public/sw.js", "utf8"), {
    self: { location: { origin: "https://pocket.test" }, navigator: { onLine: false }, addEventListener: (name: string, listener: (event: unknown) => void) => { listeners[name] = listener; } },
    URL, fetch, caches: { match: vi.fn().mockResolvedValue(cached) },
  });
  const respondWith = vi.fn();
  listeners.fetch({ request: { url: "https://pocket.test/", method: "GET", mode: "navigate" }, respondWith });
  await expect(respondWith.mock.calls[0][0]).resolves.toBe(cached);
  expect(fetch).not.toHaveBeenCalled();
});
