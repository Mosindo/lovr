import React from "react";
import { NavigationContainer } from "@react-navigation/native";
import { createBottomTabNavigator } from "@react-navigation/bottom-tabs";
import { StatusBar } from "expo-status-bar";
import DiscoverScreen from "./src/screens/DiscoverScreen";
import MatchesScreen from "./src/screens/MatchesScreen";

export type RootTabParamList = {
  Discover: undefined;
  Matches: undefined;
};

const Tab = createBottomTabNavigator<RootTabParamList>();

export default function App() {
  return (
    <NavigationContainer>
      <StatusBar style="auto" />
      <Tab.Navigator>
        <Tab.Screen name="Discover" component={DiscoverScreen} />
        <Tab.Screen name="Matches" component={MatchesScreen} />
      </Tab.Navigator>
    </NavigationContainer>
  );
}
