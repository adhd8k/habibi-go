import { useEffect, useRef, useState } from 'react'
import { Terminal as XTerm } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'
import '@xterm/xterm/css/xterm.css'
import { useAppStore } from '../store'

export function Terminal() {
  const terminalRef = useRef<HTMLDivElement>(null)
  const xtermRef = useRef<XTerm | null>(null)
  const fitAddonRef = useRef<FitAddon | null>(null)
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectFuncRef = useRef<(() => void) | null>(null)
  const [connectionStatus, setConnectionStatus] = useState<'connecting' | 'connected' | 'disconnected' | 'error'>('connecting')
  const { currentSession } = useAppStore()

  useEffect(() => {
    if (!terminalRef.current || !currentSession) return

    // Initialize xterm.js
    const terminal = new XTerm({
      cursorBlink: true,
      fontSize: 14,
      fontFamily: 'Menlo, Monaco, "Courier New", monospace',
      theme: {
        background: '#1e1e1e',
        foreground: '#d4d4d4',
        cursor: '#ffffff',
        cursorAccent: '#000000',
        selectionBackground: '#264f78',
        black: '#000000',
        red: '#cd3131',
        green: '#0dbc79',
        yellow: '#e5e510',
        blue: '#2472c8',
        magenta: '#bc3fbc',
        cyan: '#11a8cd',
        white: '#e5e5e5',
        brightBlack: '#666666',
        brightRed: '#f14c4c',
        brightGreen: '#23d18b',
        brightYellow: '#f5f543',
        brightBlue: '#3b8eea',
        brightMagenta: '#d670d6',
        brightCyan: '#29b8db',
        brightWhite: '#ffffff'
      }
    })

    const fitAddon = new FitAddon()
    const webLinksAddon = new WebLinksAddon()
    
    terminal.loadAddon(fitAddon)
    terminal.loadAddon(webLinksAddon)
    
    terminal.open(terminalRef.current)
    fitAddon.fit()

    xtermRef.current = terminal
    fitAddonRef.current = fitAddon

    // Connect to backend terminal websocket
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${window.location.host}/api/terminal/${currentSession.id}`
    const ws = new WebSocket(wsUrl)
    wsRef.current = ws

    let reconnectAttempt = 0
    const maxReconnectAttempts = 5
    
    const connectWebSocket = () => {
      setConnectionStatus('connecting')
      const newWs = new WebSocket(wsUrl)
      wsRef.current = newWs
      setupWebSocket(newWs)
      return newWs
    }
    
    const manualReconnect = () => {
      reconnectAttempt = 0 // Reset attempts for manual reconnect
      if (wsRef.current) {
        wsRef.current.close(1000)
      }
      if (xtermRef.current) {
        xtermRef.current.write('\r\n\x1b[36mReconnecting...\x1b[0m\r\n')
      }
      connectWebSocket()
    }
    
    reconnectFuncRef.current = manualReconnect
    
    const setupWebSocket = (wsInstance: WebSocket) => {
      wsInstance.onopen = () => {
        console.log('Terminal WebSocket connected')
        setConnectionStatus('connected')
        reconnectAttempt = 0 // Reset on successful connection
        terminal.writeln('Welcome to Habibi-Go Terminal')
        terminal.writeln(`Session: ${currentSession.name}`)
        terminal.writeln(`Working Directory: ${currentSession.worktree_path}`)
        terminal.writeln('')
      }

      wsInstance.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          if (data.type === 'output') {
            terminal.write(data.data)
          } else if (data.type === 'error') {
            terminal.write(`\r\n\x1b[31mError: ${data.message}\x1b[0m\r\n`)
          } else if (data.type === 'info') {
            terminal.write(`\r\n\x1b[33m${data.message}\x1b[0m\r\n`)
          }
        } catch (error) {
          // If not JSON, treat as raw output
          terminal.write(event.data)
        }
      }

      wsInstance.onerror = (error) => {
        console.error('Terminal WebSocket error:', error)
        setConnectionStatus('error')
        terminal.write('\r\n\x1b[31mConnection error\x1b[0m\r\n')
      }

      wsInstance.onclose = (event) => {
        console.log('Terminal WebSocket disconnected, code:', event.code)
        
        // If this was not a clean close and we haven't exceeded retry attempts
        if (event.code !== 1000 && reconnectAttempt < maxReconnectAttempts) {
          setConnectionStatus('connecting')
          reconnectAttempt++
          const delay = Math.min(1000 * Math.pow(2, reconnectAttempt - 1), 5000) // Exponential backoff, max 5s
          
          terminal.write(`\r\n\x1b[33mConnection lost. Reconnecting in ${delay/1000}s... (attempt ${reconnectAttempt}/${maxReconnectAttempts})\x1b[0m\r\n`)
          
          setTimeout(() => {
            try {
              connectWebSocket()
            } catch (error) {
              console.error('Failed to reconnect:', error)
              setConnectionStatus('error')
              terminal.write('\r\n\x1b[31mReconnection failed\x1b[0m\r\n')
            }
          }, delay)
        } else {
          setConnectionStatus('disconnected')
          terminal.write('\r\n\x1b[33mConnection closed\x1b[0m\r\n')
          if (reconnectAttempt >= maxReconnectAttempts) {
            terminal.write('\r\n\x1b[31mMax reconnection attempts reached. Use the reconnect button to retry.\x1b[0m\r\n')
          }
        }
      }
    }

    setupWebSocket(ws)

    // Handle terminal input
    terminal.onData((data) => {
      if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
        wsRef.current.send(JSON.stringify({
          type: 'input',
          data: data
        }))
      }
    })

    // Handle window resize
    const handleResize = () => {
      fitAddon.fit()
    }
    window.addEventListener('resize', handleResize)

    // Cleanup
    return () => {
      window.removeEventListener('resize', handleResize)
      if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
        wsRef.current.close(1000) // Clean close
      }
      terminal.dispose()
      xtermRef.current = null
      fitAddonRef.current = null
      wsRef.current = null
    }
  }, [currentSession])

  // Fit terminal when component size changes
  useEffect(() => {
    const resizeObserver = new ResizeObserver(() => {
      if (fitAddonRef.current) {
        // Small delay to ensure container has updated size
        setTimeout(() => {
          fitAddonRef.current?.fit()
        }, 10)
      }
    })

    if (terminalRef.current) {
      resizeObserver.observe(terminalRef.current)
    }

    return () => {
      resizeObserver.disconnect()
    }
  }, [])

  if (!currentSession) {
    return (
      <div className="h-full flex items-center justify-center text-gray-500">
        <div className="text-center">
          <p className="text-lg mb-2">No session selected</p>
          <p className="text-sm">Select a session to use the terminal</p>
        </div>
      </div>
    )
  }

  return (
    <div className="h-full flex flex-col">
      <div className="p-4 border-b bg-gray-50">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-lg font-semibold">Terminal</h2>
            <p className="text-sm text-gray-600">
              Session: {currentSession.name} • {currentSession.worktree_path}
            </p>
          </div>
          <div className="flex items-center gap-2">
            <div className="flex items-center gap-2">
              <div className={`w-2 h-2 rounded-full ${
                connectionStatus === 'connected' ? 'bg-green-500' :
                connectionStatus === 'connecting' ? 'bg-yellow-500 animate-pulse' :
                connectionStatus === 'error' ? 'bg-red-500' :
                'bg-gray-500'
              }`} />
              <span className="text-xs text-gray-600">
                {connectionStatus === 'connected' ? 'Connected' :
                 connectionStatus === 'connecting' ? 'Connecting...' :
                 connectionStatus === 'error' ? 'Error' :
                 'Disconnected'}
              </span>
            </div>
            
            {(connectionStatus === 'disconnected' || connectionStatus === 'error') && (
              <button
                onClick={() => reconnectFuncRef.current?.()}
                className="text-sm px-3 py-1 bg-blue-500 text-white rounded hover:bg-blue-600"
              >
                Reconnect
              </button>
            )}
            
            <button
              onClick={() => {
                if (xtermRef.current) {
                  xtermRef.current.clear()
                }
              }}
              className="text-sm px-3 py-1 bg-gray-200 rounded hover:bg-gray-300"
            >
              Clear
            </button>
          </div>
        </div>
      </div>
      
      <div className="flex-1 bg-black">
        <div ref={terminalRef} className="h-full" />
      </div>
    </div>
  )
}