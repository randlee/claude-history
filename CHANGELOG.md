# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Bookmark system for saving and organizing valuable agents
  - `bookmark add` - Create bookmarks with tags and descriptions
  - `bookmark list` - List bookmarks with optional tag filtering
  - `bookmark get` - Get detailed bookmark information
  - `bookmark search` - Search by name, description, or tags
  - `bookmark update` - Update bookmark descriptions and tags
  - `bookmark delete` - Remove bookmarks from storage
- Query integration - bookmark metadata automatically enriched in query results
  - Results show `bookmarked: true` for bookmarked agents
  - Bookmark tags participate in text search matching
  - Single result per agent (no duplicates)
- Bookmark storage in `~/.claude/bookmarks.jsonl` (JSONL format)
- Comprehensive validation for bookmark names and fields
- Atomic JSONL writes for data safety
- 82% test coverage on bookmark package (209 new tests)

## [0.3.0] - 2026-02-07

### Added - Phase 15: Query HTML Format & UX Improvements
- **HTML Format for Query Command** - Added `--format html` option to query command with auto-open in browser
- **Clickable File Paths** - Automatic detection and linking of file paths in conversation text
  - Supports Unix absolute paths (`/path/to/file.go`)
  - Supports Windows absolute paths (`C:\path\to\file.go`)
  - Supports relative paths (`./src/main.go`, `src/main.go`)
  - Creates `file://` links that open in Finder/Explorer
  - 113+ file links generated in typical sessions
- **Tool-Only Message Enhancement** - Compact display for messages with only tool calls
  - Headers show tool type: "TOOL: Read", "TOOL: Task", "TOOL: Bash", etc.
  - Inline summaries for task operations (TaskUpdate, TaskCreate, TaskList, TaskGet)
  - Single-line collapsed view, expandable on click
- **Version Management System** - Single source of truth in `pkg/version/version.go`
  - Version constant used in CLI and HTML output
  - Unit test enforces no hardcoded versions
  - Compatible with GoReleaser ldflags override

### Improved
- **HTML Export UX** - Removed "Main Session" breadcrumb badge to save vertical space
- **Navigation** - Hide empty breadcrumbs completely with CSS
- **Tool Headers** - Increased max-width from 300px to 600px for better path visibility
- **Copy Buttons** - Include full CLI commands with complete UUIDs (not just truncated IDs)
- **Role Labels** - Clean styling without boxes (like ORCHESTRATOR/AGENT)

### Fixed
- **Empty Message Bubbles** - Comprehensive filtering and detection
  - Tool-only messages no longer appear as empty "Assistant" boxes
  - Added automated test suite for empty bubble detection
- **File Path Detection** - Fixed markdown processing order to handle paths in inline code blocks
- **Windows CI Tests** - Platform-aware test skipping for Windows-specific path format tests
- **Lint Errors** - Fixed errcheck warnings and gofmt formatting

### Testing
- Added `src/pkg/export/paths_test.go` - File path detection tests (Unix and Windows)
- Added `src/pkg/export/html_empty_test.go` - Empty bubble detection tests
- Added `src/pkg/export/html_empty_full_test.go` - Full integration tests
- Added `src/pkg/export/html_empty_edgecases_test.go` - Edge case tests
- Added `src/pkg/version/version_test.go` - Version consistency unit test
- Added `scripts/check-empty-bubbles.sh` - Automated empty bubble checker

### Documentation
- Updated `README.md` with `--format html` examples and `--include-agents` flag documentation
- Updated `docs/PROJECT_PLAN.md` with Phase 15 completion details
- Added design rationale for path detection (no spaces to avoid false positives)
- Documented Windows 8.3 short filename support (e.g., `RUNNER~1`)

## [0.1.0] - 2026-02-07

### Added
- **Initial Public Release** - First official release of claude-history CLI
- Complete HTML export with Phase 10 enhancements (chat bubbles, color-coded overlays, syntax highlighting)
- Query command with smart preview mode and `--limit` flag for output control
- Session and project listing with agent hierarchy visualization
- Agent discovery by ID prefix (git-style)
- Export formats: HTML (with lazy-loaded agents) and JSONL (backup/restore)
- Copy-to-clipboard buttons for agent resurrection workflow
- Interactive controls: search, expand/collapse all, keyboard shortcuts (Ctrl+K, Ctrl+F)
- Session metadata header and keyboard shortcuts footer
- GoReleaser configuration for multi-platform releases (macOS, Linux, Windows)
- Installation script and Claude Code skill integration

