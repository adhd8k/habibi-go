#!/bin/bash

echo "=== Testing Habibi-Go Full Stack ==="
echo ""

# Clean up
echo "Cleaning up..."
rm -f habibi.db habibi.db-wal habibi.db-shm

# Build everything
echo "Building application..."
bash build_web.sh

# Start the server in background
echo ""
echo "Starting server..."
./bin/habibi-go server &
SERVER_PID=$!

# Wait for server to start
echo "Waiting for server to start..."
sleep 3

# Test API endpoints
echo ""
echo "Testing API endpoints..."
echo "1. Health check:"
curl -s http://localhost:8080/api/health | jq .

echo ""
echo "2. Creating test project:"
curl -s -X POST http://localhost:8080/api/v1/projects \
  -H "Content-Type: application/json" \
  -d '{"name": "WebUI Test", "path": "/tmp/webui-test", "repository_url": "https://github.com/test/repo"}' | jq .

echo ""
echo "3. Listing projects:"
curl -s http://localhost:8080/api/v1/projects | jq .

echo ""
echo "4. Testing web UI:"
curl -s -I http://localhost:8080/

echo ""
echo "Server is running on http://localhost:8080"
echo "Press Ctrl+C to stop the server"
echo ""

# Wait for user to stop
wait $SERVER_PID