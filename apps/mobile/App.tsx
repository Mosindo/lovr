import React, { useEffect, useState } from "react";
import { NavigationContainer } from "@react-navigation/native";
import { createBottomTabNavigator } from "@react-navigation/bottom-tabs";
import { StatusBar } from "expo-status-bar";
import { ActivityIndicator, SafeAreaView, StyleSheet } from "react-native";
import DiscoverScreen from "./src/screens/DiscoverScreen";
import MatchesScreen from "./src/screens/MatchesScreen";
import ChatsScreen from "./src/screens/ChatsScreen";
import AuthScreen from "./src/screens/AuthScreen";
import AccountScreen from "./src/screens/AccountScreen";
import { me, type AuthUser } from "./src/api/auth";
import { clearToken, getToken, saveToken } from "./src/auth/tokenStore";

export type RootTabParamList = {
  Discover: undefined;
  Matches: undefined;
  Chats: undefined;
  Account: undefined;
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
      <Tab.Navigator>
        <Tab.Screen name="Discover">{() => <DiscoverScreen token={token} />}</Tab.Screen>
        <Tab.Screen name="Matches">{() => <MatchesScreen token={token} />}</Tab.Screen>
        <Tab.Screen name="Chats">{() => <ChatsScreen currentUserId={user.id} token={token} />}</Tab.Screen>
        <Tab.Screen name="Account">
          {() => <AccountScreen onLogout={handleLogout} user={user} />}
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
