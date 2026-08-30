import { apiFetch } from "./apiClient";
import { APIClientError } from "./apiClient";

export type AuthState =
  | { status: "loading" }
  | { status: "unauthenticated" }
  | { status: "authenticated"; user: CurrentUser }
  | { status: "forbidden"; message: string; correlationID?: string };

export type CurrentUser = {
  id: string;
  email: string;
  email_verified: boolean;
  display_name: string;
  avatar_url: string;
};

const OFFLINE_USER_KEY = "mypocket.current-user.v1";

export async function loadCurrentUser(): Promise<AuthState> {
  try {
    const response = await apiFetch<{ user: CurrentUser }>("/api/v1/me");
    cacheCurrentUser(response.user);
    return { status: "authenticated", user: response.user };
  } catch (error) {
    if (error instanceof APIClientError && error.code === "FORBIDDEN") {
      localStorage.removeItem(OFFLINE_USER_KEY);
      return { status: "forbidden", message: error.message, correlationID: error.correlationID };
    }
    if (!navigator.onLine) {
      const cached = readCachedCurrentUser();
      if (cached) return { status: "authenticated", user: cached };
    }
    return { status: "unauthenticated" };
  }
}

export async function logout() {
  try {
    await apiFetch("/api/v1/auth/logout", { method: "POST" });
  } finally {
    localStorage.removeItem(OFFLINE_USER_KEY);
  }
}

function cacheCurrentUser(user: CurrentUser) {
  localStorage.setItem(OFFLINE_USER_KEY, JSON.stringify(user));
}

function readCachedCurrentUser(): CurrentUser | null {
  try {
    const parsed = JSON.parse(localStorage.getItem(OFFLINE_USER_KEY) ?? "null") as Partial<CurrentUser> | null;
    if (!parsed || typeof parsed.id !== "string" || typeof parsed.email !== "string") return null;
    return {
      id: parsed.id,
      email: parsed.email,
      email_verified: Boolean(parsed.email_verified),
      display_name: typeof parsed.display_name === "string" ? parsed.display_name : parsed.email,
      avatar_url: typeof parsed.avatar_url === "string" ? parsed.avatar_url : "",
    };
  } catch {
    localStorage.removeItem(OFFLINE_USER_KEY);
    return null;
  }
}
