package models

import "time"

// Agent represents a Claude agent instance (virtual, not a process)
type Agent struct {
	ID           int       `json:"id"`
	Status       string    `json:"status"`
	ClaudeModel  string    `json:"claude_model"`
	SessionCount int       `json:"session_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}