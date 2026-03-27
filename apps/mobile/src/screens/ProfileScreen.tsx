import React, { useCallback, useEffect, useMemo, useState } from "react";
import { FlatList, Pressable, StyleSheet, View } from "react-native";
import { type AuthUser } from "../api/auth";
import { useAuth } from "../hooks/useAuth";
import { listUsers, type PlatformUser } from "../api/platform";
import { Avatar, Button, Card, Loader, Text, colors, spacing } from "../shared/ui";
import { Header, ScreenContainer } from "../shared/layout";

type ProfileScreenProps = {
  user: AuthUser;
  token: string;
};

function formatMemberDate(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "Unknown";
  }
  return date.toLocaleDateString();
}

export default function ProfileScreen({ user, token }: ProfileScreenProps) {
  const [users, setUsers] = useState<PlatformUser[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const { logout, isLoggingOut } = useAuth();

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const nextUsers = await listUsers(token);
      setUsers(nextUsers.filter((candidate) => candidate.id !== user.id));
    } catch (loadError) {
      setError(loadError instanceof Error ? loadError.message : "could not load members");
    } finally {
      setLoading(false);
    }
  }, [token, user.id]);

  useEffect(() => {
    load();
  }, [load]);

  const directoryCountLabel = useMemo(() => {
    if (loading) {
      return "Loading members...";
    }
    if (users.length === 0) {
      return "No other members yet.";
    }
    return `${users.length} member${users.length > 1 ? "s" : ""} available`;
  }, [loading, users.length]);

  return (
    <ScreenContainer contentMaxWidth={760} testID="profile-screen">
      <Header
        action={
          <Pressable onPress={load} testID="profile-reload-button">
            <Text style={styles.reload} tone="primary" variant="label" weight="bold">
              Reload
            </Text>
          </Pressable>
        }
        eyebrow="Workspace profile"
        style={styles.header}
        title="Profile"
      />

      <Card style={styles.card}>
        <View style={styles.sessionHeader}>
          <Avatar name={user.email} size={56} />
          <View style={styles.sessionMeta}>
            <Text variant="heading" weight="bold">
              Session
            </Text>
            <Text tone="muted">{user.email}</Text>
          </View>
        </View>

        <Text style={styles.label} tone="muted" variant="eyebrow" weight="bold">
          Email
        </Text>
        <Text style={styles.value}>{user.email}</Text>

        <Text style={styles.label} tone="muted" variant="eyebrow" weight="bold">
          User ID
        </Text>
        <Text style={styles.value}>{user.id}</Text>

        <Text style={styles.label} tone="muted" variant="eyebrow" weight="bold">
          Organization
        </Text>
        <Text style={styles.value}>{user.organizationId}</Text>

        <Text style={styles.label} tone="muted" variant="eyebrow" weight="bold">
          Member since
        </Text>
        <Text style={styles.value}>{formatMemberDate(user.createdAt)}</Text>

        <Button
          fullWidth
          label={isLoggingOut ? "Logging out..." : "Logout"}
          onPress={logout}
          style={styles.button}
          testID="profile-logout-button"
          variant="secondary"
        />
      </Card>

      <View style={styles.directoryHeader}>
        <View>
          <Text style={styles.sectionTitle} variant="heading" weight="bold">
            Directory
          </Text>
          <Text style={styles.directoryMeta} tone="muted">
            {directoryCountLabel}
          </Text>
        </View>
      </View>

      {error ? (
        <Text style={styles.error} tone="danger" variant="label" weight="medium">
          {error}
        </Text>
      ) : null}

      {loading ? (
        <Loader fullScreen label="Loading members..." style={styles.loaderWrap} />
      ) : (
        <FlatList
          contentContainerStyle={styles.list}
          data={users}
          keyExtractor={(item) => item.id}
          ListEmptyComponent={
            <Text style={styles.empty} tone="muted">
              Invite teammates or create another test account.
            </Text>
          }
          renderItem={({ item }) => (
            <Card padding="sm" style={styles.memberCard} variant="muted">
              <View style={styles.memberRow}>
                <Avatar name={item.email} size={40} />
                <View style={styles.memberTextWrap}>
                  <Text style={styles.memberEmail} variant="label" weight="bold">
                    {item.email}
                  </Text>
                  <Text style={styles.memberMeta} tone="muted">
                    Joined {formatMemberDate(item.createdAt)}
                  </Text>
                </View>
              </View>
            </Card>
          )}
        />
      )}
    </ScreenContainer>
  );
}

const styles = StyleSheet.create({
  header: {
    marginBottom: spacing.md
  },
  reload: {
    color: colors.primary
  },
  card: {
    marginBottom: spacing.lg
  },
  sessionHeader: {
    flexDirection: "row",
    alignItems: "center",
    gap: spacing.md,
    marginBottom: spacing.md
  },
  sessionMeta: {
    flex: 1,
    gap: spacing.xs
  },
  sectionTitle: {
    color: colors.text,
    marginBottom: spacing.xs
  },
  label: {
    marginTop: spacing.sm
  },
  value: {
    marginTop: spacing.xs,
    color: colors.text
  },
  button: {
    marginTop: spacing.xl
  },
  directoryHeader: {
    marginBottom: spacing.sm
  },
  directoryMeta: {
    color: colors.textMuted
  },
  error: {
    marginBottom: spacing.sm
  },
  loaderWrap: {
    flex: 1
  },
  list: {
    paddingBottom: spacing.xxxl
  },
  memberCard: {
    marginBottom: spacing.sm
  },
  memberRow: {
    flexDirection: "row",
    alignItems: "center",
    gap: spacing.md
  },
  memberTextWrap: {
    flex: 1,
    gap: spacing.xs
  },
  memberEmail: {
    color: colors.text
  },
  memberMeta: {
    color: colors.textMuted
  },
  empty: {
    textAlign: "center",
    marginTop: spacing.xxl
  }
});
