import { request } from "../utils/http";

export type AuthUser = {
  id: string;
  email: string;
  createdAt: string;
};

export type AuthResponse = {
  token: string;
  user: AuthUser;
};

function authHeaders(token: string): Record<string, string> {
  return { Authorization: `Bearer ${token}` };
}

export async function register(email: string, password: string): Promise<AuthResponse> {
  return request<AuthResponse>("/auth/register", {
    method: "POST",
    body: JSON.stringify({ email, password })
  });
}

export async function login(email: string, password: string): Promise<AuthResponse> {
  return request<AuthResponse>("/auth/login", {
    method: "POST",
    body: JSON.stringify({ email, password })
  });
}

export async function me(token: string): Promise<AuthUser> {
  return request<AuthUser>("/me", {
    method: "GET",
    headers: authHeaders(token)
  });
}
