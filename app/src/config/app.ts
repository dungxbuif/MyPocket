export const APP_CONFIG = {
  API_BASE_URL: import.meta.env.VITE_API_BASE_URL?.trim() || "",
} as const;

export const APP_ROUTES = {
  AUTH_CALLBACK_PATH: "/auth/callback",
} as const;

export const API_ROUTES = {
  START_GOOGLE_AUTH: "/api/v1/auth/google",
  GOOGLE_CALLBACK: "/api/v1/auth/google/callback",
  PROFILE: "/api/v1/auth/profile",
  HOME: "/api/v1/home",
} as const;

export const STORAGE_KEYS = {
  AUTH_SESSION: "mypocket.auth.session.v1",
} as const;
