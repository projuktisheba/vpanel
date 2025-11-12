import type React from "react"
import { Navigate } from "react-router-dom"
import { useAuthCheck } from "../hooks/useAuthCheck"

interface ProtectedRouteProps {
  children: React.ReactNode
  fallback?: React.ReactNode
}

/**
 * Protected Route Component
 * Redirects to login if user is not authenticated
 */
export const ProtectedRoute: React.FC<ProtectedRouteProps> = ({ children, fallback = <div>Loading...</div> }) => {
  const { isAuthenticated, isReady } = useAuthCheck()

  if (!isReady) {
    return fallback
  }

  if (!isAuthenticated) {
    return <Navigate to="/signin" replace />
  }

  return <>{children}</>
}
