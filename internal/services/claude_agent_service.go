package services

import (
	"bufio"
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
	claudeBinaryPath string
	eventBroadcaster EventBroadcaster
}

// NewClaudeAgentService creates a new Claude agent service
func NewClaudeAgentService(
	agentRepo *repositories.AgentRepository,
	eventRepo *repositories.EventRepository,
	claudeBinaryPath string,
) *ClaudeAgentService {
	return &ClaudeAgentService{
		agentRepo:        agentRepo,
		eventRepo:        eventRepo,
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
	// Prepare Claude command
	args := []string{"--print"} // Use print mode for single response
	
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
	
	// Stream output line by line
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
	
	// Send completion event
	s.eventBroadcaster.BroadcastEvent("agent_response_complete", agent.ID, map[string]interface{}{
		"timestamp": time.Now(),
	})
	
	return nil
}

// streamClaudeOutput streams Claude's stdout to WebSocket
func (s *ClaudeAgentService) streamClaudeOutput(agent *models.Agent, stdout interface{}) {
	scanner := bufio.NewScanner(stdout.(interface{ Read([]byte) (int, error) }))
	
	for scanner.Scan() {
		line := scanner.Text()
		
		// Broadcast the output line
		s.eventBroadcaster.BroadcastEvent("agent_output", agent.ID, map[string]interface{}{
			"output":    line,
			"timestamp": time.Now(),
		})
		
		// Also create event in database
		event := models.NewAgentEvent("agent_output", agent.ID, map[string]interface{}{
			"output":    line,
			"timestamp": time.Now(),
		})
		
		if err := s.eventRepo.Create(event); err != nil {
			fmt.Printf("Failed to create output event: %v\n", err)
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

// ClaudeResponse represents a Claude response (for JSON parsing if needed)
type ClaudeResponse struct {
	SessionID string `json:"session_id,omitempty"`
	Content   string `json:"content"`
	Error     string `json:"error,omitempty"`
}