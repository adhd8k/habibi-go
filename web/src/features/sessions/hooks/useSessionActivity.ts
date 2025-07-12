import { useEffect } from 'react'
import { useAppDispatch, useAppSelector } from '../../../app/hooks'
import { useWebSocketSubscription } from '../../../shared/hooks/useWebSocketSubscription'
import { updateSessionActivity, selectCurrentSession } from '../slice/sessionsSlice'
import { sessionsApi } from '../api/sessionsApi'

export function useSessionActivity() {
  const dispatch = useAppDispatch()
  const currentSession = useAppSelector(selectCurrentSession)

  // Subscribe to session activity updates
  useWebSocketSubscription({
    messageType: 'session_activity_update',
    filter: (message) => {
      return message.data?.session_id === currentSession?.id
    },
    onMessage: (message) => {
      if (message.data?.session_id && message.data?.activity_status) {
        dispatch(updateSessionActivity({
          sessionId: message.data.session_id,
          status: message.data.activity_status,
        }))
      }
    },
  })

  // Subscribe to session updates
  useWebSocketSubscription({
    messageType: 'session_update',
    filter: (message) => {
      return message.data?.session?.id === currentSession?.id
    },
    onMessage: (message) => {
      if (message.data?.session) {
        // Invalidate the session query to refetch
        dispatch(
          sessionsApi.util.invalidateTags([
            { type: 'Session', id: message.data.session.id }
          ])
        )
      }
    },
  })

  // Mark session as viewed when switching to it
  useEffect(() => {
    if (currentSession?.activity_status === 'new') {
      // TODO: Add API endpoint to mark session as viewed
      dispatch(updateSessionActivity({
        sessionId: currentSession.id,
        status: 'viewed',
      }))
    }
  }, [currentSession, dispatch])
}