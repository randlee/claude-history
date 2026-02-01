package output

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/randlee/claude-history/pkg/models"
)

func TestFormatToolCall(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		input    map[string]any
		expected string
	}{
		{
			name:     "Bash with command",
			toolName: "Bash",
			input:    map[string]any{"command": "git status"},
			expected: "[Bash] git status",
		},
		{
			name:     "Read with file_path",
			toolName: "Read",
			input:    map[string]any{"file_path": "/path/to/file.go"},
			expected: "[Read] /path/to/file.go",
		},
		{
			name:     "Write with file_path",
			toolName: "Write",
			input:    map[string]any{"file_path": "/path/to/output.txt"},
			expected: "[Write] /path/to/output.txt",
		},
		{
			name:     "Edit with file_path",
			toolName: "Edit",
			input:    map[string]any{"file_path": "/src/main.go"},
			expected: "[Edit] /src/main.go",
		},
		{
			name:     "Grep with pattern",
			toolName: "Grep",
			input:    map[string]any{"pattern": "func.*Test"},
			expected: "[Grep] func.*Test",
		},
		{
			name:     "Glob with pattern",
			toolName: "Glob",
			input:    map[string]any{"pattern": "**/*.go"},
			expected: "[Glob] **/*.go",
		},
		{
			name:     "Task with description",
			toolName: "Task",
			input:    map[string]any{"description": "Analyze the codebase"},
			expected: "[Task] Analyze the codebase",
		},
		{
			name:     "Task with prompt (fallback)",
			toolName: "Task",
			input:    map[string]any{"prompt": "Review this code"},
			expected: "[Task] Review this code",
		},
		{
			name:     "Unknown tool with JSON fallback",
			toolName: "CustomTool",
			input:    map[string]any{"key": "value"},
			expected: `[CustomTool] {"key":"value"}`,
		},
		{
			name:     "Tool with nil input",
			toolName: "Bash",
			input:    nil,
			expected: "[Bash]",
		},
		{
			name:     "Tool with empty input",
			toolName: "Read",
			input:    map[string]any{},
			expected: "[Read]",
		},
		{
			name:     "Tool with missing expected key",
			toolName: "Bash",
			input:    map[string]any{"other_key": "value"},
			expected: `[Bash] {"other_key":"value"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatToolCall(tt.toolName, tt.input)
			if result != tt.expected {
				t.Errorf("FormatToolCall(%q, %v) = %q, want %q", tt.toolName, tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatToolCallTruncation(t *testing.T) {
	// Create a command longer than 80 characters
	longCommand := strings.Repeat("a", 100)
	input := map[string]any{"command": longCommand}

	result := FormatToolCall("Bash", input)

	// Should be "[Bash] " (7 chars) + 77 chars + "..." = 87 total for display value part
	// The display value should be truncated to 80 chars including "..."
	expectedPrefix := "[Bash] " + strings.Repeat("a", 77) + "..."

	if result != expectedPrefix {
		t.Errorf("FormatToolCall with long input = %q (len=%d), want %q (len=%d)",
			result, len(result), expectedPrefix, len(expectedPrefix))
	}

	// Verify the display value (after tool name) is exactly 80 chars
	displayValue := strings.TrimPrefix(result, "[Bash] ")
	if len(displayValue) != maxToolInputLength {
		t.Errorf("Display value length = %d, want %d", len(displayValue), maxToolInputLength)
	}
}

func TestFormatToolCallTruncationEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		inputLength int
		shouldTrunc bool
		expectedLen int
	}{
		{
			name:        "Exactly 80 chars - no truncation",
			inputLength: 80,
			shouldTrunc: false,
			expectedLen: 80,
		},
		{
			name:        "79 chars - no truncation",
			inputLength: 79,
			shouldTrunc: false,
			expectedLen: 79,
		},
		{
			name:        "81 chars - truncation",
			inputLength: 81,
			shouldTrunc: true,
			expectedLen: 80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := map[string]any{"command": strings.Repeat("x", tt.inputLength)}
			result := FormatToolCall("Bash", input)

			displayValue := strings.TrimPrefix(result, "[Bash] ")
			hasTrunc := strings.HasSuffix(displayValue, "...")

			if hasTrunc != tt.shouldTrunc {
				t.Errorf("Truncation = %v, want %v", hasTrunc, tt.shouldTrunc)
			}

			if len(displayValue) != tt.expectedLen {
				t.Errorf("Display value length = %d, want %d", len(displayValue), tt.expectedLen)
			}
		})
	}
}

func TestFormatToolCalls(t *testing.T) {
	tests := []struct {
		name     string
		tools    []ToolUse
		expected string
	}{
		{
			name:     "Empty tools",
			tools:    []ToolUse{},
			expected: "",
		},
		{
			name:     "Nil tools",
			tools:    nil,
			expected: "",
		},
		{
			name: "Single tool",
			tools: []ToolUse{
				{Name: "Bash", Input: map[string]any{"command": "ls -la"}},
			},
			expected: "[Bash] ls -la",
		},
		{
			name: "Multiple tools",
			tools: []ToolUse{
				{Name: "Bash", Input: map[string]any{"command": "git status"}},
				{Name: "Read", Input: map[string]any{"file_path": "/path/to/file.go"}},
				{Name: "Write", Input: map[string]any{"file_path": "/output.txt"}},
			},
			expected: "[Bash] git status\n[Read] /path/to/file.go\n[Write] /output.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatToolCalls(tt.tools)
			if result != tt.expected {
				t.Errorf("FormatToolCalls() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFormatToolSummary(t *testing.T) {
	tests := []struct {
		name     string
		tools    []ToolUse
		expected string
	}{
		{
			name:     "Empty tools",
			tools:    []ToolUse{},
			expected: "",
		},
		{
			name:     "Nil tools",
			tools:    nil,
			expected: "",
		},
		{
			name: "Single tool - shows full format",
			tools: []ToolUse{
				{Name: "Bash", Input: map[string]any{"command": "git status"}},
			},
			expected: "[Bash] git status",
		},
		{
			name: "Two tools - shows names only",
			tools: []ToolUse{
				{Name: "Bash", Input: map[string]any{"command": "git status"}},
				{Name: "Read", Input: map[string]any{"file_path": "/path/to/file.go"}},
			},
			expected: "[Bash, Read]",
		},
		{
			name: "Multiple tools - shows names only",
			tools: []ToolUse{
				{Name: "Bash", Input: map[string]any{}},
				{Name: "Read", Input: map[string]any{}},
				{Name: "Write", Input: map[string]any{}},
				{Name: "Edit", Input: map[string]any{}},
			},
			expected: "[Bash, Read, Write, Edit]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatToolSummary(tt.tools)
			if result != tt.expected {
				t.Errorf("FormatToolSummary() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestExtractToolDisplayValue(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		input    map[string]any
		expected string
	}{
		{
			name:     "Bash extracts command",
			toolName: "Bash",
			input:    map[string]any{"command": "npm install", "timeout": 30000},
			expected: "npm install",
		},
		{
			name:     "Read extracts file_path",
			toolName: "Read",
			input:    map[string]any{"file_path": "/src/main.go", "offset": 0},
			expected: "/src/main.go",
		},
		{
			name:     "Task prefers description over prompt",
			toolName: "Task",
			input:    map[string]any{"description": "Desc", "prompt": "Prompt"},
			expected: "Desc",
		},
		{
			name:     "Task falls back to prompt",
			toolName: "Task",
			input:    map[string]any{"prompt": "Review this"},
			expected: "Review this",
		},
		{
			name:     "Nil input returns empty",
			toolName: "Bash",
			input:    nil,
			expected: "",
		},
		{
			name:     "Empty input returns empty",
			toolName: "Bash",
			input:    map[string]any{},
			expected: "",
		},
		{
			name:     "Wrong type returns JSON",
			toolName: "Bash",
			input:    map[string]any{"command": 123},
			expected: `{"command":123}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractToolDisplayValue(tt.toolName, tt.input)
			if result != tt.expected {
				t.Errorf("extractToolDisplayValue(%q, %v) = %q, want %q",
					tt.toolName, tt.input, result, tt.expected)
			}
		})
	}
}

