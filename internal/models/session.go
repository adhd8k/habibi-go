package models

import (
	"encoding/json"
	"fmt"
	"time"
)

type Session struct {
	ID               int                    `json:"id" db:"id"`
	ProjectID        int                    `json:"project_id" db:"project_id"`
	Name             string                 `json:"name" db:"name" binding:"required"`
	BranchName       string                 `json:"branch_name" db:"branch_name" binding:"required"`
	OriginalBranch   string                 `json:"original_branch" db:"original_branch"`
	WorktreePath     string                 `json:"worktree_path" db:"worktree_path"`
	Status           string                 `json:"status" db:"status"`
	Config           map[string]interface{} `json:"config" db:"config"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
	LastUsedAt       time.Time              `json:"last_used_at" db:"last_used_at"`
	LastActivityAt   *time.Time             `json:"last_activity_at" db:"last_activity_at"`
	ActivityStatus   string                 `json:"activity_status" db:"activity_status"`
	LastViewedAt     *time.Time             `json:"last_viewed_at" db:"last_viewed_at"`
	
	// Relationships
	Project *Project `json:"project,omitempty"`
	Agents  []Agent  `json:"agents,omitempty"`
}

type SessionStatus string

const (
	SessionStatusActive  SessionStatus = "active"
	SessionStatusPaused  SessionStatus = "paused"
	SessionStatusStopped SessionStatus = "stopped"
)

type SessionActivityStatus string

const (
	ActivityStatusIdle       SessionActivityStatus = "idle"       // No recent activity
	ActivityStatusStreaming  SessionActivityStatus = "streaming"  // Currently receiving Claude response  
	ActivityStatusNewResponse SessionActivityStatus = "new"       // New response since last view
	ActivityStatusViewed     SessionActivityStatus = "viewed"     // Response has been viewed
)

type CreateSessionRequest struct {
	ProjectID  int    `json:"project_id" binding:"required"`
	Name       string `json:"name" binding:"required"`
	BranchName string `json:"branch_name" binding:"required"`
}

type UpdateSessionRequest struct {
	Name       string                 `json:"name"`
	BranchName string                 `json:"branch_name"`
	Status     string                 `json:"status"`
	Config     map[string]interface{} `json:"config"`
}

func (s *Session) MarshalConfig() (string, error) {
	if s.Config == nil {
		return "{}", nil
	}
	data, err := json.Marshal(s.Config)
	if err != nil {
		return "", fmt.Errorf("failed to marshal session config: %w", err)
	}
	return string(data), nil
}

func (s *Session) UnmarshalConfig(configStr string) error {
	if configStr == "" {
		s.Config = make(map[string]interface{})
		return nil
	}
	
	if err := json.Unmarshal([]byte(configStr), &s.Config); err != nil {
		return fmt.Errorf("failed to unmarshal session config: %w", err)
	}
	return nil
}

func (s *Session) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("session name is required")
	}
	
	if s.BranchName == "" {
		return fmt.Errorf("branch name is required")
	}
	
	if s.ProjectID == 0 {
		return fmt.Errorf("project ID is required")
	}
	
	if s.Status == "" {
		s.Status = string(SessionStatusActive)
	}
	
	if !s.IsValidStatus() {
		return fmt.Errorf("invalid session status: %s", s.Status)
	}
	
	return nil
}

func (s *Session) IsValidStatus() bool {
	switch SessionStatus(s.Status) {
	case SessionStatusActive, SessionStatusPaused, SessionStatusStopped:
		return true
	default:
		return false
	}
}

func (s *Session) BeforeCreate() {
	s.CreatedAt = time.Now()
	s.LastUsedAt = time.Now()
	
	if s.Config == nil {
		s.Config = make(map[string]interface{})
	}
	
	if s.Status == "" {
		s.Status = string(SessionStatusActive)
	}
}

func (s *Session) BeforeUpdate() {
	s.LastUsedAt = time.Now()
}

func (s *Session) UpdateLastUsed() {
	s.LastUsedAt = time.Now()
}

func (s *Session) IsActive() bool {
	return s.Status == string(SessionStatusActive)
}

func (s *Session) Activate() {
	s.Status = string(SessionStatusActive)
	s.UpdateLastUsed()
}

func (s *Session) Pause() {
	s.Status = string(SessionStatusPaused)
}

func (s *Session) Stop() {
	s.Status = string(SessionStatusStopped)
}