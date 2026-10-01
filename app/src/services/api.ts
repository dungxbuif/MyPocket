import { API_ROUTES, APP_CONFIG, STORAGE_KEYS } from "../config/app";

export type ApiErrorCode = "NETWORK" | "HTTP_ERROR" | "INVALID_RESPONSE" | "UNAUTHORIZED";

export class ApiError extends Error {
  public readonly code: ApiErrorCode;
  public readonly status: number;
  public readonly payload: unknown;

  constructor(code: ApiErrorCode, status: number, message: string, payload: unknown) {
    super(message);
    this.code = code;
    this.status = status;
    this.payload = payload;
  }
}

export type ApiResult<T> = T;

export async function apiRequest<T>(path: string, options: RequestInit = {}, token?: string): Promise<T> {
  const url = `${APP_CONFIG.API_BASE_URL}${path}`;
  const headers = new Headers(options.headers);
  const method = options.method?.toUpperCase() ?? "GET";
  let authToken = token;
  if (authToken) {
    headers.set("Authorization", `Bearer ${authToken}`);
  }
  const isFormData = typeof FormData !== "undefined" && options.body instanceof FormData;
  if (method !== "GET" && method !== "HEAD" && !isFormData && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  let response: Response;
  try {
    response = await fetch(url, {
      ...options,
      headers,
      credentials: "include",
      redirect: "follow",
    });
  } catch (error) {
    throw new ApiError("NETWORK", 0, error instanceof Error ? error.message : "Lỗi kết nối", null);
  }

  const contentType = response.headers.get("content-type") ?? "";
  const isJson = contentType.includes("application/json") || contentType.includes("application/problem+json");
  const rawText = await response.text();
  const body = isJson ? parseJsonResponse(rawText) : rawText;

  if (!response.ok && response.status === 401 && path !== API_ROUTES.REFRESH && authToken && !authToken.startsWith("mpk_")) {
    const refreshed = await refreshAccessToken();
    if (refreshed) {
      authToken = refreshed;
      headers.set("Authorization", `Bearer ${refreshed}`);
      response = await fetch(url, { ...options, headers, credentials: "include", redirect: "follow" });
      const retryContentType = response.headers.get("content-type") ?? "";
      const retryIsJson = retryContentType.includes("application/json") || retryContentType.includes("application/problem+json");
      const retryRawText = await response.text();
      const retryBody = retryIsJson ? parseJsonResponse(retryRawText) : retryRawText;
      if (response.ok) return unwrapResponse<T>(retryBody, retryRawText, retryIsJson);
      return throwApiError(response, retryBody);
    }
  }

  if (!response.ok) {
    if (response.status === 401) {
      throw new ApiError("UNAUTHORIZED", response.status, extractErrorMessage(body), body);
    }
    throw new ApiError("HTTP_ERROR", response.status, extractErrorMessage(body), body);
  }

  return unwrapResponse<T>(body, rawText, isJson);
}

function unwrapResponse<T>(body: unknown, rawText: string, isJson: boolean): T {
  if (!isJson) return (rawText.length === 0 ? null : rawText) as T;
  if (body === null) throw new ApiError("INVALID_RESPONSE", 200, "Phản hồi JSON không hợp lệ", body);
  if (typeof body === "object" && body !== null && "data" in body) return (body as { data: T }).data;
  return body as T;
}

function throwApiError(response: Response, body: unknown): never {
  const code: ApiErrorCode = response.status === 401 ? "UNAUTHORIZED" : "HTTP_ERROR";
  throw new ApiError(code, response.status, extractErrorMessage(body), body);
}

async function refreshAccessToken(): Promise<string | null> {
	if (refreshPromise) return refreshPromise;
	refreshPromise = performRefreshAccessToken();
	try {
		return await refreshPromise;
	} finally {
		refreshPromise = null;
	}
}

let refreshPromise: Promise<string | null> | null = null;

async function performRefreshAccessToken(): Promise<string | null> {
  const raw = localStorage.getItem(STORAGE_KEYS.AUTH_SESSION);
  if (!raw) return null;
  let session: Record<string, unknown>;
  try {
    session = JSON.parse(raw) as Record<string, unknown>;
  } catch {
    return null;
  }
  const refreshToken = typeof session.refreshToken === "string" ? session.refreshToken : "";
  if (!refreshToken) return null;
  try {
    const response = await fetch(`${APP_CONFIG.API_BASE_URL}${API_ROUTES.REFRESH}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "include",
      body: JSON.stringify({ refresh_token: refreshToken }),
    });
    const contentType = response.headers.get("content-type") ?? "";
    const payload = contentType.includes("json") ? parseJsonResponse(await response.text()) : null;
    if (!response.ok || !payload || typeof payload !== "object" || !("data" in payload)) return null;
    const next = (payload as { data: Record<string, unknown> }).data;
    if (typeof next.token !== "string" || typeof next.refresh_token !== "string") return null;
    localStorage.setItem(STORAGE_KEYS.AUTH_SESSION, JSON.stringify({
      ...session,
      token: next.token,
      expiresAt: typeof next.expires_at === "string" ? next.expires_at : session.expiresAt,
      refreshToken: next.refresh_token,
      refreshExpiresAt: typeof next.refresh_expires_at === "string" ? next.refresh_expires_at : session.refreshExpiresAt,
    }));
    return next.token;
  } catch {
    return null;
  }
}

function parseJsonResponse(input: string): unknown {
  if (!input) return null;
  try {
    return JSON.parse(input);
  } catch {
    return null;
  }
}

function extractErrorMessage(payload: unknown): string {
  if (!payload) {
    return "Yêu cầu API thất bại";
  }
  if (typeof payload === "object" && payload !== null && "error" in payload) {
    const maybeMessage = (payload as Record<string, unknown>).error;
    if (typeof maybeMessage === "string") {
      return maybeMessage;
    }
  }
  if (typeof payload === "object" && payload !== null && "detail" in payload) {
    const detail = (payload as Record<string, unknown>).detail;
    if (typeof detail === "string") return detail;
  }
  if (typeof payload === "string") return payload;
  return "Yêu cầu API thất bại";
}
