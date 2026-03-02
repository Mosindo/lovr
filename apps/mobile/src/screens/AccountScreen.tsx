import React from "react";
import { Pressable, SafeAreaView, StyleSheet, Text, View } from "react-native";
import { type AuthUser } from "../api/auth";

type AccountScreenProps = {
  user: AuthUser;
  onLogout: () => void;
};

export default function AccountScreen({ user, onLogout }: AccountScreenProps) {
  return (
    <SafeAreaView style={styles.container} testID="account-screen">
      <View style={styles.card}>
        <Text style={styles.title}>Account</Text>
        <Text style={styles.label}>Email</Text>
        <Text style={styles.value}>{user.email}</Text>

        <Text style={styles.label}>User ID</Text>
        <Text style={styles.value}>{user.id}</Text>

        <Pressable onPress={onLogout} style={styles.button} testID="account-logout-button">
          <Text style={styles.buttonText}>Logout</Text>
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
    backgroundColor: "#eef2ff"
  },
  card: {
    width: "88%",
    borderRadius: 16,
    backgroundColor: "#ffffff",
    padding: 20,
    elevation: 2
  },
  title: {
    fontSize: 24,
    fontWeight: "700",
    marginBottom: 14
  },
  label: {
    marginTop: 8,
    color: "#6b7280",
    fontSize: 12,
    textTransform: "uppercase"
  },
  value: {
    marginTop: 2,
    fontSize: 14,
    color: "#111827"
  },
  button: {
    marginTop: 18,
    borderRadius: 10,
    backgroundColor: "#dc2626",
    paddingVertical: 12,
    alignItems: "center"
  },
  buttonText: {
    color: "#ffffff",
    fontWeight: "700"
  }
});
