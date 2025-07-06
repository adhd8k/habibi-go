#!/bin/bash

# Simple test agent that responds to commands via stdin/stdout

echo "Test agent started. PID: $$"
echo "Ready to receive commands..."

while IFS= read -r line; do
    echo "Received command: $line"
    
    case "$line" in
        "status")
            echo "Agent is running and healthy"
            ;;
        "info")
            echo "Test agent v1.0 - PID: $$"
            ;;
        "exit")
            echo "Agent shutting down..."
            exit 0
            ;;
        *)
            echo "Executing: $line"
            eval "$line" 2>&1 || echo "Command failed with exit code: $?"
            ;;
    esac
    
    echo "---"
done