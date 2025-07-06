#!/usr/bin/env bash
# Claude Code CLI Simulator
# This simulates Claude Code CLI for testing purposes

echo "Claude Code v0.1.0"
echo "Type 'help' for available commands or start typing to chat."
echo ""

# Function to simulate Claude thinking
think() {
    echo -n "Thinking"
    for i in {1..3}; do
        sleep 0.3
        echo -n "."
    done
    echo ""
}

# Main loop
while IFS= read -r line; do
    # Skip empty lines
    if [ -z "$line" ]; then
        continue
    fi
    
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
            # Simulate a thoughtful response
            think
            echo "I understand you're asking about: \"$line\""
            echo ""
            echo "As a simulated Claude, I would help you with:"
            echo "- Code analysis and understanding"
            echo "- Writing and refactoring code"
            echo "- Debugging and problem solving"
            echo "- Best practices and recommendations"
            echo ""
            echo "This is a test response to demonstrate the chat interface is working."
            echo ""
            ;;
    esac
done