-- Drop indexes
DROP INDEX IF EXISTS idx_events_entity;
DROP INDEX IF EXISTS idx_events_created_at;
DROP INDEX IF EXISTS idx_agent_files_agent_id;
DROP INDEX IF EXISTS idx_agent_commands_agent_id;
DROP INDEX IF EXISTS idx_agents_session_id;
DROP INDEX IF EXISTS idx_sessions_project_id;

-- Drop tables in reverse order (respecting foreign key constraints)
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS agent_files;
DROP TABLE IF EXISTS agent_commands;
DROP TABLE IF EXISTS agents;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS projects;