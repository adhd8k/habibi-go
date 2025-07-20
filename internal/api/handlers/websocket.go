package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
	hub           *Hub
	claudeService *services.ClaudeSessionService
}

func NewWebSocketHandler(claudeService *services.ClaudeSessionService) *WebSocketHandler {
	handler := &WebSocketHandler{
		hub:           NewHub(),
		claudeService: claudeService,
	}
	
	// Set the event broadcaster
	claudeService.SetEventBroadcaster(handler)
	
	return handler
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
		hub:     h.hub,
		conn:    conn,
		send:    make(chan []byte, 256),
		handler: h,
	}
	
	client.hub.register <- client
	
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
	maxMessageSize = 32768 // Increased from 512 to 32KB
)

func (c *Client) readPump() {
	defer func() {
		log.Printf("Client readPump exiting")
		c.hub.unregister <- c
		c.conn.Close()
	}()
	
	log.Printf("Client readPump started")
	
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		log.Printf("Received pong")
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	
	for {
		log.Printf("Waiting for WebSocket message...")
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("ReadMessage error: %v", err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
		
		log.Printf("ReadMessage successful, handling message")
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
	Type      string      `json:"type"`
	SessionID interface{} `json:"session_id,omitempty"`
	Data      interface{} `json:"data,omitempty"`
}

func (c *Client) handleMessage(message []byte) {
	log.Printf("Received WebSocket message: %s", string(message))
	
	var msg WSMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		c.sendError(fmt.Sprintf("Invalid message format: %v", err))
		return
	}
	
	log.Printf("Parsed message type: %s", msg.Type)
	
	switch msg.Type {
	case "session_chat":
		c.handleSessionChat(msg)
	case "stop_generation":
		c.handleStopGeneration(msg)
	case "ping":
		c.sendMessage(WSMessage{Type: "pong"})
	default:
		log.Printf("Unknown message type: %s", msg.Type)
		c.sendError(fmt.Sprintf("Unknown message type: %s", msg.Type))
	}
}

func (c *Client) handleSessionChat(msg WSMessage) {
	log.Printf("Received session_chat message: %+v", msg)
	
	sessionID, ok := msg.Data.(map[string]interface{})["session_id"].(float64)
	if !ok || sessionID == 0 {
		log.Printf("Session ID error: %v", msg.Data)
		c.sendError("Session ID is required")
		return
	}
	
	message, ok := msg.Data.(map[string]interface{})["message"].(string)
	if !ok || message == "" {
		log.Printf("Message error: %v", msg.Data)
		c.sendError("Message is required")
		return
	}
	
	log.Printf("Sending message to Claude service for session %d: %s", int(sessionID), message)
	
	// Send message via Claude service
	if err := c.handler.claudeService.SendMessage(int(sessionID), message); err != nil {
		log.Printf("Claude service error: %v", err)
		c.sendError(fmt.Sprintf("Failed to send message: %v", err))
		return
	}
	
	log.Printf("Message sent successfully, sending acknowledgment")
	
	// Send acknowledgment
	c.sendMessage(WSMessage{
		Type: "chat_sent",
		Data: map[string]interface{}{
			"session_id": int(sessionID),
			"status":     "sent",
		},
	})
}

func (c *Client) handleStopGeneration(msg WSMessage) {
	log.Printf("Received stop_generation message: %+v", msg)
	
	sessionID, ok := msg.Data.(map[string]interface{})["session_id"].(float64)
	if !ok || sessionID == 0 {
		log.Printf("Session ID error: %v", msg.Data)
		c.sendError("Session ID is required")
		return
	}
	
	log.Printf("Stopping generation for session %d", int(sessionID))
	
	// Stop the generation via Claude service
	if err := c.handler.claudeService.StopGeneration(int(sessionID)); err != nil {
		log.Printf("Failed to stop generation: %v", err)
		c.sendError(fmt.Sprintf("Failed to stop generation: %v", err))
		return
	}
	
	log.Printf("Generation stopped successfully, sending acknowledgment")
	
	// Send acknowledgment
	c.sendMessage(WSMessage{
		Type: "generation_stopped",
		Data: map[string]interface{}{
			"session_id": int(sessionID),
			"status":     "stopped",
		},
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

// Implement EventBroadcaster interface
func (h *WebSocketHandler) BroadcastEvent(eventType string, agentID int, data interface{}) {
	msg := WSMessage{
		Type: eventType,
		Data: data,
	}
	
	// For session-based events, extract session_id from data
	if dataMap, ok := data.(map[string]interface{}); ok {
		if sessionID, exists := dataMap["session_id"]; exists {
			msg.SessionID = sessionID
		}
		
		// Special logging for TodoWrite
		if eventType == "claude_output" {
			if contentType, ok := dataMap["content_type"].(string); ok && contentType == "tool_use" {
				if toolName, ok := dataMap["tool_name"].(string); ok && toolName == "TodoWrite" {
					log.Printf("TODOWRITE BROADCAST: %+v", dataMap)
				}
			}
		}
	}
	
	msgData, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal broadcast message: %v", err)
		return
	}
	
	log.Printf("Broadcasting WebSocket message: %s", string(msgData))
	h.hub.broadcast <- msgData
}

// BroadcastSessionUpdate broadcasts a session update to all connected clients
func (h *WebSocketHandler) BroadcastSessionUpdate(session interface{}) {
	msg := WSMessage{
		Type: "session_update",
		Data: map[string]interface{}{
			"session": session,
		},
	}
	
	msgData, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal session update message: %v", err)
		return
	}
	
	h.hub.broadcast <- msgData
}