# History Skill Metadata

This directory contains metadata for the `/history` skill family.

## Files

### bookmarks.jsonl
Bookmarked agents that can be easily resurrected by name.

**Schema:**
```json
{
  "bookmark_name": "string",      // Friendly name for resurrection
  "session_id": "string",         // Full session ID
  "agent_id": "string",           // Agent ID (can be prefix)
  "timestamp": "string",          // ISO 8601 creation timestamp
  "description": "string",        // What this agent does
  "project_path": "string",       // Original project path
  "source_docs": ["string"],      // Optional: source documents analyzed
  "capabilities": ["string"],     // What this agent can do
  "usage": "string",              // How to use this agent
  "tags": ["string"]              // Categories for search
}
```

**Usage:**
```bash
# Resurrect by bookmark name
/resurrect --bookmark agent-design-expert

# Or manually
/resurrect --session <session_id> --agent <agent_id>
```

### resurrection-log.jsonl
Log of all agent resurrections (future implementation).

**Schema:**
```json
{
  "session_id": "string",         // Resurrected agent's session
  "agent_id": "string",           // Resurrected agent's ID
  "timestamp": "string",          // When resurrection happened
  "description": "string",        // What the agent did originally
  "project_path": "string",       // Original project
  "resurrector_session": "string",// Current session
  "resurrector_agent": "string",  // Current agent (or "main")
  "query": "string",              // Question asked
  "search_method": "string",      // "direct" | "fuzzy" | "bookmark"
  "outcome": "string",            // "success" | "failed" | "partial"
  "tokens_used": number,          // Cost of resurrection
  "duration_ms": number,          // Duration
  "error": "string"               // If outcome=failed
}
```

## Current Bookmarks

1. **agent-design-expert** (`a47c2de`)
   - Expert in Claude Code agent design best practices
   - Analyzed guidelines-0.4.md and tool-use-best-practices.md
   - Use for reviewing agent prompts before deployment
   - Session: 24a04e6d-4ea1-4a28-8abf-eb4b88962e3a

## Adding New Bookmarks

```bash
# Append to bookmarks.jsonl
echo '{"bookmark_name":"my-expert",...}' >> .sc/history/bookmarks.jsonl
```

Or use the `/history bookmark` command (future implementation).
