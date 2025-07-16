#!/bin/bash

# Release script for Habibi-Go
# Creates release artifacts for distribution

set -e

echo "🚀 Creating release artifacts for Habibi-Go..."

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Check if version is provided
if [ -z "$1" ]; then
    echo "Usage: ./release.sh <version>"
    echo "Example: ./release.sh v1.0.0"
    exit 1
fi

VERSION=$1
RELEASE_DIR="releases/$VERSION"

# Create release directory
mkdir -p "$RELEASE_DIR"

# Build all platforms
echo -e "${BLUE}🏗️  Building all platforms...${NC}"
./build.sh

# Create archives for each platform
echo -e "${BLUE}📦 Creating release archives...${NC}"

# Function to create archive
create_archive() {
    BINARY=$1
    PLATFORM=$2
    
    echo -e "${YELLOW}📦 Creating archive for $PLATFORM...${NC}"
    
    # Create temp directory
    TEMP_DIR="$RELEASE_DIR/temp-$PLATFORM"
    mkdir -p "$TEMP_DIR"
    
    # Copy files
    cp "bin/$BINARY" "$TEMP_DIR/"
    cp README.md "$TEMP_DIR/"
    cp LICENSE "$TEMP_DIR/"
    cp CLAUDE.md "$TEMP_DIR/"
    
    # Create config example
    cat > "$TEMP_DIR/config.example.yaml" << EOF
# Habibi-Go Configuration Example
server:
  host: "localhost"
  port: 8080

database:
  path: "~/.habibi-go/habibi.db"

projects:
  default_directory: "~/projects"
  auto_discover: false

logging:
  level: "info"
  format: "json"
EOF
    
    # Create archive
    cd "$RELEASE_DIR"
    if [[ "$PLATFORM" == "windows" ]]; then
        zip -r "habibi-go-$VERSION-$PLATFORM.zip" "temp-$PLATFORM"
    else
        tar -czf "habibi-go-$VERSION-$PLATFORM.tar.gz" "temp-$PLATFORM"
    fi
    cd - > /dev/null
    
    # Clean up
    rm -rf "$TEMP_DIR"
    
    echo -e "${GREEN}✅ Created: habibi-go-$VERSION-$PLATFORM${NC}"
}

# Create archives for each platform
create_archive "habibi-go-linux-amd64" "linux-amd64"
create_archive "habibi-go-darwin-amd64" "darwin-amd64"
create_archive "habibi-go-darwin-arm64" "darwin-arm64"
create_archive "habibi-go-windows-amd64.exe" "windows-amd64"

# Create checksums
echo -e "${BLUE}🔐 Creating checksums...${NC}"
cd "$RELEASE_DIR"
shasum -a 256 *.tar.gz *.zip > "habibi-go-$VERSION-checksums.txt"
cd - > /dev/null

# Create release notes template
echo -e "${BLUE}📝 Creating release notes template...${NC}"
cat > "$RELEASE_DIR/RELEASE_NOTES.md" << EOF
# Habibi-Go $VERSION

## What's New

- Feature 1
- Feature 2
- Bug fixes and improvements

## Installation

### Download Pre-built Binary

1. Download the appropriate archive for your platform
2. Extract the archive
3. Run the binary:
   \`\`\`bash
   ./habibi-go server
   \`\`\`

### Build from Source

\`\`\`bash
git clone https://github.com/yourusername/habibi-go.git
cd habibi-go
git checkout $VERSION
make build
\`\`\`

## Checksums

See \`habibi-go-$VERSION-checksums.txt\` for SHA256 checksums of all archives.

## Full Changelog

See [CHANGELOG.md](https://github.com/yourusername/habibi-go/blob/$VERSION/CHANGELOG.md)
EOF

echo -e "${GREEN}✅ Release artifacts created!${NC}"
echo ""
echo "Release files:"
ls -lh "$RELEASE_DIR"
echo ""
echo "Next steps:"
echo "1. Edit $RELEASE_DIR/RELEASE_NOTES.md with actual release notes"
echo "2. Create a new release on GitHub with tag $VERSION"
echo "3. Upload all files from $RELEASE_DIR to the release"