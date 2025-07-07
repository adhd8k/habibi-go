package api

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
	
	"github.com/gin-gonic/gin"
	"habibi-go/internal/api/handlers"
	"habibi-go/internal/api/middleware"
)

type Router struct {
	projectHandler   *handlers.ProjectHandler
	sessionHandler   *handlers.SessionHandler
	agentHandler     *handlers.AgentHandler
	websocketHandler *handlers.WebSocketHandler
	chatHandler      *handlers.ChatHandler
	webAssets        embed.FS
}

func NewRouter(projectHandler *handlers.ProjectHandler, sessionHandler *handlers.SessionHandler, agentHandler *handlers.AgentHandler, websocketHandler *handlers.WebSocketHandler, chatHandler *handlers.ChatHandler) *Router {
	return &Router{
		projectHandler:   projectHandler,
		sessionHandler:   sessionHandler,
		agentHandler:     agentHandler,
		websocketHandler: websocketHandler,
		chatHandler:      chatHandler,
	}
}

func (r *Router) SetWebAssets(assets embed.FS) {
	r.webAssets = assets
}

func (r *Router) SetupRoutes(engine *gin.Engine) {
	// Apply middleware
	engine.Use(middleware.CORS())
	engine.Use(middleware.Logger())
	
	// API routes
	api := engine.Group("/api")
	
	// Projects routes
	projects := api.Group("/projects")
	{
		projects.GET("", r.projectHandler.GetProjects)
		projects.POST("", r.projectHandler.CreateProject)
		projects.GET("/:id", r.projectHandler.GetProject)
		projects.PUT("/:id", r.projectHandler.UpdateProject)
		projects.DELETE("/:id", r.projectHandler.DeleteProject)
		projects.GET("/:id/sessions", r.sessionHandler.GetProjectSessions)
		projects.POST("/discover", r.projectHandler.DiscoverProjects)
		projects.GET("/stats", r.projectHandler.GetProjectStats)
	}
	
	// Sessions routes
	sessions := api.Group("/sessions")
	{
		sessions.GET("", r.sessionHandler.GetSessions)
		sessions.POST("", r.sessionHandler.CreateSession)
		sessions.GET("/:id", r.sessionHandler.GetSession)
		sessions.PUT("/:id", r.sessionHandler.UpdateSession)
		sessions.DELETE("/:id", r.sessionHandler.DeleteSession)
		sessions.POST("/:id/activate", r.sessionHandler.ActivateSession)
		sessions.GET("/:id/status", r.sessionHandler.GetSessionStatus)
		sessions.POST("/:id/sync", r.sessionHandler.SyncSession)
		sessions.POST("/cleanup", r.sessionHandler.CleanupSessions)
		sessions.GET("/stats", r.sessionHandler.GetSessionStats)
	}
	
	// Agents routes
	agents := api.Group("/agents")
	{
		agents.GET("", r.agentHandler.GetAgents)
		agents.POST("", r.agentHandler.CreateAgent)
		agents.GET("/:id", r.agentHandler.GetAgent)
		agents.POST("/:id/stop", r.agentHandler.StopAgent)
		agents.POST("/:id/restart", r.agentHandler.RestartAgent)
		agents.GET("/:id/status", r.agentHandler.GetAgentStatus)
		agents.POST("/:id/command", r.agentHandler.SendCommand)
		agents.GET("/:id/logs", r.agentHandler.GetAgentLogs)
		agents.GET("/:id/commands", r.agentHandler.GetCommandHistory)
		agents.GET("/:id/stats", r.agentHandler.GetCommandStats)
		agents.POST("/cleanup", r.agentHandler.CleanupAgents)
		agents.GET("/stats", r.agentHandler.GetAgentStats)
		
		// Chat history endpoints
		agents.GET("/:id/chat", r.chatHandler.GetAgentChatHistory)
		agents.DELETE("/:id/chat", r.chatHandler.DeleteAgentChatHistory)
	}
	
	// WebSocket endpoint
	api.GET("/ws", r.websocketHandler.HandleWebSocket)
	
	// Health check
	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
			"timestamp": "2024-01-01T00:00:00Z",
		})
	})
	
	// API v1 routes (for frontend compatibility)
	v1 := engine.Group("/api/v1")
	{
		// Projects routes
		v1Projects := v1.Group("/projects")
		{
			v1Projects.GET("", r.projectHandler.GetProjects)
			v1Projects.POST("", r.projectHandler.CreateProject)
			v1Projects.GET("/:id", r.projectHandler.GetProject)
			v1Projects.PUT("/:id", r.projectHandler.UpdateProject)
			v1Projects.DELETE("/:id", r.projectHandler.DeleteProject)
		}
		
		// Sessions routes
		v1Sessions := v1.Group("/sessions")
		{
			v1Sessions.GET("", r.sessionHandler.GetSessions)
			v1Sessions.POST("", r.sessionHandler.CreateSession)
			v1Sessions.GET("/:id", r.sessionHandler.GetSession)
			v1Sessions.PUT("/:id", r.sessionHandler.UpdateSession)
			v1Sessions.DELETE("/:id", r.sessionHandler.DeleteSession)
		}
		
		// Agents routes
		v1Agents := v1.Group("/agents")
		{
			v1Agents.GET("", r.agentHandler.GetAgents)
			v1Agents.POST("", r.agentHandler.CreateAgent)
			v1Agents.GET("/:id", r.agentHandler.GetAgent)
			v1Agents.GET("/:id/status", r.agentHandler.GetAgentStatus)
			v1Agents.POST("/:id/stop", r.agentHandler.StopAgent)
			v1Agents.POST("/:id/restart", r.agentHandler.RestartAgent)
			v1Agents.POST("/:id/execute", r.agentHandler.SendCommand)
			v1Agents.GET("/:id/logs", r.agentHandler.GetAgentLogs)
			v1Agents.GET("/:id/chat", r.chatHandler.GetAgentChatHistory)
			v1Agents.DELETE("/:id/chat", r.chatHandler.DeleteAgentChatHistory)
		}
	}
	
	// WebSocket endpoint (also available on root)
	engine.GET("/ws", r.websocketHandler.HandleWebSocket)
	
	// Serve static files if webAssets is set
	if r.webAssets != (embed.FS{}) {
		// Get the sub filesystem for web/dist
		distFS, err := fs.Sub(r.webAssets, "web/dist")
		if err == nil {
			// Get assets subdirectory
			assetsFS, _ := fs.Sub(distFS, "assets")
			
			// Serve static files from embedded filesystem
			engine.StaticFS("/assets", http.FS(assetsFS))
			
			// Serve index.html for root
			engine.GET("/", func(c *gin.Context) {
				data, err := fs.ReadFile(distFS, "index.html")
				if err != nil {
					c.String(500, "Failed to load index.html")
					return
				}
				c.Data(200, "text/html; charset=utf-8", data)
			})
			
			// Serve index.html for any non-API/asset routes (SPA support)
			engine.NoRoute(func(c *gin.Context) {
				path := c.Request.URL.Path
				// Don't serve index.html for API routes, WebSocket, or assets
				if !strings.HasPrefix(path, "/api") && 
				   !strings.HasPrefix(path, "/ws") &&
				   !strings.HasPrefix(path, "/assets") {
					data, err := fs.ReadFile(distFS, "index.html")
					if err != nil {
						c.String(500, "Failed to load index.html")
						return
					}
					c.Data(200, "text/html; charset=utf-8", data)
				} else {
					c.JSON(404, gin.H{"error": "Not found"})
				}
			})
		}
	}
}