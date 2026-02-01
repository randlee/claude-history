package export

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Helper to create a test session structure
func setupTestSession(t *testing.T, baseDir string) (projectDir string, sessionID string) {
	t.Helper()

	sessionID = "12345678-1234-1234-1234-123456789abc"

	// Create the project directory structure that matches Claude's storage
	// ~/.claude/projects/{encoded-path}/{sessionId}.jsonl
	projectDir = filepath.Join(baseDir, "projects", "-test-project")
	sessionDir := filepath.Join(projectDir, sessionID)
	subagentsDir := filepath.Join(sessionDir, "subagents")

	if err := os.MkdirAll(subagentsDir, 0755); err != nil {
		t.Fatalf("failed to create test directories: %v", err)
	}

	// Create main session file
	sessionContent := `{"type":"user","timestamp":"2026-02-01T10:00:00Z","sessionId":"12345678-1234-1234-1234-123456789abc","uuid":"entry-1"}
{"type":"assistant","timestamp":"2026-02-01T10:01:00Z","sessionId":"12345678-1234-1234-1234-123456789abc","uuid":"entry-2"}
`
	sessionFile := filepath.Join(projectDir, sessionID+".jsonl")
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0644); err != nil {
		t.Fatalf("failed to create session file: %v", err)
	}

	// Create agent file
	agentContent := `{"type":"user","timestamp":"2026-02-01T10:02:00Z","sessionId":"12345678-1234-1234-1234-123456789abc","uuid":"agent-entry-1"}
{"type":"assistant","timestamp":"2026-02-01T10:03:00Z","sessionId":"12345678-1234-1234-1234-123456789abc","uuid":"agent-entry-2"}
`
	agentFile := filepath.Join(subagentsDir, "agent-a1b2c3d4.jsonl")
	if err := os.WriteFile(agentFile, []byte(agentContent), 0644); err != nil {
		t.Fatalf("failed to create agent file: %v", err)
	}

	return projectDir, sessionID
}

// Helper to create nested agent structure
func setupNestedAgents(t *testing.T, projectDir, sessionID string) {
	t.Helper()

	// Create nested agent structure
	// agent-parent -> subagents -> agent-child.jsonl
	sessionDir := filepath.Join(projectDir, sessionID)
	parentAgentDir := filepath.Join(sessionDir, "subagents", "agent-parent123")
	nestedSubagentsDir := filepath.Join(parentAgentDir, "subagents")

	if err := os.MkdirAll(nestedSubagentsDir, 0755); err != nil {
		t.Fatalf("failed to create nested directories: %v", err)
	}

	// Create parent agent file
	parentContent := `{"type":"user","timestamp":"2026-02-01T11:00:00Z","uuid":"parent-entry-1"}
`
	parentFile := filepath.Join(sessionDir, "subagents", "agent-parent123.jsonl")
	if err := os.WriteFile(parentFile, []byte(parentContent), 0644); err != nil {
		t.Fatalf("failed to create parent agent file: %v", err)
	}

	// Create nested child agent file
	childContent := `{"type":"user","timestamp":"2026-02-01T11:01:00Z","uuid":"child-entry-1"}
`
	childFile := filepath.Join(nestedSubagentsDir, "agent-child456.jsonl")
	if err := os.WriteFile(childFile, []byte(childContent), 0644); err != nil {
		t.Fatalf("failed to create child agent file: %v", err)
	}
}