### Fixed
- Query output truncation bug (added `--limit` flag, default 100 chars with smart preview)
- Empty assistant blocks cluttering HTML output (filter whitespace-only entries)
- XML tag formatting in user messages (bash-stdout, bash-stderr styled with card containers)
- Task-notification display (rendered as system events with status icons ✓/✗/⏳)
- Integration test JSON encoding issues (escaped newlines)
- CI/lint failures (gofmt, staticcheck compliance)

### Documentation
- Comprehensive README with installation and usage examples
- CLAUDE.md with project instructions and development guidelines
- PROJECT_PLAN.md tracking implementation phases 1-10
- Issues tracker (docs/issues.md) with sprint planning for Phases 11-12

### Quality Metrics
- **Test Coverage**: 93.4% on export package
- **Test Pass Rate**: 100% (637+ tests passing)
- **CI Checks**: All platforms passing (macOS, Ubuntu, Windows)

---

## [0.10.0] - 2026-02-02 (Internal Development)

### Phase 10: HTML Export Enhancement - Complete

#### New Features
- **Chat Bubble Layout**: User messages align left, assistant messages align right with avatar placeholders
- **CSS Variable System**: Complete HSL color palette with semantic variable naming for dynamic theming
- **Copy-to-Clipboard**: Infrastructure for easy agent ID, file path, and session ID copying (agent resurrection workflow)
- **Color-Coded Overlays**: 14 tool types with distinct color schemes:
  - Tools: Blue/teal with wrench icon
  - Subagents: Purple/violet with agent icon
  - Thinking blocks: Gray/muted with lightbulb icon
  - System messages: Yellow/amber with info icon
- **Syntax Highlighting**: Proper code block rendering with language badges
- **Deep Dive Navigation**: Breadcrumb trails with lazy-loading for nested agent exploration
- **Interactive Controls**: Expand/collapse all, search with highlighting, keyboard shortcuts
- **Session Metadata**: Header with session stats, footer with CLI attribution and keyboard help
- **Full Integration Testing**: 20+ integration test functions covering all Phase 10 features

#### Sprint Breakdown
- **Sprint 10a**: CSS Variable System (PR #21)
- **Sprint 10b**: Chat Bubble Layout (PR #24)
- **Sprint 10c**: Copy-to-Clipboard Infrastructure (PR #22)
- **Sprint 10d**: Color-Coded Overlays (PR #25)
- **Sprint 10e**: Syntax Highlighting (in PR #21)
- **Sprint 10f**: Deep Dive Navigation (PR #27)
- **Sprint 10g**: Interactive Controls (PR #26)
- **Sprint 10h**: Header/Footer Metadata (PR #28)
- **Sprint 10i**: Integration & Polish (PR #29)

#### Quality Metrics
- **Test Coverage**: 93.4% on export package
- **Test Pass Rate**: 100% (637+ tests passing)
- **Linter Issues**: 0
- **CI Checks**: 5 required status checks (Build, Lint, Test×3)

#### Files Added/Modified
- `src/pkg/export/html.go` - Enhanced HTML generation with chat bubble layout, metadata
- `src/pkg/export/navigation.go` - Breadcrumb and deep-dive navigation components
- `src/pkg/export/overlays.go` - Tool and subagent overlay rendering
- `src/pkg/export/templates/style.css` - Complete CSS redesign with variables, dark mode, responsive layout
- `src/pkg/export/templates/clipboard.js` - Copy-to-clipboard infrastructure
- `src/pkg/export/templates/controls.js` - Interactive expand/collapse and search
- `src/pkg/export/templates/navigation.js` - Breadcrumb navigation and lazy loading
- `src/pkg/export/integration_test.go` - 866-line comprehensive integration test suite

#### Breaking Changes
None - all changes are backwards compatible

#### Notes
- Agent resurrection workflow: Copy agent ID from HTML export → paste in Claude terminal
- All Phase 10 worktrees cleaned up after release
- Branch protection enforcement enabled to prevent direct pushes
- Ready for production use

---

## [0.9.0] - Previous Release

See git history for details on Phases 1-9
