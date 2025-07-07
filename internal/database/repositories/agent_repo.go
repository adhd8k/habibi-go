package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"habibi-go/internal/models"
)

type AgentRepository struct {
	db *sql.DB
}

func NewAgentRepository(db *sql.DB) *AgentRepository {
	return &AgentRepository{db: db}
}

func (r *AgentRepository) Create(agent *models.Agent) error {
	agent.BeforeCreate()
	
	configStr, err := agent.MarshalConfig()
	if err != nil {
		return err
	}
	
	resourceUsageStr, err := agent.MarshalResourceUsage()
	if err != nil {
		return err
	}
	
	query := `
		INSERT INTO agents (session_id, agent_type, pid, status, config, command, working_directory, 
						   communication_method, input_pipe_path, output_pipe_path, last_heartbeat, 
						   resource_usage, started_at, stopped_at, claude_session_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	result, err := r.db.Exec(query, agent.SessionID, agent.AgentType, agent.PID, agent.Status,
		configStr, agent.Command, agent.WorkingDirectory, agent.CommunicationMethod,
		agent.InputPipePath, agent.OutputPipePath, agent.LastHeartbeat, resourceUsageStr,
		agent.StartedAt, agent.StoppedAt, agent.ClaudeSessionID)
	if err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get agent ID: %w", err)
	}
	
	agent.ID = int(id)
	return nil
}

func (r *AgentRepository) GetByID(id int) (*models.Agent, error) {
	query := `
		SELECT id, session_id, agent_type, pid, status, config, command, working_directory,
			   communication_method, input_pipe_path, output_pipe_path, last_heartbeat,
			   resource_usage, started_at, stopped_at, claude_session_id
		FROM agents
		WHERE id = ?
	`
	
	var agent models.Agent
	var configStr, resourceUsageStr string
	var lastHeartbeat, stoppedAt sql.NullTime
	
	err := r.db.QueryRow(query, id).Scan(
		&agent.ID, &agent.SessionID, &agent.AgentType, &agent.PID, &agent.Status,
		&configStr, &agent.Command, &agent.WorkingDirectory, &agent.CommunicationMethod,
		&agent.InputPipePath, &agent.OutputPipePath, &lastHeartbeat, &resourceUsageStr,
		&agent.StartedAt, &stoppedAt, &agent.ClaudeSessionID,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("agent not found")
		}
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}
	
	if err := agent.UnmarshalConfig(configStr); err != nil {
		return nil, err
	}
	
	if err := agent.UnmarshalResourceUsage(resourceUsageStr); err != nil {
		return nil, err
	}
	
	if lastHeartbeat.Valid {
		agent.LastHeartbeat = &lastHeartbeat.Time
	}
	
	if stoppedAt.Valid {
		agent.StoppedAt = &stoppedAt.Time
	}
	
	return &agent, nil
}

func (r *AgentRepository) GetBySessionID(sessionID int) ([]*models.Agent, error) {
	query := `
		SELECT id, session_id, agent_type, pid, status, config, command, working_directory,
			   communication_method, input_pipe_path, output_pipe_path, last_heartbeat,
			   resource_usage, started_at, stopped_at, claude_session_id
		FROM agents
		WHERE session_id = ?
		ORDER BY started_at DESC
	`
	
	rows, err := r.db.Query(query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agents: %w", err)
	}
	defer rows.Close()
	
	var agents []*models.Agent
	
	for rows.Next() {
		var agent models.Agent
		var configStr, resourceUsageStr string
		var lastHeartbeat, stoppedAt sql.NullTime
		
		err := rows.Scan(
			&agent.ID, &agent.SessionID, &agent.AgentType, &agent.PID, &agent.Status,
			&configStr, &agent.Command, &agent.WorkingDirectory, &agent.CommunicationMethod,
			&agent.InputPipePath, &agent.OutputPipePath, &lastHeartbeat, &resourceUsageStr,
			&agent.StartedAt, &stoppedAt, &agent.ClaudeSessionID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan agent: %w", err)
		}
		
		if err := agent.UnmarshalConfig(configStr); err != nil {
			return nil, err
		}
		
		if err := agent.UnmarshalResourceUsage(resourceUsageStr); err != nil {
			return nil, err
		}
		
		if lastHeartbeat.Valid {
			agent.LastHeartbeat = &lastHeartbeat.Time
		}
		
		if stoppedAt.Valid {
			agent.StoppedAt = &stoppedAt.Time
		}
		
		agents = append(agents, &agent)
	}
	
	return agents, nil
}

func (r *AgentRepository) GetAll() ([]*models.Agent, error) {
	query := `
		SELECT id, session_id, agent_type, pid, status, config, command, working_directory,
			   communication_method, input_pipe_path, output_pipe_path, last_heartbeat,
			   resource_usage, started_at, stopped_at, claude_session_id
		FROM agents
		ORDER BY started_at DESC
	`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get agents: %w", err)
	}
	defer rows.Close()
	
	var agents []*models.Agent
	
	for rows.Next() {
		var agent models.Agent
		var configStr, resourceUsageStr string
		var lastHeartbeat, stoppedAt sql.NullTime
		
		err := rows.Scan(
			&agent.ID, &agent.SessionID, &agent.AgentType, &agent.PID, &agent.Status,
			&configStr, &agent.Command, &agent.WorkingDirectory, &agent.CommunicationMethod,
			&agent.InputPipePath, &agent.OutputPipePath, &lastHeartbeat, &resourceUsageStr,
			&agent.StartedAt, &stoppedAt, &agent.ClaudeSessionID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan agent: %w", err)
		}
		
		if err := agent.UnmarshalConfig(configStr); err != nil {
			return nil, err
		}
		
		if err := agent.UnmarshalResourceUsage(resourceUsageStr); err != nil {
			return nil, err
		}
		
		if lastHeartbeat.Valid {
			agent.LastHeartbeat = &lastHeartbeat.Time
		}
		
		if stoppedAt.Valid {
			agent.StoppedAt = &stoppedAt.Time
		}
		
		agents = append(agents, &agent)
	}
	
	return agents, nil
}

func (r *AgentRepository) Update(agent *models.Agent) error {
	configStr, err := agent.MarshalConfig()
	if err != nil {
		return err
	}
	
	resourceUsageStr, err := agent.MarshalResourceUsage()
	if err != nil {
		return err
	}
	
	query := `
		UPDATE agents
		SET session_id = ?, agent_type = ?, pid = ?, status = ?, config = ?, command = ?,
			working_directory = ?, communication_method = ?, input_pipe_path = ?, 
			output_pipe_path = ?, last_heartbeat = ?, resource_usage = ?, stopped_at = ?,
			claude_session_id = ?
		WHERE id = ?
	`
	
	result, err := r.db.Exec(query, agent.SessionID, agent.AgentType, agent.PID, agent.Status,
		configStr, agent.Command, agent.WorkingDirectory, agent.CommunicationMethod,
		agent.InputPipePath, agent.OutputPipePath, agent.LastHeartbeat, resourceUsageStr,
		agent.StoppedAt, agent.ClaudeSessionID, agent.ID)
	if err != nil {
		return fmt.Errorf("failed to update agent: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("agent not found")
	}
	
	return nil
}

func (r *AgentRepository) Delete(id int) error {
	query := `DELETE FROM agents WHERE id = ?`
	
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete agent: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("agent not found")
	}
	
	return nil
}

func (r *AgentRepository) GetByStatus(status string) ([]*models.Agent, error) {
	query := `
		SELECT id, session_id, agent_type, pid, status, config, command, working_directory,
			   communication_method, input_pipe_path, output_pipe_path, last_heartbeat,
			   resource_usage, started_at, stopped_at, claude_session_id
		FROM agents
		WHERE status = ?
		ORDER BY started_at DESC
	`
	
	rows, err := r.db.Query(query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get agents by status: %w", err)
	}
	defer rows.Close()
	
	var agents []*models.Agent
	
	for rows.Next() {
		var agent models.Agent
		var configStr, resourceUsageStr string
		var lastHeartbeat, stoppedAt sql.NullTime
		
		err := rows.Scan(
			&agent.ID, &agent.SessionID, &agent.AgentType, &agent.PID, &agent.Status,
			&configStr, &agent.Command, &agent.WorkingDirectory, &agent.CommunicationMethod,
			&agent.InputPipePath, &agent.OutputPipePath, &lastHeartbeat, &resourceUsageStr,
			&agent.StartedAt, &stoppedAt, &agent.ClaudeSessionID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan agent: %w", err)
		}
		
		if err := agent.UnmarshalConfig(configStr); err != nil {
			return nil, err
		}
		
		if err := agent.UnmarshalResourceUsage(resourceUsageStr); err != nil {
			return nil, err
		}
		
		if lastHeartbeat.Valid {
			agent.LastHeartbeat = &lastHeartbeat.Time
		}
		
		if stoppedAt.Valid {
			agent.StoppedAt = &stoppedAt.Time
		}
		
		agents = append(agents, &agent)
	}
	
	return agents, nil
}

func (r *AgentRepository) GetRunningAgents() ([]*models.Agent, error) {
	return r.GetByStatus(string(models.AgentStatusRunning))
}

func (r *AgentRepository) UpdateHeartbeat(id int) error {
	query := `UPDATE agents SET last_heartbeat = CURRENT_TIMESTAMP WHERE id = ?`
	
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to update heartbeat: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("agent not found")
	}
	
	return nil
}

func (r *AgentRepository) UpdateResourceUsage(id int, resourceUsage map[string]interface{}) error {
	resourceUsageStr := "{}"
	if resourceUsage != nil {
		agent := &models.Agent{ResourceUsage: resourceUsage}
		var err error
		resourceUsageStr, err = agent.MarshalResourceUsage()
		if err != nil {
			return err
		}
	}
	
	query := `UPDATE agents SET resource_usage = ? WHERE id = ?`
	
	result, err := r.db.Exec(query, resourceUsageStr, id)
	if err != nil {
		return fmt.Errorf("failed to update resource usage: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("agent not found")
	}
	
	return nil
}

func (r *AgentRepository) GetStaleAgents(timeout time.Duration) ([]*models.Agent, error) {
	query := `
		SELECT id, session_id, agent_type, pid, status, config, command, working_directory,
			   communication_method, input_pipe_path, output_pipe_path, last_heartbeat,
			   resource_usage, started_at, stopped_at, claude_session_id
		FROM agents
		WHERE status = ? AND (last_heartbeat IS NULL OR last_heartbeat < datetime('now', '-' || ? || ' seconds'))
		ORDER BY started_at DESC
	`
	
	rows, err := r.db.Query(query, string(models.AgentStatusRunning), int(timeout.Seconds()))
	if err != nil {
		return nil, fmt.Errorf("failed to get stale agents: %w", err)
	}
	defer rows.Close()
	
	var agents []*models.Agent
	
	for rows.Next() {
		var agent models.Agent
		var configStr, resourceUsageStr string
		var lastHeartbeat, stoppedAt sql.NullTime
		
		err := rows.Scan(
			&agent.ID, &agent.SessionID, &agent.AgentType, &agent.PID, &agent.Status,
			&configStr, &agent.Command, &agent.WorkingDirectory, &agent.CommunicationMethod,
			&agent.InputPipePath, &agent.OutputPipePath, &lastHeartbeat, &resourceUsageStr,
			&agent.StartedAt, &stoppedAt, &agent.ClaudeSessionID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan agent: %w", err)
		}
		
		if err := agent.UnmarshalConfig(configStr); err != nil {
			return nil, err
		}
		
		if err := agent.UnmarshalResourceUsage(resourceUsageStr); err != nil {
			return nil, err
		}
		
		if lastHeartbeat.Valid {
			agent.LastHeartbeat = &lastHeartbeat.Time
		}
		
		if stoppedAt.Valid {
			agent.StoppedAt = &stoppedAt.Time
		}
		
		agents = append(agents, &agent)
	}
	
	return agents, nil
}

func (r *AgentRepository) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// Total agents
	var totalAgents int
	err := r.db.QueryRow("SELECT COUNT(*) FROM agents").Scan(&totalAgents)
	if err != nil {
		return nil, fmt.Errorf("failed to get total agents: %w", err)
	}
	stats["total_agents"] = totalAgents
	
	// Agents by status
	statusQuery := `
		SELECT status, COUNT(*) as count
		FROM agents
		GROUP BY status
		ORDER BY count DESC
	`
	
	rows, err := r.db.Query(statusQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent status counts: %w", err)
	}
	defer rows.Close()
	
	statusCounts := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		
		err := rows.Scan(&status, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan status count: %w", err)
		}
		
		statusCounts[status] = count
	}
	stats["status_counts"] = statusCounts
	
	// Running agents count
	var runningAgents int
	err = r.db.QueryRow("SELECT COUNT(*) FROM agents WHERE status = ?", string(models.AgentStatusRunning)).Scan(&runningAgents)
	if err != nil {
		return nil, fmt.Errorf("failed to get running agents count: %w", err)
	}
	stats["running_agents"] = runningAgents
	
	// Agents by type
	typeQuery := `
		SELECT agent_type, COUNT(*) as count
		FROM agents
		GROUP BY agent_type
		ORDER BY count DESC
	`
	
	rows, err = r.db.Query(typeQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent type counts: %w", err)
	}
	defer rows.Close()
	
	typeCounts := make(map[string]int)
	for rows.Next() {
		var agentType string
		var count int
		
		err := rows.Scan(&agentType, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan type count: %w", err)
		}
		
		typeCounts[agentType] = count
	}
	stats["type_counts"] = typeCounts
	
	return stats, nil
}

func (r *AgentRepository) UpdateClaudeSessionID(id int, sessionID string) error {
	query := `UPDATE agents SET claude_session_id = ? WHERE id = ?`
	
	result, err := r.db.Exec(query, sessionID, id)
	if err != nil {
		return fmt.Errorf("failed to update claude session ID: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("agent not found")
	}
	
	return nil
}