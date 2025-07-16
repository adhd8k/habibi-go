# 🚀 Habibi-Go

<p align="center">
  <strong>AI-Powered Development Environment Manager</strong>
</p>

<p align="center">
  A modern development environment manager that seamlessly integrates with Claude Code to provide intelligent, isolated workspaces for your projects. Manage multiple AI coding sessions across different branches with a beautiful web interface.
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/React-18+-61DAFB?style=flat&logo=react" alt="React Version">
  <img src="https://img.shields.io/badge/TypeScript-5+-3178C6?style=flat&logo=typescript" alt="TypeScript Version">
  <img src="https://img.shields.io/badge/License-MIT-green.svg" alt="License">
</p>

## ✨ Features

### 🎯 Core Features
- **🤖 Claude Code Integration** - Seamlessly manage Claude Code instances across multiple projects
- **📁 Project Management** - Organize and track all your development projects in one place
- **🌳 Git Worktree Sessions** - Create isolated development environments for different features/branches
- **💬 Real-time Chat Interface** - Interactive chat with Claude including tool use visualization
- **🖥️ Integrated Terminal** - Full terminal access with xterm.js for each session
- **🔄 Session Persistence** - Resume conversations and maintain context across sessions
- **📊 Activity Tracking** - Monitor session activity with visual indicators and notifications

### 🛠️ Technical Features
- **Pure Go Backend** - No CGO dependencies, built with Gin framework
- **React + TypeScript Frontend** - Modern, responsive UI with Tailwind CSS
- **SQLite Database** - Lightweight, embedded database with automatic migrations
- **WebSocket Support** - Real-time streaming of Claude responses
- **RESTful API** - Complete API for all operations
- **CLI Interface** - Powerful command-line tools built with Cobra

## 📋 Prerequisites

Before you begin, ensure you have the following installed:

1. **Claude Code** (required)
   ```bash
   # Install Claude Code CLI
   npm install -g @anthropic/claude-cli
   
   # Configure Claude Code with your API key
   claude login
   ```

