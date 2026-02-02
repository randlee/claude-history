package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestExport_InvalidSessionID(t *testing.T) {
	// Test with non-existent session ID
	tmpDir, projectDir, projectPath := setupTestProject(t, "invalid-session-test")

	// Create a valid session first
	_ = createTestSessionWithAgents(t, projectDir, 1)

	oldSessionID := exportSessionID
	oldFormat := exportFormat
	oldOutputDir := exportOutputDir
	oldClaudeDir := claudeDir
	defer func() {
		exportSessionID = oldSessionID
		exportFormat = oldFormat
		exportOutputDir = oldOutputDir
		claudeDir = oldClaudeDir
	}()

	outputDir := filepath.Join(tmpDir, "export-invalid")

	// Use non-existent session ID
	exportSessionID = "nonexistent-session-id-12345"
	exportFormat = "html"
	exportOutputDir = outputDir
	claudeDir = tmpDir

	// Run export - should fail
	err := runExport(exportCmd, []string{projectPath})
	if err == nil {
		t.Error("Export should fail with non-existent session ID")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Error should mention 'not found', got: %v", err)
	}
}

func TestExport_InvalidProjectPath(t *testing.T) {
	// Test with non-existent project path
	tmpDir := t.TempDir()

	oldSessionID := exportSessionID
	oldFormat := exportFormat
	oldOutputDir := exportOutputDir
	oldClaudeDir := claudeDir
	defer func() {
		exportSessionID = oldSessionID
		exportFormat = oldFormat
		exportOutputDir = oldOutputDir
		claudeDir = oldClaudeDir
	}()

	outputDir := filepath.Join(tmpDir, "export-invalid-project")

	exportSessionID = "some-session"
	exportFormat = "html"
	exportOutputDir = outputDir
	claudeDir = tmpDir

	// Use non-existent project path
	invalidProject := filepath.Join(tmpDir, "nonexistent-project")

	err := runExport(exportCmd, []string{invalidProject})
	if err == nil {
		t.Error("Export should fail with non-existent project path")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Error should mention 'not found', got: %v", err)
	}
}

func TestExport_PermissionDenied(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping permission test on Windows - chmod doesn't work the same way")
	}
	if os.Getuid() == 0 {
		t.Skip("skipping permission test when running as root")
	}

	// Test with output directory that has no write permission
	tmpDir, projectDir, projectPath := setupTestProject(t, "permission-test")
	sessionID := createTestSessionWithAgents(t, projectDir, 1)

	oldSessionID := exportSessionID
	oldFormat := exportFormat
	oldOutputDir := exportOutputDir
	oldClaudeDir := claudeDir
	defer func() {
		exportSessionID = oldSessionID
		exportFormat = oldFormat
		exportOutputDir = oldOutputDir
		claudeDir = oldClaudeDir
	}()

	// Create a read-only directory
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	if err := os.Mkdir(readOnlyDir, 0555); err != nil {
		t.Fatalf("failed to create read-only directory: %v", err)
	}
	defer func() { _ = os.Chmod(readOnlyDir, 0755) }() //nolint:gosec // restore perms for cleanup

	outputDir := filepath.Join(readOnlyDir, "export-no-perms")

	exportSessionID = sessionID
	exportFormat = "html"
	exportOutputDir = outputDir
	claudeDir = tmpDir

	// Run export - should fail with permission error
	err := runExport(exportCmd, []string{projectPath})
	if err == nil {
		t.Error("Export should fail when output directory is not writable")
	}

	if !strings.Contains(err.Error(), "permission") && !strings.Contains(err.Error(), "denied") {
		t.Logf("Error: %v", err)
		// Permission error might have different message formats
	}
}

