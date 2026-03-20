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
import { type AuthUser } from "../api/auth";
import { useAuth } from "../hooks/useAuth";
import { listUsers, type PlatformUser } from "../api/platform";

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
    <SafeAreaView style={styles.container} testID="profile-screen">
      <View style={styles.header}>
        <View>
          <Text style={styles.eyebrow}>Workspace profile</Text>
          <Text style={styles.title}>Profile</Text>
        </View>
        <Pressable onPress={load} testID="profile-reload-button">
          <Text style={styles.reload}>Reload</Text>
        </Pressable>
      </View>

      <View style={styles.card}>
        <Text style={styles.sectionTitle}>Session</Text>
        <Text style={styles.label}>Email</Text>
        <Text style={styles.value}>{user.email}</Text>

        <Text style={styles.label}>User ID</Text>
        <Text style={styles.value}>{user.id}</Text>

        <Text style={styles.label}>Organization</Text>
        <Text style={styles.value}>{user.organizationId}</Text>

        <Text style={styles.label}>Member since</Text>
        <Text style={styles.value}>{formatMemberDate(user.createdAt)}</Text>

        <Pressable disabled={isLoggingOut} onPress={logout} style={styles.button} testID="profile-logout-button">
          <Text style={styles.buttonText}>{isLoggingOut ? "Logging out..." : "Logout"}</Text>
        </Pressable>
      </View>

      <View style={styles.directoryHeader}>
        <View>
          <Text style={styles.sectionTitle}>Directory</Text>
          <Text style={styles.directoryMeta}>{directoryCountLabel}</Text>
        </View>
      </View>

      {error ? <Text style={styles.error}>{error}</Text> : null}

      {loading ? (
        <View style={styles.loaderWrap}>
          <ActivityIndicator color="#0f172a" size="large" />
        </View>
      ) : (
        <FlatList
          contentContainerStyle={styles.list}
          data={users}
          keyExtractor={(item) => item.id}
          ListEmptyComponent={<Text style={styles.empty}>Invite teammates or create another test account.</Text>}
          renderItem={({ item }) => (
            <View style={styles.memberCard}>
              <Text style={styles.memberEmail}>{item.email}</Text>
              <Text style={styles.memberMeta}>Joined {formatMemberDate(item.createdAt)}</Text>
            </View>
          )}
        />
      )}
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#fdf5ec",
    paddingHorizontal: 16,
    paddingTop: 12
  },
  header: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
    marginBottom: 12
  },
  eyebrow: {
    fontSize: 12,
    fontWeight: "700",
    textTransform: "uppercase",
    letterSpacing: 1,
    color: "#9a3412"
  },
  title: {
    fontSize: 30,
    fontWeight: "800",
    color: "#111827"
  },
  reload: {
    color: "#c2410c",
    fontWeight: "700"
  },
  card: {
    backgroundColor: "#ffffff",
    borderRadius: 18,
    padding: 18,
    borderWidth: 1,
    borderColor: "#fed7aa",
    marginBottom: 16
  },
  sectionTitle: {
    fontSize: 18,
    fontWeight: "700",
    color: "#111827",
    marginBottom: 10
  },
  label: {
    marginTop: 8,
    color: "#9a3412",
    fontSize: 12,
    textTransform: "uppercase",
    fontWeight: "700"
  },
  value: {
    marginTop: 2,
    fontSize: 14,
    color: "#111827"
  },
  button: {
    marginTop: 18,
    borderRadius: 12,
    backgroundColor: "#c2410c",
    paddingVertical: 12,
    alignItems: "center"
  },
  buttonText: {
    color: "#ffffff",
    fontWeight: "700"
  },
  directoryHeader: {
    marginBottom: 8
  },
  directoryMeta: {
    color: "#7c2d12"
  },
  error: {
    color: "#b91c1c",
    marginBottom: 8
  },
  loaderWrap: {
    flex: 1,
    alignItems: "center",
    justifyContent: "center"
  },
  list: {
    paddingBottom: 28
  },
  memberCard: {
    backgroundColor: "#fff7ed",
    borderRadius: 14,
    padding: 14,
    marginBottom: 10,
    borderWidth: 1,
    borderColor: "#fdba74"
  },
  memberEmail: {
    fontSize: 16,
    fontWeight: "700",
    color: "#111827"
  },
  memberMeta: {
    color: "#7c2d12",
    marginTop: 4
  },
  empty: {
    textAlign: "center",
    color: "#7c2d12",
    marginTop: 24
  }
});
