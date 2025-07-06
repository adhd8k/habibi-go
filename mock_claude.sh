#!/bin/bash

# Mock Claude Code CLI for testing
# This simulates the stdio interface of Claude Code

SESSION_ID=$(uuidgen 2>/dev/null || echo "session-$(date +%s)")
echo "Claude Code Mock (Session: $SESSION_ID)"
echo "Ready for conversation. Type 'exit' to quit."
echo ""

while IFS= read -r line; do
    # Handle special commands
    case "$line" in
        "exit"|"quit"|"bye")
            echo "Goodbye!"
            exit 0
            ;;
        "--help"|"help")
            echo "Available commands:"
            echo "  help - Show this help"
            echo "  exit - Exit the conversation"
            echo "  clear - Clear the screen"
            echo ""
            continue
            ;;
        "clear")
            clear
            echo "Claude Code Mock (Session: $SESSION_ID)"
            echo ""
            continue
            ;;
    esac
    
    # Simulate processing
    echo "[Processing...]"
    sleep 0.5
    
    # Generate mock response based on input
    if [[ "$line" == *"hello"* ]] || [[ "$line" == *"hi"* ]]; then
        echo "Hello! I'm Claude, your AI coding assistant. How can I help you today?"
    elif [[ "$line" == *"create"* ]] && [[ "$line" == *"function"* ]]; then
        echo "I'll help you create a function. Here's an example:"
        echo ""
        echo '```javascript'
        echo "function exampleFunction(param1, param2) {"
        echo "    // Function implementation"
        echo "    return param1 + param2;"
        echo "}"
        echo '```'
    elif [[ "$line" == *"test"* ]]; then
        echo "I'll help you with testing. What would you like to test?"
        echo "- Unit tests"
        echo "- Integration tests"
        echo "- End-to-end tests"
    else
        echo "I understand you're asking about: \"$line\""
        echo "As a mock Claude, I'm providing a simulated response."
        echo "In a real implementation, I would analyze your request and provide detailed assistance."
    fi
    
    echo ""
    echo "---"
    echo ""
done