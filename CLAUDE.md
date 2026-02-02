# Claude's Role: claude-history CLI Developer

## Project Overview

**claude-history** is a Go CLI tool that provides programmatic access to Claude Code's agent history storage. It maps between filesystem paths and Claude's internal storage format, enabling querying, filtering, and export of conversation history including subagent sessions and tool calls.

## Primary Responsibilities

1. **Maintain and extend the CLI tool** - Add features, fix bugs, improve performance
2. **Ensure cross-platform compatibility** - Windows, macOS, Linux support
3. **Write comprehensive tests** - Unit tests for all new code (>80% coverage target)
4. **Follow Go best practices** - Standard project layout, idiomatic Go code

## Repository Structure

```
claude-history/
├── CLAUDE.md              # This file - project instructions
├── src/                   # Go source code
│   ├── go.mod             # Go module definition
│   ├── go.sum             # Dependency lock file
│   ├── main.go            # CLI entry point
│   ├── cmd/               # Cobra CLI commands
│   │   ├── root.go        # Root command, global flags
│   │   ├── resolve.go     # Path resolution command
│   │   ├── list.go        # List projects/sessions
│   │   ├── query.go       # Query history
│   │   └── tree.go        # Agent hierarchy tree
│   ├── pkg/               # Public packages (importable)
│   │   ├── encoding/      # Path encoding/decoding
│   │   ├── paths/         # Claude directory resolution
│   │   ├── session/       # Session operations
│   │   ├── agent/         # Agent discovery and tree building
│   │   └── models/        # Data structures
│   └── internal/          # Private implementation
│       ├── jsonl/         # Streaming JSONL parser
│       └── output/        # Output formatters
├── test/                  # Test fixtures
│   └── fixtures/          # Sample JSONL files for testing
└── docs/                  # Documentation
    └── PROJECT_PLAN.md    # Implementation plan and status
```

## Key Technical Details

### Claude Code Storage Format

Claude Code stores conversation history in `~/.claude/projects/` using dash-encoded paths:

```
/Users/name/project → -Users-name-project
C:\Users\name\project → C--Users-name-project
```

### Entry Types

| Type | Description |
|------|-------------|
| `user` | User messages (string content = prompt, array content = tool results) |
| `assistant` | Claude responses with text and tool_use blocks |
| `system` | System events and hook summaries |
| `queue-operation` | Subagent spawn triggers |
| `progress` | Status updates |
| `file-history-snapshot` | File state captures |
| `summary` | Conversation summaries |

### Session Structure

```
~/.claude/projects/{encoded-path}/
├── sessions-index.json         # Metadata cache (may be incomplete)
├── {sessionId}.jsonl           # Main session
└── {sessionId}/
    └── subagents/
        └── agent-{agentId}.jsonl
```

## Development Commands

```bash
# Build
cd src && go build -o claude-history .

# Run tests
cd src && go test ./...

# Run tests with coverage
cd src && go test ./... -cover

# Cross-platform builds
cd src
GOOS=darwin GOARCH=arm64 go build -o ../bin/claude-history-darwin-arm64 .
GOOS=linux GOARCH=amd64 go build -o ../bin/claude-history-linux-amd64 .
GOOS=windows GOARCH=amd64 go build -o ../bin/claude-history-windows-amd64.exe .
```

## CLI Usage

```bash
# List all projects
claude-history list

# List sessions in a project
claude-history list /path/to/project

# Query history with filters
claude-history query /path --start 2026-01-31 --type user,assistant
claude-history query /path --session <session-id> --format json

# Show agent hierarchy
claude-history tree /path --session <session-id>

# Resolve paths
claude-history resolve /path/to/project
claude-history resolve /path --session <session-id> --agent <agent-id>
```

## Implementation Status

See [docs/PROJECT_PLAN.md](docs/PROJECT_PLAN.md) for detailed implementation status.

### Completed
- ✅ Phase 1: Foundation (encoding, JSONL parser, Cobra setup)
- ✅ Phase 2: Path Resolution (resolve command)
- ✅ Phase 3: Session & Agent Discovery (list, query, tree commands)
- ✅ Phase 4: Tool Filtering (`--tool`, `--tool-match` flags)
- ✅ Phase 5: Agent Discovery (`find-agent` command)
- ✅ Phase 6: HTML Export (`export` command)
- ✅ Phase 7: Prefix Matching (git-style session/agent ID prefixes)
- ✅ Phase 8: Export Integration (wire pkg/export to cmd/export)
- ✅ Phase 9: Data Model Alignment (fix agent spawn detection)

### Planned
- Future enhancements (see PROJECT_PLAN.md)

## Coding Standards

### Go Conventions
- Use `filepath.Join()` for all path operations (cross-platform)
- Use `os.TempDir()` for temp directories, not hardcoded `/tmp`
- Handle both `\r\n` and `\n` line endings
- All exported functions must have doc comments
- Error messages should be lowercase, no punctuation

### Testing Requirements
- All new code must have corresponding `_test.go` files
- Test edge cases: empty files, malformed JSON, missing fields
- Test cross-platform scenarios: Windows and Unix paths
- Use table-driven tests where appropriate

### Package Organization
- `pkg/` - Public API, can be imported by other projects
- `internal/` - Private implementation details
- `cmd/` - CLI command definitions only, delegate to packages

## Dependencies

**External** (minimal):
- `github.com/spf13/cobra` - CLI framework

**Standard Library**:
- `encoding/json` - JSON parsing
- `filepath` - Cross-platform paths
- `bufio` - Streaming file reads
- `regexp` - Pattern matching (for tool-match)
- `time` - Timestamp handling

## Important Notes

1. **Read PROJECT_PLAN.md first** - Contains detailed checklists for current phase
2. **Run tests before committing** - `go test ./...` must pass
3. **Cross-platform paths** - Never use string concatenation for paths
4. **Streaming for large files** - Session files can be 10MB+, use Scanner not ReadAll
5. **Session index may be incomplete** - Always scan JSONL files on disk, don't trust index alone

## Related Projects

- **claude-code-viewer** - Web UI for viewing Claude Code history (separate project)
- **Claude Code** - The CLI tool whose history we're querying

---

**Document Version**: 1.0
**Last Updated**: 2026-02-01
