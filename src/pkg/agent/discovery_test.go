package agent

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/randlee/claude-history/pkg/models"
)

// createTestEntry creates a JSONL entry for testing.
func createTestEntry(t *testing.T, uuid, sessionID, entryType string, timestamp time.Time, toolCalls []map[string]any) string {
	t.Helper()

	entry := map[string]any{
		"uuid":      uuid,
		"sessionId": sessionID,
		"type":      entryType,
		"timestamp": timestamp.Format(time.RFC3339Nano),
	}

	if len(toolCalls) > 0 {
		content := make([]map[string]any, 0, len(toolCalls))
		for _, tc := range toolCalls {
			content = append(content, map[string]any{
				"type":  "tool_use",
				"id":    tc["id"],
				"name":  tc["name"],
				"input": tc["input"],
			})
		}
		entry["message"] = map[string]any{
			"role":    "assistant",
			"content": content,
		}
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("Failed to marshal entry: %v", err)
	}

	return string(data) + "\n"
}

// createReadToolCall creates a Read tool call for testing.
func createReadToolCall(filePath string) map[string]any {
	return map[string]any{
		"id":    "tool_read_1",
		"name":  "Read",
		"input": map[string]any{"file_path": filePath},
	}
}

// createWriteToolCall creates a Write tool call for testing.
func createWriteToolCall(filePath string) map[string]any {
	return map[string]any{
		"id":    "tool_write_1",
		"name":  "Write",
		"input": map[string]any{"file_path": filePath, "content": "test"},
	}
}

// createEditToolCall creates an Edit tool call for testing.
func createEditToolCall(filePath string) map[string]any {
	return map[string]any{
		"id":    "tool_edit_1",
		"name":  "Edit",
		"input": map[string]any{"file_path": filePath, "old_string": "a", "new_string": "b"},
	}
}

// createBashToolCall creates a Bash tool call for testing.
func createBashToolCall(command string) map[string]any {
	return map[string]any{
		"id":    "tool_bash_1",
		"name":  "Bash",
		"input": map[string]any{"command": command},
	}
}

// createGrepToolCall creates a Grep tool call for testing.
func createGrepToolCall(pattern, path string) map[string]any {
	return map[string]any{
		"id":    "tool_grep_1",
		"name":  "Grep",
		"input": map[string]any{"pattern": pattern, "path": path},
	}
}

func TestMatchesFilePattern(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		pattern  string
		expected bool
	}{
		// Exact path match
		{
			name:     "exact path match",
			path:     "/src/main.go",
			pattern:  "/src/main.go",
			expected: true,
		},
		// Single wildcard patterns
		{
			name:     "single wildcard matches go files",
			path:     "main.go",
			pattern:  "*.go",
			expected: true,
		},
		{
			name:     "single wildcard no match",
			path:     "main.ts",
			pattern:  "*.go",
			expected: false,
		},
		{
			name:     "single wildcard matches in path",
			path:     "/src/pkg/main.go",
			pattern:  "*.go",
			expected: true,
		},
		// Double wildcard patterns
		{
			name:     "double wildcard matches nested",
			path:     "/src/pkg/models/entry.go",
			pattern:  "**/*.go",
			expected: true,
		},
		{
			name:     "double wildcard matches shallow",
			path:     "/main.go",
			pattern:  "**/*.go",
			expected: true,
		},
		{
			name:     "double wildcard no match",
			path:     "/src/pkg/models/entry.ts",
			pattern:  "**/*.go",
			expected: false,
		},
		{
			name:     "double wildcard with prefix",
			path:     "src/pkg/main.go",
			pattern:  "src/**/*.go",
			expected: true,
		},
		// Edge cases
		{
			name:     "empty pattern returns false",
			path:     "/src/main.go",
			pattern:  "",
			expected: false,
		},
		{
			name:     "question mark wildcard",
			path:     "test1.go",
			pattern:  "test?.go",
			expected: true,
		},
		// Invalid pattern handled gracefully
		{
			name:     "invalid pattern returns false",
			path:     "/src/main.go",
			pattern:  "[invalid",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesFilePattern(tt.path, tt.pattern)
			if result != tt.expected {
				t.Errorf("matchesFilePattern(%q, %q) = %v, want %v",
					tt.path, tt.pattern, result, tt.expected)
			}
		})
	}
}

