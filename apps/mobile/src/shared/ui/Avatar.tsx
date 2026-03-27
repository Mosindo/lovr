import React, { useMemo } from "react";
import { Image, StyleSheet, View, type ImageStyle, type StyleProp, type TextStyle, type ViewStyle } from "react-native";
import { Text } from "./Text";
import { colors } from "./tokens";

export type AvatarProps = {
  name?: string;
  size?: number;
  style?: StyleProp<ViewStyle>;
  textStyle?: StyleProp<TextStyle>;
  uri?: string | null;
};

function getInitials(name?: string): string {
  if (!name) {
    return "?";
  }

  const parts = name
    .trim()
    .split(/\s+/)
    .filter(Boolean);

  if (parts.length === 0) {
    return "?";
  }

  if (parts.length === 1) {
    return parts[0].slice(0, 2).toUpperCase();
  }

  return `${parts[0][0] ?? ""}${parts[1][0] ?? ""}`.toUpperCase();
}

const styles = StyleSheet.create({
  container: {
    alignItems: "center",
    justifyContent: "center",
    backgroundColor: colors.secondary,
    overflow: "hidden"
  },
  image: {
    width: "100%",
    height: "100%"
  }
});

export function Avatar({ name, size = 44, style, textStyle, uri }: AvatarProps) {
  const initials = useMemo(() => getInitials(name), [name]);

  return (
    <View
      style={[
        styles.container,
        {
          width: size,
          height: size,
          borderRadius: size / 2
        },
        style
      ]}
    >
      {uri ? (
        <Image source={{ uri }} style={styles.image as StyleProp<ImageStyle>} />
      ) : (
        <Text
          style={textStyle}
          tone="inverse"
          variant={size >= 52 ? "title" : "label"}
          weight="bold"
        >
          {initials}
        </Text>
      )}
    </View>
  );
}
