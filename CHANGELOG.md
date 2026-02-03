# Changelog

All notable changes to this project will be documented in this file.

## [0.10.0] - 2026-02-02

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
