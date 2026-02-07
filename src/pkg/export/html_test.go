package export

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/randlee/claude-history/pkg/agent"
	"github.com/randlee/claude-history/pkg/models"
)

func TestRenderConversation_BasicStructure(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-001",
			SessionID: "session-001",
			Type:      models.EntryTypeUser,
			Timestamp: "2026-01-31T10:00:00Z",
			Message:   json.RawMessage(`"Hello, Claude!"`),
		},
		{
			UUID:      "uuid-002",
			SessionID: "session-001",
			Type:      models.EntryTypeAssistant,
			Timestamp: "2026-01-31T10:00:05Z",
			Message:   json.RawMessage(`{"role": "assistant", "content": [{"type": "text", "text": "Hello! How can I help you?"}]}`),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	// Check basic structure
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("HTML missing DOCTYPE")
	}
	if !strings.Contains(html, "<html>") {
		t.Error("HTML missing <html> tag")
	}
	if !strings.Contains(html, `<meta charset="UTF-8">`) {
		t.Error("HTML missing charset meta tag")
	}
	if !strings.Contains(html, `<title>Claude Conversation Export</title>`) {
		t.Error("HTML missing title")
	}
	if !strings.Contains(html, `<link rel="stylesheet" href="static/style.css">`) {
		t.Error("HTML missing stylesheet link")
	}
	if !strings.Contains(html, `<script src="static/script.js"></script>`) {
		t.Error("HTML missing script tag")
	}
	if !strings.Contains(html, `<div class="conversation">`) {
		t.Error("HTML missing conversation container")
	}
}

func TestRenderConversation_UserMessage(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-001",
			SessionID: "session-001",
			Type:      models.EntryTypeUser,
			Timestamp: "2026-01-31T10:00:00Z",
			Message:   json.RawMessage(`"What is Go?"`),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	if !strings.Contains(html, `class="message-row user"`) {
		t.Error("HTML missing message-row user class")
	}
	if !strings.Contains(html, `data-uuid="uuid-001"`) {
		t.Error("HTML missing entry UUID attribute")
	}
	if !strings.Contains(html, "What is Go?") {
		t.Error("HTML missing user message content")
	}
}

func TestRenderConversation_AssistantMessage(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-002",
			SessionID: "session-001",
			Type:      models.EntryTypeAssistant,
			Timestamp: "2026-01-31T10:00:05Z",
			Message:   json.RawMessage(`{"role": "assistant", "content": [{"type": "text", "text": "Go is a programming language."}]}`),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	if !strings.Contains(html, `class="message-row assistant"`) {
		t.Error("HTML missing message-row assistant class")
	}
	if !strings.Contains(html, "Go is a programming language.") {
		t.Error("HTML missing assistant message content")
	}
}

func TestRenderConversation_SystemMessage(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-003",
			SessionID: "session-001",
			Type:      models.EntryTypeSystem,
			Timestamp: "2026-01-31T10:00:00Z",
			Message:   json.RawMessage(`"System initialized"`),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	if !strings.Contains(html, `class="message-row system"`) {
		t.Error("HTML missing message-row system class")
	}
}

