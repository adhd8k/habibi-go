package services

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"habibi-go/internal/database/repositories"
	"habibi-go/internal/models"
)

// ClaudeSessionService handles Claude operations directly on sessions
type ClaudeSessionService struct {
	sessionRepo      *repositories.SessionRepository
	projectRepo      *repositories.ProjectRepository
	chatRepo         *repositories.ChatMessageV2Repository
	eventRepo        *repositories.EventRepository
	claudeBinaryPath string
	eventBroadcaster EventBroadcaster
}

// NewClaudeSessionService creates a new Claude session service
func NewClaudeSessionService(
	sessionRepo *repositories.SessionRepository,
	projectRepo *repositories.ProjectRepository,
	chatRepo *repositories.ChatMessageV2Repository,
	eventRepo *repositories.EventRepository,
	claudeBinaryPath string,
) *ClaudeSessionService {
	return &ClaudeSessionService{
		sessionRepo:      sessionRepo,
		projectRepo:      projectRepo,
		chatRepo:         chatRepo,
		eventRepo:        eventRepo,
		claudeBinaryPath: claudeBinaryPath,
		eventBroadcaster: &NoOpBroadcaster{},
	}
}

// SetEventBroadcaster sets the event broadcaster
func (s *ClaudeSessionService) SetEventBroadcaster(broadcaster EventBroadcaster) {
	s.eventBroadcaster = broadcaster
}

// SendMessage sends a message to Claude for a session
func (s *ClaudeSessionService) SendMessage(sessionID int, message string) error {
	// Get session
	session, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Save user message
	userMsg := &models.ChatMessage{
		SessionID: sessionID,
		Role:      "user",
		Content:   message,
	}
	if err := s.chatRepo.Create(userMsg); err != nil {
		return fmt.Errorf("failed to save user message: %w", err)
	}

	// Broadcast the user message
	s.eventBroadcaster.BroadcastEvent("new_chat_message", 0, map[string]interface{}{
		"session_id": sessionID,
		"message":    userMsg,
	})

	// Update session activity
	if err := s.sessionRepo.UpdateActivityStatus(sessionID, string(models.ActivityStatusStreaming)); err != nil {
		fmt.Printf("Failed to update session activity status: %v\n", err)
	}

	// Execute Claude command
	go s.executeClaudeCommand(session, message)

	return nil
}

// executeClaudeCommand runs Claude in the session's worktree
func (s *ClaudeSessionService) executeClaudeCommand(session *models.Session, message string) {
	// Prepare Claude command
	claudePath := s.claudeBinaryPath
	if claudePath == "" {
		claudePath = "claude"
	}

	// Build command with -c flag to continue conversation in this directory
	args := []string{"-c", message}
	cmd := exec.Command(claudePath, args...)
	cmd.Dir = session.WorktreePath

	// Get stdout pipe
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		s.handleError(session.ID, fmt.Errorf("failed to get stdout pipe: %w", err))
		return
	}

	// Get stderr pipe for debugging
	stderr, err := cmd.StderrPipe()
	if err != nil {
		s.handleError(session.ID, fmt.Errorf("failed to get stderr pipe: %w", err))
		return
	}

	// Start command
	if err := cmd.Start(); err != nil {
		s.handleError(session.ID, fmt.Errorf("failed to start Claude: %w", err))
		return
	}

	// Read stderr in background for debugging
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Printf("Claude stderr: %s\n", line)
		}
	}()

	// Process stdout
	scanner := bufio.NewScanner(stdout)
	var assistantMessage strings.Builder
	assistantMessageID := 0

	for scanner.Scan() {
		line := scanner.Text()
		
		// Try to parse as JSON (stream-json format)
		var streamMsg map[string]interface{}
		if err := json.Unmarshal([]byte(line), &streamMsg); err == nil {
			s.handleStreamMessage(session.ID, streamMsg, &assistantMessage, &assistantMessageID)
		} else {
			// If not JSON, treat as plain text output
			assistantMessage.WriteString(line)
			assistantMessage.WriteString("\n")
			
			// Broadcast the chunk
			s.eventBroadcaster.BroadcastEvent("claude_output", 0, map[string]interface{}{
				"session_id":   session.ID,
				"content_type": "text",
				"output":       line + "\n",
				"is_chunk":     true,
			})
		}
	}

	// Wait for command to complete
	if err := cmd.Wait(); err != nil {
		s.handleError(session.ID, fmt.Errorf("Claude command failed: %w", err))
		return
	}

	// Save final assistant message if we have content
	if assistantMessage.Len() > 0 {
		msg := &models.ChatMessage{
			SessionID: session.ID,
			Role:      "assistant",
			Content:   assistantMessage.String(),
		}
		if err := s.chatRepo.Create(msg); err != nil {
			fmt.Printf("Failed to save assistant message: %v\n", err)
		}
	}

	// Update session activity
	if err := s.sessionRepo.UpdateActivityStatus(session.ID, string(models.ActivityStatusNewResponse)); err != nil {
		fmt.Printf("Failed to update session activity status: %v\n", err)
	}

	// Broadcast completion
	s.eventBroadcaster.BroadcastEvent("claude_response_complete", 0, map[string]interface{}{
		"session_id": session.ID,
	})
}

