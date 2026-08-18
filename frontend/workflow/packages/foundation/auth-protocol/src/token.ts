import { TOKEN_STORAGE_KEY, USER_INFO_STORAGE_KEY } from './constants';

export function getStoredToken(): string | null {
  return localStorage.getItem(TOKEN_STORAGE_KEY);
}

export function setStoredToken(token: string): void {
  localStorage.setItem(TOKEN_STORAGE_KEY, token);
}

export function removeStoredToken(): void {
  localStorage.removeItem(TOKEN_STORAGE_KEY);
}

export function getStoredUserInfo(): Record<string, unknown> | null {
  const raw = localStorage.getItem(USER_INFO_STORAGE_KEY);
  if (!raw) return null;
  try {
    return JSON.parse(raw);
  } catch {
    return null;
  }
}

export function setStoredUserInfo(userInfo: Record<string, unknown>): void {
  localStorage.setItem(USER_INFO_STORAGE_KEY, JSON.stringify(userInfo));
}

export function removeStoredUserInfo(): void {
  localStorage.removeItem(USER_INFO_STORAGE_KEY);
}

export function isInIframe(): boolean {
  try {
    return window.parent !== window;
  } catch {
    return true;
  }
}
