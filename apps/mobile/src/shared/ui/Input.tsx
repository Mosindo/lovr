import React from "react";
import {
  StyleSheet,
  TextInput as RNTextInput,
  View,
  type StyleProp,
  type TextInputProps as RNTextInputProps,
  type TextStyle,
  type ViewStyle
} from "react-native";
import { Text } from "./Text";
import { colors, fontSizes, radii, spacing } from "./tokens";

export type InputProps = Omit<RNTextInputProps, "style"> & {
  label?: string;
  helperText?: string;
  error?: string | null;
  style?: StyleProp<TextStyle>;
  containerStyle?: StyleProp<ViewStyle>;
};

const styles = StyleSheet.create({
  container: {
    width: "100%",
    gap: spacing.xs
  },
  input: {
    borderWidth: 1,
    borderColor: colors.border,
    borderRadius: radii.md,
    backgroundColor: colors.surface,
    color: colors.text,
    fontSize: fontSizes.md,
    minHeight: 48,
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.md
  },
  multiline: {
    minHeight: 112,
    textAlignVertical: "top"
  },
  inputError: {
    borderColor: colors.danger
  }
});

export function Input({
  containerStyle,
  error,
  helperText,
  label,
  multiline,
  style,
  ...props
}: InputProps) {
  return (
    <View style={[styles.container, containerStyle]}>
      {label ? (
        <Text tone="muted" variant="label" weight="semibold">
          {label}
        </Text>
      ) : null}
      <RNTextInput
        multiline={multiline}
        placeholderTextColor={colors.textMuted}
        style={[styles.input, multiline ? styles.multiline : null, error ? styles.inputError : null, style]}
        {...props}
      />
      {error ? (
        <Text tone="danger" variant="caption" weight="medium">
          {error}
        </Text>
      ) : helperText ? (
        <Text tone="muted" variant="caption">
          {helperText}
        </Text>
      ) : null}
    </View>
  );
}
