package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"habibi-go/internal/database/repositories"
	"habibi-go/internal/models"
)

type ProjectService struct {
	projectRepo *repositories.ProjectRepository
	eventRepo   *repositories.EventRepository
	gitService  *GitService
}

func NewProjectService(projectRepo *repositories.ProjectRepository, eventRepo *repositories.EventRepository, gitService *GitService) *ProjectService {
	return &ProjectService{
		projectRepo: projectRepo,
		eventRepo:   eventRepo,
		gitService:  gitService,
	}
}

func (s *ProjectService) CreateProject(req *models.CreateProjectRequest) (*models.Project, error) {
	// Validate request
	if req.Name == "" {
		return nil, fmt.Errorf("project name is required")
	}
	
	if req.Path == "" {
		return nil, fmt.Errorf("project path is required")
	}
	
	// Check if project already exists
	exists, err := s.projectRepo.Exists(req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check project existence: %w", err)
	}
	
	if exists {
		return nil, fmt.Errorf("project with name '%s' already exists", req.Name)
	}
	
	// Expand path
	expandedPath, err := expandPath(req.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to expand path: %w", err)
	}
	
	// Create directory if it doesn't exist
	if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
		if err := os.MkdirAll(expandedPath, 0755); err != nil {
			return nil, fmt.Errorf("failed to create project directory: %w", err)
		}
	}
	
	// Check if it's already a git repository
	if !s.isGitRepository(expandedPath) {
		// Initialize git repository only if it's not already one
		if err := s.initializeGitRepo(expandedPath); err != nil {
			return nil, fmt.Errorf("failed to initialize git repository: %w", err)
		}
	}
	
	// Create project
	project := &models.Project{
		Name:          req.Name,
		Path:          expandedPath,
		RepositoryURL: req.RepositoryURL,
		DefaultBranch: req.DefaultBranch,
		SetupCommand:  req.SetupCommand,
		Config:        make(map[string]interface{}),
	}
	
	// Set default branch to "main" if not specified
	if project.DefaultBranch == "" {
		project.DefaultBranch = "main"
	}
	
	if err := project.Validate(); err != nil {
		return nil, fmt.Errorf("project validation failed: %w", err)
	}
	
	if err := s.projectRepo.Create(project); err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}
	
	// Create project event
	event := models.NewProjectEvent(models.EventTypeProjectCreated, project.ID, map[string]interface{}{
		"name": project.Name,
		"path": project.Path,
	})
	
	if err := s.eventRepo.Create(event); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to create project event: %v\n", err)
	}
	
	return project, nil
}

func (s *ProjectService) GetProject(id int) (*models.Project, error) {
	project, err := s.projectRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	
	return project, nil
}

func (s *ProjectService) GetProjectByName(name string) (*models.Project, error) {
	project, err := s.projectRepo.GetByName(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	
	return project, nil
}

func (s *ProjectService) GetAllProjects() ([]*models.Project, error) {
	projects, err := s.projectRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}
	
	// Add current branch information to each project
	for _, project := range projects {
		if currentBranch, err := s.gitService.GetCurrentBranch(project.Path); err == nil {
			// Store in a temporary field for display - we can't modify the struct without migration
			if project.Config == nil {
				project.Config = make(map[string]interface{})
			}
			project.Config["current_branch"] = currentBranch
		}
	}
	
	return projects, nil
}

func (s *ProjectService) UpdateProject(id int, req *models.UpdateProjectRequest) (*models.Project, error) {
	project, err := s.projectRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	
	// Update fields if provided
	if req.Name != "" {
		// Check if new name conflicts with existing projects
		if req.Name != project.Name {
			exists, err := s.projectRepo.Exists(req.Name)
			if err != nil {
				return nil, fmt.Errorf("failed to check project existence: %w", err)
			}
			if exists {
				return nil, fmt.Errorf("project with name '%s' already exists", req.Name)
			}
		}
		project.Name = req.Name
	}
	
	if req.Path != "" {
		expandedPath, err := expandPath(req.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to expand path: %w", err)
		}
		
		if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("project path does not exist: %s", expandedPath)
		}
		
		project.Path = expandedPath
	}
	
	if req.RepositoryURL != "" {
		project.RepositoryURL = req.RepositoryURL
	}
	
	if req.DefaultBranch != "" {
		project.DefaultBranch = req.DefaultBranch
	}
	
	if req.SetupCommand != "" {
		project.SetupCommand = req.SetupCommand
	}
	
	if req.Config != nil {
		project.Config = req.Config
	}
	
	if err := project.Validate(); err != nil {
		return nil, fmt.Errorf("project validation failed: %w", err)
	}
	
	if err := s.projectRepo.Update(project); err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}
	
	// Create project event
	event := models.NewProjectEvent(models.EventTypeProjectUpdated, project.ID, map[string]interface{}{
		"name": project.Name,
		"path": project.Path,
	})
	
	if err := s.eventRepo.Create(event); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to create project event: %v\n", err)
	}
	
	return project, nil
}

