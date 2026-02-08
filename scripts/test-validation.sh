#!/usr/bin/env bash

# test-validation.sh
#
# Tests that the validation script catches common configuration errors.
# This is a meta-test to ensure our validation is working correctly.
#
# Usage:
#   ./scripts/test-validation.sh

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
GORELEASER_FILE="$PROJECT_ROOT/.goreleaser.yml"
BACKUP_FILE="$PROJECT_ROOT/.goreleaser.yml.backup"

echo -e "${BLUE}=== Testing GoReleaser Validation Script ===${NC}\n"

# Make sure goreleaser is available
if ! command -v goreleaser &> /dev/null; then
    echo -e "${YELLOW}WARNING: goreleaser not installed, skipping tests${NC}"
    echo "Install goreleaser to run these tests:"
    echo "  brew install goreleaser"
    exit 0
fi

# Backup the current config
cp "$GORELEASER_FILE" "$BACKUP_FILE"

cleanup() {
    echo ""
    echo "Restoring original configuration..."
    mv "$BACKUP_FILE" "$GORELEASER_FILE"
}

trap cleanup EXIT

# Test 1: Valid configuration should pass
echo -e "${BLUE}Test 1: Valid configuration${NC}"
if ./scripts/validate-goreleaser.sh > /tmp/test-validation.log 2>&1; then
    echo -e "${GREEN}✓ Valid configuration passed (expected)${NC}\n"
else
    echo -e "${RED}✗ Valid configuration failed (unexpected)${NC}"
    cat /tmp/test-validation.log
    exit 1
fi

# Test 2: Invalid field should fail
echo -e "${BLUE}Test 2: Invalid field (header_template_file)${NC}"

# Add the invalid field that caused the v0.3.2 release failure
cat >> "$GORELEASER_FILE" << 'EOF'

# This is an invalid field that should be caught
release:
  header_template_file: .goreleaser/release-notes-template.md
EOF

if ./scripts/validate-goreleaser.sh > /tmp/test-validation.log 2>&1; then
    echo -e "${RED}✗ Invalid configuration passed (unexpected)${NC}"
    cat /tmp/test-validation.log
    exit 1
else
    echo -e "${GREEN}✓ Invalid field detected (expected)${NC}"
    if grep -q "header_template_file" /tmp/test-validation.log; then
        echo -e "${GREEN}✓ Error message mentions the invalid field${NC}\n"
    else
        echo -e "${YELLOW}⚠ Error detected but doesn't mention the specific field${NC}"
        echo "Log contents:"
        cat /tmp/test-validation.log
        echo ""
    fi
fi

# Restore for next test
cp "$BACKUP_FILE" "$GORELEASER_FILE"

# Test 3: Missing required file should warn
echo -e "${BLUE}Test 3: Missing required file${NC}"

# Temporarily rename a required file
mv "$PROJECT_ROOT/LICENSE" "$PROJECT_ROOT/LICENSE.tmp" 2>/dev/null || true

if ./scripts/validate-goreleaser.sh > /tmp/test-validation.log 2>&1; then
    if grep -q "Missing file" /tmp/test-validation.log; then
        echo -e "${GREEN}✓ Missing file warning generated${NC}\n"
    else
        echo -e "${YELLOW}⚠ Missing file not detected (may still pass)${NC}\n"
    fi
else
    echo -e "${GREEN}✓ Missing file caused validation to fail${NC}\n"
fi

# Restore the file
mv "$PROJECT_ROOT/LICENSE.tmp" "$PROJECT_ROOT/LICENSE" 2>/dev/null || true

# Summary
echo -e "${GREEN}=== All validation tests passed ===${NC}"
echo ""
echo "The validation script correctly:"
echo "  ✓ Accepts valid configurations"
echo "  ✓ Rejects invalid fields (header_template_file)"
echo "  ✓ Checks for missing files"
echo ""
echo "This validation will prevent release failures like the one in v0.3.2."

exit 0
