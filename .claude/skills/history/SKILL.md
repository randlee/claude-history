---
name: history
description: Query Claude Code agent history for a project path
version: 1.0.0
author: claude-history project
parameters:
  - name: action
    type: string
    description: "Action to perform: list, query, tree, find-agent, export"
    required: false
    default: list
  - name: path
    type: string
    description: "Project path to query (defaults to current directory)"
    required: false
  - name: session
    type: string
    description: "Session ID or prefix to filter by"
    required: false
  - name: agent
    type: string
    description: "Agent ID or prefix to filter by"
    required: false
---

# Claude History Query Skill

You are using the **claude-history** CLI tool to query Claude Code's agent history storage.

## Tool Location

The `claude-history` binary is located at: `/Users/randlee/Documents/github/claude-history/bin/claude-history`

If not built, you can build it from source:
```bash
cd /Users/randlee/Documents/github/claude-history/src && go build -o ../bin/claude-history .
```

## Usage Instructions

### Available Commands

1. **list** - List projects or sessions
   ```bash
   claude-history list [path]
   ```

2. **query** - Query history with filters
   ```bash
   claude-history query <path> [flags]
   ```

3. **tree** - Show agent hierarchy
   ```bash
   claude-history tree <path> --session <session-id>
   ```

4. **find-agent** - Find agents by task description
   ```bash
   claude-history find-agent <path> [search-terms...]
   ```

5. **export** - Export session to HTML
   ```bash
   claude-history export <path> --session <session-id> [flags]
   ```

6. **resolve** - Resolve paths (debugging)
   ```bash
   claude-history resolve <path> [flags]
   ```

### Common Flags

- `--session <id>` - Filter by session ID (supports git-style prefixes)
- `--agent <id>` - Filter by agent ID (supports git-style prefixes)
- `--type <types>` - Filter by entry types (user,assistant,system,etc.)
- `--start <date>` - Filter entries after date (YYYY-MM-DD)
- `--end <date>` - Filter entries before date
- `--tool <name>` - Filter by tool name (exact match)
- `--tool-match <pattern>` - Filter by tool name pattern (regex)
- `--format <fmt>` - Output format: text, json, tree (default: text)
- `--output <file>` - Write output to file

### Workflow Based on User Request

**When user provides a path:**
1. First, run `list` to see available sessions
2. If they want to explore a specific session, use `tree` to see agent hierarchy
3. Use `query` with appropriate filters to get detailed information
4. Use `find-agent` to search for agents working on specific topics

**When user wants to search for specific work:**
1. Use `find-agent` with search terms
2. Once you identify relevant sessions/agents, use `query` to get details
3. Use `tree` to understand the agent hierarchy

**When user wants to export:**
1. Use `export` command with session ID
2. Add `--open` flag to open in browser automatically
3. Use `--template` flag for custom styling if needed

### Examples

```bash
# List all sessions for a project
claude-history list /path/to/project

# Show agent tree for a session
claude-history tree /path/to/project --session abc123

# Query all user messages in a session
claude-history query /path/to/project --session abc123 --type user

# Find agents working on "authentication"
claude-history find-agent /path/to/project authentication

# Export session to HTML
claude-history export /path/to/project --session abc123 --output report.html --open

# Query specific agent's work (supports git-style prefixes)
claude-history query /path/to/project --session abc123 --agent def456
claude-history query /path/to/project --session 8c43ec8 --agent ac8c7ba

# Query agent by full ID
claude-history query /path/to/project --session 8c43ec84-09ad-4dc7-bcf7-17f209e983f0 --agent ac8c7ba

# Find agent in tree, then query it
claude-history tree /path/to/project --session abc123  # Shows agent IDs
claude-history query /path/to/project --session abc123 --agent a059688

# Search for tool usage
claude-history query /path/to/project --tool Read --format json

# Filter by date range
claude-history query /path/to/project --start 2026-01-01 --end 2026-01-31

# Query specific agent with type filter
claude-history query /path/to/project --session abc123 --agent def456 --type assistant
```

### Entry Types Reference

- `user` - User messages (prompts or tool results)
- `assistant` - Claude responses with text and tool_use blocks
- `system` - System events and hook summaries
- `queue-operation` - Subagent spawn triggers
- `progress` - Status updates
- `file-history-snapshot` - File state captures
- `summary` - Conversation summaries

## Your Task

Based on the user's request parameters:
- **action**: ${action}
- **path**: ${path} (or use current working directory if not provided)
- **session**: ${session}
- **agent**: ${agent}

Construct and execute the appropriate `claude-history` command(s) to fulfill the user's request.

**Important:**
1. Always use the full path to the binary: `/Users/randlee/Documents/github/claude-history/bin/claude-history`
2. If the binary doesn't exist, offer to build it first
3. Parse the output and present it in a clear, readable format
4. For large outputs, consider using pagination or filtering
5. Session and agent IDs support git-style prefixes (first 7+ characters)
6. When showing results, explain what you found in context

## Response Format

After executing the command:
1. Summarize what you searched for
2. Present the key findings
3. Suggest follow-up queries if relevant
4. Explain any interesting patterns or insights
