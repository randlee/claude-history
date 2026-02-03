package export

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/randlee/claude-history/pkg/models"
)

func TestGetToolInfo_KnownTools(t *testing.T) {
	tests := []struct {
		toolName    string
		wantIcon    string
		wantHint    string
		wantClass   string
		wantDisplay string
	}{
		{"Bash", "\xf0\x9f\x94\xa7", "command execution", "tool-bash", "Bash"},
		{"Read", "\xf0\x9f\x93\x84", "file read", "tool-read", "Read"},
		{"Write", "\xf0\x9f\x93\x9d", "file write", "tool-write", "Write"},
		{"Edit", "\xe2\x9c\x8f\xef\xb8\x8f", "file edit", "tool-edit", "Edit"},
		{"Grep", "\xf0\x9f\x94\x8d", "content search", "tool-grep", "Grep"},
		{"Glob", "\xf0\x9f\x93\x81", "file pattern matching", "tool-glob", "Glob"},
		{"Task", "\xf0\x9f\xa4\x96", "spawn subagent", "tool-task", "Task"},
		{"WebFetch", "\xf0\x9f\x8c\x90", "fetch URL", "tool-webfetch", "WebFetch"},
		{"WebSearch", "\xf0\x9f\x94\x8e", "web search", "tool-websearch", "WebSearch"},
		{"NotebookEdit", "\xf0\x9f\x93\x93", "notebook edit", "tool-notebook", "NotebookEdit"},
	}

	for _, tt := range tests {
		t.Run(tt.toolName, func(t *testing.T) {
			info := GetToolInfo(tt.toolName)

			if info.Icon != tt.wantIcon {
				t.Errorf("GetToolInfo(%q).Icon = %q, want %q", tt.toolName, info.Icon, tt.wantIcon)
			}
			if info.Hint != tt.wantHint {
				t.Errorf("GetToolInfo(%q).Hint = %q, want %q", tt.toolName, info.Hint, tt.wantHint)
			}
			if info.ColorClass != tt.wantClass {
				t.Errorf("GetToolInfo(%q).ColorClass = %q, want %q", tt.toolName, info.ColorClass, tt.wantClass)
			}
			if info.DisplayName != tt.wantDisplay {
				t.Errorf("GetToolInfo(%q).DisplayName = %q, want %q", tt.toolName, info.DisplayName, tt.wantDisplay)
			}
		})
	}
}

func TestGetToolInfo_UnknownTool(t *testing.T) {
	info := GetToolInfo("UnknownTool")

	if info.Icon != "\xf0\x9f\x94\xa7" {
		t.Errorf("Unknown tool should have default icon, got %q", info.Icon)
	}
	if info.Hint != "tool" {
		t.Errorf("Unknown tool should have 'tool' hint, got %q", info.Hint)
	}
	if info.ColorClass != "tool-unknown" {
		t.Errorf("Unknown tool should have 'tool-unknown' class, got %q", info.ColorClass)
	}
	if info.DisplayName != "UnknownTool" {
		t.Errorf("Unknown tool should preserve name, got %q", info.DisplayName)
	}
}

func TestRenderToolOverlay_BashTool(t *testing.T) {
	tool := models.ToolUse{
		ID:    "toolu_01ABC123",
		Name:  "Bash",
		Input: map[string]any{"command": "git status"},
	}
	result := models.ToolResult{
		ToolUseID: "toolu_01ABC123",
		Content:   "On branch main\nnothing to commit",
		IsError:   false,
	}

	html := RenderToolOverlay(tool, result, true)

	// Check structure
	if !strings.Contains(html, `class="tool-overlay tool-bash collapsible"`) {
		t.Error("Missing tool-overlay and tool-bash classes")
	}
	if !strings.Contains(html, `data-tool-id="toolu_01ABC123"`) {
		t.Error("Missing data-tool-id attribute")
	}

	// Check header elements
	if !strings.Contains(html, `class="tool-header collapsible-trigger"`) {
		t.Error("Missing tool-header collapsible-trigger class")
	}
	if !strings.Contains(html, `onclick="toggleToolOverlay(this)"`) {
		t.Error("Missing toggleToolOverlay onclick handler")
	}
	if !strings.Contains(html, `class="tool-icon"`) {
		t.Error("Missing tool-icon class")
	}
	if !strings.Contains(html, `class="tool-name"`) {
		t.Error("Missing tool-name class")
	}
	if !strings.Contains(html, `class="tool-hint"`) {
		t.Error("Missing tool-hint class")
	}
	if !strings.Contains(html, "(command execution)") {
		t.Error("Missing tool hint text")
	}
	if !strings.Contains(html, `class="chevron down"`) {
		t.Error("Missing chevron class")
	}

	// Check body structure
	if !strings.Contains(html, `class="tool-body collapsible-content collapsed"`) {
		t.Error("Missing tool-body classes")
	}
	if !strings.Contains(html, `class="tool-section"`) {
		t.Error("Missing tool-section class")
	}
	if !strings.Contains(html, "<h4>Tool ID</h4>") {
		t.Error("Missing Tool ID section header")
	}
	if !strings.Contains(html, "<h4>Input</h4>") {
		t.Error("Missing Input section header")
	}
	if !strings.Contains(html, "<h4>Output</h4>") {
		t.Error("Missing Output section header")
	}

	// Check content
	if !strings.Contains(html, "git status") {
		t.Error("Missing command in input")
	}
	if !strings.Contains(html, "On branch main") {
		t.Error("Missing output content")
	}

	// Check copy buttons
	if !strings.Contains(html, `data-copy-type="tool-id"`) {
		t.Error("Missing tool-id copy button")
	}
}

