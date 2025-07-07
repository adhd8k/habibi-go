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
	sessionID   int
	cmd         *exec.Cmd
	pty         *os.File
	conn        *websocket.Conn
	done        chan bool
	cleanupOnce sync.Once
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
	h.terminalsMutex.RLock()
	existingTerminal, exists := h.terminals[sessionID]
	h.terminalsMutex.RUnlock()

	if exists {
		// Close existing terminal
		existingTerminal.cleanup()
	}

	// Create new terminal session
	terminal, err := h.createTerminalSession(sessionID, session.WorktreePath, conn)
	if err != nil {
		log.Printf("Failed to create terminal session for session %d: %v", sessionID, err)
		// Send error message to frontend
		conn.WriteJSON(map[string]interface{}{
			"type":    "error",
			"message": fmt.Sprintf("Failed to create terminal: %v", err),
		})
		// Don't return immediately - let the connection stay open for retry
		
		// Send a message suggesting reconnection
		conn.WriteJSON(map[string]interface{}{
			"type":    "info",
			"message": "Terminal session failed to start. You can try refreshing to reconnect.",
		})
		
		// Keep connection alive for potential retry
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				log.Printf("Terminal WebSocket connection closed for session %d: %v", sessionID, err)
				break
			}
		}
		return
	}

	// Store terminal session
	h.terminalsMutex.Lock()
	h.terminals[sessionID] = terminal
	h.terminalsMutex.Unlock()

	// Handle WebSocket messages
	h.handleTerminalMessages(terminal)

	// Cleanup when done
	h.terminalsMutex.Lock()
	delete(h.terminals, sessionID)
	h.terminalsMutex.Unlock()
	terminal.cleanup()
}

func (h *TerminalHandler) createTerminalSession(sessionID int, workingDir string, conn *websocket.Conn) (*TerminalSession, error) {
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
		sessionID: sessionID,
		cmd:       cmd,
		pty:       ptmx,
		conn:      conn,
		done:      make(chan bool),
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
				return
			}
			
			if n > 0 {
				data := string(buffer[:n])
				message := TerminalMessage{
					Type: "output",
					Data: data,
				}
				if err := terminal.conn.WriteJSON(message); err != nil {
					log.Printf("Failed to write to WebSocket: %v", err)
					return
				}
			}
		}
	}
}

func (h *TerminalHandler) handleTerminalMessages(terminal *TerminalSession) {
	for {
		select {
		case <-terminal.done:
			return
		default:
			var message TerminalMessage
			if err := terminal.conn.ReadJSON(&message); err != nil {
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
		// Close done channel
		close(ts.done)

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