// src/services/authService.ts
import axiosInstance from "../hooks/AxiosInstance";
import { LoginCredentials, AuthResponse } from "../interfaces/auth.interface";
import { tokenUtils } from "../utils/tokenUtils";

export const authService = {
  // Login user
  login: async (credentials: LoginCredentials): Promise<AuthResponse> => {
    try {
      const response = await axiosInstance.post<AuthResponse>("/auth/signin", {
        username: credentials.email,
        password: credentials.password,
      });

      const data = response.data;

      // Store tokens and user info
      tokenUtils.setTokens(data.accessToken, data.refreshToken);
      tokenUtils.setUser(data.user);

      return data;
    } catch (error: any) {
      throw new Error(error.response?.data?.message || "Login failed");
    }
  },

  // Logout user
  logout: async (): Promise<void> => {
    try {
      const refreshToken = tokenUtils.getRefreshToken();
      if (refreshToken) {
        await axiosInstance.post("/auth/signout", { refreshToken });
      }
    } catch {
      // Ignore logout errors
    } finally {
      tokenUtils.clearTokens();
    }
  },

  // Refresh access token
  refreshAccessToken: async (): Promise<string | null> => {
    try {
      const refreshToken = tokenUtils.getRefreshToken();
      if (!refreshToken) throw new Error("No refresh token available");

      const response = await axiosInstance.post<{ accessToken: string; refreshToken: string }>(
        "/auth/refresh",
        { refreshToken }
      );

      const data = response.data;
      tokenUtils.setTokens(data.accessToken, data.refreshToken);
      return data.accessToken;
    } catch (error: any) {
      tokenUtils.clearTokens();
      throw new Error(error.response?.data?.message || "Token refresh failed");
    }
  },

  // Generic request wrapper with automatic token refresh
  fetchWithAuth: async (endpoint: string, options: { method?: string; body?: any; headers?: any } = {}) => {
    try {
      const response = await axiosInstance.request({
        url: endpoint,
        method: options.method || "GET",
        data: options.body,
        headers: options.headers || {},
      });
      return response.data;
    } catch (error: any) {
      // If 401, try refreshing token once
      if (error.response?.status === 401) {
        try {
          await authService.refreshAccessToken();
          // Retry original request after refresh
          const retryResponse = await axiosInstance.request({
            url: endpoint,
            method: options.method || "GET",
            data: options.body,
            headers: options.headers || {},
          });
          return retryResponse.data;
        } catch {
          tokenUtils.clearTokens();
          throw new Error("Session expired. Please login again.");
        }
      }
      throw error;
    }
  },
};
