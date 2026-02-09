# Bookmark Implementation Plan

**Version**: 1.0
**Created**: 2026-02-09
**Status**: Design - Ready for Implementation

---

## Overview

Implement bookmark system for claude-history CLI to enable users to save and quickly access valuable agents from history.

**Phase 1 Scope**: Global personal bookmarks only
- Storage: `~/.claude/bookmarks.jsonl` (JSONL file)
- Integration: Auto-merge bookmark results into `claude-history query` commands
- Language: Go (integrated with existing CLI)

---

## Architecture

### Storage Layer

**Location**: `~/.claude/bookmarks.jsonl`

**Schema**:
```json
{
  "bookmark_id": "bmk-2026-02-09-001",
  "name": "beads-architecture-expert",
  "description": "Explored beads codebase, documented 3-tier architecture with component relationships",
  "agent_id": "agent-a1b2c3d",
  "session_id": "session-xyz789",
  "project_path": "/Users/randlee/Documents/github/beads",
  "original_timestamp": "2026-02-08T14:30:00.000Z",
  "hostname": "macbook-pro.local",
  "bookmarked_at": "2026-02-09T10:00:00.000Z",
  "bookmarked_by": "randlee",
  "scope": "global",
  "tags": ["explore", "architecture", "beads"],
  "resurrection_count": 0,
  "last_resurrected": null
}
```

### CLI Commands

**New subcommands under `claude-history bookmark`**:

```bash
# Add bookmark
claude-history bookmark add \
  --name beads-expert \
  --agent agent-abc123 \
  --session session-xyz \
  --project /path/to/beads \
  --tags "architecture,beads,explore" \
  --description "Explored beads architecture"

# List bookmarks
claude-history bookmark list
claude-history bookmark list --tag architecture
claude-history bookmark list --format json

# Get specific bookmark
claude-history bookmark get beads-expert

# Update bookmark
claude-history bookmark update beads-expert \
  --description "Updated description" \
  --add-tags "python"

# Delete bookmark
claude-history bookmark delete beads-expert

# Search bookmarks
claude-history bookmark search "beads"
```

### Query Integration

**Auto-merge bookmarks into query results**:

```bash
claude-history query /project --text "beads"
```

**Returns merged results**:
```json
[
  {
    "source": "history",
    "agent_id": "agent-abc123",
    "session_id": "session-xyz",
    "timestamp": "2026-02-08T14:30:00Z",
    "message": "Exploring beads architecture...",

    // Enriched with bookmark metadata
    "bookmarked": true,
    "bookmark_id": "bmk-001",
    "bookmark_name": "beads-architecture-expert",
    "bookmark_tags": ["beads", "architecture"]
  },
  {
    "source": "history",
    "agent_id": "agent-def456",
    "bookmarked": false
  }
]
```

