import { create } from 'zustand';
import { AuthTokens, UserRole } from '@/types';
import {
  clearStoredTokens,
  decodeJWT,
  getStoredTokens,
  refreshAccessToken,
  setStoredTokens,
  setOnSessionExpired,
} from '@/lib/api/client';
import { authApi } from '@/lib/api/services';

interface AuthState {
  accessToken: string | null;
  customerID: string | null;
  role: UserRole;
  isAuthenticated: boolean;
  isAdmin: boolean;
  isLoading: boolean;
  initialize: () => Promise<void>;
  login: (credentials: { email: string; password: string }) => Promise<void>;
  register: (payload: { email: string; password: string; name: string }) => Promise<void>;
  verifyOTPAndRegister: (payload: { email: string; otp: string }) => Promise<void>;
  setOAuthTokens: (tokens: AuthTokens) => void;
  logout: () => void;
}

export const useAuthStore = create<AuthState>((set) => {
  // Synchronize token expiry across the application
  setOnSessionExpired(() => {
    set({
      accessToken: null,
      customerID: null,
      role: 'customer',
      isAuthenticated: false,
      isAdmin: false,
    });
  });

  return {
    accessToken: null,
    customerID: null,
    role: 'customer',
    isAuthenticated: false,
    isAdmin: false,
    isLoading: true,

    initialize: async () => {
      const { accessToken, refreshToken, customerID } = getStoredTokens();
      
      // 1. If access token is still fresh
      if (accessToken && customerID) {
        const decoded = decodeJWT(accessToken);
        if (decoded && decoded.exp * 1000 > Date.now()) {
          const role = decoded.role || 'customer';
          set({
            accessToken,
            customerID,
            role,
            isAuthenticated: true,
            isAdmin: role === 'admin',
            isLoading: false,
          });
          return;
        }
      }

      // 2. If access token expired but 10-day refresh token exists, rotate & recover
      if (refreshToken) {
        try {
          const newAccessToken = await refreshAccessToken();
          if (newAccessToken) {
            const decoded = decodeJWT(newAccessToken);
            const role = decoded?.role || 'customer';
            const freshCustID = decoded?.customer_id || customerID || '';
            set({
              accessToken: newAccessToken,
              customerID: freshCustID,
              role,
              isAuthenticated: true,
              isAdmin: role === 'admin',
              isLoading: false,
            });
            return;
          }
        } catch {
          // Token refresh failed, clean up below
        }
      }

      // 3. Fallback when both access and refresh tokens are exhausted
      clearStoredTokens();
      set({
        accessToken: null,
        customerID: null,
        role: 'customer',
        isAuthenticated: false,
        isAdmin: false,
        isLoading: false,
      });
    },

  login: async (credentials) => {
    const tokens: AuthTokens = await authApi.login(credentials);
    setStoredTokens(tokens);
    const decoded = decodeJWT(tokens.accessToken);
    const role = decoded?.role || 'customer';
    set({
      accessToken: tokens.accessToken,
      customerID: tokens.customerID,
      role,
      isAuthenticated: true,
      isAdmin: role === 'admin',
    });
  },

  register: async (payload) => {
    const tokens: AuthTokens = await authApi.register(payload);
    setStoredTokens(tokens);
    const decoded = decodeJWT(tokens.accessToken);
    const role = decoded?.role || 'customer';
    set({
      accessToken: tokens.accessToken,
      customerID: tokens.customerID,
      role,
      isAuthenticated: true,
      isAdmin: role === 'admin',
    });
  },

  verifyOTPAndRegister: async (payload) => {
    const tokens: AuthTokens = await authApi.verifyRegistrationOTP(payload);
    setStoredTokens(tokens);
    const decoded = decodeJWT(tokens.accessToken);
    const role = decoded?.role || 'customer';
    set({
      accessToken: tokens.accessToken,
      customerID: tokens.customerID,
      role,
      isAuthenticated: true,
      isAdmin: role === 'admin',
    });
  },

  setOAuthTokens: (tokens: AuthTokens) => {
    setStoredTokens(tokens);
    const decoded = decodeJWT(tokens.accessToken);
    const role = decoded?.role || 'customer';
    set({
      accessToken: tokens.accessToken,
      customerID: tokens.customerID,
      role,
      isAuthenticated: true,
      isAdmin: role === 'admin',
      isLoading: false,
    });
  },

  logout: () => {
    clearStoredTokens();
    set({
      accessToken: null,
      customerID: null,
      role: 'customer',
      isAuthenticated: false,
      isAdmin: false,
    });
  },
  };
});

