# claude-history

A Go CLI tool for programmatic access to Claude Code's agent history storage.

## Overview

`claude-history` maps between filesystem paths and Claude Code's internal storage format, enabling you to:

- Query conversation history with filters
- List projects and sessions
- Display agent hierarchy trees
- Find agents by task description
- Export sessions to HTML or JSONL

## Installation

### Windows (winget)

```powershell
winget install randlee.claude-history
```

### macOS (Homebrew)

```bash
brew tap randlee/tap
brew install claude-history
```

### Linux / macOS (Install Script)

```bash
curl -fsSL https://raw.githubusercontent.com/randlee/claude-history/main/install.sh | bash
```

### Go Install

```bash
go install github.com/randlee/claude-history/src@latest
```

### Build from Source

```bash
cd src
go build -o ../bin/claude-history .
```

### Cross-Platform Builds

```bash
cd src
GOOS=darwin GOARCH=arm64 go build -o ../bin/claude-history-darwin-arm64 .
GOOS=linux GOARCH=amd64 go build -o ../bin/claude-history-linux-amd64 .
GOOS=windows GOARCH=amd64 go build -o ../bin/claude-history-windows-amd64.exe .
```

## Quick Start

```bash
# List all sessions for a project
claude-history list /path/to/project

# Show agent hierarchy for a session
claude-history tree /path/to/project --session abc123

# Query conversation history
claude-history query /path/to/project --type user,assistant

# Query specific agent by ID (supports git-style prefixes)
claude-history query /path/to/project --session abc123 --agent def456

# Find agents working on specific topics
claude-history find-agent /path/to/project authentication api

# Export session to HTML
claude-history export /path/to/project --session abc123 --open
```

## Common Workflows

### Finding and Querying Agents

```bash
# 1. List sessions for a project
claude-history list /path/to/project

# 2. Show agent tree to see all agents in a session
claude-history tree /path/to/project --session abc123

# 3. Query specific agent's work (agent IDs support git-style prefixes)
claude-history query /path/to/project --session abc123 --agent def456
```

### Searching by Topic

```bash
# 1. Find agents working on specific topics
claude-history find-agent /path/to/project authentication database

# 2. Query the session containing relevant agents
claude-history query /path/to/project --session abc123 --type assistant
```

### Exporting Session Reports

```bash
# 1. Identify the session you want to export
claude-history list /path/to/project

# 2. Export to HTML and open in browser
claude-history export /path/to/project --session abc123 --open
```

## Commands

### `list`
List all projects or sessions in a project:
```bash
claude-history list
claude-history list /path/to/project
```

### `query`
Query conversation history with filters:
```bash
claude-history query /path/to/project \
  --session abc123 \
  --type user,assistant \
  --start 2026-01-01 \
  --tool Read \
  --format json
```

**Flags:**
- `--session <id>` - Filter by session ID (supports prefixes)
- `--agent <id>` - Filter by agent ID (supports prefixes)
- `--type <types>` - Filter by entry type (user, assistant, system, etc.)
- `--start <date>` - Show entries after date (YYYY-MM-DD)
- `--end <date>` - Show entries before date
- `--tool <name>` - Filter by exact tool name
- `--tool-match <pattern>` - Filter by tool name regex
- `--format <fmt>` - Output format: text, json, tree

### `tree`
Display agent hierarchy:
```bash
claude-history tree /path/to/project --session abc123
```

### `find-agent`
Search for agents by task description:
```bash
claude-history find-agent /path/to/project [search terms...]
```

### `export`
Export session to HTML or JSONL:
```bash
claude-history export /path/to/project \
  --session abc123 \
  --output report.html \
  --open
```

**Flags:**
- `--output <file>` - Output file path
- `--format <fmt>` - Export format: html, jsonl
- `--open` - Open HTML in browser after export
- `--template <file>` - Custom HTML template

### `resolve`
Resolve filesystem paths to Claude storage (debugging):
```bash
claude-history resolve /path/to/project
claude-history resolve /path/to/project --session abc123 --agent def456
```

## Claude Code Skill

This project includes a Claude Code skill for easy querying from within Claude sessions.

### Installation

The skill is located in `.claude/skills/history/` and is automatically available when working in this project.

To use in other projects:
```bash
cp -r .claude/skills/history /path/to/your/project/.claude/skills/
```

### Usage

```
/history
/history action=list path=/Users/name/project
/history action=tree session=abc123
/history action=find-agent authentication
```

See [.claude/skills/history/README.md](.claude/skills/history/README.md) for details.

## Claude Code Storage Format

Claude Code stores history in `~/.claude/projects/` using dash-encoded paths:

```
/Users/name/project       → -Users-name-project
C:\Users\name\project     → C--Users-name-project
```

### Session Structure

```
~/.claude/projects/{encoded-path}/
├── sessions-index.json              # Metadata cache
├── {sessionId}.jsonl                # Main session
└── {sessionId}/
    └── subagents/
        └── agent-{agentId}.jsonl    # Subagent sessions
```

### Entry Types

| Type | Description |
|------|-------------|
| `user` | User messages (prompts or tool results) |
| `assistant` | Claude responses with text and tool_use |
| `system` | System events and hook summaries |
| `queue-operation` | Subagent spawn triggers |
| `progress` | Status updates |
| `file-history-snapshot` | File state captures |
| `summary` | Conversation summaries |

## Development

### Running Tests

```bash
cd src
go test ./...
go test ./... -cover
```

### Project Structure

```
claude-history/
├── .claude/
│   └── skills/
│       └── history/          # Claude Code skill
├── src/
│   ├── cmd/                  # CLI commands
│   ├── pkg/                  # Public packages
│   └── internal/             # Private implementation
├── test/fixtures/            # Test data
├── docs/                     # Documentation
└── bin/                      # Built binaries
```

See [CLAUDE.md](CLAUDE.md) for detailed project instructions and [docs/PROJECT_PLAN.md](docs/PROJECT_PLAN.md) for implementation status.

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Related Projects

- **[Claude Code](https://claude.com/claude-code)** - The CLI tool whose history we're querying
- **claude-code-viewer** - Web UI for viewing Claude Code history (separate project)
