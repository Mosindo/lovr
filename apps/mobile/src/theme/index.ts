import { colors } from "../shared/ui";
import { tokens } from "./tokens";

export { tokens };

export const themes = {
  light: {
    background: colors.background,
    backgroundHover: colors.surfaceMuted,
    backgroundPress: colors.border,
    color: colors.text,
    colorHover: colors.text,
    colorPress: colors.text,
    borderColor: colors.border,
    borderColorHover: colors.border,
    borderColorPress: colors.border,
    primary: colors.primary,
    primaryHover: colors.primary,
    primaryPress: colors.primary,
    muted: colors.surfaceMuted
  }
} as const;
