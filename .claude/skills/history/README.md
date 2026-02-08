# Claude History Skill

A Claude Code skill for querying agent history using the `claude-history` CLI tool.

## Files

- `SKILL.md` - Main skill file with YAML frontmatter and instructions
- `README.md` - This documentation file

## Prerequisites

This skill requires the `claude-history` CLI tool to be installed and available in your PATH.

### Installing claude-history

Choose one of the installation methods below:

#### Option 1: Homebrew (macOS/Linux) - Recommended

```bash
brew tap randlee/tap
brew install claude-history
```

#### Option 2: Install Script (macOS/Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/randlee/claude-history/main/install.sh | bash
```

#### Option 3: Go Install

```bash
go install github.com/randlee/claude-history/src@latest
```

#### Option 4: Download Pre-built Binary

1. Download from: https://github.com/randlee/claude-history/releases/latest
2. Extract the archive for your platform
3. Move `claude-history` to a directory in your PATH:
   ```bash
   # macOS/Linux
   sudo mv claude-history /usr/local/bin/

   # Or to user directory (no sudo needed)
   mkdir -p ~/bin
   mv claude-history ~/bin/
   export PATH="$HOME/bin:$PATH"  # Add to ~/.bashrc or ~/.zshrc
   ```

#### Option 5: Build from Source

```bash
git clone https://github.com/randlee/claude-history.git
cd claude-history/src
go build -o ../bin/claude-history .
sudo mv ../bin/claude-history /usr/local/bin/
```

### Verify Installation

```bash
claude-history --version
```

Should output the version number (e.g., `claude-history version 0.2.0`).

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

## Troubleshooting

### Command Not Found

If you see `claude-history: command not found`:

1. **Check if installed:**
   ```bash
   which claude-history
   ```

2. **Check PATH:**
   ```bash
   echo $PATH
   ```

3. **Reload shell:**
   ```bash
   source ~/.bashrc  # or ~/.zshrc
   ```

### Permission Denied

```bash
chmod +x $(which claude-history)
```

### macOS Security Warning

If macOS blocks the binary:

```bash
xattr -d com.apple.quarantine $(which claude-history)
```

Or go to **System Preferences → Security & Privacy** and click **"Allow Anyway"**.

### Outdated Version

**Homebrew:**
```bash
brew upgrade claude-history
```

**Go install:**
```bash
go install github.com/randlee/claude-history/src@latest
```

## See Also

- [GitHub Repository](https://github.com/randlee/claude-history)
- [Full Documentation](https://github.com/randlee/claude-history#readme)
- [Release Notes](https://github.com/randlee/claude-history/releases)
- [Report Issues](https://github.com/randlee/claude-history/issues)
