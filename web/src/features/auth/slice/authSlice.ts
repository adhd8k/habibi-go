import { createSlice, PayloadAction } from '@reduxjs/toolkit'
import { AuthState, AuthCredentials } from '../types'

const STORAGE_KEY = 'habibi_auth'

// Load credentials from localStorage
const loadStoredCredentials = (): AuthCredentials | null => {
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    return stored ? JSON.parse(stored) : null
  } catch {
    return null
  }
}

const initialState: AuthState = {
  isAuthenticated: !!loadStoredCredentials(),
  credentials: loadStoredCredentials(),
  isAuthRequired: false,
  error: null,
}

export const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    setCredentials: (state, action: PayloadAction<AuthCredentials>) => {
      state.credentials = action.payload
      state.isAuthenticated = true
      state.error = null
      state.isAuthRequired = false
      
      // Store in localStorage
      localStorage.setItem(STORAGE_KEY, JSON.stringify(action.payload))
    },
    
    clearCredentials: (state) => {
      state.credentials = null
      state.isAuthenticated = false
      state.error = null
      
      // Remove from localStorage
      localStorage.removeItem(STORAGE_KEY)
    },
    
    setAuthRequired: (state, action: PayloadAction<boolean>) => {
      state.isAuthRequired = action.payload
    },
    
    setAuthError: (state, action: PayloadAction<string>) => {
      state.error = action.payload
    },
  },
})

export const { setCredentials, clearCredentials, setAuthRequired, setAuthError } = authSlice.actions

export default authSlice.reducer

// Selectors
export const selectIsAuthenticated = (state: { auth: AuthState }) => state.auth.isAuthenticated
export const selectCredentials = (state: { auth: AuthState }) => state.auth.credentials
export const selectIsAuthRequired = (state: { auth: AuthState }) => state.auth.isAuthRequired
export const selectAuthError = (state: { auth: AuthState }) => state.auth.error

// Helper to get auth header
export const getAuthHeader = (credentials: AuthCredentials | null): string | null => {
  if (!credentials) return null
  return 'Basic ' + btoa(`${credentials.username}:${credentials.password}`)
}