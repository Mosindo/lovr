import { createTokens } from "tamagui";
import { colors, fontSizes, radii, spacing } from "../shared/ui";

export const tokens = createTokens({
  color: {
    background: colors.background,
    text: colors.text,
    border: colors.border,
    primary: colors.primary,
    secondary: colors.secondary,
    muted: colors.surfaceMuted
  },
  space: {
    0: 0,
    1: spacing.xs,
    2: spacing.sm,
    3: spacing.md,
    4: spacing.lg,
    5: spacing.xl,
    6: spacing.xxl,
    true: spacing.lg
  },
  size: {
    0: 0,
    1: fontSizes.xs,
    2: fontSizes.sm,
    3: fontSizes.md,
    4: fontSizes.lg,
    5: fontSizes.xl,
    6: fontSizes.xxl,
    true: fontSizes.md
  },
  radius: {
    0: 0,
    1: radii.sm,
    2: radii.md,
    3: radii.lg,
    true: radii.md
  },
  zIndex: {
    0: 0,
    1: 100,
    2: 200,
    3: 300,
    4: 400,
    5: 500
  }
});