func TestFindAgents_ExactFilePath(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "12345678-1234-1234-1234-123456789012"

	// Create main session file with Read tool call
	sessionFile := filepath.Join(tmpDir, sessionID+".jsonl")
	baseTime := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)

	content := createTestEntry(t, "1", sessionID, "user", baseTime, nil)
	content += createTestEntry(t, "2", sessionID, "assistant", baseTime.Add(time.Minute), []map[string]any{
		createReadToolCall("/src/main.go"),
	})

	mustWriteFile(t, sessionFile, []byte(content))

	matches, err := FindAgents(tmpDir, FindAgentsOptions{
		ExploredPattern: "/src/main.go",
	})

	if err != nil {
		t.Fatalf("FindAgents() error: %v", err)
	}

	if len(matches) != 1 {
		t.Fatalf("FindAgents() returned %d matches, want 1", len(matches))
	}

	if len(matches[0].MatchedFiles) != 1 || matches[0].MatchedFiles[0] != "/src/main.go" {
		t.Errorf("MatchedFiles = %v, want [/src/main.go]", matches[0].MatchedFiles)
	}
}

func TestFindAgents_GlobPattern(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "12345678-1234-1234-1234-123456789012"

	sessionFile := filepath.Join(tmpDir, sessionID+".jsonl")
	baseTime := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)

	content := createTestEntry(t, "1", sessionID, "user", baseTime, nil)
	content += createTestEntry(t, "2", sessionID, "assistant", baseTime.Add(time.Minute), []map[string]any{
		createReadToolCall("/src/main.go"),
		createWriteToolCall("/src/util.go"),
	})

	mustWriteFile(t, sessionFile, []byte(content))

	matches, err := FindAgents(tmpDir, FindAgentsOptions{
		ExploredPattern: "*.go",
	})

	if err != nil {
		t.Fatalf("FindAgents() error: %v", err)
	}

	if len(matches) != 1 {
		t.Fatalf("FindAgents() returned %d matches, want 1", len(matches))
	}

	if len(matches[0].MatchedFiles) != 2 {
		t.Errorf("MatchedFiles has %d files, want 2", len(matches[0].MatchedFiles))
	}
}

func TestFindAgents_DoubleStarGlob(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "12345678-1234-1234-1234-123456789012"

	sessionFile := filepath.Join(tmpDir, sessionID+".jsonl")
	baseTime := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)

	content := createTestEntry(t, "1", sessionID, "user", baseTime, nil)
	content += createTestEntry(t, "2", sessionID, "assistant", baseTime.Add(time.Minute), []map[string]any{
		createReadToolCall("/src/pkg/models/entry.go"),
	})

	mustWriteFile(t, sessionFile, []byte(content))

	matches, err := FindAgents(tmpDir, FindAgentsOptions{
		ExploredPattern: "**/*.go",
	})

	if err != nil {
		t.Fatalf("FindAgents() error: %v", err)
	}

	if len(matches) != 1 {
		t.Fatalf("FindAgents() returned %d matches, want 1", len(matches))
	}
}

func TestFindAgents_TimeRange(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "12345678-1234-1234-1234-123456789012"

	sessionFile := filepath.Join(tmpDir, sessionID+".jsonl")
	baseTime := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)

	// Create entries spanning multiple days
	content := createTestEntry(t, "1", sessionID, "user", baseTime, nil)
	content += createTestEntry(t, "2", sessionID, "assistant", baseTime.Add(time.Minute), []map[string]any{
		createReadToolCall("/early.go"),
	})

	mustWriteFile(t, sessionFile, []byte(content))

	// Query for time range that includes the entries
	startTime := baseTime.Add(-time.Hour)
	endTime := baseTime.Add(2 * time.Hour)

	matches, err := FindAgents(tmpDir, FindAgentsOptions{
		StartTime:       startTime,
		EndTime:         endTime,
		ExploredPattern: "*.go",
	})

	if err != nil {
		t.Fatalf("FindAgents() error: %v", err)
	}

	if len(matches) != 1 {
		t.Fatalf("FindAgents() returned %d matches, want 1", len(matches))
	}

	// Query for time range that excludes the entries
	matches, err = FindAgents(tmpDir, FindAgentsOptions{
		StartTime:       baseTime.Add(time.Hour),
		EndTime:         baseTime.Add(2 * time.Hour),
		ExploredPattern: "*.go",
	})

	if err != nil {
		t.Fatalf("FindAgents() error: %v", err)
	}

	// Should not match because the entry is outside the time range
	if len(matches) != 0 {
		t.Errorf("FindAgents() with out-of-range time returned %d matches, want 0", len(matches))
	}
}

