#!/usr/bin/env bash

# pre-commit hook template for claude-history
#
# This hook validates the .goreleaser.yml configuration before committing changes.
#
# To install this hook:
#   cp scripts/pre-commit-hook.sh .git/hooks/pre-commit
#   chmod +x .git/hooks/pre-commit
#
# To temporarily bypass this hook:
#   git commit --no-verify

# Only run if .goreleaser.yml is being committed
if git diff --cached --name-only | grep -q "^.goreleaser.yml$"; then
    echo "Validating .goreleaser.yml before commit..."

    # Run validation script
    if ! ./scripts/validate-goreleaser.sh; then
        echo ""
        echo "ERROR: .goreleaser.yml validation failed"
        echo "Please fix the errors above before committing."
        echo ""
        echo "To bypass this check (not recommended):"
        echo "  git commit --no-verify"
        exit 1
    fi

    echo "✓ .goreleaser.yml validation passed"
fi

exit 0
