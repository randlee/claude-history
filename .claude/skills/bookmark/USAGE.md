# Bookmark Skill Usage

## CLI Commands

The bookmark system is integrated into the claude-history CLI.

## Quick Examples

### Add a bookmark
```bash
claude-history bookmark add \
  --name my-agent \
  --agent agent-abc123 \
  --session session-xyz789 \
  --project /path/to/project \
  --tags "tag1,tag2,tag3" \
  --description "Description of the agent's work"
```

### List all bookmarks
```bash
# List all bookmarks (text format)
claude-history bookmark list

# Filter by tag
claude-history bookmark list --tag architecture

# JSON output for scripting
claude-history bookmark list --format json
```

### Get bookmark details
```bash
claude-history bookmark get my-agent
```

### Search bookmarks
```bash
# Search by name, description, or tags
claude-history bookmark search "beads architecture"
```

### Update bookmark
```bash
# Update description
claude-history bookmark update my-agent \
  --description "Updated description"

# Add tags
claude-history bookmark update my-agent \
  --add-tags "python,advanced"
```

### Delete bookmark
```bash
claude-history bookmark delete my-agent
```

## Query Integration

Bookmarks are automatically enriched in query results:

```bash
claude-history query /project --text "search term"
```

Results include:
- `bookmarked: true` - indicates if agent is bookmarked
- `bookmark_id` - unique bookmark identifier (e.g., "bmk-2026-02-09-001")
- `bookmark_name` - user-provided bookmark name
- `bookmark_tags` - array of tags for organization

## Storage Location

Bookmarks are stored in `~/.claude/bookmarks.jsonl` (JSONL format for easy editing and backup).

## See Also

- Main documentation: [README.md](../../../README.md)
- Detailed requirements: [requirements.txt](./requirements.txt)
- CLI help: `claude-history bookmark --help`
