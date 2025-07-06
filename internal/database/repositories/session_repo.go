package repositories

import (
	"database/sql"
	"fmt"

	"habibi-go/internal/models"
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(session *models.Session) error {
	session.BeforeCreate()
	
	configStr, err := session.MarshalConfig()
	if err != nil {
		return err
	}
	
	query := `
		INSERT INTO sessions (project_id, name, branch_name, worktree_path, status, config, created_at, last_used_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	result, err := r.db.Exec(query, session.ProjectID, session.Name, session.BranchName,
		session.WorktreePath, session.Status, configStr, session.CreatedAt, session.LastUsedAt)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get session ID: %w", err)
	}
	
	session.ID = int(id)
	return nil
}

func (r *SessionRepository) GetByID(id int) (*models.Session, error) {
	query := `
		SELECT id, project_id, name, branch_name, worktree_path, status, config, created_at, last_used_at
		FROM sessions
		WHERE id = ?
	`
	
	var session models.Session
	var configStr string
	
	err := r.db.QueryRow(query, id).Scan(
		&session.ID, &session.ProjectID, &session.Name, &session.BranchName,
		&session.WorktreePath, &session.Status, &configStr, &session.CreatedAt,
		&session.LastUsedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	
	if err := session.UnmarshalConfig(configStr); err != nil {
		return nil, err
	}
	
	return &session, nil
}

func (r *SessionRepository) GetByProjectAndName(projectID int, name string) (*models.Session, error) {
	query := `
		SELECT id, project_id, name, branch_name, worktree_path, status, config, created_at, last_used_at
		FROM sessions
		WHERE project_id = ? AND name = ?
	`
	
	var session models.Session
	var configStr string
	
	err := r.db.QueryRow(query, projectID, name).Scan(
		&session.ID, &session.ProjectID, &session.Name, &session.BranchName,
		&session.WorktreePath, &session.Status, &configStr, &session.CreatedAt,
		&session.LastUsedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	
	if err := session.UnmarshalConfig(configStr); err != nil {
		return nil, err
	}
	
	return &session, nil
}

func (r *SessionRepository) GetByProjectID(projectID int) ([]*models.Session, error) {
	query := `
		SELECT id, project_id, name, branch_name, worktree_path, status, config, created_at, last_used_at
		FROM sessions
		WHERE project_id = ?
		ORDER BY last_used_at DESC
	`
	
	rows, err := r.db.Query(query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}
	defer rows.Close()
	
	var sessions []*models.Session
	
	for rows.Next() {
		var session models.Session
		var configStr string
		
		err := rows.Scan(
			&session.ID, &session.ProjectID, &session.Name, &session.BranchName,
			&session.WorktreePath, &session.Status, &configStr, &session.CreatedAt,
			&session.LastUsedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		
		if err := session.UnmarshalConfig(configStr); err != nil {
			return nil, err
		}
		
		sessions = append(sessions, &session)
	}
	
	return sessions, nil
}

func (r *SessionRepository) GetAll() ([]*models.Session, error) {
	query := `
		SELECT id, project_id, name, branch_name, worktree_path, status, config, created_at, last_used_at
		FROM sessions
		ORDER BY last_used_at DESC
	`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}
	defer rows.Close()
	
	var sessions []*models.Session
	
	for rows.Next() {
		var session models.Session
		var configStr string
		
		err := rows.Scan(
			&session.ID, &session.ProjectID, &session.Name, &session.BranchName,
			&session.WorktreePath, &session.Status, &configStr, &session.CreatedAt,
			&session.LastUsedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		
		if err := session.UnmarshalConfig(configStr); err != nil {
			return nil, err
		}
		
		sessions = append(sessions, &session)
	}
	
	return sessions, nil
}

func (r *SessionRepository) Update(session *models.Session) error {
	session.BeforeUpdate()
	
	configStr, err := session.MarshalConfig()
	if err != nil {
		return err
	}
	
	query := `
		UPDATE sessions
		SET project_id = ?, name = ?, branch_name = ?, worktree_path = ?, status = ?, config = ?, last_used_at = ?
		WHERE id = ?
	`
	
	result, err := r.db.Exec(query, session.ProjectID, session.Name, session.BranchName,
		session.WorktreePath, session.Status, configStr, session.LastUsedAt, session.ID)
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("session not found")
	}
	
	return nil
}

func (r *SessionRepository) Delete(id int) error {
	query := `DELETE FROM sessions WHERE id = ?`
	
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("session not found")
	}
	
	return nil
}

func (r *SessionRepository) Exists(projectID int, name string) (bool, error) {
	query := `SELECT COUNT(*) FROM sessions WHERE project_id = ? AND name = ?`
	
	var count int
	err := r.db.QueryRow(query, projectID, name).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check session existence: %w", err)
	}
	
	return count > 0, nil
}

func (r *SessionRepository) UpdateLastUsed(id int) error {
	query := `UPDATE sessions SET last_used_at = CURRENT_TIMESTAMP WHERE id = ?`
	
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to update last used: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("session not found")
	}
	
	return nil
}

func (r *SessionRepository) GetActiveByProject(projectID int) ([]*models.Session, error) {
	query := `
		SELECT id, project_id, name, branch_name, worktree_path, status, config, created_at, last_used_at
		FROM sessions
		WHERE project_id = ? AND status = 'active'
		ORDER BY last_used_at DESC
	`
	
	rows, err := r.db.Query(query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active sessions: %w", err)
	}
	defer rows.Close()
	
	var sessions []*models.Session
	
	for rows.Next() {
		var session models.Session
		var configStr string
		
		err := rows.Scan(
			&session.ID, &session.ProjectID, &session.Name, &session.BranchName,
			&session.WorktreePath, &session.Status, &configStr, &session.CreatedAt,
			&session.LastUsedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		
		if err := session.UnmarshalConfig(configStr); err != nil {
			return nil, err
		}
		
		sessions = append(sessions, &session)
	}
	
	return sessions, nil
}

func (r *SessionRepository) GetByStatus(status string) ([]*models.Session, error) {
	query := `
		SELECT id, project_id, name, branch_name, worktree_path, status, config, created_at, last_used_at
		FROM sessions
		WHERE status = ?
		ORDER BY last_used_at DESC
	`
	
	rows, err := r.db.Query(query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions by status: %w", err)
	}
	defer rows.Close()
	
	var sessions []*models.Session
	
	for rows.Next() {
		var session models.Session
		var configStr string
		
		err := rows.Scan(
			&session.ID, &session.ProjectID, &session.Name, &session.BranchName,
			&session.WorktreePath, &session.Status, &configStr, &session.CreatedAt,
			&session.LastUsedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		
		if err := session.UnmarshalConfig(configStr); err != nil {
			return nil, err
		}
		
		sessions = append(sessions, &session)
	}
	
	return sessions, nil
}

func (r *SessionRepository) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// Total sessions
	var totalSessions int
	err := r.db.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&totalSessions)
	if err != nil {
		return nil, fmt.Errorf("failed to get total sessions: %w", err)
	}
	stats["total_sessions"] = totalSessions
	
	// Sessions by status
	statusQuery := `
		SELECT status, COUNT(*) as count
		FROM sessions
		GROUP BY status
		ORDER BY count DESC
	`
	
	rows, err := r.db.Query(statusQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get session status counts: %w", err)
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
	
	// Active sessions count
	var activeSessions int
	err = r.db.QueryRow("SELECT COUNT(*) FROM sessions WHERE status = 'active'").Scan(&activeSessions)
	if err != nil {
		return nil, fmt.Errorf("failed to get active sessions count: %w", err)
	}
	stats["active_sessions"] = activeSessions
	
	return stats, nil
}