func TestFindAgents_ToolType(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "12345678-1234-1234-1234-123456789012"

	sessionFile := filepath.Join(tmpDir, sessionID+".jsonl")
	baseTime := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)

	content := createTestEntry(t, "1", sessionID, "user", baseTime, nil)
	content += createTestEntry(t, "2", sessionID, "assistant", baseTime.Add(time.Minute), []map[string]any{
		createBashToolCall("go build"),
	})

	mustWriteFile(t, sessionFile, []byte(content))

	// Search for Bash tool
	matches, err := FindAgents(tmpDir, FindAgentsOptions{
		ToolTypes: []string{"Bash"},
	})

	if err != nil {
		t.Fatalf("FindAgents() error: %v", err)
	}

	if len(matches) != 1 {
		t.Fatalf("FindAgents() returned %d matches, want 1", len(matches))
	}

	// Search for non-existent tool
	matches, err = FindAgents(tmpDir, FindAgentsOptions{
		ToolTypes: []string{"WebSearch"},
	})

	if err != nil {
		t.Fatalf("FindAgents() error: %v", err)
	}

	if len(matches) != 0 {
		t.Errorf("FindAgents() for non-existent tool returned %d matches, want 0", len(matches))
	}
}

func TestFindAgents_ToolInputRegex(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "12345678-1234-1234-1234-123456789012"

	sessionFile := filepath.Join(tmpDir, sessionID+".jsonl")
	baseTime := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)

	content := createTestEntry(t, "1", sessionID, "user", baseTime, nil)
	content += createTestEntry(t, "2", sessionID, "assistant", baseTime.Add(time.Minute), []map[string]any{
		createBashToolCall("go test ./..."),
	})

	mustWriteFile(t, sessionFile, []byte(content))

	// Match regex against tool input
	matches, err := FindAgents(tmpDir, FindAgentsOptions{
		ToolMatch: "go test",
	})

	if err != nil {
		t.Fatalf("FindAgents() error: %v", err)
	}

	if len(matches) != 1 {
		t.Fatalf("FindAgents() returned %d matches, want 1", len(matches))
	}

	// Non-matching regex
	matches, err = FindAgents(tmpDir, FindAgentsOptions{
		ToolMatch: "npm install",
	})

	if err != nil {
		t.Fatalf("FindAgents() error: %v", err)
	}

	if len(matches) != 0 {
		t.Errorf("FindAgents() with non-matching regex returned %d matches, want 0", len(matches))
	}
}

func TestFindAgents_CombinedFilters(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "12345678-1234-1234-1234-123456789012"

	sessionFile := filepath.Join(tmpDir, sessionID+".jsonl")
	baseTime := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)

	content := createTestEntry(t, "1", sessionID, "user", baseTime, nil)
	content += createTestEntry(t, "2", sessionID, "assistant", baseTime.Add(time.Minute), []map[string]any{
		createReadToolCall("/src/main.go"),
		createBashToolCall("go build"),
	})

	mustWriteFile(t, sessionFile, []byte(content))

	// Combined: file pattern + tool type + time range
	matches, err := FindAgents(tmpDir, FindAgentsOptions{
		ExploredPattern: "*.go",
		ToolTypes:       []string{"Bash"},
		StartTime:       baseTime.Add(-time.Hour),
		EndTime:         baseTime.Add(2 * time.Hour),
	})

	if err != nil {
		t.Fatalf("FindAgents() error: %v", err)
	}

	if len(matches) != 1 {
		t.Fatalf("FindAgents() returned %d matches, want 1", len(matches))
	}
}

