type ApiErrorPayload = {
  error?: string;
};

const API_BASE_URL = process.env.EXPO_PUBLIC_API_URL ?? "http://localhost:18080";
const API_TIMEOUT_MS = 10000;

export async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), API_TIMEOUT_MS);
  let response: Response;
  try {
    response = await fetch(`${API_BASE_URL}${path}`, {
      ...init,
      signal: controller.signal,
      headers: {
        "Content-Type": "application/json",
        ...(init?.headers ?? {})
      }
    });
  } catch (err) {
    clearTimeout(timeoutId);
    if (err instanceof Error && err.name === "AbortError") {
      throw new Error("request timeout: check API URL/network");
    }
    throw new Error("network request failed");
  }
  clearTimeout(timeoutId);

  if (!response.ok) {
    let payload: ApiErrorPayload | null = null;
    try {
      payload = (await response.json()) as ApiErrorPayload;
    } catch {
      payload = null;
    }
    throw new Error(payload?.error ?? `request failed (${response.status})`);
  }

  return (await response.json()) as T;
}
