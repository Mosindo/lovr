import React from "react";
import { SafeAreaView, StyleSheet, Text, View } from "react-native";

export default function MatchesScreen() {
  return (
    <SafeAreaView style={styles.container}>
      <View style={styles.card}>
        <Text style={styles.title}>Matches</Text>
        <Text style={styles.subtitle}>Retrouve tes matchs ici.</Text>
      </View>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#fff7f7",
    justifyContent: "center",
    alignItems: "center"
  },
  card: {
    width: "88%",
    padding: 20,
    borderRadius: 16,
    backgroundColor: "white",
    elevation: 2
  },
  title: {
    fontSize: 28,
    fontWeight: "700",
    marginBottom: 8
  },
  subtitle: {
    color: "#4b5563",
    fontSize: 16
  }
});
