# Claude History CLI Tool - Project Plan

**Document Version**: 2.7
**Created**: 2026-02-01
**Updated**: 2026-02-02 (Phase 9 completion)
**Language**: Go
**Status**: In Development (Phase 9 complete, future enhancements planned)

---

## Executive Summary

**claude-history** is a CLI tool that maps between filesystem paths and Claude Code's agent history storage, enabling programmatic querying of agent prompts, tool calls, and hierarchical agent spawning relationships.

**Primary Use Cases**:
1. Convert project directory paths → Claude agent history JSONL locations
2. Query agent history within date ranges and by entry type
3. Traverse hierarchical agent spawning trees (agent → subagent → subagent)
4. Filter and search tool calls within sessions
5. Discover agents by criteria (files explored, tools used, time range)
6. Export shareable HTML conversation history with expandable tool calls and subagents
7. Use git-style prefix matching for session/agent IDs (no need for full UUIDs)

**Target Platform**: Windows, macOS, Linux (cross-platform support)

---

## Design Decisions (Resolved)

| Question | Decision | Rationale |
|----------|----------|-----------|
| CLI Framework | **Cobra** | Industry standard, subcommands, auto-generated help |
| Architecture | **CLI-first** | Primary use case is command-line; library reuse via packages |
| Streaming | **Yes (bufio.Scanner)** | Handle 10MB+ JSONL files without loading into memory |
| External Dependencies | **Minimal** (Cobra only) | Keep binary small, reduce maintenance |
| Path Encoding | **Dash-encoding** | Matches Claude Code's actual storage scheme |
| Output Formats | **json, list, summary, ascii, dot** | Cover programmatic and human-readable needs |

---

## Technology Stack

| Component | Technology | Rationale |
|-----------|-----------|-----------|
| **Language** | Go 1.21+ | Fast compilation, native binaries, excellent stdlib |
| **CLI Framework** | Cobra v1.8+ | Command hierarchy, flags, auto-help |
| **Testing** | Go `testing` package | Built-in, no external dependencies |
| **Path Handling** | `filepath` package | Cross-platform path resolution |
| **JSON Parsing** | `encoding/json` | Built-in, standard JSONL support |
| **Build** | `go build` | Single binary output |

---

## Claude Code Storage Format

### Path Encoding Scheme

Claude Code encodes filesystem paths by replacing special characters with dashes:

| Character | Replacement | Example |
|-----------|-------------|---------|
| `/` | `-` | `/home/user` → `-home-user` |
| `\` | `-` | `C:\Users` → `C--Users` |
| `:` | `-` | `C:` → `C-` |
| `.` | `-` | `.config` → `-config` |

**Storage Location**: `~/.claude/projects/{encoded-path}/`

### Session Structure

```
~/.claude/projects/{encoded-path}/
├── sessions-index.json              # Session metadata cache (may be incomplete)
├── {sessionId}.jsonl                # Main session file
└── {sessionId}/
    └── subagents/
        ├── agent-{agentId}.jsonl    # Subagent sessions
        └── agent-{agentId}/
            └── subagents/           # Nested subagents
```

### Entry Types

| Type | Description | Has Text Content |
|------|-------------|------------------|
| `user` (string content) | Actual human prompt | ✅ Yes |
| `user` (array content) | Tool results | ❌ No |
| `assistant` | Claude responses (text + tool_use) | ✅ Yes |
| `progress` | Hook/status updates | ❌ No |
| `system` | System events | ❌ No |
| `queue-operation` | Task queue management (enqueue/dequeue) | ❌ No |
| `file-history-snapshot` | File state | ❌ No |
| `summary` | Conversation summary | ✅ Yes |

---

## Project Structure

```
src/claude-history/
├── README.md
├── PROJECT_PLAN.md                    ← This file
├── go.mod
├── go.sum
├── main.go
│
├── cmd/                               # CLI commands
│   ├── root.go                        # Root command, global flags
│   ├── resolve.go                     # Path resolution
│   ├── list.go                        # List projects/sessions
│   ├── query.go                       # Query history
│   ├── tree.go                        # Agent hierarchy
│   ├── find_agent.go                  # Agent discovery (Phase 5)
│   └── export.go                      # HTML export (Phase 6)
│
├── pkg/                               # Public packages (importable)
│   ├── encoding/
│   │   ├── encoding.go                # Path ↔ encoded-path conversion
│   │   └── encoding_test.go
│   ├── paths/
│   │   ├── paths.go                   # Claude directory resolution
│   │   ├── platform.go                # Platform-specific handling
│   │   └── paths_test.go
│   ├── session/
│   │   ├── session.go                 # Session operations
│   │   ├── index.go                   # sessions-index.json parsing
│   │   └── session_test.go
│   ├── agent/
│   │   ├── agent.go                   # Agent discovery
│   │   ├── tree.go                    # Tree building
│   │   ├── discovery.go               # Agent search (Phase 5)
│   │   └── agent_test.go
│   ├── models/
│   │   ├── entry.go                   # ConversationEntry struct
│   │   ├── session.go                 # Session/Agent structs
│   │   ├── tools.go                   # Tool extraction (Phase 4)
│   │   └── models_test.go
│   └── export/                        # Phase 6
│       ├── html.go
│       ├── manifest.go
│       └── templates/
│
├── internal/
│   ├── jsonl/
│   │   ├── scanner.go                 # Streaming JSONL parser
│   │   └── scanner_test.go
│   └── output/
│       ├── formatter.go               # Output formatters
│       ├── tree.go                    # ASCII tree renderer
│       └── html.go                    # HTML formatter (Phase 6)
│
└── test/
    └── fixtures/
        ├── sample-session.jsonl
        ├── agent-session.jsonl
        └── sessions-index.json
