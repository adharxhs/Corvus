import { ApiError, type ApiErrorResponse } from "../types/api";
import { API_BASE_URL } from "../utils/constants";

let authToken: string | null = null;
let onUnauthorized: (() => void) | null = null;

export function setApiToken(token: string | null): void {
  authToken = token;
}

export function setOnUnauthorized(handler: (() => void) | null): void {
  onUnauthorized = handler;
}

export function isNetworkError(err: unknown): boolean {
  return err instanceof ApiError && err.status === 0;
}

export async function probeServer(timeoutMs = 5000): Promise<boolean> {
  if (typeof navigator !== "undefined" && navigator.onLine === false) {
    return false;
  }
  try {
    const controller = new AbortController();
    const timer = window.setTimeout(() => controller.abort(), timeoutMs);
    try {
      const response = await fetch(`${API_BASE_URL}/health`, {
        method: "GET",
        cache: "no-store",
        headers: { Accept: "application/json" },
        signal: controller.signal,
      });
      return response.ok;
    } finally {
      window.clearTimeout(timer);
    }
  } catch {
    return false;
  }
}

export async function apiRequest<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers = new Headers(options.headers);
  headers.set("Content-Type", "application/json");
  if (authToken) {
    headers.set("Authorization", `Bearer ${authToken}`);
  }

  if (typeof navigator !== "undefined" && navigator.onLine === false) {
    throw new ApiError("You're offline. Check your connection.", 0);
  }

  let response: Response;
  try {
    response = await fetch(`${API_BASE_URL}${path}`, {
      ...options,
      headers,
    });
  } catch {
    throw new ApiError("Unable to reach the server. Check your connection.", 0);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  const text = await response.text();
  let data: unknown;
  try {
    data = text ? JSON.parse(text) : undefined;
  } catch {
    data = undefined;
  }

  if (response.status === 401) {
    onUnauthorized?.();
  }

  if (!response.ok) {
    const retryAfter = response.headers.get("Retry-After");
    const message = isApiErrorResponse(data) ? data.error : "Request failed";
    throw new ApiError(message, response.status, retryAfter ? Number(retryAfter) : undefined);
  }

  return data as T;
}

function isApiErrorResponse(value: unknown): value is ApiErrorResponse {
  return typeof value === "object" && value !== null && typeof (value as ApiErrorResponse).error === "string";
}
