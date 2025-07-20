package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"syscall"

	"github.com/creack/pty"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"habibi-go/internal/services"
)

type TerminalHandler struct {
	sessionService *services.SessionService
	upgrader       websocket.Upgrader
	terminals      map[int]*TerminalSession // sessionID -> terminal session
	terminalsMutex sync.RWMutex
}

type TerminalSession struct {
	sessionID    int
	cmd          *exec.Cmd
	pty          *os.File
	connections  map[*websocket.Conn]bool
	connMutex    sync.RWMutex
	outputBuffer []byte // Buffer to store recent output for reconnections
	bufferMutex  sync.RWMutex
	done         chan bool
	cleanupOnce  sync.Once
	isAlive      bool
}

type TerminalMessage struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

func NewTerminalHandler(sessionService *services.SessionService) *TerminalHandler {
	return &TerminalHandler{
		sessionService: sessionService,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow connections from any origin
			},
		},
		terminals: make(map[int]*TerminalSession),
	}
}

func (h *TerminalHandler) HandleTerminalWebSocket(c *gin.Context) {
	sessionIDStr := c.Param("sessionId")
	sessionID, err := strconv.Atoi(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	// Get session to validate it exists and get worktree path
	session, err := h.sessionService.GetSession(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	// Upgrade to WebSocket
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade to WebSocket: %v", err)
		return
	}
	defer conn.Close()

	// Check if terminal already exists for this session
	h.terminalsMutex.Lock()
	terminal, exists := h.terminals[sessionID]
	
	if exists && terminal.isAlive {
		// Add this connection to existing terminal
		terminal.connMutex.Lock()
		terminal.connections[conn] = true
		terminal.connMutex.Unlock()
		h.terminalsMutex.Unlock()
		
		log.Printf("Reconnecting to existing terminal for session %d", sessionID)
		
		// Send buffered output to catch up
		terminal.bufferMutex.RLock()
		if len(terminal.outputBuffer) > 0 {
			conn.WriteJSON(TerminalMessage{
				Type: "output",
				Data: string(terminal.outputBuffer),
			})
		}
		terminal.bufferMutex.RUnlock()
		
		// Handle messages for this connection
		h.handleConnectionMessages(terminal, conn)
		
		// Remove connection when done
		terminal.connMutex.Lock()
		delete(terminal.connections, conn)
		terminal.connMutex.Unlock()
		return
	}
	
	// Create new terminal if doesn't exist or is dead
	if exists && !terminal.isAlive {
		delete(h.terminals, sessionID)
	}
	h.terminalsMutex.Unlock()

	// Create new terminal session
	terminal, err = h.createTerminalSession(sessionID, session.WorktreePath)
	if err != nil {
		log.Printf("Failed to create terminal session for session %d: %v", sessionID, err)
		// Send error message to frontend
		conn.WriteJSON(map[string]interface{}{
			"type":    "error",
			"message": fmt.Sprintf("Failed to create terminal: %v", err),
		})
		return
	}

	// Add initial connection
	terminal.connMutex.Lock()
	terminal.connections[conn] = true
	terminal.connMutex.Unlock()

	// Store terminal session
	h.terminalsMutex.Lock()
	h.terminals[sessionID] = terminal
	h.terminalsMutex.Unlock()

	// Handle WebSocket messages for this connection
	h.handleConnectionMessages(terminal, conn)

	// Remove connection when done
	terminal.connMutex.Lock()
	delete(terminal.connections, conn)
	hasConnections := len(terminal.connections) > 0
	terminal.connMutex.Unlock()
	
	// Don't cleanup terminal if there are other connections
	if hasConnections {
		log.Printf("Connection closed for session %d, but %d connections remain", sessionID, len(terminal.connections))
	}
}

func (h *TerminalHandler) createTerminalSession(sessionID int, workingDir string) (*TerminalSession, error) {
	// Find available shell (NixOS compatibility)
	var shellPath string
	possibleShells := []string{
		"/bin/bash",
		"/usr/bin/bash", 
		"/run/current-system/sw/bin/bash", // NixOS system bash
		"/bin/sh",
		"/usr/bin/sh",
	}
	
	// Check which shell exists
	for _, shell := range possibleShells {
		if _, err := os.Stat(shell); err == nil {
			shellPath = shell
			break
		}
	}
	
	// Fallback to whatever bash is in PATH
	if shellPath == "" {
		if path, err := exec.LookPath("bash"); err == nil {
			shellPath = path
		} else if path, err := exec.LookPath("sh"); err == nil {
			shellPath = path
		} else {
			return nil, fmt.Errorf("no shell found (bash or sh)")
		}
	}
	
	log.Printf("Using shell: %s for session %d", shellPath, sessionID)
	
	// Create shell command
	cmd := exec.Command(shellPath)
	cmd.Dir = workingDir
	cmd.Env = append(os.Environ(),
		"TERM=xterm-256color",
		"COLUMNS=80",
		"LINES=24",
	)

	// Start the command with a PTY
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to start shell with PTY: %w", err)
	}

	// Set PTY size
	if err := pty.Setsize(ptmx, &pty.Winsize{Rows: 24, Cols: 80}); err != nil {
		log.Printf("Failed to set PTY size: %v", err)
	}

	terminal := &TerminalSession{
		sessionID:    sessionID,
		cmd:          cmd,
		pty:          ptmx,
		connections:  make(map[*websocket.Conn]bool),
		outputBuffer: make([]byte, 0, 1024*64), // 64KB buffer
		done:         make(chan bool),
		isAlive:      true,
	}

	// Start goroutine to handle PTY output
	go h.handlePTYOutput(terminal)

	// Monitor process
	go func() {
		cmd.Wait()
		// Use cleanup method to safely close everything
		terminal.cleanup()
	}()

	return terminal, nil
}

