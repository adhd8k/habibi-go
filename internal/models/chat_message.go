package models

import (
	"time"
)

type ChatMessage struct {
	ID        int       `json:"id" db:"id"`
	AgentID   int       `json:"agent_id" db:"agent_id"`
	Role      string    `json:"role" db:"role"`
	Content   string    `json:"content" db:"content"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type CreateChatMessageRequest struct {
	AgentID int    `json:"agent_id" binding:"required"`
	Role    string `json:"role" binding:"required,oneof=user assistant system"`
	Content string `json:"content" binding:"required"`
}

func NewChatMessage(agentID int, role, content string) *ChatMessage {
	return &ChatMessage{
		AgentID:   agentID,
		Role:      role,
		Content:   content,
		CreatedAt: time.Now(),
	}
}