import React, { useMemo, useState } from "react";
import {
  ActivityIndicator,
  Pressable,
  SafeAreaView,
  StyleSheet,
  Text,
  TextInput,
  View
} from "react-native";
import { login, register, type AuthUser } from "../api/auth";

type Mode = "login" | "register";

type AuthScreenProps = {
  onAuthenticated: (token: string, user: AuthUser) => void;
};

export default function AuthScreen({ onAuthenticated }: AuthScreenProps) {
  const [mode, setMode] = useState<Mode>("login");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const title = useMemo(() => (mode === "login" ? "Login" : "Create account"), [mode]);
  const switchLabel = mode === "login" ? "Need an account? Register" : "Already have an account? Login";

  async function onSubmit() {
    const normalizedEmail = email.trim().toLowerCase();
    if (!normalizedEmail || !password) {
      setError("email and password are required");
      return;
    }

    setSubmitting(true);
    setError(null);

    try {
      const authResult =
        mode === "login"
          ? await login(normalizedEmail, password)
          : await register(normalizedEmail, password);
      onAuthenticated(authResult.token, authResult.user);
    } catch (submitError) {
      setError(submitError instanceof Error ? submitError.message : "authentication failed");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <SafeAreaView style={styles.container}>
      <View style={styles.card}>
        <Text style={styles.title}>{title}</Text>

        <TextInput
          autoCapitalize="none"
          autoComplete="email"
          keyboardType="email-address"
          onChangeText={setEmail}
          placeholder="Email"
          style={styles.input}
          testID="auth-email-input"
          value={email}
        />

        <TextInput
          autoCapitalize="none"
          onChangeText={setPassword}
          placeholder="Password"
          secureTextEntry
          style={styles.input}
          testID="auth-password-input"
          value={password}
        />

        {error ? <Text style={styles.error}>{error}</Text> : null}

        <Pressable disabled={submitting} onPress={onSubmit} style={styles.button} testID="auth-submit-button">
          {submitting ? (
            <ActivityIndicator color="#ffffff" />
          ) : (
            <Text style={styles.buttonText}>{mode === "login" ? "Login" : "Register"}</Text>
          )}
        </Pressable>

        <Pressable
          disabled={submitting}
          onPress={() => {
            setMode((prev) => (prev === "login" ? "register" : "login"));
            setError(null);
          }}
          testID="auth-switch-mode-button"
        >
          <Text style={styles.switchText}>{switchLabel}</Text>
        </Pressable>
      </View>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    alignItems: "center",
    justifyContent: "center",
    backgroundColor: "#f3f4f6"
  },
  card: {
    width: "88%",
    borderRadius: 16,
    backgroundColor: "#ffffff",
    padding: 20,
    elevation: 3
  },
  title: {
    fontSize: 24,
    fontWeight: "700",
    marginBottom: 16
  },
  input: {
    borderWidth: 1,
    borderColor: "#d1d5db",
    borderRadius: 10,
    paddingHorizontal: 12,
    paddingVertical: 10,
    marginBottom: 10
  },
  button: {
    marginTop: 4,
    borderRadius: 10,
    backgroundColor: "#111827",
    paddingVertical: 12,
    alignItems: "center"
  },
  buttonText: {
    color: "#ffffff",
    fontWeight: "700"
  },
  switchText: {
    marginTop: 14,
    color: "#2563eb",
    textAlign: "center"
  },
  error: {
    marginBottom: 6,
    color: "#b91c1c"
  }
});
