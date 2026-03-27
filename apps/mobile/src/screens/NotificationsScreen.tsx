import React, { useCallback, useEffect, useMemo, useState } from "react";
import { FlatList, Pressable, StyleSheet, View } from "react-native";
import {
  createNotification,
  listNotifications,
  markNotificationRead,
  type Notification
} from "../api/platform";
import { Header, ScreenContainer } from "../shared/layout";
import { Button, Card, Loader, Text, colors, radii, spacing } from "../shared/ui";

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
        title: "go-react-saas ready",
        body: "Your generic mobile workspace is connected to the shared platform modules."
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
      <ScreenContainer testID="notifications-screen">
        <Loader fullScreen label="Loading notifications..." />
      </ScreenContainer>
    );
  }

  return (
    <ScreenContainer contentMaxWidth={820} testID="notifications-screen">
      <Header
        action={
          <View style={styles.actions}>
            <Button
              disabled={creating || loading}
              label="Reload"
              onPress={load}
              size="sm"
              testID="notifications-reload-button"
              variant="outline"
            />
            <Button
              disabled={creating || loading}
              label={creating ? "Creating..." : "Sample"}
              onPress={onCreateSample}
              size="sm"
              testID="notifications-create-sample-button"
              variant="secondary"
            />
          </View>
        }
        eyebrow="Inbox"
        subtitle={`${unreadCount} unread`}
        style={styles.headerShell}
        title="Notifications"
      />

      {error ? (
        <Text style={styles.error} tone="danger" variant="label" weight="medium">
          {error}
        </Text>
      ) : null}

      <FlatList
        contentContainerStyle={styles.list}
        data={notifications}
        keyExtractor={(item) => item.id}
        ListEmptyComponent={
          <Text style={styles.empty} tone="muted">
            No notifications yet. Create a sample to test the flow.
          </Text>
        }
        renderItem={({ item }) => (
          <Pressable onPress={() => void onMarkRead(item)}>
            <Card style={styles.card} variant={item.isRead ? "muted" : "default"}>
              <View style={styles.cardHeader}>
                <Text style={styles.cardTitle} variant="heading" weight="bold">
                  {item.title}
                </Text>
                <Text
                  style={[styles.badge, item.isRead ? styles.badgeRead : styles.badgeUnread]}
                  tone={item.isRead ? "secondary" : "primary"}
                  variant="caption"
                  weight="bold"
                >
                  {item.isRead ? "Read" : "Unread"}
                </Text>
              </View>
              <Text style={styles.cardType} tone="secondary" variant="eyebrow" weight="bold">
                {item.type}
              </Text>
              <Text style={styles.cardBody}>{item.body}</Text>
              <Text style={styles.meta} tone="muted" variant="caption">
                {formatNotificationDate(item.createdAt)}
              </Text>
            </Card>
          </Pressable>
        )}
      />
    </ScreenContainer>
  );
}

const styles = StyleSheet.create({
  headerShell: {
    marginBottom: spacing.md,
  },
  actions: {
    flexDirection: "row",
    gap: spacing.sm,
    paddingTop: spacing.sm
  },
  error: {
    marginBottom: spacing.sm
  },
  list: {
    paddingBottom: spacing.xxxl
  },
  empty: {
    textAlign: "center",
    marginTop: spacing.xxl
  },
  card: {
    marginBottom: spacing.md
  },
  cardHeader: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
    gap: spacing.md,
    marginBottom: spacing.sm
  },
  cardTitle: {
    flex: 1,
    color: colors.text
  },
  badge: {
    borderRadius: radii.pill,
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.xs,
    overflow: "hidden"
  },
  badgeUnread: {
    backgroundColor: "#dbeafe"
  },
  badgeRead: {
    backgroundColor: "#d1fae5"
  },
  cardType: {
    marginBottom: spacing.sm
  },
  cardBody: {
    color: colors.text,
    lineHeight: 22
  },
  meta: {
    marginTop: spacing.md
  }
});
