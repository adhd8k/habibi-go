-- Add session activity tracking columns
ALTER TABLE sessions ADD COLUMN last_activity_at DATETIME;
ALTER TABLE sessions ADD COLUMN activity_status TEXT DEFAULT 'idle' CHECK(activity_status IN ('idle', 'streaming', 'new', 'viewed'));
ALTER TABLE sessions ADD COLUMN last_viewed_at DATETIME;

-- Create index for activity tracking
CREATE INDEX idx_sessions_activity_status ON sessions(activity_status);
CREATE INDEX idx_sessions_last_activity_at ON sessions(last_activity_at);