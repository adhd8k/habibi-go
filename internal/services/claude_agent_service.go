package services

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"habibi-go/internal/database/repositories"
	"habibi-go/internal/models"
)

// ClaudeAgentService handles Claude-specific agent operations
type ClaudeAgentService struct {
	agentRepo        *repositories.AgentRepository
	eventRepo        *repositories.EventRepository
	chatRepo         *repositories.ChatMessageRepository
	sessionRepo      *repositories.SessionRepository
	claudeBinaryPath string
	eventBroadcaster EventBroadcaster
}

// NewClaudeAgentService creates a new Claude agent service
func NewClaudeAgentService(
	agentRepo *repositories.AgentRepository,
	eventRepo *repositories.EventRepository,
	chatRepo *repositories.ChatMessageRepository,
	sessionRepo *repositories.SessionRepository,
	claudeBinaryPath string,
) *ClaudeAgentService {
	return &ClaudeAgentService{
		agentRepo:        agentRepo,
		eventRepo:        eventRepo,
		chatRepo:         chatRepo,
		sessionRepo:      sessionRepo,
		claudeBinaryPath: claudeBinaryPath,
		eventBroadcaster: &NoOpBroadcaster{},
	}
}

// SetEventBroadcaster sets the event broadcaster
func (s *ClaudeAgentService) SetEventBroadcaster(broadcaster EventBroadcaster) {
	s.eventBroadcaster = broadcaster
}

// SendClaudeMessage sends a message to Claude and returns the response
func (s *ClaudeAgentService) SendClaudeMessage(agent *models.Agent, message string) error {
	// Save user message
	userMsg := models.NewChatMessage(agent.ID, "user", message)
	if err := s.chatRepo.Create(userMsg); err != nil {
		fmt.Printf("Failed to save user message: %v\n", err)
	}
	
	// Set session to streaming status
	if err := s.sessionRepo.UpdateActivityStatus(agent.SessionID, string(models.ActivityStatusStreaming)); err != nil {
		fmt.Printf("Failed to update session activity status to streaming: %v\n", err)
	}
	
	// Prepare Claude command with streaming
	args := []string{"--verbose", "-c", "-p", "--dangerously-skip-permissions", "--output-format", "stream-json"}
	
	// If this agent has a Claude session ID, use --resume
	if agent.ClaudeSessionID != "" {
		args = append(args, "--resume", agent.ClaudeSessionID)
	}
	
	// Add the message
	args = append(args, message)
	
	// Execute Claude
	cmd := exec.Command(s.claudeBinaryPath, args...)
	cmd.Dir = agent.WorkingDirectory
	
	fmt.Printf("Executing Claude: %s %v\n", s.claudeBinaryPath, args)
	
	// Start the command and get stdout pipe
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start Claude: %w", err)
	}
	
	// Stream output line by line - no buffer needed, save chunks directly
	go s.streamClaudeOutput(agent, stdout)
	
	// Capture stderr for error handling and session ID extraction
	go s.processClaudeStderr(agent, stderr)
	
	// Wait for command to complete
	if err := cmd.Wait(); err != nil {
		// Check if it's just an exit code issue
		if exitErr, ok := err.(*exec.ExitError); ok {
			fmt.Printf("Claude exited with code: %d\n", exitErr.ExitCode())
			// Non-zero exit might be okay for Claude
		} else {
			return fmt.Errorf("Claude command failed: %w", err)
		}
	}
	
	// Assistant response is saved incrementally during streaming, no need to save here
	
	// Set session to new response status
	if err := s.sessionRepo.UpdateActivityStatus(agent.SessionID, string(models.ActivityStatusNewResponse)); err != nil {
		fmt.Printf("Failed to update session activity status to new response: %v\n", err)
	}
	
	// Send completion event
	s.eventBroadcaster.BroadcastEvent("agent_response_complete", agent.ID, map[string]interface{}{
		"timestamp": time.Now(),
	})
	
	return nil
}