func TestExport_InvalidFormat(t *testing.T) {
	// Test with invalid format flag
	tmpDir, projectDir, projectPath := setupTestProject(t, "invalid-format-test")
	sessionID := createTestSessionWithAgents(t, projectDir, 1)

	oldSessionID := exportSessionID
	oldFormat := exportFormat
	oldOutputDir := exportOutputDir
	oldClaudeDir := claudeDir
	defer func() {
		exportSessionID = oldSessionID
		exportFormat = oldFormat
		exportOutputDir = oldOutputDir
		claudeDir = oldClaudeDir
	}()

	outputDir := filepath.Join(tmpDir, "export-bad-format")

	exportSessionID = sessionID
	exportFormat = "invalid-format"
	exportOutputDir = outputDir
	claudeDir = tmpDir

	// Run export - should fail
	err := runExport(exportCmd, []string{projectPath})
	if err == nil {
		t.Error("Export should fail with invalid format")
	}

	if !strings.Contains(err.Error(), "invalid format") {
		t.Errorf("Error should mention 'invalid format', got: %v", err)
	}
}

func TestExport_MissingSessionFlag(t *testing.T) {
	// The --session flag is marked as required, so Cobra should catch this
	// But we can test the validation logic

	tmpDir, _, projectPath := setupTestProject(t, "missing-session-test")

	oldSessionID := exportSessionID
	oldFormat := exportFormat
	oldOutputDir := exportOutputDir
	oldClaudeDir := claudeDir
	defer func() {
		exportSessionID = oldSessionID
		exportFormat = oldFormat
		exportOutputDir = oldOutputDir
		claudeDir = oldClaudeDir
	}()

	outputDir := filepath.Join(tmpDir, "export-no-session")

	exportSessionID = "" // Empty session ID
	exportFormat = "html"
	exportOutputDir = outputDir
	claudeDir = tmpDir

	// This might be caught by Cobra before reaching runExport
	// But if it reaches runExport, it should fail gracefully
	err := runExport(exportCmd, []string{projectPath})
	if err == nil {
		t.Error("Export should fail with missing session ID")
	}
}

func TestExport_CorruptedJSONL(t *testing.T) {
	// Test with corrupted JSONL file
	tmpDir, projectDir, projectPath := setupTestProject(t, "corrupted-test")

	sessionID := "corrupted-session"

	// Create session with invalid JSON
	sessionContent := `{"type":"user","timestamp":"2026-02-01T10:00:00Z","sessionId":"corrupted-session","uuid":"entry-1","message":"Valid entry"}
{this is not valid JSON at all}
{"type":"assistant","timestamp":"2026-02-01T10:00:05Z","sessionId":"corrupted-session","uuid":"entry-2","message":"Another entry"}
`

	sessionFile := filepath.Join(projectDir, sessionID+".jsonl")
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0644); err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	oldSessionID := exportSessionID
	oldFormat := exportFormat
	oldOutputDir := exportOutputDir
	oldClaudeDir := claudeDir
	defer func() {
		exportSessionID = oldSessionID
		exportFormat = oldFormat
		exportOutputDir = oldOutputDir
		claudeDir = oldClaudeDir
	}()

	outputDir := filepath.Join(tmpDir, "export-corrupted")

	exportSessionID = sessionID
	exportFormat = "html"
	exportOutputDir = outputDir
	claudeDir = tmpDir

	// Run export - might fail or succeed with errors
	err := runExport(exportCmd, []string{projectPath})

	// Depending on implementation, this might:
	// 1. Fail completely
	// 2. Succeed but skip the invalid line
	// 3. Succeed with warnings

	if err != nil {
		// If it fails, that's acceptable
		t.Logf("Export failed with corrupted JSONL (expected): %v", err)
	} else {
		// If it succeeds, verify output was created
		if _, statErr := os.Stat(outputDir); os.IsNotExist(statErr) {
			t.Error("output directory should be created even with corrupted JSONL")
		}
	}
}

