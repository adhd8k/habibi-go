package cmd

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"habibi-go/internal/api"
	"habibi-go/internal/api/handlers"
	"habibi-go/internal/config"
	"habibi-go/internal/database"
	"habibi-go/internal/database/repositories"
	"habibi-go/internal/services"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the web server",
	Long:  `Start the web server with the specified configuration.`,
	Run:   runServer,
}

var (
	serverPort int
	serverHost string
	devMode    bool
	webAssets  embed.FS
)

func SetWebAssets(assets embed.FS) {
	webAssets = assets
}

func init() {
	serverCmd.Flags().IntVarP(&serverPort, "port", "p", 8080, "server port")
	serverCmd.Flags().StringVarP(&serverHost, "host", "H", "localhost", "server host")
	serverCmd.Flags().BoolVar(&devMode, "dev", false, "development mode")
}

func runServer(cmd *cobra.Command, args []string) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	
	// Override with command line flags
	if serverPort != 8080 {
		cfg.Server.Port = serverPort
	}
	if serverHost != "localhost" {
		cfg.Server.Host = serverHost
	}
	
	// Create necessary directories
	if err := cfg.CreateDirectories(); err != nil {
		log.Fatalf("Failed to create directories: %v", err)
	}
	
	// Initialize database
	db, err := database.New(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()
	
	// Run migrations
	if err := db.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	
	// Initialize repositories
	projectRepo := repositories.NewProjectRepository(db.DB)
	sessionRepo := repositories.NewSessionRepository(db.DB)
	agentRepo := repositories.NewAgentRepository(db.DB)
	commandRepo := repositories.NewAgentCommandRepository(db.DB)
	eventRepo := repositories.NewEventRepository(db.DB)
	chatRepo := repositories.NewChatMessageRepository(db.DB)
	
	// Initialize services
	gitService := services.NewGitService()
	projectService := services.NewProjectService(projectRepo, eventRepo, gitService)
	sessionService := services.NewSessionService(sessionRepo, projectRepo, eventRepo, gitService)
	agentService := services.NewAgentService(agentRepo, sessionRepo, eventRepo)
	
	// Configure agent service with Claude binary path
	claudeBinaryPath := "claude"
	if cfg.Agents.ClaudeBinaryPath != "" {
		claudeBinaryPath = cfg.Agents.ClaudeBinaryPath
		agentService.SetClaudeBinaryPath(cfg.Agents.ClaudeBinaryPath)
	}
	
	// Initialize Claude-specific service
	claudeService := services.NewClaudeAgentService(agentRepo, eventRepo, chatRepo, claudeBinaryPath)
	
	// Set Claude service on agent service
	agentService.SetClaudeService(claudeService)
	
	commService := services.NewAgentCommService(agentRepo, commandRepo, eventRepo, agentService, claudeService)
	
	// Initialize handlers
	projectHandler := handlers.NewProjectHandler(projectService)
	sessionHandler := handlers.NewSessionHandler(sessionService)
	agentHandler := handlers.NewAgentHandler(agentService, commService)
	websocketHandler := handlers.NewWebSocketHandler(agentService, commService)
	chatHandler := handlers.NewChatHandler(chatRepo)
	terminalHandler := handlers.NewTerminalHandler(sessionService)
	
	// Wire up WebSocket event broadcasting
	agentService.SetEventBroadcaster(websocketHandler)
	claudeService.SetEventBroadcaster(websocketHandler)
	
	// Recover agents from previous run
	if err := agentService.RecoverActiveAgents(); err != nil {
		log.Printf("Warning: failed to recover agents: %v", err)
	}
	
	// Start WebSocket hub
	websocketHandler.StartHub()
	
	// Initialize router
	router := api.NewRouter(projectHandler, sessionHandler, agentHandler, websocketHandler, chatHandler, terminalHandler)
	
	// Set web assets if available
	if webAssets != (embed.FS{}) {
		router.SetWebAssets(webAssets)
	}
	
	// Setup Gin
	if !devMode {
		gin.SetMode(gin.ReleaseMode)
	}
	
	engine := gin.New()
	router.SetupRoutes(engine)
	
	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      engine,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}
	
	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on %s:%d", cfg.Server.Host, cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()
	
	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("Server shutting down...")
	
	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()
	
	// Shutdown server
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	
	log.Println("Server exited")
}