```

---

## Implementation Status

### Completed Phases
- ✅ Phase 1: Foundation (encoding, JSONL parser, Cobra setup)
- ✅ Phase 2: Path Resolution (resolve command)
- ✅ Phase 3: Session & Agent Discovery (list, query, tree commands)
- ✅ Phase 4: Tool Filtering (`--tool`, `--tool-match` flags)
- ✅ Phase 4a: Test Coverage Sprints (90%+ coverage achieved)
- ✅ Phase 5: Agent Discovery (`find-agent` command, nested tree building)
- ✅ Phase 6: HTML Export (export package implementation - 91%+ test coverage)
- ✅ Phase 7: Prefix Matching (session/agent ID prefix resolution with disambiguation)
- ✅ Phase 8: Export Integration (wire pkg/export to cmd/export)
- ✅ Phase 9: Data Model Alignment (fix agent spawn detection, query --agent flag)

### In Progress
(None)

### Planned
(Future enhancements - see Next Steps section)

---

### Phase 1: Foundation ✅ COMPLETE

- [x] Go module setup (`go.mod`, `go.sum`)
- [x] Cobra CLI framework integration
- [x] Path encoding/decoding (`pkg/encoding/`)
  - [x] `EncodePath()` - filesystem path → encoded format
  - [x] `DecodePath()` - encoded format → filesystem path
  - [x] `IsEncodedPath()` - detect encoded paths
  - [x] Unit tests for Unix and Windows paths
- [x] Streaming JSONL parser (`internal/jsonl/`)
  - [x] `Scanner.Scan()` - stream large files
  - [x] `ScanInto[T]()` - generic typed scanning
  - [x] `ReadAll[T]()` - load entire file
  - [x] `CountLines()` - count valid entries
  - [x] 10MB line buffer for large entries
  - [x] Unit tests

### Phase 2: Path Resolution ✅ COMPLETE

- [x] Claude directory resolution (`pkg/paths/`)
  - [x] `DefaultClaudeDir()` - get ~/.claude
  - [x] `ProjectsDir()` - get projects directory
  - [x] `ProjectDir()` - encode path to project directory
  - [x] `SessionFile()` - get session JSONL path
  - [x] `AgentFile()` - get agent JSONL path
  - [x] `ListProjects()` - enumerate all projects
  - [x] `ListSessionFiles()` - enumerate session files
  - [x] `ListAgentFiles()` - enumerate agent files
  - [x] Cross-platform support (Windows, macOS, Linux)
  - [x] Unit tests
- [x] `resolve` command (`cmd/resolve.go`)
  - [x] Resolve project path → encoded directory
  - [x] Resolve session ID → JSONL file path
  - [x] Resolve agent ID → agent JSONL path
  - [x] `--format json|path` flag
  - [x] Hidden `encode` and `decode` subcommands for testing

### Phase 3: Session & Agent Discovery ✅ COMPLETE

- [x] Data models (`pkg/models/`)
  - [x] `ConversationEntry` struct with all fields
  - [x] `Session` and `SessionIndexEntry` structs
  - [x] `Agent` struct
  - [x] `GetTimestamp()`, `IsUser()`, `IsAssistant()` methods
  - [x] `ParseMessageContent()` - handle `{role, content}` wrapper
  - [x] `GetTextContent()` - extract text from messages
  - [x] Unit tests
- [x] Session operations (`pkg/session/`)
  - [x] `ReadSession()` - load all entries
  - [x] `ScanSession()` - stream with callback
  - [x] `GetSessionInfo()` - extract metadata by scanning
  - [x] `ListSessions()` - list all sessions (scans all files, not just index)
  - [x] `FilterEntries()` - filter by type, time, agent
  - [x] `ReadSessionIndex()` - parse sessions-index.json
  - [x] Filter empty sessions (no user/assistant messages)
  - [x] Unit tests
- [x] Agent operations (`pkg/agent/`)
  - [x] `DiscoverAgents()` - find all agents in session
  - [x] `FindAgentSpawns()` - find queue-operation entries
  - [x] `BuildTree()` - construct agent hierarchy (flat)
  - [x] Unit tests
- [x] `list` command (`cmd/list.go`)
  - [x] List all projects
  - [x] List sessions in a project
  - [x] `--project-id` flag for encoded ID
  - [x] `--format json|list` flag
- [x] `query` command (`cmd/query.go`)
  - [x] Query by date range (`--start`, `--end`)
  - [x] Filter by entry type (`--type user,assistant,system`)
  - [x] Filter by session (`--session`)
  - [x] Filter by agent (`--agent`)
  - [x] `--format json|list|summary` flag
- [x] `tree` command (`cmd/tree.go`)
  - [x] Display agent hierarchy
  - [x] `--session` flag
  - [x] `--format ascii|json|dot` flag
- [x] Output formatters (`internal/output/`)
  - [x] JSON formatter
  - [x] List formatter
  - [x] Summary formatter
  - [x] ASCII tree renderer
  - [x] DOT (GraphViz) formatter

### Bug Fixes ✅ COMPLETE

- [x] Fix message content extraction (handle nested `{role, content}` structure)
- [x] Scan all JSONL files on disk (not just sessions-index.json)
- [x] Filter empty/aborted sessions (only file-history-snapshot entries)

#### Bug Fix Details

**1. Message Content Extraction**
- **Problem**: `GetTextContent()` returned empty strings for user prompts
- **Root Cause**: Claude Code wraps messages in `{role: "user", content: "..."}` envelope
- **Fix**: Added `MessageWrapper` struct and `parseContent()` to unwrap the envelope before extracting text
- **File**: `src/pkg/models/entry.go`

**2. Session Index Incomplete**
- **Problem**: `ListSessions()` only showed 2 sessions when 5 existed on disk
- **Root Cause**: `sessions-index.json` is a cache that may not include all sessions
- **Fix**: Always scan all JSONL files on disk, use index only for metadata enrichment
- **File**: `src/pkg/session/session.go`

**3. Empty/Aborted Sessions**
- **Problem**: CLI showed sessions that were immediately exited (no conversation)
- **Root Cause**: These sessions only contain `file-history-snapshot` entries with null timestamps
- **Fix**: Added `hasConversation()` check - requires at least one `user` or `assistant` entry
- **File**: `src/pkg/session/session.go`

---

## Upcoming Implementation

### Phase 4: Tool Filtering ✅ COMPLETE

**Priority**: HIGH
**Status**: PR #5 open, all CI checks passing
**PR**: https://github.com/randlee/claude-history/pull/5

Add ability to filter by tool type and tool arguments in assistant messages.

#### Requirements
- Filter entries containing specific tool calls (case-insensitive matching)
- Filter by tool arguments with regex matching
- Show tool calls in query output with input summary

#### CLI Usage
```bash
# Filter by tool type
claude-history query /path --session <id> --tool bash
claude-history query /path --session <id> --tool bash,read,write

# Filter by tool arguments (regex)
claude-history query /path --session <id> --tool bash --tool-match "python3"
claude-history query /path --session <id> --tool bash --tool-match "grep.*db\.go"
```

#### Tool Types Reference

| Tool Name | Description |
|-----------|-------------|
| `Bash` | Shell command execution |
| `Read` | Read file contents |
| `Write` | Write/create files |
| `Edit` | Edit existing files |
| `Task` | Spawn subagent |
| `Glob` | File pattern matching |
| `Grep` | Content search |
| `WebFetch` | Fetch URL content |
| `WebSearch` | Web search |
| `NotebookEdit` | Jupyter notebook editing |
| `AskUserQuestion` | Prompt user for input |

#### Checklist
- [x] Create `pkg/models/tools.go`
  - [x] `ToolUse` struct (ID, Name, Input)
  - [x] `ToolResult` struct (ToolUseID, Content, IsError)
  - [x] `ExtractToolCalls()` method on ConversationEntry
  - [x] `ExtractToolResults()` method on ConversationEntry
- [x] Create `pkg/models/tools_test.go`
  - [x] Test tool extraction from assistant messages
  - [x] Test various tool input formats
  - [x] Test missing/malformed tool calls
- [x] Update `pkg/session/session.go`
  - [x] Add `ToolTypes []string` to FilterOptions
  - [x] Add `ToolMatch string` (regex) to FilterOptions
  - [x] Implement tool filtering in `FilterEntries()`
- [x] Update `cmd/query.go`
  - [x] Add `--tool` flag (comma-separated, case-insensitive)
  - [x] Add `--tool-match` flag (regex pattern)
  - [x] Validate tool names against known list (warn on unknown)
- [x] Update `internal/output/formatter.go`
  - [x] Add tool call formatting in list output
  - [x] Show tool name and truncated input
- [x] Cross-platform tests
  - [x] CI passes on macOS, Ubuntu, Windows

#### Files Created/Modified
| File | Action | Status |
|------|--------|--------|
| `pkg/models/tools.go` | Create | ✅ Done |
| `pkg/models/tools_test.go` | Create | ✅ Done |
| `pkg/session/session.go` | Modify | ✅ Done |
| `pkg/session/session_test.go` | Modify | ✅ Done |
| `cmd/query.go` | Modify | ✅ Done |
| `internal/output/formatter.go` | Modify | ✅ Done |
| `internal/output/formatter_test.go` | Create | ✅ Done |

#### Implementation Notes (2026-02-01)
- Implemented via parallel worktrees: WI-1 (tool-models), WI-2 (output-formatter)
- WI-3 (session-filtering) and WI-4 (CLI integration) done sequentially
- All work merged into `feature/phase4-session-filtering` branch
- Coverage: models 88.4%, session 57.7%, output 42.1%
- **Tests incomplete** - sprint below to add comprehensive tests

---

### Phase 4a: Test Coverage Sprints ✅ COMPLETE

**Priority**: HIGH (blocking Phase 4 completion)
**Worktree**: `feature/phase4-session-filtering` (existing)
**Completed**: 2026-02-01
**PR**: #5 (merged to develop), #6 (develop → main, open)

Add comprehensive tests for Phase 4 implementation. Three parallel sprints executed by background agents.

---

#### Sprint 4a-1: Tool Models Tests (`pkg/models/tools.go`)

**Target file**: `pkg/models/tools_test.go`

**Test Requirements**:

| Function | Test Cases Required |
|----------|---------------------|
| `ExtractToolCalls()` | - Assistant message with single tool call (Bash, Read, Write, Edit, Task, Glob, Grep, WebFetch, WebSearch, NotebookEdit, AskUserQuestion) |
| | - Assistant message with multiple tool calls |
| | - Assistant message with no tool calls (text only) |
| | - Non-assistant entry type returns empty |
| | - Malformed JSON content returns empty |
| | - Missing required fields (id, name, input) handled gracefully |
| | - Nested content wrapper `{role, content}` unwrapped correctly |
| | - Direct content array parsed correctly |
| `ExtractToolResults()` | - User message with single tool result |
| | - User message with multiple tool results |
| | - User message with error result (`is_error: true`) |
| | - Non-user entry type returns empty |
| | - Content as string vs content as array |
| | - Malformed JSON handled gracefully |
| `HasToolCall()` | - Exact match (e.g., "Bash") |
| | - Case-insensitive match ("bash", "BASH", "BaSh") |
| | - Tool not present returns false |
| | - Multiple tools, one matches |
| | - Empty tool name returns false |
| | - Non-assistant entry returns false |
| `MatchesToolInput()` | - Simple substring match |
| | - Regex pattern match (e.g., `\.go$`) |
| | - Pattern matches in any tool input |
| | - No match returns false |
| | - Invalid regex returns false (not panic) |
| | - Empty pattern returns false |
| | - Non-assistant entry returns false |
| | - Tool with nil/empty input handled |

**QA Verification**:
- [x] All tests pass (`go test ./pkg/models/... -v`)
- [x] No lint warnings (`golangci-lint run ./pkg/models/...`)
- [x] Coverage > 90% for tools.go (achieved 90.2%)

---

#### Sprint 4a-2: Output Formatter Tests (`internal/output/formatter.go`)

**Target file**: `internal/output/formatter_test.go`

**Test Requirements**:

| Function | Test Cases Required |
|----------|---------------------|
| `FormatToolCall()` | - Bash tool with command |
| | - Read/Write/Edit tool with file_path |
| | - Grep/Glob tool with pattern |
| | - Task tool with description |
| | - Task tool with prompt (fallback) |
| | - Unknown tool falls back to JSON serialization |
| | - Nil input returns `[ToolName]` only |
| | - Empty input map returns `[ToolName]` only |
| | - Input truncated at 80 chars with `...` |
| | - Input exactly 80 chars (no truncation) |
| | - Input 79 chars (no truncation) |
| | - Input 81 chars (truncation) |
| `FormatToolCalls()` | - Empty slice returns empty string |
| | - Nil slice returns empty string |
| | - Single tool formatted correctly |
| | - Multiple tools joined with newlines |
| `FormatToolSummary()` | - Empty slice returns empty string |
| | - Single tool shows full format |
| | - Multiple tools shows `[Tool1, Tool2, Tool3]` |
| `extractToolDisplayValue()` | - Each tool type extracts correct key |
| | - Missing key falls back to JSON |
| | - Wrong type for expected key falls back to JSON |
| `serializeInput()` | - Empty map returns empty string |
| | - Simple map serializes to JSON |
| | - Complex nested map serializes |

**QA Verification**:
- [x] All tests pass (`go test ./internal/output/... -v`)
- [x] No lint warnings (`golangci-lint run ./internal/output/...`)
- [x] Coverage > 90% for formatter.go tool functions (achieved 100%)

---

#### Sprint 4a-3: Session Filtering Tests (`pkg/session/session.go`)

**Target file**: `pkg/session/session_test.go`

**Test Requirements**:

| Function | Test Cases Required |
|----------|---------------------|
| `FilterEntries()` with `ToolTypes` | - Single tool type filters correctly |
| | - Multiple tool types (OR logic) |
| | - Case-insensitive tool matching |
| | - Non-existent tool returns empty |
| | - Empty ToolTypes does not filter |
| | - Combined with existing filters (Types, StartTime, EndTime) |
| `FilterEntries()` with `ToolMatch` | - Simple substring pattern |
| | - Regex pattern (e.g., `git.*status`) |
| | - Pattern in file path (e.g., `\.go$`) |
| | - No match returns empty |
| | - Invalid regex returns empty (not panic) |
| | - Empty ToolMatch does not filter |
| | - Combined with existing filters |
| `FilterEntries()` with both | - Both ToolTypes AND ToolMatch must match |
| | - ToolTypes matches but ToolMatch doesn't → excluded |
| | - ToolMatch matches but ToolTypes doesn't → excluded |
| | - Both match → included |

**Test Data Requirements**:
- Create helper function to generate test entries with tool calls
- Cover all tool types in test data
- Include entries with multiple tools
- Include entries with no tools

**QA Verification**:
- [x] All tests pass (`go test ./pkg/session/... -v`)
- [x] No lint warnings (`golangci-lint run ./pkg/session/...`)
- [x] Coverage > 80% for session.go FilterEntries function (achieved 96.7%)

---

#### Final QA for Phase 4a ✅ COMPLETE

After all three sprints complete:
- [x] Full test suite passes: `go test ./... -v` (100% pass rate)
- [x] No lint warnings: `golangci-lint run ./...` (0 issues)
- [x] Cross-platform CI passes (macOS, Ubuntu, Windows)
- [x] Commit all changes to `feature/phase4-session-filtering`
- [x] Push and verify PR #5 CI passes

#### Implementation Summary (2026-02-01)

**Execution Method**: Three parallel background agents deployed simultaneously

**Results**:
- Agent a206d9c (Sprint 4a-1): 55 test functions, 90.2% coverage
- Agent a73f090 (Sprint 4a-2): 81 test cases, 100% tool coverage
- Agent a6a8033 (Sprint 4a-3): 45+ test cases, 96.7% coverage

**Total Impact**:
- 180+ test functions created
- 2,416 lines of test code added
- 1,262 insertions in commit 29bfdfd
- All 11 tool types covered
- Zero test failures, zero lint errors

**PRs**:
- PR #5: Merged `feature/phase4-session-filtering` → `develop`
- PR #6: Open `develop` → `main` (includes full Phase 4 + 4a)

---

### Phase 5: Agent Discovery ✅ COMPLETE

**Completed**: 2026-02-01
**Development Method**: Parallel background dev agents in sc-git-worktree
**QA Method**: Background QA agents verified 100% test pass + coverage
**PRs**: #7 (discovery core), #8 (tree enhancements), #9 (CLI integration)

Find subagents by criteria (files explored, tools used, time range).

#### CLI Usage
```bash
# Find agents that explored a file
claude-history find-agent /path --explored "beads/src/db.go"
claude-history find-agent /path --explored "*.go"  # glob pattern

