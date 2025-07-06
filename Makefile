.PHONY: build clean dev test install deps build-ui embed cross-compile

# Development
dev:
	@echo "Starting development environment..."
	cd web && npm run dev &
	go run main.go server --dev

# Dependencies
deps:
	go mod download
	cd web && npm install

# Build frontend (placeholder for now)
build-ui:
	@echo "Building UI... (React implementation in Phase 4)"
	mkdir -p web/dist
	cp web/dist/index.html web/dist/index.html || true

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