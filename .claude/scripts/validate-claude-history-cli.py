#!/usr/bin/env python3
"""
PreToolUse hook to validate claude-history CLI is installed.

Returns:
    0: Tool is available
    2: Tool is missing (blocks execution)
"""
import sys
import shutil

def check_cli_installed():
    """Check if claude-history CLI is available in PATH."""
    if shutil.which('claude-history') is None:
        print("ERROR: claude-history CLI tool not found", file=sys.stderr)
        print("", file=sys.stderr)
        print("The history-search agent requires the claude-history CLI to be installed.", file=sys.stderr)
        print("", file=sys.stderr)
        print("Installation instructions: .claude/skills/history/README.md", file=sys.stderr)
        print("", file=sys.stderr)
        print("Quick install:", file=sys.stderr)
        print("  cd src && go build -o claude-history .", file=sys.stderr)
        print("  # Then add to PATH or use full path", file=sys.stderr)
        return 2
    return 0

if __name__ == "__main__":
    sys.exit(check_cli_installed())