func TestRenderConversation_ToolCall(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-001",
			SessionID: "session-001",
			Type:      models.EntryTypeAssistant,
			Timestamp: "2026-01-31T10:00:00Z",
			Message: json.RawMessage(`{
				"role": "assistant",
				"content": [
					{"type": "text", "text": "Let me check the files."},
					{"type": "tool_use", "id": "toolu_01ABC", "name": "Bash", "input": {"command": "git status"}}
				]
			}`),
		},
		{
			UUID:      "uuid-002",
			SessionID: "session-001",
			Type:      models.EntryTypeUser,
			Timestamp: "2026-01-31T10:00:01Z",
			Message: json.RawMessage(`{
				"role": "user",
				"content": [
					{"type": "tool_result", "tool_use_id": "toolu_01ABC", "content": "On branch main\nnothing to commit"}
				]
			}`),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	// Check tool call structure (now includes collapsible collapsed classes)
	if !strings.Contains(html, `class="tool-call collapsible collapsed"`) {
		t.Error("HTML missing tool-call class with collapsible collapsed")
	}
	if !strings.Contains(html, `data-tool-id="toolu_01ABC"`) {
		t.Error("HTML missing tool-id attribute")
	}
	if !strings.Contains(html, `class="tool-header collapsible-trigger"`) {
		t.Error("HTML missing tool-header collapsible-trigger class")
	}
	if !strings.Contains(html, `onclick="toggleTool(this)"`) {
		t.Error("HTML missing toggle onclick handler")
	}
	if !strings.Contains(html, "[Bash] git status") {
		t.Error("HTML missing tool summary")
	}
	if !strings.Contains(html, `class="tool-body hidden collapsible-content collapsed"`) {
		t.Error("HTML missing hidden tool-body class with collapsible content")
	}
	if !strings.Contains(html, `class="tool-input"`) {
		t.Error("HTML missing tool-input class")
	}
	if !strings.Contains(html, `class="tool-output"`) {
		t.Error("HTML missing tool-output class")
	}
	if !strings.Contains(html, "On branch main") {
		t.Error("HTML missing tool output content")
	}
}

func TestRenderConversation_ToolCallError(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-001",
			SessionID: "session-001",
			Type:      models.EntryTypeAssistant,
			Timestamp: "2026-01-31T10:00:00Z",
			Message: json.RawMessage(`{
				"role": "assistant",
				"content": [
					{"type": "tool_use", "id": "toolu_err", "name": "Read", "input": {"file_path": "/nonexistent.txt"}}
				]
			}`),
		},
		{
			UUID:      "uuid-002",
			SessionID: "session-001",
			Type:      models.EntryTypeUser,
			Timestamp: "2026-01-31T10:00:01Z",
			Message: json.RawMessage(`{
				"role": "user",
				"content": [
					{"type": "tool_result", "tool_use_id": "toolu_err", "content": "Error: file not found", "is_error": true}
				]
			}`),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	if !strings.Contains(html, `class="tool-output error"`) {
		t.Error("HTML missing error class on tool output")
	}
}

func TestRenderConversation_SubagentPlaceholder(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-001",
			SessionID: "session-001",
			Type:      models.EntryTypeQueueOperation,
			AgentID:   "a12eb64abc123",
			Timestamp: "2026-01-31T10:00:00Z",
			Message:   json.RawMessage(`"Agent spawned"`),
		},
	}

	agents := []*agent.TreeNode{
		{
			AgentID:    "a12eb64abc123",
			SessionID:  "session-001",
			EntryCount: 29,
		},
	}

	html, err := RenderConversation(entries, agents)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	if !strings.Contains(html, `class="subagent collapsible collapsed"`) {
		t.Error("HTML missing subagent class with collapsible collapsed")
	}
	if !strings.Contains(html, `data-agent-id="a12eb64abc123"`) {
		t.Error("HTML missing agent-id attribute")
	}
	if !strings.Contains(html, `onclick="loadAgent(this)"`) {
		t.Error("HTML missing loadAgent onclick handler")
	}
	if !strings.Contains(html, "Subagent: a12eb64") {
		t.Error("HTML missing truncated agent ID")
	}
	if !strings.Contains(html, "(29 entries)") {
		t.Error("HTML missing entry count")
	}
	if !strings.Contains(html, `class="subagent-content"`) {
		t.Error("HTML missing subagent-content class")
	}
}

