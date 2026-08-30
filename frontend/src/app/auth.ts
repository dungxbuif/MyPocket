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

export async function loadCurrentUser(): Promise<AuthState> {
  try {
    const response = await apiFetch<{ user: CurrentUser }>("/api/v1/me");
    return { status: "authenticated", user: response.user };
  } catch (error) {
    if (error instanceof APIClientError && error.code === "FORBIDDEN") {
      return { status: "forbidden", message: error.message, correlationID: error.correlationID };
    }
    return { status: "unauthenticated" };
  }
}

export async function logout() {
  await apiFetch("/api/v1/auth/logout", { method: "POST" });
}
