import React, { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  FlatList,
  KeyboardAvoidingView,
  Platform,
  Pressable,
  ScrollView,
  StyleSheet,
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
import { Header, ScreenContainer } from "../shared/layout";
import { Avatar, Button, Card, Input, Loader, Text, colors, radii, spacing } from "../shared/ui";

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
  const subtitle = selectedChat ? "Direct conversation" : "Conversations and team directory";

  return (
    <ScreenContainer testID="chat-screen">
      <Header
        action={
          <Pressable disabled={loading || sending} onPress={reloadCurrentView} testID="chat-reload-button">
            <Text
              style={[styles.reload, loading || sending ? styles.disabledAction : null]}
              tone={loading || sending ? "muted" : "primary"}
              variant="label"
              weight="bold"
            >
              Reload
            </Text>
          </Pressable>
        }
        eyebrow="Inbox"
        leading={
          selectedChat ? (
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
              <Text style={styles.back} tone="primary" variant="label" weight="bold">
                Back
              </Text>
            </Pressable>
          ) : null
        }
        style={styles.headerShell}
        subtitle={subtitle}
        title={title}
      />

      {error ? (
        <Card padding="sm" style={styles.errorWrap}>
          <Text style={styles.error} tone="danger" variant="label" weight="medium">
            {error}
          </Text>
          <Pressable disabled={loading || sending} onPress={reloadCurrentView}>
            <Text
              style={[styles.retry, loading || sending ? styles.disabledAction : null]}
              tone={loading || sending ? "muted" : "primary"}
              variant="label"
              weight="bold"
            >
              Retry
            </Text>
          </Pressable>
        </Card>
      ) : null}

      {!error && backgroundError ? (
        <Card padding="sm" style={styles.warnWrap} variant="muted">
          <Text style={styles.warn} tone="muted" variant="label" weight="medium">
            {backgroundError}
          </Text>
          <Pressable disabled={loading || sending} onPress={reloadCurrentView}>
            <Text
              style={[styles.retry, loading || sending ? styles.disabledAction : null]}
              tone={loading || sending ? "muted" : "primary"}
              variant="label"
              weight="bold"
            >
              Reload
            </Text>
          </Pressable>
        </Card>
      ) : null}

      {loading ? (
        <Loader
          fullScreen
          label={selectedChat ? "Loading conversation..." : "Loading chat workspace..."}
          style={styles.loaderWrap}
        />
      ) : selectedChat ? (
        <KeyboardAvoidingView behavior={Platform.OS === "ios" ? "padding" : undefined} style={styles.chatWrap}>
          <FlatList
            contentContainerStyle={styles.messageList}
            data={messages}
            keyExtractor={(item) => item.id}
            ListEmptyComponent={
              <Text style={styles.empty} tone="muted">
                No messages yet. Start the conversation.
              </Text>
            }
            renderItem={({ item }) => {
              const mine = item.senderUserId === currentUserId;
              return (
                <View style={[styles.messageBubble, mine ? styles.mine : styles.theirs]}>
                  <Text style={mine ? styles.mineText : styles.theirsText} tone={mine ? "inverse" : "default"}>
                    {item.content}
                  </Text>
                </View>
              );
            }}
          />

          <View style={styles.composer}>
            <Input
              containerStyle={styles.inputWrap}
              onChangeText={setDraft}
              placeholder="Write a message..."
              style={styles.input}
              testID="chat-message-input"
              value={draft}
            />
            <Button
              disabled={loading}
              label="Send"
              loading={sending}
              onPress={onSend}
              size="sm"
              testID="chat-send-button"
            />
          </View>
        </KeyboardAvoidingView>
      ) : (
        <ScrollView contentContainerStyle={styles.scrollContent}>
          <View style={styles.section}>
            <Text style={styles.sectionEyebrow} tone="secondary" variant="eyebrow" weight="bold">
              Inbox
            </Text>
            <Text style={styles.sectionTitle} variant="heading" weight="bold">
              Recent conversations
            </Text>
            {chatList.length === 0 ? (
              <Text style={styles.empty} tone="muted">
                No conversations yet.
              </Text>
            ) : null}
            {chatList.map((chat) => (
              <Pressable
                key={chat.user.id}
                disabled={loading || sending}
                onPress={() => void openChat(chatSummaryToTarget(chat))}
                style={[styles.chatCard, loading || sending ? styles.chatCardDisabled : null]}
                testID={`chat-open-${chat.user.id}`}
              >
                <Avatar name={chat.user.email} size={40} />
                <View style={styles.chatTextWrap}>
                  <Text style={styles.chatEmail} variant="label" weight="bold">
                    {chat.user.email}
                  </Text>
                  <Text style={styles.chatPreview} numberOfLines={1} tone="muted">
                    {chat.lastMessage?.content ?? "Open conversation"}
                  </Text>
                </View>
              </Pressable>
            ))}
          </View>

          <View style={styles.section}>
            <Text style={styles.sectionEyebrow} tone="secondary" variant="eyebrow" weight="bold">
              Directory
            </Text>
            <Text style={styles.sectionTitle} variant="heading" weight="bold">
              Start a new conversation
            </Text>
            {availableContacts.length === 0 ? (
              <Text style={styles.empty} tone="muted">
                No additional members available.
              </Text>
            ) : null}
            {availableContacts.map((user) => (
              <Pressable
                key={user.id}
                disabled={loading || sending}
                onPress={() => void openChat({ user })}
                style={[styles.contactCard, loading || sending ? styles.chatCardDisabled : null]}
              >
                <Avatar name={user.email} size={40} />
                <View style={styles.chatTextWrap}>
                  <Text style={styles.chatEmail} variant="label" weight="bold">
                    {user.email}
                  </Text>
                  <Text style={styles.chatPreview} tone="muted">
                    Message this member
                  </Text>
                </View>
              </Pressable>
            ))}
          </View>
        </ScrollView>
      )}
    </ScreenContainer>
  );
}

