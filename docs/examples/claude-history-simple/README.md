# Example: Simple Claude Code Session Export

## Description

This example shows how to export a single Claude Code session from the `claude-history` project to a standalone HTML format. The export creates a complete web application that displays the conversation flow with user prompts, Claude's responses, tool calls, and results.

## Project Information

- **Project Path**: `/Users/randlee/Documents/github/claude-history`
- **Session ID**: `967e501d-aa48-4eab-a21f-b17058bd78f0`
- **Session Date**: 2026-02-02
- **Message Count**: 13 messages
- **Total Agents**: 9 (main + 8 subagents)
- **Topic**: Discussion about Phase 9 planning and CI test failures

## Command Used

```bash
./src/claude-history export "/Users/randlee/Documents/github/claude-history" \
  --session 967e501d-aa48-4eab-a21f-b17058bd78f0 \
  --output docs/examples/claude-history-simple/export
```

## Export Structure

The export creates a complete standalone folder:

```
export/
├── index.html          # Main conversation view (570KB)
├── manifest.json       # Session metadata and agent tree
├── agents/             # Lazy-loaded subagent HTML files
│   ├── a1.html
│   ├── a2.html
│   └── ...
├── source/             # Original JSONL files (for resurrection)
│   └── 967e501d.jsonl
└── static/             # CSS and JavaScript
    ├── style.css
    └── script.js
```

## What the HTML Shows

The generated HTML export (`export/index.html`) contains:

- **User Messages**: The initial prompt and follow-up questions
- **Assistant Responses**: Claude's replies with reasoning and explanations
- **Tool Calls**: Structured tool invocations (Read, Edit, Bash, etc.)
- **Tool Results**: Output from each tool execution
- **Timestamps**: When each message was created
- **Formatted Content**: Syntax-highlighted code blocks and structured data

## Use Cases

This type of query is useful for:

1. **Session Review**: Reviewing a specific conversation to understand what was discussed
2. **Debugging**: Analyzing tool calls and their results to troubleshoot issues
3. **Documentation**: Creating records of development sessions
4. **Learning**: Studying interaction patterns between user and assistant
5. **Auditing**: Tracking what changes were made and why

## Features

### Interactive Navigation
- Click on subagent references to expand and view their conversations
- Lazy-loading keeps initial page load fast
- Collapsible sections for better organization

### Complete Content
- Full conversation history preserved
- All tool calls with syntax-highlighted inputs/outputs
- Original JSONL files included for Claude Code resurrection
- Timestamps and metadata for each message

### Portable & Shareable
- Self-contained folder with no external dependencies
- Can be opened directly in any browser
- Can be zipped and shared with others
- No server required

## Viewing the HTML

Open `export/index.html` in any web browser:

```bash
open export/index.html  # macOS
xdg-open export/index.html  # Linux
start export/index.html  # Windows
```

Or start a local server:

```bash
cd export
python3 -m http.server 8000
# Visit http://localhost:8000
```

## Export Formats

The `claude-history export` command supports two formats:

### HTML (default)
- Complete web application with styled UI
- Lazy-loaded subagent content
- Interactive navigation
- Best for: Viewing, sharing, documentation

```bash
claude-history export /path --session <id> --format html
```

### JSONL
- Raw session files only
- Smaller size
- Can be imported back into Claude Code
- Best for: Backup, archival, data processing

```bash
claude-history export /path --session <id> --format jsonl
```

## Related Examples

- See `../claude-history-subagents/` for a more complex session with 90+ agents
