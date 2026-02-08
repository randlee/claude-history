package export

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/randlee/claude-history/pkg/agent"
	"github.com/randlee/claude-history/pkg/models"
)

func TestRenderCopyButton_BasicOutput(t *testing.T) {
	result := renderCopyButton("test-value", "test-type", "Test tooltip")

	// Check button structure
	if !strings.Contains(result, `class="copy-btn"`) {
		t.Error("Copy button missing copy-btn class")
	}
	if !strings.Contains(result, `data-copy-text="test-value"`) {
		t.Error("Copy button missing data-copy-text attribute")
	}
	if !strings.Contains(result, `data-copy-type="test-type"`) {
		t.Error("Copy button missing data-copy-type attribute")
	}
	if !strings.Contains(result, `title="Test tooltip"`) {
		t.Error("Copy button missing title attribute")
	}
	if !strings.Contains(result, `<button`) {
		t.Error("Copy button should be a button element")
	}
	if !strings.Contains(result, `class="copy-icon"`) {
		t.Error("Copy button missing copy-icon span")
	}
}

func TestRenderCopyButton_EmptyText(t *testing.T) {
	result := renderCopyButton("", "test-type", "Test tooltip")

	if result != "" {
		t.Errorf("renderCopyButton with empty text should return empty string, got %q", result)
	}
}

func TestRenderCopyButton_HTMLEscaping(t *testing.T) {
	tests := []struct {
		name          string
		text          string
		copyType      string
		tooltip       string
		shouldContain string
		shouldNotHave string
	}{
		{
			name:          "text with quotes",
			text:          `path/to/"file".txt`,
			copyType:      "file-path",
			tooltip:       "Copy path",
			shouldContain: `&#34;file&#34;`,
			shouldNotHave: `"file"`,
		},
		{
			name:          "text with angle brackets",
			text:          "<script>alert(1)</script>",
			copyType:      "test",
			tooltip:       "Copy",
			shouldContain: `&lt;script&gt;`,
			shouldNotHave: `<script>`,
		},
		{
			name:          "text with ampersand",
			text:          "a&b",
			copyType:      "test",
			tooltip:       "Copy",
			shouldContain: `a&amp;b`,
			shouldNotHave: "",
		},
		{
			name:          "tooltip with special chars",
			text:          "value",
			copyType:      "test",
			tooltip:       `Copy "quoted" value`,
			shouldContain: `title="Copy &#34;quoted&#34; value"`,
			shouldNotHave: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderCopyButton(tt.text, tt.copyType, tt.tooltip)

			if !strings.Contains(result, tt.shouldContain) {
				t.Errorf("Result should contain %q, got %s", tt.shouldContain, result)
			}
			if tt.shouldNotHave != "" && strings.Contains(result, tt.shouldNotHave) {
				t.Errorf("Result should not contain unescaped %q", tt.shouldNotHave)
			}
		})
	}
}

func TestRenderCopyButton_CopyTypes(t *testing.T) {
	copyTypes := []string{
		"agent-id",
		"file-path",
		"session-id",
		"tool-id",
		"jsonl-path",
	}

	for _, copyType := range copyTypes {
		t.Run(copyType, func(t *testing.T) {
			result := renderCopyButton("value", copyType, "tooltip")

			expected := `data-copy-type="` + copyType + `"`
			if !strings.Contains(result, expected) {
				t.Errorf("Result should contain %q", expected)
			}
		})
	}
}

func TestRenderConversation_CopyButtonForAgentID(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-001",
			SessionID: "session-001",
			AgentID:   "a12eb64abc123",
			Type:      models.EntryTypeAssistant,
			Timestamp: "2026-01-31T10:00:00Z",
			Message:   json.RawMessage(`{"role": "assistant", "content": [{"type": "text", "text": "Agent response"}]}`),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	// Check for copy button in agent ID (now includes context, not just ID)
	if !strings.Contains(html, `data-copy-type="agent-id"`) {
		t.Error("HTML missing agent-id copy type")
	}
	if !strings.Contains(html, `title="Copy agent details"`) {
		t.Error("HTML missing copy agent details tooltip")
	}
	// The copy-text should now include context with the agent ID
	if !strings.Contains(html, "a12eb64abc123") {
		t.Error("HTML missing agent ID in copy context")
	}
}

