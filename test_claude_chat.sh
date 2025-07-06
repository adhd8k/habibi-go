#!/bin/bash

echo "=== Testing Claude Chat Integration ==="
echo ""
echo "This test will:"
echo "1. Start the server"
echo "2. Create a project and session"
echo "3. Automatically start a Claude agent"
echo "4. Allow you to chat with Claude through the web UI"
echo ""

# Clean database for fresh start
rm -f habibi.db habibi.db-wal habibi.db-shm

# Create test project directory
mkdir -p /tmp/claude-test-project
cd /tmp/claude-test-project
git init --initial-branch=main
echo "# Claude Test Project" > README.md
git add README.md
git config user.email "test@example.com"
git config user.name "Test User"
git commit -m "Initial commit"
cd -

echo "Starting server..."
echo "Open http://localhost:8080 in your browser"
echo ""
echo "Instructions:"
echo "1. Create a new project pointing to /tmp/claude-test-project"
echo "2. Create a session (Claude will auto-connect)"
echo "3. Start chatting with Claude!"
echo ""

./bin/habibi-go server