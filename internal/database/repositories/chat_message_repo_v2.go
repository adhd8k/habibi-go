package repositories

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"habibi-go/internal/models"
)

// ChatMessageV2Repository handles database operations for chat messages linked to sessions
type ChatMessageV2Repository struct {
	db *sql.DB
}

// NewChatMessageV2Repository creates a new chat message repository
func NewChatMessageV2Repository(db *sql.DB) *ChatMessageV2Repository {
	return &ChatMessageV2Repository{db: db}
}

// Create inserts a new chat message
func (r *ChatMessageV2Repository) Create(message *models.ChatMessage) error {
	// Handle tool metadata
	var toolInput, toolContent sql.NullString
	
	if message.ToolInput != nil {
		data, err := json.Marshal(message.ToolInput)
		if err != nil {
			return fmt.Errorf("failed to marshal tool input: %w", err)
		}
		toolInput = sql.NullString{String: string(data), Valid: true}
	}
	
	if message.ToolContent != nil {
		data, err := json.Marshal(message.ToolContent)
		if err != nil {
			return fmt.Errorf("failed to marshal tool content: %w", err)
		}
		toolContent = sql.NullString{String: string(data), Valid: true}
	}

	result, err := r.db.Exec(
		`INSERT INTO chat_messages (session_id, role, content, tool_name, tool_input, tool_use_id, tool_content)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		message.SessionID,
		message.Role,
		message.Content,
		sql.NullString{String: message.ToolName, Valid: message.ToolName != ""},
		toolInput,
		sql.NullString{String: message.ToolUseID, Valid: message.ToolUseID != ""},
		toolContent,
	)
	if err != nil {
		return fmt.Errorf("failed to insert chat message: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	message.ID = int(id)
	message.CreatedAt = time.Now()
	return nil
}

// GetBySessionID retrieves all messages for a session
func (r *ChatMessageV2Repository) GetBySessionID(sessionID int, limit int) ([]*models.ChatMessage, error) {
	query := `
		SELECT id, session_id, role, content, created_at, 
		       tool_name, tool_input, tool_use_id, tool_content
		FROM chat_messages
		WHERE session_id = ?
		ORDER BY created_at DESC, id DESC
		LIMIT ?
	`

	rows, err := r.db.Query(query, sessionID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query chat messages: %w", err)
	}
	defer rows.Close()

	var messages []*models.ChatMessage
	for rows.Next() {
		msg := &models.ChatMessage{}
		var toolName, toolInput, toolUseID, toolContent sql.NullString
		
		err := rows.Scan(
			&msg.ID,
			&msg.SessionID,
			&msg.Role,
			&msg.Content,
			&msg.CreatedAt,
			&toolName,
			&toolInput,
			&toolUseID,
			&toolContent,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan chat message: %w", err)
		}

		// Handle tool metadata
		if toolName.Valid {
			msg.ToolName = toolName.String
		}
		if toolUseID.Valid {
			msg.ToolUseID = toolUseID.String
		}
		if toolInput.Valid {
			if err := json.Unmarshal([]byte(toolInput.String), &msg.ToolInput); err != nil {
				// If unmarshal fails, store as string
				msg.ToolInput = toolInput.String
			}
		}
		if toolContent.Valid {
			if err := json.Unmarshal([]byte(toolContent.String), &msg.ToolContent); err != nil {
				// If unmarshal fails, store as string
				msg.ToolContent = toolContent.String
			}
		}

		messages = append(messages, msg)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating chat messages: %w", err)
	}

	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// GetByID retrieves a specific message by ID
func (r *ChatMessageV2Repository) GetByID(id int) (*models.ChatMessage, error) {
	msg := &models.ChatMessage{}
	var toolName, toolInput, toolUseID, toolContent sql.NullString
	
	err := r.db.QueryRow(`
		SELECT id, session_id, role, content, created_at, 
		       tool_name, tool_input, tool_use_id, tool_content
		FROM chat_messages
		WHERE id = ?
	`, id).Scan(
		&msg.ID,
		&msg.SessionID,
		&msg.Role,
		&msg.Content,
		&msg.CreatedAt,
		&toolName,
		&toolInput,
		&toolUseID,
		&toolContent,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get chat message by ID: %w", err)
	}

	// Handle tool metadata
	if toolName.Valid {
		msg.ToolName = toolName.String
	}
	if toolUseID.Valid {
		msg.ToolUseID = toolUseID.String
	}
	if toolInput.Valid {
		if err := json.Unmarshal([]byte(toolInput.String), &msg.ToolInput); err != nil {
			msg.ToolInput = toolInput.String
		}
	}
	if toolContent.Valid {
		if err := json.Unmarshal([]byte(toolContent.String), &msg.ToolContent); err != nil {
			msg.ToolContent = toolContent.String
		}
	}

	return msg, nil
}

// UpdateContent updates the content of an existing message
func (r *ChatMessageV2Repository) UpdateContent(messageID int, content string) error {
	_, err := r.db.Exec("UPDATE chat_messages SET content = ? WHERE id = ?", content, messageID)
	if err != nil {
		return fmt.Errorf("failed to update chat message content: %w", err)
	}
	return nil
}

// DeleteBySessionID deletes all messages for a session
func (r *ChatMessageV2Repository) DeleteBySessionID(sessionID int) error {
	_, err := r.db.Exec("DELETE FROM chat_messages WHERE session_id = ?", sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete chat messages: %w", err)
	}
	return nil
}