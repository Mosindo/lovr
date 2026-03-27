import { colors } from "../shared/ui";
import { tokens } from "./tokens";

export { tokens };

export const themes = {
  light: {
    background: colors.background,
    backgroundStrong: colors.backgroundElevated,
    backgroundHover: colors.surfaceMuted,
    backgroundPress: colors.surfaceSubtle,
    color: colors.text,
    colorMuted: colors.textMuted,
    colorHover: colors.text,
    colorPress: colors.text,
    borderColor: colors.border,
    borderColorHover: colors.borderStrong,
    borderColorPress: colors.borderStrong,
    primary: colors.primary,
    primaryHover: colors.primary,
    primaryPress: colors.primary,
    success: colors.success,
    warning: colors.warning,
    danger: colors.danger,
    muted: colors.surfaceMuted,
    accent: colors.surfaceAccent
  }
} as const;
