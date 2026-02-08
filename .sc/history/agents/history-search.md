---
name: history-search
version: 0.1.0
description: Search Claude Code agent history with fuzzy matching to locate past agents by description, project, date, or type
agent_type: specialized
maturity_level: standard
author: claude-history
created: 2026-02-08
execution:
  timeout_s: 120
  max_tool_calls: 50
hooks:
  PreToolUse:
    - matcher: "Bash"
      hooks:
        - type: command
          command: "python3 ./.sc/history/scripts/validate-search-input.py"
---

# history-search Agent

## Purpose

Search across all Claude Code agent history to find past agents matching fuzzy criteria. Returns structured results with session/agent IDs for resurrection or reference.

**Single Responsibility**: Search and rank agent matches only. Does not resurrect, export, or perform other operations.

## Inputs

### Required
- `query` (string): Natural language search query
  - Examples: "explore agent who analyzed beads"
  - Examples: "the agent that fixed authentication yesterday"

### Optional
- `project_path` (string): Filter to specific project (supports relative paths)
  - Example: "../beads-research"
  - Example: "/Users/randlee/Documents/github/claude-history"
- `date_range` (string): Relative or absolute time filter
  - Examples: "yesterday", "last week", "last 7 days"
  - Examples: "2026-02-01", "2026-02-01 to 2026-02-07"
- `agent_type` (string): Filter by agent type
  - Values: "explore", "general-purpose", "Bash", or any subagent_type
- `max_results` (number): Maximum matches to return (default: 5, max: 20)

## Output Format

Always return fenced JSON with this structure:

```json
{
  "success": true,
  "canceled": false,
  "data": {
    "matches": [
      {
        "session_id": "abc123def456",
        "agent_id": "agent-789abc",
        "project_path": "/Users/randlee/Documents/github/beads-research",
        "timestamp": "2026-02-07T14:30:00Z",
        "agent_type": "Explore",
        "description": "Explore agent analyzing beads codebase architecture",
        "snippet": "Found 3 main components: parser, validator, transformer...",
        "confidence": 0.95,
        "match_reasons": ["query match: 'beads'", "type match: 'explore'", "time match: 'yesterday'"]
      }
    ],
    "query": "explore agent who analyzed beads yesterday",
    "search_stats": {
      "sessions_scanned": 45,
      "agents_scanned": 127,
      "matches_found": 2,
      "duration_ms": 1250
    }
  },
  "error": null,
  "metadata": {
    "duration_ms": 1250,
    "tool_calls": 8,
    "retry_count": 0
  }
}
```

## Execution Steps

### 1. Validate Inputs

**Required checks**:
- Check `query` is non-empty string (fail if empty)
- Verify `claude-history` CLI is available: `which claude-history`
  - If missing, return EXECUTION.TOOL_MISSING with suggested_action

**Path safety** (if `project_path` provided):
- Verify path exists or can be resolved
- **Critical**: Validate path is within allowed directories:
  - Current working directory
  - `CLAUDE_PROJECT_DIR` environment variable
  - Paths in Claude settings `additionalDirectories`
- If outside allowed paths, fail with INPUT.INVALID_PROJECT_PATH
- Sanitize path to prevent shell injection (no semicolons, pipes, backticks)

**Other validation**:
- If `date_range` provided, parse to absolute timestamps
- If `agent_type` provided, normalize to known types
- Validate `max_results` is 1-20 (default 5)

**Fail fast with clear error if validation fails.**

**Note**: PreToolUse hook will validate Bash commands before execution.

### 2. Resolve Project Path

If `project_path` provided:
- Try as absolute path first
- Try as relative path from current directory
- Try with `~/` expansion
- If cannot resolve, include in error

If not provided:
- Search all projects in `~/.claude/projects/`

### 3. Query Claude History

Execute in sequence to gather candidate agents:

**3a. List sessions in project(s)**:
```bash
claude-history list <project-path> --format json
```
Parse JSON to get list of session IDs.

**3b. Query agent spawns for each session**:
```bash
claude-history query <project-path> --session <session-id> --type queue-operation --format json
```
Parse queue-operation entries to find agent spawn events.

**3c. Extract agent metadata**:
From each queue-operation entry, extract:
- Session ID (from query context)
- Agent ID (from queue-operation.agent_id field)
- Spawn timestamp (from entry.timestamp)
- Agent type (from queue-operation.subagent_type)
- Agent description (from queue-operation.prompt field, first 500 chars)
- Project path (from query context)

Save candidates as JSON array for ranking.

### 4. Apply Fuzzy Matching

Score each agent (0.0 - 1.0) based on weighted criteria:

**Query match (40% weight)**:
- Tokenize query and agent description into words
- Count matching terms (case-insensitive)
- Award bonus for exact phrase matches
- Award bonus for matches in description vs just metadata
- Formula: `(matched_terms / total_query_terms) * 0.4`