func (s *ProjectService) DeleteProject(id int) error {
	project, err := s.projectRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}
	
	if err := s.projectRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}
	
	// Create project event
	event := models.NewProjectEvent(models.EventTypeProjectDeleted, project.ID, map[string]interface{}{
		"name": project.Name,
		"path": project.Path,
	})
	
	if err := s.eventRepo.Create(event); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to create project event: %v\n", err)
	}
	
	return nil
}

func (s *ProjectService) GetProjectWithSessions(id int) (*models.Project, error) {
	project, err := s.projectRepo.GetWithSessions(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get project with sessions: %w", err)
	}
	
	return project, nil
}

func (s *ProjectService) DiscoverProjects(directory string) ([]*models.Project, error) {
	expandedDir, err := expandPath(directory)
	if err != nil {
		return nil, fmt.Errorf("failed to expand directory path: %w", err)
	}
	
	var discoveredProjects []*models.Project
	
	err = filepath.Walk(expandedDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip if not a directory
		if !info.IsDir() {
			return nil
		}
		
		// Check if it's a git repository
		gitPath := filepath.Join(path, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			// Found a git repository
			projectName := filepath.Base(path)
			
			// Check if project already exists
			exists, err := s.projectRepo.Exists(projectName)
			if err != nil {
				return fmt.Errorf("failed to check project existence: %w", err)
			}
			
			if !exists {
				discoveredProjects = append(discoveredProjects, &models.Project{
					Name:          projectName,
					Path:          path,
					DefaultBranch: "main",
					Config:        make(map[string]interface{}),
				})
			}
		}
		
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to discover projects: %w", err)
	}
	
	return discoveredProjects, nil
}

func (s *ProjectService) GetProjectStats() (map[string]interface{}, error) {
	stats, err := s.projectRepo.GetStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get project stats: %w", err)
	}
	
	return stats, nil
}

func expandPath(path string) (string, error) {
	if path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[1:]), nil
	}
	return filepath.Abs(path)
}

// isGitRepository checks if the given path is already a git repository
func (s *ProjectService) isGitRepository(path string) bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = path
	return cmd.Run() == nil
}

// initializeGitRepo initializes a new git repository with initial commit
func (s *ProjectService) initializeGitRepo(path string) error {
	// Initialize git repository
	cmd := exec.Command("git", "init")
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run git init: %w", err)
	}
	
	// Create .gitignore file
	gitignorePath := filepath.Join(path, ".gitignore")
	gitignoreContent := ".habibi-worktrees\n"
	if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
		return fmt.Errorf("failed to create .gitignore: %w", err)
	}
	
	// Add .gitignore to git
	cmd = exec.Command("git", "add", ".gitignore")
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add .gitignore: %w", err)
	}
	
	// Create initial commit
	cmd = exec.Command("git", "commit", "-m", "Initial commit with .gitignore")
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create initial commit: %w", err)
	}
	
	return nil
}

// RunProjectStartupScript runs the project's startup script in the main project directory
func (s *ProjectService) RunProjectStartupScript(id int) (string, error) {
	project, err := s.projectRepo.GetByID(id)
	if err != nil {
		return "", fmt.Errorf("failed to get project: %w", err)
	}

	if project.SetupCommand == "" {
		return "", fmt.Errorf("no startup script configured for this project")
	}

	// Run the setup command in the project directory
	cmd := exec.Command("sh", "-c", project.SetupCommand)
	cmd.Dir = project.Path
	
	// Set up environment variables
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("PROJECT_PATH=%s", project.Path))
	
	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		return string(outputBytes), fmt.Errorf("setup command failed: %w", err)
	}

	// Create event for running startup script
	event := models.NewProjectEvent("project_ran_startup_script", project.ID, map[string]interface{}{
		"success": true,
		"output":  string(outputBytes),
	})

	if err := s.eventRepo.Create(event); err != nil {
		fmt.Printf("Failed to create startup script event: %v\n", err)
	}

	return string(outputBytes), nil
}