func TestGenerateTempPath(t *testing.T) {
	tests := []struct {
		name         string
		sessionID    string
		lastModified time.Time
		wantPrefix   string
		wantContains string
	}{
		{
			name:         "standard UUID session ID",
			sessionID:    "12345678-1234-1234-1234-123456789abc",
			lastModified: time.Date(2026, 2, 1, 19, 0, 22, 0, time.UTC),
			wantPrefix:   "12345678",
			wantContains: "2026-02-01T19-00-22",
		},
		{
			name:         "short session ID",
			sessionID:    "abc",
			lastModified: time.Date(2026, 1, 15, 10, 30, 45, 0, time.UTC),
			wantPrefix:   "abc",
			wantContains: "2026-01-15T10-30-45",
		},
		{
			name:         "exactly 8 char session ID",
			sessionID:    "abcdefgh",
			lastModified: time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC),
			wantPrefix:   "abcdefgh",
			wantContains: "2026-12-31T23-59-59",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generateTempPath(tt.sessionID, tt.lastModified)
			if err != nil {
				t.Fatalf("generateTempPath() error = %v", err)
			}

			// Check it's under temp directory
			if !strings.HasPrefix(got, os.TempDir()) {
				t.Errorf("generateTempPath() = %v, should be under temp dir %v", got, os.TempDir())
			}

			// Check it contains claude-history
			if !strings.Contains(got, "claude-history") {
				t.Errorf("generateTempPath() = %v, should contain 'claude-history'", got)
			}

			// Check prefix is correct
			baseName := filepath.Base(got)
			if !strings.HasPrefix(baseName, tt.wantPrefix+"-") {
				t.Errorf("generateTempPath() base = %v, should start with %v-", baseName, tt.wantPrefix)
			}

			// Check timestamp is in the path
			if !strings.Contains(got, tt.wantContains) {
				t.Errorf("generateTempPath() = %v, should contain %v", got, tt.wantContains)
			}

			// Check path is filesystem safe (no colons in timestamp)
			if strings.Count(baseName, ":") > 0 {
				t.Errorf("generateTempPath() base = %v, should not contain colons", baseName)
			}
		})
	}
}

func TestGenerateTempPath_CrossPlatform(t *testing.T) {
	sessionID := "test-session-id"
	lastModified := time.Date(2026, 2, 1, 14, 30, 0, 0, time.UTC)

	path, err := generateTempPath(sessionID, lastModified)
	if err != nil {
		t.Fatalf("generateTempPath() error = %v", err)
	}

	// Verify path uses filepath.Join (no hardcoded slashes)
	// On all platforms, filepath.Base should work correctly
	base := filepath.Base(path)
	if base == "" {
		t.Error("generateTempPath() returned empty base name")
	}

	// Path should be valid (no invalid characters)
	if strings.ContainsAny(base, ":<>|?*\"") {
		t.Errorf("generateTempPath() base = %v, contains invalid path characters", base)
	}
}

func TestCopyFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create source file
	srcContent := []byte("test content for copy")
	srcPath := filepath.Join(tempDir, "source.txt")
	if err := os.WriteFile(srcPath, srcContent, 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	t.Run("copy to existing directory", func(t *testing.T) {
		dstPath := filepath.Join(tempDir, "dest.txt")
		if err := copyFile(srcPath, dstPath); err != nil {
			t.Fatalf("copyFile() error = %v", err)
		}

		dstContent, err := os.ReadFile(dstPath)
		if err != nil {
			t.Fatalf("failed to read destination file: %v", err)
		}

		if string(dstContent) != string(srcContent) {
			t.Errorf("copyFile() content = %v, want %v", string(dstContent), string(srcContent))
		}
	})

	t.Run("copy to nested directory (creates parents)", func(t *testing.T) {
		dstPath := filepath.Join(tempDir, "nested", "deep", "dest.txt")
		if err := copyFile(srcPath, dstPath); err != nil {
			t.Fatalf("copyFile() error = %v", err)
		}

		if _, err := os.Stat(dstPath); os.IsNotExist(err) {
			t.Error("copyFile() did not create destination file")
		}
	})

	t.Run("source does not exist", func(t *testing.T) {
		dstPath := filepath.Join(tempDir, "nonexistent-dest.txt")
		err := copyFile(filepath.Join(tempDir, "nonexistent.txt"), dstPath)
		if err == nil {
			t.Error("copyFile() expected error for nonexistent source")
		}
	})
}

func TestCopyFile_LargeFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create a larger file (1MB)
	size := 1024 * 1024
	srcContent := make([]byte, size)
	for i := range srcContent {
		srcContent[i] = byte(i % 256)
	}

	srcPath := filepath.Join(tempDir, "large.bin")
	if err := os.WriteFile(srcPath, srcContent, 0644); err != nil {
		t.Fatalf("failed to create large source file: %v", err)
	}

	dstPath := filepath.Join(tempDir, "large-copy.bin")
	if err := copyFile(srcPath, dstPath); err != nil {
		t.Fatalf("copyFile() error = %v", err)
	}

	dstContent, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("failed to read destination file: %v", err)
	}

	if len(dstContent) != size {
		t.Errorf("copyFile() size = %d, want %d", len(dstContent), size)
	}
}

