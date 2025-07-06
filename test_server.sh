#!/bin/bash

echo "Starting Habibi-Go server..."
echo "Server will run on http://localhost:8080"
echo ""

# Clean database for fresh start
rm -f habibi.db habibi.db-wal habibi.db-shm

# Start server
./bin/habibi-go server