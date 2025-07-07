export class WebSocketClient {
  private ws: WebSocket | null = null
  private reconnectTimeout: NodeJS.Timeout | null = null
  private messageHandlers: Map<string, (data: any) => void> = new Map()

  constructor(private url: string) {}

  connect() {
    this.ws = new WebSocket(this.url)

    this.ws.onopen = () => {
      console.log('WebSocket connected')
      if (this.reconnectTimeout) {
        clearTimeout(this.reconnectTimeout)
        this.reconnectTimeout = null
      }
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
      this.reconnect()
    }
  }

  private reconnect() {
    if (this.reconnectTimeout) return

    this.reconnectTimeout = setTimeout(() => {
      console.log('Attempting to reconnect WebSocket...')
      this.connect()
    }, 5000)
  }

  subscribe(agentId: number) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type: 'agent_logs_subscribe', agent_id: agentId }))
    }
  }

  unsubscribe(agentId: number) {
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
  }
}

export const wsClient = new WebSocketClient('ws://localhost:8080/ws')