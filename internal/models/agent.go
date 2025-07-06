package models

import (
	"encoding/json"
	"fmt"
	"time"
)

type Agent struct {
	ID                  int                    `json:"id" db:"id"`
	SessionID           int                    `json:"session_id" db:"session_id"`
	AgentType           string                 `json:"agent_type" db:"agent_type"`
	PID                 int                    `json:"pid" db:"pid"`
	Status              string                 `json:"status" db:"status"`
	Config              map[string]interface{} `json:"config" db:"config"`
	Command             string                 `json:"command" db:"command"`
	WorkingDirectory    string                 `json:"working_directory" db:"working_directory"`
	CommunicationMethod string                 `json:"communication_method" db:"communication_method"`
	InputPipePath       string                 `json:"input_pipe_path" db:"input_pipe_path"`
	OutputPipePath      string                 `json:"output_pipe_path" db:"output_pipe_path"`
	LastHeartbeat       *time.Time             `json:"last_heartbeat" db:"last_heartbeat"`
	ResourceUsage       map[string]interface{} `json:"resource_usage" db:"resource_usage"`
	StartedAt           time.Time              `json:"started_at" db:"started_at"`
	StoppedAt           *time.Time             `json:"stopped_at" db:"stopped_at"`
	
	// Relationships
	Session  *Session       `json:"session,omitempty"`
	Commands []AgentCommand `json:"commands,omitempty"`
	Files    []AgentFile    `json:"files,omitempty"`
}

type AgentStatus string

const (
	AgentStatusStarting AgentStatus = "starting"
	AgentStatusRunning  AgentStatus = "running"
	AgentStatusStopped  AgentStatus = "stopped"
	AgentStatusFailed   AgentStatus = "failed"
)

type CommunicationMethod string

const (
	CommunicationMethodStdio     CommunicationMethod = "stdio"
	CommunicationMethodHTTP      CommunicationMethod = "http"
	CommunicationMethodWebSocket CommunicationMethod = "websocket"
)

type ResourceUsage struct {
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryMB      int     `json:"memory_mb"`
	DiskUsageMB   int     `json:"disk_usage_mb"`
	NetworkBytesTX int64  `json:"network_bytes_tx"`
	NetworkBytesRX int64  `json:"network_bytes_rx"`
}

type CreateAgentRequest struct {
	SessionID        int                    `json:"session_id" binding:"required"`
	AgentType        string                 `json:"agent_type" binding:"required"`
	Command          string                 `json:"command" binding:"required"`
	WorkingDirectory string                 `json:"working_directory"`
	Config           map[string]interface{} `json:"config"`
}

type UpdateAgentRequest struct {
	Status string                 `json:"status"`
	Config map[string]interface{} `json:"config"`
}

func (a *Agent) MarshalConfig() (string, error) {
	if a.Config == nil {
		return "{}", nil
	}
	data, err := json.Marshal(a.Config)
	if err != nil {
		return "", fmt.Errorf("failed to marshal agent config: %w", err)
	}
	return string(data), nil
}

func (a *Agent) UnmarshalConfig(configStr string) error {
	if configStr == "" {
		a.Config = make(map[string]interface{})
		return nil
	}
	
	if err := json.Unmarshal([]byte(configStr), &a.Config); err != nil {
		return fmt.Errorf("failed to unmarshal agent config: %w", err)
	}
	return nil
}

func (a *Agent) MarshalResourceUsage() (string, error) {
	if a.ResourceUsage == nil {
		return "{}", nil
	}
	data, err := json.Marshal(a.ResourceUsage)
	if err != nil {
		return "", fmt.Errorf("failed to marshal agent resource usage: %w", err)
	}
	return string(data), nil
}

func (a *Agent) UnmarshalResourceUsage(usageStr string) error {
	if usageStr == "" {
		a.ResourceUsage = make(map[string]interface{})
		return nil
	}
	
	if err := json.Unmarshal([]byte(usageStr), &a.ResourceUsage); err != nil {
		return fmt.Errorf("failed to unmarshal agent resource usage: %w", err)
	}
	return nil
}

func (a *Agent) Validate() error {
	if a.AgentType == "" {
		return fmt.Errorf("agent type is required")
	}
	
	if a.Command == "" {
		return fmt.Errorf("command is required")
	}
	
	if a.SessionID == 0 {
		return fmt.Errorf("session ID is required")
	}
	
	if a.Status == "" {
		a.Status = string(AgentStatusStarting)
	}
	
	if !a.IsValidStatus() {
		return fmt.Errorf("invalid agent status: %s", a.Status)
	}
	
	if a.CommunicationMethod == "" {
		a.CommunicationMethod = string(CommunicationMethodStdio)
	}
	
	if !a.IsValidCommunicationMethod() {
		return fmt.Errorf("invalid communication method: %s", a.CommunicationMethod)
	}
	
	return nil
}

func (a *Agent) IsValidStatus() bool {
	switch AgentStatus(a.Status) {
	case AgentStatusStarting, AgentStatusRunning, AgentStatusStopped, AgentStatusFailed:
		return true
	default:
		return false
	}
}

func (a *Agent) IsValidCommunicationMethod() bool {
	switch CommunicationMethod(a.CommunicationMethod) {
	case CommunicationMethodStdio, CommunicationMethodHTTP, CommunicationMethodWebSocket:
		return true
	default:
		return false
	}
}

func (a *Agent) BeforeCreate() {
	a.StartedAt = time.Now()
	
	if a.Config == nil {
		a.Config = make(map[string]interface{})
	}
	
	if a.ResourceUsage == nil {
		a.ResourceUsage = make(map[string]interface{})
	}
	
	if a.Status == "" {
		a.Status = string(AgentStatusStarting)
	}
	
	if a.CommunicationMethod == "" {
		a.CommunicationMethod = string(CommunicationMethodStdio)
	}
}

func (a *Agent) IsRunning() bool {
	return a.Status == string(AgentStatusRunning)
}

func (a *Agent) IsStopped() bool {
	return a.Status == string(AgentStatusStopped)
}

func (a *Agent) IsFailed() bool {
	return a.Status == string(AgentStatusFailed)
}

func (a *Agent) Start() {
	a.Status = string(AgentStatusRunning)
	a.StartedAt = time.Now()
}

func (a *Agent) Stop() {
	a.Status = string(AgentStatusStopped)
	now := time.Now()
	a.StoppedAt = &now
}

func (a *Agent) Fail() {
	a.Status = string(AgentStatusFailed)
	now := time.Now()
	a.StoppedAt = &now
}

func (a *Agent) UpdateHeartbeat() {
	now := time.Now()
	a.LastHeartbeat = &now
}

func (a *Agent) IsHealthy(timeout time.Duration) bool {
	if a.LastHeartbeat == nil {
		return false
	}
	return time.Since(*a.LastHeartbeat) < timeout
}