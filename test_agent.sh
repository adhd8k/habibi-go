#!/bin/bash

# Test script for Phase 3: Agent Orchestration

echo "=== Habibi-Go Agent Orchestration Test ==="
echo ""

# Build the project
echo "Building Habibi-Go..."
go build -o bin/habibi-go main.go || exit 1

# Create a test project
echo ""
echo "1. Creating test project..."
./bin/habibi-go project create "Agent Test Project" "/tmp/habibi-test-project" --repo "https://github.com/example/test.git"

# List projects
echo ""
echo "2. Listing projects..."
./bin/habibi-go project list

# Create a session
echo ""
echo "3. Creating test session..."
./bin/habibi-go session create "Agent Test Project" "test-session-1" "feature-test"

# List sessions
echo ""
echo "4. Listing sessions..."
./bin/habibi-go session list 1

# Start an agent
echo ""
echo "5. Starting test agent..."
./bin/habibi-go agent start 1 "test-agent" "./test_agent_script.sh"
sleep 2  # Allow agent to start

# List agents
echo ""
echo "6. Listing agents..."
./bin/habibi-go agent list 1

# Send a command to the agent
echo ""
echo "7. Sending command to agent..."
./bin/habibi-go agent exec 1 "echo 'Hello from agent'"
sleep 1

# Check agent status
echo ""
echo "8. Checking agent status..."
./bin/habibi-go agent status 1
sleep 1

# Stop the agent
echo ""
echo "9. Stopping agent..."
./bin/habibi-go agent stop 1

# Final status check
echo ""
echo "10. Final agent list..."
./bin/habibi-go agent list 1

echo ""
echo "=== Test Complete ==="