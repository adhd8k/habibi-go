# 🤖 Habibi-Go - Agentic Coding Management Platform

A unified platform for managing AI coding agents across projects and sessions with multi-interface access (web UI, CLI, Slack bot).

## ✅ Implementation Status

- **Phase 1: Core Backend Foundation** - ✅ COMPLETED
- **Phase 2: Session Management & Git Integration** - ✅ COMPLETED  
- **Phase 3: Agent Orchestration** - ✅ COMPLETED
- **Phase 4: React Web UI** - ✅ COMPLETED
- **Phase 5: Build System & Deployment** - 🔄 In Progress

### 🎯 Features Implemented

- ✅ **SQLite Database** with modernc.org/sqlite (pure Go, no CGO)
- ✅ **Project Management** - Full CRUD operations for projects
- ✅ **Session Management** - Complete session lifecycle with Git worktrees
- ✅ **Git Integration** - Worktree creation, branch switching, status monitoring
- ✅ **Agent Orchestration** - Start, stop, and control AI coding agents
- ✅ **Process Management** - Monitor and manage agent processes
- ✅ **Real-time Communication** - WebSocket support for live agent output
- ✅ **React Web UI** - Modern interface with TypeScript and Tailwind CSS
- ✅ **Configuration System** - YAML config with environment variable support
- ✅ **REST API** - Complete API for projects, sessions, and agents
- ✅ **CLI Interface** - Cobra-based command-line tool
- ✅ **Database Migrations** - Automated schema management
- ✅ **Event System** - Audit trail for all operations

### 🚀 Quick Start

1. **Install dependencies:**
   ```bash
   make deps
   ```

2. **Build and run:**
   ```bash
   make build
   make run
   ```

3. **Access the web UI:**
   Open http://localhost:8080 in your browser

### 💻 Development

For development with hot-reload:
```bash
make dev
```

This starts:
- Vite dev server on http://localhost:3000
- Go server on http://localhost:8080

### 📋 Testing

For detailed testing instructions, see [TESTING.md](TESTING.md).

Quick test all features:
```bash
make test-all
```

### 🔧 CLI Commands

#### Project Management
```bash
# List all projects
./habibi-go project list

# Create new project
./habibi-go project create <name> <path> [--repo <url>] [--branch <branch>]

# Show project details
./habibi-go project show <name>

# Delete project
./habibi-go project delete <name>

# Auto-discover projects
./habibi-go project discover ~/workspace [--auto-create]
```

#### Server Management
```bash
# Start server
./habibi-go server [--port 8080] [--host localhost] [--dev]

# Show configuration
./habibi-go config show
```

#### Session Management
```bash
# List all sessions
./habibi-go session list

# List sessions for a project
./habibi-go session list myapp

# Create new session
./habibi-go session create <project-name> <session-name> <branch-name>

# Show session details
./habibi-go session show <session-id>

# Activate session
./habibi-go session activate <session-id>

# Delete session
./habibi-go session delete <session-id>

# Sync session with remote
./habibi-go session sync <session-id>

# Clean up stopped sessions
./habibi-go session cleanup [project-name]
```

### 🌐 API Endpoints

#### Projects
- `GET /api/projects` - List all projects
- `POST /api/projects` - Create new project
- `GET /api/projects/{id}` - Get project details
- `PUT /api/projects/{id}` - Update project
- `DELETE /api/projects/{id}` - Delete project
- `POST /api/projects/discover` - Auto-discover projects

#### Sessions
- `GET /api/sessions` - List all sessions
- `POST /api/sessions` - Create new session
- `GET /api/sessions/{id}` - Get session details
- `PUT /api/sessions/{id}` - Update session
- `DELETE /api/sessions/{id}` - Delete session
- `POST /api/sessions/{id}/activate` - Activate session
- `GET /api/sessions/{id}/status` - Get session status
- `POST /api/sessions/{id}/sync` - Sync session with remote
- `POST /api/sessions/cleanup` - Clean up stopped sessions
- `GET /api/projects/{id}/sessions` - Get sessions for project

#### System
- `GET /api/health` - Health check
- `GET /api/projects/stats` - Project statistics
- `GET /api/sessions/stats` - Session statistics

### 📁 Project Structure

```
habibi-go/
├── cmd/                      # CLI commands (Cobra)
│   ├── root.go              # Root command setup
│   ├── server.go            # Web server command
│   ├── project.go           # Project management commands
│   ├── session.go           # Session commands (Phase 2)
│   └── agent.go             # Agent commands (Phase 3)
├── internal/
│   ├── api/                 # REST API
│   │   ├── handlers/        # HTTP handlers
│   │   ├── middleware/      # CORS, logging, auth
│   │   └── routes.go        # Route definitions
│   ├── models/              # Data models
│   │   ├── project.go       # Project model
│   │   ├── session.go       # Session model (Phase 2)
│   │   ├── agent.go         # Agent model (Phase 3)
│   │   └── event.go         # Event model
│   ├── services/            # Business logic
│   │   └── project_service.go
│   ├── database/            # Database layer
│   │   ├── connection.go    # DB connection & migrations
│   │   └── repositories/    # Data access layer
│   └── config/              # Configuration management
├── web/                     # React frontend (Phase 4)
│   └── dist/               # Built frontend assets
├── configs/                 # Configuration files
└── bin/                    # Built binaries
```

### ⚙️ Configuration

Configuration is managed via YAML files. Default locations:
- `./config.yaml`
- `./configs/config.yaml`
- `~/.habibi-go/config.yaml`

Example configuration:
```yaml
server:
  host: "localhost"
  port: 8080

database:
  path: "~/.habibi-go/data.db"

projects:
  default_directory: "~/projects"
  auto_discover: true

logging:
  level: "info"
  format: "json"
```

### 🔧 Agent Management

```bash
# List agents
./habibi-go agent list [session-id]

# Start agent
./habibi-go agent start <session-id> <agent-type> <command>

# Stop agent
./habibi-go agent stop <agent-id>

# Execute command
./habibi-go agent exec <agent-id> <command>

# View agent status
./habibi-go agent status <agent-id>

# View agent logs
./habibi-go agent logs <agent-id> [--follow]
```

### 🌐 API Endpoints

#### API v1 (Frontend Compatible)
- Projects: `/api/v1/projects`
- Sessions: `/api/v1/sessions`
- Agents: `/api/v1/agents`
- WebSocket: `/ws`

See [TESTING.md](TESTING.md) for detailed API examples.

### 📊 Database Schema

The application uses SQLite with the following core tables:
- `projects` - Project definitions and metadata
- `sessions` - Development sessions with git worktrees
- `agents` - Running agent instances
- `events` - Audit trail of all operations

### 🛠️ Development

```bash
# Install dependencies
go mod download

# Build application
go build -o bin/habibi-go main.go

# Run in development mode
go run main.go server --dev

# Cross-compile for other platforms
GOOS=linux GOARCH=amd64 go build -o bin/habibi-go-linux main.go
```

### 🛠️ Development

```bash
# Install dependencies
make deps

# Development mode with hot reload
make dev

# Build for production
make build

# Cross-compile for all platforms
make cross-compile

# Clean build artifacts
make clean

# Reset database
make db-reset
```

### 📦 Deployment

The application builds into a single binary with embedded web assets:

```bash
# Build production binary
make build

# Run in production
./bin/habibi-go server --port 8080
```

### 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

### 📄 License

MIT License - see LICENSE file for details.