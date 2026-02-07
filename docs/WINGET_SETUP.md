# Windows Package Manager (winget) Setup

This document explains how to set up and maintain winget support for `claude-history`.

## What is winget?

[Windows Package Manager (winget)](https://github.com/microsoft/winget-cli) is Microsoft's official package manager for Windows, similar to `apt` on Linux or `brew` on macOS. It allows Windows users to install, update, and manage software from the command line.

## User Installation

Once published to the winget repository, Windows users can install `claude-history` with a single command:

```powershell
winget install randlee.claude-history
```

To upgrade to the latest version:

```powershell
winget upgrade randlee.claude-history
```

## How winget Works

winget pulls package manifests from the [microsoft/winget-pkgs](https://github.com/microsoft/winget-pkgs) repository. Each package has YAML manifest files that describe:

- Package metadata (name, publisher, license)
- Download URLs for installers
- SHA256 checksums for security verification
- Installation instructions

**No account or API keys required** - submission is done via GitHub Pull Request.

## Publishing a New Version

### 1. After GitHub Release

When you create a new release (e.g., `v0.2.0`), GitHub Actions automatically:
- Builds Windows binaries via goreleaser
- Creates a ZIP file: `claude-history_0.2.0_windows_amd64.zip`
- Uploads it to the GitHub Release
- Generates `checksums.txt` with SHA256 hashes

### 2. Get the SHA256 Hash

Download `checksums.txt` from the GitHub Release page and find the hash for the Windows ZIP:

```bash
# Download checksums from release
curl -sL https://github.com/randlee/claude-history/releases/download/v0.2.0/checksums.txt | grep windows_amd64.zip
```

Example output:
```
a1b2c3d4e5f6... claude-history_0.2.0_windows_amd64.zip
```

### 3. Update the Manifest

Edit `.winget/randlee.claude-history.yaml`:

```yaml
PackageVersion: 0.2.0  # Update version
Installers:
  - Architecture: x64
    InstallerType: zip
    InstallerUrl: https://github.com/randlee/claude-history/releases/download/v0.2.0/claude-history_0.2.0_windows_amd64.zip
    InstallerSha256: a1b2c3d4e5f6...  # Replace with actual hash
```

### 4. Submit to microsoft/winget-pkgs

Fork and clone the [microsoft/winget-pkgs](https://github.com/microsoft/winget-pkgs) repository:

```bash
git clone https://github.com/YOUR_USERNAME/winget-pkgs.git
cd winget-pkgs
```

Create the package directory structure:

```bash
mkdir -p manifests/r/randlee/claude-history/0.2.0
```

Copy the manifest:

```bash
cp /path/to/claude-history/.winget/randlee.claude-history.yaml \
   manifests/r/randlee/claude-history/0.2.0/randlee.claude-history.installer.yaml
```

**Note**: The manifest filename must end with `.installer.yaml` in the winget-pkgs repository.

Commit and push:

```bash
git checkout -b randlee-claude-history-0.2.0
git add manifests/r/randlee/claude-history/0.2.0/
git commit -m "Add randlee.claude-history version 0.2.0"
git push origin randlee-claude-history-0.2.0
```

Create a Pull Request to `microsoft/winget-pkgs`:
- Title: `New version: randlee.claude-history version 0.2.0`
- Description: Link to the GitHub Release

### 5. Automated Validation

The winget-pkgs repository has automated CI that will:
- Validate YAML schema
- Verify the download URL is accessible
- Check the SHA256 hash matches
- Scan for malware

If validation passes, a maintainer will merge the PR (usually within 1-3 days).

## First-Time Submission

For the **initial submission** (v0.2.0), you may need to provide additional manifests:

**`randlee.claude-history.locale.en-US.yaml`**:
```yaml
PackageIdentifier: randlee.claude-history
PackageVersion: 0.2.0
PackageLocale: en-US
Publisher: Randy Lee
PublisherUrl: https://github.com/randlee
PackageName: claude-history
PackageUrl: https://github.com/randlee/claude-history
License: MIT
LicenseUrl: https://github.com/randlee/claude-history/blob/main/LICENSE
ShortDescription: CLI tool for programmatic access to Claude Code's agent history storage
Description: |
  claude-history maps between filesystem paths and Claude Code's internal storage format,
  enabling you to query conversation history, list projects and sessions, display agent
  hierarchy trees, find agents by task description, and export sessions to HTML or JSONL.
ManifestType: defaultLocale
ManifestVersion: 1.0.0
```

**`randlee.claude-history.yaml`** (top-level):
```yaml
PackageIdentifier: randlee.claude-history
PackageVersion: 0.2.0
DefaultLocale: en-US
ManifestType: version
ManifestVersion: 1.0.0
```

Place all three files in `manifests/r/randlee/claude-history/0.2.0/` directory.

## Updating for Subsequent Releases

For subsequent releases (v0.3.0, v0.4.0, etc.):

1. Create a new version directory: `manifests/r/randlee/claude-history/0.3.0/`
2. Copy and update all three manifest files with new version and SHA256
3. Submit PR as described above

## Validation Tools

Test your manifest locally before submitting:

```powershell
# Install winget manifest validation tool
winget install --id Microsoft.WingetCreate

# Validate manifest
wingetcreate validate /path/to/manifest.yaml
```

## Troubleshooting

### Hash Mismatch
If winget reports a hash mismatch:
- Re-download `checksums.txt` from the GitHub Release
- Verify you copied the correct hash for `windows_amd64.zip`
- Ensure you didn't accidentally modify the file

### Download URL Not Found
- Verify the release tag matches the URL (e.g., `v0.2.0`)
- Check that goreleaser successfully uploaded the Windows ZIP
- Ensure the release is public, not draft

### Manifest Validation Errors
- Use `wingetcreate validate` to check YAML syntax
- Ensure all required fields are present
- Check indentation (YAML is sensitive to spaces)

## Resources

- [winget-pkgs Contributing Guide](https://github.com/microsoft/winget-pkgs/blob/master/CONTRIBUTING.md)
- [Manifest Schema Documentation](https://github.com/microsoft/winget-cli/blob/master/doc/ManifestSpecv1.0.md)
- [winget CLI Documentation](https://docs.microsoft.com/en-us/windows/package-manager/winget/)

## Notes

- **No automation possible**: Unlike Homebrew (which can be automated via GitHub Actions), winget requires manual PR submission to microsoft/winget-pkgs
- **No account needed**: You only need a GitHub account to submit PRs
- **Review time**: PRs are typically reviewed and merged within 1-3 business days
- **Single architecture**: We only build `windows/amd64`, not `windows/arm64` (see `.goreleaser.yml`)

---

**Last Updated**: 2026-02-07
