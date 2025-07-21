package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"habibi-go/internal/services"
)

type SlashCommandHandlers struct {
	slashCommandService *services.SlashCommandService
}

func NewSlashCommandHandlers(slashCommandService *services.SlashCommandService) *SlashCommandHandlers {
	return &SlashCommandHandlers{
		slashCommandService: slashCommandService,
	}
}

// GetCommands returns available slash commands for a session
func (h *SlashCommandHandlers) GetCommands(c *gin.Context) {
	sessionID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	commands, err := h.slashCommandService.GetAvailableCommands(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, commands)
}

// ExecuteCommand executes a slash command
func (h *SlashCommandHandlers) ExecuteCommand(c *gin.Context) {
	sessionID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	var req struct {
		Command string `json:"command" binding:"required"`
		Args    string `json:"args"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.slashCommandService.ExecuteCommand(c.Request.Context(), sessionID, req.Command, req.Args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}