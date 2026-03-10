import React, { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  ActivityIndicator,
  FlatList,
  KeyboardAvoidingView,
  Platform,
  Pressable,
  SafeAreaView,
  StyleSheet,
  Text,
  TextInput,
  View
} from "react-native";
import {
  chatMessages,
  chats,
  sendChatMessage,
  type ChatMessage,
  type ChatSummary
} from "../api/auth";

type ChatsScreenProps = {
  token: string;
  currentUserId: string;
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

export default function ChatsScreen({ token, currentUserId }: ChatsScreenProps) {
  const [chatList, setChatList] = useState<ChatSummary[]>([]);
  const [selectedChat, setSelectedChat] = useState<ChatSummary | null>(null);
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [draft, setDraft] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [backgroundError, setBackgroundError] = useState<string | null>(null);
  const [sending, setSending] = useState(false);
  const chatsInFlight = useRef(false);
  const messagesInFlight = useRef(false);

  const loadChats = useCallback(async (silent = false) => {
    if (chatsInFlight.current) {
      return;
    }
    chatsInFlight.current = true;
    if (!silent) {
      setLoading(true);
      setError(null);
    }
    try {
      const nextChats = await chats(token);
      setChatList(nextChats);
      setBackgroundError(null);
      setSelectedChat((current) => {
        if (!current) {
          return null;
        }
        return nextChats.find((chat) => chat.user.id === current.user.id) ?? null;
      });
    } catch (loadError) {
      const formattedError = formatChatError(loadError, "could not load chats");
      if (silent) {
        setBackgroundError(formattedError);
      } else {
        setError(formattedError);
      }
    } finally {
      if (!silent) {
        setLoading(false);
      }
      chatsInFlight.current = false;
    }
  }, [token]);

  const loadMessages = useCallback(
    async (chat: ChatSummary, silent = false): Promise<boolean> => {
      if (messagesInFlight.current) {
        return true;
      }
      messagesInFlight.current = true;
      if (!silent) {
        setLoading(true);
        setError(null);
      }
      try {
        const nextMessages = await chatMessages(token, chat.user.id);
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
    loadChats();
  }, [loadChats]);

  async function openChat(chat: ChatSummary) {
    setSelectedChat(chat);
    setError(null);
    setBackgroundError(null);
    const ok = await loadMessages(chat);
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
      await loadChats(true);
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
    await loadChats();
  }

  useEffect(() => {
    if (selectedChat) {
      return;
    }

    const intervalId = setInterval(() => {
      void loadChats(true);
    }, CHAT_LIST_POLL_MS);

    return () => {
      clearInterval(intervalId);
    };
  }, [selectedChat, loadChats]);

  useEffect(() => {
    if (!selectedChat) {
      return;
    }

    const intervalId = setInterval(() => {
      void loadMessages(selectedChat, true);
      void loadChats(true);
    }, CHAT_MESSAGES_POLL_MS);

    return () => {
      clearInterval(intervalId);
    };
  }, [selectedChat, loadMessages, loadChats]);

  const title = useMemo(() => (selectedChat ? selectedChat.user.email : "Chats"), [selectedChat]);

  return (
    <SafeAreaView style={styles.container}>
      <View style={styles.header} testID="chats-screen">
        {selectedChat ? (
          <Pressable
            onPress={() => {
              setSelectedChat(null);
              setMessages([]);
              setError(null);
              setBackgroundError(null);
            }}
            testID="chats-back-button"
          >
            <Text style={styles.back}>Back</Text>
          </Pressable>
        ) : (
          <View style={styles.backPlaceholder} />
        )}

        <Text style={styles.title}>{title}</Text>

        <Pressable
          disabled={loading || sending}
          onPress={reloadCurrentView}
          testID="chats-reload-button"
        >
          <Text style={[styles.reload, loading || sending ? styles.reloadDisabled : null]}>Reload</Text>
        </Pressable>
      </View>

      {error ? (
        <View style={styles.errorWrap}>
          <Text style={styles.error}>{error}</Text>
          <Pressable
            disabled={loading || sending}
            onPress={reloadCurrentView}
          >
            <Text style={[styles.retry, loading || sending ? styles.retryDisabled : null]}>Retry</Text>
          </Pressable>
        </View>
      ) : null}
      {!error && backgroundError ? (
        <View style={styles.warnWrap}>
          <Text style={styles.warn}>{backgroundError}</Text>
          <Pressable disabled={loading || sending} onPress={reloadCurrentView}>
            <Text style={[styles.retry, loading || sending ? styles.retryDisabled : null]}>Reload</Text>
          </Pressable>
        </View>
      ) : null}

      {loading ? (
        <View style={styles.loaderWrap}>
          <ActivityIndicator size="large" color="#111827" />
          <Text style={styles.loadingText}>{selectedChat ? "Loading conversation..." : "Loading chats..."}</Text>
        </View>
      ) : selectedChat ? (
        <KeyboardAvoidingView behavior={Platform.OS === "ios" ? "padding" : undefined} style={styles.chatWrap}>
          <FlatList
            contentContainerStyle={styles.messageList}
            data={messages}
            keyExtractor={(item) => item.id}
            ListEmptyComponent={<Text style={styles.empty}>No messages yet.</Text>}
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
              testID="chats-message-input"
              value={draft}
            />
            <Pressable
              disabled={sending || loading}
              onPress={onSend}
              style={[styles.sendButton, sending || loading ? styles.sendButtonDisabled : null]}
              testID="chats-send-button"
            >
              <Text style={styles.sendText}>{sending ? "..." : "Send"}</Text>
            </Pressable>
          </View>
        </KeyboardAvoidingView>
      ) : (
        <FlatList
          contentContainerStyle={styles.list}
          data={chatList}
          keyExtractor={(item) => item.user.id}
          ListEmptyComponent={<Text style={styles.empty}>No chats yet. Create a match first.</Text>}
          renderItem={({ item }) => (
            <Pressable
              disabled={loading || sending}
              onPress={() => openChat(item)}
              style={[styles.chatCard, loading || sending ? styles.chatCardDisabled : null]}
              testID={`chats-open-${item.user.id}`}
            >
              <Text style={styles.chatEmail}>{item.user.email}</Text>
              <Text style={styles.chatPreview}>{item.lastMessage?.content ?? "No messages yet"}</Text>
            </Pressable>
          )}
        />
      )}
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#f8fafc",
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
    width: 42
  },
  backPlaceholder: {
    width: 42
  },
  title: {
    fontSize: 24,
    fontWeight: "700",
    flex: 1,
    textAlign: "center"
  },
  reload: {
    color: "#2563eb",
    fontWeight: "700",
    width: 42,
    textAlign: "right"
  },
  reloadDisabled: {
    color: "#93c5fd"
  },
  list: {
    paddingBottom: 24
  },
  chatCard: {
    backgroundColor: "#ffffff",
    borderRadius: 12,
    padding: 14,
    marginBottom: 10,
    elevation: 2
  },
  chatCardDisabled: {
    opacity: 0.7
  },
  chatEmail: {
    fontWeight: "700",
    fontSize: 16
  },
  chatPreview: {
    marginTop: 4,
    color: "#6b7280"
  },
  empty: {
    textAlign: "center",
    color: "#6b7280",
    marginTop: 26
  },
  error: {
    color: "#b91c1c"
  },
  errorWrap: {
    marginBottom: 8,
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between"
  },
  warnWrap: {
    marginBottom: 8,
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between"
  },
  warn: {
    color: "#92400e"
  },
  retry: {
    color: "#2563eb",
    fontWeight: "700"
  },
  retryDisabled: {
    color: "#93c5fd"
  },
  loaderWrap: {
    flex: 1,
    alignItems: "center",
    justifyContent: "center",
    gap: 10
  },
  loadingText: {
    color: "#6b7280"
  },
  chatWrap: {
    flex: 1
  },
  messageList: {
    paddingBottom: 12
  },
  messageBubble: {
    maxWidth: "78%",
    borderRadius: 12,
    paddingHorizontal: 12,
    paddingVertical: 9,
    marginBottom: 8
  },
  mine: {
    alignSelf: "flex-end",
    backgroundColor: "#111827"
  },
  theirs: {
    alignSelf: "flex-start",
    backgroundColor: "#e5e7eb"
  },
  mineText: {
    color: "#ffffff"
  },
  theirsText: {
    color: "#111827"
  },
  composer: {
    flexDirection: "row",
    gap: 8,
    paddingVertical: 8
  },
  input: {
    flex: 1,
    borderWidth: 1,
    borderColor: "#d1d5db",
    borderRadius: 10,
    paddingHorizontal: 12,
    backgroundColor: "#ffffff"
  },
  sendButton: {
    backgroundColor: "#111827",
    borderRadius: 10,
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
