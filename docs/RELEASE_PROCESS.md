# Release Process

This document describes the release process for claude-history, including how release notes are generated and what happens with each release.

## Release Status by Version

- **v0.1.0, v0.2.0**: Homebrew and winget publishing not configured. Manual installation only.
- **v0.3.0+**: Homebrew and winget publishing fully automated. All installation methods functional.

## Release Notes Template

Release notes are generated automatically by GoReleaser using the template at `.goreleaser/release-notes-template.md`.

### Template Contents

The template includes:
- **Installation instructions** for all distribution methods:
  - Homebrew (macOS/Linux)
  - winget (Windows)
  - Install script (macOS/Linux)
  - Go install (all platforms)
  - Direct download (all platforms)
- **Quick Start** examples
- **Documentation** links

### Why All Methods Are Listed

The template lists **all** installation methods (Homebrew, winget, install script, etc.) even if they're not immediately available for a given release. This is intentional:

1. **Consistency** - Users see the same format for every release
2. **Documentation** - Users know what installation options exist
3. **Forward-looking** - Methods may become available shortly after release
4. **Platform parity** - Shows commitment to supporting all platforms

### Installation Method Availability (v0.3.0+)

Different installation methods become available at different times:

| Method | Availability | Notes |
|--------|--------------|-------|
| **Direct Download** | Immediate | Binaries published to GitHub Releases by GoReleaser |
| **Go install** | Immediate | Published to GitHub, works via `go install` |
| **Install Script** | Immediate | Script downloads from GitHub Releases |
| **Homebrew** | ~5-10 minutes | GoReleaser automatically pushes to `randlee/homebrew-tap` |
| **winget** | ~1-2 days | Automated PR to microsoft/winget-pkgs, requires Microsoft approval |

> **Note:** v0.2.0 had issues with Homebrew publishing due to authentication. This was fixed in develop and will work automatically starting with v0.3.0.

## Release Checklist

### Prerequisites

- [ ] All tests passing (`go test ./...`)
- [ ] Linter clean (`golangci-lint run`)
- [ ] CHANGELOG.md updated with new version
- [ ] Version number follows [Semantic Versioning](https://semver.org/)
- [ ] GoReleaser configuration validated (`./scripts/validate-goreleaser.sh`)

### Creating a Release

1. **Validate the GoReleaser configuration:**
   ```bash
   ./scripts/validate-goreleaser.sh
   ```

   This validates:
   - Configuration syntax
   - Build configuration with a snapshot build
   - No deprecated or invalid fields
   - Required files exist

   **Important**: This validation runs automatically in CI on pushes to main/develop and on PRs,
   but running it locally before tagging helps catch errors early.

2. **Tag the release:**
   ```bash
   git tag -a v0.x.0 -m "Release v0.x.0"
   git push origin v0.x.0
   ```

3. **GoReleaser builds and publishes:**
   - Builds binaries for all platforms
   - Creates GitHub Release with release notes from template
   - Publishes to GitHub Releases
   - Updates Homebrew tap formula (automatic)
   - Creates winget PR to microsoft/winget-pkgs (automatic)

4. **Verify GitHub Release:**
   - Check https://github.com/randlee/claude-history/releases/latest
   - Verify all binaries are present in Assets
   - Verify release notes rendered correctly

### Post-Release Tasks

#### Homebrew (automatic - verify only)

After ~5-10 minutes:

- [ ] Verify formula updated: https://github.com/randlee/homebrew-tap/blob/main/Formula/claude-history.rb
- [ ] Test installation: `brew upgrade claude-history` or `brew install randlee/tap/claude-history`

If the formula didn't update, check the workflow run for errors.

#### winget (automatic - monitor only)

The workflow automatically creates a PR to microsoft/winget-pkgs using winget-releaser.

After ~1-2 days:

- [ ] Check PR status: https://github.com/microsoft/winget-pkgs/pulls?q=claude-history
- [ ] Monitor for validation failures (automated tests run on the PR)
- [ ] Once merged, verify: `winget search claude-history`

**Note**: The PR is created automatically but requires Microsoft maintainer approval. Validation failures are rare if the manifest is correct.

## Troubleshooting

### GoReleaser configuration errors

If the release workflow fails with a configuration error:

```
Error: .goreleaser.yml: unknown field: header_template_file
```

**Cause**: The `.goreleaser.yml` contains fields not supported by the GoReleaser version.

**Solution**:
1. Check the error message for the invalid field name
2. Consult [GoReleaser documentation](https://goreleaser.com/customization/) for the correct field name
3. Update `.goreleaser.yml` with the correct syntax
4. Run `./scripts/validate-goreleaser.sh` to verify the fix
5. Commit the fix and push

**Prevention**: Always run `./scripts/validate-goreleaser.sh` before creating a release tag.
The validation workflow also runs automatically on pushes to main/develop and on PRs.

### Homebrew publishing fails with 403 error

If you see:
```
homebrew tap formula: could not update "Formula/claude-history.rb": 403 Resource not accessible by integration
```

**Cause**: The `HOMEBREW_TAP_TOKEN` secret is missing or expired.

**Solution**:
1. Create a new fine-grained PAT at https://github.com/settings/personal-access-tokens/new
2. Grant it `Contents: Read and write` permission for `randlee/homebrew-tap` only
3. Add it as `HOMEBREW_TAP_TOKEN` secret at https://github.com/randlee/claude-history/settings/secrets/actions

### GitHub Release fails to create

Ensure the GitHub Actions workflow has proper permissions. The default `GITHUB_TOKEN` should have `contents: write` permission (configured in `.github/workflows/release.yml`).

### Release notes template not rendering

Verify the template path in `.goreleaser.yml`:
```yaml
release:
  header_template_file: .goreleaser/release-notes-template.md
```

## Optional: Pre-commit Hook

You can optionally install a pre-commit hook that automatically validates `.goreleaser.yml` before committing:

```bash
cp scripts/pre-commit-hook.sh .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

This hook will:
- Run only when `.goreleaser.yml` is being committed
- Execute `./scripts/validate-goreleaser.sh` automatically
- Prevent the commit if validation fails

To bypass the hook (not recommended):
```bash
git commit --no-verify
```

## Modifying the Template

To update the release notes format:

1. Edit `.goreleaser/release-notes-template.md`
2. Test with a local build:
   ```bash
   goreleaser release --snapshot --clean
   cat dist/release-notes.md  # Preview generated notes
   ```
3. Commit changes
4. Next release will use updated template

### Template Variables

GoReleaser provides these variables for use in the template:

- `{{.Version}}` - Version number (e.g., "0.2.0")
- `{{.Tag}}` - Git tag (e.g., "v0.2.0")
- `{{.ProjectName}}` - Project name ("claude-history")
- `{{.Date}}` - Release date
- `{{.Commit}}` - Git commit hash

See [GoReleaser documentation](https://goreleaser.com/customization/release/) for more template options.

## References

- [GoReleaser Release Customization](https://goreleaser.com/customization/release/)
- [Semantic Versioning](https://semver.org/)
- [Homebrew Tap Setup](./HOMEBREW_SETUP.md)
- [winget Setup](./WINGET_SETUP.md)
