# Agentic Coding Management Platform - Go + React Implementation Plan

## Project Overview
Build a unified agentic coding management platform that orchestrates multiple AI coding agents across projects and sessions with multi-interface access (web UI, CLI, Slack bot).

**Core Goal**: Single binary deployment with embedded web UI, supporting git worktree-based session isolation and real-time agent monitoring.

---

## Technology Stack

### Backend (Go 1.21+)
- **Web Framework**: Gin (high performance, minimal)
- **Database**: SQLite with `modernc.org/sqlite` (pure Go, no CGO)
- **WebSocket**: `gorilla/websocket`
- **CLI**: Cobra + Viper
- **Embedding**: Go 1.16+ `embed` package
- **Configuration**: YAML with `gopkg.in/yaml.v3`
- **Logging**: `slog` (Go 1.21 structured logging)

### Frontend (React 18+)
- **Framework**: React 18 + TypeScript
- **Build Tool**: Vite
- **State Management**: Zustand (lightweight)
- **Styling**: Tailwind CSS + Headless UI
- **WebSocket**: Native WebSocket API
- **HTTP Client**: Axios with interceptors

### Key Dependencies
```go
// Go dependencies
github.com/gin-gonic/gin
github.com/gorilla/websocket
github.com/spf13/cobra
github.com/spf13/viper
modernc.org/sqlite
github.com/golang-migrate/migrate/v4
github.com/rs/cors
github.com/golang-jwt/jwt/v5
```

---

## Project Structure

```
agentic-mgr/
├── cmd/
│   ├── root.go                 # Cobra root command
│   ├── server.go              # Server start command
│   ├── project.go             # Project management commands
│   ├── session.go             # Session management commands
│   └── agent.go               # Agent management commands
├── internal/
│   ├── api/
│   │   ├── handlers/
│   │   │   ├── projects.go    # Project CRUD handlers
│   │   │   ├── sessions.go    # Session CRUD handlers
│   │   │   ├── agents.go      # Agent management handlers
│   │   │   ├── agent_comm.go  # Agent communication handlers
│   │   │   ├── agent_files.go # Agent file upload/download handlers
│   │   │   ├── events.go      # Real-time events handler
│   │   │   └── websocket.go   # WebSocket connection handler
│   │   ├── middleware/
│   │   │   ├── cors.go        # CORS middleware
│   │   │   ├── auth.go        # Authentication middleware
│   │   │   └── logging.go     # Request logging middleware
│   │   └── routes.go          # Route definitions
│   ├── models/
│   │   ├── project.go         # Project model and validation
│   │   ├── session.go         # Session model and validation
│   │   ├── agent.go           # Agent model and validation
│   │   └── event.go           # Event model for real-time updates
│   ├── services/
│   │   ├── project_service.go # Project business logic
│   │   ├── session_service.go # Session and git worktree logic
│   │   ├── agent_service.go   # Agent orchestration logic
│   │   ├── agent_comm_service.go # Agent communication handling
│   │   ├── git_service.go     # Git operations wrapper
│   │   └── event_service.go   # Event broadcasting service
│   ├── database/
│   │   ├── migrations/        # SQL migration files
│   │   ├── connection.go      # Database connection and setup
│   │   └── repositories/      # Data access layer
│   │       ├── project_repo.go
│   │       ├── session_repo.go
│   │       ├── agent_repo.go
│   │       ├── agent_command_repo.go  # Agent command history
│   │       ├── agent_file_repo.go     # Agent file management
│   │       └── event_repo.go
│   ├── slack/
│   │   ├── bot.go            # Slack bot implementation
│   │   ├── commands.go       # Slash command handlers
│   │   └── notifications.go  # Slack notification service
│   ├── config/
│   │   └── config.go         # Configuration management
│   └── util/
│       ├── process.go        # Process management utilities
│       ├── git.go            # Git utility functions
│       └── validation.go     # Input validation helpers
├── web/                      # React frontend
│   ├── src/
│   │   ├── components/
│   │   │   ├── Dashboard/
│   │   │   ├── Projects/
│   │   │   ├── Sessions/
│   │   │   ├── Agents/
│   │   │   │   ├── AgentMonitor.tsx    # Real-time agent status
│   │   │   │   ├── AgentChat.tsx       # Agent communication interface
│   │   │   │   ├── AgentControls.tsx   # Agent start/stop controls
│   │   │   │   ├── AgentLogs.tsx       # Live log streaming
│   │   │   │   └── AgentFileManager.tsx # File upload/download
│   │   │   └── common/
│   │   ├── stores/
│   │   │   ├── projectStore.ts
│   │   │   ├── sessionStore.ts
│   │   │   ├── agentStore.ts
│   │   │   └── websocketStore.ts
│   │   ├── services/
│   │   │   ├── api.ts        # API client configuration
│   │   │   ├── websocket.ts  # WebSocket service
│   │   │   ├── agentComm.ts  # Agent communication service
│   │   │   └── types.ts      # TypeScript type definitions
│   │   ├── hooks/
│   │   │   ├── useProjects.ts
│   │   │   ├── useSessions.ts
│   │   │   ├── useAgents.ts
│   │   │   ├── useAgentLogs.ts       # Agent log streaming
│   │   │   ├── useAgentCommands.ts   # Agent command handling
│   │   │   └── useWebSocket.ts
│   │   └── App.tsx
│   ├── package.json
│   ├── vite.config.ts
│   └── tailwind.config.js
├── docs/
│   ├── API.md               # API documentation
│   ├── CLI.md               # CLI usage guide
│   └── DEPLOYMENT.md        # Deployment instructions
├── scripts/
│   ├── build.sh            # Build script for cross-compilation
│   ├── install.sh          # One-line installation script
│   └── dev.sh              # Development environment setup
├── configs/
│   └── config.example.yaml # Example configuration file
├── .goreleaser.yaml        # Release configuration
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── go.mod
├── go.sum
└── README.md
```

