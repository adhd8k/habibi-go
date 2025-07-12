import { useEffect, useCallback } from 'react'
import { useAppDispatch } from '../../app/hooks'
import { WebSocketMessage } from '../../app/middleware/websocket'

interface UseWebSocketSubscriptionOptions {
  messageType?: string | string[]
  filter?: (message: WebSocketMessage) => boolean
  onMessage: (message: WebSocketMessage) => void
}

export function useWebSocketSubscription({
  messageType,
  filter,
  onMessage,
}: UseWebSocketSubscriptionOptions) {
  const dispatch = useAppDispatch()

  const handleMessage = useCallback((action: any) => {
    if (action.type !== 'websocket/messageReceived') return
    
    const message = action.payload as WebSocketMessage
    
    // Check message type if specified
    if (messageType) {
      const types = Array.isArray(messageType) ? messageType : [messageType]
      if (!types.includes(message.type)) return
    }
    
    // Apply custom filter if provided
    if (filter && !filter(message)) return
    
    // Call the handler
    onMessage(message)
  }, [messageType, filter, onMessage])

  useEffect(() => {
    // Subscribe to store changes
    const unsubscribe = (dispatch as any)(handleMessage)
    
    return () => {
      if (typeof unsubscribe === 'function') {
        unsubscribe()
      }
    }
  }, [dispatch, handleMessage])
}