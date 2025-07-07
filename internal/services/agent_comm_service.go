package services

import (
	"fmt"
	"time"

	"habibi-go/internal/database/repositories"
	"habibi-go/internal/models"
)

type AgentCommService struct {
	agentRepo        *repositories.AgentRepository
	commandRepo      *repositories.AgentCommandRepository
	eventRepo        *repositories.EventRepository
	agentService     *AgentService
	claudeService    *ClaudeAgentService
}

func NewAgentCommService(
	agentRepo *repositories.AgentRepository,
	commandRepo *repositories.AgentCommandRepository,
	eventRepo *repositories.EventRepository,
	agentService *AgentService,
	claudeService *ClaudeAgentService,
) *AgentCommService {
	return &AgentCommService{
		agentRepo:     agentRepo,
		commandRepo:   commandRepo,
		eventRepo:     eventRepo,
		agentService:  agentService,
		claudeService: claudeService,
	}
}

// SendCommand sends a command to an agent and returns the command record
func (s *AgentCommService) SendCommand(agentID int, commandText string) (*models.AgentCommand, error) {
	// Validate agent exists and is running
	agent, err := s.agentRepo.GetByID(agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}
	
	if !agent.IsRunning() {
		return nil, fmt.Errorf("agent is not running (status: %s)", agent.Status)
	}
	
	// Create command record
	command := &models.AgentCommand{
		AgentID:     agentID,
		CommandText: commandText,
		Status:      "pending",
	}
	
	if err := s.commandRepo.Create(command); err != nil {
		return nil, fmt.Errorf("failed to create command record: %w", err)
	}
	
	// Handle Claude agents differently
	if agent.AgentType == "claude-code" {
		// Use Claude-specific service for command execution
		startTime := time.Now()
		
		// Send command through Claude service
		if err := s.claudeService.SendClaudeMessage(agent, commandText); err != nil {
			s.commandRepo.MarkFailed(command.ID, fmt.Sprintf("failed to send Claude message: %v", err))
			return nil, fmt.Errorf("failed to send Claude message: %w", err)
		}
		
		// Create command event
		event := models.NewAgentEvent(models.EventTypeAgentCommand, agentID, map[string]interface{}{
			"command_id": command.ID,
			"command":    commandText,
			"timestamp":  startTime,
		})
		
		if err := s.eventRepo.Create(event); err != nil {
			fmt.Printf("Failed to create command event: %v\n", err)
		}
		
		// Mark command as completed since Claude responds immediately
		executionTime := time.Since(startTime)
		if err := s.commandRepo.MarkCompleted(command.ID, "Command sent to Claude", int(executionTime.Milliseconds())); err != nil {
			fmt.Printf("Failed to mark command as completed: %v\n", err)
		}
		
		return command, nil
	}
	
	// For non-Claude agents, use the original logic
	// Get agent instance from agent service
	s.agentService.activeAgentsMux.RLock()
	instance, exists := s.agentService.activeAgents[agentID]
	s.agentService.activeAgentsMux.RUnlock()
	
	if !exists {
		// Try to restart the agent
		fmt.Printf("Agent %d not in active agents, attempting to restart...\n", agentID)
		
		restartedAgent, err := s.agentService.GetOrRestartAgent(agentID)
		if err != nil {
			s.commandRepo.MarkFailed(command.ID, fmt.Sprintf("failed to restart agent: %v", err))
			return nil, fmt.Errorf("agent instance not found and failed to restart: %w", err)
		}
		
		// Wait a moment for the agent to fully start
		time.Sleep(1 * time.Second)
		
		// Try to get the instance again
		s.agentService.activeAgentsMux.RLock()
		instance, exists = s.agentService.activeAgents[restartedAgent.ID]
		s.agentService.activeAgentsMux.RUnlock()
		
		if !exists {
			s.commandRepo.MarkFailed(command.ID, "agent instance not found after restart")
			return nil, fmt.Errorf("agent instance not found even after restart")
		}
	}
	
	// Send command to agent
	startTime := time.Now()
	
	select {
	case instance.InputWriter <- commandText:
		// Command sent successfully
		
		// Create command event
		event := models.NewAgentEvent(models.EventTypeAgentCommand, agentID, map[string]interface{}{
			"command_id": command.ID,
			"command":    commandText,
			"timestamp":  startTime,
		})
		
		if err := s.eventRepo.Create(event); err != nil {
			fmt.Printf("Failed to create command event: %v\n", err)
		}
		
	case <-time.After(5 * time.Second):
		// Timeout sending command
		s.commandRepo.MarkFailed(command.ID, "timeout sending command to agent")
		return nil, fmt.Errorf("timeout sending command to agent")
	}
	
	return command, nil
}

