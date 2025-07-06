package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"habibi-go/internal/models"
)

type AgentCommandRepository struct {
	db *sql.DB
}

func NewAgentCommandRepository(db *sql.DB) *AgentCommandRepository {
	return &AgentCommandRepository{db: db}
}

func (r *AgentCommandRepository) Create(command *models.AgentCommand) error {
	command.CreatedAt = time.Now()
	
	query := `
		INSERT INTO agent_commands (agent_id, command_text, response_text, status, execution_time_ms, created_at, completed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	
	result, err := r.db.Exec(query, command.AgentID, command.CommandText, command.ResponseText,
		command.Status, command.ExecutionTimeMs, command.CreatedAt, command.CompletedAt)
	if err != nil {
		return fmt.Errorf("failed to create agent command: %w", err)
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get command ID: %w", err)
	}
	
	command.ID = int(id)
	return nil
}

func (r *AgentCommandRepository) GetByID(id int) (*models.AgentCommand, error) {
	query := `
		SELECT id, agent_id, command_text, response_text, status, execution_time_ms, created_at, completed_at
		FROM agent_commands
		WHERE id = ?
	`
	
	var command models.AgentCommand
	var completedAt sql.NullTime
	
	err := r.db.QueryRow(query, id).Scan(
		&command.ID, &command.AgentID, &command.CommandText, &command.ResponseText,
		&command.Status, &command.ExecutionTimeMs, &command.CreatedAt, &completedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("command not found")
		}
		return nil, fmt.Errorf("failed to get command: %w", err)
	}
	
	if completedAt.Valid {
		command.CompletedAt = &completedAt.Time
	}
	
	return &command, nil
}

func (r *AgentCommandRepository) GetByAgentID(agentID int, limit int) ([]*models.AgentCommand, error) {
	query := `
		SELECT id, agent_id, command_text, response_text, status, execution_time_ms, created_at, completed_at
		FROM agent_commands
		WHERE agent_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`
	
	rows, err := r.db.Query(query, agentID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get commands: %w", err)
	}
	defer rows.Close()
	
	var commands []*models.AgentCommand
	
	for rows.Next() {
		var command models.AgentCommand
		var completedAt sql.NullTime
		
		err := rows.Scan(
			&command.ID, &command.AgentID, &command.CommandText, &command.ResponseText,
			&command.Status, &command.ExecutionTimeMs, &command.CreatedAt, &completedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan command: %w", err)
		}
		
		if completedAt.Valid {
			command.CompletedAt = &completedAt.Time
		}
		
		commands = append(commands, &command)
	}
	
	return commands, nil
}

func (r *AgentCommandRepository) Update(command *models.AgentCommand) error {
	query := `
		UPDATE agent_commands
		SET command_text = ?, response_text = ?, status = ?, execution_time_ms = ?, completed_at = ?
		WHERE id = ?
	`
	
	result, err := r.db.Exec(query, command.CommandText, command.ResponseText, command.Status,
		command.ExecutionTimeMs, command.CompletedAt, command.ID)
	if err != nil {
		return fmt.Errorf("failed to update command: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("command not found")
	}
	
	return nil
}

func (r *AgentCommandRepository) Delete(id int) error {
	query := `DELETE FROM agent_commands WHERE id = ?`
	
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete command: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("command not found")
	}
	
	return nil
}

func (r *AgentCommandRepository) GetByStatus(status string, limit int) ([]*models.AgentCommand, error) {
	query := `
		SELECT id, agent_id, command_text, response_text, status, execution_time_ms, created_at, completed_at
		FROM agent_commands
		WHERE status = ?
		ORDER BY created_at DESC
		LIMIT ?
	`
	
	rows, err := r.db.Query(query, status, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get commands by status: %w", err)
	}
	defer rows.Close()
	
	var commands []*models.AgentCommand
	
	for rows.Next() {
		var command models.AgentCommand
		var completedAt sql.NullTime
		
		err := rows.Scan(
			&command.ID, &command.AgentID, &command.CommandText, &command.ResponseText,
			&command.Status, &command.ExecutionTimeMs, &command.CreatedAt, &completedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan command: %w", err)
		}
		
		if completedAt.Valid {
			command.CompletedAt = &completedAt.Time
		}
		
		commands = append(commands, &command)
	}
	
	return commands, nil
}

func (r *AgentCommandRepository) GetPendingCommands(limit int) ([]*models.AgentCommand, error) {
	return r.GetByStatus("pending", limit)
}

func (r *AgentCommandRepository) MarkCompleted(id int, responseText string, executionTimeMs int) error {
	now := time.Now()
	
	query := `
		UPDATE agent_commands
		SET response_text = ?, status = 'completed', execution_time_ms = ?, completed_at = ?
		WHERE id = ?
	`
	
	result, err := r.db.Exec(query, responseText, executionTimeMs, now, id)
	if err != nil {
		return fmt.Errorf("failed to mark command completed: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("command not found")
	}
	
	return nil
}

func (r *AgentCommandRepository) MarkFailed(id int, errorMessage string) error {
	now := time.Now()
	
	query := `
		UPDATE agent_commands
		SET response_text = ?, status = 'failed', completed_at = ?
		WHERE id = ?
	`
	
	result, err := r.db.Exec(query, errorMessage, now, id)
	if err != nil {
		return fmt.Errorf("failed to mark command failed: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("command not found")
	}
	
	return nil
}

func (r *AgentCommandRepository) DeleteOldCommands(agentID int, retainCount int) error {
	// Get the ID of the command to keep (retainCount-th newest)
	query := `
		SELECT id FROM agent_commands
		WHERE agent_id = ?
		ORDER BY created_at DESC
		LIMIT 1 OFFSET ?
	`
	
	var cutoffID int
	err := r.db.QueryRow(query, agentID, retainCount-1).Scan(&cutoffID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Less than retainCount commands exist, nothing to delete
			return nil
		}
		return fmt.Errorf("failed to get cutoff ID: %w", err)
	}
	
	// Delete commands older than the cutoff
	deleteQuery := `
		DELETE FROM agent_commands
		WHERE agent_id = ? AND id < ?
	`
	
	result, err := r.db.Exec(deleteQuery, agentID, cutoffID)
	if err != nil {
		return fmt.Errorf("failed to delete old commands: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	fmt.Printf("Deleted %d old commands for agent %d\n", rowsAffected, agentID)
	return nil
}

func (r *AgentCommandRepository) GetRecentCommands(agentID int, since time.Time) ([]*models.AgentCommand, error) {
	query := `
		SELECT id, agent_id, command_text, response_text, status, execution_time_ms, created_at, completed_at
		FROM agent_commands
		WHERE agent_id = ? AND created_at > ?
		ORDER BY created_at DESC
	`
	
	rows, err := r.db.Query(query, agentID, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent commands: %w", err)
	}
	defer rows.Close()
	
	var commands []*models.AgentCommand
	
	for rows.Next() {
		var command models.AgentCommand
		var completedAt sql.NullTime
		
		err := rows.Scan(
			&command.ID, &command.AgentID, &command.CommandText, &command.ResponseText,
			&command.Status, &command.ExecutionTimeMs, &command.CreatedAt, &completedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan command: %w", err)
		}
		
		if completedAt.Valid {
			command.CompletedAt = &completedAt.Time
		}
		
		commands = append(commands, &command)
	}
	
	return commands, nil
}

func (r *AgentCommandRepository) GetStats(agentID int) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// Total commands for this agent
	var totalCommands int
	err := r.db.QueryRow("SELECT COUNT(*) FROM agent_commands WHERE agent_id = ?", agentID).Scan(&totalCommands)
	if err != nil {
		return nil, fmt.Errorf("failed to get total commands: %w", err)
	}
	stats["total_commands"] = totalCommands
	
	// Commands by status for this agent
	statusQuery := `
		SELECT status, COUNT(*) as count
		FROM agent_commands
		WHERE agent_id = ?
		GROUP BY status
		ORDER BY count DESC
	`
	
	rows, err := r.db.Query(statusQuery, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get command status counts: %w", err)
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
	
	// Average execution time for completed commands
	var avgExecutionTime sql.NullFloat64
	err = r.db.QueryRow(`
		SELECT AVG(execution_time_ms)
		FROM agent_commands
		WHERE agent_id = ? AND status = 'completed' AND execution_time_ms > 0
	`, agentID).Scan(&avgExecutionTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get average execution time: %w", err)
	}
	
	if avgExecutionTime.Valid {
		stats["avg_execution_time_ms"] = avgExecutionTime.Float64
	} else {
		stats["avg_execution_time_ms"] = 0
	}
	
	return stats, nil
}