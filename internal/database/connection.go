package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
}

func New(dbPath string) (*DB, error) {
	// Create database directory if it doesn't exist
	if err := createDatabaseDir(dbPath); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}
	
	// Open database connection
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	
	// Configure connection pool
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	
	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	
	// Set SQLite pragmas to handle concurrency better
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA busy_timeout=5000",
		"PRAGMA synchronous=NORMAL",
	}
	
	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			return nil, fmt.Errorf("failed to set pragma %s: %w", pragma, err)
		}
	}
	
	return &DB{db}, nil
}

func (db *DB) RunMigrations() error {
	// For now, let's run migrations manually instead of using the migrate library
	// This will avoid file path issues during development
	
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS projects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			path TEXT NOT NULL,
			repository_url TEXT,
			default_branch TEXT DEFAULT 'main',
			setup_command TEXT,
			config TEXT DEFAULT '{}',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			project_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			branch_name TEXT NOT NULL,
			worktree_path TEXT NOT NULL,
			status TEXT DEFAULT 'active' CHECK(status IN ('active', 'paused', 'stopped')),
			config TEXT DEFAULT '{}',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_used_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
			UNIQUE(project_id, name)
		)`,
		`CREATE TABLE IF NOT EXISTS agents (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id INTEGER NOT NULL,
			agent_type TEXT NOT NULL,
			pid INTEGER,
			status TEXT DEFAULT 'starting' CHECK(status IN ('starting', 'running', 'stopped', 'failed')),
			config TEXT DEFAULT '{}',
			command TEXT NOT NULL,
			working_directory TEXT NOT NULL,
			communication_method TEXT DEFAULT 'stdio',
			input_pipe_path TEXT,
			output_pipe_path TEXT,
			last_heartbeat DATETIME,
			resource_usage TEXT DEFAULT '{}',
			started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			stopped_at DATETIME,
			claude_session_id TEXT,
			FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS agent_commands (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			agent_id INTEGER NOT NULL,
			command_text TEXT NOT NULL,
			response_text TEXT,
			status TEXT DEFAULT 'pending' CHECK(status IN ('pending', 'completed', 'failed')),
			execution_time_ms INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME,
			FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS agent_files (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			agent_id INTEGER NOT NULL,
			filename TEXT NOT NULL,
			file_path TEXT NOT NULL,
			file_size INTEGER,
			mime_type TEXT,
			direction TEXT CHECK(direction IN ('upload', 'download')),
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			event_type TEXT NOT NULL,
			entity_type TEXT NOT NULL,
			entity_id INTEGER NOT NULL,
			data TEXT DEFAULT '{}',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_project_id ON sessions(project_id)`,
		`CREATE INDEX IF NOT EXISTS idx_agents_session_id ON agents(session_id)`,
		`CREATE INDEX IF NOT EXISTS idx_agent_commands_agent_id ON agent_commands(agent_id)`,
		`CREATE INDEX IF NOT EXISTS idx_agent_files_agent_id ON agent_files(agent_id)`,
		`CREATE INDEX IF NOT EXISTS idx_events_created_at ON events(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_events_entity ON events(entity_type, entity_id)`,
		// Add setup_command column if it doesn't exist
		`ALTER TABLE projects ADD COLUMN setup_command TEXT`,
		// Add claude_session_id column if it doesn't exist
		`ALTER TABLE agents ADD COLUMN claude_session_id TEXT`,
		// Add chat messages table
		`CREATE TABLE IF NOT EXISTS chat_messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			agent_id INTEGER NOT NULL,
			role TEXT NOT NULL CHECK(role IN ('user', 'assistant', 'system')),
			content TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_chat_messages_agent_id ON chat_messages(agent_id)`,
	}
	
	for i, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			// Ignore error for ALTER TABLE if column already exists
			if i >= len(migrations)-2 && (strings.Contains(err.Error(), "duplicate column name: setup_command") || 
				strings.Contains(err.Error(), "duplicate column name: claude_session_id")) {
				continue
			}
			return fmt.Errorf("failed to execute migration %d: %w", i, err)
		}
	}
	
	return nil
}

func (db *DB) Close() error {
	return db.DB.Close()
}

func createDatabaseDir(dbPath string) error {
	dir := filepath.Dir(dbPath)
	return os.MkdirAll(dir, 0755)
}