---

## Database Schema

### SQLite Schema (migrations/001_initial.up.sql)
```sql
-- Projects table
CREATE TABLE projects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    path TEXT NOT NULL,
    repository_url TEXT,
    default_branch TEXT DEFAULT 'main',
    config TEXT DEFAULT '{}', -- JSON configuration
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Sessions table
CREATE TABLE sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    branch_name TEXT NOT NULL,
    worktree_path TEXT NOT NULL,
    status TEXT DEFAULT 'active' CHECK(status IN ('active', 'paused', 'stopped')),
    config TEXT DEFAULT '{}',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_used_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    UNIQUE(project_id, name)
);

-- Agents table
CREATE TABLE agents (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id INTEGER NOT NULL,
    agent_type TEXT NOT NULL,
    pid INTEGER,
    status TEXT DEFAULT 'starting' CHECK(status IN ('starting', 'running', 'stopped', 'failed')),
    config TEXT DEFAULT '{}',
    command TEXT NOT NULL,
    working_directory TEXT NOT NULL,
    communication_method TEXT DEFAULT 'stdio', -- stdio, http, websocket
    input_pipe_path TEXT,     -- Path to input pipe for command sending
    output_pipe_path TEXT,    -- Path to output pipe for response reading
    last_heartbeat DATETIME, -- For health monitoring
    resource_usage TEXT DEFAULT '{}', -- JSON: {"cpu": 45, "memory": 256}
    started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    stopped_at DATETIME,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

-- Agent Commands table (for command history and debugging)
CREATE TABLE agent_commands (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    agent_id INTEGER NOT NULL,
    command_text TEXT NOT NULL,
    response_text TEXT,
    status TEXT DEFAULT 'pending' CHECK(status IN ('pending', 'completed', 'failed')),
    execution_time_ms INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME,
    FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
);

-- Agent Files table (for file sharing between user and agent)
CREATE TABLE agent_files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    agent_id INTEGER NOT NULL,
    filename TEXT NOT NULL,
    file_path TEXT NOT NULL,
    file_size INTEGER,
    mime_type TEXT,
    direction TEXT CHECK(direction IN ('upload', 'download')), -- user->agent or agent->user
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
);

-- Events table (for real-time updates and audit trail)
CREATE TABLE events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_type TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id INTEGER NOT NULL,
    data TEXT DEFAULT '{}',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX idx_sessions_project_id ON sessions(project_id);
CREATE INDEX idx_agents_session_id ON agents(session_id);
CREATE INDEX idx_agent_commands_agent_id ON agent_commands(agent_id);
CREATE INDEX idx_agent_files_agent_id ON agent_files(agent_id);
CREATE INDEX idx_events_created_at ON events(created_at);
CREATE INDEX idx_events_entity ON events(entity_type, entity_id);
```

