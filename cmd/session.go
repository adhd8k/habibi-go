package cmd

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"habibi-go/internal/config"
	"habibi-go/internal/database"
	"habibi-go/internal/database/repositories"
	"habibi-go/internal/models"
	"habibi-go/internal/services"
)

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Session management commands",
	Long:  `Create, list, update, and delete sessions.`,
}

var sessionListCmd = &cobra.Command{
	Use:   "list [project-name]",
	Short: "List sessions for a project",
	Args:  cobra.MaximumNArgs(1),
	Run:   runSessionList,
}

var sessionCreateCmd = &cobra.Command{
	Use:   "create [project-name] [session-name] [branch-name]",
	Short: "Create a new session",
	Args:  cobra.ExactArgs(3),
	Run:   runSessionCreate,
}

var sessionShowCmd = &cobra.Command{
	Use:   "show [session-id]",
	Short: "Show session details",
	Args:  cobra.ExactArgs(1),
	Run:   runSessionShow,
}

var sessionActivateCmd = &cobra.Command{
	Use:   "activate [session-id]",
	Short: "Activate a session",
	Args:  cobra.ExactArgs(1),
	Run:   runSessionActivate,
}

var sessionDeleteCmd = &cobra.Command{
	Use:   "delete [session-id]",
	Short: "Delete a session",
	Args:  cobra.ExactArgs(1),
	Run:   runSessionDelete,
}

var sessionCleanupCmd = &cobra.Command{
	Use:   "cleanup [project-name]",
	Short: "Clean up stopped sessions",
	Args:  cobra.MaximumNArgs(1),
	Run:   runSessionCleanup,
}

var sessionSyncCmd = &cobra.Command{
	Use:   "sync [session-id]",
	Short: "Sync session with remote branch",
	Args:  cobra.ExactArgs(1),
	Run:   runSessionSync,
}

func init() {
	sessionCmd.AddCommand(sessionListCmd)
	sessionCmd.AddCommand(sessionCreateCmd)
	sessionCmd.AddCommand(sessionShowCmd)
	sessionCmd.AddCommand(sessionActivateCmd)
	sessionCmd.AddCommand(sessionDeleteCmd)
	sessionCmd.AddCommand(sessionCleanupCmd)
	sessionCmd.AddCommand(sessionSyncCmd)
}

func getSessionService() (*services.SessionService, *services.ProjectService) {
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
	
	projectRepo := repositories.NewProjectRepository(db.DB)
	sessionRepo := repositories.NewSessionRepository(db.DB)
	eventRepo := repositories.NewEventRepository(db.DB)
	
	gitService := services.NewGitService(cfg.Projects.WorktreeBasePath)
	sshService := services.NewSSHService()
	projectService := services.NewProjectService(projectRepo, eventRepo, gitService)
	sessionService := services.NewSessionService(sessionRepo, projectRepo, eventRepo, gitService, sshService)
	
	return sessionService, projectService
}

func runSessionList(cmd *cobra.Command, args []string) {
	sessionService, projectService := getSessionService()
	
	var sessions []*models.Session
	var err error
	
	if len(args) > 0 {
		// List sessions for specific project
		project, err := projectService.GetProjectByName(args[0])
		if err != nil {
			log.Fatalf("Failed to get project: %v", err)
		}
		
		sessions, err = sessionService.GetSessionsByProject(project.ID)
		if err != nil {
			log.Fatalf("Failed to get sessions: %v", err)
		}
		
		fmt.Printf("Sessions for project '%s':\n\n", project.Name)
	} else {
		// List all sessions
		sessions, err = sessionService.GetAllSessions()
		if err != nil {
			log.Fatalf("Failed to get sessions: %v", err)
		}
		
		fmt.Println("All sessions:\n")
	}
	
	if len(sessions) == 0 {
		fmt.Println("No sessions found")
		return
	}
	
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tBRANCH\tSTATUS\tPROJECT_ID\tLAST_USED")
	
	for _, session := range sessions {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%d\t%s\n",
			session.ID,
			session.Name,
			session.BranchName,
			session.Status,
			session.ProjectID,
			session.LastUsedAt.Format("2006-01-02 15:04"),
		)
	}
	
	w.Flush()
}

