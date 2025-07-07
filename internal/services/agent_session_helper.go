package services

import (
	"fmt"
	"habibi-go/internal/models"
)

// EnsureClaudeAgentForSession ensures a Claude agent is running for the given session
func (s *AgentService) EnsureClaudeAgentForSession(sessionID int) (*models.Agent, error) {
	// Get all agents for this session
	agents, err := s.agentRepo.GetBySessionID(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agents for session: %w", err)
	}
	
	// Look for a Claude agent
	for _, agent := range agents {
		if agent.AgentType == "claude-code" {
			// Found a Claude agent, check if it's running
			s.activeAgentsMux.RLock()
			_, isActive := s.activeAgents[agent.ID]
			s.activeAgentsMux.RUnlock()
			
			if isActive && agent.IsRunning() {
				// Agent is active and running
				return agent, nil
			}
			
			// Agent exists but not running, restart it
			fmt.Printf("Found Claude agent %d for session %d but it's not running, restarting...\n", agent.ID, sessionID)
			return s.RestartAgent(agent.ID)
		}
	}
	
	// No Claude agent found, create one
	fmt.Printf("No Claude agent found for session %d, creating one...\n", sessionID)
	
	// Get session to get working directory
	session, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	
	req := &models.CreateAgentRequest{
		SessionID:        sessionID,
		AgentType:        "claude-code",
		Command:          "claude",
		WorkingDirectory: session.WorktreePath,
		Config:           make(map[string]interface{}),
	}
	
	return s.StartAgent(req)
}