import { createApi, fetchBaseQuery, retry } from '@reduxjs/toolkit/query/react'
import type { BaseQueryFn, FetchArgs, FetchBaseQueryError } from '@reduxjs/toolkit/query'
import { RootState } from '../../app/store'
import { getAuthHeader } from '../../features/auth/slice/authSlice'
import { setAuthRequired, clearCredentials } from '../../features/auth/slice/authSlice'

// Create a base query with auth header
const baseQuery = fetchBaseQuery({
  baseUrl: '/api/v1',
  prepareHeaders: (headers, { getState }) => {
    const state = getState() as RootState
    const authHeader = getAuthHeader(state.auth.credentials)
    
    if (authHeader) {
      headers.set('Authorization', authHeader)
    }
    
    return headers
  },
})

// Add automatic retry with exponential backoff
const baseQueryWithRetry = retry(baseQuery, { maxRetries: 3 })

// Enhanced base query with auth error handling
const baseQueryWithAuth: BaseQueryFn<
  string | FetchArgs,
  unknown,
  FetchBaseQueryError
> = async (args, api, extraOptions) => {
  const result = await baseQueryWithRetry(args, api, extraOptions)
  
  // Handle 401 Unauthorized
  if (result.error && result.error.status === 401) {
    // Clear credentials and mark auth as required
    api.dispatch(clearCredentials())
    api.dispatch(setAuthRequired(true))
  }
  
  return result
}

// Create the base API slice
export const baseApi = createApi({
  reducerPath: 'api',
  baseQuery: baseQueryWithAuth,
  tagTypes: ['Project', 'Session', 'Agent', 'Chat'],
  endpoints: () => ({}),
})

// Helper type for API responses
export interface ApiResponse<T> {
  success: boolean
  data?: T
  error?: string
}

// Helper to unwrap API responses
export const unwrapApiResponse = <T>(response: ApiResponse<T>): T => {
  if (!response.success || !response.data) {
    throw new Error(response.error || 'Unknown error')
  }
  return response.data
}