-- Migration to simplify agent architecture
-- This changes chat_messages to reference sessions directly instead of agents

-- 1. Create new chat_messages table with session_id
CREATE TABLE chat_messages_v2 (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id INTEGER NOT NULL,
    role TEXT NOT NULL CHECK(role IN ('user', 'assistant', 'system', 'tool_use', 'tool_result')),
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    tool_name TEXT,
    tool_input TEXT,
    tool_use_id TEXT,
    tool_content TEXT,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

-- 2. Copy existing chat messages, linking to sessions through agents
INSERT INTO chat_messages_v2 (id, session_id, role, content, created_at, tool_name, tool_input, tool_use_id, tool_content)
SELECT 
    cm.id,
    a.session_id,
    cm.role,
    cm.content,
    cm.created_at,
    cm.tool_name,
    cm.tool_input,
    cm.tool_use_id,
    cm.tool_content
FROM chat_messages cm
JOIN agents a ON cm.agent_id = a.id;

-- 3. Drop old tables
DROP TABLE IF EXISTS agent_commands;
DROP TABLE IF EXISTS chat_messages;
DROP TABLE IF EXISTS agents;

-- 4. Rename new table
ALTER TABLE chat_messages_v2 RENAME TO chat_messages;

-- 5. Create index
CREATE INDEX idx_chat_messages_session_id ON chat_messages(session_id);