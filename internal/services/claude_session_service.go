package services

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"sync"

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
	runningProcesses map[int]*exec.Cmd
	processMutex     sync.Mutex
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
		runningProcesses: make(map[int]*exec.Cmd),
	}
}

// SetEventBroadcaster sets the event broadcaster
func (s *ClaudeSessionService) SetEventBroadcaster(broadcaster EventBroadcaster) {
	s.eventBroadcaster = broadcaster
}

// ClearChatHistory clears the chat history for a session
func (s *ClaudeSessionService) ClearChatHistory(ctx context.Context, sessionID int) error {
	return s.chatRepo.DeleteBySessionID(sessionID)
}

// GetAgent returns agent information for a given agent ID
func (s *ClaudeSessionService) GetAgent(ctx context.Context, agentID int) (*models.Agent, error) {
	// For now, return a mock agent since agents are virtual
	return &models.Agent{
		ID:           agentID,
		Status:       "active",
		ClaudeModel:  "claude-3-opus-20240229",
		SessionCount: 1,
	}, nil
}

// SendMessage sends a message to Claude for a session
func (s *ClaudeSessionService) SendMessage(ctx context.Context, sessionID int, message string) error {
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
	fmt.Printf("Saved user message with ID: %d for session: %d\n", userMsg.ID, sessionID)

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
	go s.executeClaudeCommand(session.ID, session.WorktreePath, message)

	return nil
}

// executeClaudeCommand runs Claude in the session's worktree
func (s *ClaudeSessionService) executeClaudeCommand(sessionID int, worktreePath string, message string) {
	// Prepare Claude command
	claudePath := s.claudeBinaryPath
	if claudePath == "" {
		claudePath = "claude"
	}

	// Build command with -c flag to continue conversation in this directory
	// Note: message should come after -c flag, and --verbose is required for proper output
	args := []string{"--verbose", "--output-format", "stream-json", "--dangerously-skip-permissions", "-c", message}
	cmd := exec.Command(claudePath, args...)
	cmd.Dir = worktreePath

	// Get stdout pipe
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		s.handleError(sessionID, fmt.Errorf("failed to get stdout pipe: %w", err))
		return
	}

	// Get stderr pipe for debugging
	stderr, err := cmd.StderrPipe()
	if err != nil {
		s.handleError(sessionID, fmt.Errorf("failed to get stderr pipe: %w", err))
		return
	}

	// Start command
	fmt.Printf("Starting Claude command: %s %v in directory: %s\n", claudePath, args, worktreePath)
	if err := cmd.Start(); err != nil {
		s.handleError(sessionID, fmt.Errorf("failed to start Claude: %w", err))
		return
	}
	fmt.Printf("Claude command started successfully for session %d\n", sessionID)

	// Track the running process
	s.processMutex.Lock()
	s.runningProcesses[sessionID] = cmd
	s.processMutex.Unlock()

	// Ensure we clean up the process tracking when done
	defer func() {
		s.processMutex.Lock()
		delete(s.runningProcesses, sessionID)
		s.processMutex.Unlock()
	}()

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

	fmt.Printf("Starting to read Claude output for session %d\n", sessionID)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Printf("Claude stdout line: %s\n", line)

		// Try to parse as JSON (stream-json format)
		var streamMsg map[string]interface{}
		if err := json.Unmarshal([]byte(line), &streamMsg); err == nil {
			msgType, _ := streamMsg["type"].(string)
			fmt.Printf("Parsed JSON stream message type=%v\n", msgType)

			switch msgType {
			case "assistant":
				// Handle assistant messages - each message is independent
				messageID := 0
				s.handleAssistantMessage(sessionID, streamMsg, &messageID)
			case "user":
				// Handle tool results
				s.handleToolResultMessage(sessionID, streamMsg)
			case "system", "result":
				// Log but don't process
				fmt.Printf("System/Result message: %+v\n", streamMsg)
			default:
				// Try old format handler as fallback
				s.handleStreamMessage(sessionID, streamMsg, &assistantMessage, &assistantMessageID)
			}
		} else {
			fmt.Printf("Failed to parse as JSON (error: %v), treating as plain text: %s\n", err, line)
			// If not JSON, treat as plain text output
			assistantMessage.WriteString(line)
			assistantMessage.WriteString("\n")

			// Broadcast the chunk
			s.eventBroadcaster.BroadcastEvent("claude_output", 0, map[string]interface{}{
				"session_id":   sessionID,
				"content_type": "text",
				"output":       line + "\n",
				"is_chunk":     true,
			})
		}
	}
	fmt.Printf("Finished reading Claude output for session %d\n", sessionID)

	// Wait for command to complete
	if err := cmd.Wait(); err != nil {
		s.handleError(sessionID, fmt.Errorf("Claude command failed: %w", err))
		return
	}

	// Since messages are now saved as they arrive, we don't need to do final saving
	// Just log the completion
	fmt.Printf("Claude command completed. Assistant message ID: %d\n", assistantMessageID)

	// Update session activity
	if err := s.sessionRepo.UpdateActivityStatus(sessionID, string(models.ActivityStatusNewResponse)); err != nil {
		fmt.Printf("Failed to update session activity status: %v\n", err)
	}

	// Broadcast completion
	s.eventBroadcaster.BroadcastEvent("claude_response_complete", 0, map[string]interface{}{
		"session_id": sessionID,
	})
}