---

## API Specification

### RESTful API Endpoints

#### Projects
```
GET    /api/projects                 # List all projects
POST   /api/projects                 # Create new project
GET    /api/projects/{id}            # Get project details
PUT    /api/projects/{id}            # Update project
DELETE /api/projects/{id}            # Delete project
POST   /api/projects/{id}/discover   # Auto-discover git repositories
```

#### Sessions
```
GET    /api/projects/{id}/sessions   # List project sessions
POST   /api/projects/{id}/sessions   # Create new session
GET    /api/sessions/{id}            # Get session details
PUT    /api/sessions/{id}            # Update session
DELETE /api/sessions/{id}            # Delete session
POST   /api/sessions/{id}/activate   # Switch to session
```

#### Agents
```
GET    /api/sessions/{id}/agents     # List session agents
POST   /api/sessions/{id}/agents     # Launch new agent
GET    /api/agents/{id}              # Get agent status
POST   /api/agents/{id}/stop         # Stop agent
GET    /api/agents/{id}/logs         # Get agent logs (with streaming)
POST   /api/agents/{id}/command      # Send command to agent
GET    /api/agents/{id}/logs/stream  # Server-sent events for live logs
POST   /api/agents/{id}/files        # Upload file to agent workspace
GET    /api/agents/{id}/files/{name} # Download file from agent workspace
GET    /api/agents/{id}/health       # Get detailed agent health status
POST   /api/agents/{id}/restart      # Restart agent process
```

#### System
```
GET    /api/health                   # Health check
GET    /api/status                   # System status
POST   /api/config                   # Update configuration
GET    /api/events/stream            # Server-Sent Events for real-time updates
```

#### WebSocket
```
WS     /ws                           # WebSocket endpoint for real-time updates

# WebSocket message types for agent interaction:
{
  "type": "agent_command",           # Send command to agent
  "agent_id": 123,
  "command": "analyze this file"
}

{
  "type": "agent_output",            # Receive agent response
  "agent_id": 123,
  "data": "Analysis complete..."
}

{
  "type": "agent_status_update",     # Real-time status changes
  "agent_id": 123,
  "status": "running",
  "data": {"cpu": 45, "memory": 256}
}

{
  "type": "agent_logs",              # Request log streaming
  "agent_id": 123,
  "lines": 100
}
```

### API Response Format
```json
{
  "success": true,
  "data": {},
  "error": null,
  "timestamp": "2024-01-15T10:30:00Z"
}
```

---

## Implementation Phases

### Phase 1: Core Backend Foundation
**Deliverables:**
1. **Database Setup**
   - SQLite connection with modernc.org/sqlite
   - Migration system implementation
   - Repository pattern for data access
   - Basic CRUD operations for projects

2. **Project Management**
   - Project model and validation
   - Git repository integration
   - Project configuration handling
   - Basic REST API endpoints

3. **Configuration System**
   - YAML configuration loading
   - Environment variable support
   - Default configuration values
   - Configuration validation

**Key Files to Implement:**
```
internal/database/connection.go
internal/database/migrations/001_initial.up.sql
internal/models/project.go
internal/services/project_service.go
internal/database/repositories/project_repo.go
internal/api/handlers/projects.go
internal/config/config.go
cmd/root.go
```