func TestCopyFile_PermissionError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("skipping permission test when running as root")
	}

	tempDir := t.TempDir()

	// Create source file
	srcPath := filepath.Join(tempDir, "source.txt")
	if err := os.WriteFile(srcPath, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	// Create a read-only directory
	readOnlyDir := filepath.Join(tempDir, "readonly")
	if err := os.Mkdir(readOnlyDir, 0555); err != nil {
		t.Fatalf("failed to create read-only directory: %v", err)
	}
	defer os.Chmod(readOnlyDir, 0755) // Cleanup

	dstPath := filepath.Join(readOnlyDir, "dest.txt")
	err := copyFile(srcPath, dstPath)
	if err == nil {
		t.Error("copyFile() expected error for read-only destination directory")
	}
}

func TestExportSession_Success(t *testing.T) {
	tempDir := t.TempDir()
	_, sessionID := setupTestSession(t, tempDir)

	// The project path is what would be encoded to get this projectDir
	// Since our test setup creates -test-project, we need to provide a matching path
	outputDir := filepath.Join(tempDir, "export-output")

	opts := ExportOptions{
		OutputDir: outputDir,
		ClaudeDir: tempDir, // Point to our test "home" directory
	}

	// We need to export using a path that encodes to "-test-project"
	// The encoding replaces "/" with "-" and "." with "-"
	// So "/test/project" would encode to "-test-project"
	// But our test just needs to use the existing projectDir structure

	result, err := ExportSession("/test/project", sessionID, opts)
	if err != nil {
		t.Fatalf("ExportSession() error = %v", err)
	}

	// Verify result fields
	if result.OutputDir != outputDir {
		t.Errorf("ExportSession() OutputDir = %v, want %v", result.OutputDir, outputDir)
	}

	if result.SessionID != sessionID {
		t.Errorf("ExportSession() SessionID = %v, want %v", result.SessionID, sessionID)
	}

	// Check source directory was created
	if _, err := os.Stat(result.SourceDir); os.IsNotExist(err) {
		t.Error("ExportSession() did not create source directory")
	}

	// Check main session file was copied
	if _, err := os.Stat(result.MainSessionFile); os.IsNotExist(err) {
		t.Error("ExportSession() did not copy main session file")
	}

	// Verify session file contents
	content, err := os.ReadFile(result.MainSessionFile)
	if err != nil {
		t.Fatalf("failed to read copied session file: %v", err)
	}
	if !strings.Contains(string(content), "entry-1") {
		t.Error("ExportSession() session file content mismatch")
	}

	// Check agent file was copied
	if result.TotalAgents != 1 {
		t.Errorf("ExportSession() TotalAgents = %d, want 1", result.TotalAgents)
	}

	if _, ok := result.AgentFiles["a1b2c3d4"]; !ok {
		t.Error("ExportSession() did not record agent file")
	}

	// Verify agent file exists
	for agentID, agentPath := range result.AgentFiles {
		if _, err := os.Stat(agentPath); os.IsNotExist(err) {
			t.Errorf("ExportSession() agent file %s does not exist at %s", agentID, agentPath)
		}
	}
}

func TestExportSession_WithTempDir(t *testing.T) {
	tempDir := t.TempDir()
	setupTestSession(t, tempDir)
	sessionID := "12345678-1234-1234-1234-123456789abc"

	opts := ExportOptions{
		// OutputDir is empty - should generate temp path
		ClaudeDir: tempDir,
	}

	result, err := ExportSession("/test/project", sessionID, opts)
	if err != nil {
		t.Fatalf("ExportSession() error = %v", err)
	}
	defer CleanupExport(result.OutputDir)

	// Verify output dir is under temp
	if !strings.HasPrefix(result.OutputDir, os.TempDir()) {
		t.Errorf("ExportSession() OutputDir = %v, should be under temp", result.OutputDir)
	}

	// Verify it contains claude-history
	if !strings.Contains(result.OutputDir, "claude-history") {
		t.Errorf("ExportSession() OutputDir = %v, should contain 'claude-history'", result.OutputDir)
	}

	// Verify session ID prefix is in the path
	if !strings.Contains(result.OutputDir, sessionID[:8]) {
		t.Errorf("ExportSession() OutputDir = %v, should contain session prefix %s", result.OutputDir, sessionID[:8])
	}
}

