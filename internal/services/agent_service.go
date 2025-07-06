package services

import (
	"fmt"
	"os/exec"
	"sync"
	"time"

	"habibi-go/internal/database/repositories"
	"habibi-go/internal/models"
	"habibi-go/internal/util"
)

type AgentService struct {
	agentRepo        *repositories.AgentRepository
	sessionRepo      *repositories.SessionRepository
	eventRepo        *repositories.EventRepository
	processManager   *util.ProcessManager
	
	// Active agents tracking
	activeAgents     map[int]*AgentInstance
	activeAgentsMux  sync.RWMutex
	
	// Configuration
	maxConcurrent    int
	healthInterval   time.Duration
	healthTimeout    time.Duration
}

type AgentInstance struct {
	Agent      *models.Agent
	Process    *exec.Cmd
	OutputChan chan string
	ErrorChan  chan string
	StopChan   chan struct{}
	
	// Communication channels
	InputWriter  chan string
	
	// Monitoring
	LastSeen     time.Time
	ProcessInfo  *util.ProcessInfo
}

func NewAgentService(
	agentRepo *repositories.AgentRepository,
	sessionRepo *repositories.SessionRepository,
	eventRepo *repositories.EventRepository,
) *AgentService {
	return &AgentService{
		agentRepo:       agentRepo,
		sessionRepo:     sessionRepo,
		eventRepo:       eventRepo,
		processManager:  util.NewProcessManager(),
		activeAgents:    make(map[int]*AgentInstance),
		maxConcurrent:   10,
		healthInterval:  30 * time.Second,
		healthTimeout:   5 * time.Minute,
	}
}

func (s *AgentService) StartAgent(req *models.CreateAgentRequest) (*models.Agent, error) {
	// Validate request
	if err := s.validateCreateRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}
	
	// Check concurrent limit
	if s.getActiveAgentCount() >= s.maxConcurrent {
		return nil, fmt.Errorf("maximum concurrent agents reached (%d)", s.maxConcurrent)
	}
	
	// Get session to validate it exists and get working directory
	session, err := s.sessionRepo.GetByID(req.SessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	
	// Set working directory to session's worktree path if not specified
	workingDir := req.WorkingDirectory
	if workingDir == "" {
		workingDir = session.WorktreePath
	}
	
	// Create agent record
	agent := &models.Agent{
		SessionID:           req.SessionID,
		AgentType:           req.AgentType,
		Status:              string(models.AgentStatusStarting),
		Config:              req.Config,
		Command:             req.Command,
		WorkingDirectory:    workingDir,
		CommunicationMethod: string(models.CommunicationMethodStdio),
		ResourceUsage:       make(map[string]interface{}),
	}
	
	if err := agent.Validate(); err != nil {
		return nil, fmt.Errorf("agent validation failed: %w", err)
	}
	
	// Create agent in database
	if err := s.agentRepo.Create(agent); err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}
	
	// Start the actual process
	instance, err := s.startAgentProcess(agent)
	if err != nil {
		// Clean up database record on failure
		s.agentRepo.Delete(agent.ID)
		return nil, fmt.Errorf("failed to start agent process: %w", err)
	}
	
	// Update agent with PID and status
	agent.PID = instance.Process.Process.Pid
	agent.Status = string(models.AgentStatusRunning)
	agent.Start()
	
	if err := s.agentRepo.Update(agent); err != nil {
		// Kill the process and clean up on database update failure
		s.processManager.KillProcessGroup(agent.PID)
		s.agentRepo.Delete(agent.ID)
		return nil, fmt.Errorf("failed to update agent status: %w", err)
	}
	
	// Store in active agents
	s.activeAgentsMux.Lock()
	s.activeAgents[agent.ID] = instance
	s.activeAgentsMux.Unlock()
	
	// Start monitoring goroutine
	go s.monitorAgent(instance)
	
	// Create agent event
	event := models.NewAgentEvent(models.EventTypeAgentStarted, agent.ID, map[string]interface{}{
		"agent_type":        agent.AgentType,
		"pid":              agent.PID,
		"command":          agent.Command,
		"working_directory": agent.WorkingDirectory,
		"session_id":       agent.SessionID,
	})
	
	if err := s.eventRepo.Create(event); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to create agent event: %v\n", err)
	}
	
	return agent, nil
}

func (s *AgentService) StopAgent(id int) error {
	agent, err := s.agentRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get agent: %w", err)
	}
	
	// Get agent instance
	s.activeAgentsMux.Lock()
	instance, exists := s.activeAgents[id]
	if exists {
		delete(s.activeAgents, id)
	}
	s.activeAgentsMux.Unlock()
	
	// Stop the process if it's running
	if exists && instance.Process != nil {
		// Signal shutdown
		close(instance.StopChan)
		
		// Kill the process group
		if err := s.processManager.KillProcessGroup(agent.PID); err != nil {
			fmt.Printf("Warning: failed to kill process group for agent %d: %v\n", id, err)
		}
	}
	
	// Update agent status
	agent.Stop()
	if err := s.agentRepo.Update(agent); err != nil {
		return fmt.Errorf("failed to update agent status: %w", err)
	}
	
	// Create agent event
	event := models.NewAgentEvent(models.EventTypeAgentStopped, agent.ID, map[string]interface{}{
		"agent_type": agent.AgentType,
		"pid":       agent.PID,
		"reason":    "manual_stop",
	})
	
	if err := s.eventRepo.Create(event); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to create agent event: %v\n", err)
	}
	
	return nil
}

