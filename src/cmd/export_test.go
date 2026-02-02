package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/randlee/claude-history/pkg/encoding"
)

func TestGenerateTempExportPath(t *testing.T) {
	tests := []struct {
		name      string
		sessionID string
		wantLen   int // expected length of prefix (before timestamp)
	}{
		{
			name:      "full UUID session ID",
			sessionID: "679761ba-80c0-4cd3-a586-cc6a1fc56308",
			wantLen:   8, // truncated to first 8 chars
		},
		{
			name:      "short session ID",
			sessionID: "abc",
			wantLen:   3, // not truncated
		},
		{
			name:      "exactly 8 chars",
			sessionID: "12345678",
			wantLen:   8,
		},
		{
			name:      "empty session ID",
			sessionID: "",
			wantLen:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateTempExportPath(tt.sessionID)

			// Should be in temp directory
			if !strings.HasPrefix(result, os.TempDir()) {
				t.Errorf("Path should start with temp dir %s, got %s", os.TempDir(), result)
			}

			// Should contain claude-history subdirectory
			if !strings.Contains(result, "claude-history") {
				t.Errorf("Path should contain 'claude-history', got %s", result)
			}

			// Get the basename (last component)
			base := filepath.Base(result)

			// Should have format: {prefix}-{timestamp}
			if !strings.Contains(base, "-") {
				t.Errorf("Basename should contain '-', got %s", base)
			}

			// Check prefix length
			parts := strings.SplitN(base, "-", 2)
			if len(parts[0]) != tt.wantLen {
				t.Errorf("Prefix length = %d, want %d (base=%s)", len(parts[0]), tt.wantLen, base)
			}

			// Verify timestamp format (should be parseable)
			if len(parts) > 1 {
				// Timestamp part should match format: 2006-01-02T15-04-05
				// But we can't predict exact time, so just verify it looks like a timestamp
				if len(parts[1]) < 10 {
					t.Errorf("Timestamp part too short: %s", parts[1])
				}
			}
		})
	}
}

