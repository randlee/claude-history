package version_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/randlee/claude-history/pkg/version"
)

// TestVersionConsistency ensures all version references use the single source of truth.
func TestVersionConsistency(t *testing.T) {
	// Get the version constant
	expectedVersion := version.Version
	if expectedVersion == "" {
		t.Fatal("version.Version constant is empty")
	}

	// Find the repo root (go up from src/pkg/version to src/)
	repoRoot, err := filepath.Abs("../../")
	if err != nil {
		t.Fatalf("Failed to get repo root: %v", err)
	}

	// Files to check
	filesToCheck := map[string][]string{
		"main.go": {
			`version\.Version`, // Should import and use version.Version
		},
		"pkg/export/html.go": {
			`version\.Version`, // Should use version.Version in HTML generation
		},
	}

	// Check each file
	for filename, patterns := range filesToCheck {
		fullPath := filepath.Join(repoRoot, filename)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			t.Errorf("Failed to read %s: %v", filename, err)
			continue
		}

		fileContent := string(content)

		// Verify the file uses version.Version (not hardcoded version strings)
		for _, pattern := range patterns {
			re := regexp.MustCompile(pattern)
			if !re.MatchString(fileContent) {
				t.Errorf("%s does not reference version.Version (expected pattern: %s)", filename, pattern)
			}
		}

		// Check for hardcoded version strings that should be avoided
		// Allow version strings in comments, but not in code
		hardcodedVersionPattern := regexp.MustCompile(`(?m)^\s*(?://.*)?[^/\n]*["` + "`" + `]v?\d+\.\d+\.\d+`)
		lines := strings.Split(fileContent, "\n")
		for i, line := range lines {
			// Skip comments and import statements
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "//") ||
				strings.HasPrefix(trimmed, "/*") ||
				strings.Contains(line, "import") {
				continue
			}

			// Check for hardcoded version in code (not comments)
			if hardcodedVersionPattern.MatchString(line) {
				// Verify it's not just in a comment
				if !strings.Contains(line, "//") || strings.Index(line, "//") > strings.Index(line, expectedVersion) {
					t.Errorf("%s:%d contains hardcoded version string (should use version.Version): %s",
						filename, i+1, strings.TrimSpace(line))
				}
			}
		}
	}
}

// TestVersionFormat ensures the version follows semantic versioning.
func TestVersionFormat(t *testing.T) {
	v := version.Version

	// Check format: X.Y.Z or X.Y.Z-suffix
	semverPattern := regexp.MustCompile(`^\d+\.\d+\.\d+(-[a-zA-Z0-9]+)?$`)
	if !semverPattern.MatchString(v) {
		t.Errorf("version.Version %q does not follow semantic versioning format (X.Y.Z or X.Y.Z-suffix)", v)
	}
}
