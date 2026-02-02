package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/randlee/claude-history/pkg/encoding"
	"github.com/randlee/claude-history/pkg/export"
)

func TestExportCommand_HTMLFormat(t *testing.T) {
	// Setup test environment
	tempDir := t.TempDir()
	projectPath := filepath.Join(tempDir, "test-project")
	claudeDir := filepath.Join(tempDir, ".claude")

	if err := os.MkdirAll(projectPath, 0755); err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	// Create a minimal test session
	encodedPath := encoding.EncodePath(projectPath)
	projectDir := filepath.Join(claudeDir, "projects", encodedPath)
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("Failed to create Claude project directory: %v", err)
	}

	sessionID := "test-session-123"
	sessionFile := filepath.Join(projectDir, sessionID+".jsonl")

	// Create a minimal session file with valid JSONL
	sessionContent := `{"uuid":"entry-1","type":"user","timestamp":"2026-02-01T10:00:00Z","message":[{"type":"text","text":"Hello"}]}
{"uuid":"entry-2","type":"assistant","timestamp":"2026-02-01T10:00:01Z","message":[{"type":"text","text":"Hello! How can I help?"}]}
`
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0644); err != nil {
		t.Fatalf("Failed to write session file: %v", err)
	}

	// Run export command with HTML format
	outputDir := filepath.Join(tempDir, "export-output")
	opts := export.ExportOptions{
		OutputDir: outputDir,
		ClaudeDir: claudeDir,
	}

	result, err := export.ExportSession(projectPath, sessionID, opts)
	if err != nil {
		t.Fatalf("ExportSession failed: %v", err)
	}

	// Verify JSONL files were created
	if _, err := os.Stat(result.MainSessionFile); os.IsNotExist(err) {
		t.Errorf("Main session file not created: %s", result.MainSessionFile)
	}

	// Now test HTML rendering
	if err := renderHTML(result, projectPath, projectDir, sessionID); err != nil {
		t.Fatalf("renderHTML failed: %v", err)
	}

	// Verify HTML files were created
	indexPath := filepath.Join(outputDir, "index.html")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		t.Errorf("index.html not created")
	}

	// Verify static assets were created
	cssPath := filepath.Join(outputDir, "static", "style.css")
	if _, err := os.Stat(cssPath); os.IsNotExist(err) {
		t.Errorf("style.css not created")
	}

	jsPath := filepath.Join(outputDir, "static", "script.js")
	if _, err := os.Stat(jsPath); os.IsNotExist(err) {
		t.Errorf("script.js not created")
	}

	// Verify manifest was created
	manifestPath := filepath.Join(outputDir, "manifest.json")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Errorf("manifest.json not created")
	}

	// Verify HTML message is valid
	htmlContent, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("Failed to read index.html: %v", err)
	}

	htmlStr := string(htmlContent)
	if !strings.Contains(htmlStr, "<!DOCTYPE html>") {
		t.Errorf("index.html missing DOCTYPE")
	}
	if !strings.Contains(htmlStr, "<link rel=\"stylesheet\" href=\"static/style.css\">") {
		t.Errorf("index.html missing CSS link")
	}
	if !strings.Contains(htmlStr, "<script src=\"static/script.js\">") {
		t.Errorf("index.html missing JS script tag")
	}
}

