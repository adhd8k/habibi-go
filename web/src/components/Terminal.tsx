import { useEffect, useRef, useState } from 'react'
import '@xterm/xterm/css/xterm.css'
import { useAppStore } from '../store'
import { useTerminalManager } from '../hooks/useTerminalManager'

export function Terminal() {
  const terminalRef = useRef<HTMLDivElement>(null)
  const [connectionStatus, setConnectionStatus] = useState<'connecting' | 'connected' | 'disconnected' | 'error'>('connecting')
  const { currentSession } = useAppStore()
  const { 
    switchToSession, 
    fitActiveTerminal, 
    clearActiveTerminal, 
    getActiveTerminal 
  } = useTerminalManager()

  useEffect(() => {
    if (!terminalRef.current || !currentSession) return

    // Switch to the terminal for this session
    const instance = switchToSession(currentSession.id, terminalRef.current)
    
    // Connect to backend terminal websocket if not already connected
    if (!instance.ws || instance.ws.readyState !== WebSocket.OPEN) {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      const wsUrl = `${protocol}//${window.location.host}/api/terminal/${currentSession.id}`
      
      let reconnectAttempt = 0
      const maxReconnectAttempts = 5
      
      const connectWebSocket = () => {
        setConnectionStatus('connecting')
        const newWs = new WebSocket(wsUrl)
        instance.ws = newWs
        setupWebSocket(newWs)
        return newWs
      }
      
      const setupWebSocket = (wsInstance: WebSocket) => {
        wsInstance.onopen = () => {
          console.log('Terminal WebSocket connected')
          setConnectionStatus('connected')
          reconnectAttempt = 0 // Reset on successful connection
          instance.terminal.writeln('Welcome to Habibi-Go Terminal')
          instance.terminal.writeln(`Session: ${currentSession.name}`)
          instance.terminal.writeln(`Working Directory: ${currentSession.worktree_path}`)
          instance.terminal.writeln('')
        }

        wsInstance.onmessage = (event) => {
          try {
            const data = JSON.parse(event.data)
            if (data.type === 'output') {
              instance.terminal.write(data.data)
            } else if (data.type === 'error') {
              instance.terminal.write(`\r\n\x1b[31mError: ${data.message}\x1b[0m\r\n`)
            } else if (data.type === 'info') {
              instance.terminal.write(`\r\n\x1b[33m${data.message}\x1b[0m\r\n`)
            }
          } catch (error) {
            // If not JSON, treat as raw output
            instance.terminal.write(event.data)
          }
        }

        wsInstance.onerror = (error) => {
          console.error('Terminal WebSocket error:', error)
          setConnectionStatus('error')
          instance.terminal.write('\r\n\x1b[31mConnection error\x1b[0m\r\n')
        }

        wsInstance.onclose = (event) => {
          console.log('Terminal WebSocket disconnected, code:', event.code)
          
          // If this was not a clean close and we haven't exceeded retry attempts
          if (event.code !== 1000 && reconnectAttempt < maxReconnectAttempts) {
            setConnectionStatus('connecting')
            reconnectAttempt++
            const delay = Math.min(1000 * Math.pow(2, reconnectAttempt - 1), 5000) // Exponential backoff, max 5s
            
            instance.terminal.write(`\r\n\x1b[33mConnection lost. Reconnecting in ${delay/1000}s... (attempt ${reconnectAttempt}/${maxReconnectAttempts})\x1b[0m\r\n`)
            
            setTimeout(() => {
              try {
                connectWebSocket()
              } catch (error) {
                console.error('Failed to reconnect:', error)
                setConnectionStatus('error')
                instance.terminal.write('\r\n\x1b[31mReconnection failed\x1b[0m\r\n')
              }
            }, delay)
          } else {
            setConnectionStatus('disconnected')
            instance.terminal.write('\r\n\x1b[33mConnection closed\x1b[0m\r\n')
            if (reconnectAttempt >= maxReconnectAttempts) {
              instance.terminal.write('\r\n\x1b[31mMax reconnection attempts reached. Use the reconnect button to retry.\x1b[0m\r\n')
            }
          }
        }
      }

      connectWebSocket()

      // Handle terminal input
      instance.terminal.onData((data) => {
        if (instance.ws && instance.ws.readyState === WebSocket.OPEN) {
          instance.ws.send(JSON.stringify({
            type: 'input',
            data: data
          }))
        }
      })
    } else {
      // WebSocket already connected, just update status
      setConnectionStatus('connected')
    }

    // Handle window resize
    const handleResize = () => {
      fitActiveTerminal()
    }
    window.addEventListener('resize', handleResize)

    // Cleanup - we don't dispose the terminal anymore, just remove event listener
    return () => {
      window.removeEventListener('resize', handleResize)
    }
  }, [currentSession, switchToSession, fitActiveTerminal])

  // Fit terminal when component size changes
  useEffect(() => {
    const resizeObserver = new ResizeObserver(() => {
      // Small delay to ensure container has updated size
      setTimeout(() => {
        fitActiveTerminal()
      }, 10)
    })

    if (terminalRef.current) {
      resizeObserver.observe(terminalRef.current)
    }

    return () => {
      resizeObserver.disconnect()
    }
  }, [fitActiveTerminal])
  
  const handleReconnect = () => {
    const instance = getActiveTerminal()
    if (!instance || !currentSession) return
    
    // Close existing websocket
    if (instance.ws) {
      instance.ws.close(1000)
    }
    
    instance.terminal.write('\r\n\x1b[36mReconnecting...\x1b[0m\r\n')
    
    // Force reconnect by clearing the effect dependency
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${window.location.host}/api/terminal/${currentSession.id}`
    
    setConnectionStatus('connecting')
    const newWs = new WebSocket(wsUrl)
    instance.ws = newWs
    
    // Set up the websocket handlers
    let reconnectAttempt = 0
    const maxReconnectAttempts = 5
    
    const connectWebSocket = () => {
      setConnectionStatus('connecting')
      const ws = new WebSocket(wsUrl)
      instance.ws = ws
      setupWebSocket(ws)
      return ws
    }
    
    const setupWebSocket = (wsInstance: WebSocket) => {
      wsInstance.onopen = () => {
        console.log('Terminal WebSocket connected')
        setConnectionStatus('connected')
        reconnectAttempt = 0
        instance.terminal.writeln('Reconnected successfully')
      }

      wsInstance.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          if (data.type === 'output') {
            instance.terminal.write(data.data)
          } else if (data.type === 'error') {
            instance.terminal.write(`\r\n\x1b[31mError: ${data.message}\x1b[0m\r\n`)
          } else if (data.type === 'info') {
            instance.terminal.write(`\r\n\x1b[33m${data.message}\x1b[0m\r\n`)
          }
        } catch (error) {
          instance.terminal.write(event.data)
        }
      }

      wsInstance.onerror = (error) => {
        console.error('Terminal WebSocket error:', error)
        setConnectionStatus('error')
        instance.terminal.write('\r\n\x1b[31mConnection error\x1b[0m\r\n')
      }

      wsInstance.onclose = (event) => {
        console.log('Terminal WebSocket disconnected, code:', event.code)
        
        if (event.code !== 1000 && reconnectAttempt < maxReconnectAttempts) {
          setConnectionStatus('connecting')
          reconnectAttempt++
          const delay = Math.min(1000 * Math.pow(2, reconnectAttempt - 1), 5000)
          
          instance.terminal.write(`\r\n\x1b[33mConnection lost. Reconnecting in ${delay/1000}s... (attempt ${reconnectAttempt}/${maxReconnectAttempts})\x1b[0m\r\n`)
          
          setTimeout(() => {
            try {
              connectWebSocket()
            } catch (error) {
              console.error('Failed to reconnect:', error)
              setConnectionStatus('error')
              instance.terminal.write('\r\n\x1b[31mReconnection failed\x1b[0m\r\n')
            }
          }, delay)
        } else {
          setConnectionStatus('disconnected')
          instance.terminal.write('\r\n\x1b[33mConnection closed\x1b[0m\r\n')
          if (reconnectAttempt >= maxReconnectAttempts) {
            instance.terminal.write('\r\n\x1b[31mMax reconnection attempts reached. Use the reconnect button to retry.\x1b[0m\r\n')
          }
        }
      }
    }
    
    setupWebSocket(newWs)
    
    // Re-attach input handler
    instance.terminal.onData((data) => {
      if (instance.ws && instance.ws.readyState === WebSocket.OPEN) {
        instance.ws.send(JSON.stringify({
          type: 'input',
          data: data
        }))
      }
    })
  }

  if (!currentSession) {
    return (
      <div className="h-full flex items-center justify-center text-gray-500 dark:text-gray-400">
        <div className="text-center">
          <p className="text-lg mb-2">No session selected</p>
          <p className="text-sm">Select a session to use the terminal</p>
        </div>
      </div>
    )
  }

  return (
    <div className="h-full flex flex-col">
      <div className="p-4 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Terminal</h2>
            <p className="text-sm text-gray-600 dark:text-gray-400">
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
              <span className="text-xs text-gray-600 dark:text-gray-400">
                {connectionStatus === 'connected' ? 'Connected' :
                 connectionStatus === 'connecting' ? 'Connecting...' :
                 connectionStatus === 'error' ? 'Error' :
                 'Disconnected'}
              </span>
            </div>
            
            {(connectionStatus === 'disconnected' || connectionStatus === 'error') && (
              <button
                onClick={handleReconnect}
                className="text-sm px-3 py-1 bg-blue-500 text-white rounded hover:bg-blue-600"
              >
                Reconnect
              </button>
            )}
            
            <button
              onClick={() => clearActiveTerminal()}
              className="text-sm px-3 py-1 bg-gray-200 dark:bg-gray-700 rounded hover:bg-gray-300 dark:hover:bg-gray-600 text-gray-900 dark:text-gray-100"
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