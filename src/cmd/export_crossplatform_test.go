package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/randlee/claude-history/pkg/encoding"
)

func TestExport_WindowsPaths(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-specific test")
	}

	// Test with Windows-style path
	tmpDir, projectDir, _ := setupTestProject(t, "windows-test")
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

	// Use Windows-style path for output
	outputDir := filepath.Join(tmpDir, "export-windows")

	exportSessionID = sessionID
	exportFormat = "html"
	exportOutputDir = outputDir
	claudeDir = tmpDir

	// Create a Windows project path (C:\Users\...)
	windowsProjectPath := filepath.Join("C:\\Users\\testuser\\project")
	windowsEncoded := encoding.EncodePath(windowsProjectPath)
	windowsProjectDir := filepath.Join(tmpDir, "projects", windowsEncoded)

	if err := os.MkdirAll(windowsProjectDir, 0755); err != nil {
		t.Fatalf("failed to create Windows project dir: %v", err)
	}

	// Copy session to Windows-encoded directory
	srcSessionFile := filepath.Join(projectDir, sessionID+".jsonl")
	dstSessionFile := filepath.Join(windowsProjectDir, sessionID+".jsonl")

	content, err := os.ReadFile(srcSessionFile)
	if err != nil {
		t.Fatalf("failed to read session: %v", err)
	}

	if err := os.WriteFile(dstSessionFile, content, 0644); err != nil {
		t.Fatalf("failed to write Windows session: %v", err)
	}

	// Run export with Windows path
	err = runExport(exportCmd, []string{windowsProjectPath})
	if err != nil {
		t.Fatalf("Export with Windows path failed: %v", err)
	}

	// Verify output directory created
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		t.Error("output directory not created with Windows path")
	}
}

func TestExport_UnixPaths(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix-specific test")
	}

	// Test with Unix-style path
	tmpDir, projectDir, _ := setupTestProject(t, "unix-test")
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

	outputDir := filepath.Join(tmpDir, "export-unix")

	exportSessionID = sessionID
	exportFormat = "html"
	exportOutputDir = outputDir
	claudeDir = tmpDir

	// Create a Unix project path (/home/user/project)
	unixProjectPath := "/home/testuser/project"
	unixEncoded := encoding.EncodePath(unixProjectPath)
	unixProjectDir := filepath.Join(tmpDir, "projects", unixEncoded)

	if err := os.MkdirAll(unixProjectDir, 0755); err != nil {
		t.Fatalf("failed to create Unix project dir: %v", err)
	}

	// Copy session to Unix-encoded directory
	srcSessionFile := filepath.Join(projectDir, sessionID+".jsonl")
	dstSessionFile := filepath.Join(unixProjectDir, sessionID+".jsonl")

	content, err := os.ReadFile(srcSessionFile)
	if err != nil {
		t.Fatalf("failed to read session: %v", err)
	}

	if err := os.WriteFile(dstSessionFile, content, 0644); err != nil {
		t.Fatalf("failed to write Unix session: %v", err)
	}

	// Run export with Unix path
	err = runExport(exportCmd, []string{unixProjectPath})
	if err != nil {
		t.Fatalf("Export with Unix path failed: %v", err)
	}

	// Verify output directory created
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		t.Error("output directory not created with Unix path")
	}
}

func TestExport_PathEncoding(t *testing.T) {
	// Test with paths containing spaces and special characters
	tmpDir, _, _ := setupTestProject(t, "path-encoding-test")

	// Create project with spaces in path
	projectPath := filepath.Join(tmpDir, "My Project With Spaces")
	encodedName := encoding.EncodePath(projectPath)
	projectDir := filepath.Join(tmpDir, "projects", encodedName)

	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}

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

	outputDir := filepath.Join(tmpDir, "export-encoding")

	exportSessionID = sessionID
	exportFormat = "html"
	exportOutputDir = outputDir
	claudeDir = tmpDir

	// Run export with path containing spaces
	err := runExport(exportCmd, []string{projectPath})
	if err != nil {
		t.Fatalf("Export with spaces in path failed: %v", err)
	}

	// Verify output created
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		t.Error("output directory not created for path with spaces")
	}
}

func TestExport_PathWithSpecialChars(t *testing.T) {
	// Test with paths containing special characters (that are valid in filenames)
	tmpDir, _, _ := setupTestProject(t, "special-chars-path-test")

	// Create project with special characters (use filesystem-safe characters)
	projectName := "project-test_2024"
	projectPath := filepath.Join(tmpDir, projectName)
	encodedName := encoding.EncodePath(projectPath)
	projectDir := filepath.Join(tmpDir, "projects", encodedName)

	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}

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

	outputDir := filepath.Join(tmpDir, "export-special")

	exportSessionID = sessionID
	exportFormat = "html"
	exportOutputDir = outputDir
	claudeDir = tmpDir

	// Run export
	err := runExport(exportCmd, []string{projectPath})
	if err != nil {
		t.Fatalf("Export with special characters in path failed: %v", err)
	}

	// Verify output created
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		t.Error("output directory not created for path with special characters")
	}
}

