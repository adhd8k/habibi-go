package models

import (
	"encoding/json"
	"fmt"
	"time"
)

type Event struct {
	ID         int                    `json:"id" db:"id"`
	EventType  string                 `json:"event_type" db:"event_type"`
	EntityType string                 `json:"entity_type" db:"entity_type"`
	EntityID   int                    `json:"entity_id" db:"entity_id"`
	Data       map[string]interface{} `json:"data" db:"data"`
	CreatedAt  time.Time              `json:"created_at" db:"created_at"`
}

type EventType string

const (
	// Project events
	EventTypeProjectCreated EventType = "project_created"
	EventTypeProjectUpdated EventType = "project_updated"
	EventTypeProjectDeleted EventType = "project_deleted"
	
	// Session events
	EventTypeSessionCreated   EventType = "session_created"
	EventTypeSessionUpdated   EventType = "session_updated"
	EventTypeSessionDeleted   EventType = "session_deleted"
	EventTypeSessionActivated EventType = "session_activated"
	EventTypeSessionPaused    EventType = "session_paused"
	EventTypeSessionStopped   EventType = "session_stopped"
	
	// Agent events
	EventTypeAgentCreated     EventType = "agent_created"
	EventTypeAgentStarted     EventType = "agent_started"
	EventTypeAgentStopped     EventType = "agent_stopped"
	EventTypeAgentFailed      EventType = "agent_failed"
	EventTypeAgentHeartbeat   EventType = "agent_heartbeat"
	EventTypeAgentCommand     EventType = "agent_command"
	EventTypeAgentResponse    EventType = "agent_response"
	EventTypeAgentFileUpload  EventType = "agent_file_upload"
	EventTypeAgentFileDownload EventType = "agent_file_download"
)

type EntityType string

const (
	EntityTypeProject EntityType = "project"
	EntityTypeSession EntityType = "session"
	EntityTypeAgent   EntityType = "agent"
)

type CreateEventRequest struct {
	EventType  string                 `json:"event_type" binding:"required"`
	EntityType string                 `json:"entity_type" binding:"required"`
	EntityID   int                    `json:"entity_id" binding:"required"`
	Data       map[string]interface{} `json:"data"`
}

func (e *Event) MarshalData() (string, error) {
	if e.Data == nil {
		return "{}", nil
	}
	data, err := json.Marshal(e.Data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal event data: %w", err)
	}
	return string(data), nil
}

func (e *Event) UnmarshalData(dataStr string) error {
	if dataStr == "" {
		e.Data = make(map[string]interface{})
		return nil
	}
	
	if err := json.Unmarshal([]byte(dataStr), &e.Data); err != nil {
		return fmt.Errorf("failed to unmarshal event data: %w", err)
	}
	return nil
}

func (e *Event) Validate() error {
	if e.EventType == "" {
		return fmt.Errorf("event type is required")
	}
	
	if e.EntityType == "" {
		return fmt.Errorf("entity type is required")
	}
	
	if e.EntityID == 0 {
		return fmt.Errorf("entity ID is required")
	}
	
	if !e.IsValidEventType() {
		return fmt.Errorf("invalid event type: %s", e.EventType)
	}
	
	if !e.IsValidEntityType() {
		return fmt.Errorf("invalid entity type: %s", e.EntityType)
	}
	
	return nil
}

func (e *Event) IsValidEventType() bool {
	switch EventType(e.EventType) {
	case EventTypeProjectCreated, EventTypeProjectUpdated, EventTypeProjectDeleted,
		 EventTypeSessionCreated, EventTypeSessionUpdated, EventTypeSessionDeleted,
		 EventTypeSessionActivated, EventTypeSessionPaused, EventTypeSessionStopped,
		 EventTypeAgentCreated, EventTypeAgentStarted, EventTypeAgentStopped,
		 EventTypeAgentFailed, EventTypeAgentHeartbeat, EventTypeAgentCommand,
		 EventTypeAgentResponse, EventTypeAgentFileUpload, EventTypeAgentFileDownload:
		return true
	default:
		return false
	}
}

func (e *Event) IsValidEntityType() bool {
	switch EntityType(e.EntityType) {
	case EntityTypeProject, EntityTypeSession, EntityTypeAgent:
		return true
	default:
		return false
	}
}

func (e *Event) BeforeCreate() {
	e.CreatedAt = time.Now()
	
	if e.Data == nil {
		e.Data = make(map[string]interface{})
	}
}

// Helper functions for creating specific event types
func NewProjectEvent(eventType EventType, projectID int, data map[string]interface{}) *Event {
	return &Event{
		EventType:  string(eventType),
		EntityType: string(EntityTypeProject),
		EntityID:   projectID,
		Data:       data,
	}
}

func NewSessionEvent(eventType EventType, sessionID int, data map[string]interface{}) *Event {
	return &Event{
		EventType:  string(eventType),
		EntityType: string(EntityTypeSession),
		EntityID:   sessionID,
		Data:       data,
	}
}

func NewAgentEvent(eventType EventType, agentID int, data map[string]interface{}) *Event {
	return &Event{
		EventType:  string(eventType),
		EntityType: string(EntityTypeAgent),
		EntityID:   agentID,
		Data:       data,
	}
}

// Agent command and file models
type AgentCommand struct {
	ID              int        `json:"id" db:"id"`
	AgentID         int        `json:"agent_id" db:"agent_id"`
	CommandText     string     `json:"command_text" db:"command_text"`
	ResponseText    string     `json:"response_text" db:"response_text"`
	Status          string     `json:"status" db:"status"`
	ExecutionTimeMs int        `json:"execution_time_ms" db:"execution_time_ms"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	CompletedAt     *time.Time `json:"completed_at" db:"completed_at"`
}

type AgentFile struct {
	ID        int       `json:"id" db:"id"`
	AgentID   int       `json:"agent_id" db:"agent_id"`
	Filename  string    `json:"filename" db:"filename"`
	FilePath  string    `json:"file_path" db:"file_path"`
	FileSize  int64     `json:"file_size" db:"file_size"`
	MimeType  string    `json:"mime_type" db:"mime_type"`
	Direction string    `json:"direction" db:"direction"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}