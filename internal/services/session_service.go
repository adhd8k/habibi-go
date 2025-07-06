package services

import (
	"fmt"
	"os"

	"habibi-go/internal/database/repositories"
	"habibi-go/internal/models"
)

type SessionService struct {
	sessionRepo  *repositories.SessionRepository
	projectRepo  *repositories.ProjectRepository
	eventRepo    *repositories.EventRepository
	gitService   *GitService
}

func NewSessionService(
	sessionRepo *repositories.SessionRepository,
	projectRepo *repositories.ProjectRepository,
	eventRepo *repositories.EventRepository,
	gitService *GitService,
) *SessionService {
	return &SessionService{
		sessionRepo: sessionRepo,
		projectRepo: projectRepo,
		eventRepo:   eventRepo,
		gitService:  gitService,
	}
}

func (s *SessionService) CreateSession(req *models.CreateSessionRequest) (*models.Session, error) {
	// Validate request
	if req.Name == "" {
		return nil, fmt.Errorf("session name is required")
	}
	
	if req.BranchName == "" {
		return nil, fmt.Errorf("branch name is required")
	}
	
	if req.ProjectID == 0 {
		return nil, fmt.Errorf("project ID is required")
	}
	
	// Get project to validate it exists
	project, err := s.projectRepo.GetByID(req.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	
	// Check if session already exists
	exists, err := s.sessionRepo.Exists(req.ProjectID, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check session existence: %w", err)
	}
	
	if exists {
		return nil, fmt.Errorf("session with name '%s' already exists in project '%s'", req.Name, project.Name)
	}
	
	// Validate that project path is a Git repository
	if err := s.gitService.gitUtil.ValidateRepository(project.Path); err != nil {
		return nil, fmt.Errorf("project is not a valid Git repository: %w", err)
	}
	
	// Create worktree
	worktreePath, err := s.gitService.CreateWorktree(project.Path, req.Name, req.BranchName)
	if err != nil {
		return nil, fmt.Errorf("failed to create worktree: %w", err)
	}
	
	// Create session
	session := &models.Session{
		ProjectID:    req.ProjectID,
		Name:         req.Name,
		BranchName:   req.BranchName,
		WorktreePath: worktreePath,
		Status:       string(models.SessionStatusActive),
		Config:       make(map[string]interface{}),
	}
	
	if err := session.Validate(); err != nil {
		// Clean up worktree if session validation fails
		s.gitService.RemoveWorktree(project.Path, worktreePath)
		return nil, fmt.Errorf("session validation failed: %w", err)
	}
	
	if err := s.sessionRepo.Create(session); err != nil {
		// Clean up worktree if database creation fails
		s.gitService.RemoveWorktree(project.Path, worktreePath)
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	
	// Create session event
	event := models.NewSessionEvent(models.EventTypeSessionCreated, session.ID, map[string]interface{}{
		"name":          session.Name,
		"branch_name":   session.BranchName,
		"worktree_path": session.WorktreePath,
		"project_id":    session.ProjectID,
	})
	
	if err := s.eventRepo.Create(event); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to create session event: %v\n", err)
	}
	
	return session, nil
}

func (s *SessionService) GetSession(id int) (*models.Session, error) {
	session, err := s.sessionRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	
	return session, nil
}

func (s *SessionService) GetSessionByProjectAndName(projectID int, name string) (*models.Session, error) {
	session, err := s.sessionRepo.GetByProjectAndName(projectID, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	
	return session, nil
}

func (s *SessionService) GetSessionsByProject(projectID int) ([]*models.Session, error) {
	sessions, err := s.sessionRepo.GetByProjectID(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}
	
	return sessions, nil
}

func (s *SessionService) GetAllSessions() ([]*models.Session, error) {
	sessions, err := s.sessionRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}
	
	return sessions, nil
}

func (s *SessionService) UpdateSession(id int, req *models.UpdateSessionRequest) (*models.Session, error) {
	session, err := s.sessionRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	
	// Get project for validation
	_, err = s.projectRepo.GetByID(session.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	
	// Update fields if provided
	updated := false
	
	if req.Name != "" && req.Name != session.Name {
		// Check if new name conflicts with existing sessions
		exists, err := s.sessionRepo.Exists(session.ProjectID, req.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to check session existence: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("session with name '%s' already exists", req.Name)
		}
		session.Name = req.Name
		updated = true
	}
	
	if req.BranchName != "" && req.BranchName != session.BranchName {
		// Switch to new branch in worktree
		if err := s.gitService.SwitchBranch(session.WorktreePath, req.BranchName); err != nil {
			return nil, fmt.Errorf("failed to switch branch: %w", err)
		}
		session.BranchName = req.BranchName
		updated = true
	}
	
	if req.Status != "" && req.Status != session.Status {
		// Validate new status
		oldStatus := session.Status
		session.Status = req.Status
		if !session.IsValidStatus() {
			return nil, fmt.Errorf("invalid session status: %s", req.Status)
		}
		
		// Create status change event
		eventType := models.EventTypeSessionUpdated
		switch models.SessionStatus(req.Status) {
		case models.SessionStatusActive:
			eventType = models.EventTypeSessionActivated
		case models.SessionStatusPaused:
			eventType = models.EventTypeSessionPaused
		case models.SessionStatusStopped:
			eventType = models.EventTypeSessionStopped
		}
		
		event := models.NewSessionEvent(eventType, session.ID, map[string]interface{}{
			"old_status": oldStatus,
			"new_status": req.Status,
		})
		
		if err := s.eventRepo.Create(event); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("Failed to create session status event: %v\n", err)
		}
		
		updated = true
	}
	
	if req.Config != nil {
		session.Config = req.Config
		updated = true
	}
	
	if updated {
		if err := session.Validate(); err != nil {
			return nil, fmt.Errorf("session validation failed: %w", err)
		}
		
		if err := s.sessionRepo.Update(session); err != nil {
			return nil, fmt.Errorf("failed to update session: %w", err)
		}
		
		// Create general update event if not already created above
		if req.Status == "" {
			event := models.NewSessionEvent(models.EventTypeSessionUpdated, session.ID, map[string]interface{}{
				"updated_fields": getUpdatedFields(req),
			})
			
			if err := s.eventRepo.Create(event); err != nil {
				// Log error but don't fail the operation
				fmt.Printf("Failed to create session event: %v\n", err)
			}
		}
	}
	
	return session, nil
}

func (s *SessionService) DeleteSession(id int) error {
	session, err := s.sessionRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}
	
	// Get project for worktree cleanup
	project, err := s.projectRepo.GetByID(session.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}
	
	// Remove worktree
	if err := s.gitService.RemoveWorktree(project.Path, session.WorktreePath); err != nil {
		// Log warning but continue with deletion
		fmt.Printf("Warning: failed to remove worktree: %v\n", err)
	}
	
	// Delete session from database
	if err := s.sessionRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	
	// Create deletion event
	event := models.NewSessionEvent(models.EventTypeSessionDeleted, session.ID, map[string]interface{}{
		"name":          session.Name,
		"branch_name":   session.BranchName,
		"worktree_path": session.WorktreePath,
		"project_id":    session.ProjectID,
	})
	
	if err := s.eventRepo.Create(event); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to create session event: %v\n", err)
	}
	
	return nil
}

