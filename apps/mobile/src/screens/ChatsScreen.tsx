import React, { useCallback, useEffect, useMemo, useState } from "react";
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

export default function ChatsScreen({ token, currentUserId }: ChatsScreenProps) {
  const [chatList, setChatList] = useState<ChatSummary[]>([]);
  const [selectedChat, setSelectedChat] = useState<ChatSummary | null>(null);
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [draft, setDraft] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [sending, setSending] = useState(false);

  const loadChats = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const nextChats = await chats(token);
      setChatList(nextChats);
      setSelectedChat((current) => {
        if (!current) {
          return null;
        }
        return nextChats.find((chat) => chat.user.id === current.user.id) ?? null;
      });
    } catch (loadError) {
      setError(loadError instanceof Error ? loadError.message : "could not load chats");
    } finally {
      setLoading(false);
    }
  }, [token]);

  const loadMessages = useCallback(
    async (chat: ChatSummary): Promise<boolean> => {
      setLoading(true);
      setError(null);
      try {
        const nextMessages = await chatMessages(token, chat.user.id);
        setMessages(nextMessages);
        return true;
      } catch (loadError) {
        setError(loadError instanceof Error ? loadError.message : "could not load messages");
        return false;
      } finally {
        setLoading(false);
      }
    },
    [token]
  );

  useEffect(() => {
    loadChats();
  }, [loadChats]);

  async function openChat(chat: ChatSummary) {
    setSelectedChat(chat);
    const ok = await loadMessages(chat);
    if (!ok) {
      setSelectedChat(null);
      setMessages([]);
      await loadChats();
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
    try {
      const sent = await sendChatMessage(token, selectedChat.user.id, content);
      setMessages((prev) => [...prev, sent]);
      setDraft("");
      await loadChats();
    } catch (sendError) {
      setError(sendError instanceof Error ? sendError.message : "could not send message");
    } finally {
      setSending(false);
    }
  }

  const title = useMemo(() => (selectedChat ? selectedChat.user.email : "Chats"), [selectedChat]);

  return (
    <SafeAreaView style={styles.container}>
      <View style={styles.header}>
        {selectedChat ? (
          <Pressable
            onPress={() => {
              setSelectedChat(null);
              setMessages([]);
              setError(null);
            }}
          >
            <Text style={styles.back}>Back</Text>
          </Pressable>
        ) : (
          <View style={styles.backPlaceholder} />
        )}

        <Text style={styles.title}>{title}</Text>

        <Pressable
          onPress={async () => {
            if (selectedChat) {
              await loadMessages(selectedChat);
            } else {
              await loadChats();
            }
          }}
        >
          <Text style={styles.reload}>Reload</Text>
        </Pressable>
      </View>

      {error ? <Text style={styles.error}>{error}</Text> : null}

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
              value={draft}
            />
            <Pressable disabled={sending} onPress={onSend} style={styles.sendButton}>
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
            <Pressable onPress={() => openChat(item)} style={styles.chatCard}>
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
    color: "#b91c1c",
    marginBottom: 8
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
  sendText: {
    color: "#ffffff",
    fontWeight: "700"
  }
});
