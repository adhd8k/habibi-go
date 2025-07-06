# 🤖 Habibi-Go - Agentic Coding Management Platform

A unified platform for managing AI coding agents across projects and sessions with multi-interface access (web UI, CLI, Slack bot).

## ✅ Phase 1 Implementation Status

**Core Backend Foundation - COMPLETED**

### 🎯 Features Implemented

- ✅ **SQLite Database** with modernc.org/sqlite (pure Go, no CGO)
- ✅ **Project Management** - Full CRUD operations for projects
- ✅ **Configuration System** - YAML config with environment variable support
- ✅ **REST API** - Complete project management endpoints
- ✅ **CLI Interface** - Cobra-based command-line tool
- ✅ **Database Migrations** - Automated schema management
- ✅ **Event System** - Audit trail for all operations

### 🚀 Quick Start

1. **Build the application:**
   ```bash
   go build -o bin/habibi-go main.go
   ```

2. **Create your first project:**
   ```bash
   ./bin/habibi-go project create myapp /path/to/your/project --repo https://github.com/user/repo
   ```

3. **List all projects:**
   ```bash
   ./bin/habibi-go project list
   ```

4. **Start the web server:**
   ```bash
   ./bin/habibi-go server --port 8080
   ```

5. **Access the web interface:**
   Open http://localhost:8080 in your browser

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

### 🌐 API Endpoints

#### Projects
- `GET /api/projects` - List all projects
- `POST /api/projects` - Create new project
- `GET /api/projects/{id}` - Get project details
- `PUT /api/projects/{id}` - Update project
- `DELETE /api/projects/{id}` - Delete project
- `POST /api/projects/discover` - Auto-discover projects

#### System
- `GET /api/health` - Health check
- `GET /api/projects/stats` - Project statistics

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

### 🔄 What's Next

**Phase 2: Session Management & Git Integration** (Coming Next)
- Git worktree integration
- Session lifecycle management
- Branch switching and synchronization

**Phase 3: Agent Orchestration**
- Process launching and monitoring
- Agent communication protocols
- Real-time status updates
- WebSocket integration

**Phase 4: React Web UI**
- Real-time dashboard
- Agent monitoring interface
- Interactive project management
- Live log streaming

**Phase 5: Advanced Features**
- Slack integration
- Cross-platform builds
- Docker deployment
- Documentation

### 🧪 Testing

Test the implementation:

```bash
# Test CLI
./bin/habibi-go --help
./bin/habibi-go project create test-project .
./bin/habibi-go project list

# Test API
./bin/habibi-go server --port 8081 &
curl http://localhost:8081/api/health
curl http://localhost:8081/api/projects
```

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

---

**Phase 1 Complete! ✅**

The core backend foundation is now fully implemented and functional. The application provides a solid base for project management with a complete CLI interface, REST API, and database persistence.

Ready to proceed to Phase 2: Session Management & Git Integration!