func TestExportSession_NestedAgents(t *testing.T) {
	tempDir := t.TempDir()
	projectDir, sessionID := setupTestSession(t, tempDir)
	setupNestedAgents(t, projectDir, sessionID)

	outputDir := filepath.Join(tempDir, "export-nested")

	opts := ExportOptions{
		OutputDir: outputDir,
		ClaudeDir: tempDir,
	}

	result, err := ExportSession("/test/project", sessionID, opts)
	if err != nil {
		t.Fatalf("ExportSession() error = %v", err)
	}

	// Should have 3 agents: a1b2c3d4, parent123, child456
	if result.TotalAgents != 3 {
		t.Errorf("ExportSession() TotalAgents = %d, want 3", result.TotalAgents)
	}

	// Check parent agent was copied
	if _, ok := result.AgentFiles["parent123"]; !ok {
		t.Error("ExportSession() did not copy parent agent")
	}

	// Check nested child agent was copied
	if _, ok := result.AgentFiles["child456"]; !ok {
		t.Error("ExportSession() did not copy nested child agent")
	}

	// Verify child is in the nested directory structure
	childPath := result.AgentFiles["child456"]
	if !strings.Contains(childPath, "subagents") {
		t.Errorf("ExportSession() child agent path = %v, should contain 'subagents'", childPath)
	}
}

func TestExportSession_SessionNotFound(t *testing.T) {
	tempDir := t.TempDir()

	// Create minimal directory structure without session file
	projectDir := filepath.Join(tempDir, "projects", "-test-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}

	opts := ExportOptions{
		OutputDir: filepath.Join(tempDir, "output"),
		ClaudeDir: tempDir,
	}

	_, err := ExportSession("/test/project", "nonexistent-session", opts)
	if err == nil {
		t.Error("ExportSession() expected error for nonexistent session")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("ExportSession() error = %v, should contain 'not found'", err)
	}
}

func TestExportSession_NoAgents(t *testing.T) {
	tempDir := t.TempDir()

	sessionID := "session-no-agents"
	projectDir := filepath.Join(tempDir, "projects", "-test-project")

	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}

	// Create session file without any agents directory
	sessionContent := `{"type":"user","timestamp":"2026-02-01T10:00:00Z","sessionId":"session-no-agents","uuid":"entry-1"}
{"type":"assistant","timestamp":"2026-02-01T10:01:00Z","sessionId":"session-no-agents","uuid":"entry-2"}
`
	sessionFile := filepath.Join(projectDir, sessionID+".jsonl")
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0644); err != nil {
		t.Fatalf("failed to create session file: %v", err)
	}

	opts := ExportOptions{
		OutputDir: filepath.Join(tempDir, "output"),
		ClaudeDir: tempDir,
	}

	result, err := ExportSession("/test/project", sessionID, opts)
	if err != nil {
		t.Fatalf("ExportSession() error = %v", err)
	}

	if result.TotalAgents != 0 {
		t.Errorf("ExportSession() TotalAgents = %d, want 0", result.TotalAgents)
	}

	if len(result.AgentFiles) != 0 {
		t.Errorf("ExportSession() AgentFiles = %v, want empty", result.AgentFiles)
	}
}

func TestExportSession_OutputDirAlreadyExists(t *testing.T) {
	tempDir := t.TempDir()
	_, sessionID := setupTestSession(t, tempDir)

	// Create output directory with existing content
	outputDir := filepath.Join(tempDir, "existing-output")
	sourceDir := filepath.Join(outputDir, "source")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("failed to create existing output dir: %v", err)
	}

	// Create an existing file
	existingFile := filepath.Join(sourceDir, "existing.txt")
	if err := os.WriteFile(existingFile, []byte("existing content"), 0644); err != nil {
		t.Fatalf("failed to create existing file: %v", err)
	}

	opts := ExportOptions{
		OutputDir: outputDir,
		ClaudeDir: tempDir,
	}

	result, err := ExportSession("/test/project", sessionID, opts)
	if err != nil {
		t.Fatalf("ExportSession() error = %v", err)
	}

	// Export should succeed even with existing directory
	if result.OutputDir != outputDir {
		t.Errorf("ExportSession() OutputDir = %v, want %v", result.OutputDir, outputDir)
	}

	// New session file should be created
	if _, err := os.Stat(result.MainSessionFile); os.IsNotExist(err) {
		t.Error("ExportSession() did not create session file in existing directory")
	}
}

