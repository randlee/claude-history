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

You are using the **claude-history** CLI tool to query Claude Code's agent
history storage.

## Step 1 — Verify `claude-history` Installation

```bash
which claude-history && claude-history --version
python3 --version
```

If not found on PATH, also check common install locations:

```bash
for p in "$HOME/.local/bin/claude-history" \
  "$(python3 -m site --user-base 2>/dev/null)/bin/claude-history" \
  "/opt/homebrew/bin/claude-history"; do
  [ -x "$p" ] && echo "Found at: $p" && break
done
```

If found at a non-PATH location, use the full path for all commands, or export
that directory to PATH for this session:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

This skill requires `claude-history` CLI `v0.2.0+` and `python3`.

If `claude-history` is not installed or is too old, read
[`references/installation-and-troubleshooting.md`](references/installation-and-troubleshooting.md)
before proceeding.

## Synaptic Canvas

If you are using Synaptic Canvas instead of manual installation:

```bash
sc install claude-history
sc upgrade claude-history
sc uninstall claude-history
```

## Agent Delegation (Required)

This skill must **delegate all tool use** to the `history-search` agent. Do not call `claude-history` directly from the skill.

1. Invoke the `history-search` agent (from `.claude/agents/registry.yaml`).
2. Pass a fenced JSON payload that matches the agent's Input Contract.
3. Receive a fenced JSON envelope and format the result for the user.

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
- `--format <fmt>` - Output format: text, json, tree, html (default: text)
- `--output <file>` - Write output to file

### Workflow Based on User Request

**When user provides a path:**
1. Delegate `list` to the agent to see available sessions
2. If they want to explore a specific session, delegate `tree` to see agent hierarchy
3. Delegate `query` with appropriate filters to get detailed information
4. Delegate `find-agent` to search for agents working on specific topics

**When user wants to search for specific work:**
1. Delegate `find-agent` with search terms
2. Once you identify relevant sessions/agents, delegate `query` to get details
3. Delegate `tree` to understand the agent hierarchy

**When user wants to export:**
1. Delegate `export` with session ID to create HTML files
2. Or delegate `query --format html` to generate and auto-open HTML report in browser

### Examples

These command examples are executed by the `history-search` agent on the skill's behalf.

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
claude-history export /path/to/project --session abc123

# Generate and auto-open HTML report
claude-history query /path/to/project --session abc123 --format html

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

## Passing Agent to Subagents by Reference

**Use Case:** Share the results of a previous agent's work with new agents without duplicating the analysis.

Instead of re-analyzing a repository or re-exploring code, you can pass a **reference** to an agent that already did the work. New agents can query the previous agent's findings directly.

### Example: Extracting Final Analysis from a Subagent

```bash
claude-history query <project-path> --session <session-id> --agent <agent-id> --type assistant --format json | jq -r '.[-1].message.content[0].text'
```

**What this does:**
- Queries a specific session (the main conversation)
- Extracts output from a specific agent (e.g., the explore agent that analyzed the repository)
- Filters to `assistant` messages only (the agent's responses)
- Returns JSON format to avoid truncation
- Uses `jq` to extract just the text from the final message

**Result:** Gets the complete final analysis (e.g., ~26,000 token comprehensive report) without re-analyzing.

### When to Use This Pattern

✅ **Use pass-by-reference when:**
- Multiple agents need the same analysis (e.g., updating multiple docs from one repository analysis)
- The analysis is expensive (time/tokens) to regenerate
- You want consistent information across agents
- Previous agent did deep exploration you want to reuse

❌ **Don't use when:**
- Information is outdated (repository changed significantly)
- Each agent needs different depth/focus
- The reference agent didn't cover what you need

### Benefits

- **10x faster** - Agents skip re-exploration
- **Token efficient** - Share one analysis across many agents
- **Consistency** - All agents work from same source
- **Scalability** - One explore agent → N documentation agents

### Passing to Subagents

When giving this pattern to subagents, provide:

1. **Exact command** in a fenced code block (prevents typos)
2. **Session and agent IDs** (the reference to query)
3. **Stop instruction** - "Stop immediately if command fails"
4. **What to expect** - "Returns ~26,000 token analysis"

**Template:**
```markdown
Use the Bash tool to run this exact command:

```bash
claude-history query <project-path> --session <session-id> --agent <agent-id> --type assistant --format json | jq -r '.[-1].message.content[0].text'
```

This returns the complete analysis from agent <agent-id>.

**CRITICAL:** Stop immediately if command fails or returns empty output.
```

## Your Task

Based on the user's request parameters:
- **action**: ${action}
- **path**: ${path} (or use current working directory if not provided)
- **session**: ${session}
- **agent**: ${agent}

Delegate the request to `history-search` with a fenced JSON payload. Do not run
`claude-history` directly from the skill. The agent owns CLI execution,
structured results, and retry/error handling.

**Important:**
1. Ensure the delegated payload matches the agent Input Contract.
2. If the CLI is missing or too old, stop and direct the user to `references/installation-and-troubleshooting.md`.
3. Session and agent IDs support git-style prefixes (first 7+ characters).
4. Format the agent result clearly for the user and explain any relevant findings.

## Response Format

After executing the command:
1. Summarize what you searched for
2. Present the key findings
3. Suggest follow-up queries if relevant
4. Explain any interesting patterns or insights
