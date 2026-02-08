## Claude History v{{.Version}}

CLI tool for programmatic access to Claude Code's agent history storage.

### Installation

**Homebrew (macOS/Linux):**
```bash
brew tap randlee/tap
brew install claude-history
```

**winget (Windows):**
```powershell
winget install randlee.claude-history
```

**Install Script (macOS/Linux):**
```bash
curl -fsSL https://raw.githubusercontent.com/randlee/claude-history/main/install.sh | bash
```

**Go install:**
```bash
go install github.com/randlee/claude-history/src@{{.Tag}}
```

**Direct Download:**

Download pre-built binaries from the [Assets](#assets) section below for your platform:
- macOS: `claude-history_{{.Version}}_darwin_amd64.tar.gz` (Intel) or `claude-history_{{.Version}}_darwin_arm64.tar.gz` (Apple Silicon)
- Linux: `claude-history_{{.Version}}_linux_amd64.tar.gz` or `claude-history_{{.Version}}_linux_arm64.tar.gz`
- Windows: `claude-history_{{.Version}}_windows_amd64.zip`

Extract and move the `claude-history` binary to a directory in your `$PATH`.

### Quick Start

```bash
# List sessions for a project
claude-history list /path/to/project

# Show agent hierarchy
claude-history tree /path/to/project --session abc123

# Query conversation history
claude-history query /path/to/project --type user,assistant

# Export to HTML
claude-history export /path/to/project --session abc123 --open
```

### Documentation

- [README](https://github.com/randlee/claude-history#readme) - Full documentation and usage examples
- [CLAUDE.md](https://github.com/randlee/claude-history/blob/main/CLAUDE.md) - Development guidelines