func TestRenderToolOverlay_ReadToolWithFilePath(t *testing.T) {
	tool := models.ToolUse{
		ID:    "toolu_read_01",
		Name:  "Read",
		Input: map[string]any{"file_path": "/path/to/file.go"},
	}

	html := RenderToolOverlay(tool, models.ToolResult{}, false)

	// Check Read-specific styling
	if !strings.Contains(html, "tool-read") {
		t.Error("Missing tool-read class")
	}

	// Check file path copy button
	if !strings.Contains(html, `data-copy-type="file-path"`) {
		t.Error("Missing file-path copy button")
	}
	if !strings.Contains(html, `data-copy-text="/path/to/file.go"`) {
		t.Error("Missing file path in copy button data")
	}
}

func TestRenderToolOverlay_NoResult(t *testing.T) {
	tool := models.ToolUse{
		ID:    "toolu_no_result",
		Name:  "Bash",
		Input: map[string]any{"command": "echo test"},
	}

	html := RenderToolOverlay(tool, models.ToolResult{}, false)

	// Should NOT have Output section
	if strings.Contains(html, "<h4>Output</h4>") {
		t.Error("Should not have Output section when hasResult is false")
	}
}

func TestRenderToolOverlay_ErrorResult(t *testing.T) {
	tool := models.ToolUse{
		ID:    "toolu_error",
		Name:  "Read",
		Input: map[string]any{"file_path": "/nonexistent.txt"},
	}
	result := models.ToolResult{
		ToolUseID: "toolu_error",
		Content:   "Error: file not found",
		IsError:   true,
	}

	html := RenderToolOverlay(tool, result, true)

	// Check error class on output
	if !strings.Contains(html, `class="tool-output error"`) {
		t.Error("Missing error class on tool output")
	}
}

func TestRenderToolOverlay_AllToolTypes(t *testing.T) {
	toolNames := []string{"Bash", "Read", "Write", "Edit", "Grep", "Glob", "Task", "WebFetch", "WebSearch", "NotebookEdit"}

	for _, name := range toolNames {
		t.Run(name, func(t *testing.T) {
			tool := models.ToolUse{
				ID:    "toolu_" + strings.ToLower(name),
				Name:  name,
				Input: map[string]any{"test": "value"},
			}

			html := RenderToolOverlay(tool, models.ToolResult{}, false)

			info := GetToolInfo(name)
			if !strings.Contains(html, info.ColorClass) {
				t.Errorf("Missing color class %q for tool %s", info.ColorClass, name)
			}
			if !strings.Contains(html, info.Icon) {
				t.Errorf("Missing icon for tool %s", name)
			}
		})
	}
}

func TestRenderToolOverlay_CharacterCount(t *testing.T) {
	tests := []struct {
		name           string
		inputSize      int
		outputSize     int
		hasResult      bool
		expectedPhrase string
	}{
		{"small input only", 50, 0, false, "50 chars in"},
		{"large input only", 5000, 0, false, "5.0K chars in"},
		{"input and output", 100, 200, true, "100 chars in / 200 chars out"},
		{"large input and output", 15000, 25000, true, "15K chars in / 25K chars out"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := map[string]any{"data": strings.Repeat("x", tt.inputSize-20)} // Account for JSON formatting
			tool := models.ToolUse{
				ID:    "toolu_test",
				Name:  "Bash",
				Input: input,
			}
			result := models.ToolResult{
				Content: strings.Repeat("y", tt.outputSize),
			}

			html := RenderToolOverlay(tool, result, tt.hasResult)

			if !strings.Contains(html, "tool-char-count") {
				t.Error("Missing tool-char-count class")
			}
		})
	}
}

