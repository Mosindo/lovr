import { type AuthUser } from "./auth";
import { request } from "../utils/http";

export type PlatformUser = AuthUser;

export type Post = {
  id: string;
  authorUserId: string;
  title: string;
  body: string;
  createdAt: string;
  updatedAt: string;
};

export type ChatMessage = {
  id: string;
  senderUserId: string;
  recipientUserId: string;
  content: string;
  createdAt: string;
};

export type ChatSummary = {
  user: PlatformUser;
  lastMessage?: {
    content: string;
    createdAt: string;
  };
};

export type Notification = {
  id: string;
  userId: string;
  type: string;
  title: string;
  body: string;
  isRead: boolean;
  createdAt: string;
  readAt?: string;
};

type UsersResponse = {
  users: PlatformUser[];
};

type PostsResponse = {
  posts: Post[];
};

type ChatsResponse = {
  chats: ChatSummary[];
};

type ChatMessagesResponse = {
  messages: ChatMessage[];
};

type NotificationsResponse = {
  notifications: Notification[];
};

export type CreateNotificationInput = {
  type: string;
  title: string;
  body: string;
};

function authHeaders(token: string): Record<string, string> {
  return { Authorization: `Bearer ${token}` };
}

export async function listUsers(token: string): Promise<PlatformUser[]> {
  const payload = await request<UsersResponse>("/users", {
    method: "GET",
    headers: authHeaders(token)
  });
  return payload.users;
}

export async function listPosts(token: string): Promise<Post[]> {
  const payload = await request<PostsResponse>("/posts", {
    method: "GET",
    headers: authHeaders(token)
  });
  return payload.posts;
}

export async function createPost(token: string, title: string, body: string): Promise<Post> {
  return request<Post>("/posts", {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify({ title, body })
  });
}

export async function listChats(token: string): Promise<ChatSummary[]> {
  const payload = await request<ChatsResponse>("/chats", {
    method: "GET",
    headers: authHeaders(token)
  });
  return payload.chats;
}

export async function listChatMessages(token: string, userId: string): Promise<ChatMessage[]> {
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

export async function listNotifications(token: string): Promise<Notification[]> {
  const payload = await request<NotificationsResponse>("/notifications", {
    method: "GET",
    headers: authHeaders(token)
  });
  return payload.notifications;
}

export async function createNotification(token: string, input: CreateNotificationInput): Promise<Notification> {
  return request<Notification>("/notifications", {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify(input)
  });
}

export async function markNotificationRead(token: string, notificationId: string): Promise<Notification> {
  return request<Notification>(`/notifications/${notificationId}/read`, {
    method: "POST",
    headers: authHeaders(token)
  });
}
