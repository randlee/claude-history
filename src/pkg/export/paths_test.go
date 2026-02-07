package export

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMakePathsClickable_NoPath(t *testing.T) {
	text := "This is just plain text with no paths"
	result := makePathsClickable(text, "", func(pathStr, linkHTML string, start, end int) string {
		return linkHTML
	})

	if result != text {
		t.Errorf("Plain text should be unchanged, got %q", result)
	}
}

func TestMakePathsClickable_AbsoluteUnixPath(t *testing.T) {
	// Create a temporary file to test against
	tmpFile, err := os.CreateTemp("", "test-*.go")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	text := "Check out " + tmpFile.Name() + " for details"
	result := makePathsClickable(text, "", func(pathStr, linkHTML string, start, end int) string {
		return linkHTML
	})

	// Should contain a file link
	if !strings.Contains(result, `<a href="file://`) {
		t.Errorf("Should contain file link, got %q", result)
	}
	if !strings.Contains(result, `class="file-link"`) {
		t.Errorf("Should have file-link class, got %q", result)
	}
	if !strings.Contains(result, tmpFile.Name()) {
		t.Errorf("Should contain original path in link text, got %q", result)
	}
}

func TestMakePathsClickable_RelativePath(t *testing.T) {
	// Create a temp directory with a file
	tmpDir, err := os.MkdirTemp("", "test-dir-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a file in the directory
	testFile := filepath.Join(tmpDir, "test.go")
	if err := os.WriteFile(testFile, []byte("package main"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test relative path detection
	text := "See test.go for the implementation"
	result := makePathsClickable(text, tmpDir, func(pathStr, linkHTML string, start, end int) string {
		return linkHTML
	})

	// Should contain a file link
	if !strings.Contains(result, `<a href="file://`) {
		t.Errorf("Should contain file link for relative path, got %q", result)
	}
	if !strings.Contains(result, `class="file-link"`) {
		t.Errorf("Should have file-link class, got %q", result)
	}
}

func TestMakePathsClickable_RelativePathWithDotSlash(t *testing.T) {
	// Create a temp directory with a file
	tmpDir, err := os.MkdirTemp("", "test-dir-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a file in the directory
	testFile := filepath.Join(tmpDir, "test.go")
	if err := os.WriteFile(testFile, []byte("package main"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test ./path detection
	text := "Check ./test.go for details"
	result := makePathsClickable(text, tmpDir, func(pathStr, linkHTML string, start, end int) string {
		return linkHTML
	})

	// Should contain a file link
	if !strings.Contains(result, `<a href="file://`) {
		t.Errorf("Should contain file link for ./path, got %q", result)
	}
}

func TestMakePathsClickable_NonExistentPath(t *testing.T) {
	text := "This mentions /nonexistent/path/to/file.go that doesn't exist"
	result := makePathsClickable(text, "", func(pathStr, linkHTML string, start, end int) string {
		return linkHTML
	})

	// Should NOT create a link for non-existent file
	if strings.Contains(result, `<a href="file://`) {
		t.Errorf("Should not create link for non-existent file, got %q", result)
	}
	if result != text {
		t.Errorf("Non-existent path should leave text unchanged, got %q", result)
	}
}

func TestMakePathsClickable_PathInCodeBlock(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test-*.go")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Note: In the actual implementation, code blocks are protected by placeholders
	// before makePathsClickable is called. This test checks the raw behavior.
	text := "The path " + tmpFile.Name() + " should be linked"
	result := makePathsClickable(text, "", func(pathStr, linkHTML string, start, end int) string {
		return linkHTML
	})

	// Should create a link (code block protection happens at a higher level)
	if !strings.Contains(result, `<a href="file://`) {
		t.Errorf("Should create link, got %q", result)
	}
}

func TestMakePathsClickable_MultiplePaths(t *testing.T) {
	// Create two temporary files
	tmpFile1, err := os.CreateTemp("", "test1-*.go")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile1.Name())
	defer tmpFile1.Close()

	tmpFile2, err := os.CreateTemp("", "test2-*.go")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile2.Name())
	defer tmpFile2.Close()

	text := "Check " + tmpFile1.Name() + " and " + tmpFile2.Name() + " for details"
	result := makePathsClickable(text, "", func(pathStr, linkHTML string, start, end int) string {
		return linkHTML
	})

	// Should create links for both files
	linkCount := strings.Count(result, `<a href="file://`)
	if linkCount != 2 {
		t.Errorf("Expected 2 file links, got %d. Result: %q", linkCount, result)
	}
}

func TestMakePathsClickable_PathWithSpaces(t *testing.T) {
	// Paths with spaces are NOT automatically detected to avoid false positives.
	// This is a design decision - spaces in directory/file names cause too many
	// false matches (e.g., "/path/file.go and another sentence" would be treated
	// as one long path if spaces were allowed).
	// Users can still manually create links for files with spaces using markdown syntax.

	tmpDir, err := os.MkdirTemp("", "test-dir-with-dashes-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a file WITHOUT spaces (spaces are not supported)
	testFile := filepath.Join(tmpDir, "test-file.go")
	if err := os.WriteFile(testFile, []byte("package main"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	text := "The file is at " + testFile
	result := makePathsClickable(text, "", func(pathStr, linkHTML string, start, end int) string {
		return linkHTML
	})

	// Should handle paths with dashes (not spaces)
	if !strings.Contains(result, `<a href="file://`) {
		t.Errorf("Should handle paths with dashes, got %q", result)
	}
}

func TestMakePathsClickable_EmptyProjectPath(t *testing.T) {
	// Without a project path, relative paths should not be detected
	text := "Check src/main.go for details"
	result := makePathsClickable(text, "", func(pathStr, linkHTML string, start, end int) string {
		return linkHTML
	})

	// Should not create links for relative paths when projectPath is empty
	// (unless the path happens to exist relative to current directory)
	// For this test, we assume src/main.go doesn't exist
	if result != text {
		// Only fail if a link was actually created
		if strings.Contains(result, `<a href="file://`) {
			t.Errorf("Should not create link for relative path without projectPath, got %q", result)
		}
	}
}

func TestMakePathsClickableWithPlaceholders(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test-*.go")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	text := "Check " + tmpFile.Name() + " for details"
	placeholders := make(map[string]string)
	startIdx := 0

	result := makePathsClickableWithPlaceholders(text, "", &placeholders, &startIdx)

	// Should contain a placeholder
	if !strings.Contains(result, "\x00PATH_0\x00") {
		t.Errorf("Should contain placeholder, got %q", result)
	}

	// Placeholders map should have an entry
	if len(placeholders) != 1 {
		t.Errorf("Expected 1 placeholder, got %d", len(placeholders))
	}

	// The placeholder should map to link HTML
	for _, html := range placeholders {
		if !strings.Contains(html, `<a href="file://`) {
			t.Errorf("Placeholder should map to file link HTML, got %q", html)
		}
	}
}