func TestRenderAgentFragment_BasicContent(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-agent-001",
			SessionID: "session-001",
			AgentID:   "a12eb64",
			Type:      models.EntryTypeUser,
			Timestamp: "2026-01-31T10:00:00Z",
			Message:   json.RawMessage(`"Agent task"`),
		},
		{
			UUID:      "uuid-agent-002",
			SessionID: "session-001",
			AgentID:   "a12eb64",
			Type:      models.EntryTypeAssistant,
			Timestamp: "2026-01-31T10:00:05Z",
			Message:   json.RawMessage(`{"role": "assistant", "content": [{"type": "text", "text": "Working on it."}]}`),
		},
	}

	html, err := RenderAgentFragment("a12eb64", entries)
	if err != nil {
		t.Fatalf("RenderAgentFragment() error = %v", err)
	}

	// Fragment should not have full HTML structure
	if strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("Fragment should not contain DOCTYPE")
	}
	if strings.Contains(html, "<html>") {
		t.Error("Fragment should not contain <html> tag")
	}

	// But should have entry content
	if !strings.Contains(html, "Agent task") {
		t.Error("Fragment missing user message")
	}
	if !strings.Contains(html, "Working on it.") {
		t.Error("Fragment missing assistant message")
	}
}

func TestEscapeHTML_XSSPrevention(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "script tag",
			input:    "<script>alert('XSS')</script>",
			expected: "&lt;script&gt;alert(&#39;XSS&#39;)&lt;/script&gt;",
		},
		{
			name:     "ampersand",
			input:    "a & b",
			expected: "a &amp; b",
		},
		{
			name:     "double quotes",
			input:    `say "hello"`,
			expected: "say &#34;hello&#34;",
		},
		{
			name:     "less than greater than",
			input:    "1 < 2 > 0",
			expected: "1 &lt; 2 &gt; 0",
		},
		{
			name:     "html attributes",
			input:    `<img src="x" onerror="alert(1)">`,
			expected: "&lt;img src=&#34;x&#34; onerror=&#34;alert(1)&#34;&gt;",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "plain text",
			input:    "Hello World",
			expected: "Hello World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeHTML(tt.input)
			if result != tt.expected {
				t.Errorf("escapeHTML(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRenderConversation_XSSInContent(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-001",
			SessionID: "session-001",
			Type:      models.EntryTypeUser,
			Timestamp: "2026-01-31T10:00:00Z",
			Message:   json.RawMessage(`"<script>alert('XSS')</script>"`),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	// Script tag should be escaped
	if strings.Contains(html, "<script>alert") {
		t.Error("HTML contains unescaped script tag - XSS vulnerability!")
	}
	if !strings.Contains(html, "&lt;script&gt;") {
		t.Error("Script tag should be HTML escaped")
	}
}

func TestRenderConversation_XSSInToolInput(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-001",
			SessionID: "session-001",
			Type:      models.EntryTypeAssistant,
			Timestamp: "2026-01-31T10:00:00Z",
			Message: json.RawMessage(`{
				"role": "assistant",
				"content": [
					{"type": "tool_use", "id": "toolu_xss", "name": "Bash", "input": {"command": "<script>evil()</script>"}}
				]
			}`),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	if strings.Contains(html, "<script>evil") {
		t.Error("HTML contains unescaped script in tool input - XSS vulnerability!")
	}
}

func TestRenderConversation_EmptySession(t *testing.T) {
	entries := []models.ConversationEntry{}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	// Should still produce valid HTML structure
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("Empty session HTML missing DOCTYPE")
	}
	if !strings.Contains(html, `<div class="conversation">`) {
		t.Error("Empty session HTML missing conversation container")
	}
}

func TestRenderConversation_NullContent(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-001",
			SessionID: "session-001",
			Type:      models.EntryTypeUser,
			Timestamp: "2026-01-31T10:00:00Z",
			Message:   nil,
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	// Null content should be skipped (no message-row rendered)
	if strings.Contains(html, `class="message-row user"`) {
		t.Error("HTML should not render message-row for null content")
	}
}

func TestRenderConversation_WhitespaceOnlyContent(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-whitespace-001",
			SessionID: "session-001",
			Type:      models.EntryTypeAssistant,
			Timestamp: "2026-01-31T10:00:00Z",
			Message:   json.RawMessage(`{"role": "assistant", "content": [{"type": "text", "text": "   \n\t  "}]}`),
		},
		{
			UUID:      "uuid-whitespace-002",
			SessionID: "session-001",
			Type:      models.EntryTypeAssistant,
			Timestamp: "2026-01-31T10:00:05Z",
			Message:   json.RawMessage(`{"role": "assistant", "content": [{"type": "text", "text": ""}]}`),
		},
		{
			UUID:      "uuid-whitespace-003",
			SessionID: "session-001",
			Type:      models.EntryTypeUser,
			Timestamp: "2026-01-31T10:00:10Z",
			Message:   json.RawMessage(`"     "`),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	// Whitespace-only assistant entries should be skipped
	if strings.Contains(html, `data-uuid="uuid-whitespace-001"`) {
		t.Error("HTML should not render assistant entry with only whitespace")
	}
	if strings.Contains(html, `data-uuid="uuid-whitespace-002"`) {
		t.Error("HTML should not render assistant entry with empty text")
	}
	if strings.Contains(html, `data-uuid="uuid-whitespace-003"`) {
		t.Error("HTML should not render user entry with only whitespace")
	}
}

func TestRenderConversation_AssistantWithOnlyToolCalls(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-toolonly-001",
			SessionID: "session-001",
			Type:      models.EntryTypeAssistant,
			Timestamp: "2026-01-31T10:00:00Z",
			Message: json.RawMessage(`{
				"role": "assistant",
				"content": [
					{"type": "text", "text": "  "},
					{"type": "tool_use", "id": "toolu_01", "name": "Read", "input": {"file_path": "/test.go"}}
				]
			}`),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	// Assistant entries with tool calls should render even if text is only whitespace
	if !strings.Contains(html, `data-uuid="uuid-toolonly-001"`) {
		t.Error("HTML should render assistant entry with tool calls even if text is whitespace")
	}
	if !strings.Contains(html, "[Read] /test.go") {
		t.Error("HTML should contain tool call summary")
	}
}

func TestRenderConversation_SpecialCharacters(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-special",
			SessionID: "session-001",
			Type:      models.EntryTypeUser,
			Timestamp: "2026-01-31T10:00:00Z",
			Message:   json.RawMessage(`"Special chars: \u0000 \n \t \r unicode: \u4e2d\u6587"`),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	// Should handle special characters without error
	if !strings.Contains(html, "Special chars:") {
		t.Error("HTML missing special character content")
	}
}

func TestRenderConversation_VeryLongContent(t *testing.T) {
	// Create a very long message
	longText := strings.Repeat("a", 100000)
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-long",
			SessionID: "session-001",
			Type:      models.EntryTypeUser,
			Timestamp: "2026-01-31T10:00:00Z",
			Message:   json.RawMessage(`"` + longText + `"`),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	// Should handle long content
	if len(html) < 100000 {
		t.Error("HTML should contain the long content")
	}
}

func TestRenderConversation_MultipleToolCalls(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-001",
			SessionID: "session-001",
			Type:      models.EntryTypeAssistant,
			Timestamp: "2026-01-31T10:00:00Z",
			Message: json.RawMessage(`{
				"role": "assistant",
				"content": [
					{"type": "tool_use", "id": "toolu_01", "name": "Read", "input": {"file_path": "/file1.go"}},
					{"type": "tool_use", "id": "toolu_02", "name": "Read", "input": {"file_path": "/file2.go"}},
					{"type": "tool_use", "id": "toolu_03", "name": "Bash", "input": {"command": "go test"}}
				]
			}`),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	// Count tool-call divs (now with collapsible collapsed classes)
	count := strings.Count(html, `class="tool-call collapsible collapsed"`)
	if count != 3 {
		t.Errorf("HTML has %d tool-call divs, want 3", count)
	}

	if !strings.Contains(html, "[Read] /file1.go") {
		t.Error("HTML missing first Read tool")
	}
	if !strings.Contains(html, "[Read] /file2.go") {
		t.Error("HTML missing second Read tool")
	}
	if !strings.Contains(html, "[Bash] go test") {
		t.Error("HTML missing Bash tool")
	}
}

func TestRenderEntry_AllEntryTypes(t *testing.T) {
	tests := []struct {
		entryType     models.EntryType
		expectedClass string
	}{
		{models.EntryTypeUser, "user"},
		{models.EntryTypeAssistant, "assistant"},
		{models.EntryTypeSystem, "system"},
		{models.EntryTypeQueueOperation, "queue-operation"},
		{models.EntryTypeSummary, "summary"},
	}

	for _, tt := range tests {
		t.Run(string(tt.entryType), func(t *testing.T) {
			entry := models.ConversationEntry{
				UUID:      "uuid-test",
				Type:      tt.entryType,
				Timestamp: "2026-01-31T10:00:00Z",
				Message:   json.RawMessage(`"test"`),
			}

			html := renderEntry(entry, nil)

			if !strings.Contains(html, `class="message-row `+tt.expectedClass+`"`) {
				t.Errorf("Entry type %s should have message-row class %s", tt.entryType, tt.expectedClass)
			}
		})
	}
}

func TestGetEntryClass_UnknownType(t *testing.T) {
	result := getEntryClass(models.EntryType("unknown-type"))
	if result != "unknown" {
		t.Errorf("getEntryClass(unknown-type) = %q, want %q", result, "unknown")
	}
}

func TestFormatTimestamp_ValidTimestamp(t *testing.T) {
	result := formatTimestamp("2026-01-31T14:30:45.123456789Z")
	if result != "14:30:45" {
		t.Errorf("formatTimestamp() = %q, want %q", result, "14:30:45")
	}
}

func TestFormatTimestamp_InvalidTimestamp(t *testing.T) {
	result := formatTimestamp("not-a-timestamp")
	// Should return original string when parsing fails
	if result != "not-a-timestamp" {
		t.Errorf("formatTimestamp(invalid) = %q, want original string", result)
	}
}

func TestFormatToolSummary_AllToolTypes(t *testing.T) {
	tests := []struct {
		name     string
		tool     models.ToolUse
		expected string
	}{
		{
			name:     "Bash",
			tool:     models.ToolUse{Name: "Bash", Input: map[string]any{"command": "ls -la"}},
			expected: "[Bash] ls -la",
		},
		{
			name:     "Read",
			tool:     models.ToolUse{Name: "Read", Input: map[string]any{"file_path": "/path/to/file.go"}},
			expected: "[Read] /path/to/file.go",
		},
		{
			name:     "Write",
			tool:     models.ToolUse{Name: "Write", Input: map[string]any{"file_path": "/new/file.go"}},
			expected: "[Write] /new/file.go",
		},
		{
			name:     "Edit",
			tool:     models.ToolUse{Name: "Edit", Input: map[string]any{"file_path": "/edit/file.go"}},
			expected: "[Edit] /edit/file.go",
		},
		{
			name:     "Grep",
			tool:     models.ToolUse{Name: "Grep", Input: map[string]any{"pattern": "func.*Test"}},
			expected: "[Grep] func.*Test",
		},
		{
			name:     "Glob",
			tool:     models.ToolUse{Name: "Glob", Input: map[string]any{"pattern": "**/*.go"}},
			expected: "[Glob] **/*.go",
		},
		{
			name:     "Task",
			tool:     models.ToolUse{Name: "Task", Input: map[string]any{"description": "Explore code"}},
			expected: "[Task] Explore code",
		},
		{
			name:     "Task with prompt",
			tool:     models.ToolUse{Name: "Task", Input: map[string]any{"prompt": "Find bugs"}},
			expected: "[Task] Find bugs",
		},
		{
			name:     "WebFetch",
			tool:     models.ToolUse{Name: "WebFetch", Input: map[string]any{"url": "https://example.com"}},
			expected: "[WebFetch] https://example.com",
		},
		{
			name:     "WebSearch",
			tool:     models.ToolUse{Name: "WebSearch", Input: map[string]any{"query": "go testing"}},
			expected: "[WebSearch] go testing",
		},
		{
			name:     "Unknown tool",
			tool:     models.ToolUse{Name: "CustomTool", Input: map[string]any{"custom": "value"}},
			expected: "[CustomTool]",
		},
		{
			name:     "No input",
			tool:     models.ToolUse{Name: "Bash", Input: nil},
			expected: "[Bash]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatToolSummary(tt.tool)
			if result != tt.expected {
				t.Errorf("formatToolSummary() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFormatToolSummary_LongInput(t *testing.T) {
	longCommand := strings.Repeat("a", 100)
	tool := models.ToolUse{
		Name:  "Bash",
		Input: map[string]any{"command": longCommand},
	}

	result := formatToolSummary(tool)

	// Should be truncated
	if len(result) > 70 {
		t.Errorf("formatToolSummary() length = %d, should be truncated", len(result))
	}
	if !strings.HasSuffix(result, "...") {
		t.Error("Truncated summary should end with ...")
	}
}

func TestFormatToolInput_NilInput(t *testing.T) {
	result := formatToolInput(nil)
	if result != "{}" {
		t.Errorf("formatToolInput(nil) = %q, want %q", result, "{}")
	}
}

func TestFormatToolInput_EmptyInput(t *testing.T) {
	result := formatToolInput(map[string]any{})
	if result != "{}" {
		t.Errorf("formatToolInput({}) = %q, want %q", result, "{}")
	}
}

func TestFormatToolInput_ComplexInput(t *testing.T) {
	input := map[string]any{
		"command":     "ls -la",
		"description": "List files",
	}

	result := formatToolInput(input)

	// Should be valid JSON
	if !strings.HasPrefix(result, "{") {
		t.Error("formatToolInput should return JSON object")
	}
	if !strings.Contains(result, "command") {
		t.Error("formatToolInput should contain 'command'")
	}
	if !strings.Contains(result, "ls -la") {
		t.Error("formatToolInput should contain command value")
	}
}

func TestBuildAgentMap_EmptyAgents(t *testing.T) {
	result := buildAgentMap(nil)
	if len(result) != 0 {
		t.Errorf("buildAgentMap(nil) returned %d entries, want 0", len(result))
	}

	result = buildAgentMap([]*agent.TreeNode{})
	if len(result) != 0 {
		t.Errorf("buildAgentMap([]) returned %d entries, want 0", len(result))
	}
}

func TestBuildAgentMap_SingleAgent(t *testing.T) {
	agents := []*agent.TreeNode{
		{
			AgentID:    "agent-001",
			EntryCount: 15,
		},
	}

	result := buildAgentMap(agents)

	if count, ok := result["agent-001"]; !ok || count != 15 {
		t.Errorf("buildAgentMap() agent-001 = %d, want 15", count)
	}
}

func TestBuildAgentMap_NestedAgents(t *testing.T) {
	agents := []*agent.TreeNode{
		{
			AgentID:    "parent-agent",
			EntryCount: 10,
			Children: []*agent.TreeNode{
				{
					AgentID:    "child-agent-1",
					EntryCount: 5,
				},
				{
					AgentID:    "child-agent-2",
					EntryCount: 8,
					Children: []*agent.TreeNode{
						{
							AgentID:    "grandchild-agent",
							EntryCount: 3,
						},
					},
				},
			},
		},
	}

	result := buildAgentMap(agents)

	expected := map[string]int{
		"parent-agent":     10,
		"child-agent-1":    5,
		"child-agent-2":    8,
		"grandchild-agent": 3,
	}

	for id, count := range expected {
		if result[id] != count {
			t.Errorf("buildAgentMap() %s = %d, want %d", id, result[id], count)
		}
	}
}

func TestBuildToolResultsMap_Empty(t *testing.T) {
	result := buildToolResultsMap(nil)
	if len(result) != 0 {
		t.Errorf("buildToolResultsMap(nil) returned %d entries, want 0", len(result))
	}
}

func TestBuildToolResultsMap_WithResults(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			Type: models.EntryTypeUser,
			Message: json.RawMessage(`{
				"role": "user",
				"content": [
					{"type": "tool_result", "tool_use_id": "toolu_01", "content": "Result 1"},
					{"type": "tool_result", "tool_use_id": "toolu_02", "content": "Result 2", "is_error": true}
				]
			}`),
		},
	}

	result := buildToolResultsMap(entries)

	if len(result) != 2 {
		t.Fatalf("buildToolResultsMap() returned %d entries, want 2", len(result))
	}

	if r := result["toolu_01"]; r.Content != "Result 1" || r.IsError {
		t.Errorf("Result toolu_01: %+v", r)
	}
	if r := result["toolu_02"]; r.Content != "Result 2" || !r.IsError {
		t.Errorf("Result toolu_02: %+v", r)
	}
}

func TestBuildToolResultsMap_IgnoresNonUserEntries(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			Type: models.EntryTypeAssistant,
			Message: json.RawMessage(`{
				"role": "assistant",
				"content": [
					{"type": "tool_use", "id": "toolu_01", "name": "Bash"}
				]
			}`),
		},
	}

	result := buildToolResultsMap(entries)

	if len(result) != 0 {
		t.Errorf("buildToolResultsMap() should ignore assistant entries, got %d", len(result))
	}
}

func TestRenderToolCall_NoResult(t *testing.T) {
	tool := models.ToolUse{
		ID:    "toolu_no_result",
		Name:  "Bash",
		Input: map[string]any{"command": "echo test"},
	}

	html := renderToolCall(tool, models.ToolResult{}, false)

	// Should have tool-call structure (with collapsible collapsed classes)
	if !strings.Contains(html, `class="tool-call collapsible collapsed"`) {
		t.Error("HTML missing tool-call class with collapsible collapsed")
	}
	if !strings.Contains(html, `class="tool-input"`) {
		t.Error("HTML missing tool-input class")
	}
	// Should NOT have tool-output when hasResult is false
	if strings.Contains(html, `class="tool-output"`) {
		t.Error("HTML should not have tool-output when hasResult is false")
	}
}

func TestRenderConversation_TimestampInHeader(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-001",
			SessionID: "session-001",
			Type:      models.EntryTypeUser,
			Timestamp: "2026-01-31T14:30:45Z",
			Message:   json.RawMessage(`"test"`),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	if !strings.Contains(html, `class="timestamp"`) {
		t.Error("HTML missing timestamp class")
	}
	// Updated to check for readable timestamp format (2:30 PM)
	if !strings.Contains(html, "2:30 PM") {
		t.Error("HTML missing formatted timestamp (expected 2:30 PM)")
	}
}

func TestRenderConversation_RoleInHeader(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-001",
			SessionID: "session-001",
			Type:      models.EntryTypeUser,
			Timestamp: "2026-01-31T10:00:00Z",
			Message:   json.RawMessage(`"test"`),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	// Chat bubble layout uses "role" instead of "type"
	if !strings.Contains(html, `class="role"`) {
		t.Error("HTML missing role class")
	}
	if !strings.Contains(html, ">User<") {
		t.Error("HTML missing role label in header")
	}
}

func TestRenderConversation_AgentIDInHeader(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-001",
			SessionID: "session-001",
			AgentID:   "a12eb64",
			Type:      models.EntryTypeAssistant,
			Timestamp: "2026-01-31T10:00:00Z",
			Message:   json.RawMessage(`{"role": "assistant", "content": [{"type": "text", "text": "Agent response"}]}`),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	if !strings.Contains(html, `class="agent-id"`) {
		t.Error("HTML missing agent-id class")
	}
	if !strings.Contains(html, "[a12eb64]") {
		t.Error("HTML missing agent ID in header")
	}
}

func TestRenderConversation_MalformedJSON(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-001",
			SessionID: "session-001",
			Type:      models.EntryTypeAssistant,
			Timestamp: "2026-01-31T10:00:00Z",
			Message:   json.RawMessage(`{invalid json`),
		},
	}

	// Should not panic, should handle gracefully
	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() should not error on malformed JSON: %v", err)
	}

	// Malformed content with no parseable text should be skipped
	if strings.Contains(html, `class="message-row assistant"`) {
		t.Error("HTML should not render message-row for malformed content with no text")
	}
}