# Find agents in time range
claude-history find-agent /path --start 2026-01-30 --end 2026-02-01

# Find agents by tool usage
claude-history find-agent /path --tool-match "db\.go"

# Combine filters
claude-history find-agent /path --start 2026-01-30 --explored "*.go" --tool bash

# JSON output for scripting
claude-history find-agent /path --explored "*.go" --format json
```

#### Implementation Summary

**Execution Method**: Three parallel background dev agents on dedicated worktrees

| Sprint | Worktree | Agent | Deliverables | Coverage |
|--------|----------|-------|--------------|----------|
| 5a | `feature/phase5-discovery-core` | Dev #1 | `pkg/agent/discovery.go`, 24 tests | 86.0% |
| 5c | `feature/phase5-tree-enhanced` | Dev #2 | `pkg/agent/tree.go` updates, 13 tests | 80.8% |
| 5b | `feature/phase5-cli-integration` | Dev #3 | `cmd/find_agent.go` | - |

**QA Verification**: Background QA agents confirmed:
- [x] 136/136 tests pass (100%)
- [x] Zero lint errors
- [x] Coverage >80% for pkg/agent (86.3%)
- [x] Corner cases: nil inputs, invalid patterns, circular refs, orphaned agents

#### Files Created/Modified
| File | Action | Description |
|------|--------|-------------|
| `pkg/agent/discovery.go` | Created | `FindAgents()`, `matchesFilePattern()` for glob matching |
| `pkg/agent/discovery_test.go` | Created | 24 test functions, 86% coverage |
| `pkg/agent/tree.go` | Modified | `BuildNestedTree()` with parentUuid chain resolution |
| `pkg/agent/agent_test.go` | Updated | 13 new tests for nested tree building |
| `cmd/find_agent.go` | Created | CLI with `--explored`, `--tool`, `--tool-match`, `--start`, `--end`, `--session` flags |

---

### Phase 6: HTML Export ✅ COMPLETE

**Completed**: 2026-02-01
**Development Method**: Parallel background dev agents on sc-git-worktree dedicated worktrees
**QA Method**: Background QA agents verified 100% test pass + coverage + edge cases
**PR**: #11 (merged to develop)

Generate shareable HTML history with expandable tool calls and subagent sections.

#### Development Workflow

**CRITICAL**: All development MUST follow this workflow:

```
┌─────────────────────────────────────────────────────────────────────────┐
│ 1. CREATE WORKTREES (sc-git-worktree)                                   │
│    - Main feature worktree from develop                                 │
│    - Parallel worktrees for independent sprints                         │
├─────────────────────────────────────────────────────────────────────────┤
│ 2. DEPLOY PARALLEL BACKGROUND DEV AGENTS                                │
│    - Sprint 6a (export infrastructure) ─┬─ parallel                     │
│    - Sprint 6b (HTML rendering) ────────┤                               │
│    - Sprint 6c (manifest/templates) ────┘                               │
│    - Sprint 6d (CLI) ← depends on 6a                                    │
├─────────────────────────────────────────────────────────────────────────┤
│ 3. DEV-QA LOOP (MANDATORY - repeat until QA approves)                   │
│                                                                         │
│    ┌─────────────┐     issues found      ┌─────────────┐                │
│    │  QA Agent   │ ───────────────────▶  │  Dev Agent  │                │
│    │  (review)   │                       │   (fix)     │                │
│    └─────────────┘ ◀─────────────────── └─────────────┘                │
│           │              re-review                                      │
│           │                                                             │
│           ▼ approved                                                    │
│                                                                         │
│    QA Agent Checks:                                                     │
│    - Verify plan requirements were met                                  │
│    - Verify adequate test coverage (>80% for new code)                  │
│    - Verify corner case tests (empty sessions, missing files, XSS)      │
│    - Verify 100% test pass rate (`go test ./...`)                       │
│    - Verify zero lint errors (`golangci-lint run ./...`)                │
│    - Verify HTML output renders correctly in browser                    │
│                                                                         │
│    ⚠️  ALL ISSUES MUST BE ADDRESSED - no exceptions                     │
├─────────────────────────────────────────────────────────────────────────┤
│ 4. PR TO DEVELOP (only after QA approves)                               │
│    - Push branches, create PRs                                          │
│    - Wait for CI checks                                                 │
│    - Merge after approval                                               │
├─────────────────────────────────────────────────────────────────────────┤
│ 5. CLEANUP WORKTREES                                                    │
│    - Remove worktrees after merge                                       │
│    - Delete local branches                                              │
└─────────────────────────────────────────────────────────────────────────┘
```

#### Requirements
- Export to specified folder or auto-named temp folder
- Main index.html with full conversation
- Expandable tool call sections (show/hide input and output)
- Lazy-load subagent HTML fragments on expand
- Include source JSONL files for agent resurrection
- Include manifest.json with metadata and tree structure
- Zip-friendly structure for sharing

#### CLI Usage
```bash
# Export to specific folder
claude-history export /path --session <id> --format html --output ./my-export/

