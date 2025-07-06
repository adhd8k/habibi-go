package models

import (
	"encoding/json"
	"fmt"
	"time"
)

type Project struct {
	ID            int                    `json:"id" db:"id"`
	Name          string                 `json:"name" db:"name" binding:"required"`
	Path          string                 `json:"path" db:"path" binding:"required"`
	RepositoryURL string                 `json:"repository_url" db:"repository_url"`
	DefaultBranch string                 `json:"default_branch" db:"default_branch"`
	SetupCommand  string                 `json:"setup_command" db:"setup_command"`
	Config        map[string]interface{} `json:"config" db:"config"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at" db:"updated_at"`
	
	// Relationships
	Sessions []Session `json:"sessions,omitempty"`
}

type ProjectConfig struct {
	GitRemote     string            `json:"git_remote,omitempty"`
	AgentDefaults map[string]string `json:"agent_defaults,omitempty"`
	Notifications bool              `json:"notifications,omitempty"`
}

type CreateProjectRequest struct {
	Name          string `json:"name" binding:"required"`
	Path          string `json:"path" binding:"required"`
	RepositoryURL string `json:"repository_url"`
	DefaultBranch string `json:"default_branch"`
	SetupCommand  string `json:"setup_command"`
}

type UpdateProjectRequest struct {
	Name          string                 `json:"name"`
	Path          string                 `json:"path"`
	RepositoryURL string                 `json:"repository_url"`
	DefaultBranch string                 `json:"default_branch"`
	SetupCommand  string                 `json:"setup_command"`
	Config        map[string]interface{} `json:"config"`
}

func (p *Project) MarshalConfig() (string, error) {
	if p.Config == nil {
		return "{}", nil
	}
	data, err := json.Marshal(p.Config)
	if err != nil {
		return "", fmt.Errorf("failed to marshal project config: %w", err)
	}
	return string(data), nil
}

func (p *Project) UnmarshalConfig(configStr string) error {
	if configStr == "" {
		p.Config = make(map[string]interface{})
		return nil
	}
	
	if err := json.Unmarshal([]byte(configStr), &p.Config); err != nil {
		return fmt.Errorf("failed to unmarshal project config: %w", err)
	}
	return nil
}

func (p *Project) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("project name is required")
	}
	
	if p.Path == "" {
		return fmt.Errorf("project path is required")
	}
	
	if p.DefaultBranch == "" {
		p.DefaultBranch = "main"
	}
	
	return nil
}

func (p *Project) BeforeCreate() {
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	
	if p.Config == nil {
		p.Config = make(map[string]interface{})
	}
}

func (p *Project) BeforeUpdate() {
	p.UpdatedAt = time.Now()
}