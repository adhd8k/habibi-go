package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"habibi-go/internal/models"
	"habibi-go/internal/services"
)

type AgentHandler struct {
	agentService *services.AgentService
	commService  *services.AgentCommService
}

func NewAgentHandler(agentService *services.AgentService, commService *services.AgentCommService) *AgentHandler {
	return &AgentHandler{
		agentService: agentService,
		commService:  commService,
	}
}

func (h *AgentHandler) GetAgents(c *gin.Context) {
	sessionIDStr := c.Query("session_id")
	
	if sessionIDStr != "" {
		// Get agents for specific session
		sessionID, err := strconv.Atoi(sessionIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid session ID",
			})
			return
		}
		
		agents, err := h.agentService.GetAgentsBySession(sessionID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    agents,
		})
		return
	}
	
	// Get all agents
	agents, err := h.agentService.GetAllAgents()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    agents,
	})
}

func (h *AgentHandler) CreateAgent(c *gin.Context) {
	var req models.CreateAgentRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	
	agent, err := h.agentService.StartAgent(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    agent,
	})
}

func (h *AgentHandler) GetAgent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid agent ID",
		})
		return
	}
	
	agent, err := h.agentService.GetAgent(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    agent,
	})
}

func (h *AgentHandler) StopAgent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid agent ID",
		})
		return
	}
	
	if err := h.agentService.StopAgent(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Agent stopped successfully",
	})
}

func (h *AgentHandler) RestartAgent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid agent ID",
		})
		return
	}
	
	agent, err := h.agentService.RestartAgent(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    agent,
	})
}

func (h *AgentHandler) GetAgentStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid agent ID",
		})
		return
	}
	
	status, err := h.agentService.GetAgentStatus(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    status,
	})
}

func (h *AgentHandler) SendCommand(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid agent ID",
		})
		return
	}
	
	type CommandRequest struct {
		Command string `json:"command" binding:"required"`
		Wait    bool   `json:"wait"`
		Timeout int    `json:"timeout"` // seconds
	}
	
	var req CommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	
	if req.Wait {
		// Send command and wait for response
		timeout := 30 * time.Second
		if req.Timeout > 0 {
			timeout = time.Duration(req.Timeout) * time.Second
		}
		
		result, err := h.commService.SendCommandAndWait(id, req.Command, timeout)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    result,
		})
	} else {
		// Send command without waiting
		command, err := h.commService.SendCommand(id, req.Command)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    command,
		})
	}
}

func (h *AgentHandler) GetAgentLogs(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid agent ID",
		})
		return
	}
	
	// Parse since parameter
	sinceStr := c.Query("since")
	var since time.Time
	if sinceStr != "" {
		if parsedTime, err := time.Parse(time.RFC3339, sinceStr); err == nil {
			since = parsedTime
		} else {
			// Try parsing as duration (e.g., "1h", "30m")
			if duration, err := time.ParseDuration(sinceStr); err == nil {
				since = time.Now().Add(-duration)
			}
		}
	} else {
		// Default to last hour
		since = time.Now().Add(-1 * time.Hour)
	}
	
	logs, err := h.commService.GetAgentLogs(id, since)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    logs,
	})
}

func (h *AgentHandler) GetCommandHistory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid agent ID",
		})
		return
	}
	
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}
	
	commands, err := h.commService.GetCommandHistory(id, limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    commands,
	})
}

func (h *AgentHandler) GetAgentStats(c *gin.Context) {
	stats, err := h.agentService.GetAgentStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

func (h *AgentHandler) CleanupAgents(c *gin.Context) {
	if err := h.agentService.CleanupStaleAgents(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Stale agents cleaned up successfully",
	})
}

func (h *AgentHandler) GetSessionAgents(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid session ID",
		})
		return
	}
	
	agents, err := h.agentService.GetAgentsBySession(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    agents,
	})
}

func (h *AgentHandler) GetCommandStats(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid agent ID",
		})
		return
	}
	
	stats, err := h.commService.GetCommandStats(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}