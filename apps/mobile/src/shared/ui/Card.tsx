import React from "react";
import { StyleSheet, View, type StyleProp, type ViewProps, type ViewStyle } from "react-native";
import { colors, radii, shadows, spacing } from "./tokens";

export type CardVariant = "default" | "muted" | "accent";
export type CardPadding = "sm" | "md" | "lg" | "xl";

export type CardProps = Omit<ViewProps, "style"> & {
  variant?: CardVariant;
  padding?: CardPadding;
  style?: StyleProp<ViewStyle>;
};

const styles = StyleSheet.create({
  base: {
    borderRadius: radii.lg,
    borderWidth: 1,
    borderColor: colors.border,
    backgroundColor: colors.surface,
    ...shadows.card
  },
  muted: {
    backgroundColor: colors.surfaceMuted
  },
  accent: {
    backgroundColor: "#eff6ff",
    borderColor: "#bfdbfe"
  },
  paddingSm: {
    padding: spacing.md
  },
  paddingMd: {
    padding: spacing.lg
  },
  paddingLg: {
    padding: spacing.xl
  },
  paddingXl: {
    padding: spacing.xxl
  }
});

const variantStyles: Record<CardVariant, ViewStyle | null> = {
  default: null,
  muted: styles.muted,
  accent: styles.accent
};

const paddingStyles: Record<CardPadding, ViewStyle> = {
  sm: styles.paddingSm,
  md: styles.paddingMd,
  lg: styles.paddingLg,
  xl: styles.paddingXl
};

export function Card({ children, padding = "md", style, variant = "default", ...props }: CardProps) {
  return (
    <View style={[styles.base, variantStyles[variant], paddingStyles[padding], style]} {...props}>
      {children}
    </View>
  );
}