func TestSerializeInput(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected string
	}{
		{
			name:     "Empty map returns empty string",
			input:    map[string]any{},
			expected: "",
		},
		{
			name:     "Simple map",
			input:    map[string]any{"key": "value"},
			expected: `{"key":"value"}`,
		},
		{
			name:     "Multiple keys",
			input:    map[string]any{"a": 1, "b": "two"},
			expected: "", // Will check contains instead due to map ordering
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := serializeInput(tt.input)
			if tt.name == "Multiple keys" {
				// Map ordering is not guaranteed, so just check it contains expected parts
				if !strings.Contains(result, `"a":1`) || !strings.Contains(result, `"b":"two"`) {
					t.Errorf("serializeInput() = %q, expected to contain 'a':1 and 'b':'two'", result)
				}
			} else {
				if result != tt.expected {
					t.Errorf("serializeInput(%v) = %q, want %q", tt.input, result, tt.expected)
				}
			}
		})
	}
}

// Existing formatter tests
func TestParseFormat(t *testing.T) {
	tests := []struct {
		input    string
		expected Format
	}{
		{"json", FormatJSON},
		{"JSON", FormatJSON},
		{"list", FormatList},
		{"summary", FormatSummary},
		{"ascii", FormatASCII},
		{"dot", FormatDOT},
		{"path", FormatPath},
		{"unknown", FormatList},
		{"", FormatList},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseFormat(tt.input)
			if result != tt.expected {
				t.Errorf("ParseFormat(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestWriteJSON(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{"key": "value"}

	err := WriteJSON(&buf, data)
	if err != nil {
		t.Fatalf("WriteJSON() error = %v", err)
	}

	expected := "{\n  \"key\": \"value\"\n}\n"
	if buf.String() != expected {
		t.Errorf("WriteJSON() = %q, want %q", buf.String(), expected)
	}
}

func TestWriteList(t *testing.T) {
	var buf bytes.Buffer
	items := []string{"item1", "item2", "item3"}

	WriteList(&buf, items)

	expected := "item1\nitem2\nitem3\n"
	if buf.String() != expected {
		t.Errorf("WriteList() = %q, want %q", buf.String(), expected)
	}
}

func TestWriteSessions(t *testing.T) {
	sessions := []models.Session{
		{
			ID:           "session-123",
			Modified:     time.Date(2026, 1, 31, 10, 30, 0, 0, time.UTC),
			MessageCount: 5,
			FirstPrompt:  "Hello world",
		},
	}

	t.Run("list format", func(t *testing.T) {
		var buf bytes.Buffer
		err := WriteSessions(&buf, sessions, FormatList)
		if err != nil {
			t.Fatalf("WriteSessions() error = %v", err)
		}

		result := buf.String()
		if !strings.Contains(result, "session-123") {
			t.Error("Expected session ID in output")
		}
		if !strings.Contains(result, "5 msgs") {
			t.Error("Expected message count in output")
		}
	})

	t.Run("json format", func(t *testing.T) {
		var buf bytes.Buffer
		err := WriteSessions(&buf, sessions, FormatJSON)
		if err != nil {
			t.Fatalf("WriteSessions() error = %v", err)
		}

		result := buf.String()
		if !strings.Contains(result, "\"sessionId\": \"session-123\"") {
			t.Error("Expected JSON formatted session ID")
		}
	})
}

func TestWritePath(t *testing.T) {
	var buf bytes.Buffer
	WritePath(&buf, "/path/to/project")

	expected := "/path/to/project\n"
	if buf.String() != expected {
		t.Errorf("WritePath() = %q, want %q", buf.String(), expected)
	}
}
