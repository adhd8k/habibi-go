package cmd

import (
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"habibi-go/internal/config"
	"habibi-go/internal/database"
	"habibi-go/internal/database/repositories"
	"habibi-go/internal/models"
	"habibi-go/internal/services"
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Project management commands",
	Long:  `Create, list, update, and delete projects.`,
}

var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects",
	Run:   runProjectList,
}

var projectCreateCmd = &cobra.Command{
	Use:   "create [name] [path]",
	Short: "Create a new project",
	Args:  cobra.ExactArgs(2),
	Run:   runProjectCreate,
}

var projectShowCmd = &cobra.Command{
	Use:   "show [name]",
	Short: "Show project details",
	Args:  cobra.ExactArgs(1),
	Run:   runProjectShow,
}

var projectDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a project",
	Args:  cobra.ExactArgs(1),
	Run:   runProjectDelete,
}

var projectDiscoverCmd = &cobra.Command{
	Use:   "discover [directory]",
	Short: "Auto-discover projects in a directory",
	Args:  cobra.ExactArgs(1),
	Run:   runProjectDiscover,
}

var (
	projectRepo        string
	projectBranch      string
	projectAutoCreate  bool
)

func init() {
	projectCmd.AddCommand(projectListCmd)
	projectCmd.AddCommand(projectCreateCmd)
	projectCmd.AddCommand(projectShowCmd)
	projectCmd.AddCommand(projectDeleteCmd)
	projectCmd.AddCommand(projectDiscoverCmd)
	
	// Flags for create command
	projectCreateCmd.Flags().StringVar(&projectRepo, "repo", "", "repository URL")
	projectCreateCmd.Flags().StringVar(&projectBranch, "branch", "main", "default branch")
	
	// Flags for discover command
	projectDiscoverCmd.Flags().BoolVar(&projectAutoCreate, "auto-create", false, "automatically create discovered projects")
}

func getProjectService() *services.ProjectService {
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
	eventRepo := repositories.NewEventRepository(db.DB)
	
	return services.NewProjectService(projectRepo, eventRepo)
}

func runProjectList(cmd *cobra.Command, args []string) {
	projectService := getProjectService()
	
	projects, err := projectService.GetAllProjects()
	if err != nil {
		log.Fatalf("Failed to get projects: %v", err)
	}
	
	if len(projects) == 0 {
		fmt.Println("No projects found")
		return
	}
	
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tPATH\tREPOSITORY\tBRANCH\tCREATED")
	
	for _, project := range projects {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\n",
			project.ID,
			project.Name,
			project.Path,
			project.RepositoryURL,
			project.DefaultBranch,
			project.CreatedAt.Format("2006-01-02 15:04"),
		)
	}
	
	w.Flush()
}

func runProjectCreate(cmd *cobra.Command, args []string) {
	projectService := getProjectService()
	
	req := &models.CreateProjectRequest{
		Name:          args[0],
		Path:          args[1],
		RepositoryURL: projectRepo,
		DefaultBranch: projectBranch,
	}
	
	project, err := projectService.CreateProject(req)
	if err != nil {
		log.Fatalf("Failed to create project: %v", err)
	}
	
	fmt.Printf("Project '%s' created successfully (ID: %d)\n", project.Name, project.ID)
}

func runProjectShow(cmd *cobra.Command, args []string) {
	projectService := getProjectService()
	
	project, err := projectService.GetProjectByName(args[0])
	if err != nil {
		log.Fatalf("Failed to get project: %v", err)
	}
	
	fmt.Printf("Project: %s (ID: %d)\n", project.Name, project.ID)
	fmt.Printf("Path: %s\n", project.Path)
	fmt.Printf("Repository: %s\n", project.RepositoryURL)
	fmt.Printf("Default Branch: %s\n", project.DefaultBranch)
	fmt.Printf("Created: %s\n", project.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated: %s\n", project.UpdatedAt.Format("2006-01-02 15:04:05"))
	
	if len(project.Config) > 0 {
		fmt.Println("Configuration:")
		for key, value := range project.Config {
			fmt.Printf("  %s: %v\n", key, value)
		}
	}
}

func runProjectDelete(cmd *cobra.Command, args []string) {
	projectService := getProjectService()
	
	project, err := projectService.GetProjectByName(args[0])
	if err != nil {
		log.Fatalf("Failed to get project: %v", err)
	}
	
	if err := projectService.DeleteProject(project.ID); err != nil {
		log.Fatalf("Failed to delete project: %v", err)
	}
	
	fmt.Printf("Project '%s' deleted successfully\n", project.Name)
}

func runProjectDiscover(cmd *cobra.Command, args []string) {
	projectService := getProjectService()
	
	projects, err := projectService.DiscoverProjects(args[0])
	if err != nil {
		log.Fatalf("Failed to discover projects: %v", err)
	}
	
	if len(projects) == 0 {
		fmt.Println("No git repositories found")
		return
	}
	
	fmt.Printf("Found %d git repositories:\n\n", len(projects))
	
	for _, project := range projects {
		fmt.Printf("- %s (%s)\n", project.Name, project.Path)
		
		if projectAutoCreate {
			req := &models.CreateProjectRequest{
				Name:          project.Name,
				Path:          project.Path,
				DefaultBranch: "main",
			}
			
			if _, err := projectService.CreateProject(req); err != nil {
				fmt.Printf("  Error creating project: %v\n", err)
			} else {
				fmt.Printf("  ✓ Created project\n")
			}
		}
	}
	
	if !projectAutoCreate {
		fmt.Println("\nUse --auto-create flag to automatically create these projects")
	}
}