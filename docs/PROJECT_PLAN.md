# Claude History CLI Tool - Project Plan

**Document Version**: 2.3
**Created**: 2026-02-01
**Updated**: 2026-02-01 (Phase 6 completion)
**Language**: Go
**Status**: In Development

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
| `queue-operation` | Subagent spawns | ❌ No |
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
- ✅ Phase 6: HTML Export (`export` command)

### In Progress
- None

### Upcoming Phases
- All phases complete! Future enhancements TBD.

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

**Next Steps**: All planned phases complete. Consider future enhancements:
- Agent resurrection command integration
- Interactive HTML export viewer improvements
- Additional export formats (Markdown, PDF)
- Session comparison/diff tooling
