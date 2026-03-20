import React, { useEffect, useState } from "react";
import { NavigationContainer } from "@react-navigation/native";
import { createBottomTabNavigator } from "@react-navigation/bottom-tabs";
import { StatusBar } from "expo-status-bar";
import { ActivityIndicator, SafeAreaView, StyleSheet } from "react-native";
import AuthScreen from "./src/screens/AuthScreen";
import ChatScreen from "./src/screens/ChatScreen";
import HomeScreen from "./src/screens/HomeScreen";
import NotificationsScreen from "./src/screens/NotificationsScreen";
import ProfileScreen from "./src/screens/ProfileScreen";
import { me, type AuthUser } from "./src/api/auth";
import { clearToken, getToken, saveToken } from "./src/store/tokenStore";

export type RootTabParamList = {
  Home: undefined;
  Chat: undefined;
  Notifications: undefined;
  Profile: undefined;
};

const Tab = createBottomTabNavigator<RootTabParamList>();

export default function App() {
  const [booting, setBooting] = useState(true);
  const [token, setToken] = useState<string | null>(null);
  const [user, setUser] = useState<AuthUser | null>(null);

  useEffect(() => {
    let active = true;

    async function restoreSession() {
      try {
        const storedToken = await getToken();
        if (!storedToken || !active) {
          return;
        }

        const meUser = await me(storedToken);
        if (!active) {
          return;
        }

        setToken(storedToken);
        setUser(meUser);
      } catch {
        await clearToken();
      } finally {
        if (active) {
          setBooting(false);
        }
      }
    }

    restoreSession();

    return () => {
      active = false;
    };
  }, []);

  async function handleAuthenticated(nextToken: string, nextUser: AuthUser) {
    await saveToken(nextToken);
    setToken(nextToken);
    setUser(nextUser);
  }

  async function handleLogout() {
    await clearToken();
    setToken(null);
    setUser(null);
  }

  if (booting) {
    return (
      <SafeAreaView style={styles.bootContainer}>
        <StatusBar style="auto" />
        <ActivityIndicator size="large" color="#111827" />
      </SafeAreaView>
    );
  }

  if (!token || !user) {
    return (
      <>
        <StatusBar style="auto" />
        <AuthScreen onAuthenticated={handleAuthenticated} />
      </>
    );
  }

  return (
    <NavigationContainer>
      <StatusBar style="auto" />
      <Tab.Navigator screenOptions={{ headerShown: false }}>
        <Tab.Screen name="Home">
          {() => <HomeScreen currentUserId={user.id} token={token} />}
        </Tab.Screen>
        <Tab.Screen name="Chat">
          {() => <ChatScreen currentUserId={user.id} token={token} />}
        </Tab.Screen>
        <Tab.Screen name="Notifications">
          {() => <NotificationsScreen token={token} />}
        </Tab.Screen>
        <Tab.Screen name="Profile">
          {() => <ProfileScreen onLogout={handleLogout} token={token} user={user} />}
        </Tab.Screen>
      </Tab.Navigator>
    </NavigationContainer>
  );
}

const styles = StyleSheet.create({
  bootContainer: {
    flex: 1,
    alignItems: "center",
    justifyContent: "center",
    backgroundColor: "#f9fafb"
  }
});
