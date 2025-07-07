package cmd

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"habibi-go/internal/config"
	"habibi-go/internal/database"
	"habibi-go/internal/database/repositories"
	"habibi-go/internal/models"
	"habibi-go/internal/services"
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Agent management commands",
	Long:  `Start, stop, and monitor coding agents.`,
}

var agentListCmd = &cobra.Command{
	Use:   "list [session-id]",
	Short: "List agents for a session",
	Args:  cobra.MaximumNArgs(1),
	Run:   runAgentList,
}

var agentStartCmd = &cobra.Command{
	Use:   "start [session-id] [agent-type] [command]",
	Short: "Start a new agent",
	Args:  cobra.ExactArgs(3),
	Run:   runAgentStart,
}

var agentStopCmd = &cobra.Command{
	Use:   "stop [agent-id]",
	Short: "Stop an agent",
	Args:  cobra.ExactArgs(1),
	Run:   runAgentStop,
}

var agentStatusCmd = &cobra.Command{
	Use:   "status [agent-id]",
	Short: "Show agent status",
	Args:  cobra.ExactArgs(1),
	Run:   runAgentStatus,
}

var agentExecCmd = &cobra.Command{
	Use:   "exec [agent-id] [command]",
	Short: "Execute command on agent",
	Args:  cobra.ExactArgs(2),
	Run:   runAgentExec,
}

var agentLogsCmd = &cobra.Command{
	Use:   "logs [agent-id]",
	Short: "Show agent logs",
	Args:  cobra.ExactArgs(1),
	Run:   runAgentLogs,
}

var agentRestartCmd = &cobra.Command{
	Use:   "restart [agent-id]",
	Short: "Restart an agent",
	Args:  cobra.ExactArgs(1),
	Run:   runAgentRestart,
}

var agentCleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Clean up stale agents",
	Args:  cobra.NoArgs,
	Run:   runAgentCleanup,
}

var (
	agentTimeout    int
	agentFollow     bool
	agentSince      string
)

func init() {
	agentCmd.AddCommand(agentListCmd)
	agentCmd.AddCommand(agentStartCmd)
	agentCmd.AddCommand(agentStopCmd)
	agentCmd.AddCommand(agentStatusCmd)
	agentCmd.AddCommand(agentExecCmd)
	agentCmd.AddCommand(agentLogsCmd)
	agentCmd.AddCommand(agentRestartCmd)
	agentCmd.AddCommand(agentCleanupCmd)
	
	// Flags for exec command
	agentExecCmd.Flags().IntVar(&agentTimeout, "timeout", 30, "timeout in seconds")
	
	// Flags for logs command
	agentLogsCmd.Flags().BoolVar(&agentFollow, "follow", false, "follow log output")
	agentLogsCmd.Flags().StringVar(&agentSince, "since", "1h", "show logs since (duration or RFC3339 timestamp)")
}

