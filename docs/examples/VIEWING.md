# Viewing the Example Exports

## Quick Start

### Option 1: Open Directly in Browser (Simplest)

```bash
# macOS
open docs/examples/claude-history-simple/export/index.html
open docs/examples/claude-history-subagents/export/index.html

# Linux
xdg-open docs/examples/claude-history-simple/export/index.html
xdg-open docs/examples/claude-history-subagents/export/index.html

# Windows
start docs\examples\claude-history-simple\export\index.html
start docs\examples\claude-history-subagents\export\index.html
```

### Option 2: Use Local HTTP Server (Recommended for Large Exports)

For better performance, especially with the large subagents example:

```bash
# Start server for simple example
cd docs/examples/claude-history-simple/export
python3 -m http.server 8000
# Visit http://localhost:8000

# Start server for subagents example (different terminal)
cd docs/examples/claude-history-subagents/export
python3 -m http.server 8001
# Visit http://localhost:8001
```

## What You'll See

### Simple Session Export
- **Main Conversation**: 13 messages showing Phase 9 planning discussion
- **Agent References**: Click to expand 8 subagent conversations
- **Tool Calls**: Syntax-highlighted Read, Edit, and Bash operations
- **Navigation**: Collapsible sections and smooth scrolling

### Large Session with Subagents
- **Main Conversation**: 2041 messages from ARCH-HIST autonomous agent
- **90 Agents**: Explore agents, prompt suggestions, compact operations
- **Lazy Loading**: Subagents load on demand for fast initial page load
- **Hierarchy View**: See parent-child relationships between agents

## Navigation Tips

### Finding Specific Content
1. Use browser's Find (Ctrl+F / Cmd+F) to search for keywords
2. Check `manifest.json` to see full agent tree structure
3. Collapse sections you're not interested in
4. Use browser DevTools to inspect data structures

### Performance Optimization
- **Simple Export**: Loads instantly, no optimization needed
- **Large Export**: Use HTTP server, not file:// protocol
- **Lazy Loading**: Subagents only load when you click them
- **Browser Choice**: Modern browsers (Chrome, Firefox, Safari, Edge) all work

## Exploring the Structure

### Files to Check First
1. **index.html** - Main conversation view (start here)
2. **manifest.json** - Session metadata and agent tree
3. **agents/*.html** - Individual subagent conversations
4. **source/*.jsonl** - Original session data (for resurrection)

### What to Look For
- **Spawn Points**: Where main conversation launches background agents
- **Tool Patterns**: Common sequences (Read → Edit → Bash)
- **Agent Types**: Different agent roles (explore, prompt_suggestion, compact)
- **Timestamps**: How long different operations took
- **Conversation Flow**: User prompts → Assistant reasoning → Tool execution

## Sharing Exports

Each export is completely self-contained:

```bash
# Zip for sharing
zip -r simple-session.zip docs/examples/claude-history-simple/export/
zip -r large-session.zip docs/examples/claude-history-subagents/export/

# Recipient can unzip and open index.html
# No installation or dependencies required
```

## Troubleshooting

### Export Doesn't Load
- **Check Path**: Make sure you're opening index.html, not a subdirectory
- **Try HTTP Server**: Some features work better over HTTP than file://
- **Check Console**: Open browser DevTools to see any JavaScript errors

### Slow Performance
- **Use HTTP Server**: Much faster than file:// for large exports
- **Close Other Tabs**: Free up browser memory
- **Try Smaller Export**: Start with simple example first

### Subagents Don't Load
- **Click to Load**: Subagents are lazy-loaded, you must click them
- **Check Network Tab**: See if agent files are loading correctly
- **Verify Files**: Make sure agents/ directory has .html files

## Comparing Examples

Open both exports side-by-side to see:
- How agent count affects structure
- Performance differences (9 vs 90 agents)
- Different agent types in action
- Simple vs complex conversation patterns

## Next Steps

After viewing the examples:
1. Try exporting your own sessions with `claude-history export`
2. Experiment with different session sizes
3. Share exports with teammates
4. Use as documentation for development sessions

For more details, see the README.md in each example directory.