func TestRenderConversation_CopyButtonForToolID(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-001",
			SessionID: "session-001",
			Type:      models.EntryTypeAssistant,
			Timestamp: "2026-01-31T10:00:00Z",
			Message: json.RawMessage(`{
				"role": "assistant",
				"content": [
					{"type": "tool_use", "id": "toolu_01ABC123", "name": "Bash", "input": {"command": "ls"}}
				]
			}`),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	// Check for copy button for tool ID
	if !strings.Contains(html, `data-copy-text="toolu_01ABC123"`) {
		t.Error("HTML missing copy button with tool ID")
	}
	if !strings.Contains(html, `data-copy-type="tool-id"`) {
		t.Error("HTML missing tool-id copy type")
	}
	if !strings.Contains(html, `title="Copy tool ID"`) {
		t.Error("HTML missing copy tool ID tooltip")
	}
}

func TestRenderConversation_CopyButtonForFilePath(t *testing.T) {
	tests := []struct {
		name       string
		toolName   string
		input      map[string]any
		expectPath string
	}{
		{
			name:       "Read tool",
			toolName:   "Read",
			input:      map[string]any{"file_path": "/path/to/file.go"},
			expectPath: "/path/to/file.go",
		},
		{
			name:       "Write tool",
			toolName:   "Write",
			input:      map[string]any{"file_path": "/output/new-file.txt"},
			expectPath: "/output/new-file.txt",
		},
		{
			name:       "Edit tool",
			toolName:   "Edit",
			input:      map[string]any{"file_path": "/src/main.go"},
			expectPath: "/src/main.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputJSON, _ := json.Marshal(tt.input)
			entries := []models.ConversationEntry{
				{
					UUID:      "uuid-001",
					SessionID: "session-001",
					Type:      models.EntryTypeAssistant,
					Timestamp: "2026-01-31T10:00:00Z",
					Message: json.RawMessage(`{
						"role": "assistant",
						"content": [
							{"type": "tool_use", "id": "toolu_test", "name": "` + tt.toolName + `", "input": ` + string(inputJSON) + `}
						]
					}`),
				},
			}

			html, err := RenderConversation(entries, nil)
			if err != nil {
				t.Fatalf("RenderConversation() error = %v", err)
			}

			if !strings.Contains(html, `data-copy-text="`+escapeHTML(tt.expectPath)+`"`) {
				t.Errorf("HTML missing copy button with file path %s", tt.expectPath)
			}
			if !strings.Contains(html, `data-copy-type="file-path"`) {
				t.Error("HTML missing file-path copy type")
			}
			if !strings.Contains(html, `title="Copy file path"`) {
				t.Error("HTML missing copy file path tooltip")
			}
		})
	}
}

func TestRenderConversation_NoFilePathForNonFileTool(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-001",
			SessionID: "session-001",
			Type:      models.EntryTypeAssistant,
			Timestamp: "2026-01-31T10:00:00Z",
			Message: json.RawMessage(`{
				"role": "assistant",
				"content": [
					{"type": "tool_use", "id": "toolu_test", "name": "Bash", "input": {"command": "ls -la"}}
				]
			}`),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	// Should have tool-id but not file-path
	if !strings.Contains(html, `data-copy-type="tool-id"`) {
		t.Error("HTML missing tool-id copy type")
	}
	// Bash doesn't have a file_path, so no file-path button
	if strings.Contains(html, `data-copy-type="file-path"`) {
		t.Error("HTML should not have file-path copy type for Bash tool")
	}
}

func TestRenderSubagentPlaceholder_CopyButton(t *testing.T) {
	agentMap := map[string]int{
		"a12eb64abc123def456": 29,
	}

	html := renderSubagentPlaceholder("a12eb64abc123def456", agentMap, "session-123", "/test/project")

	// Check for copy button with full agent ID in context
	if !strings.Contains(html, `a12eb64abc123def456`) {
		t.Error("Subagent placeholder missing copy button with full agent ID")
	}
	if !strings.Contains(html, `data-copy-type="agent-id"`) {
		t.Error("Subagent placeholder missing agent-id copy type")
	}
	if !strings.Contains(html, `title="Copy agent details"`) {
		t.Error("Subagent placeholder missing copy tooltip")
	}
}

