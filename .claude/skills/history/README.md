# Claude History Skill

A Claude Code skill for querying agent history using the `claude-history` CLI tool.

## Files

- `SKILL.md` - Main skill file with YAML frontmatter and instructions
- `README.md` - This documentation file

## Installation

This skill is automatically available when working in the claude-history project directory.

To use it in other projects, copy this skill directory to your project's `.claude/skills/` directory:

```bash
cp -r /Users/randlee/Documents/github/claude-history/.claude/skills/history \
      /path/to/your/project/.claude/skills/
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

## Requirements

- The `claude-history` binary must be built:
  ```bash
  cd src && go build -o ../bin/claude-history .
  ```

## See Also

- [claude-history CLI documentation](../../../docs/PROJECT_PLAN.md)
- [CLAUDE.md](../../../CLAUDE.md) - Project instructions
