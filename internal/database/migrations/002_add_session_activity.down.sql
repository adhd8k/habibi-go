-- Remove session activity tracking columns
DROP INDEX IF EXISTS idx_sessions_last_activity_at;
DROP INDEX IF EXISTS idx_sessions_activity_status;

-- Note: SQLite doesn't support DROP COLUMN directly
-- In production, you would need to recreate the table without these columns
-- For now, we'll leave the columns in place as they're nullable