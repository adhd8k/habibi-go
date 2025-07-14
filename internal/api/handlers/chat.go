package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"habibi-go/internal/database/repositories"
	"habibi-go/internal/models"
)

type ChatHandler struct {
	chatRepo    *repositories.ChatMessageV2Repository
	sessionRepo *repositories.SessionRepository
}

func NewChatHandler(chatRepo *repositories.ChatMessageV2Repository, sessionRepo *repositories.SessionRepository) *ChatHandler {
	return &ChatHandler{
		chatRepo:    chatRepo,
		sessionRepo: sessionRepo,
	}
}

func (h *ChatHandler) GetSessionChatHistory(c *gin.Context) {
	sessionIDStr := c.Param("id")
	sessionID, err := strconv.Atoi(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	messages, err := h.chatRepo.GetBySessionID(sessionID, 100) // Get last 100 messages
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get chat history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    messages,
		"success": true,
	})
}

func (h *ChatHandler) DeleteSessionChatHistory(c *gin.Context) {
	sessionIDStr := c.Param("id")
	sessionID, err := strconv.Atoi(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	err = h.chatRepo.DeleteBySessionID(sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete chat history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Chat history deleted",
	})
}

func (h *ChatHandler) SendChatMessage(c *gin.Context) {
	sessionIDStr := c.Param("id")
	sessionID, err := strconv.Atoi(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	var request struct {
		Content string `json:"content" binding:"required"`
		Role    string `json:"role"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Default role to user if not specified
	if request.Role == "" {
		request.Role = "user"
	}

	// Validate role
	if request.Role != "user" && request.Role != "assistant" && request.Role != "system" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
		return
	}

	// Check if session exists
	_, err = h.sessionRepo.GetByID(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	// Create message
	message := &models.ChatMessage{
		SessionID: sessionID,
		Role:      request.Role,
		Content:   request.Content,
	}

	// Save message
	if err := h.chatRepo.Create(message); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save message"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data":    message,
		"success": true,
	})
}