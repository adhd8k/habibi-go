# Claude Integration

Habibi-Go now has full integration with Claude Code CLI for AI-powered coding assistance.

## Features

1. **Automatic Claude Connection**: When you create a new session, Claude automatically starts and connects
2. **Real-time Chat Interface**: Clean chat UI with message history and live streaming responses
3. **Smart Binary Detection**: Automatically finds Claude in PATH or common installation locations
4. **WebSocket Communication**: Real-time bidirectional communication between UI and Claude

## Setup

### 1. Install Claude Code CLI

First, ensure you have Claude Code CLI installed:
```bash
# Check if claude is installed
which claude

# If not found, install Claude Code CLI from Anthropic
```

### 2. Configure Claude Path (Optional)

If Claude is installed in a non-standard location, you can configure it in `config.yaml`:

```yaml
agents:
  claude_binary_path: "/path/to/claude"
```

### 3. Run Habibi-Go

```bash
./bin/habibi-go server
```

## Usage

1. **Create a Project**: Click "New Project" and provide a name and path (must be a Git repository)
2. **Create a Session**: Select your project and create a session with a branch name
3. **Chat with Claude**: Claude automatically connects when the session is created
4. **Send Messages**: Type your questions or requests in the chat interface

## How It Works

- When a session is created, the backend starts a Claude process with `agent_type: 'claude-code'`
- The agent service looks for the `claude` binary in:
  - System PATH
  - `/usr/local/bin/claude`
  - `/usr/bin/claude`
  - `/opt/claude/bin/claude`
  - Configuration file path (if specified)
- Communication happens via stdio pipes
- WebSocket streams the output to the browser in real-time
- The chat interface handles message formatting and display

## Troubleshooting

### Claude Not Found
If you see "claude binary not found", ensure:
1. Claude Code CLI is installed
2. It's available in your PATH: `export PATH=$PATH:/path/to/claude/bin`
3. Or configure the path in `config.yaml`

### Process Dies Immediately
This usually means:
1. Claude binary is not executable: `chmod +x /path/to/claude`
2. Missing dependencies or incorrect installation
3. Check server logs for detailed error messages

### No Response from Claude
1. Check if the agent is running in the UI (should show "Connected")
2. Look at server logs for any errors
3. Try restarting the agent from the UI

## Configuration Example

```yaml
# ~/.habibi-go/config.yaml
agents:
  claude_binary_path: "/usr/local/bin/claude"
  default_timeout: "30m"
  max_concurrent: 10
```