func (s *SessionService) ActivateSession(id int) (*models.Session, error) {
	session, err := s.sessionRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	
	// Update session status and last used timestamp
	session.Activate()
	
	if err := s.sessionRepo.Update(session); err != nil {
		return nil, fmt.Errorf("failed to activate session: %w", err)
	}
	
	// Create activation event
	event := models.NewSessionEvent(models.EventTypeSessionActivated, session.ID, map[string]interface{}{
		"name":          session.Name,
		"branch_name":   session.BranchName,
		"worktree_path": session.WorktreePath,
	})
	
	if err := s.eventRepo.Create(event); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to create session event: %v\n", err)
	}
	
	return session, nil
}

func (s *SessionService) GetSessionStatus(id int) (*SessionStatusInfo, error) {
	session, err := s.sessionRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	
	// Get worktree status
	worktreeStatus, err := s.gitService.GetWorktreeStatus(session.WorktreePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree status: %w", err)
	}
	
	// Check if worktree path exists
	worktreeExists := true
	if _, err := os.Stat(session.WorktreePath); os.IsNotExist(err) {
		worktreeExists = false
	}
	
	return &SessionStatusInfo{
		Session:         session,
		WorktreeStatus:  worktreeStatus,
		WorktreeExists:  worktreeExists,
	}, nil
}