func (s *AgentService) GetAgent(id int) (*models.Agent, error) {
	agent, err := s.agentRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}
	
	return agent, nil
}

func (s *AgentService) GetAgentsBySession(sessionID int) ([]*models.Agent, error) {
	agents, err := s.agentRepo.GetBySessionID(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agents: %w", err)
	}
	
	return agents, nil
}

func (s *AgentService) GetAllAgents() ([]*models.Agent, error) {
	agents, err := s.agentRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get agents: %w", err)
	}
	
	return agents, nil
}

func (s *AgentService) GetAgentStatus(id int) (*AgentStatusInfo, error) {
	agent, err := s.agentRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}
	
	status := &AgentStatusInfo{
		Agent: agent,
	}
	
	// Check if agent is in active agents
	s.activeAgentsMux.RLock()
	instance, isActive := s.activeAgents[id]
	s.activeAgentsMux.RUnlock()
	
	status.IsActive = isActive
	
	if isActive {
		status.LastSeen = instance.LastSeen
		status.ProcessInfo = instance.ProcessInfo
	}
	
	// Check if process is actually running
	if agent.PID > 0 {
		status.ProcessExists = s.processManager.ProcessExists(agent.PID)
		
		if status.ProcessExists {
			if processInfo, err := s.processManager.GetProcessInfo(agent.PID); err == nil {
				status.ProcessInfo = processInfo
			}
		}
	}
	
	// Check health
	status.IsHealthy = status.IsActive && status.ProcessExists && agent.IsHealthy(s.healthTimeout)
	
	return status, nil
}

type AgentStatusInfo struct {
	Agent         *models.Agent      `json:"agent"`
	IsActive      bool              `json:"is_active"`
	ProcessExists bool              `json:"process_exists"`
	IsHealthy     bool              `json:"is_healthy"`
	LastSeen      time.Time         `json:"last_seen"`
	ProcessInfo   *util.ProcessInfo `json:"process_info,omitempty"`
}

func (s *AgentService) RestartAgent(id int) (*models.Agent, error) {
	// Stop the agent first
	if err := s.StopAgent(id); err != nil {
		return nil, fmt.Errorf("failed to stop agent: %w", err)
	}
	
	// Get agent details
	agent, err := s.agentRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}
	
	// Create restart request
	req := &models.CreateAgentRequest{
		SessionID:        agent.SessionID,
		AgentType:        agent.AgentType,
		Command:          agent.Command,
		WorkingDirectory: agent.WorkingDirectory,
		Config:           agent.Config,
	}
	
	// Start new agent
	newAgent, err := s.StartAgent(req)
	if err != nil {
		return nil, fmt.Errorf("failed to restart agent: %w", err)
	}
	
	return newAgent, nil
}

func (s *AgentService) CleanupStaleAgents() error {
	// Get stale agents
	staleAgents, err := s.agentRepo.GetStaleAgents(s.healthTimeout)
	if err != nil {
		return fmt.Errorf("failed to get stale agents: %w", err)
	}
	
	for _, agent := range staleAgents {
		fmt.Printf("Cleaning up stale agent %d (PID: %d)\n", agent.ID, agent.PID)
		
		// Check if process still exists
		if agent.PID > 0 && s.processManager.ProcessExists(agent.PID) {
			// Kill the process
			if err := s.processManager.KillProcessGroup(agent.PID); err != nil {
				fmt.Printf("Warning: failed to kill stale agent process %d: %v\n", agent.PID, err)
			}
		}
		
		// Remove from active agents
		s.activeAgentsMux.Lock()
		delete(s.activeAgents, agent.ID)
		s.activeAgentsMux.Unlock()
		
		// Update status to failed
		agent.Fail()
		if err := s.agentRepo.Update(agent); err != nil {
			fmt.Printf("Warning: failed to update stale agent status: %v\n", err)
		}
		
		// Create failure event
		event := models.NewAgentEvent(models.EventTypeAgentFailed, agent.ID, map[string]interface{}{
			"reason": "stale_timeout",
			"pid":    agent.PID,
		})
		
		if err := s.eventRepo.Create(event); err != nil {
			fmt.Printf("Warning: failed to create agent failure event: %v\n", err)
		}
	}
	
	return nil
}

func (s *AgentService) GetAgentStats() (map[string]interface{}, error) {
	stats, err := s.agentRepo.GetStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get agent stats: %w", err)
	}
	
	// Add runtime stats
	s.activeAgentsMux.RLock()
	stats["active_agents"] = len(s.activeAgents)
	s.activeAgentsMux.RUnlock()
	
	return stats, nil
}

