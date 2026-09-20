import { APP_CONFIG } from "../config/app";

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
  const authToken = token;
  if (authToken) {
    headers.set("Authorization", `Bearer ${authToken}`);
  }
  if (method !== "GET" && method !== "HEAD" && !headers.has("Content-Type")) {
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

  if (!response.ok) {
    if (response.status === 401) {
      throw new ApiError("UNAUTHORIZED", response.status, extractErrorMessage(body), body);
    }
    throw new ApiError("HTTP_ERROR", response.status, extractErrorMessage(body), body);
  }

  if (!isJson) {
    if (rawText.length === 0) {
      return null as T;
    }
    return rawText as unknown as T;
  }

  if (body === null) {
    throw new ApiError("INVALID_RESPONSE", response.status, "Phản hồi JSON không hợp lệ", body);
  }
  if (typeof body === "object" && body !== null && "data" in body) {
    return (body as { data: T }).data;
  }
  return body as T;
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
