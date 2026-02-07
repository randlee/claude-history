#!/bin/bash
set -e

# Installation script for claude-history CLI tool

REPO="randlee/claude-history"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
  x86_64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)
    echo -e "${RED}Error: Unsupported architecture: $ARCH${NC}"
    exit 1
    ;;
esac

# Get version (default to latest)
VERSION=${1:-latest}
if [ "$VERSION" = "latest" ]; then
  echo "Fetching latest version..."
  VERSION=$(curl -s https://api.github.com/repos/$REPO/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
  if [ -z "$VERSION" ]; then
    echo -e "${RED}Error: Could not fetch latest version${NC}"
    exit 1
  fi
fi

echo -e "${GREEN}Installing claude-history $VERSION for $OS/$ARCH...${NC}"

# Construct download URL
DOWNLOAD_URL="https://github.com/$REPO/releases/download/${VERSION}/claude-history_${VERSION#v}_${OS}_${ARCH}.tar.gz"

# Download and extract
TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"

echo "Downloading from $DOWNLOAD_URL..."
if ! curl -fsSL "$DOWNLOAD_URL" -o claude-history.tar.gz; then
  echo -e "${RED}Error: Download failed${NC}"
  exit 1
fi

echo "Extracting..."
tar xzf claude-history.tar.gz

# Install binary
echo "Installing to $INSTALL_DIR..."
if [ -w "$INSTALL_DIR" ]; then
  mv claude-history "$INSTALL_DIR/"
else
  sudo mv claude-history "$INSTALL_DIR/"
fi

chmod +x "$INSTALL_DIR/claude-history"

# Cleanup
cd - > /dev/null
rm -rf "$TMP_DIR"

echo -e "${GREEN}✓ claude-history installed successfully!${NC}"
echo ""
echo "Run 'claude-history --help' to get started"
echo ""

# Verify installation
if command -v claude-history >/dev/null 2>&1; then
  claude-history --version
else
  echo -e "${YELLOW}Warning: $INSTALL_DIR may not be in your PATH${NC}"
  echo "Add this to your shell profile:"
  echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
fi