func TestCopyAgentFiles_EmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()

	sessionDir := filepath.Join(tempDir, "empty-session")
	destDir := filepath.Join(tempDir, "dest")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatalf("failed to create dest dir: %v", err)
	}

	result := &ExportResult{
		AgentFiles: make(map[string]string),
	}

	// Should not error on missing subagents directory
	if err := copyAgentFiles(sessionDir, destDir, result); err != nil {
		t.Errorf("copyAgentFiles() error = %v, want nil", err)
	}

	if result.TotalAgents != 0 {
		t.Errorf("copyAgentFiles() TotalAgents = %d, want 0", result.TotalAgents)
	}
}

func TestCopyAgentFiles_EmptySubagentsDirectory(t *testing.T) {
	tempDir := t.TempDir()

	sessionDir := filepath.Join(tempDir, "session")
	subagentsDir := filepath.Join(sessionDir, "subagents")
	destDir := filepath.Join(tempDir, "dest")

	if err := os.MkdirAll(subagentsDir, 0755); err != nil {
		t.Fatalf("failed to create subagents dir: %v", err)
	}
	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatalf("failed to create dest dir: %v", err)
	}

	result := &ExportResult{
		AgentFiles: make(map[string]string),
	}

	if err := copyAgentFiles(sessionDir, destDir, result); err != nil {
		t.Errorf("copyAgentFiles() error = %v, want nil", err)
	}

	if result.TotalAgents != 0 {
		t.Errorf("copyAgentFiles() TotalAgents = %d, want 0", result.TotalAgents)
	}
}

