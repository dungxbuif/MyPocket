import { afterEach, describe, expect, it, vi } from "vitest";
import { loadCurrentUser } from "./auth";

describe("loadCurrentUser", () => {
  it("rejects an expired session even when navigator reports offline", async () => {
    localStorage.setItem("mypocket.current-user.v1", JSON.stringify({ id: "a", email: "a@example.com" }));
    Object.defineProperty(navigator, "onLine", { configurable: true, value: false });
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(JSON.stringify({ error: { code: "AUTH_REQUIRED" } }), { status: 401 })));
    await expect(loadCurrentUser()).resolves.toEqual({ status: "unauthenticated" });
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("uses the cached authenticated user after a transport failure even when the browser reports online", async () => {
    localStorage.setItem("mypocket.current-user.v1", JSON.stringify({
      id: "user_123",
      email: "offline@example.com",
      email_verified: true,
      display_name: "Offline User",
      avatar_url: "",
    }));
    Object.defineProperty(navigator, "onLine", { configurable: true, value: true });
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new TypeError("network unavailable")));

    await expect(loadCurrentUser()).resolves.toMatchObject({
      status: "authenticated",
      user: { id: "user_123", email: "offline@example.com" },
    });
  });
});
