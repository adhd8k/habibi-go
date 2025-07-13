import { useRef, useCallback } from 'react'
import { Terminal as XTerm } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'

interface TerminalInstance {
  terminal: XTerm
  fitAddon: FitAddon
  ws: WebSocket | null
  sessionId: number
  container: HTMLDivElement | null
}

export function useTerminalManager() {
  const terminalsRef = useRef<Map<number, TerminalInstance>>(new Map())
  const activeSessionIdRef = useRef<number | null>(null)

  const createTerminal = useCallback((sessionId: number): TerminalInstance => {
    // Create new terminal
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

    return {
      terminal,
      fitAddon,
      ws: null,
      sessionId,
      container: null
    }
  }, [])

  const getOrCreateTerminal = useCallback((sessionId: number): TerminalInstance => {
    let instance = terminalsRef.current.get(sessionId)
    if (!instance) {
      instance = createTerminal(sessionId)
      terminalsRef.current.set(sessionId, instance)
    }
    return instance
  }, [createTerminal])

  const switchToSession = useCallback((sessionId: number, container: HTMLDivElement) => {
    // Hide current terminal if any
    if (activeSessionIdRef.current !== null && activeSessionIdRef.current !== sessionId) {
      const currentInstance = terminalsRef.current.get(activeSessionIdRef.current)
      if (currentInstance?.container) {
        // Clear the container but keep the terminal instance
        currentInstance.container.innerHTML = ''
        currentInstance.container = null
      }
    }

    // Get or create terminal for new session
    const instance = getOrCreateTerminal(sessionId)
    
    // Attach to new container
    if (!instance.container || instance.container !== container) {
      container.innerHTML = '' // Clear any existing content
      instance.terminal.open(container)
      instance.fitAddon.fit()
      instance.container = container
    }

    activeSessionIdRef.current = sessionId
    return instance
  }, [getOrCreateTerminal])

  const closeTerminal = useCallback((sessionId: number) => {
    const instance = terminalsRef.current.get(sessionId)
    if (instance) {
      if (instance.ws && instance.ws.readyState === WebSocket.OPEN) {
        instance.ws.close(1000)
      }
      instance.terminal.dispose()
      terminalsRef.current.delete(sessionId)
      
      if (activeSessionIdRef.current === sessionId) {
        activeSessionIdRef.current = null
      }
    }
  }, [])

  const fitActiveTerminal = useCallback(() => {
    if (activeSessionIdRef.current !== null) {
      const instance = terminalsRef.current.get(activeSessionIdRef.current)
      if (instance) {
        instance.fitAddon.fit()
      }
    }
  }, [])

  const clearActiveTerminal = useCallback(() => {
    if (activeSessionIdRef.current !== null) {
      const instance = terminalsRef.current.get(activeSessionIdRef.current)
      if (instance) {
        instance.terminal.clear()
      }
    }
  }, [])

  const getActiveTerminal = useCallback(() => {
    if (activeSessionIdRef.current !== null) {
      return terminalsRef.current.get(activeSessionIdRef.current)
    }
    return null
  }, [])

  return {
    switchToSession,
    closeTerminal,
    fitActiveTerminal,
    clearActiveTerminal,
    getActiveTerminal,
    activeSessionId: activeSessionIdRef.current
  }
}