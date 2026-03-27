import React, { useEffect } from "react";
import { StatusBar } from "expo-status-bar";
import { StyleSheet, View } from "react-native";
import AuthScreen from "./src/screens/AuthScreen";
import { AuthProvider, useAuth } from "./src/hooks/useAuth";
import { BottomNavigation, SafeAreaLayout } from "./src/shared/layout";
import { clearGlobalError, ErrorView, LoadingView, useGlobalFeedback } from "./src/shared/feedback";
import { colors, spacing } from "./src/shared/ui";

function AppShell() {
  const { accessToken, isAuthenticated, isBooting, user } = useAuth();
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
        <StatusBar style="dark" />
        <LoadingView fullScreen label="Restoring session..." />
      </SafeAreaLayout>
    );
  }

  if (!isAuthenticated || !accessToken || !user) {
    return (
      <>
        <StatusBar style="dark" />
        <AuthScreen />
      </>
    );
  }

  return (
    <>
      <StatusBar style="dark" />
      <BottomNavigation accessToken={accessToken} key={user.id} user={user} />
      {globalFeedback.error ? (
        <View pointerEvents="box-none" style={styles.bannerWrap}>
          <ErrorView
            actionLabel="Dismiss"
            compact
            message={globalFeedback.error}
            onAction={clearGlobalError}
            style={styles.banner}
            title="Request issue"
          />
        </View>
      ) : null}
      {globalFeedback.loadingCount > 0 && !isBooting ? (
        <View style={styles.overlay}>
          <LoadingView label={globalFeedback.loadingLabel ?? "Working..."} style={styles.overlayCard} />
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
  bannerWrap: {
    position: "absolute",
    left: spacing.lg,
    right: spacing.lg,
    bottom: spacing.xxl
  },
  banner: {},
  overlay: {
    position: "absolute",
    top: 0,
    right: 0,
    bottom: 0,
    left: 0,
    backgroundColor: colors.overlay,
    alignItems: "center",
    justifyContent: "center",
    paddingHorizontal: spacing.xxl
  },
  overlayCard: {
    width: "100%",
    maxWidth: 320
  }
});
