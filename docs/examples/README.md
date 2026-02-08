# claude-history Export Examples

This directory contains real-world examples of `claude-history export` output, demonstrating HTML exports of Claude Code sessions with varying complexity.

**Quick Links**: [Viewing Guide](VIEWING.md) | [Project Documentation](../../README.md)

## Available Examples

### 1. Simple Session Export
**Directory**: `claude-history-simple/`

A straightforward example showing how to export a Claude Code session with 9 agents. Perfect for understanding the export format and interactive HTML structure.

- **Session Size**: 13 messages (main conversation)
- **Total Agents**: 9
- **Export Size**: 570KB (index.html)
- **Command**: `claude-history export`
- **Best For**: Learning export basics, understanding HTML structure

[View README](claude-history-simple/README.md) | [Open Export](claude-history-simple/export/index.html)

### 2. Large Session with Multiple Subagents
**Directory**: `claude-history-subagents/`

An advanced example showing a complex session with 90 subagents, including background explore agents, prompt suggestions, and compact operations. Demonstrates lazy-loading and hierarchical navigation.

- **Session Size**: 2041 messages (main conversation)
- **Total Agents**: 90
- **Export Size**: 1.4MB (index.html)
- **Command**: `claude-history export`
- **Best For**: Understanding agent hierarchies, viewing large sessions, performance patterns

[View README](claude-history-subagents/README.md) | [Open Export](claude-history-subagents/export/index.html)

## Quick Start

To create your own exports:

```bash
# Build the tool
cd src && go build -o claude-history .

# List available sessions
./claude-history list /path/to/your/project

# Export to HTML (default format)
./claude-history export /path/to/project --session <session-id>

# Export to specific folder
./claude-history export /path/to/project --session <session-id> --output ./my-export/

# Export just JSONL (smaller, for backup)
./claude-history export /path/to/project --session <session-id> --format jsonl
```

## Example Structure

Each example directory contains:
- **README.md**: Documentation including commands used, project info, and use cases
- **export/**: Complete standalone HTML export that can be opened in any browser
  - `index.html`: Main conversation view
  - `manifest.json`: Session metadata and agent tree
  - `agents/`: Lazy-loaded subagent content
  - `source/`: Original JSONL files
  - `static/`: CSS and JavaScript

## Reproducing Examples

All examples use actual sessions from the `claude-history` project itself. The commands shown in each README are copy-pasteable and will work on any system with `claude-history` installed.

## Export vs Query

The `claude-history` tool provides two ways to access session data:

### Export Command
- Creates standalone HTML with interactive UI
- Includes source JSONL files for resurrection
- Lazy-loads subagents for performance
- Self-contained and shareable
- **Best for**: Viewing, sharing, archiving

```bash
claude-history export /path --session <id>
```

### Query Command
- Outputs to stdout (text, JSON, or colored text)
- Faster for scripting and pipelines
- Can filter by type, date, tool usage
- **Best for**: Analysis, automation, filtering

```bash
claude-history query /path --session <id> --format json
```

## Tips for Creating Exports

1. **Start Small**: Export short sessions first to understand the structure
2. **Check the Tree**: Use `claude-history tree` to preview agent hierarchy
3. **Use Prefix Matching**: Save typing with short session IDs (e.g., `abc123` instead of full UUID)
4. **Local Server**: For large exports, serve via HTTP for better performance
5. **JSONL Backups**: Use `--format jsonl` for lightweight backups
6. **Verify Manifest**: Check `manifest.json` to understand session metadata

## Related Documentation

- [Project Plan](../PROJECT_PLAN.md) - Implementation phases and feature status
- [README](../../README.md) - Main project documentation
- [CLAUDE.md](../../CLAUDE.md) - Development guidelines

## Contributing Examples

To add a new example:

1. Create a new directory: `docs/examples/your-example-name/`
2. Generate export: Run `claude-history export` command
3. Write README.md: Document the commands, use cases, and key insights
4. Update this file: Add your example to the list above

Keep examples focused on demonstrating specific features or use cases.

## What Makes a Good Example

- **Clear Purpose**: Explains what feature or workflow it demonstrates
- **Reproducible**: Includes exact commands that others can run
- **Well-Documented**: README explains what to look for in the export
- **Appropriate Size**: Large enough to be interesting, small enough to load quickly
- **Real Usage**: Shows actual development sessions, not contrived examples