func TestCleanupExport(t *testing.T) {
	// Create a valid temp export directory
	tempBase := filepath.Join(os.TempDir(), "claude-history")
	exportDir := filepath.Join(tempBase, "test-cleanup-session")

	if err := os.MkdirAll(exportDir, 0755); err != nil {
		t.Fatalf("failed to create export dir: %v", err)
	}

	// Create some files in it
	testFile := filepath.Join(exportDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Cleanup should succeed
	if err := CleanupExport(exportDir); err != nil {
		t.Errorf("CleanupExport() error = %v", err)
	}

	// Directory should be gone
	if _, err := os.Stat(exportDir); !os.IsNotExist(err) {
		t.Error("CleanupExport() did not remove directory")
	}
}

func TestCleanupExport_SafetyCheck(t *testing.T) {
	tempDir := t.TempDir()

	// Create a directory outside claude-history
	outsideDir := filepath.Join(tempDir, "outside-claude-history")
	if err := os.MkdirAll(outsideDir, 0755); err != nil {
		t.Fatalf("failed to create outside dir: %v", err)
	}

	// Cleanup should refuse to delete it
	err := CleanupExport(outsideDir)
	if err == nil {
		t.Error("CleanupExport() should refuse to delete directory outside claude-history temp")
	}

	if !strings.Contains(err.Error(), "refusing") {
		t.Errorf("CleanupExport() error = %v, should contain 'refusing'", err)
	}

	// Directory should still exist
	if _, err := os.Stat(outsideDir); os.IsNotExist(err) {
		t.Error("CleanupExport() deleted directory it shouldn't have")
	}
}

func TestCleanupExport_NonexistentDirectory(t *testing.T) {
	tempBase := filepath.Join(os.TempDir(), "claude-history")
	nonexistentDir := filepath.Join(tempBase, "nonexistent-for-cleanup-test")

	// Should not error on nonexistent directory
	if err := CleanupExport(nonexistentDir); err != nil {
		t.Errorf("CleanupExport() error = %v, want nil for nonexistent dir", err)
	}
}

func TestExportResult_Structure(t *testing.T) {
	result := &ExportResult{
		OutputDir:       "/tmp/claude-history/test-export",
		SessionID:       "test-session",
		SourceDir:       "/tmp/claude-history/test-export/source",
		MainSessionFile: "/tmp/claude-history/test-export/source/session.jsonl",
		AgentFiles: map[string]string{
			"agent1": "/tmp/claude-history/test-export/source/agents/agent-agent1.jsonl",
			"agent2": "/tmp/claude-history/test-export/source/agents/agent-agent2.jsonl",
		},
		TotalAgents: 2,
		Errors:      []string{"warning: some minor issue"},
	}

	// Verify structure
	if result.OutputDir == "" {
		t.Error("ExportResult OutputDir should not be empty")
	}

	if result.SessionID == "" {
		t.Error("ExportResult SessionID should not be empty")
	}

	if len(result.AgentFiles) != 2 {
		t.Errorf("ExportResult AgentFiles len = %d, want 2", len(result.AgentFiles))
	}

	if result.TotalAgents != 2 {
		t.Errorf("ExportResult TotalAgents = %d, want 2", result.TotalAgents)
	}
}

func TestExportOptions_Defaults(t *testing.T) {
	opts := ExportOptions{}

	// Verify zero values
	if opts.OutputDir != "" {
		t.Errorf("ExportOptions default OutputDir = %v, want empty", opts.OutputDir)
	}

	if opts.ClaudeDir != "" {
		t.Errorf("ExportOptions default ClaudeDir = %v, want empty", opts.ClaudeDir)
	}
}

func TestCopyAgentFilesRecursive_FilesOnly(t *testing.T) {
	tempDir := t.TempDir()

	// Create a subagents directory with only files (no nested dirs)
	srcDir := filepath.Join(tempDir, "subagents")
	destDir := filepath.Join(tempDir, "dest")

	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("failed to create src dir: %v", err)
	}
	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatalf("failed to create dest dir: %v", err)
	}

	// Create agent files
	agents := []string{"agent-a1.jsonl", "agent-b2.jsonl", "agent-c3.jsonl"}
	for _, name := range agents {
		path := filepath.Join(srcDir, name)
		if err := os.WriteFile(path, []byte(`{"test": true}`), 0644); err != nil {
			t.Fatalf("failed to create agent file %s: %v", name, err)
		}
	}

	// Create a non-agent file (should be ignored)
	otherFile := filepath.Join(srcDir, "other.txt")
	if err := os.WriteFile(otherFile, []byte("not an agent"), 0644); err != nil {
		t.Fatalf("failed to create other file: %v", err)
	}

	result := &ExportResult{
		AgentFiles: make(map[string]string),
	}

	if err := copyAgentFilesRecursive(srcDir, destDir, "", result); err != nil {
		t.Fatalf("copyAgentFilesRecursive() error = %v", err)
	}

	// Should have 3 agents
	if result.TotalAgents != 3 {
		t.Errorf("copyAgentFilesRecursive() TotalAgents = %d, want 3", result.TotalAgents)
	}

	// Check all agent IDs were recorded
	expectedIDs := []string{"a1", "b2", "c3"}
	for _, id := range expectedIDs {
		if _, ok := result.AgentFiles[id]; !ok {
			t.Errorf("copyAgentFilesRecursive() missing agent ID %s", id)
		}
	}

	// Verify files exist
	for id, path := range result.AgentFiles {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("copyAgentFilesRecursive() agent file %s does not exist at %s", id, path)
		}
	}

	// Verify other.txt was not copied
	if _, err := os.Stat(filepath.Join(destDir, "other.txt")); !os.IsNotExist(err) {
		t.Error("copyAgentFilesRecursive() should not copy non-agent files")
	}
}