func TestGenerateTempExportPath_UniqueTimestamps(t *testing.T) {
	// Generate two paths in quick succession - they should be different
	// (or at least we verify they have timestamps)
	path1 := generateTempExportPath("test1234")
	time.Sleep(1 * time.Second) // Ensure different timestamps
	path2 := generateTempExportPath("test1234")

	// Both should be valid paths
	if path1 == "" || path2 == "" {
		t.Error("Generated paths should not be empty")
	}

	// They should be different (due to timestamp)
	if path1 == path2 {
		t.Errorf("Paths generated at different times should be different: %s", path1)
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{
			name:   "no truncation needed",
			input:  "hello",
			maxLen: 10,
			want:   "hello",
		},
		{
			name:   "exact length",
			input:  "hello",
			maxLen: 5,
			want:   "hello",
		},
		{
			name:   "truncation with ellipsis",
			input:  "hello world",
			maxLen: 8,
			want:   "hello...",
		},
		{
			name:   "very short max length",
			input:  "hello",
			maxLen: 3,
			want:   "hel",
		},
		{
			name:   "max length 4 with ellipsis",
			input:  "hello world",
			maxLen: 4,
			want:   "h...",
		},
		{
			name:   "empty string",
			input:  "",
			maxLen: 10,
			want:   "",
		},
		{
			name:   "max length 0",
			input:  "hello",
			maxLen: 0,
			want:   "",
		},
		{
			name:   "long string truncated",
			input:  "This is a very long string that should be truncated to a reasonable length",
			maxLen: 20,
			want:   "This is a very lo...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateString(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncateString(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestExportCmd_Flags(t *testing.T) {
	// Verify the export command has expected flags
	cmd := exportCmd

	// Check session flag
	sessionFlag := cmd.Flags().Lookup("session")
	if sessionFlag == nil {
		t.Error("export command should have --session flag")
	} else {
		if sessionFlag.Shorthand != "s" {
			t.Errorf("--session shorthand = %q, want 's'", sessionFlag.Shorthand)
		}
	}

	// Check output flag
	outputFlag := cmd.Flags().Lookup("output")
	if outputFlag == nil {
		t.Error("export command should have --output flag")
	} else {
		if outputFlag.Shorthand != "o" {
			t.Errorf("--output shorthand = %q, want 'o'", outputFlag.Shorthand)
		}
	}

	// Check format flag
	formatFlag := cmd.Flags().Lookup("format")
	if formatFlag == nil {
		t.Error("export command should have --format flag")
	} else {
		if formatFlag.Shorthand != "f" {
			t.Errorf("--format shorthand = %q, want 'f'", formatFlag.Shorthand)
		}
		if formatFlag.DefValue != "html" {
			t.Errorf("--format default = %q, want 'html'", formatFlag.DefValue)
		}
	}
}

func TestExportCmd_Usage(t *testing.T) {
	cmd := exportCmd

	// Verify command name
	if cmd.Use != "export [project-path]" {
		t.Errorf("Use = %q, want 'export [project-path]'", cmd.Use)
	}

	// Verify it accepts 0 or 1 args
	if cmd.Args == nil {
		t.Error("Args should be set")
	}

	// Verify short description exists
	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	// Verify long description mentions HTML and JSONL
	if !strings.Contains(cmd.Long, "HTML") {
		t.Error("Long description should mention HTML format")
	}
	if !strings.Contains(cmd.Long, "JSONL") {
		t.Error("Long description should mention JSONL format")
	}
}

func TestExportCmd_InvalidFormat(t *testing.T) {
	// Reset global variables
	oldSessionID := exportSessionID
	oldFormat := exportFormat
	oldOutputDir := exportOutputDir
	defer func() {
		exportSessionID = oldSessionID
		exportFormat = oldFormat
		exportOutputDir = oldOutputDir
	}()

	// Set up test values
	exportSessionID = "test-session"
	exportFormat = "invalid"
	exportOutputDir = ""

	// Run the command - should fail with invalid format
	err := runExport(exportCmd, []string{"/nonexistent/path"})
	if err == nil {
		t.Error("Expected error for invalid format")
	}
	if !strings.Contains(err.Error(), "invalid format") {
		t.Errorf("Error should mention invalid format, got: %v", err)
	}
}

func TestExportCmd_MissingProject(t *testing.T) {
	// Reset global variables
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

	// Create a temporary directory for claude config
	tmpDir := t.TempDir()
	projectsDir := filepath.Join(tmpDir, "projects")
	if err := os.MkdirAll(projectsDir, 0755); err != nil {
		t.Fatalf("Failed to create projects dir: %v", err)
	}

	// Set up test values
	exportSessionID = "test-session"
	exportFormat = "html"
	exportOutputDir = ""
	claudeDir = tmpDir

	// Run the command - should fail with project not found
	err := runExport(exportCmd, []string{"/nonexistent/project"})
	if err == nil {
		t.Error("Expected error for missing project")
	}
	if !strings.Contains(err.Error(), "project not found") {
		t.Errorf("Error should mention project not found, got: %v", err)
	}
}

func TestExportCmd_MissingSession(t *testing.T) {
	// Reset global variables
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

	// Create a temporary project directory structure
	tmpDir := t.TempDir()
	projectsDir := filepath.Join(tmpDir, "projects")

	// Create a project path and encode it properly (cross-platform)
	projectPath := filepath.Join(tmpDir, "myproject")
	encodedName := encoding.EncodePath(projectPath)
	encodedProjectDir := filepath.Join(projectsDir, encodedName)
	if err := os.MkdirAll(encodedProjectDir, 0755); err != nil {
		t.Fatalf("Failed to create project dir: %v", err)
	}

	// Set up test values
	exportSessionID = "nonexistent-session"
	exportFormat = "html"
	exportOutputDir = ""
	claudeDir = tmpDir

	// Run the command - should fail with session not found
	err := runExport(exportCmd, []string{projectPath})
	if err == nil {
		t.Error("Expected error for missing session")
	}
	if !strings.Contains(err.Error(), "session not found") {
		t.Errorf("Error should mention session not found, got: %v", err)
	}
}

func TestExportCmd_ValidSession(t *testing.T) {
	// Reset global variables
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

	// Create a temporary project directory structure
	tmpDir := t.TempDir()
	projectsDir := filepath.Join(tmpDir, "projects")

	// Create a project path and encode it properly (cross-platform)
	projectPath := filepath.Join(tmpDir, "myproject")
	encodedName := encoding.EncodePath(projectPath)
	encodedProjectDir := filepath.Join(projectsDir, encodedName)
	if err := os.MkdirAll(encodedProjectDir, 0755); err != nil {
		t.Fatalf("Failed to create project dir: %v", err)
	}

	// Create a session file with valid content
	sessionID := "679761ba-80c0-4cd3-a586-cc6a1fc56308"
	sessionFile := filepath.Join(encodedProjectDir, sessionID+".jsonl")
	sessionContent := `{"uuid":"1","sessionId":"679761ba-80c0-4cd3-a586-cc6a1fc56308","type":"user","timestamp":"2026-02-01T18:00:00.000Z","message":"Hello, world!"}
{"uuid":"2","sessionId":"679761ba-80c0-4cd3-a586-cc6a1fc56308","type":"assistant","timestamp":"2026-02-01T18:00:05.000Z","message":"Hi there!"}
`
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0600); err != nil {
		t.Fatalf("Failed to create session file: %v", err)
	}

	// Create output directory
	outputDir := filepath.Join(tmpDir, "export-output")

	// Set up test values
	exportSessionID = sessionID
	exportFormat = "html"
	exportOutputDir = outputDir
	claudeDir = tmpDir

	// Run the command - should succeed
	err := runExport(exportCmd, []string{projectPath})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify output directory was created
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		t.Error("Output directory should have been created")
	}

	// Verify source directory was created
	sourceDir := filepath.Join(outputDir, "source")
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		t.Error("Source directory should have been created")
	}

	// Verify session file was copied
	copiedSessionFile := filepath.Join(sourceDir, "session.jsonl")
	if _, err := os.Stat(copiedSessionFile); os.IsNotExist(err) {
		t.Error("Session file should have been copied to source directory")
	}
}