2. **For Development** (optional)
   - **Go** 1.23 or higher - [Install Go](https://golang.org/doc/install)
   - **Node.js** 18+ and npm - [Install Node.js](https://nodejs.org/)

## 🚀 Quick Start

### Option 1: Download Pre-built Binary (Recommended)

1. Download the latest release for your platform from the [Releases](https://github.com/yourusername/habibi-go/releases) page

2. Make it executable (Linux/macOS):
   ```bash
   chmod +x habibi-go-<platform>
   ```

3. Run the application:
   ```bash
   ./habibi-go-<platform> server
   ```

4. Open your browser at http://localhost:8080

### Option 2: Build from Source

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/habibi-go.git
   cd habibi-go
   ```

2. Install dependencies:
   ```bash
   make deps
   ```

3. Build and run:
   ```bash
   make build
   ./bin/habibi-go server
   ```

4. Open your browser at http://localhost:8080

## 🏗️ Building

### Build for Current Platform
```bash
make build
```

### Build for All Platforms
```bash
make cross-compile
```

This creates binaries for:
- Linux (amd64): `bin/habibi-go-linux-amd64`
- macOS (Intel): `bin/habibi-go-darwin-amd64`
- macOS (Apple Silicon): `bin/habibi-go-darwin-arm64`
- Windows: `bin/habibi-go-windows-amd64.exe`

### Custom Build
```bash
# Linux
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o habibi-go-linux main.go

# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o habibi-go-macos-intel main.go

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o habibi-go-macos-arm64 main.go

# Windows
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o habibi-go-windows.exe main.go
```

## 💻 Usage

### Web Interface

1. Start the server:
   ```bash
   habibi-go server
   ```

2. Open http://localhost:8080 in your browser

3. Create a new project or import existing ones

4. Start a new session to begin coding with Claude

### CLI Commands

#### Project Management
```bash
# List all projects
habibi-go project list

# Create a new project
habibi-go project create <name> <path>

# Import existing git repository
habibi-go project create <name> <path> --repo <git-url>

# Auto-discover projects
habibi-go project discover ~/workspace --auto-create
```

#### Session Management
```bash
# Create a new session
habibi-go session create <project-name> <session-name> <branch-name>

# List sessions
habibi-go session list

# Activate a session
habibi-go session activate <session-id>
```

## 🔧 Configuration

Configuration can be set via:
1. Configuration file: `~/.habibi-go/config.yaml`
2. Environment variables
3. Command-line flags

### Example Configuration
```yaml
server:
  host: "localhost"
  port: 8080

database:
  path: "~/.habibi-go/habibi.db"

projects:
  default_directory: "~/projects"
  auto_discover: true

logging:
  level: "info"
  format: "json"
```

### Environment Variables
- `HABIBI_GO_HOME`: Override default data directory (default: `~/.habibi-go`)
- `HABIBI_GO_PORT`: Server port (default: 8080)
- `HABIBI_GO_HOST`: Server host (default: localhost)

## 🤝 Contributing

This project was developed entirely with AI assistance, and we encourage using AI tools for contributions!

### Development Setup

1. Fork and clone the repository
2. Install dependencies:
   ```bash
   make deps
   ```

3. Start development mode with hot-reload:
   ```bash
   make dev
   ```

### Using AI for Contributions

We recommend using Claude Code or Cursor for development:

1. **Claude Code Integration**
   - The project includes a `CLAUDE.md` file with context for Claude
   - Use Claude Code for implementing features and fixing bugs

2. **Cursor Rules** (coming soon)
   - `.cursorrules` file will be added for Cursor IDE integration

3. **AI Development Guidelines**
   - Let AI understand the codebase through the CLAUDE.md file
   - Use AI to maintain consistent code style
   - Leverage AI for writing tests and documentation

### Submitting Changes

1. Create a feature branch
2. Make your changes (preferably with AI assistance)
3. Run tests: `make test`
4. Submit a Pull Request

## 📁 Project Structure

```
habibi-go/
├── cmd/                    # CLI commands
├── internal/               # Core application code
│   ├── api/               # HTTP handlers and routes
│   ├── database/          # Database layer
│   ├── models/            # Data models
│   └── services/          # Business logic
├── web/                   # React frontend
│   ├── src/
│   │   ├── components/    # React components
│   │   ├── api/          # API client
│   │   ├── hooks/        # Custom hooks
│   │   └── types/        # TypeScript types
│   └── dist/             # Built frontend
├── CLAUDE.md             # AI context file
├── Makefile              # Build automation
└── README.md             # This file
```

## 🧪 Testing

Run all tests:
```bash
make test
```

Test specific features:
```bash
# Backend tests
go test ./...

# Frontend tests
cd web && npm test
```

## 🚨 Troubleshooting

### Common Issues

1. **Port already in use**
   ```bash
   # Use a different port
   habibi-go server --port 8081
   ```

2. **Database issues**
   ```bash
   # Reset database
   make db-reset
   ```

3. **Claude Code not found**
   - Ensure Claude Code is installed: `npm install -g @anthropic/claude-cli`
   - Verify installation: `which claude`

## 📊 Monitoring

The application provides built-in monitoring:
- Health check endpoint: `GET /api/health`
- Project statistics: `GET /api/projects/stats`
- Session statistics: `GET /api/sessions/stats`

## 🔒 Security

- All user inputs are validated and sanitized
- File paths are restricted to project directories
- No credentials or secrets are stored in the codebase
- Uses secure WebSocket connections for real-time features

## 📄 License

This project is licensed under the MIT License - see below for details:

```
MIT License

Copyright (c) 2024 Habibi-Go Contributors

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

## 🙏 Acknowledgments

- Built with ❤️ using AI assistance
- Powered by [Claude](https://claude.ai) from Anthropic
- UI components inspired by modern development tools
- Community feedback and contributions

## 🔗 Links

- [Documentation](https://github.com/yourusername/habibi-go/wiki)
- [Issue Tracker](https://github.com/yourusername/habibi-go/issues)
- [Discussions](https://github.com/yourusername/habibi-go/discussions)

---

<p align="center">
Made with 🤖 by AI and humans working together
</p>