func TestRenderSubagentPlaceholder_ShortAgentID(t *testing.T) {
	agentMap := map[string]int{
		"abc": 5,
	}

	html := renderSubagentPlaceholder("abc", agentMap)

	// Short IDs should not be truncated
	if !strings.Contains(html, "Subagent: abc") {
		t.Error("Short agent ID should not be truncated")
	}
}

func TestRenderSubagentPlaceholder_LongAgentID(t *testing.T) {
	agentMap := map[string]int{
		"a12eb64abc123def456": 10,
	}

	html := renderSubagentPlaceholder("a12eb64abc123def456", agentMap)

	// Long IDs should be truncated to 7 chars in display
	if !strings.Contains(html, "Subagent: a12eb64") {
		t.Error("Long agent ID should be truncated in display")
	}
	// But full ID should be in data attribute
	if !strings.Contains(html, `data-agent-id="a12eb64abc123def456"`) {
		t.Error("Full agent ID should be in data attribute")
	}
}

func TestRenderSubagentPlaceholder_ZeroEntries(t *testing.T) {
	agentMap := map[string]int{}

	html := renderSubagentPlaceholder("agent-x", agentMap)

	// Should show 0 entries when not in map
	if !strings.Contains(html, "(0 entries)") {
		t.Error("Unknown agent should show 0 entries")
	}
}