func TestExportCommand_HTMLWithAgents(t *testing.T) {
	// Setup test environment
	tempDir := t.TempDir()
	projectPath := filepath.Join(tempDir, "test-project")
	claudeDir := filepath.Join(tempDir, ".claude")

	if err := os.MkdirAll(projectPath, 0755); err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	// Create a test session with agents
	encodedPath := encoding.EncodePath(projectPath)
	projectDir := filepath.Join(claudeDir, "projects", encodedPath)
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("Failed to create Claude project directory: %v", err)
	}

	sessionID := "test-session-456"
	sessionFile := filepath.Join(projectDir, sessionID+".jsonl")

	// Create a session with a queue-operation that spawns an agent
	sessionContent := `{"uuid":"entry-1","type":"user","timestamp":"2026-02-01T10:00:00Z","message":[{"type":"text","text":"Hello"}]}
{"uuid":"entry-2","type":"queue-operation","timestamp":"2026-02-01T10:00:01Z","agentId":"agent-abc123"}
`
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0644); err != nil {
		t.Fatalf("Failed to write session file: %v", err)
	}

	// Create subagent directory and file
	sessionDir := filepath.Join(projectDir, sessionID)
	subagentsDir := filepath.Join(sessionDir, "subagents")
	if err := os.MkdirAll(subagentsDir, 0755); err != nil {
		t.Fatalf("Failed to create subagents directory: %v", err)
	}

	agentFile := filepath.Join(subagentsDir, "agent-abc123.jsonl")
	agentContent := `{"uuid":"agent-1","type":"user","timestamp":"2026-02-01T10:00:02Z","message":[{"type":"text","text":"Agent task"}]}
{"uuid":"agent-2","type":"assistant","timestamp":"2026-02-01T10:00:03Z","message":[{"type":"text","text":"Agent response"}]}
`
	if err := os.WriteFile(agentFile, []byte(agentContent), 0644); err != nil {
		t.Fatalf("Failed to write agent file: %v", err)
	}

	// Run export
	outputDir := filepath.Join(tempDir, "export-output-agents")
	opts := export.ExportOptions{
		OutputDir: outputDir,
		ClaudeDir: claudeDir,
	}

	result, err := export.ExportSession(projectPath, sessionID, opts)
	if err != nil {
		t.Fatalf("ExportSession failed: %v", err)
	}

	if result.TotalAgents != 1 {
		t.Errorf("Expected 1 agent, got %d", result.TotalAgents)
	}

	// Test HTML rendering
	if err := renderHTML(result, projectPath, projectDir, sessionID); err != nil {
		t.Fatalf("renderHTML failed: %v", err)
	}

	// Verify agent HTML fragment was created
	// Agent ID "agent-abc123" becomes "abc123" in the map, which is 6 chars, so stays "abc123"
	agentsDir := filepath.Join(outputDir, "agents")
	agentHTMLPath := filepath.Join(agentsDir, "abc123.html")
	if _, err := os.Stat(agentHTMLPath); os.IsNotExist(err) {
		t.Errorf("Agent HTML not created: %s", agentHTMLPath)
	}

	// Verify agent HTML message
	agentHTML, err := os.ReadFile(agentHTMLPath)
	if err != nil {
		t.Fatalf("Failed to read agent HTML: %v", err)
	}

	agentHTMLStr := string(agentHTML)
	if !strings.Contains(agentHTMLStr, "Agent task") {
		t.Errorf("Agent HTML missing expected content. Content: %s", agentHTMLStr)
	}
}

func TestRenderHTML(t *testing.T) {
	// Setup test environment
	tempDir := t.TempDir()
	projectPath := filepath.Join(tempDir, "test-project")
	claudeDir := filepath.Join(tempDir, ".claude")

	if err := os.MkdirAll(projectPath, 0755); err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	encodedPath := encoding.EncodePath(projectPath)
	projectDir := filepath.Join(claudeDir, "projects", encodedPath)
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("Failed to create Claude project directory: %v", err)
	}

	sessionID := "test-session-789"
	sessionFile := filepath.Join(projectDir, sessionID+".jsonl")

	// Create a minimal session
	sessionContent := `{"uuid":"entry-1","type":"user","timestamp":"2026-02-01T10:00:00Z","message":[{"type":"text","text":"Test"}]}
`
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0644); err != nil {
		t.Fatalf("Failed to write session file: %v", err)
	}

	// Export JSONL
	outputDir := filepath.Join(tempDir, "export-output-render")
	opts := export.ExportOptions{
		OutputDir: outputDir,
		ClaudeDir: claudeDir,
	}

	result, err := export.ExportSession(projectPath, sessionID, opts)
	if err != nil {
		t.Fatalf("ExportSession failed: %v", err)
	}

	// Test renderHTML directly
	if err := renderHTML(result, projectPath, projectDir, sessionID); err != nil {
		t.Errorf("renderHTML failed: %v", err)
	}

	// Verify outputs
	if _, err := os.Stat(filepath.Join(outputDir, "index.html")); os.IsNotExist(err) {
		t.Errorf("index.html not created")
	}
	if _, err := os.Stat(filepath.Join(outputDir, "static", "style.css")); os.IsNotExist(err) {
		t.Errorf("style.css not created")
	}
	if _, err := os.Stat(filepath.Join(outputDir, "static", "script.js")); os.IsNotExist(err) {
		t.Errorf("script.js not created")
	}
	if _, err := os.Stat(filepath.Join(outputDir, "manifest.json")); os.IsNotExist(err) {
		t.Errorf("manifest.json not created")
	}
}

func TestRenderAgentFragments(t *testing.T) {
	// Setup test environment
	tempDir := t.TempDir()
	outputDir := filepath.Join(tempDir, "export-output-fragments")
	sourceDir := filepath.Join(outputDir, "source")
	agentsDir := filepath.Join(sourceDir, "agents")

	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("Failed to create agents directory: %v", err)
	}

	// Create agent JSONL file
	agentID := "test-agent-xyz"
	agentFile := filepath.Join(agentsDir, "agent-"+agentID+".jsonl")
	agentContent := `{"uuid":"a-1","type":"user","timestamp":"2026-02-01T10:00:00Z","message":[{"type":"text","text":"Task"}]}
{"uuid":"a-2","type":"assistant","timestamp":"2026-02-01T10:00:01Z","message":[{"type":"text","text":"Done"}]}
`
	if err := os.WriteFile(agentFile, []byte(agentContent), 0644); err != nil {
		t.Fatalf("Failed to write agent file: %v", err)
	}

	// Create ExportResult
	result := &export.ExportResult{
		OutputDir: outputDir,
		SessionID: "test-session",
		AgentFiles: map[string]string{
			agentID: agentFile,
		},
	}

	// Test renderAgentFragments
	if err := renderAgentFragments(result, nil); err != nil {
		t.Errorf("renderAgentFragments failed: %v", err)
	}

	// Verify agent HTML was created
	// Agent ID "test-agent-xyz" becomes "test-age" when truncated to 8 chars
	agentHTMLPath := filepath.Join(outputDir, "agents", "test-age.html")
	if _, err := os.Stat(agentHTMLPath); os.IsNotExist(err) {
		t.Errorf("Agent HTML not created: %s", agentHTMLPath)
	}
}

