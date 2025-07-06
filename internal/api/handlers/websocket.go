package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"habibi-go/internal/services"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

type WebSocketHandler struct {
	hub         *Hub
	agentService *services.AgentService
	commService  *services.AgentCommService
}

func NewWebSocketHandler(agentService *services.AgentService, commService *services.AgentCommService) *WebSocketHandler {
	return &WebSocketHandler{
		hub:          NewHub(),
		agentService: agentService,
		commService:  commService,
	}
}

func (h *WebSocketHandler) StartHub() {
	go h.hub.Run()
}

func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	
	client := &Client{
		hub:      h.hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		handler:  h,
	}
	
	client.hub.register <- client
	
	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Printf("Client connected. Total clients: %d", len(h.clients))
			
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("Client disconnected. Total clients: %d", len(h.clients))
			}
			
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

type Client struct {
	hub     *Hub
	conn    *websocket.Conn
	send    chan []byte
	handler *WebSocketHandler
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
		
		// Handle incoming message
		c.handleMessage(message)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			
			// Add queued messages to the current message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}
			
			if err := w.Close(); err != nil {
				return
			}
			
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

type WSMessage struct {
	Type    string      `json:"type"`
	AgentID int         `json:"agent_id,omitempty"`
	Command string      `json:"command,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func (c *Client) handleMessage(message []byte) {
	var msg WSMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		c.sendError(fmt.Sprintf("Invalid message format: %v", err))
		return
	}
	
	switch msg.Type {
	case "agent_command":
		c.handleAgentCommand(msg)
	case "agent_logs_subscribe":
		c.handleAgentLogsSubscribe(msg)
	case "agent_status_request":
		c.handleAgentStatusRequest(msg)
	case "ping":
		c.sendMessage(WSMessage{Type: "pong"})
	default:
		c.sendError(fmt.Sprintf("Unknown message type: %s", msg.Type))
	}
}

func (c *Client) handleAgentCommand(msg WSMessage) {
	if msg.AgentID == 0 {
		c.sendError("Agent ID is required")
		return
	}
	
	if msg.Command == "" {
		c.sendError("Command is required")
		return
	}
	
	// Send command to agent
	command, err := c.handler.commService.SendCommand(msg.AgentID, msg.Command)
	if err != nil {
		c.sendError(fmt.Sprintf("Failed to send command: %v", err))
		return
	}
	
	// Send acknowledgment
	c.sendMessage(WSMessage{
		Type:    "command_sent",
		AgentID: msg.AgentID,
		Data: map[string]interface{}{
			"command_id": command.ID,
			"status":     "sent",
		},
	})
}

func (c *Client) handleAgentLogsSubscribe(msg WSMessage) {
	if msg.AgentID == 0 {
		c.sendError("Agent ID is required")
		return
	}
	
	// Start streaming logs for this agent
	logStream, err := c.handler.commService.StreamAgentOutput(msg.AgentID)
	if err != nil {
		c.sendError(fmt.Sprintf("Failed to stream logs: %v", err))
		return
	}
	
	// Start goroutine to forward logs
	go func() {
		for log := range logStream {
			c.sendMessage(WSMessage{
				Type:    "agent_log",
				AgentID: msg.AgentID,
				Data: map[string]interface{}{
					"message":   log,
					"timestamp": time.Now(),
				},
			})
		}
	}()
	
	// Send subscription confirmation
	c.sendMessage(WSMessage{
		Type:    "logs_subscribed",
		AgentID: msg.AgentID,
		Data:    map[string]string{"status": "subscribed"},
	})
}

func (c *Client) handleAgentStatusRequest(msg WSMessage) {
	if msg.AgentID == 0 {
		c.sendError("Agent ID is required")
		return
	}
	
	// Get agent status
	status, err := c.handler.agentService.GetAgentStatus(msg.AgentID)
	if err != nil {
		c.sendError(fmt.Sprintf("Failed to get agent status: %v", err))
		return
	}
	
	c.sendMessage(WSMessage{
		Type:    "agent_status",
		AgentID: msg.AgentID,
		Data:    status,
	})
}

func (c *Client) sendMessage(msg WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}
	
	select {
	case c.send <- data:
	default:
		close(c.send)
		c.hub.unregister <- c
	}
}

func (c *Client) sendError(message string) {
	c.sendMessage(WSMessage{
		Type: "error",
		Data: map[string]string{"error": message},
	})
}

// BroadcastEvent broadcasts an event to all connected clients
func (h *WebSocketHandler) BroadcastEvent(eventType string, agentID int, data interface{}) {
	msg := WSMessage{
		Type:    eventType,
		AgentID: agentID,
		Data:    data,
	}
	
	msgData, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal broadcast message: %v", err)
		return
	}
	
	h.hub.broadcast <- msgData
}

// Helper function to get agent ID from URL parameter
func getAgentIDFromPath(c *gin.Context) (int, error) {
	idStr := c.Param("id")
	return strconv.Atoi(idStr)
}