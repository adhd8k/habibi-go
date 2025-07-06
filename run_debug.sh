#!/bin/bash

echo "Starting Habibi-Go with debugging..."
echo ""
echo "1. Open http://localhost:8080 in your browser"
echo "2. Open the browser's Developer Console (F12)"
echo "3. Look for any error messages in the console"
echo "4. The API responses will be logged to help debug the issue"
echo ""
echo "Server starting..."
echo ""

# Clean database for fresh start
rm -f habibi.db habibi.db-wal habibi.db-shm

# Run server
./bin/habibi-go server