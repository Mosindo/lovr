import React, { useMemo, useState } from "react";
import { Pressable, StyleSheet } from "react-native";
import { useLogin, useRegister } from "../hooks/useAuth";
import { Button, Card, Input, Text, colors, spacing } from "../shared/ui";
import { Header, ScreenContainer } from "../shared/layout";

type Mode = "login" | "register";

export default function AuthScreen() {
  const [mode, setMode] = useState<Mode>("login");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const loginMutation = useLogin();
  const registerMutation = useRegister();

  const title = useMemo(() => (mode === "login" ? "Login" : "Create account"), [mode]);
  const switchLabel = mode === "login" ? "Need an account? Register" : "Already have an account? Login";
  const submitting = loginMutation.isPending || registerMutation.isPending;

  async function onSubmit() {
    const normalizedEmail = email.trim().toLowerCase();
    if (!normalizedEmail || !password) {
      setError("email and password are required");
      return;
    }

    setError(null);

    try {
      if (mode === "login") {
        await loginMutation.mutateAsync({ email: normalizedEmail, password });
      } else {
        await registerMutation.mutateAsync({ email: normalizedEmail, password });
      }
    } catch (submitError) {
      setError(submitError instanceof Error ? submitError.message : "authentication failed");
    }
  }

  return (
    <ScreenContainer centered contentMaxWidth={440}>
      <Card padding="lg" style={styles.card}>
        <Header
          centered
          eyebrow="Access"
          style={styles.header}
          subtitle="Use your workspace credentials to continue."
          title={title}
        />

        <Input
          autoCapitalize="none"
          autoComplete="email"
          containerStyle={styles.field}
          keyboardType="email-address"
          label="Email"
          onChangeText={setEmail}
          placeholder="Email"
          testID="auth-email-input"
          value={email}
        />

        <Input
          autoCapitalize="none"
          containerStyle={styles.field}
          label="Password"
          onChangeText={setPassword}
          placeholder="Password"
          secureTextEntry
          testID="auth-password-input"
          value={password}
        />

        {error ? (
          <Text style={styles.error} tone="danger" variant="label" weight="medium">
            {error}
          </Text>
        ) : null}

        <Button
          fullWidth
          label={mode === "login" ? "Login" : "Register"}
          loading={submitting}
          onPress={onSubmit}
          testID="auth-submit-button"
        />

        <Pressable
          disabled={submitting}
          onPress={() => {
            setMode((prev) => (prev === "login" ? "register" : "login"));
            setError(null);
          }}
          testID="auth-switch-mode-button"
        >
          <Text style={styles.switchText} tone="primary" variant="label" weight="semibold">
            {switchLabel}
          </Text>
        </Pressable>
      </Card>
    </ScreenContainer>
  );
}

const styles = StyleSheet.create({
  card: {
    width: "100%"
  },
  header: {
    marginBottom: spacing.lg
  },
  field: {
    marginBottom: spacing.md
  },
  switchText: {
    marginTop: spacing.lg,
    textAlign: "center"
  },
  error: {
    marginBottom: spacing.sm
  }
});
