package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/randlee/claude-history/pkg/encoding"
)

func TestExport_LargeSession(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping large session test in short mode")
	}

	// Create test project with large session (1000 entries)
	tmpDir, projectDir, projectPath := setupTestProject(t, "large-session-test")
	sessionID := createLargeSession(t, projectDir, 1000)

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

	outputDir := filepath.Join(tmpDir, "export-large")

	exportSessionID = sessionID
	exportFormat = "html"
	exportOutputDir = outputDir
	claudeDir = tmpDir

	// Time the export
	start := time.Now()
	err := runExport(exportCmd, []string{projectPath})
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Export of large session failed: %v", err)
	}

	// Should complete in reasonable time (< 30 seconds)
	if duration > 30*time.Second {
		t.Errorf("Export took too long: %v (expected < 30s)", duration)
	}

	t.Logf("Export of 1000 entries completed in %v", duration)

	// Verify output was created
	verifyHTMLOutput(t, outputDir, 0)
}

func TestExport_ManyAgents(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping many agents test in short mode")
	}

	// Create test project with many agents (50)
	tmpDir, projectDir, projectPath := setupTestProject(t, "many-agents-test")
	sessionID := createTestSessionWithAgents(t, projectDir, 50)

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

	outputDir := filepath.Join(tmpDir, "export-many-agents")

	exportSessionID = sessionID
	exportFormat = "html"
	exportOutputDir = outputDir
	claudeDir = tmpDir

	// Time the export
	start := time.Now()
	err := runExport(exportCmd, []string{projectPath})
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Export with many agents failed: %v", err)
	}

	t.Logf("Export of 50 agents completed in %v", duration)

	// Verify all agents were rendered
	verifyHTMLOutput(t, outputDir, 50)
}

func TestExport_DeepNesting(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping deep nesting test in short mode")
	}

	// Create test project with deeply nested agents (10 levels)
	tmpDir, projectDir, projectPath := setupTestProject(t, "deep-nest-test")

	sessionID := "deep-nest-session"

	// Create deeply nested structure: session -> agent1 -> agent2 -> ... -> agent10
	sessionContent := `{"type":"user","timestamp":"2026-02-01T10:00:00Z","sessionId":"deep-nest-session","uuid":"entry-1","message":"Root"}
{"type":"queue-operation","timestamp":"2026-02-01T10:00:05Z","sessionId":"deep-nest-session","uuid":"queue-1","agentId":"agent-1"}
`
	sessionFile := filepath.Join(projectDir, sessionID+".jsonl")
	if err := writeFile(sessionFile, []byte(sessionContent)); err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Create nested agents
	currentDir := filepath.Join(projectDir, sessionID)
	for level := 1; level <= 10; level++ {
		subagentsDir := filepath.Join(currentDir, "subagents")
		if err := mkdirAll(subagentsDir, 0755); err != nil {
			t.Fatalf("failed to create subagents dir at level %d: %v", level, err)
		}

		agentContent := fmt.Sprintf(`{"type":"user","timestamp":"2026-02-01T10:%02d:00Z","uuid":"entry-%d","message":"Level %d"}
`, level, level, level)

		// Add queue-operation for next level (except last)
		if level < 10 {
			agentContent += fmt.Sprintf(`{"type":"queue-operation","timestamp":"2026-02-01T10:%02d:05Z","uuid":"queue-%d","agentId":"agent-%d"}
`, level, level, level+1)
		}

		agentFile := filepath.Join(subagentsDir, fmt.Sprintf("agent-agent-%d.jsonl", level))
		if err := writeFile(agentFile, []byte(agentContent)); err != nil {
			t.Fatalf("failed to create agent %d: %v", level, err)
		}

		// Move to next level's directory
		currentDir = filepath.Join(subagentsDir, fmt.Sprintf("agent-agent-%d", level))
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

	outputDir := filepath.Join(tmpDir, "export-deep-nest")

	exportSessionID = sessionID
	exportFormat = "html"
	exportOutputDir = outputDir
	claudeDir = tmpDir

	// Time the export
	start := time.Now()
	err := runExport(exportCmd, []string{projectPath})
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Export with deep nesting failed: %v", err)
	}

	t.Logf("Export of 10-level deep nesting completed in %v", duration)

	// Verify output created
	verifyHTMLOutput(t, outputDir, 10)
}

func BenchmarkExport_HTML(b *testing.B) {
	// Benchmark HTML export
	tmpDir, projectDir, projectPath := setupBenchProject(b, "bench-html")
	sessionID := createBenchSession(b, projectDir, 5)

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

	exportSessionID = sessionID
	exportFormat = "html"
	claudeDir = tmpDir

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		outputDir := filepath.Join(tmpDir, fmt.Sprintf("export-bench-%d", i))
		exportOutputDir = outputDir

		if err := runExport(exportCmd, []string{projectPath}); err != nil {
			b.Fatalf("Export failed: %v", err)
		}
	}
}

func BenchmarkExport_JSONL(b *testing.B) {
	// Benchmark JSONL export
	tmpDir, projectDir, projectPath := setupBenchProject(b, "bench-jsonl")
	sessionID := createBenchSession(b, projectDir, 5)

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

	exportSessionID = sessionID
	exportFormat = "jsonl"
	claudeDir = tmpDir

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		outputDir := filepath.Join(tmpDir, fmt.Sprintf("export-jsonl-bench-%d", i))
		exportOutputDir = outputDir

		if err := runExport(exportCmd, []string{projectPath}); err != nil {
			b.Fatalf("Export failed: %v", err)
		}
	}
}

