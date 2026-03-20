import React from "react";
import { NavigationContainer } from "@react-navigation/native";
import { createBottomTabNavigator } from "@react-navigation/bottom-tabs";
import { StatusBar } from "expo-status-bar";
import { ActivityIndicator, SafeAreaView, StyleSheet } from "react-native";
import AuthScreen from "./src/screens/AuthScreen";
import ChatScreen from "./src/screens/ChatScreen";
import HomeScreen from "./src/screens/HomeScreen";
import NotificationsScreen from "./src/screens/NotificationsScreen";
import ProfileScreen from "./src/screens/ProfileScreen";
import { AuthProvider, useAuth } from "./src/hooks/useAuth";

export type RootTabParamList = {
  Home: undefined;
  Chat: undefined;
  Notifications: undefined;
  Profile: undefined;
};

const Tab = createBottomTabNavigator<RootTabParamList>();

function AppShell() {
  const { accessToken, authError, isAuthenticated, isBooting, user } = useAuth();

  if (isBooting) {
    return (
      <SafeAreaView style={styles.bootContainer}>
        <StatusBar style="auto" />
        <ActivityIndicator size="large" color="#111827" />
      </SafeAreaView>
    );
  }

  if (!isAuthenticated || !accessToken || !user) {
    return (
      <>
        <StatusBar style="auto" />
        <AuthScreen />
        {authError ? <SafeAreaView style={styles.authErrorWrap}><ActivityIndicator size="small" color="#111827" /></SafeAreaView> : null}
      </>
    );
  }

  return (
    <NavigationContainer>
      <StatusBar style="auto" />
      <Tab.Navigator screenOptions={{ headerShown: false }}>
        <Tab.Screen name="Home">
          {() => <HomeScreen currentUserId={user.id} token={accessToken} />}
        </Tab.Screen>
        <Tab.Screen name="Chat">
          {() => <ChatScreen currentUserId={user.id} token={accessToken} />}
        </Tab.Screen>
        <Tab.Screen name="Notifications">
          {() => <NotificationsScreen token={accessToken} />}
        </Tab.Screen>
        <Tab.Screen name="Profile">
          {() => <ProfileScreen token={accessToken} user={user} />}
        </Tab.Screen>
      </Tab.Navigator>
    </NavigationContainer>
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
    alignItems: "center",
    justifyContent: "center",
    backgroundColor: "#f9fafb"
  },
  authErrorWrap: {
    position: "absolute",
    bottom: 24,
    left: 24,
    right: 24,
    alignItems: "center"
  }
});