### Phase 2: Session Management & Git Integration
**Deliverables:**
1. **Git Worktree Integration**
   - Git repository operations wrapper
   - Worktree creation and management
   - Branch switching and synchronization
   - Cleanup and error handling

2. **Session Management**
   - Session lifecycle management
   - Worktree path management
   - Session switching logic
   - Session persistence

3. **CLI Interface Foundation**
   - Cobra CLI structure
   - Project management commands
   - Session management commands
   - Configuration commands

**Key Files to Implement:**
```
internal/services/git_service.go
internal/services/session_service.go
internal/models/session.go
internal/database/repositories/session_repo.go
internal/api/handlers/sessions.go
cmd/project.go
cmd/session.go
internal/util/git.go
```

### Phase 3: Agent Orchestration
**Deliverables:**
1. **Agent Management**
   - Process launching and monitoring
   - Agent lifecycle management
   - Resource usage tracking
   - Error handling and recovery

2. **Agent Communication**
   - Standard input/output handling
   - Health check implementation
   - Log collection and streaming
   - Graceful shutdown procedures

3. **Event System**
   - Event broadcasting system
   - Real-time updates infrastructure
   - Event persistence for audit trail
   - WebSocket connection management

**Key Files to Implement:**
```
internal/services/agent_service.go
internal/services/agent_comm_service.go  # NEW: Agent communication
internal/models/agent.go
internal/database/repositories/agent_repo.go
internal/database/repositories/agent_command_repo.go  # NEW: Command history
internal/database/repositories/agent_file_repo.go     # NEW: File management
internal/api/handlers/agents.go
internal/api/handlers/agent_comm.go      # NEW: Communication endpoints
internal/api/handlers/agent_files.go     # NEW: File upload/download
internal/services/event_service.go
internal/api/handlers/websocket.go
internal/util/process.go
cmd/agent.go
```

### Phase 4: Web UI Development
**Deliverables:**
1. **React Application Setup**
   - Vite build configuration
   - TypeScript setup
   - Tailwind CSS integration
   - API client configuration

2. **Core Components**
   - Dashboard layout
   - Project management interface
   - Session management interface
   - Agent monitoring interface

3. **Real-time Features**
   - WebSocket integration
   - Live status updates
   - Real-time log streaming
   - Event notifications

**Key Files to Implement:**
```
web/src/App.tsx
web/src/services/api.ts
web/src/services/websocket.ts
web/src/services/agentComm.ts           # NEW: Agent communication service
web/src/stores/projectStore.ts
web/src/stores/sessionStore.ts
web/src/stores/agentStore.ts
web/src/components/Dashboard/
web/src/components/Projects/
web/src/components/Sessions/
web/src/components/Agents/
web/src/components/Agents/AgentMonitor.tsx      # NEW: Real-time monitoring
web/src/components/Agents/AgentChat.tsx         # NEW: Chat interface  
web/src/components/Agents/AgentControls.tsx     # NEW: Control panel
web/src/components/Agents/AgentLogs.tsx         # NEW: Log streaming
web/src/components/Agents/AgentFileManager.tsx  # NEW: File management
web/src/hooks/useAgentLogs.ts           # NEW: Log streaming hook
web/src/hooks/useAgentCommands.ts       # NEW: Command handling hook
```

### Phase 5: Advanced Features & Polish
**Deliverables:**
1. **Slack Integration**
   - Slack bot implementation
   - Slash command handlers
   - Interactive message components
   - Notification system

2. **Build & Deployment**
   - Cross-platform build scripts
   - Asset embedding
   - Single binary compilation
   - Installation scripts

3. **Documentation & Testing**
   - API documentation
   - CLI usage guide
   - Unit and integration tests
   - End-to-end testing

**Key Files to Implement:**
```
internal/slack/bot.go
internal/slack/commands.go
scripts/build.sh
scripts/install.sh
Makefile
.goreleaser.yaml
```

---

## Configuration Schema

