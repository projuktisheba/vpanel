export interface AuthTokens {
  accessToken: string
  refreshToken: string
}

export interface DecodedToken {
  sub: string
  email: string
  iat: number
  exp: number
}

export interface LoginCredentials {
  email: string
  password: string
}

export interface AuthResponse {
  user: {
    id: string
    email: string
    name: string
  }
  accessToken: string
  refreshToken: string
}

export interface AuthContextType {
  user: AuthResponse["user"] | null
  loading: boolean
  isAuthenticated: boolean
  error: string | null
  login: (credentials: LoginCredentials) => Promise<void>
  logout: () => Promise<void>
  refreshAccessToken: () => Promise<boolean>
}
