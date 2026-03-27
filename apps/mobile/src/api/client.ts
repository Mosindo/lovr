import { Platform } from "react-native";
import { getAccessToken } from "../store/tokenStore";
import { showGlobalError } from "../shared/feedback";

export type ApiErrorPayload = {
  error?: string;
};

function normalizeBaseUrl(value: string): string {
  return value.replace(/\/+$/, "");
}

function resolveApiBaseUrl(): string {
  const explicitBaseUrl = process.env.EXPO_PUBLIC_API_URL?.trim();
  if (explicitBaseUrl) {
    return normalizeBaseUrl(explicitBaseUrl);
  }

  if (Platform.OS === "android") {
    return "http://10.0.2.2:18080";
  }

  return "http://localhost:18080";
}

export const API_BASE_URL = resolveApiBaseUrl();

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
  return getAccessToken();
}

function mergeHeaders(base: HeadersInit | undefined, injected?: Record<string, string>): Headers {
  const headers = new Headers();

  headers.set("Accept", "application/json");

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

function shouldSetJsonContentType(body: BodyInit | null | undefined): boolean {
  return body !== undefined && body !== null && !(body instanceof FormData);
}

function formatNetworkHint(): string {
  if (process.env.EXPO_PUBLIC_API_URL) {
    return "Network error. Check your connection and API availability, then try again.";
  }

  if (Platform.OS === "android") {
    return "Network error. Android emulators usually need 10.0.2.2, and physical devices need EXPO_PUBLIC_API_URL set to your LAN IP.";
  }

  if (Platform.OS === "ios") {
    return "Network error. iOS simulators can use localhost, but physical devices need EXPO_PUBLIC_API_URL set to your LAN IP.";
  }

  return "Network error. Check your connection and try again.";
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
    if (!headers.has("Content-Type") && shouldSetJsonContentType(options?.body)) {
      headers.set("Content-Type", "application/json");
    }
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
      let message = apiPayload?.error ?? `API request failed (${response.status})`;
      if (response.status === 401) {
        message = apiPayload?.error ?? "Your session is no longer valid. Please sign in again.";
      }
      if (response.status >= 500 || response.status === 401) {
        showGlobalError(message);
      }
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
      showGlobalError("The request timed out. Please try again.");
      throw new Error("API request timeout (10s)");
    }

    if (err instanceof TypeError) {
      showGlobalError(formatNetworkHint());
    }
    throw err;
  } finally {
    clearTimeout(timeoutId);
  }
}