func TestExport_LineEndings_CRLF(t *testing.T) {
	// Test JSONL files with Windows line endings (\r\n)
	tmpDir, projectDir, projectPath := setupTestProject(t, "crlf-test")

	sessionID := "crlf-session"

	// Create session with CRLF line endings
	sessionContent := "{{\"type\":\"user\",\"timestamp\":\"2026-02-01T10:00:00Z\",\"sessionId\":\"crlf-session\",\"uuid\":\"entry-1\",\"message\":\"Test\"}}\r\n{{\"type\":\"assistant\",\"timestamp\":\"2026-02-01T10:00:05Z\",\"sessionId\":\"crlf-session\",\"uuid\":\"entry-2\",\"message\":\"Response\"}}\r\n"

	// Replace {{ }} with real braces (workaround for formatting)
	sessionContent = strings.ReplaceAll(sessionContent, "{{", "{")
	sessionContent = strings.ReplaceAll(sessionContent, "}}", "}")

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

	outputDir := filepath.Join(tmpDir, "export-crlf")

	exportSessionID = sessionID
	exportFormat = "html"
	exportOutputDir = outputDir
	claudeDir = tmpDir

	// Run export - should handle CRLF correctly
	err := runExport(exportCmd, []string{projectPath})
	if err != nil {
		t.Fatalf("Export with CRLF line endings failed: %v", err)
	}

	// Verify output created
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		t.Error("output directory not created for CRLF session")
	}

	// Verify HTML was generated
	indexPath := filepath.Join(outputDir, "index.html")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		t.Error("index.html not created for CRLF session")
	}
}

func TestExport_LineEndings_LF(t *testing.T) {
	// Test JSONL files with Unix line endings (\n)
	tmpDir, projectDir, projectPath := setupTestProject(t, "lf-test")

	sessionID := "lf-session"

	// Create session with LF line endings
	sessionContent := `{"type":"user","timestamp":"2026-02-01T10:00:00Z","sessionId":"lf-session","uuid":"entry-1","message":"Test"}
{"type":"assistant","timestamp":"2026-02-01T10:00:05Z","sessionId":"lf-session","uuid":"entry-2","message":"Response"}
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

	outputDir := filepath.Join(tmpDir, "export-lf")

	exportSessionID = sessionID
	exportFormat = "html"
	exportOutputDir = outputDir
	claudeDir = tmpDir

	// Run export
	err := runExport(exportCmd, []string{projectPath})
	if err != nil {
		t.Fatalf("Export with LF line endings failed: %v", err)
	}

	// Verify output created
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		t.Error("output directory not created for LF session")
	}
}

func TestExport_FileSeparators(t *testing.T) {
	// Test that filepath operations use correct separators for the platform
	tmpDir, projectDir, projectPath := setupTestProject(t, "separators-test")
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

	outputDir := filepath.Join(tmpDir, "export-separators")

	exportSessionID = sessionID
	exportFormat = "html"
	exportOutputDir = outputDir
	claudeDir = tmpDir

	// Run export
	err := runExport(exportCmd, []string{projectPath})
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify all paths use correct separator
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("failed to read output dir: %v", err)
	}

	for _, entry := range entries {
		name := entry.Name()
		// Verify no wrong separators in filenames
		if runtime.GOOS == "windows" {
			if strings.Contains(name, "/") {
				t.Errorf("filename contains forward slash on Windows: %s", name)
			}
		} else {
			if strings.Contains(name, "\\") {
				t.Errorf("filename contains backslash on Unix: %s", name)
			}
		}
	}
}

func TestExport_TempDirectory(t *testing.T) {
	// Test that os.TempDir() returns valid path on all platforms
	tmpDir, projectDir, projectPath := setupTestProject(t, "temp-dir-test")
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

	// Don't specify output directory - use auto-generated temp path
	exportSessionID = sessionID
	exportFormat = "html"
	exportOutputDir = ""
	claudeDir = tmpDir

	// Run export
	err := runExport(exportCmd, []string{projectPath})
	if err != nil {
		t.Fatalf("Export with temp directory failed: %v", err)
	}

	// Verify temp directory is platform-appropriate
	tempDir := os.TempDir()

	if runtime.GOOS == "windows" {
		// Windows temp should be like C:\Users\...\AppData\Local\Temp
		if !filepath.IsAbs(tempDir) {
			t.Errorf("Windows temp directory should be absolute: %s", tempDir)
		}
	} else {
		// Unix temp typically /tmp or /var/tmp
		if !strings.HasPrefix(tempDir, "/") {
			t.Errorf("Unix temp directory should start with /: %s", tempDir)
		}
	}
}

func TestExport_AbsoluteVsRelative(t *testing.T) {
	// Test that both absolute and relative paths work
	tmpDir, projectDir, _ := setupTestProject(t, "abs-rel-test")
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

	// Test with absolute output path
	absOutputDir := filepath.Join(tmpDir, "export-absolute")

	exportSessionID = sessionID
	exportFormat = "html"
	exportOutputDir = absOutputDir
	claudeDir = tmpDir

	// First export with absolute path
	projectPath := filepath.Join(tmpDir, "abs-rel-test")
	err := runExport(exportCmd, []string{projectPath})
	if err != nil {
		t.Fatalf("Export with absolute path failed: %v", err)
	}

	if _, err := os.Stat(absOutputDir); os.IsNotExist(err) {
		t.Error("absolute output directory not created")
	}

	// Test with relative output path
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	relOutputDir := "export-relative"
	exportOutputDir = relOutputDir

	err = runExport(exportCmd, []string{"abs-rel-test"})
	if err != nil {
		t.Fatalf("Export with relative path failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, relOutputDir)); os.IsNotExist(err) {
		t.Error("relative output directory not created")
	}
}