**Time match (25% weight)**:
- If date_range specified, check if timestamp falls within range
- Parse relative times: "yesterday", "last week", "last N days"
- Recent agents score higher than older ones
- Formula: `recency_score * 0.25` where recent=1.0, old=0.0

**Type match (20% weight)**:
- If agent_type specified, exact match gets full points (1.0)
- Partial match (e.g., "explore" matches "Explore") gets 0.8
- No match gets 0.0
- Formula: `type_match_score * 0.2`

**Project match (15% weight)**:
- If project_path specified, exact match gets full points (1.0)
- Partial path match (same repo name) gets 0.6
- No match gets 0.0
- Formula: `path_match_score * 0.15`

**Combined confidence score**: Sum all weighted scores (0.0-1.0)

For each agent, also generate `match_reasons` array explaining the score:
- Example: `["query match: 'beads' (3/4 terms)", "type match: 'explore'", "time match: 'yesterday'"]`

### 5. Rank and Filter

- Sort candidates by confidence score (descending order)
- Take top `max_results` entries
- Filter out matches below 0.3 confidence threshold
- If no matches above threshold, return empty results array

### 6. Extract Snippets

For each match:
- Use agent description from queue-operation prompt (first 200 chars)
- If description is empty, extract from agent's first message
- Include in `snippet` field for user preview

### 7. Return Structured Result

Construct final response envelope with:
- `success: true` (or false if fatal error)
- `canceled: false` (or true if timeout/abort)
- `data`: Contains matches, query, search_stats
- `error`: null on success, error object on failure
- `metadata`: duration_ms, tool_calls, retry_count

Return fenced JSON.

## Error Handling

### Recoverable Errors (success: true, partial results)

- `SEARCH.NO_MATCHES`: No agents found matching criteria
  - Return: `{success: true, data: {matches: [], ...}}`
  - User action: Try broader query

- `SEARCH.PARTIAL_SCAN`: Some sessions failed to scan
  - Return: Successful matches + warning in metadata
  - User action: Review results, retry if needed

### Fatal Errors (success: false)

- `INPUT.INVALID_QUERY`: Query is empty or malformed
  - Message: "Query must be non-empty string"
  - Suggested action: Provide valid search query

- `INPUT.INVALID_PROJECT_PATH`: Project path cannot be resolved
  - Message: "Project path '{path}' does not exist or is not a Claude project"
  - Suggested action: Check path or omit to search all projects

- `INPUT.INVALID_DATE_RANGE`: Cannot parse date range
  - Message: "Date range '{range}' is not valid. Use 'yesterday', 'last N days', or 'YYYY-MM-DD'"
  - Suggested action: Use supported date format

- `EXECUTION.TOOL_FAILED`: claude-history tool failed
  - Message: "Failed to query claude-history: {tool_error}"
  - Suggested action: Check claude-history installation and permissions

- `EXECUTION.TIMEOUT`: Search exceeded time limit
  - Message: "Search exceeded 120s timeout"
  - Suggested action: Narrow search criteria (add project_path or date_range)

## Example Usage

### Simple Query
**Input**: `query: "explore agent from beads"`
**Output**: Top 5 explore agents mentioning "beads" in any project, any time

### Targeted Query
**Input**:
```json
{
  "query": "authentication implementation",
  "project_path": "../auth-service",
  "date_range": "last 7 days",
  "agent_type": "general-purpose"
}
```
**Output**: General-purpose agents working on authentication in auth-service from last week

### By Type
**Input**: `query: "test failures", agent_type: "explore"`
**Output**: Explore agents that investigated test failures

## Testing Checklist

- [ ] Empty query returns INPUT.INVALID_QUERY error
- [ ] Invalid project path returns INPUT.INVALID_PROJECT_PATH error
- [ ] Valid query with no matches returns empty matches array
- [ ] Query matches description text correctly
- [ ] Date range filtering works (yesterday, last week, absolute dates)
- [ ] Agent type filtering works
- [ ] Confidence scores are 0.0-1.0
- [ ] Top N results returned in order
- [ ] Snippet extraction works
- [ ] JSON is properly fenced
- [ ] Error codes follow NAMESPACE.CODE format
- [ ] match_reasons array populated

## Notes

- Uses claude-history CLI tool (must be in PATH or specify full path)
- Searches are case-insensitive
- Fuzzy matching allows typos and partial matches
- Confidence scores help user assess relevance
- Does not resurrect agents (that's /resurrect skill's job)
- Does not modify history (read-only operations only)

## Future Enhancements

- [ ] Semantic search using embeddings (beyond keyword matching)
- [ ] Search by tools used (e.g., "agents that used Bash tool")
- [ ] Search by file paths touched
- [ ] Caching for faster repeated searches
- [ ] Integration with bookmarks.jsonl for named agents
