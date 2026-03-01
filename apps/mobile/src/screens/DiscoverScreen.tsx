import React, { useCallback, useEffect, useState } from "react";
import {
  ActivityIndicator,
  FlatList,
  Pressable,
  SafeAreaView,
  StyleSheet,
  Text,
  View
} from "react-native";
import { block, discover, like, type DiscoverUser } from "../api/auth";

type DiscoverScreenProps = {
  token: string;
};

export default function DiscoverScreen({ token }: DiscoverScreenProps) {
  const [users, setUsers] = useState<DiscoverUser[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [feedback, setFeedback] = useState<string | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const nextUsers = await discover(token);
      setUsers(nextUsers);
    } catch (loadError) {
      setError(loadError instanceof Error ? loadError.message : "could not load users");
    } finally {
      setLoading(false);
    }
  }, [token]);

  useEffect(() => {
    load();
  }, [load]);

  async function onLike(user: DiscoverUser) {
    try {
      const result = await like(token, user.id);
      setUsers((prev) => prev.filter((candidate) => candidate.id !== user.id));
      setFeedback(result.matched ? `Match with ${user.email}` : `Liked ${user.email}`);
    } catch (likeError) {
      setError(likeError instanceof Error ? likeError.message : "could not send like");
    }
  }

  async function onBlock(user: DiscoverUser) {
    try {
      await block(token, user.id);
      setUsers((prev) => prev.filter((candidate) => candidate.id !== user.id));
      setFeedback(`Blocked ${user.email}`);
    } catch (blockError) {
      setError(blockError instanceof Error ? blockError.message : "could not block user");
    }
  }

  if (loading) {
    return (
      <SafeAreaView style={styles.container}>
        <ActivityIndicator size="large" color="#111827" />
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView style={styles.container}>
      <View style={styles.header}>
        <Text style={styles.title}>Discover</Text>
        <Pressable onPress={load}>
          <Text style={styles.reload}>Reload</Text>
        </Pressable>
      </View>

      {error ? <Text style={styles.error}>{error}</Text> : null}
      {feedback ? <Text style={styles.feedback}>{feedback}</Text> : null}

      <FlatList
        contentContainerStyle={styles.list}
        data={users}
        keyExtractor={(item) => item.id}
        ListEmptyComponent={<Text style={styles.empty}>No more profiles right now.</Text>}
        renderItem={({ item }) => (
          <View style={styles.card}>
            <Text style={styles.email}>{item.email}</Text>
            <Text style={styles.id}>{item.id}</Text>
            <View style={styles.actions}>
              <Pressable onPress={() => onLike(item)} style={styles.likeButton}>
                <Text style={styles.likeText}>Like</Text>
              </Pressable>
              <Pressable onPress={() => onBlock(item)} style={styles.blockButton}>
                <Text style={styles.blockText}>Block</Text>
              </Pressable>
            </View>
          </View>
        )}
      />
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#f4f7ff",
    paddingHorizontal: 16,
    paddingTop: 8
  },
  header: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
    marginBottom: 12
  },
  title: {
    fontSize: 28,
    fontWeight: "700"
  },
  reload: {
    color: "#2563eb",
    fontWeight: "600"
  },
  list: {
    paddingBottom: 24
  },
  card: {
    backgroundColor: "#ffffff",
    borderRadius: 14,
    padding: 16,
    marginBottom: 12,
    elevation: 2
  },
  email: {
    fontSize: 18,
    fontWeight: "700"
  },
  id: {
    marginTop: 6,
    fontSize: 12,
    color: "#6b7280"
  },
  likeButton: {
    backgroundColor: "#111827",
    paddingVertical: 10,
    borderRadius: 8,
    alignItems: "center",
    flex: 1
  },
  likeText: {
    color: "#ffffff",
    fontWeight: "700"
  },
  actions: {
    marginTop: 14,
    flexDirection: "row",
    gap: 10
  },
  blockButton: {
    backgroundColor: "#b91c1c",
    paddingVertical: 10,
    borderRadius: 8,
    alignItems: "center",
    flex: 1
  },
  blockText: {
    color: "#ffffff",
    fontWeight: "700"
  },
  error: {
    color: "#b91c1c",
    marginBottom: 8
  },
  feedback: {
    color: "#047857",
    marginBottom: 8
  },
  empty: {
    textAlign: "center",
    color: "#6b7280",
    marginTop: 24
  }
});
