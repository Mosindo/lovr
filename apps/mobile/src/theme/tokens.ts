import { createTokens } from "tamagui";

export const tokens = createTokens({
  color: {
    background: "#f9fafb",
    text: "#111827",
    border: "#d1d5db",
    primary: "#2563eb",
    muted: "#e5e7eb"
  },
  space: {
    0: 0,
    1: 4,
    2: 8,
    3: 12,
    4: 16,
    5: 20,
    6: 24,
    true: 16
  },
  size: {
    0: 0,
    1: 12,
    2: 14,
    3: 16,
    4: 18,
    5: 20,
    6: 24,
    true: 16
  },
  radius: {
    0: 0,
    1: 4,
    2: 8,
    3: 12,
    true: 8
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
