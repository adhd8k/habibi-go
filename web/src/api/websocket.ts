export class WebSocketClient {
  private ws: WebSocket | null = null
  private reconnectTimeout: NodeJS.Timeout | null = null
  private messageHandlers: Map<string, Set<(data: any) => void>> = new Map()
  private subscribedAgents: Set<number> = new Set()
  private reconnectAttempts = 0
  private maxReconnectAttempts = 10
  private isConnecting = false

  constructor(private url: string) {}

  connect() {
    if (this.isConnecting || (this.ws && this.ws.readyState === WebSocket.OPEN)) {
      return
    }
    
    this.isConnecting = true
    this.ws = new WebSocket(this.url)

    this.ws.onopen = () => {
      console.log('WebSocket connected')
      this.isConnecting = false
      this.reconnectAttempts = 0
      
      if (this.reconnectTimeout) {
        clearTimeout(this.reconnectTimeout)
        this.reconnectTimeout = null
      }
      
      // Resubscribe to all previously subscribed agents
      this.subscribedAgents.forEach(agentId => {
        this.ws!.send(JSON.stringify({ type: 'agent_logs_subscribe', agent_id: agentId }))
      })
    }

    this.ws.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data)
        console.log('WebSocket received message:', message)
        const handlers = this.messageHandlers.get(message.type)
        if (handlers && handlers.size > 0) {
          console.log(`Found ${handlers.size} handler(s) for message type:`, message.type)
          // Call all registered handlers
          handlers.forEach(handler => {
            try {
              handler(message)
            } catch (error) {
              console.error('Error in WebSocket handler:', error)
            }
          })
        } else {
          console.log('No handler found for message type:', message.type)
        }
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error)
      }
    }

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error)
    }

    this.ws.onclose = () => {
      console.log('WebSocket disconnected')
      this.isConnecting = false
      this.reconnect()
    }
  }

  private reconnect() {
    if (this.reconnectTimeout || this.reconnectAttempts >= this.maxReconnectAttempts) {
      if (this.reconnectAttempts >= this.maxReconnectAttempts) {
        console.error('Max reconnection attempts reached')
      }
      return
    }

    this.reconnectAttempts++
    const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts - 1), 30000) // Exponential backoff, max 30s
    
    this.reconnectTimeout = setTimeout(() => {
      console.log(`Attempting to reconnect WebSocket... (attempt ${this.reconnectAttempts})`)
      this.reconnectTimeout = null
      this.connect()
    }, delay)
  }

  subscribe(agentId: number) {
    this.subscribedAgents.add(agentId)
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type: 'agent_logs_subscribe', agent_id: agentId }))
    } else {
      // Try to connect if not connected
      this.connect()
    }
  }

  unsubscribe(agentId: number) {
    this.subscribedAgents.delete(agentId)
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type: 'agent_logs_unsubscribe', agent_id: agentId }))
    }
  }

  send(data: any) {
    console.log('WebSocket send called with data:', data)
    console.log('WebSocket state:', this.ws?.readyState)
    if (this.ws?.readyState === WebSocket.OPEN) {
      const message = JSON.stringify(data)
      console.log('Sending WebSocket message:', message)
      this.ws.send(message)
    } else {
      console.error('WebSocket is not connected, state:', this.ws?.readyState)
      // Try to connect if not connected
      this.connect()
    }
  }

  on(event: string, handler: (data: any) => void) {
    if (!this.messageHandlers.has(event)) {
      this.messageHandlers.set(event, new Set())
    }
    const handlers = this.messageHandlers.get(event)!
    handlers.add(handler)
    console.log(`Added handler for ${event}, total handlers: ${handlers.size}`)
  }

  off(event: string, handler?: (data: any) => void) {
    const handlers = this.messageHandlers.get(event)
    if (handlers) {
      if (handler) {
        handlers.delete(handler)
        console.log(`Removed specific handler for ${event}, remaining: ${handlers.size}`)
      } else {
        // Remove all handlers for this event if no specific handler provided
        this.messageHandlers.delete(event)
        console.log(`Removed all handlers for ${event}`)
      }
    }
  }

  disconnect() {
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout)
      this.reconnectTimeout = null
    }
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
    this.subscribedAgents.clear()
    this.reconnectAttempts = 0
    this.isConnecting = false
  }

  isConnected() {
    return this.ws?.readyState === WebSocket.OPEN
  }

  getReadyState() {
    return this.ws?.readyState
  }
}

// Get WebSocket URL with auth if needed
const getWebSocketUrl = () => {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const host = window.location.host
  const auth = localStorage.getItem('habibi_auth')
  
  if (auth) {
    const { username, password } = JSON.parse(auth)
    return `${protocol}//${username}:${password}@${host}/ws`
  }
  
  return `${protocol}//${host}/ws`
}

export const wsClient = new WebSocketClient(getWebSocketUrl())