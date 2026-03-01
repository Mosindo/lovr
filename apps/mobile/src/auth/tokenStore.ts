import * as SecureStore from "expo-secure-store";

const TOKEN_KEY = "lovr.auth.token";
let memoryToken: string | null = null;

export async function saveToken(token: string): Promise<void> {
  memoryToken = token;
  try {
    await SecureStore.setItemAsync(TOKEN_KEY, token);
  } catch {
    // Fallback keeps session in memory when secure storage is unavailable.
  }
}

export async function getToken(): Promise<string | null> {
  try {
    const token = await SecureStore.getItemAsync(TOKEN_KEY);
    if (token) {
      memoryToken = token;
      return token;
    }
  } catch {
    // Ignore and fallback to memory.
  }
  return memoryToken;
}

export async function clearToken(): Promise<void> {
  memoryToken = null;
  try {
    await SecureStore.deleteItemAsync(TOKEN_KEY);
  } catch {
    // Ignore cleanup errors on unsupported platforms.
  }
}
