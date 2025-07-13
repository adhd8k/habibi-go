package models

import (
	"time"
)

type ChatMessage struct {
	ID          int         `json:"id" db:"id"`
	AgentID     int         `json:"agent_id" db:"agent_id"`     // Deprecated - will be removed
	SessionID   int         `json:"session_id" db:"session_id"` // New field
	Role        string      `json:"role" db:"role"`
	Content     string      `json:"content" db:"content"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	ToolName    string      `json:"tool_name,omitempty" db:"tool_name"`
	ToolInput   interface{} `json:"tool_input,omitempty" db:"tool_input"`
	ToolUseID   string      `json:"tool_use_id,omitempty" db:"tool_use_id"`
	ToolContent interface{} `json:"tool_content,omitempty" db:"tool_content"`
}

type CreateChatMessageRequest struct {
	AgentID int    `json:"agent_id" binding:"required"`
	Role    string `json:"role" binding:"required,oneof=user assistant system tool_use tool_result"`
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

func NewToolUseMessage(agentID int, toolName, toolUseID string, toolInput interface{}) *ChatMessage {
	return &ChatMessage{
		AgentID:   agentID,
		Role:      "tool_use",
		Content:   "",
		ToolName:  toolName,
		ToolInput: toolInput,
		ToolUseID: toolUseID,
		CreatedAt: time.Now(),
	}
}

func NewToolResultMessage(agentID int, toolUseID string, toolContent interface{}) *ChatMessage {
	return &ChatMessage{
		AgentID:     agentID,
		Role:        "tool_result",
		Content:     "",
		ToolUseID:   toolUseID,
		ToolContent: toolContent,
		CreatedAt:   time.Now(),
	}
}