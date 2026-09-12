import { beforeEach, describe, expect, it, vi } from "vitest";

const { apiFetch } = vi.hoisted(() => ({ apiFetch: vi.fn() }));
vi.mock("./apiClient", () => ({ apiFetch }));

import { loadDashboard } from "./analytics";
import { loadAssets } from "./portfolio";
import { clearUserDataCaches, userDataCacheKey } from "./userDataCache";

describe("user-owned data cache", () => {
  beforeEach(() => {
    localStorage.clear();
    apiFetch.mockReset();
  });

  it("never serves an offline dashboard cached by another user", async () => {
    localStorage.setItem(userDataCacheKey("user-a", "dashboard"), JSON.stringify({ net_worth_vnd: 120000 }));
    apiFetch.mockRejectedValue(new TypeError("Network unavailable"));

    await expect(loadDashboard("user-b")).rejects.toThrow("Network unavailable");
    await expect(loadDashboard("user-a")).resolves.toMatchObject({ net_worth_vnd: 120000 });
  });

  it("uses an owner-scoped portfolio cache and clears it on logout", async () => {
    localStorage.setItem(userDataCacheKey("user-a", "assets"), JSON.stringify([{ id: "asset-a" }]));
    apiFetch.mockRejectedValue(new TypeError("Network unavailable"));

    await expect(loadAssets("user-b")).rejects.toThrow("Network unavailable");
    await expect(loadAssets("user-a")).resolves.toEqual([{ id: "asset-a" }]);

    clearUserDataCaches("user-a");
    expect(localStorage.getItem(userDataCacheKey("user-a", "assets"))).toBeNull();
  });

  it("does not hide an authorization failure behind a cached dashboard", async () => {
    localStorage.setItem(userDataCacheKey("user-a", "dashboard"), JSON.stringify({ net_worth_vnd: 120000 }));
    apiFetch.mockRejectedValue(new Error("Authentication required"));
    await expect(loadDashboard("user-a")).rejects.toThrow("Authentication required");
  });

  it("returns the fresh API result even when browser storage is full", async () => {
    const fresh = { net_worth_vnd: 200000 };
    apiFetch.mockResolvedValue({ report: fresh });
    const spy = vi.spyOn(Storage.prototype, "setItem").mockImplementation(() => { throw new DOMException("Full", "QuotaExceededError"); });
    try { await expect(loadDashboard("user-a")).resolves.toEqual(fresh); }
    finally { spy.mockRestore(); }
  });
});
