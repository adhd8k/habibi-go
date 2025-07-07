#!/usr/bin/env bash
# Claude Code CLI Simulator
# This simulates Claude Code CLI for testing purposes

echo "Claude Code v0.1.0"
echo "Type 'help' for available commands or start typing to chat."
echo ""

# Function to simulate Claude thinking and responding
respond() {
    local input="$1"
    
    # Simulate thinking delay
    sleep 0.5
    
    # Generate a multi-line response
    echo "I understand you're asking about: \"$input\""
    echo ""
    echo "Here's my response:"
    echo "- This is line 1 of the response"
    echo "- This is line 2 showing streaming works"
    sleep 0.2
    echo "- Each line appears separately"
    echo "- Demonstrating real-time output"
    echo ""
    echo "The chat interface is working correctly!"
}

# Main loop
while IFS= read -r line; do
    # Skip empty lines
    if [ -z "$line" ]; then
        continue
    fi
    
    # Log received input for debugging
    # echo "[Received: $line]" >&2
    
    # Simulate different responses based on input
    case "$line" in
        "help"|"?")
            echo "Available commands:"
            echo "  help     - Show this help message"
            echo "  exit     - Exit Claude Code"
            echo "  clear    - Clear the screen"
            echo ""
            echo "Or just type your question and I'll help you!"
            ;;
        "exit"|"quit")
            echo "Goodbye!"
            exit 0
            ;;
        "clear")
            clear
            echo "Claude Code v0.1.0"
            echo ""
            ;;
        *)
            # Simulate a response with streaming
            respond "$line"
            ;;
    esac
done