func TestRenderToolOverlay_XSSPrevention(t *testing.T) {
	tool := models.ToolUse{
		ID:    "<script>alert('xss')</script>",
		Name:  "Bash",
		Input: map[string]any{"command": "<script>evil()</script>"},
	}
	result := models.ToolResult{
		Content: "<script>malicious()</script>",
	}

	html := RenderToolOverlay(tool, result, true)

	// Ensure no unescaped script tags
	if strings.Contains(html, "<script>") {
		t.Error("HTML contains unescaped script tag - XSS vulnerability!")
	}
	if !strings.Contains(html, "&lt;script&gt;") {
		t.Error("Script tags should be HTML escaped")
	}
}

func TestRenderSubagentOverlay_BasicStructure(t *testing.T) {
	html := RenderSubagentOverlay("a12eb64abc123def456", 29, nil)

	// Check structure
	if !strings.Contains(html, `class="subagent-overlay agent-overlay collapsible"`) {
		t.Error("Missing subagent-overlay classes")
	}
	if !strings.Contains(html, `data-agent-id="a12eb64abc123def456"`) {
		t.Error("Missing data-agent-id attribute")
	}

	// Check header elements
	if !strings.Contains(html, `class="agent-icon"`) {
		t.Error("Missing agent-icon class")
	}
	if !strings.Contains(html, "\xf0\x9f\xa4\x96") { // robot emoji
		t.Error("Missing robot icon")
	}
	if !strings.Contains(html, "Subagent: a12eb64") {
		t.Error("Missing truncated agent ID in title")
	}
	if !strings.Contains(html, "(29 entries)") {
		t.Error("Missing entry count")
	}

	// Check copy button
	if !strings.Contains(html, `data-copy-text="a12eb64abc123def456"`) {
		t.Error("Missing full agent ID in copy button")
	}
	if !strings.Contains(html, `data-copy-type="agent-id"`) {
		t.Error("Missing agent-id copy type")
	}

	// Check Deep Dive button
	if !strings.Contains(html, `class="deep-dive-btn"`) {
		t.Error("Missing deep-dive-btn class")
	}
	if !strings.Contains(html, "Deep Dive") {
		t.Error("Missing Deep Dive button text")
	}
	if !strings.Contains(html, `onclick="deepDiveAgent('a12eb64abc123def456', event)"`) {
		t.Error("Missing deepDiveAgent onclick handler")
	}

	// Check content container
	if !strings.Contains(html, `class="subagent-content collapsible-content collapsed"`) {
		t.Error("Missing subagent-content classes")
	}
}

func TestRenderSubagentOverlay_ShortAgentID(t *testing.T) {
	html := RenderSubagentOverlay("abc", 5, nil)

	// Short IDs should not be truncated
	if !strings.Contains(html, "Subagent: abc") {
		t.Error("Short agent ID should not be truncated")
	}
}

func TestRenderSubagentOverlay_WithMetadata(t *testing.T) {
	metadata := map[string]string{
		"Session":  "fbd51e2b",
		"Duration": "5m 32s",
	}

	html := RenderSubagentOverlay("agent123", 10, metadata)

	// Check metadata section
	if !strings.Contains(html, `class="subagent-metadata"`) {
		t.Error("Missing subagent-metadata class")
	}
	if !strings.Contains(html, "Session:") {
		t.Error("Missing Session metadata key")
	}
	if !strings.Contains(html, "fbd51e2b") {
		t.Error("Missing Session metadata value")
	}
}

func TestRenderSubagentOverlay_NoMetadata(t *testing.T) {
	html := RenderSubagentOverlay("agent123", 10, nil)

	// Should not have metadata section when nil
	if strings.Contains(html, `class="subagent-metadata"`) {
		t.Error("Should not have metadata section when metadata is nil")
	}
}

