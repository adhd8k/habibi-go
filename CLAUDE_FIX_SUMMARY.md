# Claude Integration Fix Summary

## Issues Fixed

1. **Process Death Detection**: Added detailed logging to understand why Claude was dying
2. **Binary Path Configuration**: Implemented proper configuration support for Claude binary path
3. **Environment Setup**: Added proper environment variables (TERM=dumb, NO_COLOR=1)
4. **Error Handling**: Added immediate exit detection with 500ms timeout
5. **Simulator for Testing**: Created claude-simulator.sh to test the integration

## Implementation Details

### 1. Agent Service Updates
- Added `claudeBinaryPath` field to AgentService
- Added `SetClaudeBinaryPath()` method to configure the path
- Enhanced Claude detection to check configured path first, then PATH, then common locations
- Added detailed logging for process startup and death

### 2. Configuration
- The `config.yaml` now supports `agents.claude_binary_path` setting
- Default is "claude" (looks in PATH)
- Can specify absolute path to Claude binary or simulator

### 3. Claude Simulator
Created `/claude-simulator.sh` for testing:
- Simulates Claude Code CLI behavior
- Responds to input with simulated responses
- Useful for testing the integration without real Claude

### 4. Process Monitoring
- Added stderr capture for Claude to debug startup issues
- Added process exit code logging
- Improved error messages

## Testing

1. **With Simulator**:
   ```yaml
   agents:
     claude_binary_path: "/home/moe/mnt/d/code/habibi-go/claude-simulator.sh"
   ```

2. **With Real Claude**:
   ```yaml
   agents:
     claude_binary_path: "/home/moe/.npm-packages/bin/claude"
   ```

## Next Steps

If Claude is still dying with the real binary:
1. Check Claude's requirements (may need PTY support)
2. Review Claude's command-line flags and options
3. Consider using Claude's programmatic API mode if available
4. Add PTY support using a library like github.com/creack/pty

## Usage

1. Configure the Claude path in `config.yaml`
2. Run the server: `./bin/habibi-go server`
3. Create a project and session in the UI
4. Claude will automatically start when the session is created
5. Use the chat interface to interact with Claude