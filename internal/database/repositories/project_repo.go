package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"habibi-go/internal/models"
)

type ProjectRepository struct {
	db *sql.DB
}

func NewProjectRepository(db *sql.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Create(project *models.Project) error {
	project.BeforeCreate()
	
	configStr, err := project.MarshalConfig()
	if err != nil {
		return err
	}
	
	query := `
		INSERT INTO projects (name, path, repository_url, default_branch, config, setup_command, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	result, err := r.db.Exec(query, project.Name, project.Path, project.RepositoryURL, 
		project.DefaultBranch, configStr, project.SetupCommand, project.CreatedAt, project.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get project ID: %w", err)
	}
	
	project.ID = int(id)
	return nil
}

func (r *ProjectRepository) GetByID(id int) (*models.Project, error) {
	query := `
		SELECT id, name, path, repository_url, default_branch, config, setup_command, created_at, updated_at
		FROM projects
		WHERE id = ?
	`
	
	var project models.Project
	var configStr string
	
	err := r.db.QueryRow(query, id).Scan(
		&project.ID, &project.Name, &project.Path, &project.RepositoryURL,
		&project.DefaultBranch, &configStr, &project.SetupCommand, &project.CreatedAt, &project.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("project not found")
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	
	if err := project.UnmarshalConfig(configStr); err != nil {
		return nil, err
	}
	
	return &project, nil
}

func (r *ProjectRepository) GetByName(name string) (*models.Project, error) {
	query := `
		SELECT id, name, path, repository_url, default_branch, config, setup_command, created_at, updated_at
		FROM projects
		WHERE name = ?
	`
	
	var project models.Project
	var configStr string
	
	err := r.db.QueryRow(query, name).Scan(
		&project.ID, &project.Name, &project.Path, &project.RepositoryURL,
		&project.DefaultBranch, &configStr, &project.SetupCommand, &project.CreatedAt, &project.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("project not found")
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	
	if err := project.UnmarshalConfig(configStr); err != nil {
		return nil, err
	}
	
	return &project, nil
}

func (r *ProjectRepository) GetAll() ([]*models.Project, error) {
	query := `
		SELECT id, name, path, repository_url, default_branch, config, setup_command, created_at, updated_at
		FROM projects
		ORDER BY name
	`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}
	defer rows.Close()
	
	var projects []*models.Project
	
	for rows.Next() {
		var project models.Project
		var configStr string
		
		err := rows.Scan(
			&project.ID, &project.Name, &project.Path, &project.RepositoryURL,
			&project.DefaultBranch, &configStr, &project.SetupCommand, &project.CreatedAt, &project.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		
		if err := project.UnmarshalConfig(configStr); err != nil {
			return nil, err
		}
		
		projects = append(projects, &project)
	}
	
	return projects, nil
}

func (r *ProjectRepository) Update(project *models.Project) error {
	project.BeforeUpdate()
	
	configStr, err := project.MarshalConfig()
	if err != nil {
		return err
	}
	
	query := `
		UPDATE projects
		SET name = ?, path = ?, repository_url = ?, default_branch = ?, config = ?, setup_command = ?, updated_at = ?
		WHERE id = ?
	`
	
	result, err := r.db.Exec(query, project.Name, project.Path, project.RepositoryURL,
		project.DefaultBranch, configStr, project.SetupCommand, project.UpdatedAt, project.ID)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("project not found")
	}
	
	return nil
}

func (r *ProjectRepository) Delete(id int) error {
	query := `DELETE FROM projects WHERE id = ?`
	
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("project not found")
	}
	
	return nil
}

func (r *ProjectRepository) Exists(name string) (bool, error) {
	query := `SELECT COUNT(*) FROM projects WHERE name = ?`
	
	var count int
	err := r.db.QueryRow(query, name).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check project existence: %w", err)
	}
	
	return count > 0, nil
}

func (r *ProjectRepository) GetWithSessions(id int) (*models.Project, error) {
	project, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}
	
	// Get sessions for this project
	sessionQuery := `
		SELECT id, project_id, name, branch_name, worktree_path, status, config, created_at, last_used_at
		FROM sessions
		WHERE project_id = ?
		ORDER BY last_used_at DESC
	`
	
	rows, err := r.db.Query(sessionQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get project sessions: %w", err)
	}
	defer rows.Close()
	
	var sessions []models.Session
	
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
		
		sessions = append(sessions, session)
	}
	
	project.Sessions = sessions
	return project, nil
}

func (r *ProjectRepository) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// Total projects
	var totalProjects int
	err := r.db.QueryRow("SELECT COUNT(*) FROM projects").Scan(&totalProjects)
	if err != nil {
		return nil, fmt.Errorf("failed to get total projects: %w", err)
	}
	stats["total_projects"] = totalProjects
	
	// Projects with sessions
	var projectsWithSessions int
	err = r.db.QueryRow(`
		SELECT COUNT(DISTINCT p.id)
		FROM projects p
		JOIN sessions s ON p.id = s.project_id
	`).Scan(&projectsWithSessions)
	if err != nil {
		return nil, fmt.Errorf("failed to get projects with sessions: %w", err)
	}
	stats["projects_with_sessions"] = projectsWithSessions
	
	// Most recently created project
	var recentProject string
	var recentTime time.Time
	err = r.db.QueryRow(`
		SELECT name, created_at
		FROM projects
		ORDER BY created_at DESC
		LIMIT 1
	`).Scan(&recentProject, &recentTime)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get recent project: %w", err)
	}
	if err != sql.ErrNoRows {
		stats["most_recent_project"] = recentProject
		stats["most_recent_project_time"] = recentTime
	}
	
	return stats, nil
}