"use client"

import type React from "react"
import { createContext, useContext, useState, useEffect, type ReactNode } from "react"
import { AuthContextType, AuthResponse, LoginCredentials } from "../interfaces/auth.interface"
import { authService } from "../services/auth.service"
import { tokenUtils } from "../utils/tokenUtils"

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export const AuthProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<AuthResponse["user"] | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // Initialize auth state on mount
  useEffect(() => {
    const initializeAuth = () => {
      const storedUser = tokenUtils.getUser()
      const accessToken = tokenUtils.getAccessToken()

      if (storedUser && accessToken && !tokenUtils.isTokenExpired(accessToken)) {
        setUser(storedUser)
      } else {
        tokenUtils.clearTokens()
      }
      setLoading(false)
    }

    initializeAuth()
  }, [])

  const login = async (credentials: LoginCredentials) => {
    try {
      setError(null)
      const response = await authService.login(credentials)
      setUser(response.user)
    } catch (err) {
      const message = err instanceof Error ? err.message : "Login failed"
      setError(message)
      throw err
    }
  }

  const logout = async () => {
    try {
      setError(null)
      await authService.logout()
      setUser(null)
    } catch (err) {
      const message = err instanceof Error ? err.message : "Logout failed"
      setError(message)
      throw err
    }
  }

  const refreshAccessToken = async (): Promise<boolean> => {
    try {
      await authService.refreshAccessToken()
      return true
    } catch (err) {
      setUser(null)
      return false
    }
  }

  const value: AuthContextType = {
    user,
    loading,
    isAuthenticated: !!user,
    error,
    login,
    logout,
    refreshAccessToken,
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export const useAuth = () => {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error("useAuth must be used within an AuthProvider")
  }
  return context
}
