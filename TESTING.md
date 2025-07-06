# Testing Habibi-Go Locally

## Quick Start

### 1. Install Dependencies
```bash
make deps
```

### 2. Build and Run
```bash
# Build everything (backend + frontend)
make build

# Run the server
make run
```

The application will be available at http://localhost:8080

## Development Mode

For development with hot-reload:

```bash
# Terminal 1: Run both frontend and backend in dev mode
make dev
```

This starts:
- Vite dev server on http://localhost:3000 (with hot module replacement)
- Go server on http://localhost:8080 (with --dev flag)

## Testing Each Phase

### Phase 1: Core Backend
```bash
# Create a project
./bin/habibi-go project create "My Project" "/path/to/project"

# List projects
./bin/habibi-go project list

# View project details
./bin/habibi-go project get 1
```

### Phase 2: Session Management
```bash
# Create a session
./bin/habibi-go session create "My Project" "feature-session" "feature-branch"

# List sessions
./bin/habibi-go session list 1
```

### Phase 3: Agent Orchestration
```bash
# Start an agent
./bin/habibi-go agent start 1 "test-agent" "echo 'Hello World'"

# List agents
./bin/habibi-go agent list

# Execute command on agent
./bin/habibi-go agent exec 1 "pwd"

# Stop agent
./bin/habibi-go agent stop 1
```

### Phase 4: Web UI
1. Build and run: `make run`
2. Open http://localhost:8080
3. Use the web interface to:
   - Create and manage projects
   - Create sessions within projects
   - Start and control agents
   - View real-time agent output

## API Testing

Test API endpoints directly:

```bash
# Health check
curl http://localhost:8080/api/health

# List projects
curl http://localhost:8080/api/v1/projects

# Create project
curl -X POST http://localhost:8080/api/v1/projects \
  -H "Content-Type: application/json" \
  -d '{"name": "Test Project", "path": "/tmp/test"}'

# List sessions
curl http://localhost:8080/api/v1/sessions

# Create agent
curl -X POST http://localhost:8080/api/v1/agents \
  -H "Content-Type: application/json" \
  -d '{"session_id": 1, "agent_type": "test", "command": "echo test"}'
```

## WebSocket Testing

Connect to WebSocket endpoint for real-time updates:
- Endpoint: `ws://localhost:8080/ws`
- Subscribe to agent: `{"type": "subscribe", "agent_id": 1}`

## Troubleshooting

### Reset Database
```bash
make db-reset
```

### Clean Build
```bash
make clean
make build
```

### Check Logs
The server logs all requests and errors to stdout.

### Common Issues

1. **Port already in use**: Kill existing processes or change port
   ```bash
   # Change port
   ./bin/habibi-go server --port 8081
   ```

2. **Database locked**: Reset the database
   ```bash
   make db-reset
   ```

3. **Frontend not building**: Check Node.js version (requires 18+)
   ```bash
   node --version
   npm --version
   ```

## Full Demo

Run a complete demo:
```bash
make test-all
```

This will:
1. Create a test project
2. Create a session
3. Start an agent
4. Launch the server

Then visit http://localhost:8080 to see everything in action!