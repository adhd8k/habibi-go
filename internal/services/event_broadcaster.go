package services

// EventBroadcaster interface for broadcasting events
type EventBroadcaster interface {
	BroadcastEvent(eventType string, agentID int, data interface{})
}

// NoOpBroadcaster is a no-op implementation for when WebSocket is not available
type NoOpBroadcaster struct{}

func (n *NoOpBroadcaster) BroadcastEvent(eventType string, agentID int, data interface{}) {
	// Do nothing
}