**Key behaviors**:
- Bookmark tags participate in search matching
- History results enriched with bookmark metadata when agent is bookmarked
- Single result per agent (no duplicates)
- Agents know if result is bookmarked (don't ask to bookmark again)

---

## Go Package Structure

```
src/
├── cmd/
│   └── bookmark.go              # CLI subcommands
├── pkg/
│   └── bookmarks/
│       ├── bookmark.go          # Core types
│       ├── storage.go           # JSONL read/write
│       ├── validation.go        # Schema validation
│       └── query_integration.go # Merge with history results
└── pkg/
    └── session/
        └── session.go           # Enhanced to include bookmark data
```

### Core Types

```go
// pkg/bookmarks/bookmark.go

type Bookmark struct {
    BookmarkID        string    `json:"bookmark_id"`
    Name              string    `json:"name"`
    Description       string    `json:"description"`
    AgentID           string    `json:"agent_id"`
    SessionID         string    `json:"session_id"`
    ProjectPath       string    `json:"project_path"`
    OriginalTimestamp time.Time `json:"original_timestamp"`
    Hostname          string    `json:"hostname"`
    BookmarkedAt      time.Time `json:"bookmarked_at"`
    BookmarkedBy      string    `json:"bookmarked_by"`
    Scope             string    `json:"scope"`
    Tags              []string  `json:"tags"`
    ResurrectionCount int       `json:"resurrection_count"`
    LastResurrected   *time.Time `json:"last_resurrected,omitempty"`
}
```

### Storage Interface

```go
// pkg/bookmarks/storage.go

type Storage interface {
    // Read all bookmarks
    List() ([]Bookmark, error)

    // Find bookmark by name
    Get(name string) (*Bookmark, error)

    // Add new bookmark
    Add(bookmark Bookmark) error

    // Update existing bookmark
    Update(name string, updates map[string]interface{}) error

    // Delete bookmark
    Delete(name string) error

    // Search by tags/name/description
    Search(query string) ([]Bookmark, error)
}

// JSONLStorage implements Storage for JSONL files
type JSONLStorage struct {
    path string
}
```

### Query Integration

```go
// pkg/bookmarks/query_integration.go

// EnrichWithBookmarks takes history query results and enriches them
// with bookmark metadata if the agent is bookmarked
func EnrichWithBookmarks(results []session.Entry, bookmarks []Bookmark) []session.Entry {
    // Build agent_id → bookmark map
    bookmarkMap := make(map[string]*Bookmark)
    for i := range bookmarks {
        bookmarkMap[bookmarks[i].AgentID] = &bookmarks[i]
    }

    // Enrich results
    for i := range results {
        if bookmark, exists := bookmarkMap[results[i].AgentID]; exists {
            results[i].Bookmarked = true
            results[i].BookmarkID = bookmark.BookmarkID
            results[i].BookmarkName = bookmark.Name
            results[i].BookmarkTags = bookmark.Tags
        }
    }

    return results
}

// MatchBookmarkTags checks if bookmark tags match query
func MatchBookmarkTags(bookmark Bookmark, query string) bool {
    // Check if query matches:
    // - Bookmark name
    // - Any tags
    // - Description
    query = strings.ToLower(query)

    if strings.Contains(strings.ToLower(bookmark.Name), query) {
        return true
    }

    for _, tag := range bookmark.Tags {
        if strings.Contains(strings.ToLower(tag), query) {
            return true
        }
    }

    if strings.Contains(strings.ToLower(bookmark.Description), query) {
        return true
    }

    return false
}
```

---

## Implementation Tasks

### Task 1: Core Bookmark Package
**File**: `src/pkg/bookmarks/bookmark.go`
- [ ] Define `Bookmark` struct
- [ ] Define `Storage` interface
- [ ] Implement `JSONLStorage` struct
- [ ] Implement CRUD operations (Add, Get, List, Update, Delete)
- [ ] Implement bookmark ID generation (`bmk-YYYY-MM-DD-NNN`)

### Task 2: Validation
**File**: `src/pkg/bookmarks/validation.go`
- [ ] Validate bookmark schema (required fields)
- [ ] Validate bookmark name (alphanumeric, hyphens, underscores)
- [ ] Validate agent exists via existing query code
- [ ] Validate project path exists
- [ ] Validate hostname format

### Task 3: CLI Commands
**File**: `src/cmd/bookmark.go`
- [ ] Implement `bookmark add` command
- [ ] Implement `bookmark list` command
- [ ] Implement `bookmark get` command
- [ ] Implement `bookmark update` command
- [ ] Implement `bookmark delete` command
- [ ] Implement `bookmark search` command
- [ ] Add global flags (--format json|text)

### Task 4: Query Integration
**File**: `src/pkg/bookmarks/query_integration.go`
- [ ] Implement `EnrichWithBookmarks()` function
- [ ] Implement `MatchBookmarkTags()` function
- [ ] Modify `src/cmd/query.go` to load bookmarks
- [ ] Auto-merge bookmark data into results
- [ ] Add `bookmarked` field to query output

### Task 5: Session Model Enhancement
**File**: `src/pkg/session/session.go`
- [ ] Add bookmark fields to Entry struct:
  - `Bookmarked bool`
  - `BookmarkID string`
  - `BookmarkName string`
  - `BookmarkTags []string`
- [ ] Update JSON serialization

### Task 6: Testing
**Files**: `src/pkg/bookmarks/*_test.go`
- [ ] Unit tests for storage operations
- [ ] Unit tests for validation
- [ ] Integration tests for CLI commands
- [ ] Integration tests for query enrichment
- [ ] Test JSONL concurrent access (file locking)

### Task 7: Documentation
**Files**: `README.md`, `docs/`
- [ ] Update main README with bookmark commands
- [ ] Add bookmark usage examples
- [ ] Document query integration behavior
- [ ] Update `.claude/skills/bookmark/` with CLI usage

---

## Migration Path to Dolt (Phase 2+)

**Phase 1 (Now)**: JSONL storage
- Simple file format
- No dependencies
- Fast iteration

**Phase 2 (Later)**: Dolt database
- Add `pkg/bookmarks/dolt_storage.go`
- Implement `Storage` interface for Dolt
- Dual-write mode (JSONL + Dolt)
- Migration script: JSONL → Dolt

**Interface-based design** ensures storage backend is swappable.

---

## Testing Strategy

### Unit Tests
```go
func TestBookmarkStorage_Add(t *testing.T) {
    storage := NewJSONLStorage(tempFile())
    bookmark := Bookmark{...}
    err := storage.Add(bookmark)
    assert.NoError(t, err)
}
```

### Integration Tests
```bash
# Test full workflow
go test -tags=integration ./cmd/... -run TestBookmarkWorkflow
```

### Manual Testing
```bash
# Build CLI
cd src && go build -o ../bin/claude-history .

# Test commands
bin/claude-history bookmark add --name test-bookmark --agent agent-123 --session session-456
bin/claude-history bookmark list
bin/claude-history query . --text "test"  # Should show bookmarked: true
```

---

## Success Criteria

- [ ] Can add bookmark via CLI
- [ ] Can list/get/update/delete bookmarks
- [ ] Bookmarks persist to `~/.claude/bookmarks.jsonl`
- [ ] Query results enriched with bookmark metadata
- [ ] Bookmark tags participate in search
- [ ] Results show `bookmarked: true` for bookmarked agents
- [ ] No duplicate results (single result per agent)
- [ ] All tests pass
- [ ] Documentation complete

---

## Future Enhancements (Later Phases)

- Local project bookmarks (`.sc/history/bookmarks.jsonl`)
- Team sharing via Dolt database
- Bookmark collections/folders
- Smart recommendations based on current task
- Bookmark versioning
- Import/export bookmark manifests

---

**Document Status**: Design Complete - Ready for Implementation
**Next Steps**: Begin Task 1 - Implement core bookmark package
