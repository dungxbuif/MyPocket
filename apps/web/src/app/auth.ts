import { apiFetch } from "./apiClient";

export type AuthState =
  | { status: "loading" }
  | { status: "unauthenticated" }
  | { status: "authenticated"; user: CurrentUser };

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
  } catch {
    return { status: "unauthenticated" };
  }
}

export async function logout() {
  await apiFetch("/api/v1/auth/logout", { method: "POST" });
}