// Helper methods

func (s *AgentService) validateCreateRequest(req *models.CreateAgentRequest) error {
	if req.SessionID == 0 {
		return fmt.Errorf("session ID is required")
	}
	
	if req.AgentType == "" {
		return fmt.Errorf("agent type is required")
	}
	
	if req.Command == "" {
		return fmt.Errorf("command is required")
	}
	
	return nil
}

func (s *AgentService) getActiveAgentCount() int {
	s.activeAgentsMux.RLock()
	defer s.activeAgentsMux.RUnlock()
	return len(s.activeAgents)
}

func (s *AgentService) startAgentProcess(agent *models.Agent) (*AgentInstance, error) {
	// Parse command and arguments
	cmd := exec.Command("sh", "-c", agent.Command)
	cmd.Dir = agent.WorkingDirectory
	
	// Create pipes for communication
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	
	// Start the process
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start process: %w", err)
	}
	
	// Create agent instance
	instance := &AgentInstance{
		Agent:       agent,
		Process:     cmd,
		OutputChan:  make(chan string, 100),
		ErrorChan:   make(chan string, 100),
		StopChan:    make(chan struct{}),
		InputWriter: make(chan string, 10),
		LastSeen:    time.Now(),
	}
	
	// Start output streaming goroutines
	go s.processManager.StreamOutput(stdout, instance.OutputChan)
	go s.processManager.StreamOutput(stderr, instance.ErrorChan)
	
	// Start input handling goroutine
	go func() {
		defer stdin.Close()
		for {
			select {
			case input := <-instance.InputWriter:
				if _, err := stdin.Write([]byte(input + "\n")); err != nil {
					fmt.Printf("Failed to write to agent stdin: %v\n", err)
					return
				}
			case <-instance.StopChan:
				return
			}
		}
	}()
	
	return instance, nil
}

func (s *AgentService) monitorAgent(instance *AgentInstance) {
	defer func() {
		// Clean up when monitoring stops
		s.activeAgentsMux.Lock()
		delete(s.activeAgents, instance.Agent.ID)
		s.activeAgentsMux.Unlock()
	}()
	
	ticker := time.NewTicker(s.healthInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-instance.StopChan:
			return
		case <-ticker.C:
			// Update heartbeat
			if err := s.agentRepo.UpdateHeartbeat(instance.Agent.ID); err != nil {
				fmt.Printf("Failed to update heartbeat for agent %d: %v\n", instance.Agent.ID, err)
			}
			
			// Update process info
			if instance.Process != nil && instance.Process.Process != nil {
				if processInfo, err := s.processManager.GetProcessInfo(instance.Process.Process.Pid); err == nil {
					instance.ProcessInfo = processInfo
					instance.LastSeen = time.Now()
					
					// Update resource usage in database
					resourceUsage := map[string]interface{}{
						"cpu_percent": processInfo.CPUPercent,
						"memory_mb":   processInfo.MemoryMB,
						"status":      processInfo.Status,
					}
					
					if err := s.agentRepo.UpdateResourceUsage(instance.Agent.ID, resourceUsage); err != nil {
						fmt.Printf("Failed to update resource usage for agent %d: %v\n", instance.Agent.ID, err)
					}
				}
			}
			
		case output := <-instance.OutputChan:
			// Broadcast output event
			event := models.NewAgentEvent("agent_output", instance.Agent.ID, map[string]interface{}{
				"output":    output,
				"timestamp": time.Now(),
			})
			
			if err := s.eventRepo.Create(event); err != nil {
				fmt.Printf("Failed to create agent output event: %v\n", err)
			}
			
		case errorOutput := <-instance.ErrorChan:
			// Broadcast error output event
			event := models.NewAgentEvent("agent_error", instance.Agent.ID, map[string]interface{}{
				"error":     errorOutput,
				"timestamp": time.Now(),
			})
			
			if err := s.eventRepo.Create(event); err != nil {
				fmt.Printf("Failed to create agent error event: %v\n", err)
			}
		}
		
		// Check if process is still running
		if instance.Process != nil {
			select {
			case <-time.After(100 * time.Millisecond):
				// Non-blocking check
			default:
				if err := instance.Process.Process.Signal(nil); err != nil {
					// Process is dead
					fmt.Printf("Agent %d process died\n", instance.Agent.ID)
					
					// Update status
					instance.Agent.Fail()
					if err := s.agentRepo.Update(instance.Agent); err != nil {
						fmt.Printf("Failed to update dead agent status: %v\n", err)
					}
					
					// Create failure event
					event := models.NewAgentEvent(models.EventTypeAgentFailed, instance.Agent.ID, map[string]interface{}{
						"reason": "process_died",
						"pid":    instance.Agent.PID,
					})
					
					if err := s.eventRepo.Create(event); err != nil {
						fmt.Printf("Failed to create agent failure event: %v\n", err)
					}
					
					return
				}
			}
		}
	}
}