### Default Configuration (configs/config.example.yaml)
```yaml
# Server configuration
server:
  host: "localhost"
  port: 8080
  read_timeout: "30s"
  write_timeout: "30s"
  shutdown_timeout: "10s"

# Database configuration
database:
  path: "~/.agentic-mgr/data.db"
  backup_enabled: true
  backup_interval: "24h"
  max_connections: 10

# Projects configuration
projects:
  default_directory: "~/projects"
  auto_discover: true
  worktree_base_path: ".worktrees"

# Agent configuration
agents:
  default_timeout: "30m"
  max_concurrent: 10
  health_check_interval: "30s"
  log_retention_days: 7
  resource_limits:
    memory_mb: 1024
    cpu_percent: 50

# Slack integration (optional)
slack:
  enabled: false
  bot_token: ""
  app_token: ""
  signing_secret: ""
  notification_channel: "#dev"

# Logging configuration
logging:
  level: "info"
  format: "json"
  file_path: "~/.agentic-mgr/logs/app.log"
  max_size_mb: 100
  max_backups: 5
```

---

## Build & Deployment

### Makefile
```makefile
.PHONY: build clean dev test install deps build-ui embed cross-compile

# Development
dev:
	@echo "Starting development environment..."
	cd web && npm run dev &
	go run main.go server --dev

# Dependencies
deps:
	go mod download
	cd web && npm install

# Build frontend
build-ui:
	cd web && npm run build

# Embed assets and build
build: build-ui
	go build -ldflags="-s -w" -o bin/agentic-mgr main.go

# Cross-compile for all platforms
cross-compile: build-ui
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/agentic-mgr-linux-amd64 main.go
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o bin/agentic-mgr-darwin-amd64 main.go
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o bin/agentic-mgr-darwin-arm64 main.go
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o bin/agentic-mgr-windows-amd64.exe main.go

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf web/dist/

# Run tests
test:
	go test ./...
	cd web && npm test

# Install locally
install: build
	cp bin/agentic-mgr /usr/local/bin/
```

### Installation Script (scripts/install.sh)
```bash
#!/bin/bash
set -e

# Agentic Coding Manager - One-line installer
REPO="your-org/agentic-mgr"
BINARY_NAME="agentic-mgr"
INSTALL_DIR="/usr/local/bin"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH" && exit 1 ;;
esac

# Download URL
DOWNLOAD_URL="https://github.com/${REPO}/releases/latest/download/${BINARY_NAME}-${OS}-${ARCH}"
if [ "$OS" = "windows" ]; then
    DOWNLOAD_URL="${DOWNLOAD_URL}.exe"
    BINARY_NAME="${BINARY_NAME}.exe"
fi

echo "Downloading ${BINARY_NAME} for ${OS}-${ARCH}..."

# Download and install
curl -fsSL "$DOWNLOAD_URL" -o "/tmp/$BINARY_NAME"
chmod +x "/tmp/$BINARY_NAME"

# Install (may require sudo)
if [ -w "$INSTALL_DIR" ]; then
    mv "/tmp/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
else
    sudo mv "/tmp/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
fi

echo "✅ ${BINARY_NAME} installed successfully!"
echo "🚀 Run 'agentic-mgr --help' to get started"
```

---

## CLI Command Structure

### Root Command (cmd/root.go)
```go
var rootCmd = &cobra.Command{
    Use:   "agentic-mgr",
    Short: "Agentic Coding Management Platform",
    Long:  `A unified platform for managing AI coding agents across projects and sessions.`,
}

func init() {
    rootCmd.AddCommand(serverCmd)
    rootCmd.AddCommand(projectCmd)
    rootCmd.AddCommand(sessionCmd)
    rootCmd.AddCommand(agentCmd)
    rootCmd.AddCommand(configCmd)
}
```

### Server Command (cmd/server.go)
```bash
# Start the server
agentic-mgr server --port 8080 --host localhost

# Start in development mode
agentic-mgr server --dev

# Start with custom config
agentic-mgr server --config ./config.yaml
```

