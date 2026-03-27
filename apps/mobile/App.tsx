import React, { useEffect } from "react";
import { StatusBar } from "expo-status-bar";
import { Pressable, StyleSheet, View } from "react-native";
import AuthScreen from "./src/screens/AuthScreen";
import { AuthProvider, useAuth } from "./src/hooks/useAuth";
import { BottomNavigation, SafeAreaLayout } from "./src/shared/layout";
import { clearGlobalError, useGlobalFeedback } from "./src/shared/feedback";
import { Card, Loader, Text, colors, spacing } from "./src/shared/ui";

function AppShell() {
  const { accessToken, authError, isAuthenticated, isBooting, user } = useAuth();
  const globalFeedback = useGlobalFeedback();

  useEffect(() => {
    if (!globalFeedback.error) {
      return;
    }

    const timeoutId = setTimeout(() => {
      clearGlobalError();
    }, 5000);

    return () => {
      clearTimeout(timeoutId);
    };
  }, [globalFeedback.error]);

  if (isBooting) {
    return (
      <SafeAreaLayout style={styles.bootContainer}>
        <StatusBar style="auto" />
        <Loader fullScreen label="Restoring session..." />
      </SafeAreaLayout>
    );
  }

  if (!isAuthenticated || !accessToken || !user) {
    return (
      <>
        <StatusBar style="auto" />
        <AuthScreen />
        {authError ? (
          <View style={styles.authErrorWrap}>
            <Text tone="danger" variant="caption" weight="medium">
              {authError}
            </Text>
          </View>
        ) : null}
      </>
    );
  }

  return (
    <>
      <StatusBar style="auto" />
      <BottomNavigation accessToken={accessToken} key={user.id} user={user} />
      {globalFeedback.error ? (
        <View pointerEvents="box-none" style={styles.bannerWrap}>
          <Card padding="sm" style={styles.banner}>
            <Text style={styles.bannerText} tone="danger" variant="label" weight="medium">
              {globalFeedback.error}
            </Text>
            <Pressable onPress={clearGlobalError}>
              <Text tone="primary" variant="caption" weight="bold">
                Dismiss
              </Text>
            </Pressable>
          </Card>
        </View>
      ) : null}
      {globalFeedback.loadingCount > 0 && !isBooting ? (
        <View style={styles.overlay}>
          <Card padding="md" style={styles.overlayCard}>
            <Loader label={globalFeedback.loadingLabel ?? "Working..."} />
          </Card>
        </View>
      ) : null}
    </>
  );
}

export default function App() {
  return (
    <AuthProvider>
      <AppShell />
    </AuthProvider>
  );
}

const styles = StyleSheet.create({
  bootContainer: {
    flex: 1,
    backgroundColor: colors.background
  },
  authErrorWrap: {
    position: "absolute",
    bottom: spacing.xxl,
    left: spacing.xxl,
    right: spacing.xxl,
    alignItems: "center"
  },
  bannerWrap: {
    position: "absolute",
    left: spacing.lg,
    right: spacing.lg,
    bottom: spacing.xxl
  },
  banner: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
    gap: spacing.md
  },
  bannerText: {
    flex: 1
  },
  overlay: {
    position: "absolute",
    top: 0,
    right: 0,
    bottom: 0,
    left: 0,
    backgroundColor: "rgba(15, 23, 42, 0.16)",
    alignItems: "center",
    justifyContent: "center",
    paddingHorizontal: spacing.xxl
  },
  overlayCard: {
    width: "100%",
    maxWidth: 320
  }
});
