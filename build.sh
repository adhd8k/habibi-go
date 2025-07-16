#!/bin/bash

# Build script for Habibi-Go
# Creates optimized binaries for Linux and macOS

set -e

echo "🚀 Building Habibi-Go..."

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Create bin directory if it doesn't exist
mkdir -p bin

# Build frontend first
echo -e "${BLUE}📦 Building frontend...${NC}"
cd web
npm install
npm run build
cd ..

# Get version from git tag or use dev
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(date +"%Y-%m-%d %H:%M:%S")
LDFLAGS="-s -w -X 'main.Version=$VERSION' -X 'main.BuildTime=$BUILD_TIME'"

# Function to build for a specific platform
build_platform() {
    GOOS=$1
    GOARCH=$2
    OUTPUT=$3
    
    echo -e "${YELLOW}🔨 Building for $GOOS/$GOARCH...${NC}"
    
    GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="$LDFLAGS" -o "$OUTPUT" main.go
    
    # Make executable (for non-Windows)
    if [[ "$GOOS" != "windows" ]]; then
        chmod +x "$OUTPUT"
    fi
    
    # Get file size
    SIZE=$(ls -lh "$OUTPUT" | awk '{print $5}')
    echo -e "${GREEN}✅ Built: $OUTPUT (${SIZE})${NC}"
}

# Build for all platforms
echo -e "${BLUE}🏗️  Building binaries...${NC}"

# Linux AMD64
build_platform "linux" "amd64" "bin/habibi-go-linux-amd64"

# macOS Intel
build_platform "darwin" "amd64" "bin/habibi-go-darwin-amd64"

# macOS Apple Silicon
build_platform "darwin" "arm64" "bin/habibi-go-darwin-arm64"

# Windows AMD64
build_platform "windows" "amd64" "bin/habibi-go-windows-amd64.exe"

# Create version file
echo "$VERSION" > bin/VERSION

echo -e "${GREEN}✅ Build complete!${NC}"
echo ""
echo "Binaries created:"
ls -lh bin/habibi-go-*
echo ""
echo "To run on your current platform:"
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    echo "  ./bin/habibi-go-linux-amd64 server"
elif [[ "$OSTYPE" == "darwin"* ]]; then
    if [[ $(uname -m) == "arm64" ]]; then
        echo "  ./bin/habibi-go-darwin-arm64 server"
    else
        echo "  ./bin/habibi-go-darwin-amd64 server"
    fi
fi