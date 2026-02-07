# Homebrew Tap Setup

This document explains the Homebrew tap configuration for claude-history.

## What is a Homebrew Tap?

A "tap" is a third-party repository of Homebrew formulas. Users can add your tap and install your tool with:

```bash
brew tap randlee/tap
brew install claude-history
```

## How It Works

1. **Repository**: GoReleaser automatically creates/updates `github.com/randlee/homebrew-tap`
2. **Formula**: Creates `Formula/claude-history.rb` with installation instructions
3. **Auto-update**: Each release updates the formula with new version/checksums

## Configuration

See `.goreleaser.yml` section:

```yaml
brews:
  - name: claude-history
    repository:
      owner: randlee
      name: homebrew-tap
```

## First-Time Setup

GoReleaser will **automatically create** the `homebrew-tap` repository on first release.

**No manual setup required!**

## User Installation

After the next release (v0.2.0+), users install with:

```bash
# Add the tap (one-time)
brew tap randlee/tap

# Install claude-history
brew install claude-history

# Update to latest
brew upgrade claude-history
```

## Verification

After the next release, verify:

1. Repository exists: https://github.com/randlee/homebrew-tap
2. Formula exists: `homebrew-tap/Formula/claude-history.rb`
3. Users can install: `brew tap randlee/tap && brew install claude-history`

## Notes

- The tap name is `randlee/tap` (short form of `randlee/homebrew-tap`)
- GoReleaser handles all formula updates automatically
- Works on both macOS and Linux
- Formula includes SHA256 checksums for security
