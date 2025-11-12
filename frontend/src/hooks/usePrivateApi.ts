"use client"

import { useCallback } from "react"
import { useAuth } from "../context/AuthContext"
import { authService } from "../services/auth.service"

/**
 * Custom hook for making authenticated API calls
 * Automatically handles token refresh and error handling
 */
export const usePrivateApi = () => {
  const { refreshAccessToken } = useAuth()

  const request = useCallback(
    async (url: string, options: RequestInit = {}): Promise<Response> => {
      try {
        const response = await authService.fetchWithAuth(url, options)
        return response
      } catch (error) {
        // Try refreshing token if there's an auth error
        const refreshed = await refreshAccessToken()
        if (refreshed) {
          return authService.fetchWithAuth(url, options)
        }
        throw error
      }
    },
    [refreshAccessToken],
  )

  return { request }
}