// SendCommandAndWait sends a command and waits for a response (with timeout)
func (s *AgentCommService) SendCommandAndWait(agentID int, commandText string, timeout time.Duration) (*CommandResult, error) {
	// Send the command
	command, err := s.SendCommand(agentID, commandText)
	if err != nil {
		return nil, err
	}
	
	// Wait for completion
	startTime := time.Now()
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-time.After(timeout):
			// Timeout waiting for response
			s.commandRepo.MarkFailed(command.ID, "timeout waiting for response")
			return nil, fmt.Errorf("timeout waiting for command response")
			
		case <-ticker.C:
			// Check command status
			updatedCommand, err := s.commandRepo.GetByID(command.ID)
			if err != nil {
				continue // Keep trying
			}
			
			if updatedCommand.Status == "completed" {
				executionTime := time.Since(startTime)
				return &CommandResult{
					Command:       updatedCommand,
					Response:      updatedCommand.ResponseText,
					ExecutionTime: executionTime,
					Success:       true,
				}, nil
			}
			
			if updatedCommand.Status == "failed" {
				return &CommandResult{
					Command:       updatedCommand,
					Response:      updatedCommand.ResponseText,
					ExecutionTime: time.Since(startTime),
					Success:       false,
				}, nil
			}
		}
	}
}

type CommandResult struct {
	Command       *models.AgentCommand `json:"command"`
	Response      string               `json:"response"`
	ExecutionTime time.Duration        `json:"execution_time"`
	Success       bool                 `json:"success"`
}

// GetCommandHistory returns command history for an agent
func (s *AgentCommService) GetCommandHistory(agentID int, limit int) ([]*models.AgentCommand, error) {
	commands, err := s.commandRepo.GetByAgentID(agentID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get command history: %w", err)
	}
	
	return commands, nil
}

// GetPendingCommands returns pending commands across all agents
func (s *AgentCommService) GetPendingCommands(limit int) ([]*models.AgentCommand, error) {
	commands, err := s.commandRepo.GetPendingCommands(limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending commands: %w", err)
	}
	
	return commands, nil
}

// CompleteCommand marks a command as completed (typically called by agent output processing)
func (s *AgentCommService) CompleteCommand(commandID int, response string, executionTimeMs int) error {
	if err := s.commandRepo.MarkCompleted(commandID, response, executionTimeMs); err != nil {
		return fmt.Errorf("failed to complete command: %w", err)
	}
	
	// Get command details for event
	command, err := s.commandRepo.GetByID(commandID)
	if err != nil {
		return fmt.Errorf("failed to get command: %w", err)
	}
	
	// Create response event
	event := models.NewAgentEvent(models.EventTypeAgentResponse, command.AgentID, map[string]interface{}{
		"command_id":       commandID,
		"response":         response,
		"execution_time_ms": executionTimeMs,
		"timestamp":        time.Now(),
	})
	
	if err := s.eventRepo.Create(event); err != nil {
		fmt.Printf("Failed to create response event: %v\n", err)
	}
	
	return nil
}

// FailCommand marks a command as failed
func (s *AgentCommService) FailCommand(commandID int, errorMessage string) error {
	if err := s.commandRepo.MarkFailed(commandID, errorMessage); err != nil {
		return fmt.Errorf("failed to fail command: %w", err)
	}
	
	// Get command details for event
	command, err := s.commandRepo.GetByID(commandID)
	if err != nil {
		return fmt.Errorf("failed to get command: %w", err)
	}
	
	// Create failure event
	event := models.NewAgentEvent("agent_command_failed", command.AgentID, map[string]interface{}{
		"command_id": commandID,
		"error":      errorMessage,
		"timestamp":  time.Now(),
	})
	
	if err := s.eventRepo.Create(event); err != nil {
		fmt.Printf("Failed to create command failure event: %v\n", err)
	}
	
	return nil
}

