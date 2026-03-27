import React, { type ReactNode } from "react";
import {
  ActivityIndicator,
  Pressable,
  StyleSheet,
  View,
  type PressableProps,
  type StyleProp,
  type TextStyle,
  type ViewStyle
} from "react-native";
import { Text } from "./Text";
import { colors, radii, spacing } from "./tokens";

export type ButtonVariant = "primary" | "secondary" | "outline" | "ghost";
export type ButtonSize = "sm" | "md" | "lg";

export type ButtonProps = Omit<PressableProps, "children" | "style"> & {
  children?: ReactNode;
  label?: string;
  loading?: boolean;
  variant?: ButtonVariant;
  size?: ButtonSize;
  fullWidth?: boolean;
  style?: StyleProp<ViewStyle>;
  textStyle?: StyleProp<TextStyle>;
};

const styles = StyleSheet.create({
  base: {
    borderRadius: radii.md,
    alignItems: "center",
    justifyContent: "center"
  },
  fullWidth: {
    width: "100%"
  },
  small: {
    minHeight: 40,
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm
  },
  medium: {
    minHeight: 48,
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.md
  },
  large: {
    minHeight: 54,
    paddingHorizontal: spacing.xl,
    paddingVertical: spacing.md
  },
  primary: {
    backgroundColor: colors.primary
  },
  secondary: {
    backgroundColor: colors.secondary
  },
  outline: {
    backgroundColor: colors.surface,
    borderWidth: 1,
    borderColor: colors.border
  },
  ghost: {
    backgroundColor: "transparent"
  },
  disabled: {
    opacity: 0.6
  },
  pressed: {
    opacity: 0.88
  },
  content: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "center",
    gap: spacing.sm
  }
});

const sizeStyles: Record<ButtonSize, ViewStyle> = {
  sm: styles.small,
  md: styles.medium,
  lg: styles.large
};

const variantStyles: Record<ButtonVariant, ViewStyle> = {
  primary: styles.primary,
  secondary: styles.secondary,
  outline: styles.outline,
  ghost: styles.ghost
};

const textToneByVariant: Record<ButtonVariant, "inverse" | "primary" | "secondary"> = {
  primary: "inverse",
  secondary: "inverse",
  outline: "primary",
  ghost: "secondary"
};

const spinnerColorByVariant: Record<ButtonVariant, string> = {
  primary: colors.inverse,
  secondary: colors.inverse,
  outline: colors.primary,
  ghost: colors.secondary
};

export function Button({
  children,
  disabled,
  fullWidth = false,
  label,
  loading = false,
  size = "md",
  style,
  textStyle,
  variant = "primary",
  ...props
}: ButtonProps) {
  const buttonDisabled = disabled || loading;
  const textTone = textToneByVariant[variant];
  const labelNode = typeof children === "string" ? children : label;

  return (
    <Pressable
      accessibilityRole="button"
      disabled={buttonDisabled}
      style={({ pressed }) => [
        styles.base,
        sizeStyles[size],
        variantStyles[variant],
        fullWidth ? styles.fullWidth : null,
        buttonDisabled ? styles.disabled : null,
        pressed ? styles.pressed : null,
        style
      ]}
      {...props}
    >
      <View style={styles.content}>
        {loading ? <ActivityIndicator color={spinnerColorByVariant[variant]} size="small" /> : null}
        {labelNode ? (
          <Text style={textStyle} tone={textTone} variant="button" weight="bold">
            {labelNode}
          </Text>
        ) : (
          children
        )}
      </View>
    </Pressable>
  );
}
