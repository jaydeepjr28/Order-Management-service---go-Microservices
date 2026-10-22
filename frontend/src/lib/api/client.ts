import { AuthTokens, DecodedToken, UserRole } from '@/types';

const API_BASE = import.meta.env.VITE_API_BASE_URL || '';

const TOKEN_KEY = 'stratum_access_token';
const REFRESH_TOKEN_KEY = 'stratum_refresh_token';
const CUSTOMER_ID_KEY = 'stratum_customer_id';

export function getStoredTokens(): { accessToken: string | null; refreshToken: string | null; customerID: string | null } {
  return {
    accessToken: localStorage.getItem(TOKEN_KEY) || localStorage.getItem('nexus_access_token'),
    refreshToken: localStorage.getItem(REFRESH_TOKEN_KEY) || localStorage.getItem('nexus_refresh_token'),
    customerID: localStorage.getItem(CUSTOMER_ID_KEY) || localStorage.getItem('nexus_customer_id'),
  };
}

export function setStoredTokens(tokens: AuthTokens) {
  localStorage.setItem(TOKEN_KEY, tokens.accessToken);
  localStorage.setItem(REFRESH_TOKEN_KEY, tokens.refreshToken);
  localStorage.setItem(CUSTOMER_ID_KEY, tokens.customerID);
}

export function clearStoredTokens() {
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(REFRESH_TOKEN_KEY);
  localStorage.removeItem(CUSTOMER_ID_KEY);
  localStorage.removeItem('nexus_access_token');
  localStorage.removeItem('nexus_refresh_token');
  localStorage.removeItem('nexus_customer_id');
}

export function decodeJWT(token: string): DecodedToken | null {
  try {
    const parts = token.split('.');
    if (parts.length !== 3) return null;
    const base64Url = parts[1];
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
    const jsonPayload = decodeURIComponent(
      atob(base64)
        .split('')
        .map((c) => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
        .join('')
    );
    return JSON.parse(jsonPayload) as DecodedToken;
  } catch {
    return null;
  }
}

export function getCurrentUserRole(): UserRole {
  const { accessToken } = getStoredTokens();
  if (!accessToken) return 'customer';
  const decoded = decodeJWT(accessToken);
  return decoded?.role || 'customer';
}

// Session Expiry Notification Handler
let sessionExpiredHandler: (() => void) | null = null;

export function setOnSessionExpired(handler: () => void) {
  sessionExpiredHandler = handler;
}

function notifySessionExpired() {
  clearStoredTokens();
  if (sessionExpiredHandler) {
    sessionExpiredHandler();
  }
}

// Anti-refresh storm mutex
let isRefreshing = false;
let refreshSubscribers: Array<(newToken: string | null) => void> = [];

function subscribeTokenRefresh(cb: (newToken: string | null) => void) {
  refreshSubscribers.push(cb);
}

function onRefreshed(newToken: string | null) {
  refreshSubscribers.forEach((cb) => cb(newToken));
  refreshSubscribers = [];
}

export async function refreshAccessToken(): Promise<string | null> {
  const { refreshToken, customerID } = getStoredTokens();
  if (!refreshToken) {
    notifySessionExpired();
    return null;
  }

  try {
    const response = await fetch(`${API_BASE}/api/users/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refreshToken }),
    });

    if (!response.ok) {
      notifySessionExpired();
      return null;
    }

    const data = await response.json();
    const newTokens: AuthTokens = {
      accessToken: data.AccessToken || data.accessToken,
      refreshToken: data.RefreshToken || data.refreshToken || refreshToken,
      customerID: data.CustomerID || data.customerID || customerID || '',
    };

    setStoredTokens(newTokens);
    return newTokens.accessToken;
  } catch {
    notifySessionExpired();
    return null;
  }
}

export interface ApiRequestOptions extends RequestInit {
  skipAuth?: boolean;
  skipRefreshRetry?: boolean;
}

export async function apiClient<T>(endpoint: string, options: ApiRequestOptions = {}): Promise<T> {
  const url = `${API_BASE}${endpoint.startsWith('/') ? endpoint : `/${endpoint}`}`;
  const headers = new Headers(options.headers || {});

  if (!headers.has('Content-Type') && !(options.body instanceof FormData)) {
    headers.set('Content-Type', 'application/json');
  }

  const { accessToken } = getStoredTokens();
  if (!options.skipAuth && accessToken && !headers.has('Authorization')) {
    headers.set('Authorization', `Bearer ${accessToken}`);
  }

  const config: RequestInit = {
    ...options,
    headers,
  };

  let response: Response;
  try {
    response = await fetch(url, config);
  } catch (err: any) {
    if (err?.name === 'TypeError' && String(err?.message).toLowerCase().includes('fetch')) {
      throw new Error('Network error: Unable to connect to Stratum Gateway service. Please verify server status.');
    }
    throw err;
  }

  // Handle Token Expiry (401 Unauthorized)
  if (response.status === 401 && !options.skipAuth && !options.skipRefreshRetry) {
    if (!isRefreshing) {
      isRefreshing = true;
      const newToken = await refreshAccessToken();
      isRefreshing = false;
      onRefreshed(newToken);

      if (newToken) {
        // Re-attempt original request with new token
        headers.set('Authorization', `Bearer ${newToken}`);
        return apiClient<T>(endpoint, {
          ...options,
          headers,
          skipRefreshRetry: true,
        });
      }
      throw new Error('Session expired. Please sign in again.');
    } else {
      // Another request is currently refreshing the token; wait for it
      return new Promise<T>((resolve, reject) => {
        subscribeTokenRefresh(async (newToken) => {
          if (!newToken) {
            reject(new Error('Session expired. Please sign in again.'));
            return;
          }
          headers.set('Authorization', `Bearer ${newToken}`);
          try {
            const retryRes = await apiClient<T>(endpoint, {
              ...options,
              headers,
              skipRefreshRetry: true,
            });
            resolve(retryRes);
          } catch (err) {
            reject(err);
          }
        });
      });
    }
  }

  if (!response.ok) {
    let errorMessage = `Request failed (${response.status})`;
    try {
      const errData = await response.json();
      errorMessage = errData.error || errData.message || errorMessage;
    } catch {
      // Keep default error message
    }
    throw new Error(errorMessage);
  }

  if (response.status === 204) {
    return {} as T;
  }

  const text = await response.text();
  if (!text) {
    return {} as T;
  }

  try {
    return JSON.parse(text) as T;
  } catch {
    return text as unknown as T;
  }
}
