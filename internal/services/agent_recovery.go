package services

import (
	"fmt"
	"habibi-go/internal/models"
)

// RecoverActiveAgents recovers agents that were running before server restart
func (s *AgentService) RecoverActiveAgents() error {
	// Get all agents with "running" status
	agents, err := s.agentRepo.GetAll()
	if err != nil {
		return fmt.Errorf("failed to get agents for recovery: %w", err)
	}

	recoveredCount := 0
	for _, agent := range agents {
		if agent.Status == string(models.AgentStatusRunning) {
			fmt.Printf("Found agent %d marked as running, updating status to failed\n", agent.ID)
			
			// Mark as failed since the process is gone after restart
			agent.Status = string(models.AgentStatusFailed)
			agent.Stop()
			
			if err := s.agentRepo.Update(agent); err != nil {
				fmt.Printf("Failed to update agent %d status: %v\n", agent.ID, err)
				continue
			}
			
			// Create failure event
			event := models.NewAgentEvent(models.EventTypeAgentFailed, agent.ID, map[string]interface{}{
				"reason": "server_restart",
				"pid":    agent.PID,
			})
			
			if err := s.eventRepo.Create(event); err != nil {
				fmt.Printf("Failed to create agent failure event: %v\n", err)
			}
			
			recoveredCount++
		}
	}
	
	if recoveredCount > 0 {
		fmt.Printf("Marked %d agents as failed after server restart\n", recoveredCount)
	}
	
	return nil
}

// GetOrRestartAgent gets an agent and restarts it if it's not running
func (s *AgentService) GetOrRestartAgent(agentID int) (*models.Agent, error) {
	// Check if agent is in active agents
	s.activeAgentsMux.RLock()
	_, exists := s.activeAgents[agentID]
	s.activeAgentsMux.RUnlock()
	
	if exists {
		// Agent is active, just return it
		return s.agentRepo.GetByID(agentID)
	}
	
	// Agent not active, check database
	agent, err := s.agentRepo.GetByID(agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}
	
	// If agent is marked as running but not in active agents, it needs restart
	if agent.Status == string(models.AgentStatusRunning) {
		fmt.Printf("Agent %d marked as running but not active, restarting...\n", agentID)
		return s.RestartAgent(agentID)
	}
	
	// If agent is not running, try to start it
	if agent.Status != string(models.AgentStatusRunning) {
		fmt.Printf("Agent %d is not running (status: %s), starting...\n", agentID, agent.Status)
		return s.RestartAgent(agentID)
	}
	
	return agent, nil
}