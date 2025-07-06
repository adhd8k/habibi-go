#!/bin/bash

# Clean test script for Phase 3: Agent Orchestration

echo "=== Habibi-Go Agent Orchestration Test ==="
echo ""

# Clean up previous test data
echo "Cleaning up previous test data..."
rm -f habibi.db
rm -rf /tmp/habibi-test-project
mkdir -p /tmp/habibi-test-project

# Build the project
echo "Building Habibi-Go..."
go build -o bin/habibi-go main.go || exit 1

# Initialize a git repo in the test project
echo ""
echo "Initializing test git repository..."
cd /tmp/habibi-test-project
git init --initial-branch=main
echo "# Test Project" > README.md
git add README.md
git config user.email "test@example.com"
git config user.name "Test User"
git commit -m "Initial commit"
cd -

# Create a test project
echo ""
echo "1. Creating test project..."
./bin/habibi-go project create "AgentTestProject" "/tmp/habibi-test-project"

# List projects
echo ""
echo "2. Listing projects..."
./bin/habibi-go project list

# Get project ID
PROJECT_ID=$(./bin/habibi-go project list | grep AgentTestProject | awk '{print $1}')
echo "Project ID: $PROJECT_ID"

# Create a session
echo ""
echo "3. Creating test session..."
./bin/habibi-go session create "AgentTestProject" "test-session-1" "feature-test"

# List sessions
echo ""
echo "4. Listing sessions..."
./bin/habibi-go session list $PROJECT_ID

# Get session ID
SESSION_ID=$(./bin/habibi-go session list $PROJECT_ID | grep test-session-1 | awk '{print $1}')
echo "Session ID: $SESSION_ID"

# Start an agent
echo ""
echo "5. Starting test agent..."
./bin/habibi-go agent start $SESSION_ID "test-agent" "./test_agent_script.sh"
sleep 3  # Allow agent to fully start

# List agents
echo ""
echo "6. Listing agents..."
./bin/habibi-go agent list $SESSION_ID

# Get agent ID
AGENT_ID=$(./bin/habibi-go agent list $SESSION_ID | grep test-agent | grep running | head -1 | awk '{print $1}')
echo "Agent ID: $AGENT_ID"

if [ -z "$AGENT_ID" ]; then
    echo "Error: No running agent found"
    exit 1
fi

# Send a command to the agent
echo ""
echo "7. Sending command to agent..."
./bin/habibi-go agent exec $AGENT_ID "echo 'Hello from agent'"
sleep 1

# Send another command
echo ""
echo "8. Sending info command..."
./bin/habibi-go agent exec $AGENT_ID "info"
sleep 1

# Check agent status
echo ""
echo "9. Checking agent status..."
./bin/habibi-go agent status $AGENT_ID

# Stop the agent
echo ""
echo "10. Stopping agent..."
./bin/habibi-go agent stop $AGENT_ID

# Final status check
echo ""
echo "11. Final agent list..."
./bin/habibi-go agent list $SESSION_ID

echo ""
echo "=== Test Complete ==="