export const colors = {
  primary: "#2563eb",
  secondary: "#0f766e",
  background: "#f8fafc",
  surface: "#ffffff",
  surfaceMuted: "#f1f5f9",
  border: "#dbe4ee",
  text: "#0f172a",
  textMuted: "#64748b",
  danger: "#b91c1c",
  success: "#047857",
  inverse: "#ffffff"
} as const;

export const spacing = {
  xs: 4,
  sm: 8,
  md: 12,
  lg: 16,
  xl: 20,
  xxl: 24,
  xxxl: 32
} as const;

export const radii = {
  sm: 8,
  md: 12,
  lg: 16,
  xl: 20,
  pill: 999
} as const;

export const fontSizes = {
  xs: 12,
  sm: 14,
  md: 16,
  lg: 18,
  xl: 24,
  xxl: 30
} as const;

export const shadows = {
  card: {
    shadowColor: "#0f172a",
    shadowOpacity: 0.06,
    shadowRadius: 10,
    shadowOffset: { width: 0, height: 4 },
    elevation: 2
  }
} as const;

export const ui = {
  colors,
  spacing,
  radii,
  fontSizes,
  shadows
} as const;

export type ColorToken = keyof typeof colors;