// streamClaudeOutput streams Claude's stdout JSON to WebSocket and database
func (s *ClaudeAgentService) streamClaudeOutput(agent *models.Agent, stdout interface{}) {
	scanner := bufio.NewScanner(stdout.(interface{ Read([]byte) (int, error) }))
	
	// Track the current assistant message being built for this streaming session
	var currentAssistantMessage *models.ChatMessage
	var currentClaudeMessageID string
	
	for scanner.Scan() {
		line := scanner.Text()
		
		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}
		
		// Parse JSON line
		var streamMsg ClaudeStreamMessage
		if err := json.Unmarshal([]byte(line), &streamMsg); err != nil {
			// If it's not JSON, treat as plain text (fallback)
			fmt.Printf("Non-JSON output: %s\n", line)
			continue
		}
		
		// Extract session ID for future use
		if streamMsg.SessionID != "" && agent.ClaudeSessionID != streamMsg.SessionID {
			agent.ClaudeSessionID = streamMsg.SessionID
			if err := s.agentRepo.UpdateClaudeSessionID(agent.ID, streamMsg.SessionID); err != nil {
				fmt.Printf("Failed to update Claude session ID: %v\n", err)
			}
		}
		
		// Handle different message types
		if (streamMsg.Type == "assistant" || streamMsg.Type == "user") && streamMsg.Message != nil {
			// Parse the message as ClaudeMessage
			msgBytes, err := json.Marshal(streamMsg.Message)
			if err != nil {
				continue
			}
			
			var claudeMsg ClaudeMessage
			if err := json.Unmarshal(msgBytes, &claudeMsg); err != nil {
				continue
			}
			
			// Process each content block in the message
			for _, content := range claudeMsg.Content {
				switch content.Type {
				case "text":
					if content.Text != "" {
						// Get or create the assistant message in database for this Claude message ID
						if currentAssistantMessage == nil || currentClaudeMessageID != claudeMsg.ID {
							// Starting a new Claude message - create new DB message
							currentClaudeMessageID = claudeMsg.ID
							currentAssistantMessage = models.NewChatMessage(agent.ID, "assistant", "")
							if err := s.chatRepo.Create(currentAssistantMessage); err != nil {
								fmt.Printf("Failed to create assistant message: %v\n", err)
								continue
							}
						}
						
						// Append the new content to the message
						currentAssistantMessage.Content += content.Text
						
						// Update the message in database
						if err := s.chatRepo.Update(currentAssistantMessage); err != nil {
							fmt.Printf("Failed to update assistant message: %v\n", err)
						}
						
						// Broadcast the content chunk
						s.eventBroadcaster.BroadcastEvent("agent_output", agent.ID, map[string]interface{}{
							"output":        content.Text,
							"message_id":    claudeMsg.ID,
							"db_message_id": currentAssistantMessage.ID,
							"timestamp":     time.Now(),
							"is_chunk":      true,
							"content_type":  "text",
						})
					}
					
				case "tool_use":
					// Create a separate message for tool use
					toolUseMsg := models.NewToolUseMessage(agent.ID, content.Name, content.ID, content.Input)
					if err := s.chatRepo.Create(toolUseMsg); err != nil {
						fmt.Printf("Failed to create tool use message: %v\n", err)
						continue
					}
					
					// Broadcast tool use event
					s.eventBroadcaster.BroadcastEvent("agent_output", agent.ID, map[string]interface{}{
						"tool_use_id":   content.ID,
						"tool_name":     content.Name,
						"tool_input":    content.Input,
						"message_id":    claudeMsg.ID,
						"db_message_id": toolUseMsg.ID,
						"timestamp":     time.Now(),
						"content_type":  "tool_use",
					})
					
				case "tool_result":
					// Create a separate message for tool result
					toolResultMsg := models.NewToolResultMessage(agent.ID, content.ID, content.Content)
					if err := s.chatRepo.Create(toolResultMsg); err != nil {
						fmt.Printf("Failed to create tool result message: %v\n", err)
						continue
					}
					
					// Broadcast tool result event
					s.eventBroadcaster.BroadcastEvent("agent_output", agent.ID, map[string]interface{}{
						"tool_use_id":   content.ID,
						"tool_content":  content.Content,
						"message_id":    claudeMsg.ID,
						"db_message_id": toolResultMsg.ID,
						"timestamp":     time.Now(),
						"content_type":  "tool_result",
					})
				}
			}
		}
		
		// Handle system messages (like init, tool use, etc.)
		if streamMsg.Type == "system" {
			s.eventBroadcaster.BroadcastEvent("agent_system", agent.ID, map[string]interface{}{
				"type":      streamMsg.Type,
				"subtype":   streamMsg.Subtype,
				"message":   streamMsg.Message,
				"timestamp": time.Now(),
			})
		}
	}
}