func TestExportSession_JSONLContent(t *testing.T) {
	tempDir := t.TempDir()
	_, sessionID := setupTestSession(t, tempDir)

	outputDir := filepath.Join(tempDir, "export-content")

	opts := ExportOptions{
		OutputDir: outputDir,
		ClaudeDir: tempDir,
	}

	result, err := ExportSession("/test/project", sessionID, opts)
	if err != nil {
		t.Fatalf("ExportSession() error = %v", err)
	}

	// Read and verify session content is valid JSONL
	content, err := os.ReadFile(result.MainSessionFile)
	if err != nil {
		t.Fatalf("failed to read exported session: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 2 {
		t.Errorf("exported session has %d lines, want 2", len(lines))
	}

	// Verify each line is valid JSON (basic check)
	for i, line := range lines {
		if !strings.HasPrefix(line, "{") || !strings.HasSuffix(line, "}") {
			t.Errorf("line %d is not valid JSON: %s", i, line)
		}
	}
}

func TestExportSession_DirectoryStructure(t *testing.T) {
	tempDir := t.TempDir()
	setupTestSession(t, tempDir)
	sessionID := "12345678-1234-1234-1234-123456789abc"

	outputDir := filepath.Join(tempDir, "export-structure")

	opts := ExportOptions{
		OutputDir: outputDir,
		ClaudeDir: tempDir,
	}

	result, err := ExportSession("/test/project", sessionID, opts)
	if err != nil {
		t.Fatalf("ExportSession() error = %v", err)
	}

	// Verify expected directory structure
	expectedDirs := []string{
		result.OutputDir,
		result.SourceDir,
		filepath.Join(result.SourceDir, "agents"),
	}

	for _, dir := range expectedDirs {
		info, err := os.Stat(dir)
		if os.IsNotExist(err) {
			t.Errorf("expected directory does not exist: %s", dir)
			continue
		}
		if !info.IsDir() {
			t.Errorf("expected %s to be a directory", dir)
		}
	}

	// Verify session.jsonl is in source directory
	expectedFile := filepath.Join(result.SourceDir, "session.jsonl")
	if result.MainSessionFile != expectedFile {
		t.Errorf("MainSessionFile = %s, want %s", result.MainSessionFile, expectedFile)
	}
}

func TestGetExportTreeInfo(t *testing.T) {
	tempDir := t.TempDir()
	projectDir, sessionID := setupTestSession(t, tempDir)

	tree, err := GetExportTreeInfo(projectDir, sessionID)
	if err != nil {
		t.Fatalf("GetExportTreeInfo() error = %v", err)
	}

	if tree == nil {
		t.Fatal("GetExportTreeInfo() returned nil tree")
	}

	if !tree.IsRoot {
		t.Error("GetExportTreeInfo() root node should have IsRoot=true")
	}

	if tree.SessionID != sessionID {
		t.Errorf("GetExportTreeInfo() SessionID = %v, want %v", tree.SessionID, sessionID)
	}

	// Should have at least the main session entries
	if tree.EntryCount == 0 {
		t.Error("GetExportTreeInfo() EntryCount should not be 0")
	}

	// Should have child agents
	if len(tree.Children) == 0 {
		t.Error("GetExportTreeInfo() should have child agents")
	}
}

func TestGetExportTreeInfo_NoAgents(t *testing.T) {
	tempDir := t.TempDir()

	sessionID := "session-no-agents-tree"
	projectDir := filepath.Join(tempDir, "projects", "-test-project")

	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}

	// Create session file without any agents
	sessionContent := `{"type":"user","timestamp":"2026-02-01T10:00:00Z","sessionId":"session-no-agents-tree"}
`
	sessionFile := filepath.Join(projectDir, sessionID+".jsonl")
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0644); err != nil {
		t.Fatalf("failed to create session file: %v", err)
	}

	tree, err := GetExportTreeInfo(projectDir, sessionID)
	if err != nil {
		t.Fatalf("GetExportTreeInfo() error = %v", err)
	}

	if len(tree.Children) != 0 {
		t.Errorf("GetExportTreeInfo() Children = %d, want 0", len(tree.Children))
	}
}