func TestFindAgents_NestedAgents(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "12345678-1234-1234-1234-123456789012"
	agentID := "a12eb64"

	// Create main session file
	sessionFile := filepath.Join(tmpDir, sessionID+".jsonl")
	baseTime := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)

	mainContent := createTestEntry(t, "1", sessionID, "user", baseTime, nil)
	mainContent += createTestEntry(t, "2", sessionID, "assistant", baseTime.Add(time.Minute), []map[string]any{
		createReadToolCall("/main.go"),
	})
	mainContent += createTestEntry(t, "3", sessionID, "queue-operation", baseTime.Add(2*time.Minute), nil)

	mustWriteFile(t, sessionFile, []byte(mainContent))

	// Create subagent directory and file
	sessionDir := filepath.Join(tmpDir, sessionID)
	subagentsDir := filepath.Join(sessionDir, "subagents")
	mustMkdirAll(t, subagentsDir)

	agentContent := createTestEntry(t, "a1", sessionID, "user", baseTime.Add(3*time.Minute), nil)
	agentContent += createTestEntry(t, "a2", sessionID, "assistant", baseTime.Add(4*time.Minute), []map[string]any{
		createReadToolCall("/agent.go"),
	})

	mustWriteFile(t, filepath.Join(subagentsDir, "agent-"+agentID+".jsonl"), []byte(agentContent))

	// Search for all agents with .go files
	matches, err := FindAgents(tmpDir, FindAgentsOptions{
		ExploredPattern: "*.go",
	})

	if err != nil {
		t.Fatalf("FindAgents() error: %v", err)
	}

	// Should find both main session and subagent
	if len(matches) != 2 {
		t.Errorf("FindAgents() returned %d matches, want 2", len(matches))
	}

	// Verify we found the agent
	foundAgent := false
	for _, m := range matches {
		if m.AgentID == agentID {
			foundAgent = true
			break
		}
	}
	if !foundAgent {
		t.Error("FindAgents() did not find the subagent")
	}
}

func TestFindAgents_NoMatchesReturnsEmptySlice(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "12345678-1234-1234-1234-123456789012"

	sessionFile := filepath.Join(tmpDir, sessionID+".jsonl")
	baseTime := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)

	// Create session without matching files
	content := createTestEntry(t, "1", sessionID, "user", baseTime, nil)
	content += createTestEntry(t, "2", sessionID, "assistant", baseTime.Add(time.Minute), []map[string]any{
		createReadToolCall("/src/main.ts"),
	})

	mustWriteFile(t, sessionFile, []byte(content))

	matches, err := FindAgents(tmpDir, FindAgentsOptions{
		ExploredPattern: "*.go",
	})

	if err != nil {
		t.Fatalf("FindAgents() error: %v", err)
	}

	// Should return empty slice, not nil
	if matches == nil {
		t.Error("FindAgents() returned nil, want empty slice")
	}

	if len(matches) != 0 {
		t.Errorf("FindAgents() returned %d matches, want 0", len(matches))
	}
}

func TestFindAgents_InvalidProjectDirectory(t *testing.T) {
	_, err := FindAgents("/nonexistent/path/that/does/not/exist", FindAgentsOptions{})

	if err == nil {
		t.Error("FindAgents() with invalid directory should return error")
	}
}

func TestFindAgents_InvalidRegex(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := FindAgents(tmpDir, FindAgentsOptions{
		ToolMatch: "[invalid regex(",
	})

	if err == nil {
		t.Error("FindAgents() with invalid regex should return error")
	}
}

func TestFindAgents_SessionScope(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID1 := "12345678-1234-1234-1234-123456789012"
	sessionID2 := "87654321-4321-4321-4321-210987654321"

	baseTime := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)

	// Create two sessions
	content1 := createTestEntry(t, "1", sessionID1, "user", baseTime, nil)
	content1 += createTestEntry(t, "2", sessionID1, "assistant", baseTime.Add(time.Minute), []map[string]any{
		createReadToolCall("/session1.go"),
	})
	mustWriteFile(t, filepath.Join(tmpDir, sessionID1+".jsonl"), []byte(content1))

	content2 := createTestEntry(t, "1", sessionID2, "user", baseTime, nil)
	content2 += createTestEntry(t, "2", sessionID2, "assistant", baseTime.Add(time.Minute), []map[string]any{
		createReadToolCall("/session2.go"),
	})
	mustWriteFile(t, filepath.Join(tmpDir, sessionID2+".jsonl"), []byte(content2))

	// Search with session scope
	matches, err := FindAgents(tmpDir, FindAgentsOptions{
		SessionID:       sessionID1,
		ExploredPattern: "*.go",
	})

	if err != nil {
		t.Fatalf("FindAgents() error: %v", err)
	}

	if len(matches) != 1 {
		t.Fatalf("FindAgents() with session scope returned %d matches, want 1", len(matches))
	}

	if matches[0].SessionID != sessionID1 {
		t.Errorf("SessionID = %q, want %q", matches[0].SessionID, sessionID1)
	}
}

