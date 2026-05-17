# Installation And Troubleshooting

This skill requires:
- `claude-history` CLI `v0.2.0+`
- `python3`

## Check First

```bash
which claude-history && claude-history --version
python3 --version
```

If both commands work, skip installation.

## Find Existing Install

Check common locations if `claude-history` is not on PATH:

```bash
for p in "$HOME/.local/bin/claude-history" \
  "$(python3 -m site --user-base 2>/dev/null)/bin/claude-history" \
  "/opt/homebrew/bin/claude-history"; do
  [ -x "$p" ] && echo "Found at: $p" && break
done
```

You can also point the skill at a local build with `.sc/history/config.yml`:

```yaml
cli:
  path: /full/path/to/claude-history
```

## Install

Choose one of these install paths:

### Homebrew (macOS/Linux)

```bash
brew tap randlee/tap
brew install claude-history
```

### Install Script (macOS/Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/randlee/claude-history/main/install.sh | bash
```

### Go Install

```bash
go install github.com/randlee/claude-history/src@latest
```

### Prebuilt Release

Download the current release archive from:
- `https://github.com/randlee/claude-history/releases/latest`

Extract `claude-history` and move it into a directory on PATH.

### Build From Source

```bash
git clone https://github.com/randlee/claude-history.git
cd claude-history/src
go build -o ../bin/claude-history .
```

### Winget (Windows)

```powershell
winget install randlee.claude-history
```

## Minimum Version

This skill expects `claude-history` CLI `v0.2.0` or newer.

Upgrade examples:

```bash
brew upgrade claude-history
go install github.com/randlee/claude-history/src@latest
```

## PATH Troubleshooting

Claude Code may run with a smaller PATH than your interactive shell.

Use the binary directly for the current session:

```bash
/full/path/to/claude-history --version
```

Or export the containing directory:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

## Validation

```bash
claude-history --version
python3 --version
```

Expected result:
- `claude-history` prints a version `>= 0.2.0`
- `python3` resolves on PATH

## Synaptic Canvas

If you are using Synaptic Canvas package management, use:

```bash
sc install claude-history
sc upgrade claude-history
sc uninstall claude-history
```

These commands automate installation and lifecycle management. The manual steps
above remain the normal path for non-Synaptic Canvas users.

## Known Issues

### `claude-history: command not found`

- verify PATH
- check the fallback locations above
- use `.sc/history/config.yml` for a local build

### Version too old

Upgrade with your original install method, then rerun:

```bash
claude-history --version
```

### macOS quarantine warning

```bash
xattr -d com.apple.quarantine /full/path/to/claude-history
```

### Local build not on PATH

Create `.sc/history/config.yml` and point it at the built binary.
