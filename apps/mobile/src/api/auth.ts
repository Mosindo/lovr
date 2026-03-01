export type AuthUser = {
  id: string;
  email: string;
  createdAt: string;
};

export type AuthResponse = {
  token: string;
  user: AuthUser;
};

export type DiscoverUser = {
  id: string;
  email: string;
  createdAt: string;
};

type DiscoverResponse = {
  users: DiscoverUser[];
};

type MatchesResponse = {
  matches: DiscoverUser[];
};

type LikeResponse = {
  matched: boolean;
};

export type ChatMessage = {
  id: string;
  senderUserId: string;
  recipientUserId: string;
  content: string;
  createdAt: string;
};

export type ChatSummary = {
  user: DiscoverUser;
  lastMessage?: {
    content: string;
    createdAt: string;
  };
};

type ChatsResponse = {
  chats: ChatSummary[];
};

type ChatMessagesResponse = {
  messages: ChatMessage[];
};

type ApiErrorPayload = {
  error?: string;
};

const API_BASE_URL = process.env.EXPO_PUBLIC_API_URL ?? "http://localhost:18080";
const API_TIMEOUT_MS = 10000;

async function request<T>(path: string, init?: RequestInit): Promise<T> {
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

export async function discover(token: string): Promise<DiscoverUser[]> {
  const payload = await request<DiscoverResponse>("/discover", {
    method: "GET",
    headers: authHeaders(token)
  });
  return payload.users;
}

export async function like(token: string, toUserId: string): Promise<LikeResponse> {
  return request<LikeResponse>("/likes", {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify({ toUserId })
  });
}

export async function block(token: string, toUserId: string): Promise<void> {
  await request<{ blocked: boolean }>("/block", {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify({ toUserId })
  });
}

export async function matches(token: string): Promise<DiscoverUser[]> {
  const payload = await request<MatchesResponse>("/matches", {
    method: "GET",
    headers: authHeaders(token)
  });
  return payload.matches;
}

export async function chats(token: string): Promise<ChatSummary[]> {
  const payload = await request<ChatsResponse>("/chats", {
    method: "GET",
    headers: authHeaders(token)
  });
  return payload.chats;
}

export async function chatMessages(token: string, userId: string): Promise<ChatMessage[]> {
  const payload = await request<ChatMessagesResponse>(`/chats/${userId}/messages`, {
    method: "GET",
    headers: authHeaders(token)
  });
  return payload.messages;
}

export async function sendChatMessage(token: string, userId: string, content: string): Promise<ChatMessage> {
  return request<ChatMessage>(`/chats/${userId}/messages`, {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify({ content })
  });
}
