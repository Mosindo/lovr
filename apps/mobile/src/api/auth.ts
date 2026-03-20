import { apiRequest } from "./client";

export type AuthUser = {
  id: string;
  email: string;
  organizationId: string;
  createdAt: string;
};

export type AuthResponse = {
  token?: string;
  accessToken?: string;
  refreshToken: string;
  user: AuthUser;
};

export type AuthSession = {
  accessToken: string;
  refreshToken: string;
  user: AuthUser;
};

function normalizeSession(response: AuthResponse): AuthSession {
  const accessToken = response.accessToken ?? response.token;
  if (!accessToken) {
    throw new Error("authentication response did not include an access token");
  }

  return {
    accessToken,
    refreshToken: response.refreshToken,
    user: response.user
  };
}

export async function register(email: string, password: string): Promise<AuthSession> {
  const response = await apiRequest<AuthResponse>("/auth/register", {
    method: "POST",
    body: JSON.stringify({ email, password })
  });
  return normalizeSession(response);
}

export async function login(email: string, password: string): Promise<AuthSession> {
  const response = await apiRequest<AuthResponse>("/auth/login", {
    method: "POST",
    body: JSON.stringify({ email, password })
  });
  return normalizeSession(response);
}

export async function refreshSession(refreshToken: string): Promise<AuthSession> {
  const response = await apiRequest<AuthResponse>("/auth/refresh", {
    method: "POST",
    body: JSON.stringify({ refreshToken })
  });
  return normalizeSession(response);
}

export async function logout(refreshToken: string): Promise<void> {
  await apiRequest("/auth/logout", {
    method: "POST",
    body: JSON.stringify({ refreshToken })
  });
}

export async function me(): Promise<AuthUser> {
  return apiRequest<AuthUser>("/me", { method: "GET" });
}
