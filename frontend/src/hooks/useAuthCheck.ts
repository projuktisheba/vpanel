"use client"

import { useEffect, useState } from "react"
import { useAuth } from "../context/AuthContext"

/**
 * Custom hook to check authentication status
 * Useful for protecting routes and conditional rendering
 */
export const useAuthCheck = () => {
  const { isAuthenticated, loading } = useAuth()
  const [isReady, setIsReady] = useState(false)

  useEffect(() => {
    if (!loading) {
      setIsReady(true)
    }
  }, [loading])

  return {
    isAuthenticated,
    isReady,
    loading,
  }
}