type SessionStatusInfo struct {
	Session        *models.Session         `json:"session"`
	WorktreeStatus *WorktreeStatus         `json:"worktree_status"`
	WorktreeExists bool                    `json:"worktree_exists"`
}

func (s *SessionService) SyncSessionBranch(id int) error {
	session, err := s.sessionRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}
	
	// Pull latest changes
	if err := s.gitService.PullLatest(session.WorktreePath); err != nil {
		return fmt.Errorf("failed to sync branch: %w", err)
	}
	
	// Update last used timestamp
	if err := s.sessionRepo.UpdateLastUsed(session.ID); err != nil {
		return fmt.Errorf("failed to update last used: %w", err)
	}
	
	return nil
}

func (s *SessionService) CleanupStoppedSessions(projectID int) error {
	// Get stopped sessions
	stoppedSessions, err := s.sessionRepo.GetByStatus(string(models.SessionStatusStopped))
	if err != nil {
		return fmt.Errorf("failed to get stopped sessions: %w", err)
	}
	
	// Filter by project if specified
	var sessionsToClean []*models.Session
	for _, session := range stoppedSessions {
		if projectID == 0 || session.ProjectID == projectID {
			sessionsToClean = append(sessionsToClean, session)
		}
	}
	
	// Get project for worktree cleanup
	var project *models.Project
	if projectID != 0 {
		project, err = s.projectRepo.GetByID(projectID)
		if err != nil {
			return fmt.Errorf("failed to get project: %w", err)
		}
	}
	
	// Clean up each session
	for _, session := range sessionsToClean {
		if project == nil {
			// Get project for this session
			sessionProject, err := s.projectRepo.GetByID(session.ProjectID)
			if err != nil {
				fmt.Printf("Warning: failed to get project for session %s: %v\n", session.Name, err)
				continue
			}
			project = sessionProject
		}
		
		// Remove worktree
		if err := s.gitService.RemoveWorktree(project.Path, session.WorktreePath); err != nil {
			fmt.Printf("Warning: failed to remove worktree for session %s: %v\n", session.Name, err)
		}
		
		// Delete session
		if err := s.sessionRepo.Delete(session.ID); err != nil {
			fmt.Printf("Warning: failed to delete session %s: %v\n", session.Name, err)
		} else {
			fmt.Printf("Cleaned up session: %s\n", session.Name)
		}
	}
	
	return nil
}

func (s *SessionService) GetSessionStats() (map[string]interface{}, error) {
	stats, err := s.sessionRepo.GetStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get session stats: %w", err)
	}
	
	return stats, nil
}

// Helper function to determine what fields were updated
func getUpdatedFields(req *models.UpdateSessionRequest) []string {
	var fields []string
	
	if req.Name != "" {
		fields = append(fields, "name")
	}
	if req.BranchName != "" {
		fields = append(fields, "branch_name")
	}
	if req.Status != "" {
		fields = append(fields, "status")
	}
	if req.Config != nil {
		fields = append(fields, "config")
	}
	
	return fields
}