func getAgentServices() (*services.AgentService, *services.AgentCommService, *services.SessionService) {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	
	// Create necessary directories
	if err := cfg.CreateDirectories(); err != nil {
		log.Fatalf("Failed to create directories: %v", err)
	}
	
	db, err := database.New(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	
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
	
	// Initialize services
	gitService := services.NewGitService()
	agentService := services.NewAgentService(agentRepo, sessionRepo, eventRepo)
	
	// Initialize Claude service
	claudeBinaryPath := "claude"
	if cfg.Agents.ClaudeBinaryPath != "" {
		claudeBinaryPath = cfg.Agents.ClaudeBinaryPath
	}
	claudeService := services.NewClaudeAgentService(agentRepo, eventRepo, claudeBinaryPath)
	agentService.SetClaudeService(claudeService)
	
	commService := services.NewAgentCommService(agentRepo, commandRepo, eventRepo, agentService, claudeService)
	sessionService := services.NewSessionService(sessionRepo, projectRepo, eventRepo, gitService)
	
	return agentService, commService, sessionService
}

func runAgentList(cmd *cobra.Command, args []string) {
	agentService, _, _ := getAgentServices()
	
	var agents []*models.Agent
	var err error
	
	if len(args) > 0 {
		// List agents for specific session
		sessionID, err := strconv.Atoi(args[0])
		if err != nil {
			log.Fatalf("Invalid session ID: %v", err)
		}
		
		agents, err = agentService.GetAgentsBySession(sessionID)
		if err != nil {
			log.Fatalf("Failed to get agents: %v", err)
		}
		
		fmt.Printf("Agents for session %d:\n\n", sessionID)
	} else {
		// List all agents
		agents, err = agentService.GetAllAgents()
		if err != nil {
			log.Fatalf("Failed to get agents: %v", err)
		}
		
		fmt.Println("All agents:\n")
	}
	
	if len(agents) == 0 {
		fmt.Println("No agents found")
		return
	}
	
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTYPE\tSTATUS\tPID\tSESSION_ID\tSTARTED")
	
	for _, agent := range agents {
		fmt.Fprintf(w, "%d\t%s\t%s\t%d\t%d\t%s\n",
			agent.ID,
			agent.AgentType,
			agent.Status,
			agent.PID,
			agent.SessionID,
			agent.StartedAt.Format("2006-01-02 15:04"),
		)
	}
	
	w.Flush()
}

func runAgentStart(cmd *cobra.Command, args []string) {
	agentService, _, _ := getAgentServices()
	
	sessionID, err := strconv.Atoi(args[0])
	if err != nil {
		log.Fatalf("Invalid session ID: %v", err)
	}
	
	req := &models.CreateAgentRequest{
		SessionID: sessionID,
		AgentType: args[1],
		Command:   args[2],
	}
	
	agent, err := agentService.StartAgent(req)
	if err != nil {
		log.Fatalf("Failed to start agent: %v", err)
	}
	
	fmt.Printf("Agent started successfully (ID: %d)\n", agent.ID)
	fmt.Printf("Type: %s\n", agent.AgentType)
	fmt.Printf("PID: %d\n", agent.PID)
	fmt.Printf("Command: %s\n", agent.Command)
	fmt.Printf("Working Directory: %s\n", agent.WorkingDirectory)
}

func runAgentStop(cmd *cobra.Command, args []string) {
	agentService, _, _ := getAgentServices()
	
	id, err := strconv.Atoi(args[0])
	if err != nil {
		log.Fatalf("Invalid agent ID: %v", err)
	}
	
	if err := agentService.StopAgent(id); err != nil {
		log.Fatalf("Failed to stop agent: %v", err)
	}
	
	fmt.Printf("Agent %d stopped successfully\n", id)
}

func runAgentStatus(cmd *cobra.Command, args []string) {
	agentService, _, _ := getAgentServices()
	
	id, err := strconv.Atoi(args[0])
	if err != nil {
		log.Fatalf("Invalid agent ID: %v", err)
	}
	
	status, err := agentService.GetAgentStatus(id)
	if err != nil {
		log.Fatalf("Failed to get agent status: %v", err)
	}
	
	agent := status.Agent
	
	fmt.Printf("Agent: %d\n", agent.ID)
	fmt.Printf("Type: %s\n", agent.AgentType)
	fmt.Printf("Status: %s\n", agent.Status)
	fmt.Printf("PID: %d\n", agent.PID)
	fmt.Printf("Session ID: %d\n", agent.SessionID)
	fmt.Printf("Command: %s\n", agent.Command)
	fmt.Printf("Working Directory: %s\n", agent.WorkingDirectory)
	fmt.Printf("Communication Method: %s\n", agent.CommunicationMethod)
	fmt.Printf("Started: %s\n", agent.StartedAt.Format("2006-01-02 15:04:05"))
	
	if agent.StoppedAt != nil {
		fmt.Printf("Stopped: %s\n", agent.StoppedAt.Format("2006-01-02 15:04:05"))
	}
	
	if agent.LastHeartbeat != nil {
		fmt.Printf("Last Heartbeat: %s\n", agent.LastHeartbeat.Format("2006-01-02 15:04:05"))
	}
	
	fmt.Printf("\nStatus Info:\n")
	fmt.Printf("Is Active: %t\n", status.IsActive)
	fmt.Printf("Process Exists: %t\n", status.ProcessExists)
	fmt.Printf("Is Healthy: %t\n", status.IsHealthy)
	
	if !status.LastSeen.IsZero() {
		fmt.Printf("Last Seen: %s\n", status.LastSeen.Format("2006-01-02 15:04:05"))
	}
	
	if status.ProcessInfo != nil {
		fmt.Printf("\nProcess Info:\n")
		fmt.Printf("CPU: %.1f%%\n", status.ProcessInfo.CPUPercent)
		fmt.Printf("Memory: %d MB\n", status.ProcessInfo.MemoryMB)
		fmt.Printf("Process Status: %s\n", status.ProcessInfo.Status)
	}
}

func runAgentExec(cmd *cobra.Command, args []string) {
	_, commService, _ := getAgentServices()
	
	id, err := strconv.Atoi(args[0])
	if err != nil {
		log.Fatalf("Invalid agent ID: %v", err)
	}
	
	command := args[1]
	timeout := time.Duration(agentTimeout) * time.Second
	
	fmt.Printf("Executing command on agent %d: %s\n", id, command)
	
	result, err := commService.SendCommandAndWait(id, command, timeout)
	if err != nil {
		log.Fatalf("Failed to execute command: %v", err)
	}
	
	fmt.Printf("\nCommand executed successfully\n")
	fmt.Printf("Execution time: %v\n", result.ExecutionTime)
	fmt.Printf("Success: %t\n", result.Success)
	
	if result.Response != "" {
		fmt.Printf("\nResponse:\n%s\n", result.Response)
	}
}

func runAgentLogs(cmd *cobra.Command, args []string) {
	_, commService, _ := getAgentServices()
	
	id, err := strconv.Atoi(args[0])
	if err != nil {
		log.Fatalf("Invalid agent ID: %v", err)
	}
	
	// Parse since parameter
	var since time.Time
	if duration, err := time.ParseDuration(agentSince); err == nil {
		since = time.Now().Add(-duration)
	} else if parsedTime, err := time.Parse(time.RFC3339, agentSince); err == nil {
		since = parsedTime
	} else {
		log.Fatalf("Invalid since format: %s", agentSince)
	}
	
	logs, err := commService.GetAgentLogs(id, since)
	if err != nil {
		log.Fatalf("Failed to get agent logs: %v", err)
	}
	
	if len(logs) == 0 {
		fmt.Printf("No logs found since %s\n", since.Format("2006-01-02 15:04:05"))
		return
	}
	
	fmt.Printf("Agent %d logs since %s:\n\n", id, since.Format("2006-01-02 15:04:05"))
	
	for _, logLine := range logs {
		fmt.Println(logLine)
	}
	
	if agentFollow {
		fmt.Println("\n--- Following logs (Ctrl+C to stop) ---")
		
		// Stream real-time logs
		outputChan, err := commService.StreamAgentOutput(id)
		if err != nil {
			log.Fatalf("Failed to stream logs: %v", err)
		}
		
		for output := range outputChan {
			fmt.Printf("[%s] %s\n", time.Now().Format("15:04:05"), output)
		}
	}
}

func runAgentRestart(cmd *cobra.Command, args []string) {
	agentService, _, _ := getAgentServices()
	
	id, err := strconv.Atoi(args[0])
	if err != nil {
		log.Fatalf("Invalid agent ID: %v", err)
	}
	
	agent, err := agentService.RestartAgent(id)
	if err != nil {
		log.Fatalf("Failed to restart agent: %v", err)
	}
	
	fmt.Printf("Agent restarted successfully\n")
	fmt.Printf("New ID: %d\n", agent.ID)
	fmt.Printf("New PID: %d\n", agent.PID)
}

func runAgentCleanup(cmd *cobra.Command, args []string) {
	agentService, _, _ := getAgentServices()
	
	if err := agentService.CleanupStaleAgents(); err != nil {
		log.Fatalf("Failed to cleanup agents: %v", err)
	}
	
	fmt.Println("Stale agents cleaned up successfully")
}