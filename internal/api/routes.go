package api

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
	
	"github.com/gin-gonic/gin"
	"habibi-go/internal/api/handlers"
	"habibi-go/internal/api/middleware"
	"habibi-go/internal/config"
)

type Router struct {
	projectHandler   *handlers.ProjectHandler
	sessionHandler   *handlers.SessionHandler
	websocketHandler *handlers.WebSocketHandler
	chatHandler      *handlers.ChatHandler
	terminalHandler  *handlers.TerminalHandler
	webAssets        embed.FS
	authConfig       *config.AuthConfig
}

func NewRouter(
	projectHandler *handlers.ProjectHandler, 
	sessionHandler *handlers.SessionHandler, 
	websocketHandler *handlers.WebSocketHandler,
	chatHandler *handlers.ChatHandler,
	terminalHandler *handlers.TerminalHandler,
) *Router {
	return &Router{
		projectHandler:   projectHandler,
		sessionHandler:   sessionHandler,
		websocketHandler: websocketHandler,
		chatHandler:      chatHandler,
		terminalHandler:  terminalHandler,
	}
}

func (r *Router) SetAuthConfig(authConfig *config.AuthConfig) {
	r.authConfig = authConfig
}

func (r *Router) SetWebAssets(assets embed.FS) {
	r.webAssets = assets
}

func (r *Router) SetupRoutes(engine *gin.Engine) {
	// Apply middleware
	engine.Use(middleware.CORS())
	engine.Use(middleware.Logger())
	
	// Apply auth middleware if configured
	if r.authConfig != nil {
		engine.Use(middleware.BasicAuth(r.authConfig))
	}
	
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
		sessions.GET("/:id/diffs", r.sessionHandler.GetSessionDiffs)
		sessions.POST("/:id/rebase", r.sessionHandler.RebaseSession)
		sessions.POST("/:id/push", r.sessionHandler.PushSession)
		sessions.POST("/:id/merge", r.sessionHandler.MergeSession)
		sessions.POST("/:id/merge-to-original", r.sessionHandler.MergeSessionToOriginal)
		sessions.POST("/:id/close", r.sessionHandler.CloseSession)
		
		// Chat history for sessions
		sessions.GET("/:id/chat", r.chatHandler.GetSessionChatHistory)
		sessions.DELETE("/:id/chat", r.chatHandler.DeleteSessionChatHistory)
		sessions.POST("/:id/chat", r.chatHandler.SendChatMessage)
	}
	
	// WebSocket endpoint
	api.GET("/ws", r.websocketHandler.HandleWebSocket)
	
	// Terminal WebSocket endpoint
	api.GET("/terminal/:sessionId", r.terminalHandler.HandleTerminalWebSocket)
	
	// WebSocket endpoint (also available on root for compatibility)
	engine.GET("/ws", r.websocketHandler.HandleWebSocket)
	
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
			v1Projects.GET("/:id/sessions", r.sessionHandler.GetProjectSessions)
			v1Projects.POST("/discover", r.projectHandler.DiscoverProjects)
			v1Projects.GET("/stats", r.projectHandler.GetProjectStats)
		}
		
		// Sessions routes
		v1Sessions := v1.Group("/sessions")
		{
			v1Sessions.GET("", r.sessionHandler.GetSessions)
			v1Sessions.POST("", r.sessionHandler.CreateSession)
			v1Sessions.GET("/:id", r.sessionHandler.GetSession)
			v1Sessions.PUT("/:id", r.sessionHandler.UpdateSession)
			v1Sessions.DELETE("/:id", r.sessionHandler.DeleteSession)
			v1Sessions.GET("/:id/diffs", r.sessionHandler.GetSessionDiffs)
			v1Sessions.POST("/:id/rebase", r.sessionHandler.RebaseSession)
			v1Sessions.POST("/:id/push", r.sessionHandler.PushSession)
			v1Sessions.POST("/:id/merge", r.sessionHandler.MergeSession)
			v1Sessions.POST("/:id/merge-to-original", r.sessionHandler.MergeSessionToOriginal)
			v1Sessions.POST("/:id/close", r.sessionHandler.CloseSession)
			v1Sessions.GET("/:id/chat", r.chatHandler.GetSessionChatHistory)
			v1Sessions.DELETE("/:id/chat", r.chatHandler.DeleteSessionChatHistory)
			v1Sessions.POST("/:id/chat", r.chatHandler.SendChatMessage)
		}
	}
	
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