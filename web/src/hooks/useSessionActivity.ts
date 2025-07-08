import { useEffect } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { wsClient } from '../api/websocket'

export function useSessionActivity() {
  const queryClient = useQueryClient()

  useEffect(() => {
    // Handler for session activity updates
    const handleSessionActivityUpdate = (message: any) => {
      if (message.type === 'session_activity_update' && message.data) {
        const { session_id, activity_status } = message.data
        
        // Invalidate session queries to refresh the UI
        queryClient.invalidateQueries({ queryKey: ['sessions'] })
        queryClient.invalidateQueries({ queryKey: ['session', session_id] })
        
        // Update session data in cache if available
        queryClient.setQueriesData(
          { queryKey: ['sessions'] },
          (oldData: any) => {
            if (!oldData?.data) return oldData
            
            return {
              ...oldData,
              data: oldData.data.map((session: any) =>
                session.id === session_id
                  ? { ...session, activity_status }
                  : session
              )
            }
          }
        )
      }
    }

    // Register handler
    wsClient.on('session_activity_update', handleSessionActivityUpdate)

    return () => {
      wsClient.off('session_activity_update')
    }
  }, [queryClient])
}