### Project Commands (cmd/project.go)
```bash
# List all projects
agentic-mgr project list

# Create new project
agentic-mgr project create myapp --repo https://github.com/user/repo

# Show project details
agentic-mgr project show myapp

# Update project
agentic-mgr project update myapp --default-branch develop

# Delete project
agentic-mgr project delete myapp

# Auto-discover projects
agentic-mgr project discover ~/workspace
```

### Session Commands (cmd/session.go)
```bash
# List sessions in project
agentic-mgr session list myapp

# Create new session
agentic-mgr session create myapp feature-x --branch feature/auth

# Show session details
agentic-mgr session show myapp/feature-x

# Switch to session
agentic-mgr session switch myapp/feature-x

# Delete session
agentic-mgr session delete myapp/feature-x

# Clean up stopped sessions
agentic-mgr session cleanup myapp
```

### Agent Commands (cmd/agent.go)
```bash
# List agents in session
agentic-mgr agent list myapp/feature-x

# Start new agent
agentic-mgr agent start myapp/feature-x --type claude-code --config agent.yaml

# Show agent status
agentic-mgr agent status <agent-id>

# Stop agent
agentic-mgr agent stop <agent-id>

# View agent logs
agentic-mgr agent logs <agent-id> --follow

# Send command to agent (NEW)
agentic-mgr agent exec <agent-id> "analyze this codebase"

# Interactive chat with agent (NEW)
agentic-mgr agent shell <agent-id>

# Send file to agent (NEW)
agentic-mgr agent send-file <agent-id> ./config.json

# Get file from agent (NEW)
agentic-mgr agent get-file <agent-id> ./output.txt

# Agent health check (NEW)
agentic-mgr agent health <agent-id>

# Restart agent (NEW)
agentic-mgr agent restart <agent-id>

# Kill all agents in session
agentic-mgr agent kill-all myapp/feature-x
```

## **Agent Interaction Architecture**

### User Interaction Points

#### **1. Web UI Dashboard** (Primary Interface)
- **Real-time Agent Monitoring**: Live status cards showing agent health, resource usage, and current activity
- **Interactive Chat Interface**: Send commands/questions to agents and receive structured responses
- **File Management**: Drag-and-drop file upload to agent workspace, download agent-generated files
- **Live Log Streaming**: Terminal-style component with real-time log updates and filtering
- **Agent Control Panel**: Start/stop/restart controls, configuration editor, emergency kill switch

#### **2. CLI Interface** (Automation & Scripting)
- **Direct Command Execution**: `agentic-mgr agent exec <id> "command"`
- **Interactive Shell**: `agentic-mgr agent shell <id>` for persistent communication
- **File Transfer**: Upload/download files to/from agent workspace
- **Status Monitoring**: Detailed agent health and performance metrics
- **Log Streaming**: Real-time log following with filtering options

#### **3. Slack Integration** (Team Coordination)
- **Status Commands**: `/agentic agent status` for quick team updates
- **Command Delegation**: `/agentic agent exec` for remote agent control
- **Notifications**: Automatic alerts for agent status changes and task completion
- **Log Sharing**: Agent output snippets shared in team channels

### Agent Communication Protocol

#### **Bidirectional Communication Channels**
- **Standard I/O Pipes**: Direct stdin/stdout communication with agent processes
- **WebSocket Streams**: Real-time bidirectional communication for web UI
- **File System Interface**: Shared workspace for file exchange
- **HTTP API**: RESTful endpoints for command execution and status queries

#### **Message Types & Flow**
1. **Command Execution**: User sends command → Agent processes → Response streamed back
2. **File Operations**: User uploads file → Available in agent workspace → Agent can process and generate outputs
3. **Status Updates**: Agent health/resource usage → Broadcast to all connected interfaces
4. **Event Notifications**: Agent lifecycle events → Real-time updates to dashboards and notifications

### Implementation Details

