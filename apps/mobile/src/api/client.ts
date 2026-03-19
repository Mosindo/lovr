import { getToken as getTokenFromStore } from "../store/tokenStore";

export type ApiErrorPayload = {
  error?: string;
};

export const API_BASE_URL = process.env.EXPO_PUBLIC_API_URL ?? "http://localhost:18080";

export class ApiError extends Error {
  public status: number;
  public data: unknown;

  constructor(status: number, data: unknown, message?: string) {
    super(message ?? `API request failed (${status})`);
    this.name = "ApiError";
    this.status = status;
    this.data = data;
  }
}

// Placeholder indirection for JWT injection.
async function getToken(): Promise<string | null> {
  return getTokenFromStore();
}

function mergeHeaders(base: HeadersInit | undefined, injected?: Record<string, string>): Headers {
  const headers = new Headers();

  // Ensure consistent JSON API behavior.
  headers.set("Content-Type", "application/json");

  if (base) {
    new Headers(base).forEach((value, key) => {
      headers.set(key, value);
    });
  }

  if (injected) {
    for (const [k, v] of Object.entries(injected)) {
      if (!headers.has(k)) {
        headers.set(k, v);
      }
    }
  }

  return headers;
}

export async function apiRequest<T = unknown>(path: string, options?: RequestInit): Promise<T> {
  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), 10_000);

  const token = await getToken();
  const authHeader = token ? { Authorization: `Bearer ${token}` } : undefined;

  try {
    if (options?.signal) {
      if (options.signal.aborted) {
        controller.abort();
      } else {
        options.signal.addEventListener("abort", () => controller.abort(), { once: true });
      }
    }

    const headers = mergeHeaders(options?.headers, authHeader);
    const response = await fetch(`${API_BASE_URL}${path}`, {
      ...options,
      headers,
      signal: controller.signal
    });

    if (!response.ok) {
      const contentType = response.headers.get("content-type") ?? "";
      let data: unknown = null;
      try {
        data = contentType.includes("application/json") ? await response.json() : await response.text();
      } catch {
        data = null;
      }

      const apiPayload = data as ApiErrorPayload | null;
      const message = apiPayload?.error ?? `API request failed (${response.status})`;
      throw new ApiError(response.status, data, message);
    }

    // Some endpoints could return empty bodies (204); keep it predictable.
    const text = await response.text();
    if (!text) {
      return undefined as T;
    }
    return JSON.parse(text) as T;
  } catch (err) {
    const name = (err as { name?: string }).name;
    if (name === "AbortError") {
      throw new Error("API request timeout (10s)");
    }
    throw err;
  } finally {
    clearTimeout(timeoutId);
  }
}

