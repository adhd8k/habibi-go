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

// Helper function to check if a column exists
func (db *DB) columnExists(table, column string) bool {
	query := fmt.Sprintf(`SELECT COUNT(*) FROM pragma_table_info('%s') WHERE name = ?`, table)
	var count int
	err := db.QueryRow(query, column).Scan(&count)
	return err == nil && count > 0
}

// Helper function to add column if it doesn't exist
func (db *DB) addColumnIfNotExists(table, column, columnDef string) error {
	if !db.columnExists(table, column) {
		query := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, columnDef)
		_, err := db.Exec(query)
		return err
	}
	return nil
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
		// Add chat messages table
		`CREATE TABLE IF NOT EXISTS chat_messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			agent_id INTEGER NOT NULL,
			role TEXT NOT NULL CHECK(role IN ('user', 'assistant', 'system', 'tool_use', 'tool_result')),
			content TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_chat_messages_agent_id ON chat_messages(agent_id)`,
	}
	
	for i, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("failed to execute migration %d: %w", i, err)
		}
	}
	
	// Add columns if they don't exist (for existing databases)
	if err := db.addColumnIfNotExists("projects", "setup_command", "TEXT"); err != nil {
		return fmt.Errorf("failed to add setup_command column: %w", err)
	}
	
	if err := db.addColumnIfNotExists("agents", "claude_session_id", "TEXT"); err != nil {
		return fmt.Errorf("failed to add claude_session_id column: %w", err)
	}
	
	// Add session activity tracking columns
	if err := db.addColumnIfNotExists("sessions", "last_activity_at", "DATETIME"); err != nil {
		return fmt.Errorf("failed to add last_activity_at column: %w", err)
	}
	
	if err := db.addColumnIfNotExists("sessions", "activity_status", "TEXT DEFAULT 'idle' CHECK(activity_status IN ('idle', 'streaming', 'new', 'viewed'))"); err != nil {
		return fmt.Errorf("failed to add activity_status column: %w", err)
	}
	
	if err := db.addColumnIfNotExists("sessions", "last_viewed_at", "DATETIME"); err != nil {
		return fmt.Errorf("failed to add last_viewed_at column: %w", err)
	}
	
	// Add tool metadata columns for chat messages
	if err := db.addColumnIfNotExists("chat_messages", "tool_name", "TEXT"); err != nil {
		return fmt.Errorf("failed to add tool_name column: %w", err)
	}
	
	if err := db.addColumnIfNotExists("chat_messages", "tool_input", "TEXT"); err != nil {
		return fmt.Errorf("failed to add tool_input column: %w", err)
	}
	
	if err := db.addColumnIfNotExists("chat_messages", "tool_use_id", "TEXT"); err != nil {
		return fmt.Errorf("failed to add tool_use_id column: %w", err)
	}
	
	if err := db.addColumnIfNotExists("chat_messages", "tool_content", "TEXT"); err != nil {
		return fmt.Errorf("failed to add tool_content column: %w", err)
	}
	
	// Fix the role constraint for existing databases
	// SQLite doesn't support ALTER TABLE to modify constraints, so we need to recreate the table
	if err := db.fixChatMessagesRoleConstraint(); err != nil {
		return fmt.Errorf("failed to fix chat messages role constraint: %w", err)
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

// fixChatMessagesRoleConstraint fixes the role constraint to include tool_use and tool_result
func (db *DB) fixChatMessagesRoleConstraint() error {
	// Check if the constraint already includes tool_use
	var constraintCheck string
	err := db.QueryRow(`
		SELECT sql FROM sqlite_master 
		WHERE type='table' AND name='chat_messages'
	`).Scan(&constraintCheck)
	
	if err != nil {
		return fmt.Errorf("failed to check chat_messages table: %w", err)
	}
	
	// If the constraint already includes tool_use, we're good
	if strings.Contains(constraintCheck, "tool_use") {
		return nil
	}
	
	// Otherwise, we need to recreate the table with the updated constraint
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	// Create new table with updated constraint
	_, err = tx.Exec(`
		CREATE TABLE chat_messages_new (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			agent_id INTEGER NOT NULL,
			role TEXT NOT NULL CHECK(role IN ('user', 'assistant', 'system', 'tool_use', 'tool_result')),
			content TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			tool_name TEXT,
			tool_input TEXT,
			tool_use_id TEXT,
			tool_content TEXT,
			FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create new chat_messages table: %w", err)
	}
	
	// Copy data from old table
	_, err = tx.Exec(`
		INSERT INTO chat_messages_new (id, agent_id, role, content, created_at, tool_name, tool_input, tool_use_id, tool_content)
		SELECT id, agent_id, role, content, created_at, tool_name, tool_input, tool_use_id, tool_content
		FROM chat_messages
	`)
	if err != nil {
		return fmt.Errorf("failed to copy chat messages: %w", err)
	}
	
	// Drop old table
	_, err = tx.Exec(`DROP TABLE chat_messages`)
	if err != nil {
		return fmt.Errorf("failed to drop old chat_messages table: %w", err)
	}
	
	// Rename new table
	_, err = tx.Exec(`ALTER TABLE chat_messages_new RENAME TO chat_messages`)
	if err != nil {
		return fmt.Errorf("failed to rename chat_messages table: %w", err)
	}
	
	// Recreate index
	_, err = tx.Exec(`CREATE INDEX IF NOT EXISTS idx_chat_messages_agent_id ON chat_messages(agent_id)`)
	if err != nil {
		return fmt.Errorf("failed to recreate index: %w", err)
	}
	
	return tx.Commit()
}