func TestRenderAgentFragment_CopyButtonsForAgentEntries(t *testing.T) {
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
					{"type": "tool_use", "id": "toolu_agent_01", "name": "Read", "input": {"file_path": "/agent/file.go"}}
				]
			}`),
		},
	}

	html, err := RenderAgentFragment("a12eb64", entries)
	if err != nil {
		t.Fatalf("RenderAgentFragment() error = %v", err)
	}

	// Check for tool ID copy button
	if !strings.Contains(html, `data-copy-text="toolu_agent_01"`) {
		t.Error("Fragment missing tool ID copy button")
	}
	// Check for file path copy button
	if !strings.Contains(html, `data-copy-text="/agent/file.go"`) {
		t.Error("Fragment missing file path copy button")
	}
}

func TestExtractFilePath_AllFileTools(t *testing.T) {
	tests := []struct {
		name       string
		toolName   string
		input      map[string]any
		expectPath string
	}{
		{
			name:       "Read with file_path",
			toolName:   "Read",
			input:      map[string]any{"file_path": "/path/file.go"},
			expectPath: "/path/file.go",
		},
		{
			name:       "Write with file_path",
			toolName:   "Write",
			input:      map[string]any{"file_path": "/output/file.txt"},
			expectPath: "/output/file.txt",
		},
		{
			name:       "Edit with file_path",
			toolName:   "Edit",
			input:      map[string]any{"file_path": "/src/main.go"},
			expectPath: "/src/main.go",
		},
		{
			name:       "NotebookEdit with notebook_path",
			toolName:   "NotebookEdit",
			input:      map[string]any{"notebook_path": "/notebooks/analysis.ipynb"},
			expectPath: "/notebooks/analysis.ipynb",
		},
		{
			name:       "Bash (no file path)",
			toolName:   "Bash",
			input:      map[string]any{"command": "ls"},
			expectPath: "",
		},
		{
			name:       "Grep (no file path)",
			toolName:   "Grep",
			input:      map[string]any{"pattern": "test"},
			expectPath: "",
		},
		{
			name:       "Read with nil input",
			toolName:   "Read",
			input:      nil,
			expectPath: "",
		},
		{
			name:       "Read with missing file_path",
			toolName:   "Read",
			input:      map[string]any{"other": "value"},
			expectPath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractFilePath(tt.toolName, tt.input)
			if result != tt.expectPath {
				t.Errorf("extractFilePath(%q, %v) = %q, want %q", tt.toolName, tt.input, result, tt.expectPath)
			}
		})
	}
}

func TestRenderToolCall_HasBothToolIDAndFilePathCopyButtons(t *testing.T) {
	tool := models.ToolUse{
		ID:    "toolu_01XYZ",
		Name:  "Read",
		Input: map[string]any{"file_path": "/test/file.go"},
	}

	html := renderToolCall(tool, models.ToolResult{}, false)

	// Should have both tool ID and file path copy buttons
	toolIDCount := strings.Count(html, `data-copy-type="tool-id"`)
	filePathCount := strings.Count(html, `data-copy-type="file-path"`)

	if toolIDCount != 1 {
		t.Errorf("Expected 1 tool-id copy button, got %d", toolIDCount)
	}
	if filePathCount != 1 {
		t.Errorf("Expected 1 file-path copy button, got %d", filePathCount)
	}
}

func TestRenderToolCall_OnlyToolIDForNonFileTool(t *testing.T) {
	tool := models.ToolUse{
		ID:    "toolu_01XYZ",
		Name:  "WebSearch",
		Input: map[string]any{"query": "test query"},
	}

	html := renderToolCall(tool, models.ToolResult{}, false)

	// Should have tool ID but no file path
	if !strings.Contains(html, `data-copy-type="tool-id"`) {
		t.Error("Missing tool-id copy button")
	}
	if strings.Contains(html, `data-copy-type="file-path"`) {
		t.Error("Should not have file-path copy button for WebSearch")
	}
}

func TestGetClipboardJS_ReturnsContent(t *testing.T) {
	content := GetClipboardJS()

	if content == "" {
		t.Fatal("GetClipboardJS() returned empty string")
	}

	// Check for key functions
	if !strings.Contains(content, "copyToClipboard") {
		t.Error("clipboard.js missing copyToClipboard function")
	}
	if !strings.Contains(content, "showCopySuccess") {
		t.Error("clipboard.js missing showCopySuccess function")
	}
	if !strings.Contains(content, "showCopyToast") {
		t.Error("clipboard.js missing showCopyToast function")
	}
	if !strings.Contains(content, "navigator.clipboard") {
		t.Error("clipboard.js missing navigator.clipboard API usage")
	}
	if !strings.Contains(content, "handleCopyClick") {
		t.Error("clipboard.js missing handleCopyClick function")
	}
	if !strings.Contains(content, "initCopyButtons") {
		t.Error("clipboard.js missing initCopyButtons function")
	}
}

func TestHTMLFooter_IncludesClipboardScript(t *testing.T) {
	if !strings.Contains(htmlFooter, `<script src="static/clipboard.js"></script>`) {
		t.Error("HTML footer missing clipboard.js script tag")
	}
}

func TestRenderConversation_MultipleAgentIDCopyButtons(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-001",
			SessionID: "session-001",
			AgentID:   "agent-alpha",
			Type:      models.EntryTypeAssistant,
			Timestamp: "2026-01-31T10:00:00Z",
			Message:   json.RawMessage(`{"role": "assistant", "content": [{"type": "text", "text": "First agent"}]}`),
		},
		{
			UUID:      "uuid-002",
			SessionID: "session-001",
			AgentID:   "agent-beta",
			Type:      models.EntryTypeAssistant,
			Timestamp: "2026-01-31T10:00:05Z",
			Message:   json.RawMessage(`{"role": "assistant", "content": [{"type": "text", "text": "Second agent"}]}`),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	// Should have copy buttons for both agents (in copy context, not exact match)
	if !strings.Contains(html, "agent-alpha") {
		t.Error("Missing copy button context for agent-alpha")
	}
	if !strings.Contains(html, "agent-beta") {
		t.Error("Missing copy button context for agent-beta")
	}
}

func TestRenderConversation_SubagentWithCopyButton(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-001",
			SessionID: "session-001",
			Type:      models.EntryTypeQueueOperation,
			AgentID:   "spawned-agent-123",
			Timestamp: "2026-01-31T10:00:00Z",
			Message:   json.RawMessage(`"Agent spawned"`),
		},
	}

	agents := []*agent.TreeNode{
		{
			AgentID:    "spawned-agent-123",
			SessionID:  "session-001",
			EntryCount: 15,
		},
	}

	html, err := RenderConversation(entries, agents)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	// Subagent header should have copy button with full agent ID context
	if !strings.Contains(html, `spawned-agent-123`) {
		t.Error("Subagent placeholder missing copy button with full agent ID")
	}
	// Should include copy button with context
	if !strings.Contains(html, `class="copy-btn"`) {
		t.Error("Subagent placeholder missing copy button")
	}
}

func TestCSSContainsCopyButtonStyles(t *testing.T) {
	css := GetStyleCSS()

	requiredClasses := []string{
		".copy-btn",
		".copy-btn:hover",
		".copy-btn.copy-success",
		".copy-btn.copy-error",
		".copy-toast",
		".copy-toast-visible",
		".copy-toast-error",
		".copy-icon",
	}

	for _, class := range requiredClasses {
		if !strings.Contains(css, class) {
			t.Errorf("CSS missing required class: %s", class)
		}
	}
}

func TestRenderCopyButton_AllDataAttributesPresent(t *testing.T) {
	result := renderCopyButton("test-text", "test-type", "Test Tooltip")

	// Check all required attributes
	requiredAttrs := []string{
		`class="copy-btn"`,
		`data-copy-text="test-text"`,
		`data-copy-type="test-type"`,
		`title="Test Tooltip"`,
	}

	for _, attr := range requiredAttrs {
		if !strings.Contains(result, attr) {
			t.Errorf("Copy button missing required attribute: %s", attr)
		}
	}
}

func TestRenderToolCall_ToolSummaryAndCopyButtonsStructure(t *testing.T) {
	tool := models.ToolUse{
		ID:    "toolu_struct",
		Name:  "Read",
		Input: map[string]any{"file_path": "/test.go"},
	}

	html := renderToolCall(tool, models.ToolResult{}, false)

	// Check structure - tool summary should be in a span
	if !strings.Contains(html, `class="tool-summary"`) {
		t.Error("Tool header missing tool-summary span")
	}

	// Check tool ID span
	if !strings.Contains(html, `class="tool-id"`) {
		t.Error("Tool header missing tool-id span")
	}

	// Check file path button span
	if !strings.Contains(html, `class="file-path-btn"`) {
		t.Error("Tool header missing file-path-btn span")
	}
}

func TestRenderSubagentPlaceholder_StructureWithCopyButton(t *testing.T) {
	agentMap := map[string]int{
		"test-agent": 10,
	}

	html := renderSubagentPlaceholder("test-agent", agentMap, "session-abc", "/test/path")

	// Check structure
	if !strings.Contains(html, `class="subagent-title"`) {
		t.Error("Subagent header missing subagent-title span")
	}
	if !strings.Contains(html, `class="subagent-meta"`) {
		t.Error("Subagent header missing subagent-meta span")
	}
	if !strings.Contains(html, `class="copy-btn"`) {
		t.Error("Subagent header missing copy button")
	}
}
