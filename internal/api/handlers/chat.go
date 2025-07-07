package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"habibi-go/internal/database/repositories"
)

type ChatHandler struct {
	chatRepo *repositories.ChatMessageRepository
}

func NewChatHandler(chatRepo *repositories.ChatMessageRepository) *ChatHandler {
	return &ChatHandler{
		chatRepo: chatRepo,
	}
}

// GetAgentChatHistory returns chat history for an agent
func (h *ChatHandler) GetAgentChatHistory(c *gin.Context) {
	agentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid agent ID",
			"success": false,
		})
		return
	}
	
	// Get limit from query param, default to 100
	limit := 100
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	
	messages, err := h.chatRepo.GetByAgentID(agentID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"success":  true,
	})
}

// DeleteAgentChatHistory deletes all chat history for an agent
func (h *ChatHandler) DeleteAgentChatHistory(c *gin.Context) {
	agentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid agent ID",
			"success": false,
		})
		return
	}
	
	if err := h.chatRepo.DeleteByAgentID(agentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Chat history deleted successfully",
		"success": true,
	})
}