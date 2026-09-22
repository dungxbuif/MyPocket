import { API_ROUTES, APP_ROUTES, STORAGE_KEYS } from "../config/app";
import { apiRequest } from "./api";

export type UserProfile = {
  id: string;
  name: string;
  email: string;
  timezone: string;
  timezone_confirmed: boolean;
  created_at: string;
};

export type LoginOutput = {
  token: string;
  expires_at: string;
  user: UserProfile;
};

export type HomeOutput = {
  greeting: string;
  stats: {
    wallet_count: number;
    notes_count: number;
  };
  notice: string;
};

export type AuthSession = {
  token: string;
  user: UserProfile;
  expiresAt: string;
};

export function readSession(): AuthSession | null {
  const raw = localStorage.getItem(STORAGE_KEYS.AUTH_SESSION);
  if (!raw) {
    return null;
  }
  try {
    const parsed = JSON.parse(raw) as Partial<AuthSession>;
    if (!parsed || typeof parsed.token !== "string" || typeof parsed.user?.id !== "string") return null;
    return {
      token: parsed.token,
      user: parsed.user as UserProfile,
      expiresAt: typeof parsed.expiresAt === "string" ? parsed.expiresAt : "",
    };
  } catch {
    localStorage.removeItem(STORAGE_KEYS.AUTH_SESSION);
    return null;
  }
}

export function writeSession(session: AuthSession): void {
  localStorage.setItem(STORAGE_KEYS.AUTH_SESSION, JSON.stringify(session));
}

export function clearSession(): void {
  localStorage.removeItem(STORAGE_KEYS.AUTH_SESSION);
}

export function getStoredToken(): string {
  return readSession()?.token ?? "";
}

export function isAuthCallbackPath(pathname: string): boolean {
  return pathname === APP_ROUTES.AUTH_CALLBACK_PATH || pathname === API_ROUTES.GOOGLE_CALLBACK;
}

export async function startGoogleLogin(): Promise<LoginOutput | null> {
  const data = await apiRequest<unknown>(API_ROUTES.START_GOOGLE_AUTH, {
    method: "GET",
  });
  if (!data || typeof data !== "object") return null;
  const candidate = data as Partial<LoginOutput>;
  if (typeof candidate.token !== "string" || typeof candidate.user?.email !== "string") {
    return null;
  }
  return {
    token: candidate.token,
    expires_at: typeof candidate.expires_at === "string" ? candidate.expires_at : "",
    user: {
      id: String(candidate.user.id ?? ""),
      name: String(candidate.user.name ?? ""),
      email: candidate.user.email,
      timezone: String(candidate.user.timezone ?? "Asia/Ho_Chi_Minh"),
      timezone_confirmed: Boolean(candidate.user.timezone_confirmed),
      created_at: String(candidate.user.created_at ?? ""),
    },
  };
}

export async function loginWithGoogleCode(code: string): Promise<LoginOutput> {
  const url = `${API_ROUTES.GOOGLE_CALLBACK}?code=${encodeURIComponent(code)}`;
  return apiRequest<LoginOutput>(url, {
    method: "GET",
  });
}

export async function loginWithGoogleFixture(query: URLSearchParams): Promise<LoginOutput> {
  const subject = query.get("subject")?.trim() ?? "";
  const email = query.get("email")?.trim() ?? "";
  const emailVerified = query.get("email_verified")?.trim() ?? "true";
  const name = query.get("name")?.trim() ?? "";
  const avatarURL = query.get("avatar_url")?.trim() ?? "";

  if (!subject || !email) {
    throw new Error("Thiếu dữ liệu callback cho login giả lập");
  }

  const url = `${API_ROUTES.GOOGLE_CALLBACK}?subject=${encodeURIComponent(subject)}&email=${encodeURIComponent(email)}&email_verified=${encodeURIComponent(emailVerified)}&name=${encodeURIComponent(name)}&avatar_url=${encodeURIComponent(avatarURL)}`;
  return apiRequest<LoginOutput>(url, {
    method: "GET",
  });
}

export function extractFixtureLoginFromQueryParams(query: URLSearchParams): LoginOutput | null {
  const token = query.get("token")?.trim() ?? "";
  if (!token) return null;

  const user: UserProfile = {
    id: query.get("user_id")?.trim() ?? query.get("id")?.trim() ?? "",
    name: query.get("name")?.trim() ?? "",
    email: query.get("email")?.trim() ?? "",
    timezone: "Asia/Ho_Chi_Minh",
    timezone_confirmed: false,
    created_at: query.get("created_at")?.trim() ?? "",
  };

  if (!user.id || !user.email) {
    return null;
  }

  return {
    token,
    expires_at: query.get("expires_at")?.trim() ?? "",
    user,
  };
}

export async function fetchProfile(token?: string): Promise<UserProfile> {
  return apiRequest<UserProfile>(API_ROUTES.PROFILE, {}, token);
}

export async function updateAccountTimezone(timezone: string, initializeOnly = false, token?: string): Promise<UserProfile> {
  return apiRequest<UserProfile>(API_ROUTES.PROFILE, {
    method: "PATCH",
    body: JSON.stringify({ timezone, initialize_only: initializeOnly }),
  }, token);
}

export async function initializeAccountTimezone(profile: UserProfile, token: string): Promise<UserProfile> {
  if (profile.timezone_confirmed) return profile;
  const browserTimezone = Intl.DateTimeFormat().resolvedOptions().timeZone;
  if (!browserTimezone) return profile;
  try {
    return await updateAccountTimezone(browserTimezone, true, token);
  } catch {
    return profile;
  }
}

export async function fetchHome(token?: string): Promise<HomeOutput> {
  return apiRequest<HomeOutput>(API_ROUTES.HOME, {}, token);
}