func TestRenderAgentFragments_MissingFile(t *testing.T) {
	// Test with missing agent file (should add to errors but not fail)
	tempDir := t.TempDir()
	outputDir := filepath.Join(tempDir, "export-output-missing")

	result := &export.ExportResult{
		OutputDir: outputDir,
		SessionID: "test-session",
		AgentFiles: map[string]string{
			"missing-agent": "/nonexistent/path.jsonl",
		},
	}

	// Should return error for missing file
	err := renderAgentFragments(result, nil)
	if err == nil {
		t.Errorf("Expected error for missing agent file, got nil")
	}
}

func TestHTMLOutput_ValidStructure(t *testing.T) {
	// Setup test environment
	tempDir := t.TempDir()
	projectPath := filepath.Join(tempDir, "test-project")
	claudeDir := filepath.Join(tempDir, ".claude")

	if err := os.MkdirAll(projectPath, 0755); err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	encodedPath := encoding.EncodePath(projectPath)
	projectDir := filepath.Join(claudeDir, "projects", encodedPath)
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("Failed to create Claude project directory: %v", err)
	}

	sessionID := "test-session-valid"
	sessionFile := filepath.Join(projectDir, sessionID+".jsonl")

	// Create a session with various entry types
	sessionContent := `{"uuid":"e-1","type":"user","timestamp":"2026-02-01T10:00:00Z","message":[{"type":"text","text":"Test user message"}]}
{"uuid":"e-2","type":"assistant","timestamp":"2026-02-01T10:00:01Z","message":[{"type":"text","text":"Test response"},{"type":"tool_use","id":"tool-1","name":"Bash","input":{"command":"echo test"}}]}
{"uuid":"e-3","type":"user","timestamp":"2026-02-01T10:00:02Z","message":[{"type":"tool_result","tool_use_id":"tool-1","message":"test","is_error":false}]}
`
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0644); err != nil {
		t.Fatalf("Failed to write session file: %v", err)
	}

	// Export and render
	outputDir := filepath.Join(tempDir, "export-output-valid")
	opts := export.ExportOptions{
		OutputDir: outputDir,
		ClaudeDir: claudeDir,
	}

	result, err := export.ExportSession(projectPath, sessionID, opts)
	if err != nil {
		t.Fatalf("ExportSession failed: %v", err)
	}

	if err := renderHTML(result, projectPath, projectDir, sessionID); err != nil {
		t.Fatalf("renderHTML failed: %v", err)
	}

	// Read and validate HTML structure
	indexPath := filepath.Join(outputDir, "index.html")
	htmlContent, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("Failed to read index.html: %v", err)
	}

	htmlStr := string(htmlContent)

	// Check for expected HTML structure
	expectedElements := []string{
		"<!DOCTYPE html>",
		"<html>",
		"<head>",
		"<title>Claude Conversation Export</title>",
		"<link rel=\"stylesheet\" href=\"static/style.css\">",
		"<body>",
		"<div class=\"conversation\">",
		"<div class=\"entry user\"",
		"<div class=\"entry assistant\"",
		"<div class=\"tool-call\"",
		"<script src=\"static/script.js\">",
		"</body>",
		"</html>",
	}

	for _, elem := range expectedElements {
		if !strings.Contains(htmlStr, elem) {
			t.Errorf("HTML missing expected element: %s", elem)
		}
	}

	// Check for content presence
	if !strings.Contains(htmlStr, "Test user message") {
		t.Errorf("HTML missing user message content")
	}
	if !strings.Contains(htmlStr, "Test response") {
		t.Errorf("HTML missing assistant response content")
	}
	if !strings.Contains(htmlStr, "Bash") {
		t.Errorf("HTML missing tool name")
	}
}

func TestTruncateAgentID(t *testing.T) {
	tests := []struct {
		name     string
		agentID  string
		expected string
	}{
		{
			name:     "long agent ID",
			agentID:  "agent-12345678-1234-1234-1234-123456789012",
			expected: "agent-12",
		},
		{
			name:     "short agent ID",
			agentID:  "abc123",
			expected: "abc123",
		},
		{
			name:     "exactly 8 chars",
			agentID:  "12345678",
			expected: "12345678",
		},
		{
			name:     "empty string",
			agentID:  "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateAgentID(tt.agentID)
			if result != tt.expected {
				t.Errorf("truncateAgentID(%q) = %q, expected %q", tt.agentID, result, tt.expected)
			}
		})
	}
}