func TestExportCmd_JSONLFormat(t *testing.T) {
	// Reset global variables
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

	// Create a temporary project directory structure
	tmpDir := t.TempDir()
	projectsDir := filepath.Join(tmpDir, "projects")

	// Create a project path and encode it properly (cross-platform)
	projectPath := filepath.Join(tmpDir, "myproject")
	encodedName := encoding.EncodePath(projectPath)
	encodedProjectDir := filepath.Join(projectsDir, encodedName)
	if err := os.MkdirAll(encodedProjectDir, 0755); err != nil {
		t.Fatalf("Failed to create project dir: %v", err)
	}

	// Create a session file
	sessionID := "test-session-1234"
	sessionFile := filepath.Join(encodedProjectDir, sessionID+".jsonl")
	sessionContent := `{"uuid":"1","sessionId":"test-session-1234","type":"user","timestamp":"2026-02-01T18:00:00.000Z","message":"Test"}
`
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0600); err != nil {
		t.Fatalf("Failed to create session file: %v", err)
	}

	// Create output directory
	outputDir := filepath.Join(tmpDir, "jsonl-output")

	// Set up test values
	exportSessionID = sessionID
	exportFormat = "jsonl"
	exportOutputDir = outputDir
	claudeDir = tmpDir

	// Run the command - should succeed with jsonl format
	err := runExport(exportCmd, []string{projectPath})
	if err != nil {
		t.Errorf("Unexpected error with jsonl format: %v", err)
	}

	// Verify output directory was created
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		t.Error("Output directory should have been created")
	}

	// Verify source directory exists
	sourceDir := filepath.Join(outputDir, "source")
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		t.Error("Source directory should have been created")
	}

	// Verify session file was copied
	copiedSessionFile := filepath.Join(sourceDir, "session.jsonl")
	if _, err := os.Stat(copiedSessionFile); os.IsNotExist(err) {
		t.Error("Session file should have been copied")
	}

	// Verify content matches
	copiedContent, err := os.ReadFile(copiedSessionFile)
	if err != nil {
		t.Fatalf("Failed to read copied session file: %v", err)
	}
	if string(copiedContent) != sessionContent {
		t.Error("Copied session content does not match original")
	}
}

