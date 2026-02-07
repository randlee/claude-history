package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/randlee/claude-history/pkg/paths"
)

// TestExport_SessionResolution tests that export command can locate sessions found by list command.
// This reproduces the bug where export fails with "session not found: file does not exist"
// even though list command successfully finds the session.
func TestExport_SessionResolution(t *testing.T) {
	// Setup: Create mock .claude/projects directory with real filesystem project path
	tempDir := t.TempDir()
	mockClaudeDir := filepath.Join(tempDir, ".claude")

	// Create a mock filesystem project that will be encoded into Claude storage
	realProjectPath := filepath.Join(tempDir, "Users", "testuser", "my-project")
	if err := os.MkdirAll(realProjectPath, 0755); err != nil {
		t.Fatalf("failed to create mock project path: %v", err)
	}

	// Get the encoded project directory path
	projectDir, err := paths.ProjectDir(mockClaudeDir, realProjectPath)
	if err != nil {
		t.Fatalf("failed to get project directory: %v", err)
	}

	// Create the project directory in Claude's storage
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("failed to create project directory: %v", err)
	}

	// Create test session with full UUID
	sessionID := "c04f363a-5592-4c56-b4af-8d9533b5bca1"
	sessionFile := filepath.Join(projectDir, sessionID+".jsonl")

	// Create a minimal valid session JSONL with at least one user message
	sessionContent := `{"type":"user","timestamp":"2026-02-01T12:00:00Z","sessionId":"c04f363a-5592-4c56-b4af-8d9533b5bca1","uuid":"entry-1","message":"Test prompt"}
{"type":"assistant","timestamp":"2026-02-01T12:00:01Z","sessionId":"c04f363a-5592-4c56-b4af-8d9533b5bca1","uuid":"entry-2","message":"Test response"}
`

	// Write test session file
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0644); err != nil {
		t.Fatalf("failed to write test session file: %v", err)
	}

	// Verify session exists and can be listed (this should work)
	sessionFiles, err := paths.ListSessionFiles(projectDir)
	if err != nil {
		t.Fatalf("failed to list session files: %v", err)
	}

	if _, exists := sessionFiles[sessionID]; !exists {
		t.Fatalf("session %s not found in session files", sessionID)
	}

	// Test 1: Export with full session ID
	t.Run("export_with_full_session_id", func(t *testing.T) {
		// This should work - full session ID provided
		oldClaudeDir := claudeDir
		defer func() { claudeDir = oldClaudeDir }()
		claudeDir = mockClaudeDir

		// Create a temporary output directory
		outputDir := filepath.Join(tempDir, "export-output-full")

		// Mock the command arguments
		exportSessionID = sessionID
		exportOutputDir = outputDir
		exportFormat = "jsonl"

		err := runExport(nil, []string{realProjectPath})
		if err != nil {
			t.Fatalf("export with full session ID failed: %v", err)
		}

		// Verify output was created
		if _, err := os.Stat(outputDir); os.IsNotExist(err) {
			t.Fatalf("output directory was not created")
		}
	})

	// Test 2: Export with session ID prefix (reproduces the bug)
	t.Run("export_with_session_prefix", func(t *testing.T) {
		// This is the test case that reproduces the bug
		// When using a prefix like "c04f363a", export should resolve it like list does
		oldClaudeDir := claudeDir
		defer func() { claudeDir = oldClaudeDir }()
		claudeDir = mockClaudeDir

		// Create a temporary output directory
		outputDir := filepath.Join(tempDir, "export-output-prefix")

		// Use only first 8 characters of session ID (prefix)
		sessionIDPrefix := sessionID[:8]

		// Mock the command arguments
		exportSessionID = sessionIDPrefix
		exportOutputDir = outputDir
		exportFormat = "jsonl"

		// This should NOT fail - export should resolve the prefix like list command does
		err := runExport(nil, []string{realProjectPath})
		if err != nil {
			t.Fatalf("export with session ID prefix failed: %v", err)
		}

		// Verify output was created
		if _, err := os.Stat(outputDir); os.IsNotExist(err) {
			t.Fatalf("output directory was not created for prefix-resolved session")
		}
	})
}

// TestExport_SessionResolution_WithEncodedPath tests export works with encoded project paths.
func TestExport_SessionResolution_WithEncodedPath(t *testing.T) {
	// Setup: Create mock .claude/projects directory with real filesystem project path
	tempDir := t.TempDir()
	mockClaudeDir := filepath.Join(tempDir, ".claude")

	// Create a mock filesystem project that will be encoded into Claude storage
	realProjectPath := filepath.Join(tempDir, "Users", "anotheruser", "another-project")
	if err := os.MkdirAll(realProjectPath, 0755); err != nil {
		t.Fatalf("failed to create mock project path: %v", err)
	}

	// Get the encoded project directory path
	projectDir, err := paths.ProjectDir(mockClaudeDir, realProjectPath)
	if err != nil {
		t.Fatalf("failed to get project directory: %v", err)
	}

	// Create the project directory in Claude's storage
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("failed to create project directory: %v", err)
	}

	// Create test session with full UUID
	sessionID := "a1b2c3d4-5678-90ab-cdef-1234567890ab"
	sessionFile := filepath.Join(projectDir, sessionID+".jsonl")

	// Create a minimal valid session JSONL
	sessionContent := `{"type":"user","timestamp":"2026-02-01T14:00:00Z","sessionId":"a1b2c3d4-5678-90ab-cdef-1234567890ab","uuid":"entry-1","message":"Another test prompt"}
`

	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0644); err != nil {
		t.Fatalf("failed to write test session file: %v", err)
	}

	t.Run("export_with_prefix_and_encoded_path", func(t *testing.T) {
		oldClaudeDir := claudeDir
		defer func() { claudeDir = oldClaudeDir }()
		claudeDir = mockClaudeDir

		outputDir := filepath.Join(tempDir, "export-encoded-path")

		// Use prefix of session ID
		sessionIDPrefix := sessionID[:12]

		exportSessionID = sessionIDPrefix
		exportOutputDir = outputDir
		exportFormat = "jsonl"

		err := runExport(nil, []string{realProjectPath})
		if err != nil {
			t.Fatalf("export with prefix and encoded path failed: %v", err)
		}

		if _, err := os.Stat(outputDir); os.IsNotExist(err) {
			t.Fatalf("output directory was not created")
		}
	})
}
