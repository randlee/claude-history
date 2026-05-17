# Claude History Skill

A Claude Code skill for querying agent history with the `claude-history` CLI.

## Files

- `SKILL.md` - skill entry point
- `README.md` - package overview
- `references/installation-and-troubleshooting.md` - install, upgrade, and troubleshooting guide

## Prerequisites

This skill requires:
- `claude-history` CLI `v0.2.0+`
- `python3`

For installation, upgrade, PATH help, and Synaptic Canvas commands, read:
- [`references/installation-and-troubleshooting.md`](references/installation-and-troubleshooting.md)

## Installing the Skill

To use this skill in other projects, copy the skill directory to your project's `.claude/skills/` directory:

```bash
cp -r /Users/randlee/Documents/github/claude-history/.claude/skills/history \
      /path/to/your/project/.claude/skills/
```

Or create a symlink for automatic updates:

```bash
ln -s /Users/randlee/Documents/github/claude-history/.claude/skills/history \
      /path/to/your/project/.claude/skills/history
```

## Usage

Invoke the skill using the `/history` command in Claude Code:

```
/history
```

### Examples

**List sessions for current project:**
```
/history
```

**List sessions for specific path:**
```
/history action=list path=/Users/name/project
```

**Show agent tree:**
```
/history action=tree path=/Users/name/project session=abc123
```

**Find agents working on specific topic:**
```
/history action=find-agent path=/Users/name/project authentication
```

**Query session history:**
```
/history action=query path=/Users/name/project session=abc123
```

**Query specific agent by ID:**
```
/history action=query path=/Users/name/project session=abc123 agent=def456
```

**Export session to HTML:**
```
/history action=export path=/Users/name/project session=abc123
```

## Parameters

- `action` - Command to run: list, query, tree, find-agent, export (default: list)
- `path` - Project path to query (default: current directory)
- `session` - Session ID or prefix to filter by
- `agent` - Agent ID or prefix to filter by

## Features

- **Session Listing**: See all Claude Code sessions for a project
- **Agent Tree**: Visualize the hierarchy of agents and subagents
- **History Query**: Search conversation history with filters
- **Agent Search**: Find agents working on specific topics
- **HTML Export**: Generate readable HTML reports of sessions
- **Pass-by-Reference**: Share agent analysis across subagents for efficiency

## Agent Delegation

This skill delegates all tool use to the `history-search` agent registered in `.claude/agents/registry.yaml`. The skill should not call `claude-history` directly.

## Troubleshooting

See:
- [`references/installation-and-troubleshooting.md`](references/installation-and-troubleshooting.md)

## See Also

- [GitHub Repository](https://github.com/randlee/claude-history)
- [Full Documentation](https://github.com/randlee/claude-history#readme)
- [Release Notes](https://github.com/randlee/claude-history/releases)
- [Report Issues](https://github.com/randlee/claude-history/issues)
