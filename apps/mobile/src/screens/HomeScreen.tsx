import React, { useCallback, useEffect, useMemo, useState } from "react";
import {
  ActivityIndicator,
  FlatList,
  Pressable,
  SafeAreaView,
  StyleSheet,
  Text,
  TextInput,
  View
} from "react-native";
import { createPost, listPosts, type Post } from "../api/platform";

type HomeScreenProps = {
  token: string;
  currentUserId: string;
};

function formatPostDate(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "Unknown date";
  }
  return date.toLocaleString();
}

export default function HomeScreen({ token, currentUserId }: HomeScreenProps) {
  const [posts, setPosts] = useState<Post[]>([]);
  const [title, setTitle] = useState("");
  const [body, setBody] = useState("");
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [feedback, setFeedback] = useState<string | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const nextPosts = await listPosts(token);
      setPosts(nextPosts);
    } catch (loadError) {
      setError(loadError instanceof Error ? loadError.message : "could not load posts");
    } finally {
      setLoading(false);
    }
  }, [token]);

  useEffect(() => {
    load();
  }, [load]);

  async function onSubmit() {
    const normalizedTitle = title.trim();
    const normalizedBody = body.trim();
    if (!normalizedTitle || !normalizedBody) {
      setError("title and body are required");
      return;
    }

    setSubmitting(true);
    setError(null);
    setFeedback(null);
    try {
      const created = await createPost(token, normalizedTitle, normalizedBody);
      setPosts((prev) => [created, ...prev.filter((item) => item.id !== created.id)]);
      setTitle("");
      setBody("");
      setFeedback("Post published.");
    } catch (submitError) {
      setError(submitError instanceof Error ? submitError.message : "could not create post");
    } finally {
      setSubmitting(false);
    }
  }

  const emptyState = useMemo(() => {
    if (loading) {
      return null;
    }
    return <Text style={styles.empty}>No posts yet. Publish the first update.</Text>;
  }, [loading]);

  if (loading && posts.length === 0) {
    return (
      <SafeAreaView style={styles.loaderContainer}>
        <ActivityIndicator color="#0f172a" size="large" />
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView style={styles.container}>
      <View style={styles.header} testID="home-screen">
        <View>
          <Text style={styles.eyebrow}>Workspace</Text>
          <Text style={styles.title}>Home</Text>
        </View>
        <Pressable disabled={loading || submitting} onPress={load} testID="home-reload-button">
          <Text style={[styles.reload, loading || submitting ? styles.mutedAction : null]}>Reload</Text>
        </Pressable>
      </View>

      <View style={styles.composer}>
        <Text style={styles.sectionTitle}>Publish an update</Text>
        <TextInput
          onChangeText={setTitle}
          placeholder="Post title"
          style={styles.titleInput}
          testID="home-post-title-input"
          value={title}
        />
        <TextInput
          multiline
          numberOfLines={4}
          onChangeText={setBody}
          placeholder="Share progress, notes, or an announcement..."
          style={styles.bodyInput}
          testID="home-post-body-input"
          value={body}
        />
        <Pressable
          disabled={submitting || loading}
          onPress={onSubmit}
          style={[styles.publishButton, submitting || loading ? styles.publishButtonDisabled : null]}
          testID="home-post-submit-button"
        >
          <Text style={styles.publishText}>{submitting ? "Publishing..." : "Publish"}</Text>
        </Pressable>
      </View>

      {error ? <Text style={styles.error}>{error}</Text> : null}
      {feedback ? <Text style={styles.feedback}>{feedback}</Text> : null}

      <FlatList
        contentContainerStyle={styles.list}
        data={posts}
        keyExtractor={(item) => item.id}
        ListEmptyComponent={emptyState}
        renderItem={({ item }) => (
          <View style={styles.card}>
            <View style={styles.cardHeader}>
              <Text style={styles.cardTitle}>{item.title}</Text>
              <Text style={styles.badge}>{item.authorUserId === currentUserId ? "You" : "Member"}</Text>
            </View>
            <Text style={styles.cardBody}>{item.body}</Text>
            <Text style={styles.meta}>{formatPostDate(item.createdAt)}</Text>
          </View>
        )}
      />
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#f6f7f2",
    paddingHorizontal: 16,
    paddingTop: 12
  },
  loaderContainer: {
    flex: 1,
    alignItems: "center",
    justifyContent: "center",
    backgroundColor: "#f6f7f2"
  },
  header: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
    marginBottom: 16
  },
  eyebrow: {
    fontSize: 12,
    fontWeight: "700",
    textTransform: "uppercase",
    letterSpacing: 1,
    color: "#7c6f64"
  },
  title: {
    fontSize: 30,
    fontWeight: "800",
    color: "#1f2937"
  },
  reload: {
    color: "#166534",
    fontWeight: "700"
  },
  mutedAction: {
    color: "#9ca3af"
  },
  composer: {
    backgroundColor: "#fffdf7",
    borderRadius: 18,
    padding: 16,
    borderWidth: 1,
    borderColor: "#e7e5d4",
    marginBottom: 12
  },
  sectionTitle: {
    fontSize: 17,
    fontWeight: "700",
    color: "#1f2937",
    marginBottom: 12
  },
  titleInput: {
    borderWidth: 1,
    borderColor: "#d6d3c4",
    borderRadius: 12,
    paddingHorizontal: 12,
    paddingVertical: 10,
    backgroundColor: "#ffffff",
    marginBottom: 10
  },
  bodyInput: {
    borderWidth: 1,
    borderColor: "#d6d3c4",
    borderRadius: 12,
    paddingHorizontal: 12,
    paddingVertical: 12,
    minHeight: 96,
    textAlignVertical: "top",
    backgroundColor: "#ffffff"
  },
  publishButton: {
    marginTop: 12,
    borderRadius: 12,
    backgroundColor: "#14532d",
    paddingVertical: 12,
    alignItems: "center"
  },
  publishButtonDisabled: {
    opacity: 0.7
  },
  publishText: {
    color: "#ffffff",
    fontWeight: "700"
  },
  error: {
    color: "#b91c1c",
    marginBottom: 8
  },
  feedback: {
    color: "#166534",
    marginBottom: 8
  },
  list: {
    paddingBottom: 32
  },
  empty: {
    color: "#6b7280",
    marginTop: 32,
    textAlign: "center"
  },
  card: {
    backgroundColor: "#ffffff",
    borderRadius: 18,
    padding: 16,
    marginBottom: 12,
    borderWidth: 1,
    borderColor: "#ece7da"
  },
  cardHeader: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
    marginBottom: 10,
    gap: 10
  },
  cardTitle: {
    flex: 1,
    fontSize: 18,
    fontWeight: "700",
    color: "#111827"
  },
  badge: {
    backgroundColor: "#ecfccb",
    color: "#365314",
    paddingHorizontal: 10,
    paddingVertical: 4,
    borderRadius: 999,
    fontSize: 12,
    fontWeight: "700"
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