func TestRenderSubagentOverlay_EmptyMetadata(t *testing.T) {
	html := RenderSubagentOverlay("agent123", 10, map[string]string{})

	// Should not have metadata section when empty
	if strings.Contains(html, `class="subagent-metadata"`) {
		t.Error("Should not have metadata section when metadata is empty")
	}
}

func TestRenderSubagentOverlay_XSSPrevention(t *testing.T) {
	metadata := map[string]string{
		"<script>": "<script>alert('xss')</script>",
	}

	html := RenderSubagentOverlay("<script>evil</script>", 10, metadata)

	// Ensure no unescaped script tags
	if strings.Contains(html, "<script>") {
		t.Error("HTML contains unescaped script tag - XSS vulnerability!")
	}
}

func TestRenderThinkingOverlay_BasicStructure(t *testing.T) {
	thinking := ThinkingBlock{
		Content: "Let me think about this problem step by step...",
	}

	html := RenderThinkingOverlay(thinking)

	// Check structure
	if !strings.Contains(html, `class="thinking-overlay collapsible"`) {
		t.Error("Missing thinking-overlay classes")
	}

	// Check header elements
	if !strings.Contains(html, `class="thinking-header overlay-header collapsible-trigger"`) {
		t.Error("Missing thinking-header classes")
	}
	if !strings.Contains(html, `onclick="toggleThinking(this)"`) {
		t.Error("Missing toggleThinking onclick handler")
	}
	if !strings.Contains(html, `class="thinking-icon"`) {
		t.Error("Missing thinking-icon class")
	}
	if !strings.Contains(html, "\xf0\x9f\x92\xa1") { // lightbulb emoji
		t.Error("Missing lightbulb icon")
	}
	if !strings.Contains(html, "Thinking") {
		t.Error("Missing 'Thinking' title")
	}
	if !strings.Contains(html, `class="thinking-char-count"`) {
		t.Error("Missing thinking-char-count class")
	}

	// Check body
	if !strings.Contains(html, `class="thinking-body collapsible-content collapsed"`) {
		t.Error("Missing thinking-body classes")
	}
	if !strings.Contains(html, `class="thinking-content"`) {
		t.Error("Missing thinking-content class")
	}
	if !strings.Contains(html, "Let me think about this problem step by step...") {
		t.Error("Missing thinking content")
	}
}

func TestRenderThinkingOverlay_CharacterCount(t *testing.T) {
	tests := []struct {
		contentLen int
		expected   string
	}{
		{50, "50 chars"},
		{1500, "1.5K chars"},
		{15000, "15K chars"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			thinking := ThinkingBlock{
				Content: strings.Repeat("a", tt.contentLen),
			}

			html := RenderThinkingOverlay(thinking)

			if !strings.Contains(html, tt.expected) {
				t.Errorf("Expected character count %q not found in HTML", tt.expected)
			}
		})
	}
}

func TestRenderThinkingOverlay_XSSPrevention(t *testing.T) {
	thinking := ThinkingBlock{
		Content: "<script>alert('xss')</script>",
	}

	html := RenderThinkingOverlay(thinking)

	// Ensure no unescaped script tags
	if strings.Contains(html, "<script>") {
		t.Error("HTML contains unescaped script tag - XSS vulnerability!")
	}
	if !strings.Contains(html, "&lt;script&gt;") {
		t.Error("Script tags should be HTML escaped")
	}
}

func TestExtractThinkingBlocks_AssistantWithThinking(t *testing.T) {
	entry := models.ConversationEntry{
		Type: models.EntryTypeAssistant,
		Message: json.RawMessage(`{
			"role": "assistant",
			"content": [
				{"type": "thinking", "text": "Let me analyze this..."},
				{"type": "text", "text": "Here is my response."},
				{"type": "thinking", "text": "I should also consider..."}
			]
		}`),
	}

	blocks := ExtractThinkingBlocks(entry)

	if len(blocks) != 2 {
		t.Fatalf("Expected 2 thinking blocks, got %d", len(blocks))
	}
	if blocks[0].Content != "Let me analyze this..." {
		t.Errorf("First thinking block content = %q, want %q", blocks[0].Content, "Let me analyze this...")
	}
	if blocks[1].Content != "I should also consider..." {
		t.Errorf("Second thinking block content = %q, want %q", blocks[1].Content, "I should also consider...")
	}
}

