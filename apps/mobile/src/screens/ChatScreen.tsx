import React, { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  ActivityIndicator,
  FlatList,
  KeyboardAvoidingView,
  Platform,
  Pressable,
  SafeAreaView,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  View
} from "react-native";
import {
  listChatMessages,
  listChats,
  listUsers,
  sendChatMessage,
  type ChatMessage,
  type ChatSummary,
  type PlatformUser
} from "../api/platform";

type ChatScreenProps = {
  token: string;
  currentUserId: string;
};

type ChatTarget = {
  user: PlatformUser;
  lastMessage?: ChatSummary["lastMessage"];
};

const CHAT_LIST_POLL_MS = 8000;
const CHAT_MESSAGES_POLL_MS = 3000;

function formatChatError(error: unknown, fallback: string): string {
  if (!(error instanceof Error) || !error.message) {
    return fallback;
  }
  const message = error.message.toLowerCase();
  if (message.includes("network") || message.includes("failed to fetch")) {
    return "Network error. Check connection and retry.";
  }
  if (message.includes("timeout")) {
    return "Request timed out. Please retry.";
  }
  return error.message;
}

function chatSummaryToTarget(chat: ChatSummary): ChatTarget {
  return { user: chat.user, lastMessage: chat.lastMessage };
}

export default function ChatScreen({ token, currentUserId }: ChatScreenProps) {
  const [chatList, setChatList] = useState<ChatSummary[]>([]);
  const [directory, setDirectory] = useState<PlatformUser[]>([]);
  const [selectedChat, setSelectedChat] = useState<ChatTarget | null>(null);
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [draft, setDraft] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [backgroundError, setBackgroundError] = useState<string | null>(null);
  const [sending, setSending] = useState(false);
  const lobbyInFlight = useRef(false);
  const messagesInFlight = useRef(false);

  const loadLobby = useCallback(async (silent = false) => {
    if (lobbyInFlight.current) {
      return;
    }

    lobbyInFlight.current = true;
    if (!silent) {
      setLoading(true);
      setError(null);
    }

    try {
      const [nextChats, nextUsers] = await Promise.all([listChats(token), listUsers(token)]);
      const filteredUsers = nextUsers.filter((user) => user.id !== currentUserId);

      setChatList(nextChats);
      setDirectory(filteredUsers);
      setBackgroundError(null);
      setSelectedChat((current) => {
        if (!current) {
          return null;
        }

        const existingChat = nextChats.find((chat) => chat.user.id === current.user.id);
        if (existingChat) {
          return chatSummaryToTarget(existingChat);
        }

        const existingUser = filteredUsers.find((user) => user.id === current.user.id);
        if (existingUser) {
          return { user: existingUser, lastMessage: current.lastMessage };
        }

        return current;
      });
    } catch (loadError) {
      const formattedError = formatChatError(loadError, "could not load conversations");
      if (silent) {
        setBackgroundError(formattedError);
      } else {
        setError(formattedError);
      }
    } finally {
      if (!silent) {
        setLoading(false);
      }
      lobbyInFlight.current = false;
    }
  }, [token, currentUserId]);

  const loadMessages = useCallback(
    async (chat: ChatTarget, silent = false): Promise<boolean> => {
      if (messagesInFlight.current) {
        return true;
      }

      messagesInFlight.current = true;
      if (!silent) {
        setLoading(true);
        setError(null);
      }

      try {
        const nextMessages = await listChatMessages(token, chat.user.id);
        setMessages(nextMessages);
        setBackgroundError(null);
        return true;
      } catch (loadError) {
        const formattedError = formatChatError(loadError, "could not load messages");
        if (silent) {
          setBackgroundError(formattedError);
        } else {
          setError(formattedError);
        }
        return false;
      } finally {
        if (!silent) {
          setLoading(false);
        }
        messagesInFlight.current = false;
      }
    },
    [token]
  );

  useEffect(() => {
    loadLobby();
  }, [loadLobby]);

  async function openChat(target: ChatTarget) {
    setSelectedChat(target);
    setError(null);
    setBackgroundError(null);
    const ok = await loadMessages(target);
    if (!ok) {
      setMessages([]);
    }
  }

  async function onSend() {
    if (!selectedChat) {
      return;
    }

    const content = draft.trim();
    if (!content) {
      return;
    }

    setSending(true);
    setError(null);
    setBackgroundError(null);
    try {
      const sent = await sendChatMessage(token, selectedChat.user.id, content);
      setMessages((prev) => [...prev, sent]);
      setDraft("");
      await loadLobby(true);
      await loadMessages(selectedChat, true);
    } catch (sendError) {
      setError(formatChatError(sendError, "could not send message"));
    } finally {
      setSending(false);
    }
  }

  async function reloadCurrentView() {
    setBackgroundError(null);
    if (selectedChat) {
      await loadMessages(selectedChat);
      return;
    }
    await loadLobby();
  }

  useEffect(() => {
    if (selectedChat) {
      return;
    }

    const intervalId = setInterval(() => {
      void loadLobby(true);
    }, CHAT_LIST_POLL_MS);

    return () => {
      clearInterval(intervalId);
    };
  }, [selectedChat, loadLobby]);

  useEffect(() => {
    if (!selectedChat) {
      return;
    }

    const intervalId = setInterval(() => {
      void loadMessages(selectedChat, true);
      void loadLobby(true);
    }, CHAT_MESSAGES_POLL_MS);

    return () => {
      clearInterval(intervalId);
    };
  }, [selectedChat, loadMessages, loadLobby]);

  const availableContacts = useMemo(
    () => directory.filter((user) => !chatList.some((chat) => chat.user.id === user.id)),
    [directory, chatList]
  );

  const title = useMemo(() => (selectedChat ? selectedChat.user.email : "Chat"), [selectedChat]);

  return (
    <SafeAreaView style={styles.container}>
      <View style={styles.header} testID="chat-screen">
        {selectedChat ? (
          <Pressable
            onPress={() => {
              setSelectedChat(null);
              setMessages([]);
              setDraft("");
              setError(null);
              setBackgroundError(null);
            }}
            testID="chat-back-button"
          >
            <Text style={styles.back}>Back</Text>
          </Pressable>
        ) : (
          <View style={styles.backPlaceholder} />
        )}

        <Text style={styles.title}>{title}</Text>

        <Pressable disabled={loading || sending} onPress={reloadCurrentView} testID="chat-reload-button">
          <Text style={[styles.reload, loading || sending ? styles.disabledAction : null]}>Reload</Text>
        </Pressable>
      </View>

      {error ? (
        <View style={styles.errorWrap}>
          <Text style={styles.error}>{error}</Text>
          <Pressable disabled={loading || sending} onPress={reloadCurrentView}>
            <Text style={[styles.retry, loading || sending ? styles.disabledAction : null]}>Retry</Text>
          </Pressable>
        </View>
      ) : null}

      {!error && backgroundError ? (
        <View style={styles.warnWrap}>
          <Text style={styles.warn}>{backgroundError}</Text>
          <Pressable disabled={loading || sending} onPress={reloadCurrentView}>
            <Text style={[styles.retry, loading || sending ? styles.disabledAction : null]}>Reload</Text>
          </Pressable>
        </View>
      ) : null}

      {loading ? (
        <View style={styles.loaderWrap}>
          <ActivityIndicator size="large" color="#1d4ed8" />
          <Text style={styles.loadingText}>
            {selectedChat ? "Loading conversation..." : "Loading chat workspace..."}
          </Text>
        </View>
      ) : selectedChat ? (
        <KeyboardAvoidingView behavior={Platform.OS === "ios" ? "padding" : undefined} style={styles.chatWrap}>
          <FlatList
            contentContainerStyle={styles.messageList}
            data={messages}
            keyExtractor={(item) => item.id}
            ListEmptyComponent={<Text style={styles.empty}>No messages yet. Start the conversation.</Text>}
            renderItem={({ item }) => {
              const mine = item.senderUserId === currentUserId;
              return (
                <View style={[styles.messageBubble, mine ? styles.mine : styles.theirs]}>
                  <Text style={mine ? styles.mineText : styles.theirsText}>{item.content}</Text>
                </View>
              );
            }}
          />

          <View style={styles.composer}>
            <TextInput
              onChangeText={setDraft}
              placeholder="Write a message..."
              style={styles.input}
              testID="chat-message-input"
              value={draft}
            />
            <Pressable
              disabled={sending || loading}
              onPress={onSend}
              style={[styles.sendButton, sending || loading ? styles.sendButtonDisabled : null]}
              testID="chat-send-button"
            >
              <Text style={styles.sendText}>{sending ? "..." : "Send"}</Text>
            </Pressable>
          </View>
        </KeyboardAvoidingView>
      ) : (
        <ScrollView contentContainerStyle={styles.scrollContent}>
          <View style={styles.section}>
            <Text style={styles.sectionEyebrow}>Inbox</Text>
            <Text style={styles.sectionTitle}>Recent conversations</Text>
            {chatList.length === 0 ? <Text style={styles.empty}>No conversations yet.</Text> : null}
            {chatList.map((chat) => (
              <Pressable
                key={chat.user.id}
                disabled={loading || sending}
                onPress={() => void openChat(chatSummaryToTarget(chat))}
                style={[styles.chatCard, loading || sending ? styles.chatCardDisabled : null]}
                testID={`chat-open-${chat.user.id}`}
              >
                <Text style={styles.chatEmail}>{chat.user.email}</Text>
                <Text style={styles.chatPreview}>{chat.lastMessage?.content ?? "Open conversation"}</Text>
              </Pressable>
            ))}
          </View>

          <View style={styles.section}>
            <Text style={styles.sectionEyebrow}>Directory</Text>
            <Text style={styles.sectionTitle}>Start a new conversation</Text>
            {availableContacts.length === 0 ? <Text style={styles.empty}>No additional members available.</Text> : null}
            {availableContacts.map((user) => (
              <Pressable
                key={user.id}
                disabled={loading || sending}
                onPress={() => void openChat({ user })}
                style={[styles.contactCard, loading || sending ? styles.chatCardDisabled : null]}
              >
                <Text style={styles.chatEmail}>{user.email}</Text>
                <Text style={styles.chatPreview}>Message this member</Text>
              </Pressable>
            ))}
          </View>
        </ScrollView>
      )}
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#eef6ff",
    paddingHorizontal: 16,
    paddingTop: 8
  },
  header: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
    marginBottom: 12
  },
  back: {
    color: "#2563eb",
    fontWeight: "700",
    width: 52
  },
  backPlaceholder: {
    width: 52
  },
  title: {
    fontSize: 26,
    fontWeight: "800",
    flex: 1,
    textAlign: "center",
    color: "#0f172a"
  },
  reload: {
    color: "#2563eb",
    fontWeight: "700",
    width: 52,
    textAlign: "right"
  },
  disabledAction: {
    color: "#93c5fd"
  },
  scrollContent: {
    paddingBottom: 28
  },
  section: {
    marginBottom: 24
  },
  sectionEyebrow: {
    fontSize: 12,
    fontWeight: "700",
    letterSpacing: 1,
    textTransform: "uppercase",
    color: "#64748b",
    marginBottom: 4
  },
  sectionTitle: {
    fontSize: 18,
    fontWeight: "700",
    color: "#0f172a",
    marginBottom: 12
  },
  chatCard: {
    backgroundColor: "#ffffff",
    borderRadius: 14,
    padding: 14,
    marginBottom: 10,
    borderWidth: 1,
    borderColor: "#dbeafe"
  },
  contactCard: {
    backgroundColor: "#eff6ff",
    borderRadius: 14,
    padding: 14,
    marginBottom: 10,
    borderWidth: 1,
    borderColor: "#bfdbfe"
  },
  chatCardDisabled: {
    opacity: 0.7
  },
  chatEmail: {
    fontWeight: "700",
    fontSize: 16,
    color: "#0f172a"
  },
  chatPreview: {
    marginTop: 4,
    color: "#475569"
  },
  empty: {
    color: "#64748b",
    marginTop: 6
  },
  errorWrap: {
    marginBottom: 8,
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
    gap: 12
  },
  warnWrap: {
    marginBottom: 8,
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
    gap: 12
  },
  error: {
    flex: 1,
    color: "#b91c1c"
  },
  warn: {
    flex: 1,
    color: "#b45309"
  },
  retry: {
    color: "#2563eb",
    fontWeight: "700"
  },
  loaderWrap: {
    flex: 1,
    alignItems: "center",
    justifyContent: "center",
    gap: 10
  },
  loadingText: {
    color: "#64748b"
  },
  chatWrap: {
    flex: 1
  },
  messageList: {
    paddingBottom: 12
  },
  messageBubble: {
    maxWidth: "78%",
    borderRadius: 14,
    paddingHorizontal: 12,
    paddingVertical: 10,
    marginBottom: 8
  },
  mine: {
    alignSelf: "flex-end",
    backgroundColor: "#0f172a"
  },
  theirs: {
    alignSelf: "flex-start",
    backgroundColor: "#dbeafe"
  },
  mineText: {
    color: "#ffffff"
  },
  theirsText: {
    color: "#0f172a"
  },
  composer: {
    flexDirection: "row",
    gap: 8,
    paddingVertical: 8
  },
  input: {
    flex: 1,
    borderWidth: 1,
    borderColor: "#bfdbfe",
    borderRadius: 12,
    paddingHorizontal: 12,
    backgroundColor: "#ffffff"
  },
  sendButton: {
    backgroundColor: "#1d4ed8",
    borderRadius: 12,
    paddingHorizontal: 14,
    alignItems: "center",
    justifyContent: "center"
  },
  sendButtonDisabled: {
    opacity: 0.75
  },
  sendText: {
    color: "#ffffff",
    fontWeight: "700"
  }
});
