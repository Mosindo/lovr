import React from "react";
import { NavigationContainer } from "@react-navigation/native";
import { createBottomTabNavigator } from "@react-navigation/bottom-tabs";
import { type AuthUser } from "../../api/auth";
import ChatScreen from "../../screens/ChatScreen";
import HomeScreen from "../../screens/HomeScreen";
import NotificationsScreen from "../../screens/NotificationsScreen";
import ProfileScreen from "../../screens/ProfileScreen";
import { colors, spacing } from "../ui";

export type RootTabParamList = {
  Home: undefined;
  Chat: undefined;
  Notifications: undefined;
  Profile: undefined;
};

type BottomNavigationProps = {
  accessToken: string;
  user: AuthUser;
};

const Tab = createBottomTabNavigator<RootTabParamList>();

export function BottomNavigation({ accessToken, user }: BottomNavigationProps) {
  return (
    <NavigationContainer>
      <Tab.Navigator
        initialRouteName="Home"
        screenOptions={{
          headerShown: false,
          tabBarActiveTintColor: colors.primary,
          tabBarInactiveTintColor: colors.textMuted,
          tabBarStyle: {
            backgroundColor: colors.surface,
            borderTopColor: colors.border,
            paddingTop: spacing.xs
          },
          tabBarLabelStyle: {
            paddingBottom: spacing.xs
          }
        }}
      >
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
