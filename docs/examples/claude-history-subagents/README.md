# Example: Large Session Export with Multiple Subagents

## Description

This example demonstrates exporting a complex Claude Code session with 90+ subagents. The HTML export creates a rich, interactive view of the hierarchical conversation structure, showing how the main session coordinates with background explore agents, prompt suggestions, and other parallel work streams.

## Project Information

- **Project Path**: `/Users/randlee/Documents/github/claude-history`
- **Session ID**: `c04f363a-5592-4c56-b4af-8d9533b5bca1`
- **Session Date**: 2026-02-03
- **Main Conversation**: 2041 messages
- **Total Agents**: 90 (including main session)
- **Export Size**: ~1.4MB (index.html)
- **Topic**: Architecture and history exploration (ARCH-HIST autonomous agent role)

## Command Used

```bash
./src/claude-history export "/Users/randlee/Documents/github/claude-history" \
  --session c04f363a-5592-4c56-b4af-8d9533b5bca1 \
  --output docs/examples/claude-history-subagents/export
```

## Export Structure

The export creates a comprehensive folder with lazy-loaded content:

```
export/
├── index.html          # Main conversation view (1.4MB)
├── manifest.json       # Session metadata and 90-agent tree (29KB)
├── agents/             # Individual subagent HTML files (30 files)
│   ├── aprompt_suggestion-db2774.html
│   ├── a70e460.html
│   ├── ad2309b.html
│   └── ... (27 more)
├── source/             # Original JSONL files for all 90 agents
│   ├── c04f363a.jsonl
│   └── c04f363a/subagents/
└── static/             # Shared CSS and JavaScript
    ├── style.css
    └── script.js
```

## What the HTML Shows

The generated HTML export (`export/index.html`) provides:

### Main Session
- User prompts and Claude's responses
- Tool calls and results from the primary conversation
- Spawn points where subagents were created

### Subagent Hierarchies
- **Background Agents**: Long-running explore or analysis agents
- **Prompt Suggestions**: Quick AI-generated suggestions (prefix: `aprompt_suggestion-`)
- **Compact Operations**: Internal optimization agents (prefix: `acompact-`)
- **Each Subagent Shows**:
  - Agent ID and type
  - Complete conversation history
  - Tool calls and results
  - Final summaries or deliverables

### Visual Organization
- Hierarchical indentation showing parent-child relationships
- Color coding or styling to distinguish agent types
- Navigation aids for large conversation trees
- Timestamps showing parallel execution

## Use Cases

Querying with `--include-agents` is essential for:

1. **Understanding Complex Workflows**: See how main conversation coordinates with background agents
2. **Performance Analysis**: Track how long subagents took and what they accomplished
3. **Debugging Agent Interactions**: Identify issues in agent spawning or communication
4. **Complete Session Records**: Archive entire conversation trees for documentation
5. **Learning Agent Patterns**: Study how Claude uses background agents for different tasks
6. **Quality Assurance**: Review all work done across parallel agents

## Viewing the HTML

Open `export/index.html` in any web browser:

```bash
open export/index.html  # macOS
xdg-open export/index.html  # Linux
start export/index.html  # Windows
```

Or start a local server for better performance:

```bash
cd export
python3 -m http.server 8000
# Visit http://localhost:8000
```

**Performance Note**: The main index.html is 1.4MB, but subagents are lazy-loaded only when clicked, keeping initial page load reasonable.

## Agent Types in This Session

This example includes several types of subagents:

- **Explore Agents** (e.g., `a70e460`, `ad2309b`): Deep codebase analysis
- **Prompt Suggestions** (e.g., `aprompt_suggestion-db2774`): AI-generated follow-up ideas
- **Compact Operations** (e.g., `acompact-53a066`): Internal conversation optimization

## Comparison with Simple Exports

| Feature | Simple Session | Large Session with Agents |
|---------|----------------|---------------------------|
| **Agents** | 9 | 90 |
| **Index Size** | 570KB | 1.4MB |
| **Manifest** | 3.5KB | 29KB |
| **Load Strategy** | All inline | Lazy-loaded |
| **Best For** | Quick review | Deep analysis |
| **Typical Use** | Single task | Complex multi-agent work |

## Agent Breakdown

This session contains various types of subagents:

- **Explore Agents**: Deep codebase analysis (e.g., `a70e460`, `ad2309b`)
- **Prompt Suggestions**: AI-generated follow-up ideas (e.g., `aprompt_suggestion-*`)
- **Compact Operations**: Internal conversation optimization (e.g., `acompact-*`)

Each type serves a different purpose in the overall workflow.

## Related Examples

- See `../claude-history-simple/` for a smaller session (9 agents, easier to explore)

## Export vs Query Commands

### Export Command
- Creates standalone HTML with styling and interactivity
- Includes source JSONL files for resurrection
- Lazy-loads subagents for performance
- Best for: Sharing, archiving, viewing

```bash
claude-history export /path --session <id>
```

### Query Command
- Outputs to stdout (text or JSON)
- Faster for scripting and pipelines
- Can filter by type, date, tool usage
- Best for: Analysis, automation, filtering

```bash
claude-history query /path --session <id> --format json
```

## Tips for Large Sessions

1. **Use prefix matching**: `c04f363a` instead of full ID
2. **Browse with local server**: Better performance than file://
3. **Check manifest.json first**: See agent tree before diving in
4. **Export to JSONL**: Smaller backup format
5. **Filter before exporting**: Use query command to test filters
