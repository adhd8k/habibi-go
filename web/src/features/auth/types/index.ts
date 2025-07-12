export interface AuthCredentials {
  username: string
  password: string
}

export interface AuthState {
  isAuthenticated: boolean
  credentials: AuthCredentials | null
  isAuthRequired: boolean
  error: string | null
}