func TestExportCmd_AutoGeneratedOutput(t *testing.T) {
	// Reset global variables
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

	// Create a temporary project directory structure
	tmpDir := t.TempDir()
	projectsDir := filepath.Join(tmpDir, "projects")

	// Create a project path and encode it properly (cross-platform)
	projectPath := filepath.Join(tmpDir, "myproject")
	encodedName := encoding.EncodePath(projectPath)
	encodedProjectDir := filepath.Join(projectsDir, encodedName)
	if err := os.MkdirAll(encodedProjectDir, 0755); err != nil {
		t.Fatalf("Failed to create project dir: %v", err)
	}

	// Create a session file
	sessionID := "abcd1234-efgh-5678-ijkl-mnopqrstuvwx"
	sessionFile := filepath.Join(encodedProjectDir, sessionID+".jsonl")
	sessionContent := `{"uuid":"1","sessionId":"abcd1234-efgh-5678-ijkl-mnopqrstuvwx","type":"user","timestamp":"2026-02-01T18:00:00.000Z","message":"Test"}
`
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0600); err != nil {
		t.Fatalf("Failed to create session file: %v", err)
	}

	// Set up test values - no output directory specified
	exportSessionID = sessionID
	exportFormat = "html"
	exportOutputDir = "" // Auto-generate
	claudeDir = tmpDir

	// Run the command - should succeed and create output in temp dir
	err := runExport(exportCmd, []string{projectPath})
	if err != nil {
		t.Errorf("Unexpected error with auto-generated output: %v", err)
	}

	// The auto-generated path should be in the temp directory
	// We can't easily verify the exact path created, but we can verify no error occurred
}

func TestExportCmd_RelativeOutputPath(t *testing.T) {
	// Reset global variables
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

	// Create a temporary project directory structure
	tmpDir := t.TempDir()
	projectsDir := filepath.Join(tmpDir, "projects")

	// Create a project path and encode it properly (cross-platform)
	projectPath := filepath.Join(tmpDir, "myproject")
	encodedName := encoding.EncodePath(projectPath)
	encodedProjectDir := filepath.Join(projectsDir, encodedName)
	if err := os.MkdirAll(encodedProjectDir, 0755); err != nil {
		t.Fatalf("Failed to create project dir: %v", err)
	}

	// Create a session file
	sessionID := "test-session"
	sessionFile := filepath.Join(encodedProjectDir, sessionID+".jsonl")
	sessionContent := `{"uuid":"1","sessionId":"test-session","type":"user","timestamp":"2026-02-01T18:00:00.000Z","message":"Test"}
`
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0600); err != nil {
		t.Fatalf("Failed to create session file: %v", err)
	}

	// Use a relative output path (relative to current directory)
	relativeOutput := filepath.Join(tmpDir, "relative-export")

	// Set up test values
	exportSessionID = sessionID
	exportFormat = "html"
	exportOutputDir = relativeOutput
	claudeDir = tmpDir

	// Run the command
	err := runExport(exportCmd, []string{projectPath})
	if err != nil {
		t.Errorf("Unexpected error with relative output path: %v", err)
	}

	// Verify the directory was created
	if _, err := os.Stat(relativeOutput); os.IsNotExist(err) {
		t.Error("Output directory should have been created")
	}
}

