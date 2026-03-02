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
import { block, matches, type DiscoverUser } from "../api/auth";

type MatchesScreenProps = {
  token: string;
};

export default function MatchesScreen({ token }: MatchesScreenProps) {
  const [items, setItems] = useState<DiscoverUser[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [feedback, setFeedback] = useState<string | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    setFeedback(null);
    try {
      const nextMatches = await matches(token);
      setItems(nextMatches);
    } catch (loadError) {
      setError(loadError instanceof Error ? loadError.message : "could not load matches");
    } finally {
      setLoading(false);
    }
  }, [token]);

  useEffect(() => {
    load();
  }, [load]);

  async function onBlock(user: DiscoverUser) {
    try {
      await block(token, user.id);
      setItems((prev) => prev.filter((candidate) => candidate.id !== user.id));
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
      <View style={styles.header} testID="matches-screen">
        <Text style={styles.title}>Matches</Text>
        <Pressable onPress={load} testID="matches-reload-button">
          <Text style={styles.reload}>Reload</Text>
        </Pressable>
      </View>
      {error ? <Text style={styles.error}>{error}</Text> : null}
      {feedback ? <Text style={styles.feedback} testID="matches-feedback">{feedback}</Text> : null}

      <FlatList
        contentContainerStyle={styles.list}
        data={items}
        keyExtractor={(item) => item.id}
        ListEmptyComponent={<Text style={styles.empty}>No matches yet.</Text>}
        renderItem={({ item }) => (
          <View style={styles.card}>
            <Text style={styles.email}>{item.email}</Text>
            <Text style={styles.id}>{item.id}</Text>
            <Pressable onPress={() => onBlock(item)} style={styles.blockButton} testID={`matches-block-${item.id}`}>
              <Text style={styles.blockText}>Block</Text>
            </Pressable>
          </View>
        )}
      />
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#fff7f7",
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
  empty: {
    textAlign: "center",
    color: "#6b7280",
    marginTop: 24
  },
  error: {
    color: "#b91c1c",
    marginBottom: 8
  },
  feedback: {
    color: "#047857",
    marginBottom: 8
  },
  blockButton: {
    marginTop: 12,
    backgroundColor: "#b91c1c",
    paddingVertical: 10,
    borderRadius: 8,
    alignItems: "center"
  },
  blockText: {
    color: "#ffffff",
    fontWeight: "700"
  }
});