func TestRenderConversation_MultilineContent(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-001",
			SessionID: "session-001",
			Type:      models.EntryTypeUser,
			Timestamp: "2026-01-31T10:00:00Z",
			Message:   json.RawMessage(`"Line 1\nLine 2\nLine 3"`),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	// Newlines should be preserved (as \n in content)
	if !strings.Contains(html, "Line 1\nLine 2\nLine 3") {
		t.Error("HTML should preserve newlines in content")
	}
}

func TestRenderAgentFragment_WithToolCalls(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-agent-001",
			SessionID: "session-001",
			AgentID:   "a12eb64",
			Type:      models.EntryTypeAssistant,
			Timestamp: "2026-01-31T10:00:00Z",
			Message: json.RawMessage(`{
				"role": "assistant",
				"content": [
					{"type": "tool_use", "id": "toolu_agent_01", "name": "Read", "input": {"file_path": "/file.go"}}
				]
			}`),
		},
		{
			UUID:      "uuid-agent-002",
			SessionID: "session-001",
			AgentID:   "a12eb64",
			Type:      models.EntryTypeUser,
			Timestamp: "2026-01-31T10:00:01Z",
			Message: json.RawMessage(`{
				"role": "user",
				"content": [
					{"type": "tool_result", "tool_use_id": "toolu_agent_01", "content": "file contents"}
				]
			}`),
		},
	}

	html, err := RenderAgentFragment("a12eb64", entries)
	if err != nil {
		t.Fatalf("RenderAgentFragment() error = %v", err)
	}

	if !strings.Contains(html, "[Read] /file.go") {
		t.Error("Fragment should contain tool call")
	}
	if !strings.Contains(html, "file contents") {
		t.Error("Fragment should contain tool result")
	}
}

func TestRenderConversation_ToolCallWithoutMatchingResult(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-001",
			SessionID: "session-001",
			Type:      models.EntryTypeAssistant,
			Timestamp: "2026-01-31T10:00:00Z",
			Message: json.RawMessage(`{
				"role": "assistant",
				"content": [
					{"type": "tool_use", "id": "toolu_orphan", "name": "Bash", "input": {"command": "echo test"}}
				]
			}`),
		},
		// No matching tool result
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	// Should still render the tool call
	if !strings.Contains(html, "[Bash] echo test") {
		t.Error("HTML should contain tool call even without result")
	}
	// But should not have tool-output
	if strings.Contains(html, `class="tool-output"`) {
		t.Error("HTML should not have tool-output for orphan tool call")
	}
}