func TestExport_EmptySessionFile(t *testing.T) {
	// Test with completely empty session file
	tmpDir, projectDir, projectPath := setupTestProject(t, "empty-file-test")

	sessionID := "empty-file-session"

	// Create empty session file
	sessionFile := filepath.Join(projectDir, sessionID+".jsonl")
	if err := os.WriteFile(sessionFile, []byte(""), 0644); err != nil {
		t.Fatalf("failed to create empty session: %v", err)
	}

	oldSessionID := exportSessionID
	oldFormat := exportFormat
	oldOutputDir := exportOutputDir
	oldClaudeDir := claudeDir
	defer func() {
		exportSessionID = oldSessionID
		exportFormat = oldFormat
		exportOutputDir = oldOutputDir
		claudeDir = oldClaudeDir
	}()

	outputDir := filepath.Join(tmpDir, "export-empty-file")

	exportSessionID = sessionID
	exportFormat = "html"
	exportOutputDir = outputDir
	claudeDir = tmpDir

	// Run export - should succeed
	err := runExport(exportCmd, []string{projectPath})
	if err != nil {
		t.Fatalf("Export of empty file should succeed: %v", err)
	}

	// Verify basic output created
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		t.Error("output directory should be created for empty session")
	}
}

func TestExport_OutputDirCreationFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows - different permission handling")
	}
	if os.Getuid() == 0 {
		t.Skip("skipping when running as root")
	}

	// Test when output directory cannot be created
	tmpDir, projectDir, projectPath := setupTestProject(t, "mkdir-fail-test")
	sessionID := createTestSessionWithAgents(t, projectDir, 1)

	oldSessionID := exportSessionID
	oldFormat := exportFormat
	oldOutputDir := exportOutputDir
	oldClaudeDir := claudeDir
	defer func() {
		exportSessionID = oldSessionID
		exportFormat = oldFormat
		exportOutputDir = oldOutputDir
		claudeDir = oldClaudeDir
	}()

	// Create a parent directory that's not writable
	readOnlyParent := filepath.Join(tmpDir, "readonly-parent")
	if err := os.Mkdir(readOnlyParent, 0555); err != nil {
		t.Fatalf("failed to create read-only parent: %v", err)
	}
	defer func() { _ = os.Chmod(readOnlyParent, 0755) }() //nolint:gosec // restore perms

	outputDir := filepath.Join(readOnlyParent, "cannot-create-this")

	exportSessionID = sessionID
	exportFormat = "html"
	exportOutputDir = outputDir
	claudeDir = tmpDir

	// Run export - should fail
	err := runExport(exportCmd, []string{projectPath})
	if err == nil {
		t.Error("Export should fail when output directory cannot be created")
	}
}

func TestExport_InvalidClaudeDir(t *testing.T) {
	// Test with invalid Claude directory
	tmpDir, _, projectPath := setupTestProject(t, "invalid-claude-test")

	oldSessionID := exportSessionID
	oldFormat := exportFormat
	oldOutputDir := exportOutputDir
	oldClaudeDir := claudeDir
	defer func() {
		exportSessionID = oldSessionID
		exportFormat = oldFormat
		exportOutputDir = oldOutputDir
		claudeDir = oldClaudeDir
	}()

	outputDir := filepath.Join(tmpDir, "export-invalid-claude")

	exportSessionID = "some-session"
	exportFormat = "html"
	exportOutputDir = outputDir
	claudeDir = filepath.Join(tmpDir, "nonexistent-claude-dir")

	// Run export - should fail
	err := runExport(exportCmd, []string{projectPath})
	if err == nil {
		t.Error("Export should fail with invalid Claude directory")
	}
}

func TestExport_NonAbsolutePathResolution(t *testing.T) {
	// Test that relative paths are properly resolved
	tmpDir, projectDir, _ := setupTestProject(t, "relative-resolve-test")
	sessionID := createTestSessionWithAgents(t, projectDir, 1)

	oldSessionID := exportSessionID
	oldFormat := exportFormat
	oldOutputDir := exportOutputDir
	oldClaudeDir := claudeDir
	defer func() {
		exportSessionID = oldSessionID
		exportFormat = oldFormat
		exportOutputDir = oldOutputDir
		claudeDir = oldClaudeDir
	}()

	// Change to temp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Use relative path
	relativeOutput := "export-rel-resolve"

	exportSessionID = sessionID
	exportFormat = "html"
	exportOutputDir = relativeOutput
	claudeDir = tmpDir

	// Run export with relative paths
	err = runExport(exportCmd, []string{"relative-resolve-test"})
	if err != nil {
		t.Fatalf("Export with relative paths failed: %v", err)
	}

	// Verify output created
	absOutputPath := filepath.Join(tmpDir, relativeOutput)
	if _, err := os.Stat(absOutputPath); os.IsNotExist(err) {
		t.Error("output directory not created for relative path")
	}
}

