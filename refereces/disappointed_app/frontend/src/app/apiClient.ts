export type APIErrorCode =
  | "AUTH_REQUIRED"
  | "FORBIDDEN"
  | "VALIDATION_FAILED"
  | "CSRF_REQUIRED"
  | "INTERNAL_RETRYABLE"
  | "INTERNAL_FAILURE"
  | "NOT_FOUND"
  | "VERSION_CONFLICT"
  | "RATE_LIMITED"
  | "RECENT_AUTH_REQUIRED"
  | "CAPABILITY_UNAVAILABLE"
  | "IDEMPOTENCY_CONFLICT";

export class APIClientError extends Error {
  constructor(
    readonly code: APIErrorCode,
    message: string,
    readonly correlationID?: string,
  ) {
    super(message);
  }
}

export async function apiFetch<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers = new Headers(options.headers);
  const method = options.method?.toUpperCase() ?? "GET";
  if (method !== "GET" && method !== "HEAD") {
    const token = readCookie("mypocket_csrf");
    if (token) headers.set("X-CSRF-Token", token);
  }

  const response = await fetch(`${apiBaseURL()}${path}`, {
    ...options,
    headers,
    credentials: "include",
  });
  const body = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw new APIClientError(body.error?.code ?? "INTERNAL_FAILURE", body.error?.message ?? "Request failed", body.correlation_id);
  }
  return body as T;
}

export function apiBaseURL() {
  return import.meta.env.VITE_API_BASE_URL ?? "";
}

function readCookie(name: string) {
  return document.cookie
    .split(";")
    .map((part) => part.trim())
    .find((part) => part.startsWith(`${name}=`))
    ?.slice(name.length + 1);
}
