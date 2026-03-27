import React from "react";
import { StyleSheet, Text as RNText, type StyleProp, type TextProps as RNTextProps, type TextStyle } from "react-native";
import { colors, fontSizes } from "./tokens";

export type TextVariant = "body" | "label" | "title" | "heading" | "caption" | "eyebrow" | "button";
export type TextTone = "default" | "muted" | "primary" | "secondary" | "danger" | "success" | "inverse";
export type TextWeight = "regular" | "medium" | "semibold" | "bold";

export type TextProps = RNTextProps & {
  variant?: TextVariant;
  tone?: TextTone;
  weight?: TextWeight;
  style?: StyleProp<TextStyle>;
};

const styles = StyleSheet.create({
  base: {
    color: colors.text,
    fontSize: fontSizes.md
  },
  body: {
    fontSize: fontSizes.md,
    lineHeight: 22
  },
  label: {
    fontSize: fontSizes.sm,
    lineHeight: 20
  },
  title: {
    fontSize: fontSizes.xxl,
    lineHeight: 36
  },
  heading: {
    fontSize: fontSizes.xl,
    lineHeight: 30
  },
  caption: {
    fontSize: fontSizes.xs,
    lineHeight: 18
  },
  eyebrow: {
    fontSize: fontSizes.xs,
    lineHeight: 16,
    textTransform: "uppercase",
    letterSpacing: 0.8
  },
  button: {
    fontSize: fontSizes.md,
    lineHeight: 20
  },
  regular: {
    fontWeight: "400"
  },
  medium: {
    fontWeight: "500"
  },
  semibold: {
    fontWeight: "600"
  },
  bold: {
    fontWeight: "700"
  },
  defaultTone: {
    color: colors.text
  },
  mutedTone: {
    color: colors.textMuted
  },
  primaryTone: {
    color: colors.primary
  },
  secondaryTone: {
    color: colors.secondary
  },
  dangerTone: {
    color: colors.danger
  },
  successTone: {
    color: colors.success
  },
  inverseTone: {
    color: colors.inverse
  }
});

const variantStyles: Record<TextVariant, TextStyle> = {
  body: styles.body,
  label: styles.label,
  title: styles.title,
  heading: styles.heading,
  caption: styles.caption,
  eyebrow: styles.eyebrow,
  button: styles.button
};

const weightStyles: Record<TextWeight, TextStyle> = {
  regular: styles.regular,
  medium: styles.medium,
  semibold: styles.semibold,
  bold: styles.bold
};

const toneStyles: Record<TextTone, TextStyle> = {
  default: styles.defaultTone,
  muted: styles.mutedTone,
  primary: styles.primaryTone,
  secondary: styles.secondaryTone,
  danger: styles.dangerTone,
  success: styles.successTone,
  inverse: styles.inverseTone
};

export function Text({
  children,
  style,
  tone = "default",
  variant = "body",
  weight = "regular",
  ...props
}: TextProps) {
  return (
    <RNText style={[styles.base, variantStyles[variant], weightStyles[weight], toneStyles[tone], style]} {...props}>
      {children}
    </RNText>
  );
}