// handleStreamMessage processes a JSON message from Claude's stream
func (s *ClaudeSessionService) handleStreamMessage(sessionID int, msg map[string]interface{}, assistantMessage *strings.Builder, assistantMessageID *int) {
	msgType, _ := msg["type"].(string)
	
	switch msgType {
	case "message_start":
		// New message started
		if message, ok := msg["message"].(map[string]interface{}); ok {
			if _, ok := message["id"].(string); ok {
				// Create new assistant message
				newMsg := &models.ChatMessage{
					SessionID: sessionID,
					Role:      "assistant",
					Content:   "",
				}
				if err := s.chatRepo.Create(newMsg); err == nil {
					*assistantMessageID = newMsg.ID
				}
			}
		}

	case "content_block_delta":
		// Text content
		if delta, ok := msg["delta"].(map[string]interface{}); ok {
			if text, ok := delta["text"].(string); ok {
				assistantMessage.WriteString(text)
				
				// Broadcast the chunk
				s.eventBroadcaster.BroadcastEvent("claude_output", 0, map[string]interface{}{
					"session_id":     sessionID,
					"content_type":   "text",
					"output":         text,
					"is_chunk":       true,
					"db_message_id":  *assistantMessageID,
				})
			}
		}

	case "tool_use":
		// Tool use message
		if toolName, ok := msg["name"].(string); ok {
			toolUseID, _ := msg["id"].(string)
			toolInput := msg["input"]
			
			// Save tool use message
			toolMsg := &models.ChatMessage{
				SessionID:   sessionID,
				Role:        "tool_use",
				Content:     "",
				ToolName:    toolName,
				ToolInput:   toolInput,
				ToolUseID:   toolUseID,
			}
			if err := s.chatRepo.Create(toolMsg); err == nil {
				// Broadcast tool use
				s.eventBroadcaster.BroadcastEvent("claude_output", 0, map[string]interface{}{
					"session_id":     sessionID,
					"content_type":   "tool_use",
					"tool_name":      toolName,
					"tool_input":     toolInput,
					"tool_use_id":    toolUseID,
					"db_message_id":  toolMsg.ID,
				})
			}
		}

	case "tool_result":
		// Tool result message
		if toolUseID, ok := msg["tool_use_id"].(string); ok {
			content := msg["content"]
			
			// Save tool result message
			toolMsg := &models.ChatMessage{
				SessionID:   sessionID,
				Role:        "tool_result",
				Content:     "",
				ToolUseID:   toolUseID,
				ToolContent: content,
			}
			if err := s.chatRepo.Create(toolMsg); err == nil {
				// Broadcast tool result
				s.eventBroadcaster.BroadcastEvent("claude_output", 0, map[string]interface{}{
					"session_id":     sessionID,
					"content_type":   "tool_result",
					"tool_use_id":    toolUseID,
					"tool_content":   content,
					"db_message_id":  toolMsg.ID,
				})
			}
		}
	}
}

// handleError handles errors during Claude execution
func (s *ClaudeSessionService) handleError(sessionID int, err error) {
	fmt.Printf("Claude error for session %d: %v\n", sessionID, err)
	
	// Update session status
	if err := s.sessionRepo.UpdateActivityStatus(sessionID, string(models.ActivityStatusIdle)); err != nil {
		fmt.Printf("Failed to update session status: %v\n", err)
	}
	
	// Broadcast error
	s.eventBroadcaster.BroadcastEvent("claude_error", 0, map[string]interface{}{
		"session_id": sessionID,
		"error":      err.Error(),
	})
}

// GetChatHistory retrieves chat history for a session
func (s *ClaudeSessionService) GetChatHistory(sessionID int, limit int) ([]*models.ChatMessage, error) {
	return s.chatRepo.GetBySessionID(sessionID, limit)
}