// processClaudeStderr processes Claude's stderr for session IDs and errors
func (s *ClaudeAgentService) processClaudeStderr(agent *models.Agent, stderr interface{}) {
	scanner := bufio.NewScanner(stderr.(interface{ Read([]byte) (int, error) }))
	
	// Regex to extract session ID from Claude output
	sessionIDRegex := regexp.MustCompile(`session[_-]?id[:\s]+([a-f0-9-]+)`)
	
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Printf("Claude stderr: %s\n", line)
		
		// Try to extract session ID
		if matches := sessionIDRegex.FindStringSubmatch(strings.ToLower(line)); len(matches) > 1 {
			sessionID := matches[1]
			fmt.Printf("Extracted Claude session ID: %s\n", sessionID)
			
			// Update agent with session ID
			agent.ClaudeSessionID = sessionID
			if err := s.agentRepo.UpdateClaudeSessionID(agent.ID, sessionID); err != nil {
				fmt.Printf("Failed to update Claude session ID: %v\n", err)
			}
		}
		
		// Check for specific error patterns
		if strings.Contains(line, "error") || strings.Contains(line, "Error") {
			// Broadcast error
			s.eventBroadcaster.BroadcastEvent("agent_error", agent.ID, map[string]interface{}{
				"error":     line,
				"timestamp": time.Now(),
			})
		}
	}
}

// StartClaudeAgent creates a new Claude agent (virtual, not a process)
func (s *ClaudeAgentService) StartClaudeAgent(sessionID int, workingDirectory string) (*models.Agent, error) {
	// Create agent record
	agent := &models.Agent{
		SessionID:           sessionID,
		AgentType:           "claude-code",
		Status:              string(models.AgentStatusRunning),
		Command:             "claude",
		WorkingDirectory:    workingDirectory,
		CommunicationMethod: string(models.CommunicationMethodStdio),
		Config:              make(map[string]interface{}),
		ResourceUsage:       make(map[string]interface{}),
	}
	
	if err := agent.Validate(); err != nil {
		return nil, fmt.Errorf("agent validation failed: %w", err)
	}
	
	agent.Start()
	
	// Create in database
	if err := s.agentRepo.Create(agent); err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}
	
	// Create start event
	event := models.NewAgentEvent(models.EventTypeAgentStarted, agent.ID, map[string]interface{}{
		"agent_type": agent.AgentType,
		"session_id": agent.SessionID,
	})
	
	if err := s.eventRepo.Create(event); err != nil {
		fmt.Printf("Failed to create agent event: %v\n", err)
	}
	
	return agent, nil
}

// ClaudeStreamMessage represents a single JSON message from Claude's stream-json output
type ClaudeStreamMessage struct {
	Type    string      `json:"type"`
	Subtype string      `json:"subtype,omitempty"`
	Message interface{} `json:"message,omitempty"`
	SessionID string    `json:"session_id,omitempty"`
}

// ClaudeMessage represents the message part of Claude stream
type ClaudeMessage struct {
	ID      string                 `json:"id"`
	Type    string                 `json:"type"`
	Role    string                 `json:"role"`
	Model   string                 `json:"model"`
	Content []ClaudeContentBlock   `json:"content"`
	Usage   map[string]interface{} `json:"usage,omitempty"`
}

// ClaudeContentBlock represents a content block in Claude's response
type ClaudeContentBlock struct {
	Type    string                 `json:"type"`
	Text    string                 `json:"text,omitempty"`
	ID      string                 `json:"id,omitempty"`
	Name    string                 `json:"name,omitempty"`
	Input   map[string]interface{} `json:"input,omitempty"`
	Content interface{}            `json:"content,omitempty"`
}