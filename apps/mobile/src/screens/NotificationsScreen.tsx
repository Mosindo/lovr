import React, { useCallback, useEffect, useMemo, useState } from "react";
import {
  ActivityIndicator,
  FlatList,
  Pressable,
  SafeAreaView,
  StyleSheet,
  Text,
  View
} from "react-native";
import {
  createNotification,
  listNotifications,
  markNotificationRead,
  type Notification
} from "../api/platform";

type NotificationsScreenProps = {
  token: string;
};

function formatNotificationDate(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "Unknown date";
  }
  return date.toLocaleString();
}

export default function NotificationsScreen({ token }: NotificationsScreenProps) {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const nextNotifications = await listNotifications(token);
      setNotifications(nextNotifications);
    } catch (loadError) {
      setError(loadError instanceof Error ? loadError.message : "could not load notifications");
    } finally {
      setLoading(false);
    }
  }, [token]);

  useEffect(() => {
    load();
  }, [load]);

  async function onCreateSample() {
    setCreating(true);
    setError(null);
    try {
      const created = await createNotification(token, {
        type: "system",
        title: "Boilerplate ready",
        body: "Your generic mobile workspace is connected to the platform modules."
      });
      setNotifications((prev) => [created, ...prev.filter((item) => item.id !== created.id)]);
    } catch (createError) {
      setError(createError instanceof Error ? createError.message : "could not create notification");
    } finally {
      setCreating(false);
    }
  }

  async function onMarkRead(item: Notification) {
    if (item.isRead) {
      return;
    }

    try {
      const updated = await markNotificationRead(token, item.id);
      setNotifications((prev) => prev.map((current) => (current.id === updated.id ? updated : current)));
    } catch (markError) {
      setError(markError instanceof Error ? markError.message : "could not update notification");
    }
  }

  const unreadCount = useMemo(
    () => notifications.filter((item) => !item.isRead).length,
    [notifications]
  );

  if (loading && notifications.length === 0) {
    return (
      <SafeAreaView style={styles.loaderContainer}>
        <ActivityIndicator color="#7c3aed" size="large" />
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView style={styles.container} testID="notifications-screen">
      <View style={styles.header}>
        <View>
          <Text style={styles.eyebrow}>Inbox</Text>
          <Text style={styles.title}>Notifications</Text>
          <Text style={styles.subtitle}>{unreadCount} unread</Text>
        </View>

        <View style={styles.actions}>
          <Pressable disabled={creating || loading} onPress={load} testID="notifications-reload-button">
            <Text style={[styles.actionText, creating || loading ? styles.actionDisabled : null]}>Reload</Text>
          </Pressable>
          <Pressable
            disabled={creating || loading}
            onPress={onCreateSample}
            testID="notifications-create-sample-button"
          >
            <Text style={[styles.actionText, creating || loading ? styles.actionDisabled : null]}>
              {creating ? "..." : "Sample"}
            </Text>
          </Pressable>
        </View>
      </View>

      {error ? <Text style={styles.error}>{error}</Text> : null}

      <FlatList
        contentContainerStyle={styles.list}
        data={notifications}
        keyExtractor={(item) => item.id}
        ListEmptyComponent={<Text style={styles.empty}>No notifications yet. Create a sample to test the flow.</Text>}
        renderItem={({ item }) => (
          <Pressable onPress={() => void onMarkRead(item)} style={[styles.card, item.isRead ? styles.readCard : styles.unreadCard]}>
            <View style={styles.cardHeader}>
              <Text style={styles.cardTitle}>{item.title}</Text>
              <Text style={[styles.badge, item.isRead ? styles.badgeRead : styles.badgeUnread]}>
                {item.isRead ? "Read" : "Unread"}
              </Text>
            </View>
            <Text style={styles.cardType}>{item.type}</Text>
            <Text style={styles.cardBody}>{item.body}</Text>
            <Text style={styles.meta}>{formatNotificationDate(item.createdAt)}</Text>
          </Pressable>
        )}
      />
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#f9f5ff",
    paddingHorizontal: 16,
    paddingTop: 12
  },
  loaderContainer: {
    flex: 1,
    alignItems: "center",
    justifyContent: "center",
    backgroundColor: "#f9f5ff"
  },
  header: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "flex-start",
    marginBottom: 12,
    gap: 12
  },
  eyebrow: {
    fontSize: 12,
    fontWeight: "700",
    letterSpacing: 1,
    textTransform: "uppercase",
    color: "#6d28d9"
  },
  title: {
    fontSize: 30,
    fontWeight: "800",
    color: "#1f2937"
  },
  subtitle: {
    marginTop: 2,
    color: "#6b21a8"
  },
  actions: {
    flexDirection: "row",
    gap: 14,
    paddingTop: 8
  },
  actionText: {
    color: "#7c3aed",
    fontWeight: "700"
  },
  actionDisabled: {
    color: "#c4b5fd"
  },
  error: {
    color: "#b91c1c",
    marginBottom: 8
  },
  list: {
    paddingBottom: 28
  },
  empty: {
    textAlign: "center",
    color: "#6b7280",
    marginTop: 24
  },
  card: {
    borderRadius: 18,
    padding: 16,
    marginBottom: 12,
    borderWidth: 1
  },
  unreadCard: {
    backgroundColor: "#ffffff",
    borderColor: "#c4b5fd"
  },
  readCard: {
    backgroundColor: "#f3e8ff",
    borderColor: "#ddd6fe"
  },
  cardHeader: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
    gap: 12,
    marginBottom: 8
  },
  cardTitle: {
    flex: 1,
    fontSize: 17,
    fontWeight: "700",
    color: "#111827"
  },
  badge: {
    borderRadius: 999,
    paddingHorizontal: 10,
    paddingVertical: 4,
    fontSize: 12,
    fontWeight: "700"
  },
  badgeUnread: {
    backgroundColor: "#ede9fe",
    color: "#6d28d9"
  },
  badgeRead: {
    backgroundColor: "#e9d5ff",
    color: "#7e22ce"
  },
  cardType: {
    color: "#7c3aed",
    fontWeight: "700",
    textTransform: "uppercase",
    fontSize: 12,
    marginBottom: 8
  },
  cardBody: {
    color: "#374151",
    lineHeight: 21
  },
  meta: {
    marginTop: 12,
    color: "#6b7280",
    fontSize: 12
  }
});