func runSessionCreate(cmd *cobra.Command, args []string) {
	sessionService, projectService := getSessionService()
	
	// Get project by name
	project, err := projectService.GetProjectByName(args[0])
	if err != nil {
		log.Fatalf("Failed to get project: %v", err)
	}
	
	req := &models.CreateSessionRequest{
		ProjectID:  project.ID,
		Name:       args[1],
		BranchName: args[2],
	}
	
	session, err := sessionService.CreateSession(req)
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}
	
	fmt.Printf("Session '%s' created successfully (ID: %d)\n", session.Name, session.ID)
	fmt.Printf("Worktree path: %s\n", session.WorktreePath)
	fmt.Printf("Branch: %s\n", session.BranchName)
}

func runSessionShow(cmd *cobra.Command, args []string) {
	sessionService, _ := getSessionService()
	
	id, err := strconv.Atoi(args[0])
	if err != nil {
		log.Fatalf("Invalid session ID: %v", err)
	}
	
	status, err := sessionService.GetSessionStatus(id)
	if err != nil {
		log.Fatalf("Failed to get session status: %v", err)
	}
	
	session := status.Session
	
	fmt.Printf("Session: %s (ID: %d)\n", session.Name, session.ID)
	fmt.Printf("Project ID: %d\n", session.ProjectID)
	fmt.Printf("Branch: %s\n", session.BranchName)
	fmt.Printf("Status: %s\n", session.Status)
	fmt.Printf("Worktree Path: %s\n", session.WorktreePath)
	fmt.Printf("Worktree Exists: %t\n", status.WorktreeExists)
	fmt.Printf("Created: %s\n", session.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Last Used: %s\n", session.LastUsedAt.Format("2006-01-02 15:04:05"))
	
	if status.WorktreeStatus != nil {
		fmt.Printf("\nGit Status:\n")
		fmt.Printf("Current Branch: %s\n", status.WorktreeStatus.Branch)
		fmt.Printf("Commit: %s\n", status.WorktreeStatus.Commit)
		fmt.Printf("Uncommitted Changes: %t\n", status.WorktreeStatus.HasUncommittedChanges)
		
		if status.WorktreeStatus.GitStatus != "" {
			fmt.Printf("Git Status Output:\n%s\n", status.WorktreeStatus.GitStatus)
		}
	}
	
	if len(session.Config) > 0 {
		fmt.Println("\nConfiguration:")
		for key, value := range session.Config {
			fmt.Printf("  %s: %v\n", key, value)
		}
	}
}

func runSessionActivate(cmd *cobra.Command, args []string) {
	sessionService, _ := getSessionService()
	
	id, err := strconv.Atoi(args[0])
	if err != nil {
		log.Fatalf("Invalid session ID: %v", err)
	}
	
	session, err := sessionService.ActivateSession(id)
	if err != nil {
		log.Fatalf("Failed to activate session: %v", err)
	}
	
	fmt.Printf("Session '%s' activated successfully\n", session.Name)
}

func runSessionDelete(cmd *cobra.Command, args []string) {
	sessionService, _ := getSessionService()
	
	id, err := strconv.Atoi(args[0])
	if err != nil {
		log.Fatalf("Invalid session ID: %v", err)
	}
	
	// Get session details first for confirmation
	session, err := sessionService.GetSession(id)
	if err != nil {
		log.Fatalf("Failed to get session: %v", err)
	}
	
	if err := sessionService.DeleteSession(id); err != nil {
		log.Fatalf("Failed to delete session: %v", err)
	}
	
	fmt.Printf("Session '%s' deleted successfully\n", session.Name)
}

func runSessionCleanup(cmd *cobra.Command, args []string) {
	sessionService, projectService := getSessionService()
	
	var projectID int
	if len(args) > 0 {
		project, err := projectService.GetProjectByName(args[0])
		if err != nil {
			log.Fatalf("Failed to get project: %v", err)
		}
		projectID = project.ID
		fmt.Printf("Cleaning up stopped sessions for project '%s'...\n", project.Name)
	} else {
		fmt.Println("Cleaning up all stopped sessions...")
	}
	
	if err := sessionService.CleanupStoppedSessions(projectID); err != nil {
		log.Fatalf("Failed to cleanup sessions: %v", err)
	}
	
	fmt.Println("Session cleanup completed successfully")
}

func runSessionSync(cmd *cobra.Command, args []string) {
	sessionService, _ := getSessionService()
	
	id, err := strconv.Atoi(args[0])
	if err != nil {
		log.Fatalf("Invalid session ID: %v", err)
	}
	
	if err := sessionService.SyncSessionBranch(id); err != nil {
		log.Fatalf("Failed to sync session: %v", err)
	}
	
	fmt.Printf("Session synced successfully\n")
}