func (h *TerminalHandler) handlePTYOutput(terminal *TerminalSession) {
	buffer := make([]byte, 1024)
	for {
		select {
		case <-terminal.done:
			return
		default:
			n, err := terminal.pty.Read(buffer)
			if err != nil {
				if err != io.EOF {
					log.Printf("Error reading from PTY: %v", err)
				}
				terminal.isAlive = false
				return
			}
			
			if n > 0 {
				data := buffer[:n]
				
				// Update output buffer
				terminal.bufferMutex.Lock()
				terminal.outputBuffer = append(terminal.outputBuffer, data...)
				// Keep buffer size reasonable (last 64KB)
				if len(terminal.outputBuffer) > 64*1024 {
					terminal.outputBuffer = terminal.outputBuffer[len(terminal.outputBuffer)-64*1024:]
				}
				terminal.bufferMutex.Unlock()
				
				// Broadcast to all connections
				message := TerminalMessage{
					Type: "output",
					Data: string(data),
				}
				
				terminal.connMutex.RLock()
				for conn := range terminal.connections {
					if err := conn.WriteJSON(message); err != nil {
						log.Printf("Failed to write to WebSocket: %v", err)
					}
				}
				terminal.connMutex.RUnlock()
			}
		}
	}
}

func (h *TerminalHandler) handleConnectionMessages(terminal *TerminalSession, conn *websocket.Conn) {
	for {
		select {
		case <-terminal.done:
			return
		default:
			var message TerminalMessage
			if err := conn.ReadJSON(&message); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket error: %v", err)
				}
				return
			}

			switch message.Type {
			case "input":
				// Send input to PTY
				if _, err := terminal.pty.Write([]byte(message.Data)); err != nil {
					log.Printf("Failed to write to PTY: %v", err)
					return
				}
			case "resize":
				// Handle terminal resize (future enhancement)
				// Would need to parse cols/rows from message and call pty.Setsize()
			}
		}
	}
}

func (ts *TerminalSession) cleanup() {
	ts.cleanupOnce.Do(func() {
		// Mark as not alive
		ts.isAlive = false
		
		// Close done channel
		close(ts.done)

		// Close all connections
		ts.connMutex.Lock()
		for conn := range ts.connections {
			conn.Close()
		}
		ts.connections = make(map[*websocket.Conn]bool)
		ts.connMutex.Unlock()

		// Close PTY
		if ts.pty != nil {
			ts.pty.Close()
			ts.pty = nil
		}

		// Kill the process
		if ts.cmd != nil && ts.cmd.Process != nil {
			// Send SIGTERM first
			ts.cmd.Process.Signal(syscall.SIGTERM)
			
			// If it doesn't exit in a reasonable time, force kill
			// For now, just force kill immediately
			ts.cmd.Process.Kill()
			ts.cmd = nil
		}
	})
}

// CleanupSessionTerminal closes the terminal for a specific session
func (h *TerminalHandler) CleanupSessionTerminal(sessionID int) {
	h.terminalsMutex.Lock()
	defer h.terminalsMutex.Unlock()
	
	if terminal, exists := h.terminals[sessionID]; exists {
		log.Printf("Cleaning up terminal for session %d", sessionID)
		terminal.cleanup()
		delete(h.terminals, sessionID)
	}
}

// RestartTerminal restarts the terminal for a specific session
func (h *TerminalHandler) RestartTerminal(c *gin.Context) {
	sessionIDStr := c.Param("sessionId")
	sessionID, err := strconv.Atoi(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	// Clean up existing terminal if it exists
	h.terminalsMutex.Lock()
	if terminal, exists := h.terminals[sessionID]; exists {
		log.Printf("Restarting terminal for session %d", sessionID)
		terminal.cleanup()
		delete(h.terminals, sessionID)
	}
	h.terminalsMutex.Unlock()

	c.JSON(http.StatusOK, gin.H{"message": "Terminal restarted successfully"})
}