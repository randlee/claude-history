#!/usr/bin/env python3
"""
PreToolUse hook to validate claude-history CLI is installed.

Returns:
    0: Tool is available
    2: Tool is missing (blocks execution)
"""
import sys
import shutil
from pathlib import Path

try:
    import yaml
except ImportError:
    yaml = None

def find_cli():
    """
    Find claude-history CLI path.

    Priority:
    1. Check PATH (Homebrew, winget, go install)
    2. Check .sc/history/config.yml (custom local build)

    Returns:
        str: Path to CLI if found, None otherwise
    """
    # 1. Check PATH (happy path)
    cli = shutil.which('claude-history')
    if cli:
        return cli

    # 2. Check .sc/history/config.yml (local build fallback)
    config_path = Path('.sc/history/config.yml')
    if config_path.exists() and yaml:
        try:
            with open(config_path) as f:
                config = yaml.safe_load(f)
                if config and 'cli' in config and 'path' in config['cli']:
                    custom_path = Path(config['cli']['path'])
                    if custom_path.exists():
                        return str(custom_path)
        except Exception:
            pass

    return None

def check_cli_installed():
    """Check if claude-history CLI is available."""
    if find_cli() is None:
        print("ERROR: claude-history CLI tool not found", file=sys.stderr)
        print("", file=sys.stderr)
        print("The history-search agent requires the claude-history CLI to be installed.", file=sys.stderr)
        print("", file=sys.stderr)
        print("Installation instructions: .claude/skills/history/README.md", file=sys.stderr)
        print("", file=sys.stderr)
        print("Recommended install methods (add to PATH automatically):", file=sys.stderr)
        print("  - Homebrew: brew install randlee/tap/claude-history", file=sys.stderr)
        print("  - winget: winget install randlee.claude-history", file=sys.stderr)
        print("  - Go: go install github.com/randlee/claude-history/src@latest", file=sys.stderr)
        print("", file=sys.stderr)
        print("For local builds, create .sc/history/config.yml:", file=sys.stderr)
        print("  cli:", file=sys.stderr)
        print("    path: /full/path/to/claude-history", file=sys.stderr)
        return 2
    return 0

if __name__ == "__main__":
    sys.exit(check_cli_installed())
