package repositories

import (
	"database/sql"
	"fmt"

	"habibi-go/internal/models"
)

type EventRepository struct {
	db *sql.DB
}

func NewEventRepository(db *sql.DB) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) Create(event *models.Event) error {
	event.BeforeCreate()
	
	dataStr, err := event.MarshalData()
	if err != nil {
		return err
	}
	
	query := `
		INSERT INTO events (event_type, entity_type, entity_id, data, created_at)
		VALUES (?, ?, ?, ?, ?)
	`
	
	result, err := r.db.Exec(query, event.EventType, event.EntityType, 
		event.EntityID, dataStr, event.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get event ID: %w", err)
	}
	
	event.ID = int(id)
	return nil
}

func (r *EventRepository) GetByID(id int) (*models.Event, error) {
	query := `
		SELECT id, event_type, entity_type, entity_id, data, created_at
		FROM events
		WHERE id = ?
	`
	
	var event models.Event
	var dataStr string
	
	err := r.db.QueryRow(query, id).Scan(
		&event.ID, &event.EventType, &event.EntityType, &event.EntityID,
		&dataStr, &event.CreatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("event not found")
		}
		return nil, fmt.Errorf("failed to get event: %w", err)
	}
	
	if err := event.UnmarshalData(dataStr); err != nil {
		return nil, err
	}
	
	return &event, nil
}

func (r *EventRepository) GetByEntity(entityType string, entityID int, limit int) ([]*models.Event, error) {
	query := `
		SELECT id, event_type, entity_type, entity_id, data, created_at
		FROM events
		WHERE entity_type = ? AND entity_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`
	
	rows, err := r.db.Query(query, entityType, entityID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}
	defer rows.Close()
	
	var events []*models.Event
	
	for rows.Next() {
		var event models.Event
		var dataStr string
		
		err := rows.Scan(
			&event.ID, &event.EventType, &event.EntityType, &event.EntityID,
			&dataStr, &event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		
		if err := event.UnmarshalData(dataStr); err != nil {
			return nil, err
		}
		
		events = append(events, &event)
	}
	
	return events, nil
}

func (r *EventRepository) GetRecent(limit int) ([]*models.Event, error) {
	query := `
		SELECT id, event_type, entity_type, entity_id, data, created_at
		FROM events
		ORDER BY created_at DESC
		LIMIT ?
	`
	
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent events: %w", err)
	}
	defer rows.Close()
	
	var events []*models.Event
	
	for rows.Next() {
		var event models.Event
		var dataStr string
		
		err := rows.Scan(
			&event.ID, &event.EventType, &event.EntityType, &event.EntityID,
			&dataStr, &event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		
		if err := event.UnmarshalData(dataStr); err != nil {
			return nil, err
		}
		
		events = append(events, &event)
	}
	
	return events, nil
}

func (r *EventRepository) GetByType(eventType string, limit int) ([]*models.Event, error) {
	query := `
		SELECT id, event_type, entity_type, entity_id, data, created_at
		FROM events
		WHERE event_type = ?
		ORDER BY created_at DESC
		LIMIT ?
	`
	
	rows, err := r.db.Query(query, eventType, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get events by type: %w", err)
	}
	defer rows.Close()
	
	var events []*models.Event
	
	for rows.Next() {
		var event models.Event
		var dataStr string
		
		err := rows.Scan(
			&event.ID, &event.EventType, &event.EntityType, &event.EntityID,
			&dataStr, &event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		
		if err := event.UnmarshalData(dataStr); err != nil {
			return nil, err
		}
		
		events = append(events, &event)
	}
	
	return events, nil
}

func (r *EventRepository) DeleteOldEvents(daysOld int) error {
	query := `
		DELETE FROM events
		WHERE created_at < datetime('now', '-' || ? || ' days')
	`
	
	result, err := r.db.Exec(query, daysOld)
	if err != nil {
		return fmt.Errorf("failed to delete old events: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	fmt.Printf("Deleted %d old events\n", rowsAffected)
	return nil
}

func (r *EventRepository) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// Total events
	var totalEvents int
	err := r.db.QueryRow("SELECT COUNT(*) FROM events").Scan(&totalEvents)
	if err != nil {
		return nil, fmt.Errorf("failed to get total events: %w", err)
	}
	stats["total_events"] = totalEvents
	
	// Events by type
	eventTypeQuery := `
		SELECT event_type, COUNT(*) as count
		FROM events
		GROUP BY event_type
		ORDER BY count DESC
	`
	
	rows, err := r.db.Query(eventTypeQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get event types: %w", err)
	}
	defer rows.Close()
	
	eventTypes := make(map[string]int)
	for rows.Next() {
		var eventType string
		var count int
		
		err := rows.Scan(&eventType, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event type: %w", err)
		}
		
		eventTypes[eventType] = count
	}
	stats["event_types"] = eventTypes
	
	// Events by entity type
	entityTypeQuery := `
		SELECT entity_type, COUNT(*) as count
		FROM events
		GROUP BY entity_type
		ORDER BY count DESC
	`
	
	rows, err = r.db.Query(entityTypeQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get entity types: %w", err)
	}
	defer rows.Close()
	
	entityTypes := make(map[string]int)
	for rows.Next() {
		var entityType string
		var count int
		
		err := rows.Scan(&entityType, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan entity type: %w", err)
		}
		
		entityTypes[entityType] = count
	}
	stats["entity_types"] = entityTypes
	
	return stats, nil
}