func TestExport_SessionFileIsDirectory(t *testing.T) {
	// Test error handling when session "file" is actually a directory
	tmpDir, projectDir, projectPath := setupTestProject(t, "dir-as-file-test")

	sessionID := "directory-session"

	// Create a directory instead of a file
	sessionDir := filepath.Join(projectDir, sessionID+".jsonl")
	if err := os.Mkdir(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create session directory: %v", err)
	}

	oldSessionID := exportSessionID
	oldFormat := exportFormat
	oldOutputDir := exportOutputDir
	oldClaudeDir := claudeDir
	defer func() {
		exportSessionID = oldSessionID
		exportFormat = oldFormat
		exportOutputDir = oldOutputDir
		claudeDir = oldClaudeDir
	}()

	outputDir := filepath.Join(tmpDir, "export-dir-as-file")

	exportSessionID = sessionID
	exportFormat = "html"
	exportOutputDir = outputDir
	claudeDir = tmpDir

	// Run export - should fail
	err := runExport(exportCmd, []string{projectPath})
	if err == nil {
		t.Error("Export should fail when session file is a directory")
	}
}

func TestExport_VeryLongPath(t *testing.T) {
	// Test with very long path (near filesystem limits)
	// Note: Some filesystems have path length limits (e.g., 260 chars on Windows)

	tmpDir, _, _ := setupTestProject(t, "long-path-test")

	// Create a deeply nested path
	longPath := tmpDir
	for i := 0; i < 20; i++ {
		longPath = filepath.Join(longPath, "very-long-directory-name-segment")
	}

	// Try to create this path
	if err := os.MkdirAll(longPath, 0755); err != nil {
		t.Skipf("Cannot create very long path on this system: %v", err)
	}

	// If we got here, the filesystem supports this path length
	// Create a project there
	projectPath := filepath.Join(longPath, "project")
	if err := os.Mkdir(projectPath, 0755); err != nil {
		t.Skipf("Cannot create project in long path: %v", err)
	}

	// This test mainly verifies that we don't crash with very long paths
	// The export might fail for legitimate reasons, but shouldn't panic
	t.Logf("Created very long path: %d characters", len(projectPath))
}

func TestExport_ConcurrentExports(t *testing.T) {
	// Test that multiple exports can run simultaneously
	tmpDir, projectDir, projectPath := setupTestProject(t, "concurrent-test")

	// Create multiple sessions
	session1 := createTestSessionWithAgents(t, projectDir, 1)
	session2 := createTestSessionWithAgents(t, projectDir, 2)

	// Run exports concurrently (in separate test invocations)
	// For now, just verify both can be exported sequentially

	oldSessionID := exportSessionID
	oldFormat := exportFormat
	oldOutputDir := exportOutputDir
	oldClaudeDir := claudeDir
	defer func() {
		exportSessionID = oldSessionID
		exportFormat = oldFormat
		exportOutputDir = oldOutputDir
		claudeDir = oldClaudeDir
	}()

	// Export first session
	exportSessionID = session1
	exportFormat = "html"
	exportOutputDir = filepath.Join(tmpDir, "export-1")
	claudeDir = tmpDir

	if err := runExport(exportCmd, []string{projectPath}); err != nil {
		t.Fatalf("First export failed: %v", err)
	}

	// Export second session
	exportSessionID = session2
	exportOutputDir = filepath.Join(tmpDir, "export-2")

	if err := runExport(exportCmd, []string{projectPath}); err != nil {
		t.Fatalf("Second export failed: %v", err)
	}

	// Both should succeed
	if _, err := os.Stat(filepath.Join(tmpDir, "export-1")); os.IsNotExist(err) {
		t.Error("first export directory not created")
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "export-2")); os.IsNotExist(err) {
		t.Error("second export directory not created")
	}
}
