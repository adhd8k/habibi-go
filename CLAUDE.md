# Habibi-Go Claude Integration Guide

This file contains important information for Claude to understand and work effectively with the Habibi-Go project.

## Project Overview

Habibi-Go is a development environment manager that integrates with Claude Code. It manages:
- **Projects**: Git repositories with development environments
- **Sessions**: Isolated worktrees for different features/branches
- **Agents**: Claude Code instances that assist with development tasks

## Key Architecture Components

### Backend (Go)
- **Framework**: Gin web framework
- **Database**: SQLite with modernc.org/sqlite driver
- **Architecture**: Repository pattern with service layers
- **WebSocket**: Real-time communication for agent output

### Frontend (React + TypeScript)
- **Build Tool**: Vite
- **State Management**: Zustand (via useAppStore)
- **UI Components**: Custom components with Tailwind CSS
- **Real-time Updates**: WebSocket client for live streaming

## Important Commands

### Running the Application
```bash
# Backend
go run . server

# Frontend (in web/ directory)
npm run dev

# Build frontend
npm run build

# Run tests
go test ./...
```

### Linting and Type Checking
```bash
# Go linting (if available)
golangci-lint run

# TypeScript type checking
cd web && npm run typecheck

# Frontend linting
cd web && npm run lint
```

## Database Schema

### Key Tables
1. **projects**: Project configurations and paths
2. **sessions**: Development sessions with worktrees
3. **agents**: Claude instances (virtual, not processes)
4. **chat_messages**: Conversation history with tool use support

### Recent Schema Changes
- Added `tool_use` and `tool_result` roles to chat_messages
- Added activity tracking columns to sessions (activity_status, last_activity_at, last_viewed_at)
- Added tool metadata columns (tool_name, tool_input, tool_use_id, tool_content)

## Claude Integration Details

### Claude Agent Service
- Uses `claude` CLI with `--output-format stream-json`
- Supports session resumption with `--resume` flag
- Tracks Claude session IDs for continuity
- Streams responses in real-time via WebSocket

### Message Types
- **user**: User messages
- **assistant**: Claude's text responses
- **tool_use**: Tool invocations by Claude
- **tool_result**: Results from tool executions

### Activity Status
Sessions track activity with these states:
- **idle**: No recent activity
- **streaming**: Currently receiving Claude response
- **new**: New response since last viewed
- **viewed**: Response has been viewed

## Key Features

### Terminal Emulator
- Uses xterm.js for frontend terminal
- PTY (pseudo-terminal) backend for proper terminal control
- Automatic reconnection on server restart
- NixOS compatibility with shell detection

### Git Integration
- Automatic worktree management
- Branch diff visualization against base branch
- Uses `git merge-base` for accurate comparisons
- Supports rebase and push operations

### Real-time Features
- WebSocket for streaming Claude responses
- Session activity indicators
- Audio notifications for completed responses
- Persistent notification preferences

## Development Tips

### Adding New Features
1. Update database schema in `internal/database/connection.go`
2. Update models in `internal/models/`
3. Update repositories in `internal/database/repositories/`
4. Update frontend types in `web/src/types/index.ts`
5. Test database migrations on existing databases

### Common Issues and Solutions
1. **Database constraint errors**: Run migrations or recreate database
2. **Terminal not working**: Check if server restarted, use reconnect button
3. **Messages not showing**: Ensure activity columns are selected in queries
4. **Tool messages missing**: Check that frontend parses tool metadata

### Testing Database Changes
```bash
# Backup existing database
cp ~/.habibi-go/habibi.db ~/.habibi-go/habibi.db.backup

# Test migrations
go run . server

# If issues, restore backup
cp ~/.habibi-go/habibi.db.backup ~/.habibi-go/habibi.db
```

## Project Structure
```
habibi-go/
├── cmd/              # CLI commands
├── internal/         # Core application code
│   ├── api/         # HTTP handlers and routes
│   ├── database/    # Database connection and repositories
│   ├── models/      # Data models
│   └── services/    # Business logic
├── web/             # React frontend
│   ├── src/
│   │   ├── components/  # React components
│   │   ├── api/        # API client
│   │   ├── hooks/      # Custom React hooks
│   │   └── types/      # TypeScript types
│   └── public/         # Static assets
└── CLAUDE.md          # This file
```

## Environment Variables
- `HABIBI_GO_HOME`: Override default data directory (default: ~/.habibi-go)
- `HABIBI_GO_PORT`: Server port (default: 8080)
- `HABIBI_GO_HOST`: Server host (default: localhost)

## Security Considerations
- Never commit sensitive data or credentials
- Use defensive coding for security features
- Validate all user inputs
- Sanitize file paths and commands