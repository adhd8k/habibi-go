// Utility function to extract error message from RTK Query errors
export function getErrorMessage(error: unknown): string {
  // Handle RTK Query error structure
  if (error && typeof error === 'object' && 'data' in error) {
    const data = (error as any).data
    if (data && typeof data === 'object') {
      // Check for common error message fields
      if ('error' in data && typeof data.error === 'string') {
        return data.error
      }
      if ('message' in data && typeof data.message === 'string') {
        return data.message
      }
    }
  }
  
  // Handle standard Error objects
  if (error instanceof Error) {
    return error.message
  }
  
  // Handle string errors
  if (typeof error === 'string') {
    return error
  }
  
  // Default error message
  return 'An unexpected error occurred'
}