const styles = StyleSheet.create({
  headerShell: {
    marginBottom: spacing.md,
  },
  back: {
    color: colors.primary
  },
  reload: {
    color: colors.primary
  },
  disabledAction: {
    color: colors.textMuted
  },
  scrollContent: {
    paddingBottom: spacing.xxxl
  },
  section: {
    marginBottom: spacing.xxl
  },
  sectionEyebrow: {
    marginBottom: spacing.xs
  },
  sectionTitle: {
    marginBottom: spacing.md
  },
  chatCard: {
    flexDirection: "row",
    alignItems: "center",
    gap: spacing.md,
    backgroundColor: colors.surface,
    borderRadius: radii.md,
    padding: spacing.md,
    marginBottom: spacing.sm,
    borderWidth: 1,
    borderColor: colors.border
  },
  contactCard: {
    flexDirection: "row",
    alignItems: "center",
    gap: spacing.md,
    backgroundColor: colors.surfaceMuted,
    borderRadius: radii.md,
    padding: spacing.md,
    marginBottom: spacing.sm,
    borderWidth: 1,
    borderColor: colors.border
  },
  chatCardDisabled: {
    opacity: 0.7
  },
  chatTextWrap: {
    flex: 1,
    gap: spacing.xs
  },
  chatEmail: {
    color: colors.text
  },
  chatPreview: {
    color: colors.textMuted
  },
  empty: {
    color: colors.textMuted,
    marginTop: spacing.xs
  },
  errorWrap: {
    marginBottom: spacing.sm,
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
    gap: spacing.md
  },
  warnWrap: {
    marginBottom: spacing.sm,
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
    gap: spacing.md
  },
  error: {
    flex: 1
  },
  warn: {
    flex: 1
  },
  retry: {
    color: colors.primary
  },
  loaderWrap: {
    flex: 1
  },
  chatWrap: {
    flex: 1
  },
  messageList: {
    paddingBottom: spacing.md
  },
  messageBubble: {
    maxWidth: "78%",
    borderRadius: radii.md,
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm,
    marginBottom: spacing.sm
  },
  mine: {
    alignSelf: "flex-end",
    backgroundColor: colors.primary
  },
  theirs: {
    alignSelf: "flex-start",
    backgroundColor: colors.surfaceMuted
  },
  mineText: {
    color: colors.inverse
  },
  theirsText: {
    color: colors.text
  },
  composer: {
    flexDirection: "row",
    alignItems: "flex-end",
    gap: spacing.sm,
    paddingVertical: spacing.sm
  },
  inputWrap: {
    flex: 1
  },
  input: {
    minHeight: 48
  }
});
