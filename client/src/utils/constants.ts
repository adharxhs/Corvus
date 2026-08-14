const DEFAULT_API_BASE_URL = "http://localhost:8080";
const DEFAULT_WS_BASE_URL = "ws://localhost:8080";

function resolveBaseUrls(): { api: string; ws: string } {
  const serverUrl = import.meta.env.VITE_SERVER_URL as string | undefined;
  if (serverUrl) {
    const base = serverUrl.replace(/\/+$/, "");
    return { api: base, ws: base.replace(/^http/, "ws") };
  }
  const wsUrl = (import.meta.env.VITE_WS_URL as string | undefined) ?? DEFAULT_WS_BASE_URL;
  return { api: DEFAULT_API_BASE_URL, ws: wsUrl };
}

const urls = resolveBaseUrls();

export const API_BASE_URL = urls.api;
export const WS_BASE_URL = urls.ws;
export const AUTH_STORAGE_KEY = "corvus-auth";
export const THEME_STORAGE_KEY = "corvus-theme";
export const PROFILE_PICTURE_MAX_BYTES = 2 * 1024 * 1024;
export const CHATS_STORAGE_KEY = "corvus-chats";
export const CONTACTS_STORAGE_KEY = "corvus-contacts";
export const PENDING_MESSAGES_STORAGE_KEY = "corvus-pending-messages";
export const USERNAME_CACHE_KEY = "corvus-username-cache";
