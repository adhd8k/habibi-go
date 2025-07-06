.PHONY: build clean dev test install deps build-ui embed cross-compile

# Development
dev:
	@echo "Starting development environment..."
	@echo "Starting Vite dev server on http://localhost:3000"
	@echo "Starting Go server on http://localhost:8080"
	@echo "Press Ctrl+C to stop"
	@trap 'kill %1' INT; \
	cd web && npm run dev & \
	go run main.go server --dev

# Dependencies
deps:
	go mod download
	cd web && npm install

# Build frontend
build-ui:
	@echo "Building React UI..."
	cd web && npm run build

# Embed assets and build
build: build-ui
	go build -ldflags="-s -w" -o bin/habibi-go main.go

# Cross-compile for all platforms
cross-compile: build-ui
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/habibi-go-linux-amd64 main.go
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o bin/habibi-go-darwin-amd64 main.go
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o bin/habibi-go-darwin-arm64 main.go
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o bin/habibi-go-windows-amd64.exe main.go

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf web/dist/

# Run tests
test:
	go test ./...
	cd web && npm test || true

# Install locally
install: build
	cp bin/habibi-go /usr/local/bin/

# Quick start
start: build
	./bin/habibi-go server

# CLI commands for testing
demo: build
	@echo "=== Habibi-Go Demo ==="
	@echo "1. Listing projects:"
	./bin/habibi-go project list
	@echo ""
	@echo "2. Show configuration:"
	./bin/habibi-go config show
	@echo ""
	@echo "3. Available commands:"
	./bin/habibi-go --help

# Run full stack locally
run: build
	@echo "=== Starting Habibi-Go ==="
	@echo "Server running on http://localhost:8080"
	@echo "Press Ctrl+C to stop"
	./bin/habibi-go server

# Quick test of all phases
test-all: build
	@echo "=== Testing All Phases ==="
	@echo "Creating test project..."
	@mkdir -p /tmp/test-project && cd /tmp/test-project && git init
	./bin/habibi-go project create "Test Project" "/tmp/test-project"
	@echo ""
	@echo "Creating session..."
	./bin/habibi-go session create "Test Project" "test-session" "feature-test"
	@echo ""
	@echo "Starting agent..."
	./bin/habibi-go agent start 1 "test-agent" "echo 'Hello from agent'"
	@echo ""
	@echo "Server is ready at http://localhost:8080"
	./bin/habibi-go server

# Database management
db-reset:
	@echo "Resetting database..."
	rm -f habibi.db habibi.db-wal habibi.db-shm
	@echo "Database reset complete"

# Development with auto-reload
watch:
	@which air > /dev/null || (echo "Installing air..." && go install github.com/cosmtrek/air@latest)
	@echo "Starting with auto-reload..."
	air