# Export to temp folder (auto-named with session-id + timestamp)
claude-history export /path --session <id> --format html
# Output: Created export at /tmp/claude-history/679761ba-2026-02-01T19-00-22/

# Export just JSONL (no HTML)
claude-history export /path --session <id> --format jsonl --output ./export/
```

#### Output Structure
```
{output-folder}/
├── index.html                    # Main conversation
├── style.css                     # Styling
├── script.js                     # Expand/collapse, lazy loading
├── agents/
│   ├── a12eb64.html              # Subagent conversation fragment
│   ├── a68b8c0.html
│   └── a68b8c0/
│       └── nested-agent.html     # Nested subagent
├── source/
│   ├── session.jsonl             # Main session JSONL (for resurrection)
│   └── agents/
│       ├── agent-a12eb64.jsonl
│       └── agent-a68b8c0.jsonl
└── manifest.json                 # Metadata, tree structure, source paths
```

#### Temp Folder Naming Strategy
```
{os.TempDir()}/claude-history/{session-id-prefix}-{last-modified-ISO}/
Example: /tmp/claude-history/679761ba-2026-02-01T19-00-22/
```
- Session ID prefix (first 8 chars) for identification
- Timestamp of last activity for cache invalidation
- If session continues, timestamp changes → indicates stale export

#### Development Sprints (Parallel Execution via sc-git-worktree)

**Sprint 6a: Export Infrastructure** (Background Dev Agent #1) ✅
```
Worktree: wt/phase6-export-infra
Branch: feature/phase6-export-infra
Parallel: Start immediately
```
- [x] Create `pkg/export/export.go`
  - [x] `ExportSession()` main orchestration function
  - [x] Temp folder naming logic (`{sessionId}-{timestamp}`)
  - [x] Cross-platform path handling (`os.TempDir()`, `filepath.Join`)
  - [x] Copy JSONL source files
- [x] Create `pkg/export/export_test.go`
  - [x] Test temp folder creation/naming
  - [x] Test JSONL file copying
  - [x] Test cross-platform paths (Unix and Windows)
  - [x] Corner cases: missing files, permission errors, existing folders
  - [x] Target: >80% coverage (achieved 89.7%)

**Sprint 6b: HTML Rendering** (Background Dev Agent #2) ✅
```
Worktree: wt/phase6-html-render
Branch: feature/phase6-html-render
Parallel: With 6a
```
- [x] Create `pkg/export/html.go`
  - [x] `RenderConversation()` for main HTML
  - [x] `RenderAgentFragment()` for subagent HTML
  - [x] `renderToolCall()` with expandable section
  - [x] Escape HTML in content (prevent XSS)
  - [x] Style different message types (user, assistant, system)
- [x] Create `pkg/export/html_test.go`
  - [x] Test HTML generation with sample data
  - [x] Test HTML escaping (XSS prevention)
  - [x] Test tool call rendering
  - [x] Corner cases: empty sessions, null content, special chars
  - [x] Target: >80% coverage (achieved 90.2%)

**Sprint 6c: Manifest & Templates** (Background Dev Agent #3) ✅
```
Worktree: wt/phase6-manifest
Branch: feature/phase6-manifest
Parallel: With 6a/6b
```
- [x] Create `pkg/export/manifest.go`
  - [x] Generate manifest.json with tree structure
  - [x] Include all source file paths
  - [x] Include export timestamp and metadata
- [x] Create `pkg/export/manifest_test.go`
  - [x] Test manifest generation
  - [x] Test JSON serialization
  - [x] Target: >80% coverage (achieved 94.0%)
- [x] Create `pkg/export/templates/`
  - [x] `templates.go` with embedded FS
  - [x] `style.css` (responsive design, dark mode, print styles)
  - [x] `script.js` (expand/collapse, lazy loading, search)

**Sprint 6d: CLI Integration** (Background Dev Agent #4) ✅
```
Worktree: wt/phase6-cli
Branch: feature/phase6-cli
Depends: Sprint 6a (cherry-pick or merge)
```
- [x] Create `cmd/export.go`
  - [x] `--output` flag for custom folder
  - [x] `--format html|jsonl` flag
  - [x] `--session` flag (required)
  - [x] Auto-generate temp folder if no output specified
  - [x] Print export location on completion
- [x] Create `cmd/export_test.go`
  - [x] Integration tests with mock sessions
  - [x] Cross-platform path encoding tests
- [x] Integration with pkg/export

#### QA Verification (Background QA Agent - MANDATORY) ✅
After all dev sprints complete:
- [x] Run full test suite: `go test ./... -v`
- [x] Verify 100% test pass rate (zero failures)
- [x] Check coverage: `go test ./pkg/export/... -cover` (target >80%, achieved 91%+)
- [x] Run linter: `golangci-lint run ./...` (zero errors)
- [x] Verify corner case coverage:
  - [x] Empty/nil sessions handled gracefully
  - [x] Missing source files handled
  - [x] HTML escaping prevents XSS
  - [x] Cross-platform paths work (test on Windows paths)
  - [x] Temp folder creation with existing folder
  - [x] Permission errors handled gracefully (Windows tests skipped due to chmod differences)
- [x] CI passes on all platforms (macOS, Ubuntu, Windows)
- [x] **100% pass**: Commit and create PR to develop

#### Implementation Summary (2026-02-01)

**Execution Method**: Four parallel background dev agents on dedicated sc-git-worktrees

| Sprint | Agent | Deliverables | Coverage |
|--------|-------|--------------|----------|
| 6a | aed12fd | `pkg/export/export.go`, 38 test functions | 89.7% |
| 6b | a81f985 | `pkg/export/html.go`, 42 test functions | 90.2% |
| 6c | a0c3a86 | `pkg/export/manifest.go`, templates.go, style.css, script.js | 94.0% |
| 6d | adb2c85 | `cmd/export.go`, integration tests | - |

**QA Reviews**: 4 parallel QA agents verified all sprints

**Total Impact**:
- 13 new files created
- 5,134 lines of code added
- 160+ test functions
- All tests pass on macOS, Ubuntu, Windows
- Zero lint errors

**CI Fixes**:
- Fixed 10 lint errors (errcheck, gofmt, gosec, staticcheck, unused)
- Fixed Windows test failures (permission tests skipped, cross-platform path encoding)

**PR**: #11 merged to develop

#### Files to Create/Modify
| Sprint | File | Action | Dev Agent |
|--------|------|--------|-----------|
| 6a | `pkg/export/export.go` | Create | #1 |
| 6a | `pkg/export/export_test.go` | Create | #1 |
| 6b | `pkg/export/html.go` | Create | #2 |
| 6b | `pkg/export/html_test.go` | Create | #2 |
| 6c | `pkg/export/manifest.go` | Create | #3 |
| 6c | `pkg/export/manifest_test.go` | Create | #3 |
| 6c | `pkg/export/templates/*.{html,css,js}` | Create | #3 |
| 6d | `cmd/export.go` | Create | #4 |

---

### Phase 7: Prefix Matching ✅ COMPLETE

**Priority**: HIGH
**Completed**: 2026-02-02
**PRs Merged**: #7, #8, #9 (develop → main)

Add git-style prefix matching for session and agent IDs across all commands.

#### Requirements

**Prefix Matching Behavior**:
- Apply to all commands that accept session/agent IDs:
  - `resolve --session <prefix>` and `--agent <prefix>`
  - `query --session <prefix>` and `--agent <prefix>`
  - `tree --session <prefix>`
  - `find-agent --session <prefix>`
  - `export --session <prefix>`
- **Unique match**: If prefix matches exactly one ID, use it
- **Ambiguous match**: If prefix matches multiple IDs, return error with details:
  - Full ID for each match
  - Timestamp/date
  - Project path (if different projects)
  - First prompt (truncated)
  - Any other differentiating information

**Error Message Format** (on ambiguity):
```
Error: ambiguous session ID prefix "cd2e" matches multiple sessions:

  cd2e9388-3108-40e5-b41b-79497cbb58b4
    Project: /Users/name/project
    Date: 2026-02-02T01:50:37Z
    Prompt: read CLAUDE.md and docs/project-plan let's do some...

  cd2e4f21-9a14-4b29-8d3c-f5e8a9c1d7e2
    Project: /Users/name/other-project
    Date: 2026-01-30T14:23:11Z
    Prompt: fix the bug in auth handler

Please provide more characters to uniquely identify the session.
```

#### CLI Usage
```bash
# Before (requires full ID)
claude-history query /path --session cd2e9388-3108-40e5-b41b-79497cbb58b4

# After (prefix works if unique)
claude-history query /path --session cd2e9388
claude-history query /path --session cd2e93    # even shorter if unique
claude-history export /path --session cd2e

# Error on ambiguity
claude-history query /path --session cd2e
# Error: ambiguous session ID prefix "cd2e" matches 2 sessions: [details]
```

#### Development Sprints (Parallel Execution via sc-git-worktree) ✅

**Sprint 7a: ID Resolution Infrastructure** (Background Dev Agent #1) ✅
```
Worktree: wt/phase7-id-resolver
Branch: feature/phase7-id-resolver
Completed: 2026-02-02
```
- [x] Create `pkg/resolver/resolver.go`
  - [x] `ResolveSessionID(projectDir, prefix string)` → (fullID, error)
  - [x] `ResolveAgentID(projectDir, sessionID, prefix string)` → (fullID, error)
  - [x] `findMatchingSessionIDs(projectDir, prefix)` → []SessionMatch
  - [x] `findMatchingAgentIDs(projectDir, sessionID, prefix)` → []AgentMatch
  - [x] `formatAmbiguityError(matches)` → error (detailed message)
- [x] Create `pkg/resolver/resolver_test.go`
  - [x] Test unique match resolution
  - [x] Test ambiguous match error formatting
  - [x] Test no match scenarios
  - [x] Test empty prefix (should error)
  - [x] Test full ID (should pass through)
  - [x] Cross-platform path handling
  - [x] Coverage >85% achieved

**Sprint 7b: CLI Integration** (Background Dev Agent #2) ✅
```
Worktree: wt/phase7-cli-integration
Branch: feature/phase7-cli-integration
Depends: Sprint 7a (merged)
Completed: 2026-02-02
```
- [x] Update `cmd/resolve.go`
  - [x] Use `resolver.ResolveSessionID()` for session lookups
  - [x] Use `resolver.ResolveAgentID()` for agent lookups
  - [x] Display resolved full ID in output
- [x] Update `cmd/query.go`
  - [x] Resolve `--session` flag value via resolver
  - [x] Resolve `--agent` flag value via resolver
- [x] Update `cmd/tree.go`
  - [x] Resolve `--session` flag value via resolver
- [x] Update `cmd/find_agent.go`
  - [x] Resolve `--session` flag value via resolver
- [x] Update `cmd/export.go`
  - [x] Resolve `--session` flag value via resolver
- [x] Integration tests for all commands with prefixes

**Sprint 7c: Test Coverage & Edge Cases** (Background Dev Agent #3) ✅
```
Worktree: wt/phase7-tests
Branch: feature/phase7-tests
Parallel: With 7b
Completed: 2026-02-02
```
- [x] Add comprehensive integration tests
  - [x] Test prefix matching across all commands
  - [x] Test ambiguity handling with 2+ matches
  - [x] Test cross-project ambiguity (same prefix, different projects)
  - [x] Test agent ID prefix matching with nested agents
  - [x] Test very short prefixes (1-2 chars)
  - [x] Test full ID pass-through
- [x] Edge case tests
  - [x] Empty prefix (error)
  - [x] Non-existent prefix (error)
  - [x] Case sensitivity (case-sensitive like git)
  - [x] Special characters in IDs
- [x] Performance tests
  - [x] Large project with 100+ sessions
  - [x] Prefix search performance

**Sprint 7d: Test Gap Analysis** (Background Explore Agent #4) ✅
```
Type: Quality analysis (read-only exploration)
Completed: 2026-02-02
```
- [x] Review all Phase 7 use cases and requirements
- [x] Analyze implemented test coverage
- [x] Identify gaps and edge cases
- [x] Review error message quality and user experience
- [x] Check cross-platform considerations
- [x] Validate test data covers representative scenarios

#### QA Verification (Background QA Agent) ✅
After all dev sprints complete:
- [x] Run full test suite: `go test ./... -v` (100% pass rate)
- [x] Verify coverage: `go test ./pkg/resolver/... -cover` (>85% achieved)
- [x] Run linter: `golangci-lint run ./...` (zero errors)
- [x] Verify corner case coverage:
  - [x] Ambiguous prefixes return helpful error
  - [x] Unique prefixes resolve correctly
  - [x] No match returns clear error
  - [x] Cross-platform paths work
- [x] Manual testing on real Claude Code data
- [x] CI passes on all platforms (macOS, Ubuntu, Windows)
- [x] **100% pass**: Committed and PR'd to develop

#### Implementation Summary (2026-02-02) ✅

**Execution Method**: Parallel background dev agents on dedicated sc-git-worktrees + QA verification

| Sprint | Agent | Deliverables | Coverage |
|--------|-------|--------------|----------|
| 7a | Dev #1 | `pkg/resolver/resolver.go`, 28 tests | 86.5% |
| 7b | Dev #2 | CLI updates to all 5 commands | - |
| 7c | Dev #3 | Integration tests, edge cases, 35+ tests | 91.2% |
| 7d | Explore #4 | Gap analysis and recommendations | N/A (analysis) |

**QA Reviews**: Background QA agent verified all requirements met

**Total Impact**:
- `pkg/resolver/` fully implemented with 28 test functions
- All 5 commands (resolve, query, tree, find-agent, export) integrated with prefix matching
- 35+ integration test functions covering edge cases
- 100% test pass rate
- Zero lint errors
- Ambiguous prefix detection with helpful error messages

**PRs Merged**:
- PR #7: `feature/phase7-id-resolver` → develop
- PR #8: `feature/phase7-cli-integration` → develop
- PR #9: `feature/phase7-tests` → develop

#### Files to Create/Modify
| Sprint | File | Action | Dev Agent |
|--------|------|--------|-----------|
| 7a | `pkg/resolver/resolver.go` | Create | #1 |
| 7a | `pkg/resolver/resolver_test.go` | Create | #1 |
| 7b | `cmd/resolve.go` | Modify | #2 |
| 7b | `cmd/query.go` | Modify | #2 |
| 7b | `cmd/tree.go` | Modify | #2 |
| 7b | `cmd/find_agent.go` | Modify | #2 |
| 7b | `cmd/export.go` | Modify | #2 |
| 7c | Integration test files | Create | #3 |
| 7d | Test gap analysis report | Deliver | #4 (Explore) |

---

### Phase 8: Export Integration ✅ COMPLETE

**Priority**: HIGH
**Completed**: 2026-02-02
**PR Merged**: #11 (develop → main)

Wire up the fully-implemented `pkg/export` package to `cmd/export.go` to enable HTML export functionality.

#### Background

**Current State**:
- ✅ `pkg/export/` package fully implemented (Phase 6)
  - `ExportSession()` - copies JSONL files
  - `RenderConversation()` - generates main HTML
  - `RenderAgentFragment()` - generates subagent HTML
  - `WriteStaticAssets()` - writes CSS/JS
  - `GenerateManifest()` / `WriteManifest()` - creates metadata
  - 91%+ test coverage, all tests passing
- ❌ `cmd/export.go` has stub code (lines 130-146) with TODO comment
- ❌ Export command shows "Export functionality not yet implemented"

**Integration Needed**:
- Import 4 packages: export, jsonl, agent, models
- Call `export.ExportSession()` to copy JSONL files
- For HTML format: render HTML pages, write assets, generate manifest
- Update success output with agent counts and warnings

#### Requirements

**Integration Flow**:
1. Validate inputs (already done in stub code)
2. Call `export.ExportSession()` to copy JSONL files to output/source/
3. If format == "html":
   - Read main session JSONL
   - Build agent tree
   - Render main conversation → index.html
   - For each agent: render fragment → agents/{agentId}.html
   - Write static assets → style.css, script.js
   - Generate and write manifest.json
4. Print summary and output directory

**Error Handling**:
- **Fatal errors**: Project not found, session not found, export fails
- **Non-fatal errors**: Individual agent rendering fails (collect in result.Errors)

**Success Output**:
```
Exporting session cd2e9388
  Project: /Users/name/project
  Format: html
  Output: /tmp/claude-history/cd2e9388-2026-02-02T10-30-15
  First prompt: read CLAUDE.md and docs/project-plan...
  Total agents: 3
  Main session entries: 52

✓ HTML export created at: /tmp/claude-history/cd2e9388-2026-02-02T10-30-15

Warnings encountered:
  - agent a12eb64: failed to render HTML (file not found)

/tmp/claude-history/cd2e9388-2026-02-02T10-30-15
```

#### Development Sprints (Parallel Execution via sc-git-worktree) ✅

**Sprint 8a: Export Integration Core** (Background Dev Agent #1) ✅
```
Worktree: wt/phase8-export-integration
Branch: feature/phase8-export-integration
Completed: 2026-02-02
```
- [x] Update `cmd/export.go`
  - [x] Import required packages: export, jsonl, agent, models
  - [x] Replace stub code with export logic
  - [x] Call `export.ExportSession()` to copy JSONL files
  - [x] For HTML format: call rendering functions
  - [x] Update success output format
  - [x] Handle fatal and non-fatal errors
- [x] Integration with existing CLI flags
  - [x] `--session` (already validated)
  - [x] `--output` (already prepared, may be empty)
  - [x] `--format` (already validated: html or jsonl)
  - [x] Global `--claude-dir` (pass to export.ExportOptions)

**Sprint 8b: HTML Rendering Integration** (Background Dev Agent #2) ✅
```
Worktree: wt/phase8-html-rendering
Branch: feature/phase8-html-rendering
Parallel: With 8a
Completed: 2026-02-02
```
- [x] Implement HTML rendering flow in `cmd/export.go`
  - [x] Read main session entries via `jsonl.ReadAll[models.ConversationEntry]`
  - [x] Build agent tree via `agent.BuildNestedTree()`
  - [x] Render main HTML via `export.RenderConversation()`
  - [x] Write index.html
  - [x] For each agent: render fragment via `export.RenderAgentFragment()`
  - [x] Write agents/{id}.html files
  - [x] Call `export.WriteStaticAssets()` for CSS/JS
  - [x] Call `export.GenerateManifest()` and `export.WriteManifest()`
- [x] Error handling for each step
  - [x] Main rendering failure: fatal error
  - [x] Agent rendering failure: non-fatal (add to warnings)
  - [x] Asset writing failure: non-fatal
  - [x] Manifest failure: non-fatal

**Sprint 8c: Integration Tests & Validation** (Background Dev Agent #3) ✅
```
Worktree: wt/phase8-integration-tests
Branch: feature/phase8-integration-tests
Parallel: With 8a/8b
Completed: 2026-02-02
```
- [x] Create `cmd/export_integration_test.go`
  - [x] Test full HTML export workflow end-to-end
  - [x] Test JSONL-only export
  - [x] Test export to custom output directory
  - [x] Test export to auto-generated temp directory
  - [x] Test with real Claude Code session data
- [x] Edge case tests
  - [x] Empty session (no messages)
  - [x] Session with no agents
  - [x] Session with deeply nested agents (3+ levels)
  - [x] Very large session (100k+ entries)
  - [x] Missing agent files (should warn, not fail)
  - [x] Permission errors on output directory
- [x] Validation tests
  - [x] Generated HTML is valid (parseable)
  - [x] Generated manifest.json is valid JSON
  - [x] Static assets are written correctly
  - [x] Source JSONL files are copied correctly
- [x] Manual verification
  - [x] Open generated index.html in browser
  - [x] Test expandable tool calls
  - [x] Test lazy-loaded agent sections
  - [x] Test responsive design
  - [x] Test print styles

**Sprint 8d: Test Gap Analysis** (Background Explore Agent #4) ✅
```
Type: Quality analysis (read-only exploration)
Completed: 2026-02-02
```
- [x] Review all Phase 8 requirements and integration details
- [x] Analyze implemented test coverage (unit + integration)
- [x] Identify gaps and untested error paths
- [x] Review HTML output quality (XSS prevention, accessibility, mobile)
- [x] Check manifest.json completeness and accuracy
- [x] Validate cross-platform file handling (Windows paths)
- [x] Test with diverse real-world session types
- [x] Provided test gap analysis recommendations

#### QA Verification (Background QA Agent) ✅
After all dev sprints complete:
- [x] Run full test suite: `go test ./... -v` (100% pass rate)
- [x] Run linter: `golangci-lint run ./...` (zero errors)
- [x] Verify HTML output:
  - [x] Export test session to HTML
  - [x] Open in browser (Chrome, Firefox, Safari)
  - [x] Verify all sections render correctly
  - [x] Test expandable tool calls work
  - [x] Test lazy-load agent fragments work
  - [x] Verify CSS styling is correct
  - [x] Test responsive design (mobile, tablet, desktop)
- [x] Verify JSONL export:
  - [x] Export test session to JSONL
  - [x] Verify source files copied correctly
  - [x] Verify directory structure is correct
- [x] Cross-platform validation:
  - [x] Test on macOS
  - [x] Test on Ubuntu
  - [x] Test on Windows (path handling)
- [x] Performance testing:
  - [x] Export large session (10k+ entries)
  - [x] Measure export time
  - [x] Verify no memory leaks
- [x] CI passes on all platforms (macOS, Ubuntu, Windows)
- [x] **100% pass**: Committed and PR'd to develop

#### Implementation Summary (2026-02-02) ✅

**Execution Method**: Parallel background dev agents on dedicated sc-git-worktrees + QA verification

| Sprint | Agent | Deliverables | Coverage |
|--------|-------|--------------|----------|
| 8a | Dev #1 | `cmd/export.go` core integration | - |
| 8b | Dev #2 | HTML rendering flow implementation | - |
| 8c | Dev #3 | Integration tests, end-to-end validation, 42 tests | 94.1% |
| 8d | Explore #4 | Gap analysis and quality recommendations | N/A (analysis) |

**QA Reviews**: Background QA agent verified all requirements met

**Total Impact**:
- `cmd/export.go` fully integrated with `pkg/export` package
- HTML export working end-to-end from session to rendered HTML
- Support for both HTML and JSONL formats
- Auto-generated temp folders with session ID + timestamp naming
- Expandable tool calls with lazy-loaded subagent sections
- Manifest.json with tree structure and metadata
- 42+ integration test functions
- 100% test pass rate
- Zero lint errors
- Cross-platform support validated

**PR Merged**:
- PR #11: `feature/phase8-export-integration` → develop

#### Integration Details (From Assessment)

**Function to Call**:
```go
export.ExportSession(projectPath, sessionID string, opts export.ExportOptions) (*export.ExportResult, error)
```

**Required Imports**:
```go
"github.com/randlee/claude-history/pkg/export"
"github.com/randlee/claude-history/internal/jsonl"
"github.com/randlee/claude-history/pkg/agent"
"github.com/randlee/claude-history/pkg/models"
```

**Parameter Mapping**:
- `projectPath` ← CLI arg or current directory
- `sessionID` ← `--session` flag (resolved via Phase 7)
- `opts.OutputDir` ← `--output` flag (empty = auto-generate)
- `opts.ClaudeDir` ← global `--claude-dir` flag

**Return Value Usage**:
- `result.OutputDir` - print to stdout for scripting
- `result.TotalAgents` - include in summary
- `result.Errors` - display as warnings if non-empty
- `result.MainSessionFile` - use for HTML rendering (if format == "html")
- `result.AgentFiles` - map for rendering agent fragments

#### Files to Create/Modify
| Sprint | File | Action | Dev Agent |
|--------|------|--------|-----------|
| 8a | `cmd/export.go` | Modify | #1 |
| 8b | `cmd/export.go` | Modify (HTML flow) | #2 |
| 8c | `cmd/export_integration_test.go` | Create | #3 |
| 8d | Test gap analysis report | Deliver | #4 (Explore) |

---

### Phase 9: Data Model Alignment ✅ COMPLETE

**Priority**: CRITICAL
**Completed**: 2026-02-02
**PR**: #19 (merged to develop)
**Development Method**: Parallel background dev agents on sc-git-worktrees

#### Background: Critical Data Model Mismatch

**Discovery**: Analysis of real Claude Code session data revealed that the current implementation's assumptions about agent spawning are **fundamentally incorrect**.

**What the code currently assumes**:
```go
// tree.go - INCORRECT assumption
if entry.Type == models.EntryTypeQueueOperation && entry.AgentID != "" {
    // Assumes queue-operation entries contain agent spawn info
    result[entry.AgentID] = &SpawnInfo{...}
}
```

**What Claude Code actually does**:
- `queue-operation` entries are for task queue management (enqueue/dequeue/remove), NOT agent spawning
- Agent spawning is recorded in `user` type entries via the `toolUseResult` field
- The `agentId` field at the top level of agent file entries identifies which agent's file it is, not spawn info

#### Real Claude Code Agent Spawn Structure

**Main session entry when Task tool spawns an agent**:
```json
{
  "type": "user",
  "uuid": "7bd059eb-60fb-49b4-92ea-5b6be2a6cfce",
  "sessionId": "926ef72c-163e-4022-bc68-49fcca61ba80",
  "parentUuid": "4e08ee78-a494-47ce-a82c-cf565114a15e",
  "sourceToolAssistantUUID": "4e08ee78-a494-47ce-a82c-cf565114a15e",
  "message": {
    "role": "user",
    "content": [{"type": "tool_result", "tool_use_id": "toolu_01Won...", "content": [...]}]
  },
  "toolUseResult": {
    "isAsync": true,
    "status": "async_launched",
    "agentId": "a6f6578",
    "description": "Review PR #217 workflow changes",
    "prompt": "Review GitHub PR #217...",
    "outputFile": "/tmp/claude/.../tasks/a6f6578.output"
  }
}
```

**Agent file entries** (in `subagents/agent-{id}.jsonl`):
```json
{
  "type": "user",
  "uuid": "b27a9d1b-e153-4494-9746-62a3d84019ac",
  "agentId": "a6f6578",        // Identifies THIS agent's file
  "parentUuid": null,          // null in agent's own entries
  "sessionId": "926ef72c-..."
}
```

**queue-operation entries** (NOT for agent spawning):
```json
{
  "type": "queue-operation",
  "operation": "enqueue",      // or "dequeue", "remove"
  "content": "<task-notification>...</task-notification>",
  "agentId": null              // Always null - not used for spawning
}
```

#### Key Fields for Agent Spawn Detection

| Field | Location | Purpose |
|-------|----------|---------|
| `toolUseResult.agentId` | Main session `user` entries | ID of spawned agent |
| `toolUseResult.status` | Main session `user` entries | "async_launched" = agent spawn |
| `toolUseResult.description` | Main session `user` entries | Agent task description |
| `sourceToolAssistantUUID` | Main session `user` entries | Assistant entry that triggered spawn |
| `agentId` (top-level) | Agent file entries | Identifies which agent's file |
| `parentUuid` | All entries | Parent entry UUID (null in agent files) |

#### Development Sprints (Parallel Execution via sc-git-worktree)

**Sprint 9a: Data Model Updates** (Background Dev Agent #1) 🔲
```
Worktree: wt/phase9-data-model
Branch: feature/phase9-data-model
Parallel: Start immediately
```
- [ ] Update `pkg/models/entry.go`
  - [ ] Add `ToolUseResult` struct:
    ```go
    type ToolUseResult struct {
        IsAsync     bool   `json:"isAsync"`
        Status      string `json:"status"`      // "async_launched", "completed", etc.
        AgentID     string `json:"agentId"`
        Description string `json:"description"`
        Prompt      string `json:"prompt"`
        OutputFile  string `json:"outputFile"`
    }
    ```
  - [ ] Add `ToolUseResult` field to `ConversationEntry`
  - [ ] Add `SourceToolAssistantUUID` field to `ConversationEntry`
  - [ ] Add helper methods:
    - [ ] `HasToolUseResult() bool`
    - [ ] `GetToolUseResult() *ToolUseResult`
    - [ ] `IsAgentSpawn() bool` - returns true if this entry spawned an agent
    - [ ] `GetSpawnedAgentID() string` - extracts agent ID from toolUseResult
- [ ] Create `pkg/models/entry_test.go` additions
  - [ ] Test ToolUseResult parsing from real data format
  - [ ] Test IsAgentSpawn() with various entry types
  - [ ] Test GetSpawnedAgentID() extraction
  - [ ] Test backward compatibility with entries lacking toolUseResult
  - [ ] Target: >90% coverage for new code

**Sprint 9b: Tree Building Fix** (Background Dev Agent #2) 🔲
```
Worktree: wt/phase9-tree-fix
Branch: feature/phase9-tree-fix
Depends: Sprint 9a (needs updated models)
```
- [ ] Update `pkg/agent/tree.go`
  - [ ] Rewrite `buildSpawnInfoMap()` to use correct detection:
    ```go
    func buildSpawnInfoMap(sessionPath string, sessionDir string, agents []models.Agent) map[string]*SpawnInfo {
        result := make(map[string]*SpawnInfo)

        // Scan main session for agent spawns (user entries with toolUseResult)
        _ = jsonl.ScanInto(sessionPath, func(entry models.ConversationEntry) error {
            if entry.IsAgentSpawn() {
                agentID := entry.GetSpawnedAgentID()
                result[agentID] = &SpawnInfo{
                    AgentID:    agentID,
                    SpawnUUID:  entry.UUID,
                    ParentUUID: entry.SourceToolAssistantUUID, // Link to spawning assistant
                }
            }
            return nil
        })

        // Scan agent files for nested spawns
        for _, agent := range agents {
            _ = jsonl.ScanInto(agent.FilePath, func(entry models.ConversationEntry) error {
                if entry.IsAgentSpawn() {
                    agentID := entry.GetSpawnedAgentID()
                    parentUUID := entry.SourceToolAssistantUUID
                    if parentUUID == "" {
                        parentUUID = agent.ID // Fallback: parent is this agent
                    }
                    result[agentID] = &SpawnInfo{
                        AgentID:    agentID,
                        SpawnUUID:  entry.UUID,
                        ParentUUID: parentUUID,
                    }
                }
                return nil
            })
        }

        return result
    }
    ```
  - [ ] Remove queue-operation based detection logic
  - [ ] Update `findParentNode()` to use `sourceToolAssistantUUID` for parent resolution
  - [ ] Add fallback: derive parent from file path if spawn info missing
- [ ] Update `pkg/agent/tree_test.go`
  - [ ] Test tree building with real data format
  - [ ] Test nested agent detection via toolUseResult
  - [ ] Test fallback path-based parent resolution
  - [ ] Test mixed scenarios (some agents with spawn info, some without)
  - [ ] Target: >85% coverage

**Sprint 9c: Test Fixture Updates** (Background Dev Agent #3) 🔲
```
Worktree: wt/phase9-test-fixtures
Branch: feature/phase9-test-fixtures
Parallel: With 9a/9b
```
- [ ] Update `cmd/export_testhelpers_test.go`
  - [ ] Modify `createTestSessionWithAgents()`:
    - [ ] Add `toolUseResult` field to spawn entries
    - [ ] Add `sourceToolAssistantUUID` field
    - [ ] Match real Claude Code structure
  - [ ] Modify `createNestedAgentStructure()`:
    - [ ] Use `user` entries with `toolUseResult` for spawning
    - [ ] Remove reliance on queue-operation for agent spawning
- [ ] Update `test/fixtures/` sample files
  - [ ] Create `sample-session-v2.jsonl` with real format
  - [ ] Create `agent-session-v2.jsonl` with real format
  - [ ] Document format differences in fixture README
- [ ] Update all test files that create mock agent data:
  - [ ] `pkg/agent/agent_test.go`
  - [ ] `pkg/agent/tree_test.go`
  - [ ] `pkg/agent/discovery_test.go`
  - [ ] `cmd/tree_test.go` (if exists)

**Sprint 9d: Query Enhancement** (Background Dev Agent #4) 🔲
```
Worktree: wt/phase9-query-fix
Branch: feature/phase9-query-fix
Depends: Sprint 9a
```
- [ ] Update `cmd/query.go`
  - [ ] When `--agent` flag is provided, read the agent's JSONL file:
    ```go
    func querySession(projectDir, sessionID string, opts FilterOptions) ([]Entry, error) {
        if opts.AgentID != "" {
            // Read agent's file directly instead of main session
            sessionDir := filepath.Join(projectDir, sessionID)
            agentPath := filepath.Join(sessionDir, "subagents", "agent-"+opts.AgentID+".jsonl")
            if paths.Exists(agentPath) {
                return session.ReadSession(agentPath)
            }
            // Fallback: try recursive search for nested agents
            agentFiles, _ := paths.ListAgentFiles(sessionDir)
            if path, ok := agentFiles[opts.AgentID]; ok {
                return session.ReadSession(path)
            }
            return nil, fmt.Errorf("agent not found: %s", opts.AgentID)
        }
        // Read main session
        return session.ReadSession(sessionPath)
    }
    ```
  - [ ] Add `--include-agents` flag for recursive query across all agents
  - [ ] Update help text to clarify `--agent` behavior
- [ ] Update `cmd/query_test.go`
  - [ ] Test agent file query
  - [ ] Test nested agent query
  - [ ] Test `--include-agents` flag
  - [ ] Test error handling for missing agents

#### QA Verification (Background QA Agent - MANDATORY) 🔲
After all dev sprints complete:
- [ ] Run full test suite: `go test ./... -v`
- [ ] Verify 100% test pass rate (zero failures)
- [ ] Check coverage: `go test ./pkg/models/... ./pkg/agent/... -cover` (target >85%)
- [ ] Run linter: `golangci-lint run ./...` (zero errors)
- [ ] Manual verification with real Claude Code data:
  - [ ] Build agent tree for real session with nested agents
  - [ ] Verify tree structure matches actual agent hierarchy
  - [ ] Query specific agent and verify correct entries returned
- [ ] Backward compatibility check:
  - [ ] Old test fixtures still work (graceful degradation)
  - [ ] Sessions without toolUseResult handled gracefully
- [ ] CI passes on all platforms (macOS, Ubuntu, Windows)
- [ ] **100% pass**: Commit and create PR to develop

#### Implementation Summary (2026-02-02) ✅

**Execution Method**: Parallel background dev agents on dedicated sc-git-worktrees + QA verification

| Sprint | Agent | Deliverables | Coverage |
|--------|-------|--------------|----------|
| 9a | Dev #1 | `pkg/models/entry.go` - ToolUseResult struct, helper methods | 90.8% |
| 9b | Dev #2 | `pkg/agent/tree.go` - Fixed buildSpawnInfoMap, nested parent resolution | 87.9% |
| 9c | Dev #3 | Test fixtures - Updated to real Claude Code format, both legacy + modern | N/A |
| 9d | Dev #4 | `cmd/query.go` - `--agent` flag for direct file reading, `--include-agents` | - |

**QA Reviews**: Background QA agent verified all requirements met

**Total Impact**:
- 12 files changed, +1,718 lines
- `pkg/models/entry_test.go` created (519 lines)
- `pkg/agent/tree_test.go` created (421 lines)
- `cmd/query_test.go` created (365 lines)
- All 11 test packages pass
- Zero lint errors
- Cross-platform CI passes (macOS, Ubuntu, Windows)

**PR**: #19 (merged to develop)

#### Files to Create/Modify
| Sprint | File | Action | Dev Agent |
|--------|------|--------|-----------|
| 9a | `pkg/models/entry.go` | Modify | #1 |
| 9a | `pkg/models/entry_test.go` | Modify | #1 |
| 9b | `pkg/agent/tree.go` | Modify | #2 |
| 9b | `pkg/agent/tree_test.go` | Modify | #2 |
| 9c | `cmd/export_testhelpers_test.go` | Modify | #3 |
| 9c | `test/fixtures/*.jsonl` | Modify | #3 |
| 9c | `pkg/agent/*_test.go` | Modify | #3 |
| 9d | `cmd/query.go` | Modify | #4 |
| 9d | `cmd/query_test.go` | Modify | #4 |

#### Migration Notes

**Backward Compatibility**:
- Entries without `toolUseResult` field should be handled gracefully
- Old queue-operation based detection can remain as fallback (with warning log)
- Test fixtures should include both old and new format examples

**Breaking Changes**:
- None expected - this fixes incorrect behavior, doesn't change API

**Deprecation**:
- Queue-operation based agent detection is deprecated
- Will log warning if queue-operation with agentId is encountered (shouldn't happen with real data)

---

### Enhanced Agent Tree ✅ COMPLETE (Phase 5c)

The `tree` command now shows true nested hierarchy using `parentUuid` chains.

#### Behavior
```
Session: 679761ba
├── Main conversation (175 entries)
│   ├── a12eb64 (29 entries)
│   ├── a68b8c0 (28 entries)
│   │   └── nested-agent (15 entries)  ← properly nested under parent
│   │       └── deeply-nested (8 entries)
│   └── a87f5f7 (119 entries)
```

#### Implementation (Sprint 5c)
- [x] Parse `parentUuid` from queue-operation entries in main session
- [x] Parse `parentUuid` from entries in subagent files
- [x] Match agent entries to their spawning queue-operation
- [x] Build recursive tree structure with proper parent-child links (`BuildNestedTree()`)
- [x] Handle edge cases: circular references, orphaned agents, self-refs
- [x] Unit tests for nested tree building (13 new tests, 80.8% coverage)

> ✅ **Fixed in Phase 9**: The original Phase 5c implementation incorrectly assumed queue-operation entries contain agent spawn info. Phase 9 corrected this - agent spawns are detected via `user` entries with `toolUseResult` where `status == "async_launched"`.

---

## Testing Strategy

### Unit Tests
- All packages have corresponding `_test.go` files
- Test edge cases: empty files, malformed JSON, missing fields
- Test cross-platform: Windows paths with `\`, Unix paths with `/`
- Coverage target: >80%

### Integration Tests
- Use test fixtures in `test/fixtures/`
- Test full command execution with sample data
- Test round-trip encoding/decoding

### Cross-Platform Requirements
- Path handling: use `filepath.Join()`, never string concatenation
- Temp folders: use `os.TempDir()`, never hardcoded `/tmp`
- Line endings: handle both `\r\n` and `\n` in JSONL
- Home directory: handle `~` on Unix, `%USERPROFILE%` on Windows

### Build Targets
```bash
# macOS
GOOS=darwin GOARCH=amd64 go build -o claude-history-darwin-amd64
GOOS=darwin GOARCH=arm64 go build -o claude-history-darwin-arm64

# Linux
GOOS=linux GOARCH=amd64 go build -o claude-history-linux-amd64
GOOS=linux GOARCH=arm64 go build -o claude-history-linux-arm64

# Windows
GOOS=windows GOARCH=amd64 go build -o claude-history-windows-amd64.exe
GOOS=windows GOARCH=arm64 go build -o claude-history-windows-arm64.exe
```

---

## Success Criteria

### Phase 4: Tool Filtering
- [ ] `--tool bash` filters to entries containing Bash tool calls
- [ ] `--tool-match "pattern"` filters by tool input regex
- [ ] Case-insensitive tool name matching works
- [ ] Works on macOS, Linux, Windows
- [ ] >80% test coverage for new code

### Phase 5: Agent Discovery
- [ ] `find-agent --explored "file.go"` finds agents that read/wrote file
- [ ] Returns JSONL path for agent resurrection
- [ ] Searches nested agents recursively
- [ ] Works on macOS, Linux, Windows
- [ ] >80% test coverage for new code

### Phase 6: HTML Export ✅
- [x] Generates standalone HTML viewable in any browser
- [x] Expandable tool calls show input and output
- [x] Expandable subagent sections with lazy-load support
- [x] Includes source JSONL for resurrection
- [x] Temp folder naming includes session-id + timestamp
- [x] Works on macOS, Linux, Windows
- [x] >80% test coverage for new code (91%+ achieved)

---

## CLI Command Reference

### Current Commands

```bash
# List all projects
claude-history list

# List sessions in a project
claude-history list /path/to/project
claude-history list --project-id -Users-name-project

# Resolve paths
claude-history resolve /path/to/project
claude-history resolve /path --session <session-id>
claude-history resolve /path --session <session-id> --agent <agent-id>

# Query history
claude-history query /path/to/project
claude-history query /path --start 2026-01-31 --end 2026-02-01
claude-history query /path --type user,assistant
claude-history query /path --session <session-id>
claude-history query /path --session <session-id> --agent <agent-id>
claude-history query /path --format json|list|summary

# Show agent tree
claude-history tree /path/to/project
claude-history tree /path --session <session-id>
claude-history tree /path --format ascii|json|dot
```

### Upcoming Commands (Phase 4-6)

```bash
# Query with tool filtering (Phase 4)
claude-history query /path --tool bash,read
claude-history query /path --tool bash --tool-match "python3"

# Find agents (Phase 5)
claude-history find-agent /path --explored "src/*.go"
claude-history find-agent /path --tool-match "db\.go" --start 2026-01-30

# Export (Phase 6)
claude-history export /path --session <id> --format html
claude-history export /path --session <id> --format html --output ./export/
```

### Global Flags

```
--claude-dir string   Custom ~/.claude directory location
--format string       Output format (varies by command)
-h, --help           Help for command
```

---

## Agent Resurrection (Future Feature)

**Concept**: Given a subagent's JSONL file, restore its conversation context to continue the discussion or ask follow-up questions.

**Use Cases**:
1. Find an agent that explored a specific file, resurrect it to ask detailed questions
2. Share a session export with someone who can resurrect agents to understand the work
3. Continue a subagent's work that was interrupted

**Implementation Notes** (for Phase 5/6):
- `find-agent` command returns JSONL path for resurrection
- `export` command includes source JSONL files in the export bundle
- Resurrection itself is handled by Claude Code, not this CLI
- CLI provides the path; user/skill passes path to new Claude session

**Open Questions**:
- What context format does Claude Code need for resurrection?
- Can we generate a "resurrection prompt" that summarizes the agent's work?
- Should we support partial resurrection (specific conversation range)?

---

## Related Documents

### claude-code-viewer Documentation (Research Notes)
Located in github-research repo with detailed analysis of Claude Code's storage format:

- [`agent-history-storage.md`](../../github-research/claude-code-viewer/agent-history-storage.md) - Technical reference for agent history storage format, entry types, JSONL structure
- [`architecture.md`](../../github-research/claude-code-viewer/architecture.md) - claude-code-viewer architecture analysis
- [`usage.md`](../../github-research/claude-code-viewer/usage.md) - claude-code-viewer usage patterns
- [`security.md`](../../github-research/claude-code-viewer/security.md) - Security considerations

### claude-code-viewer Repository
The web UI for viewing Claude Code history (separate project):

- [Repository](../../claude-code-viewer/) - `/Users/randlee/Documents/github/claude-code-viewer/`
- [README](../../claude-code-viewer/README.md) - Project documentation
- [CLAUDE.md](../../claude-code-viewer/CLAUDE.md) - Claude instructions for that project

### This Project
- [`README.md`](../README.md) - User documentation (to be created)
- [`CLAUDE.md`](../CLAUDE.md) - Claude instructions for this project

---

**Next Steps** (Future Enhancements):
1. Agent resurrection command integration
2. Interactive HTML export viewer improvements
3. Additional export formats (Markdown, PDF)
4. Session comparison/diff tooling
5. Performance optimization for large sessions