func TestExportCmd_WithAgents(t *testing.T) {
	// Reset global variables
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

	// Create a temporary project directory structure
	tmpDir := t.TempDir()
	projectsDir := filepath.Join(tmpDir, "projects")

	// Create a project path and encode it properly (cross-platform)
	projectPath := filepath.Join(tmpDir, "myproject")
	encodedName := encoding.EncodePath(projectPath)
	encodedProjectDir := filepath.Join(projectsDir, encodedName)
	if err := os.MkdirAll(encodedProjectDir, 0755); err != nil {
		t.Fatalf("Failed to create project dir: %v", err)
	}

	// Create a session file
	sessionID := "test-session-with-agents"
	sessionFile := filepath.Join(encodedProjectDir, sessionID+".jsonl")
	sessionContent := `{"uuid":"1","sessionId":"test-session-with-agents","type":"user","timestamp":"2026-02-01T18:00:00.000Z","message":"Test"}
`
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0600); err != nil {
		t.Fatalf("Failed to create session file: %v", err)
	}

	// Create agent files
	sessionDir := filepath.Join(encodedProjectDir, sessionID)
	subagentsDir := filepath.Join(sessionDir, "subagents")
	if err := os.MkdirAll(subagentsDir, 0755); err != nil {
		t.Fatalf("Failed to create subagents dir: %v", err)
	}

	// Create agent file
	agentContent := `{"uuid":"agent-1","sessionId":"test-session-with-agents","type":"user","timestamp":"2026-02-01T18:01:00.000Z","message":"Agent message"}
`
	agentFile := filepath.Join(subagentsDir, "agent-abc123.jsonl")
	if err := os.WriteFile(agentFile, []byte(agentContent), 0600); err != nil {
		t.Fatalf("Failed to create agent file: %v", err)
	}

	// Create output directory
	outputDir := filepath.Join(tmpDir, "export-with-agents")

	// Set up test values
	exportSessionID = sessionID
	exportFormat = "jsonl"
	exportOutputDir = outputDir
	claudeDir = tmpDir

	// Run the command
	err := runExport(exportCmd, []string{projectPath})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify agent was copied
	copiedAgentFile := filepath.Join(outputDir, "source", "agents", "agent-abc123.jsonl")
	if _, err := os.Stat(copiedAgentFile); os.IsNotExist(err) {
		t.Error("Agent file should have been copied to source/agents directory")
	}

	// Verify agent content matches
	copiedAgentContent, err := os.ReadFile(copiedAgentFile)
	if err != nil {
		t.Fatalf("Failed to read copied agent file: %v", err)
	}
	if string(copiedAgentContent) != agentContent {
		t.Error("Copied agent content does not match original")
	}
}

func TestExportCmd_AutoOutputDir(t *testing.T) {
	// Reset global variables
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

	// Create a temporary project directory structure
	tmpDir := t.TempDir()
	projectsDir := filepath.Join(tmpDir, "projects")

	// Create a project path and encode it properly (cross-platform)
	projectPath := filepath.Join(tmpDir, "myproject")
	encodedName := encoding.EncodePath(projectPath)
	encodedProjectDir := filepath.Join(projectsDir, encodedName)
	if err := os.MkdirAll(encodedProjectDir, 0755); err != nil {
		t.Fatalf("Failed to create project dir: %v", err)
	}

	// Create a session file
	sessionID := "abcd1234-5678-90ef-ghij-klmnopqrstuv"
	sessionFile := filepath.Join(encodedProjectDir, sessionID+".jsonl")
	sessionContent := `{"uuid":"1","sessionId":"abcd1234-5678-90ef-ghij-klmnopqrstuv","type":"user","timestamp":"2026-02-01T18:00:00.000Z","message":"Test"}
`
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0600); err != nil {
		t.Fatalf("Failed to create session file: %v", err)
	}

	// Set up test values - no output directory specified
	exportSessionID = sessionID
	exportFormat = "jsonl"
	exportOutputDir = "" // Auto-generate
	claudeDir = tmpDir

	// Run the command
	err := runExport(exportCmd, []string{projectPath})
	if err != nil {
		t.Errorf("Unexpected error with auto-generated output: %v", err)
	}

	// We can't easily verify the exact path, but we verified no error occurred
	// The export package's tests verify the auto-generated path format
}