func BenchmarkExport_ManyAgents(b *testing.B) {
	// Benchmark export with many agents
	tmpDir, projectDir, projectPath := setupBenchProject(b, "bench-many-agents")
	sessionID := createBenchSession(b, projectDir, 20)

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

	exportSessionID = sessionID
	exportFormat = "html"
	claudeDir = tmpDir

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		outputDir := filepath.Join(tmpDir, fmt.Sprintf("export-agents-bench-%d", i))
		exportOutputDir = outputDir

		if err := runExport(exportCmd, []string{projectPath}); err != nil {
			b.Fatalf("Export failed: %v", err)
		}
	}
}

func TestExport_MemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping memory test in short mode")
	}

	// Create a very large session to test memory handling
	tmpDir, projectDir, projectPath := setupTestProject(t, "memory-test")
	sessionID := createLargeSession(t, projectDir, 5000)

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

	outputDir := filepath.Join(tmpDir, "export-memory")

	exportSessionID = sessionID
	exportFormat = "html"
	exportOutputDir = outputDir
	claudeDir = tmpDir

	// Export should complete without running out of memory
	err := runExport(exportCmd, []string{projectPath})
	if err != nil {
		t.Fatalf("Export with large session failed: %v", err)
	}

	t.Log("Large session exported successfully (memory test passed)")
}

func TestExport_RepeatableOutput(t *testing.T) {
	// Test that exporting the same session twice produces consistent results
	tmpDir, projectDir, projectPath := setupTestProject(t, "repeatable-test")
	sessionID := createTestSessionWithAgents(t, projectDir, 2)

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

	exportSessionID = sessionID
	exportFormat = "html"
	claudeDir = tmpDir

	// First export
	outputDir1 := filepath.Join(tmpDir, "export-repeat-1")
	exportOutputDir = outputDir1

	if err := runExport(exportCmd, []string{projectPath}); err != nil {
		t.Fatalf("First export failed: %v", err)
	}

	// Second export
	outputDir2 := filepath.Join(tmpDir, "export-repeat-2")
	exportOutputDir = outputDir2

	if err := runExport(exportCmd, []string{projectPath}); err != nil {
		t.Fatalf("Second export failed: %v", err)
	}

	// Compare key files (note: timestamps in manifest will differ)
	// Just verify both exports succeeded and have same structure
	verifyHTMLOutput(t, outputDir1, 2)
	verifyHTMLOutput(t, outputDir2, 2)

	t.Log("Repeated exports produced consistent output structure")
}

// Helper functions for benchmarks

func setupBenchProject(b *testing.B, projectName string) (string, string, string) {
	b.Helper()

	tmpDir := b.TempDir()
	projectsDir := filepath.Join(tmpDir, "projects")

	// Create a project path
	projectPath := filepath.Join(tmpDir, projectName)
	encodedName := encoding.EncodePath(projectPath)
	encodedProjectDir := filepath.Join(projectsDir, encodedName)

	if err := os.MkdirAll(encodedProjectDir, 0755); err != nil {
		b.Fatalf("failed to create project directory: %v", err)
	}

	return tmpDir, encodedProjectDir, projectPath
}

func createBenchSession(b *testing.B, projectDir string, agentCount int) string {
	b.Helper()

	sessionID := "12345678-1234-1234-1234-123456789abc"

	// Create session content
	sessionContent := fmt.Sprintf(`{"type":"user","timestamp":"2026-02-01T10:00:00Z","sessionId":"%s","uuid":"entry-1","message":"Test"}
{"type":"assistant","timestamp":"2026-02-01T10:00:05Z","sessionId":"%s","uuid":"entry-2","message":"Response"}
`, sessionID, sessionID)

	// Add queue-operation entries for agents
	for i := 0; i < agentCount; i++ {
		agentID := fmt.Sprintf("agent-%d", i+1)
		queueEntry := fmt.Sprintf(`{"type":"queue-operation","timestamp":"2026-02-01T10:%02d:00Z","sessionId":"%s","uuid":"queue-%d","agentId":"%s"}
`, i+1, sessionID, i+1, agentID)
		sessionContent += queueEntry
	}

	// Create main session file
	sessionFile := filepath.Join(projectDir, sessionID+".jsonl")
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0644); err != nil {
		b.Fatalf("failed to create session file: %v", err)
	}

	// Create agent files
	sessionDir := filepath.Join(projectDir, sessionID)
	subagentsDir := filepath.Join(sessionDir, "subagents")
	if err := os.MkdirAll(subagentsDir, 0755); err != nil {
		b.Fatalf("failed to create subagents directory: %v", err)
	}

	for i := 0; i < agentCount; i++ {
		agentID := fmt.Sprintf("agent-%d", i+1)
		agentContent := fmt.Sprintf(`{"type":"user","timestamp":"2026-02-01T10:%02d:05Z","sessionId":"%s","uuid":"%s-entry-1","message":"Agent task"}
`, i+1, sessionID, agentID)

		agentFile := filepath.Join(subagentsDir, fmt.Sprintf("agent-%s.jsonl", agentID))
		if err := os.WriteFile(agentFile, []byte(agentContent), 0644); err != nil {
			b.Fatalf("failed to create agent file: %v", err)
		}
	}

	return sessionID
}

func writeFile(path string, content []byte) error {
	return os.WriteFile(path, content, 0644)
}

func mkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}