func TestExtractThinkingBlocks_NoThinking(t *testing.T) {
	entry := models.ConversationEntry{
		Type: models.EntryTypeAssistant,
		Message: json.RawMessage(`{
			"role": "assistant",
			"content": [
				{"type": "text", "text": "Just a regular response."}
			]
		}`),
	}

	blocks := ExtractThinkingBlocks(entry)

	if len(blocks) != 0 {
		t.Errorf("Expected 0 thinking blocks, got %d", len(blocks))
	}
}

func TestExtractThinkingBlocks_NonAssistant(t *testing.T) {
	entry := models.ConversationEntry{
		Type:    models.EntryTypeUser,
		Message: json.RawMessage(`"User message"`),
	}

	blocks := ExtractThinkingBlocks(entry)

	if blocks != nil {
		t.Errorf("Expected nil for non-assistant entry, got %v", blocks)
	}
}

func TestExtractThinkingBlocks_MalformedJSON(t *testing.T) {
	entry := models.ConversationEntry{
		Type:    models.EntryTypeAssistant,
		Message: json.RawMessage(`{invalid json`),
	}

	blocks := ExtractThinkingBlocks(entry)

	if blocks != nil {
		t.Errorf("Expected nil for malformed JSON, got %v", blocks)
	}
}

func TestExtractThinkingBlocks_EmptyThinking(t *testing.T) {
	entry := models.ConversationEntry{
		Type: models.EntryTypeAssistant,
		Message: json.RawMessage(`{
			"role": "assistant",
			"content": [
				{"type": "thinking", "text": ""},
				{"type": "thinking", "text": "Valid thinking"}
			]
		}`),
	}

	blocks := ExtractThinkingBlocks(entry)

	// Should only include non-empty thinking blocks
	if len(blocks) != 1 {
		t.Fatalf("Expected 1 thinking block (non-empty), got %d", len(blocks))
	}
	if blocks[0].Content != "Valid thinking" {
		t.Errorf("Thinking block content = %q, want %q", blocks[0].Content, "Valid thinking")
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		chars    int
		expected string
	}{
		{0, "0 chars"},
		{50, "50 chars"},
		{999, "999 chars"},
		{1000, "1.0K chars"},
		{1500, "1.5K chars"},
		{9999, "10.0K chars"},
		{10000, "10K chars"},
		{50000, "50K chars"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatSize(tt.chars)
			if result != tt.expected {
				t.Errorf("formatSize(%d) = %q, want %q", tt.chars, result, tt.expected)
			}
		})
	}
}

func TestFormatCharCount(t *testing.T) {
	tests := []struct {
		inputChars  int
		outputChars int
		expected    string
	}{
		{100, 0, "100 chars in"},
		{100, 200, "100 chars in / 200 chars out"},
		{5000, 0, "5.0K chars in"},
		{5000, 10000, "5.0K chars in / 10K chars out"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatCharCount(tt.inputChars, tt.outputChars)
			if result != tt.expected {
				t.Errorf("formatCharCount(%d, %d) = %q, want %q", tt.inputChars, tt.outputChars, result, tt.expected)
			}
		})
	}
}

func TestRenderToolOverlay_NilInput(t *testing.T) {
	tool := models.ToolUse{
		ID:    "toolu_nil",
		Name:  "Bash",
		Input: nil,
	}

	html := RenderToolOverlay(tool, models.ToolResult{}, false)

	// Should handle nil input gracefully
	if !strings.Contains(html, "{}") {
		t.Error("Nil input should render as empty JSON object")
	}
}

func TestRenderToolOverlay_ComplexInput(t *testing.T) {
	tool := models.ToolUse{
		ID:   "toolu_complex",
		Name: "Bash",
		Input: map[string]any{
			"command":     "complex command",
			"description": "A description",
			"nested": map[string]any{
				"key": "value",
			},
		},
	}

	html := RenderToolOverlay(tool, models.ToolResult{}, false)

	// Should render complex input as JSON
	if !strings.Contains(html, "complex command") {
		t.Error("Missing command in rendered input")
	}
	if !strings.Contains(html, "description") {
		t.Error("Missing description in rendered input")
	}
}

func TestToolRegistry_Completeness(t *testing.T) {
	// Ensure commonly used tools are in the registry
	requiredTools := []string{
		"Bash", "Read", "Write", "Edit", "Grep", "Glob", "Task",
		"WebFetch", "WebSearch", "NotebookEdit",
	}

	for _, tool := range requiredTools {
		info := GetToolInfo(tool)
		if info.ColorClass == "tool-unknown" {
			t.Errorf("Tool %q should be in registry but has unknown class", tool)
		}
	}
}