func TestCopyAgentFilesRecursive_WithNestedDirectories(t *testing.T) {
	tempDir := t.TempDir()

	// Create a complex nested structure
	srcDir := filepath.Join(tempDir, "subagents")
	nestedDir := filepath.Join(srcDir, "agent-nested123", "subagents")
	deepNestedDir := filepath.Join(srcDir, "agent-deep456", "subagents")

	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("failed to create nested dir: %v", err)
	}
	if err := os.MkdirAll(deepNestedDir, 0755); err != nil {
		t.Fatalf("failed to create deep nested dir: %v", err)
	}

	destDir := filepath.Join(tempDir, "dest")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatalf("failed to create dest dir: %v", err)
	}

	// Create agent files at different levels
	topLevelAgent := filepath.Join(srcDir, "agent-top.jsonl")
	nestedAgent := filepath.Join(nestedDir, "agent-nested-child.jsonl")
	deepAgent := filepath.Join(deepNestedDir, "agent-deep-child.jsonl")

	for _, path := range []string{topLevelAgent, nestedAgent, deepAgent} {
		if err := os.WriteFile(path, []byte(`{"test": true}`), 0644); err != nil {
			t.Fatalf("failed to create agent file %s: %v", path, err)
		}
	}

	result := &ExportResult{
		AgentFiles: make(map[string]string),
	}

	if err := copyAgentFilesRecursive(srcDir, destDir, "", result); err != nil {
		t.Fatalf("copyAgentFilesRecursive() error = %v", err)
	}

	// Should have all agents including nested ones
	if result.TotalAgents < 3 {
		t.Errorf("copyAgentFilesRecursive() TotalAgents = %d, want >= 3", result.TotalAgents)
	}

	// Check specific agents
	expectedAgents := []string{"top", "nested-child", "deep-child"}
	for _, id := range expectedAgents {
		if _, ok := result.AgentFiles[id]; !ok {
			t.Errorf("copyAgentFilesRecursive() missing agent ID %s", id)
		}
	}
}

func TestCopyAgentFilesRecursive_ReadDirError(t *testing.T) {
	tempDir := t.TempDir()
	destDir := filepath.Join(tempDir, "dest")

	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatalf("failed to create dest dir: %v", err)
	}

	result := &ExportResult{
		AgentFiles: make(map[string]string),
	}

	// Try to read a directory that doesn't exist
	nonexistentDir := filepath.Join(tempDir, "nonexistent")
	err := copyAgentFilesRecursive(nonexistentDir, destDir, "", result)

	// Should not return error for nonexistent directory (graceful handling)
	if err != nil {
		t.Errorf("copyAgentFilesRecursive() error = %v, want nil", err)
	}
}

func TestExportSession_InvalidProjectPath(t *testing.T) {
	tempDir := t.TempDir()

	// Create minimal structure but no session
	projectDir := filepath.Join(tempDir, "projects")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}

	opts := ExportOptions{
		OutputDir: filepath.Join(tempDir, "output"),
		ClaudeDir: tempDir,
	}

	// Project path that doesn't have an existing directory
	_, err := ExportSession("/nonexistent/project", "some-session", opts)
	if err == nil {
		t.Error("ExportSession() expected error for invalid project path")
	}
}

func TestCleanupExport_PathTraversal(t *testing.T) {
	// Test that path traversal attempts are blocked
	tempBase := filepath.Join(os.TempDir(), "claude-history")
	traversalPath := filepath.Join(tempBase, "..", "should-not-delete")

	err := CleanupExport(traversalPath)
	if err == nil {
		t.Error("CleanupExport() should refuse path traversal attempts")
	}
}

func TestExportSession_ErrorCopyingAgents(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("skipping permission test when running as root")
	}

	tempDir := t.TempDir()
	_, sessionID := setupTestSession(t, tempDir)

	outputDir := filepath.Join(tempDir, "export-error")

	// Create the output directory
	if err := os.MkdirAll(filepath.Join(outputDir, "source", "agents"), 0755); err != nil {
		t.Fatalf("failed to create output dir: %v", err)
	}

	// Make agents directory read-only to cause copy errors
	agentsDir := filepath.Join(outputDir, "source", "agents")
	if err := os.Chmod(agentsDir, 0555); err != nil {
		t.Fatalf("failed to chmod agents dir: %v", err)
	}
	defer os.Chmod(agentsDir, 0755) // Cleanup

	opts := ExportOptions{
		OutputDir: outputDir,
		ClaudeDir: tempDir,
	}

	result, err := ExportSession("/test/project", sessionID, opts)
	if err != nil {
		t.Fatalf("ExportSession() error = %v (should succeed with warnings)", err)
	}

	// Should have errors recorded
	if len(result.Errors) == 0 {
		t.Error("ExportSession() should have recorded errors for failed agent copies")
	}
}