// handleStreamMessage processes a JSON message from Claude's stream
func (s *ClaudeSessionService) handleStreamMessage(sessionID int, msg map[string]interface{}, assistantMessage *strings.Builder, assistantMessageID *int) {
	msgType, _ := msg["type"].(string)

	switch msgType {
	case "message_start":
		// New message started
		fmt.Printf("message_start: creating new assistant message for session %d\n", sessionID)
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
					fmt.Printf("Created assistant message with ID: %d\n", newMsg.ID)
				} else {
					fmt.Printf("Failed to create assistant message: %v\n", err)
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
					"session_id":    sessionID,
					"content_type":  "text",
					"output":        text,
					"is_chunk":      true,
					"db_message_id": *assistantMessageID,
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
				SessionID: sessionID,
				Role:      "tool_use",
				Content:   "",
				ToolName:  toolName,
				ToolInput: toolInput,
				ToolUseID: toolUseID,
			}
			if err := s.chatRepo.Create(toolMsg); err == nil {
				// Broadcast tool use
				s.eventBroadcaster.BroadcastEvent("claude_output", 0, map[string]interface{}{
					"session_id":    sessionID,
					"content_type":  "tool_use",
					"tool_name":     toolName,
					"tool_input":    toolInput,
					"tool_use_id":   toolUseID,
					"db_message_id": toolMsg.ID,
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
					"session_id":    sessionID,
					"content_type":  "tool_result",
					"tool_use_id":   toolUseID,
					"tool_content":  content,
					"db_message_id": toolMsg.ID,
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

// handleAssistantMessage handles assistant messages in the new format
func (s *ClaudeSessionService) handleAssistantMessage(sessionID int, msg map[string]interface{}, assistantMessageID *int) {
	message, ok := msg["message"].(map[string]interface{})
	if !ok {
		fmt.Printf("No message field in assistant message\n")
		return
	}

	content, ok := message["content"].([]interface{})
	if !ok {
		fmt.Printf("No content field in assistant message\n")
		return
	}

	// Process each content block
	for _, block := range content {
		blockMap, ok := block.(map[string]interface{})
		if !ok {
			continue
		}

		blockType, _ := blockMap["type"].(string)

		switch blockType {
		case "text":
			text, _ := blockMap["text"].(string)
			if text != "" {
				// Always create new assistant message for each message event
				// (Claude sends complete messages, not deltas)
				newMsg := &models.ChatMessage{
					SessionID: sessionID,
					Role:      "assistant",
					Content:   text,
				}
				if err := s.chatRepo.Create(newMsg); err == nil {
					*assistantMessageID = newMsg.ID
					fmt.Printf("Created assistant message with ID: %d, content: %s\n", newMsg.ID, text)
					
					// Broadcast the text
					s.eventBroadcaster.BroadcastEvent("claude_output", 0, map[string]interface{}{
						"session_id":    sessionID,
						"content_type":  "text",
						"output":        text,
						"is_chunk":      false,
						"db_message_id": newMsg.ID,
					})
				} else {
					fmt.Printf("Failed to create assistant message: %v\n", err)
				}
			}

		case "tool_use":
			toolName, _ := blockMap["name"].(string)
			toolUseID, _ := blockMap["id"].(string)
			toolInput := blockMap["input"]

			// Save tool use message
			toolMsg := &models.ChatMessage{
				SessionID: sessionID,
				Role:      "tool_use",
				Content:   "",
				ToolName:  toolName,
				ToolInput: toolInput,
				ToolUseID: toolUseID,
			}
			if err := s.chatRepo.Create(toolMsg); err == nil {
				fmt.Printf("Created tool_use message: %s with input: %+v\n", toolName, toolInput)
				// Broadcast tool use
				s.eventBroadcaster.BroadcastEvent("claude_output", 0, map[string]interface{}{
					"session_id":    sessionID,
					"content_type":  "tool_use",
					"tool_name":     toolName,
					"tool_input":    toolInput,
					"tool_use_id":   toolUseID,
					"db_message_id": toolMsg.ID,
				})
				fmt.Printf("Broadcasted tool_use event for %s\n", toolName)
			}
		}
	}
}

// handleToolResultMessage handles tool result messages
func (s *ClaudeSessionService) handleToolResultMessage(sessionID int, msg map[string]interface{}) {
	message, ok := msg["message"].(map[string]interface{})
	if !ok {
		return
	}

	content, ok := message["content"].([]interface{})
	if !ok {
		return
	}

	for _, item := range content {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		if itemMap["type"] == "tool_result" {
			toolUseID, _ := itemMap["tool_use_id"].(string)
			toolContent := itemMap["content"]

			// Save tool result message
			toolMsg := &models.ChatMessage{
				SessionID:   sessionID,
				Role:        "tool_result",
				Content:     "",
				ToolUseID:   toolUseID,
				ToolContent: toolContent,
			}
			if err := s.chatRepo.Create(toolMsg); err == nil {
				fmt.Printf("Created tool_result message\n")
				// Broadcast tool result
				s.eventBroadcaster.BroadcastEvent("claude_output", 0, map[string]interface{}{
					"session_id":    sessionID,
					"content_type":  "tool_result",
					"tool_use_id":   toolUseID,
					"tool_content":  toolContent,
					"db_message_id": toolMsg.ID,
				})
			}
		}
	}
}

// GetChatHistory retrieves chat history for a session
func (s *ClaudeSessionService) GetChatHistory(sessionID int, limit int) ([]*models.ChatMessage, error) {
	return s.chatRepo.GetBySessionID(sessionID, limit)
}

// StopGeneration stops the Claude process for a session
func (s *ClaudeSessionService) StopGeneration(sessionID int) error {
	s.processMutex.Lock()
	cmd, exists := s.runningProcesses[sessionID]
	s.processMutex.Unlock()

	if !exists {
		return fmt.Errorf("no running process for session %d", sessionID)
	}

	// Kill the process
	if err := cmd.Process.Kill(); err != nil {
		return fmt.Errorf("failed to kill process: %w", err)
	}

	// Update session status
	if err := s.sessionRepo.UpdateActivityStatus(sessionID, string(models.ActivityStatusIdle)); err != nil {
		fmt.Printf("Failed to update session status after stopping: %v\n", err)
	}

	// Broadcast that generation was stopped
	s.eventBroadcaster.BroadcastEvent("claude_generation_stopped", 0, map[string]interface{}{
		"session_id": sessionID,
	})

	return nil
}
