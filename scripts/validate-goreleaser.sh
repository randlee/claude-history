#!/usr/bin/env bash

# validate-goreleaser.sh
#
# Validates the .goreleaser.yml configuration before committing or creating a release.
# This script helps catch configuration errors early in the development process.
#
# Usage:
#   ./scripts/validate-goreleaser.sh
#
# Exit codes:
#   0 - Configuration is valid
#   1 - Configuration has errors
#   2 - Required tools not found

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo -e "${BLUE}=== GoReleaser Configuration Validator ===${NC}\n"

# Check if goreleaser is installed
if ! command -v goreleaser &> /dev/null; then
    echo -e "${RED}ERROR: goreleaser is not installed${NC}"
    echo ""
    echo "Install goreleaser:"
    echo "  macOS:   brew install goreleaser"
    echo "  Linux:   brew install goreleaser (or use apt/yum)"
    echo "  Windows: choco install goreleaser"
    echo "  Go:      go install github.com/goreleaser/goreleaser@latest"
    echo ""
    exit 2
fi

# Check if .goreleaser.yml exists
if [ ! -f "$PROJECT_ROOT/.goreleaser.yml" ]; then
    echo -e "${RED}ERROR: .goreleaser.yml not found${NC}"
    echo "Expected location: $PROJECT_ROOT/.goreleaser.yml"
    exit 1
fi

echo -e "${BLUE}Step 1: Checking goreleaser version${NC}"
goreleaser --version
echo ""

# Step 1: Run goreleaser check
echo -e "${BLUE}Step 2: Running 'goreleaser check'${NC}"
cd "$PROJECT_ROOT"

if goreleaser check; then
    echo -e "${GREEN}✓ Configuration syntax is valid${NC}"
    echo ""
else
    echo -e "${RED}✗ Configuration syntax check failed${NC}"
    echo ""
    echo "Fix the errors above in .goreleaser.yml"
    exit 1
fi

# Step 2: Try a snapshot build to validate the full configuration
echo -e "${BLUE}Step 3: Running snapshot build to validate full configuration${NC}"
echo "This will build for your current platform only..."
echo ""

if goreleaser build --snapshot --clean --single-target > /tmp/goreleaser-build.log 2>&1; then
    echo -e "${GREEN}✓ Snapshot build succeeded${NC}"
    echo ""

    # Show what was built
    if [ -d "$PROJECT_ROOT/dist" ]; then
        echo -e "${BLUE}Build artifacts:${NC}"
        find "$PROJECT_ROOT/dist" -name 'claude-history*' -type f -exec ls -lh {} \;
        echo ""
    fi
else
    echo -e "${RED}✗ Snapshot build failed${NC}"
    echo ""
    echo "Build log:"
    cat /tmp/goreleaser-build.log
    echo ""
    echo "This indicates a problem with the .goreleaser.yml configuration."
    echo "The build configuration may reference invalid paths, missing files,"
    echo "or use unsupported GoReleaser features."
    exit 1
fi

# Step 3: Check for common configuration issues
echo -e "${BLUE}Step 4: Checking for common configuration issues${NC}"

# Check for deprecated fields (add more as needed)
DEPRECATED_FIELDS=(
    "header_template_file"  # Not supported in v1.26.2, use header_template instead
)

for field in "${DEPRECATED_FIELDS[@]}"; do
    if grep -q "^[[:space:]]*${field}:" "$PROJECT_ROOT/.goreleaser.yml"; then
        echo -e "${RED}✗ Found deprecated/invalid field: ${field}${NC}"
        echo "  This field is not supported by GoReleaser v1.26.2+"
        echo "  Remove or replace it in .goreleaser.yml"
        exit 1
    fi
done

echo -e "${GREEN}✓ No deprecated fields found${NC}"
echo ""

# Step 4: Check that required files exist
echo -e "${BLUE}Step 5: Checking required files${NC}"

REQUIRED_FILES=(
    "LICENSE"
    "README.md"
    "CLAUDE.md"
    "src/go.mod"
)

ALL_EXIST=true
for file in "${REQUIRED_FILES[@]}"; do
    if [ ! -f "$PROJECT_ROOT/$file" ]; then
        echo -e "${YELLOW}⚠ Missing file referenced in .goreleaser.yml: ${file}${NC}"
        ALL_EXIST=false
    fi
done

if [ "$ALL_EXIST" = true ]; then
    echo -e "${GREEN}✓ All required files exist${NC}"
else
    echo -e "${YELLOW}⚠ Some files are missing but build may still succeed${NC}"
fi
echo ""

# Success
echo -e "${GREEN}=== All validation checks passed ===${NC}"
echo ""
echo "Your .goreleaser.yml configuration appears valid."
echo "You can now commit your changes or create a release tag."
echo ""
echo "To create a release:"
echo "  git tag -a v0.x.0 -m 'Release v0.x.0'"
echo "  git push origin v0.x.0"

exit 0
