export class WebSocketClient {
  private ws: WebSocket | null = null
  private reconnectTimeout: NodeJS.Timeout | null = null
  private messageHandlers: Map<string, (data: any) => void> = new Map()
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
        const handler = this.messageHandlers.get(message.type)
        if (handler) {
          // Pass the full message to handlers, not just data
          handler(message)
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

  on(event: string, handler: (data: any) => void) {
    this.messageHandlers.set(event, handler)
  }

  off(event: string) {
    this.messageHandlers.delete(event)
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
}

export const wsClient = new WebSocketClient('ws://localhost:8080/ws')