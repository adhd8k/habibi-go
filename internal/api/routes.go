package api

import (
	"github.com/gin-gonic/gin"
	"habibi-go/internal/api/handlers"
	"habibi-go/internal/api/middleware"
)

type Router struct {
	projectHandler *handlers.ProjectHandler
	sessionHandler *handlers.SessionHandler
}

func NewRouter(projectHandler *handlers.ProjectHandler, sessionHandler *handlers.SessionHandler) *Router {
	return &Router{
		projectHandler: projectHandler,
		sessionHandler: sessionHandler,
	}
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
	
	// Health check
	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
			"timestamp": "2024-01-01T00:00:00Z",
		})
	})
}