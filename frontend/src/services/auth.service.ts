import axios, { AxiosInstance } from "axios";
import { LoginCredentials, AuthResponse } from "../interfaces/auth.interface";
import { tokenUtils } from "../utils/tokenUtils";
import { API_BASE_URL } from "../config/apiConfig";

// Create axios instance
const api: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    "Content-Type": "application/json",
  },
  withCredentials: true, // Optional: if backend uses cookies
});

export const authService = {
  // Login user
  login: async (credentials: LoginCredentials): Promise<AuthResponse> => {
    try {
      const response = await api.post<AuthResponse>("/auth/signin", {
        username: credentials.email,
        password: credentials.password,
      });

      const data = response.data;
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
        await api.post(
          "/auth/signout",
          { refreshToken },
          {
            headers: {
              Authorization: `Bearer ${tokenUtils.getAccessToken()}`,
            },
          }
        );
      }
    } catch (error) {
      // Ignore logout error
    } finally {
      tokenUtils.clearTokens();
    }
  },

  // Refresh access token
  refreshAccessToken: async (): Promise<string | null> => {
    try {
      const refreshToken = tokenUtils.getRefreshToken();
      if (!refreshToken) throw new Error("No refresh token available");

      const response = await api.post("/auth/refresh", { refreshToken });
      const data = response.data;

      tokenUtils.setTokens(data.accessToken, data.refreshToken);
      return data.accessToken;
    } catch (error: any) {
      tokenUtils.clearTokens();
      throw new Error(error.response?.data?.message || "Token refresh failed");
    }
  },

  // Generic request wrapper with token handling
  fetchWithAuth: async (endpoint: string, options: any = {}): Promise<any> => {
    let accessToken = tokenUtils.getAccessToken();

    // Refresh if token is expired
    if (accessToken && tokenUtils.isTokenExpired(accessToken)) {
      try {
        const newToken = await authService.refreshAccessToken();
        accessToken = newToken;
      } catch {
        tokenUtils.clearTokens();
        throw new Error("Session expired. Please login again.");
      }
    }

    try {
      const response = await api.request({
        url: endpoint,
        method: options.method || "GET",
        data: options.body || undefined,
        headers: {
          Authorization: `Bearer ${accessToken}`,
          ...(options.headers || {}),
        },
      });

      return response;
    } catch (error: any) {
      // If 401, try one refresh attempt
      if (error.response?.status === 401) {
        try {
          const newToken = await authService.refreshAccessToken();
          const retryResponse = await api.request({
            url: endpoint,
            method: options.method || "GET",
            data: options.body || undefined,
            headers: {
              Authorization: `Bearer ${newToken}`,
              ...(options.headers || {}),
            },
          });
          return retryResponse;
        } catch {
          tokenUtils.clearTokens();
          throw new Error("Session expired. Please login again.");
        }
      }
      throw error;
    }
  },
};
