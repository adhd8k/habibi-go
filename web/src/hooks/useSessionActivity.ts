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

    // Handler for session updates
    const handleSessionUpdate = (message: any) => {
      if (message.type === 'session_update' && message.data?.session) {
        const session = message.data.session

        // Invalidate session queries to refresh the UI
        queryClient.invalidateQueries({ queryKey: ['sessions'] })
        queryClient.invalidateQueries({ queryKey: ['session', session.id] })
        queryClient.invalidateQueries({ queryKey: ['projects'] })
      }
    }

    // Handler for session creation
    const handleSessionCreated = (message: any) => {
      if (message.type === 'session_created' && message.data) {
        const { project_id } = message.data

        // Invalidate queries to show new session
        queryClient.invalidateQueries({ queryKey: ['sessions'] })
        queryClient.invalidateQueries({ queryKey: ['projects'] })
        if (project_id) {
          queryClient.invalidateQueries({ queryKey: ['project', project_id, 'sessions'] })
        }
      }
    }

    // Handler for session deletion
    const handleSessionDeleted = (message: any) => {
      if (message.type === 'session_deleted' && message.data) {
        const { project_id } = message.data

        // Invalidate queries to remove deleted session
        queryClient.invalidateQueries({ queryKey: ['sessions'] })
        queryClient.invalidateQueries({ queryKey: ['projects'] })
        if (project_id) {
          queryClient.invalidateQueries({ queryKey: ['project', project_id, 'sessions'] })
        }
      }
    }

    // Register handlers
    wsClient.on('session_activity_update', handleSessionActivityUpdate)
    wsClient.on('session_update', handleSessionUpdate)
    wsClient.on('session_created', handleSessionCreated)
    wsClient.on('session_deleted', handleSessionDeleted)

    return () => {
      wsClient.off('session_activity_update')
      wsClient.off('session_update')
      wsClient.off('session_created')
      wsClient.off('session_deleted')
    }
  }, [queryClient])
}