// GetAgentLogs returns recent output from an agent
func (s *AgentCommService) GetAgentLogs(agentID int, since time.Time) ([]string, error) {
	// Get recent command outputs
	commands, err := s.commandRepo.GetRecentCommands(agentID, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent commands: %w", err)
	}
	
	var logs []string
	for _, command := range commands {
		if command.ResponseText != "" {
			logs = append(logs, fmt.Sprintf("[%s] Command: %s", 
				command.CreatedAt.Format("15:04:05"), command.CommandText))
			logs = append(logs, fmt.Sprintf("[%s] Response: %s", 
				command.CreatedAt.Format("15:04:05"), command.ResponseText))
		}
	}
	
	return logs, nil
}

// StreamAgentOutput returns a channel for real-time agent output
func (s *AgentCommService) StreamAgentOutput(agentID int) (<-chan string, error) {
	// Validate agent exists
	agent, err := s.agentRepo.GetByID(agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}
	
	if !agent.IsRunning() {
		return nil, fmt.Errorf("agent is not running")
	}
	
	// Get agent instance
	s.agentService.activeAgentsMux.RLock()
	instance, exists := s.agentService.activeAgents[agentID]
	s.agentService.activeAgentsMux.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("agent instance not found")
	}
	
	// Create a new output channel that mirrors the agent's output
	outputChan := make(chan string, 100)
	
	go func() {
		defer close(outputChan)
		
		// Subscribe to agent output
		for {
			select {
			case output := <-instance.OutputChan:
				select {
				case outputChan <- output:
				default:
					// Channel full, skip this output
				}
			case <-instance.StopChan:
				return
			}
		}
	}()
	
	return outputChan, nil
}

// GetCommandStats returns statistics about commands for an agent
func (s *AgentCommService) GetCommandStats(agentID int) (map[string]interface{}, error) {
	stats, err := s.commandRepo.GetStats(agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get command stats: %w", err)
	}
	
	return stats, nil
}

// CleanupOldCommands removes old command history to save space
func (s *AgentCommService) CleanupOldCommands(agentID int, retainCount int) error {
	if err := s.commandRepo.DeleteOldCommands(agentID, retainCount); err != nil {
		return fmt.Errorf("failed to cleanup old commands: %w", err)
	}
	
	return nil
}

// InteractiveSession starts an interactive session with an agent
func (s *AgentCommService) StartInteractiveSession(agentID int) (*InteractiveSession, error) {
	// Validate agent exists and is running
	agent, err := s.agentRepo.GetByID(agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}
	
	if !agent.IsRunning() {
		return nil, fmt.Errorf("agent is not running")
	}
	
	// Get agent instance
	s.agentService.activeAgentsMux.RLock()
	instance, exists := s.agentService.activeAgents[agentID]
	s.agentService.activeAgentsMux.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("agent instance not found")
	}
	
	session := &InteractiveSession{
		AgentID:     agentID,
		InputChan:   make(chan string, 10),
		OutputChan:  make(chan string, 100),
		StopChan:    make(chan struct{}),
		commService: s,
		instance:    instance,
	}
	
	// Start session handler
	go session.handleSession()
	
	return session, nil
}

type InteractiveSession struct {
	AgentID     int
	InputChan   chan string
	OutputChan  chan string
	StopChan    chan struct{}
	commService *AgentCommService
	instance    *AgentInstance
}

func (is *InteractiveSession) handleSession() {
	defer close(is.OutputChan)
	
	for {
		select {
		case input := <-is.InputChan:
			// Send command to agent
			if _, err := is.commService.SendCommand(is.AgentID, input); err != nil {
				is.OutputChan <- fmt.Sprintf("Error: %v", err)
			}
			
		case output := <-is.instance.OutputChan:
			// Forward output to session
			select {
			case is.OutputChan <- output:
			default:
				// Channel full
			}
			
		case <-is.StopChan:
			return
		}
	}
}

func (is *InteractiveSession) SendInput(input string) {
	select {
	case is.InputChan <- input:
	default:
		// Channel full, input dropped
	}
}

func (is *InteractiveSession) Stop() {
	close(is.StopChan)
}