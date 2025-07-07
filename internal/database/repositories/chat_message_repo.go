package repositories

import (
	"database/sql"
	"fmt"

	"habibi-go/internal/models"
)

type ChatMessageRepository struct {
	db *sql.DB
}

func NewChatMessageRepository(db *sql.DB) *ChatMessageRepository {
	return &ChatMessageRepository{db: db}
}

func (r *ChatMessageRepository) Create(message *models.ChatMessage) error {
	query := `
		INSERT INTO chat_messages (agent_id, role, content, created_at)
		VALUES (?, ?, ?, ?)
	`
	
	result, err := r.db.Exec(query, message.AgentID, message.Role, message.Content, message.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create chat message: %w", err)
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get message ID: %w", err)
	}
	
	message.ID = int(id)
	return nil
}

func (r *ChatMessageRepository) GetByAgentID(agentID int, limit int) ([]*models.ChatMessage, error) {
	query := `
		SELECT id, agent_id, role, content, created_at
		FROM chat_messages
		WHERE agent_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`
	
	rows, err := r.db.Query(query, agentID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat messages: %w", err)
	}
	defer rows.Close()
	
	var messages []*models.ChatMessage
	
	for rows.Next() {
		var message models.ChatMessage
		
		err := rows.Scan(
			&message.ID, &message.AgentID, &message.Role,
			&message.Content, &message.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		
		messages = append(messages, &message)
	}
	
	// Reverse the messages to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	
	return messages, nil
}

func (r *ChatMessageRepository) GetAll(agentID int) ([]*models.ChatMessage, error) {
	query := `
		SELECT id, agent_id, role, content, created_at
		FROM chat_messages
		WHERE agent_id = ?
		ORDER BY created_at ASC
	`
	
	rows, err := r.db.Query(query, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get all chat messages: %w", err)
	}
	defer rows.Close()
	
	var messages []*models.ChatMessage
	
	for rows.Next() {
		var message models.ChatMessage
		
		err := rows.Scan(
			&message.ID, &message.AgentID, &message.Role,
			&message.Content, &message.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		
		messages = append(messages, &message)
	}
	
	return messages, nil
}

func (r *ChatMessageRepository) DeleteByAgentID(agentID int) error {
	query := `DELETE FROM chat_messages WHERE agent_id = ?`
	
	_, err := r.db.Exec(query, agentID)
	if err != nil {
		return fmt.Errorf("failed to delete chat messages: %w", err)
	}
	
	return nil
}