func TestFindAgents_MultipleToolTypes(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "12345678-1234-1234-1234-123456789012"

	sessionFile := filepath.Join(tmpDir, sessionID+".jsonl")
	baseTime := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)

	content := createTestEntry(t, "1", sessionID, "user", baseTime, nil)
	content += createTestEntry(t, "2", sessionID, "assistant", baseTime.Add(time.Minute), []map[string]any{
		createGrepToolCall("func main", "/src"),
	})

	mustWriteFile(t, sessionFile, []byte(content))

	// Search for multiple tool types (should match on any)
	matches, err := FindAgents(tmpDir, FindAgentsOptions{
		ToolTypes: []string{"Grep", "Bash"},
	})

	if err != nil {
		t.Fatalf("FindAgents() error: %v", err)
	}

	if len(matches) != 1 {
		t.Fatalf("FindAgents() returned %d matches, want 1", len(matches))
	}
}

func TestFindAgents_CaseInsensitiveToolMatch(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "12345678-1234-1234-1234-123456789012"

	sessionFile := filepath.Join(tmpDir, sessionID+".jsonl")
	baseTime := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)

	content := createTestEntry(t, "1", sessionID, "user", baseTime, nil)
	content += createTestEntry(t, "2", sessionID, "assistant", baseTime.Add(time.Minute), []map[string]any{
		createBashToolCall("echo test"),
	})

	mustWriteFile(t, sessionFile, []byte(content))

	// Search with different case (should still match)
	matches, err := FindAgents(tmpDir, FindAgentsOptions{
		ToolTypes: []string{"bash"}, // lowercase
	})

	if err != nil {
		t.Fatalf("FindAgents() error: %v", err)
	}

	if len(matches) != 1 {
		t.Errorf("FindAgents() with lowercase tool type returned %d matches, want 1", len(matches))
	}
}

func TestIsFileOperationTool(t *testing.T) {
	tests := []struct {
		toolName string
		expected bool
	}{
		{"Read", true},
		{"read", true},
		{"Write", true},
		{"Edit", true},
		{"Glob", true},
		{"Grep", true},
		{"NotebookEdit", true},
		{"Bash", false},
		{"WebSearch", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.toolName, func(t *testing.T) {
			result := isFileOperationTool(tt.toolName)
			if result != tt.expected {
				t.Errorf("isFileOperationTool(%q) = %v, want %v", tt.toolName, result, tt.expected)
			}
		})
	}
}

func TestExtractFilePathFromTool(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected string
	}{
		{
			name:     "file_path field",
			input:    map[string]any{"file_path": "/src/main.go"},
			expected: "/src/main.go",
		},
		{
			name:     "path field",
			input:    map[string]any{"path": "/src/util.go"},
			expected: "/src/util.go",
		},
		{
			name:     "no path field",
			input:    map[string]any{"command": "ls"},
			expected: "",
		},
		{
			name:     "nil input",
			input:    nil,
			expected: "",
		},
		{
			name:     "non-string path",
			input:    map[string]any{"path": 123},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := models.ToolUse{Input: tt.input}
			result := extractFilePathFromTool(tool)
			if result != tt.expected {
				t.Errorf("extractFilePathFromTool() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestPathError(t *testing.T) {
	err := &PathError{
		Path: "/some/path",
		Op:   "find agents",
		Err:  "directory does not exist",
	}

	expected := "find agents /some/path: directory does not exist"
	if err.Error() != expected {
		t.Errorf("PathError.Error() = %q, want %q", err.Error(), expected)
	}
}

func TestPatternError(t *testing.T) {
	err := &PatternError{
		Pattern: "[invalid",
		Err:     "invalid regex",
	}

	expected := "pattern [invalid: invalid regex"
	if err.Error() != expected {
		t.Errorf("PatternError.Error() = %q, want %q", err.Error(), expected)
	}
}

// Ensure models package is used
var _ = models.ToolUse{}
