package paths

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/randlee/claude-history/pkg/encoding"
)

// mustMkdirAll creates directories or fails the test
func mustMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0750); err != nil {
		t.Fatalf("MkdirAll(%q) failed: %v", path, err)
	}
}

// mustWriteFile writes a file or fails the test
func mustWriteFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("WriteFile(%q) failed: %v", path, err)
	}
}

func TestDefaultClaudeDir(t *testing.T) {
	dir, err := DefaultClaudeDir()
	if err != nil {
		t.Fatalf("DefaultClaudeDir() error: %v", err)
	}

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".claude")

	if dir != expected {
		t.Errorf("DefaultClaudeDir() = %q, want %q", dir, expected)
	}
}

func TestProjectDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test directory to ensure cross-platform compatibility
	testPath := filepath.Join(tmpDir, "test-project")
	mustMkdirAll(t, testPath)

	dir, err := ProjectDir(tmpDir, testPath)
	if err != nil {
		t.Fatalf("ProjectDir() error: %v", err)
	}

	// Expected encoding depends on the absolute path format
	absTestPath, _ := filepath.Abs(testPath)
	expectedEncoded := encoding.EncodePath(absTestPath)
	expected := filepath.Join(tmpDir, "projects", expectedEncoded)

	if dir != expected {
		t.Errorf("ProjectDir() = %q, want %q", dir, expected)
	}
}

func TestProjectDirRelativePath(t *testing.T) {
	tmpDir := t.TempDir()

	// Test with "." (current directory)
	dir, err := ProjectDir(tmpDir, ".")
	if err != nil {
		t.Fatalf("ProjectDir() with '.' error: %v", err)
	}

	// Should resolve to absolute path and encode it
	// Expected path should start with tmpDir/projects/
	if !filepath.IsAbs(dir) {
		t.Error("ProjectDir() should return absolute path")
	}
	if !strings.HasPrefix(dir, filepath.Join(tmpDir, "projects")) {
		t.Errorf("ProjectDir() = %q, should start with %q", dir, filepath.Join(tmpDir, "projects"))
	}

	// Test with ".." (parent directory)
	dir, err = ProjectDir(tmpDir, "..")
	if err != nil {
		t.Fatalf("ProjectDir() with '..' error: %v", err)
	}

	if !filepath.IsAbs(dir) {
		t.Error("ProjectDir() should return absolute path for relative input")
	}
	if !strings.HasPrefix(dir, filepath.Join(tmpDir, "projects")) {
		t.Errorf("ProjectDir() = %q, should start with %q", dir, filepath.Join(tmpDir, "projects"))
	}

	// Test with relative path like "./subdir"
	dir, err = ProjectDir(tmpDir, "./subdir")
	if err != nil {
		t.Fatalf("ProjectDir() with './subdir' error: %v", err)
	}

	if !filepath.IsAbs(dir) {
		t.Error("ProjectDir() should return absolute path for relative input")
	}
}

func TestSessionFile(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "679761ba-80c0-4cd3-a586-cc6a1fc56308"

	file, err := SessionFile(tmpDir, "/Users/test/project", sessionID)
	if err != nil {
		t.Fatalf("SessionFile() error: %v", err)
	}

	if !filepath.IsAbs(file) {
		t.Error("SessionFile() should return absolute path")
	}

	if filepath.Ext(file) != ".jsonl" {
		t.Error("SessionFile() should have .jsonl extension")
	}

	if filepath.Base(file) != sessionID+".jsonl" {
		t.Errorf("SessionFile() filename = %q, want %q", filepath.Base(file), sessionID+".jsonl")
	}
}

func TestAgentFile(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "679761ba-80c0-4cd3-a586-cc6a1fc56308"
	agentID := "a12eb64"

	file, err := AgentFile(tmpDir, "/Users/test/project", sessionID, agentID)
	if err != nil {
		t.Fatalf("AgentFile() error: %v", err)
	}

	if !filepath.IsAbs(file) {
		t.Error("AgentFile() should return absolute path")
	}

	expectedName := "agent-" + agentID + ".jsonl"
	if filepath.Base(file) != expectedName {
		t.Errorf("AgentFile() filename = %q, want %q", filepath.Base(file), expectedName)
	}

	// Should be in subagents directory
	if filepath.Base(filepath.Dir(file)) != "subagents" {
		t.Error("AgentFile() should be in subagents directory")
	}
}

func TestListProjects(t *testing.T) {
	tmpDir := t.TempDir()
	projectsDir := filepath.Join(tmpDir, "projects")
	mustMkdirAll(t, projectsDir)

	// Create some project directories
	mustMkdirAll(t, filepath.Join(projectsDir, "-Users-test-project1"))
	mustMkdirAll(t, filepath.Join(projectsDir, "-Users-test-project2"))
	mustMkdirAll(t, filepath.Join(projectsDir, "not-encoded")) // Should be excluded

	projects, err := ListProjects(tmpDir)
	if err != nil {
		t.Fatalf("ListProjects() error: %v", err)
	}

	if len(projects) != 2 {
		t.Errorf("ListProjects() returned %d projects, want 2", len(projects))
	}

	if _, ok := projects["-Users-test-project1"]; !ok {
		t.Error("ListProjects() missing -Users-test-project1")
	}
}

func TestListSessionFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create session files
	mustWriteFile(t, filepath.Join(tmpDir, "679761ba-80c0-4cd3-a586-cc6a1fc56308.jsonl"), []byte("{}"))
	mustWriteFile(t, filepath.Join(tmpDir, "12345678-1234-1234-1234-123456789012.jsonl"), []byte("{}"))
	mustWriteFile(t, filepath.Join(tmpDir, "sessions-index.json"), []byte("{}")) // Should be excluded
	mustMkdirAll(t, filepath.Join(tmpDir, "some-dir"))                           // Should be excluded

	sessions, err := ListSessionFiles(tmpDir)
	if err != nil {
		t.Fatalf("ListSessionFiles() error: %v", err)
	}

	if len(sessions) != 2 {
		t.Errorf("ListSessionFiles() returned %d sessions, want 2", len(sessions))
	}
}

func TestListAgentFiles(t *testing.T) {
	tmpDir := t.TempDir()
	subagentsDir := filepath.Join(tmpDir, "subagents")
	mustMkdirAll(t, subagentsDir)

	// Create agent files
	mustWriteFile(t, filepath.Join(subagentsDir, "agent-a12eb64.jsonl"), []byte("{}"))
	mustWriteFile(t, filepath.Join(subagentsDir, "agent-aprompt_suggestion-abc.jsonl"), []byte("{}"))
	mustWriteFile(t, filepath.Join(subagentsDir, "other.jsonl"), []byte("{}")) // Should be excluded

	agents, err := ListAgentFiles(tmpDir)
	if err != nil {
		t.Fatalf("ListAgentFiles() error: %v", err)
	}

	if len(agents) != 2 {
		t.Errorf("ListAgentFiles() returned %d agents, want 2", len(agents))
	}

	if _, ok := agents["a12eb64"]; !ok {
		t.Error("ListAgentFiles() missing a12eb64")
	}
}

func TestLooksLikeUUID(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"679761ba-80c0-4cd3-a586-cc6a1fc56308", true},
		{"12345678-1234-1234-1234-123456789012", true},
		{"not-a-uuid", false},
		{"679761ba80c04cd3a586cc6a1fc56308", false}, // No dashes
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := looksLikeUUID(tt.input)
			if result != tt.expected {
				t.Errorf("looksLikeUUID(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
