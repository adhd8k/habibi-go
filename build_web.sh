#!/bin/bash

echo "Building Habibi-Go Web UI..."

# Build the React app
cd web
npm run build

# Go back to root
cd ..

# Build the Go app with embedded web assets
echo "Building Go application with embedded web assets..."
go build -o bin/habibi-go main.go

echo "Build complete!"