#### **Agent Communication Service** (`internal/services/agent_comm_service.go`)
```go
type AgentCommunication struct {
    AgentID         int
    InputChannel    chan string          // Commands from user to agent
    OutputChannel   chan string          // Responses from agent to user  
    StatusChannel   chan AgentStatus     // Status updates
    FileChannel     chan FileOperation   // File upload/download operations
    HealthTicker    *time.Ticker         // Periodic health checks
}

// Send command and wait for response
func (ac *AgentCommunication) ExecuteCommand(command string, timeout time.Duration) (*CommandResult, error)

// Stream live output from agent
func (ac *AgentCommunication) StreamOutput() <-chan string

// Upload file to agent workspace
func (ac *AgentCommunication) UploadFile(filename string, content []byte) error

// Download file from agent workspace  
func (ac *AgentCommunication) DownloadFile(filename string) ([]byte, error)
```

#### **Web UI Agent Components**
- **AgentChat.tsx**: Chat interface with command history, autocomplete, and rich message formatting
- **AgentMonitor.tsx**: Real-time dashboards showing CPU/memory usage, task progress, and health status
- **AgentFileManager.tsx**: File browser for agent workspace with upload/download capabilities
- **AgentLogs.tsx**: Live log viewer with search, filtering, and export functionality

#### **Database Schema for Agent Interaction**
- **agent_commands**: Command history with execution time and responses
- **agent_files**: File transfer tracking with metadata and access logs
- **agent_events**: Detailed audit trail of all agent interactions and status changes

This comprehensive agent interaction architecture ensures users can seamlessly communicate with agents across all interfaces while maintaining full visibility into agent status, performance, and outputs.

### 1. Git Worktree Management
```go
// internal/services/git_service.go
func (s *GitService) CreateWorktree(projectPath, sessionName, branchName string) error {
    worktreePath := filepath.Join(projectPath, ".worktrees", sessionName)
    
    // Create worktree directory
    if err := os.MkdirAll(filepath.Dir(worktreePath), 0755); err != nil {
        return err
    }
    
    // Add git worktree
    cmd := exec.Command("git", "worktree", "add", worktreePath, branchName)
    cmd.Dir = projectPath
    return cmd.Run()
}
```

### 2. Agent Process Management
```go
// internal/services/agent_service.go
type AgentProcess struct {
    ID      int
    Cmd     *exec.Cmd
    Status  string
    LogFile *os.File
    InputPipe  io.WriteCloser  // NEW: For sending commands
    OutputPipe io.ReadCloser   // NEW: For receiving responses
    ErrorPipe  io.ReadCloser   // NEW: For error handling
}

func (s *AgentService) StartAgent(sessionID int, config AgentConfig) (*AgentProcess, error) {
    // Create log file
    logPath := filepath.Join(s.logDir, fmt.Sprintf("agent-%d.log", sessionID))
    logFile, err := os.Create(logPath)
    if err != nil {
        return nil, err
    }
    
    // Create pipes for communication
    inputPipe, err := cmd.StdinPipe()
    if err != nil {
        return nil, err
    }
    
    outputPipe, err := cmd.StdoutPipe()
    if err != nil {
        return nil, err
    }
    
    errorPipe, err := cmd.StderrPipe()
    if err != nil {
        return nil, err
    }
    
    // Start process
    cmd := exec.Command(config.Command, config.Args...)
    cmd.Dir = config.WorkingDir
    
    if err := cmd.Start(); err != nil {
        return nil, err
    }
    
    return &AgentProcess{
        Cmd:        cmd,
        Status:     "running",
        LogFile:    logFile,
        InputPipe:  inputPipe,
        OutputPipe: outputPipe,
        ErrorPipe:  errorPipe,
    }, nil
}

// NEW: Send command to agent
func (s *AgentService) SendCommand(agentID int, command string) error {
    agent := s.GetAgent(agentID)
    if agent == nil {
        return fmt.Errorf("agent %d not found", agentID)
    }
    
    // Send command to agent's stdin
    _, err := agent.InputPipe.Write([]byte(command + "\n"))
    if err != nil {
        return fmt.Errorf("failed to send command: %w", err)
    }
    
    // Store command in database for history
    s.storeCommand(agentID, command)
    
    return nil
}

// NEW: Stream agent output
func (s *AgentService) StreamOutput(agentID int) (<-chan string, error) {
    agent := s.GetAgent(agentID)
    if agent == nil {
        return nil, fmt.Errorf("agent %d not found", agentID)
    }
    
    outputChan := make(chan string, 100)
    
    go func() {
        defer close(outputChan)
        scanner := bufio.NewScanner(agent.OutputPipe)
        for scanner.Scan() {
            line := scanner.Text()
            outputChan <- line
            
            // Also broadcast via WebSocket
            s.eventService.BroadcastEvent("agent_output", map[string]interface{}{
                "agent_id": agentID,
                "output":   line,
                "timestamp": time.Now(),
            })
        }
    }()
    
    return outputChan, nil
}
```

