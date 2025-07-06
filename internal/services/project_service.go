package services

import (
	"fmt"
	"os"
	"path/filepath"

	"habibi-go/internal/database/repositories"
	"habibi-go/internal/models"
)

type ProjectService struct {
	projectRepo *repositories.ProjectRepository
	eventRepo   *repositories.EventRepository
}

func NewProjectService(projectRepo *repositories.ProjectRepository, eventRepo *repositories.EventRepository) *ProjectService {
	return &ProjectService{
		projectRepo: projectRepo,
		eventRepo:   eventRepo,
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
	
	// Validate path exists
	if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("project path does not exist: %s", expandedPath)
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