import { Middleware } from '@reduxjs/toolkit'
import { RootState } from '../store'

export interface WebSocketMessage {
  type: string
  data?: any
  agent_id?: number
  session_id?: number
}

interface WebSocketConnectAction {
  type: 'websocket/connect'
  payload: {
    url: string
    onOpen?: () => void
    onError?: (error: Event) => void
  }
}

interface WebSocketDisconnectAction {
  type: 'websocket/disconnect'
}

interface WebSocketSendAction {
  type: 'websocket/send'
  payload: WebSocketMessage
}

interface WebSocketMessageReceivedAction {
  type: 'websocket/messageReceived'
  payload: WebSocketMessage
}

interface WebSocketErrorAction {
  type: 'websocket/error'
  payload: string
}

interface WebSocketConnectedAction {
  type: 'websocket/connected'
}

interface WebSocketDisconnectedAction {
  type: 'websocket/disconnected'
}

export type WebSocketAction =
  | WebSocketConnectAction
  | WebSocketDisconnectAction
  | WebSocketSendAction
  | WebSocketMessageReceivedAction
  | WebSocketErrorAction
  | WebSocketConnectedAction
  | WebSocketDisconnectedAction

export const websocketMiddleware: Middleware<{}, RootState> = (store) => {
  let socket: WebSocket | null = null
  let reconnectAttempts = 0
  const maxReconnectAttempts = 10
  let reconnectTimeout: NodeJS.Timeout | null = null
  let isConnecting = false

  const connect = (url: string, onOpen?: () => void, onError?: (error: Event) => void) => {
    if (isConnecting || (socket && socket.readyState === WebSocket.OPEN)) {
      return
    }

    isConnecting = true
    socket = new WebSocket(url)

    socket.onopen = () => {
      console.log('WebSocket connected')
      isConnecting = false
      reconnectAttempts = 0
      
      if (reconnectTimeout) {
        clearTimeout(reconnectTimeout)
        reconnectTimeout = null
      }
      
      store.dispatch({ type: 'websocket/connected' })
      onOpen?.()
    }

    socket.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data) as WebSocketMessage
        store.dispatch({
          type: 'websocket/messageReceived',
          payload: message,
        })
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error)
      }
    }

    socket.onerror = (error) => {
      console.error('WebSocket error:', error)
      store.dispatch({
        type: 'websocket/error',
        payload: 'WebSocket connection error',
      })
      onError?.(error)
    }

    socket.onclose = () => {
      console.log('WebSocket disconnected')
      isConnecting = false
      socket = null
      store.dispatch({ type: 'websocket/disconnected' })
      
      // Attempt to reconnect
      if (reconnectAttempts < maxReconnectAttempts) {
        reconnectAttempts++
        const delay = Math.min(1000 * Math.pow(2, reconnectAttempts - 1), 30000)
        
        reconnectTimeout = setTimeout(() => {
          console.log(`Attempting to reconnect WebSocket... (attempt ${reconnectAttempts})`)
          reconnectTimeout = null
          connect(url, onOpen, onError)
        }, delay)
      }
    }
  }

  return (next) => (action) => {
    const typedAction = action as WebSocketAction

    switch (typedAction.type) {
      case 'websocket/connect':
        const { url, onOpen, onError } = typedAction.payload
        connect(url, onOpen, onError)
        break

      case 'websocket/disconnect':
        if (reconnectTimeout) {
          clearTimeout(reconnectTimeout)
          reconnectTimeout = null
        }
        reconnectAttempts = maxReconnectAttempts // Prevent reconnection
        socket?.close()
        socket = null
        break

      case 'websocket/send':
        if (socket?.readyState === WebSocket.OPEN) {
          socket.send(JSON.stringify(typedAction.payload))
        } else {
          console.error('WebSocket is not connected')
          store.dispatch({
            type: 'websocket/error',
            payload: 'Cannot send message: WebSocket is not connected',
          })
        }
        break
    }

    return next(action)
  }
}

// Action creators
export const websocketConnect = (url: string, onOpen?: () => void, onError?: (error: Event) => void): WebSocketConnectAction => ({
  type: 'websocket/connect',
  payload: { url, onOpen, onError },
})

export const websocketDisconnect = (): WebSocketDisconnectAction => ({
  type: 'websocket/disconnect',
})

export const websocketSend = (message: WebSocketMessage): WebSocketSendAction => ({
  type: 'websocket/send',
  payload: message,
})