### 3. WebSocket Event Broadcasting
```go
// internal/api/handlers/websocket.go
type Hub struct {
    clients    map[*Client]bool
    broadcast  chan []byte
    register   chan *Client
    unregister chan *Client
}

type WSMessage struct {
    Type    string      `json:"type"`
    AgentID int         `json:"agent_id,omitempty"`
    Command string      `json:"command,omitempty"`
    Data    interface{} `json:"data,omitempty"`
}

func (h *Hub) BroadcastEvent(eventType string, data interface{}) {
    event := Event{
        Type:      eventType,
        Data:      data,
        Timestamp: time.Now(),
    }
    
    message, _ := json.Marshal(event)
    h.broadcast <- message
}

// NEW: Handle agent-specific WebSocket messages
func (h *WebSocketHandler) handleAgentMessage(conn *websocket.Conn, msg WSMessage) {
    switch msg.Type {
    case "agent_command":
        // Send command to agent
        err := h.agentService.SendCommand(msg.AgentID, msg.Command)
        if err != nil {
            h.sendError(conn, fmt.Sprintf("Failed to send command: %v", err))
            return
        }
        
        // Acknowledge command sent
        h.sendMessage(conn, WSMessage{
            Type: "command_sent",
            AgentID: msg.AgentID,
            Data: map[string]string{"status": "sent"},
        })
        
    case "agent_logs_subscribe":
        // Start streaming logs for this agent
        logStream, err := h.agentService.StreamOutput(msg.AgentID)
        if err != nil {
            h.sendError(conn, fmt.Sprintf("Failed to stream logs: %v", err))
            return
        }
        
        go func() {
            for log := range logStream {
                h.sendMessage(conn, WSMessage{
                    Type: "agent_log",
                    AgentID: msg.AgentID,
                    Data: map[string]interface{}{
                        "message": log,
                        "timestamp": time.Now(),
                    },
                })
            }
        }()
        
    case "agent_file_upload":
        // Handle file upload to agent workspace
        err := h.agentService.HandleFileUpload(msg.AgentID, msg.Data)
        if err != nil {
            h.sendError(conn, fmt.Sprintf("File upload failed: %v", err))
            return
        }
        
        h.sendMessage(conn, WSMessage{
            Type: "file_uploaded",
            AgentID: msg.AgentID,
            Data: map[string]string{"status": "success"},
        })
    }
}
```

### 4. Embedded Assets
```go
// main.go
//go:embed web/dist/*
var webAssets embed.FS

func setupStaticFiles(r *gin.Engine) {
    // Serve embedded React build
    r.StaticFS("/static", http.FS(webAssets))
    
    // Catch-all for React router
    r.NoRoute(func(c *gin.Context) {
        data, _ := webAssets.ReadFile("web/dist/index.html")
        c.Data(http.StatusOK, "text/html; charset=utf-8", data)
    })
}
```

This comprehensive implementation plan provides claude-code with all the necessary details, file structure, code examples, and step-by-step phases to build the complete agentic coding management platform using Go + React.