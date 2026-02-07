package export

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/randlee/claude-history/pkg/agent"
	"github.com/randlee/claude-history/pkg/models"
)

// makeMessage converts a value to json.RawMessage for test entries.
func makeMessage(v interface{}) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}

// TestIntegration_FullHTMLExport tests the complete HTML export workflow with all Phase 10 features.
func TestIntegration_FullHTMLExport(t *testing.T) {
	// Create test entries with various types
	entries := []models.ConversationEntry{
		{
			Type:      models.EntryTypeUser,
			Timestamp: "2026-02-01T10:00:00Z",
			SessionID: "test-session-001",
			UUID:      "entry-1",
			Message:   makeMessage("Create a Go CLI application"),
		},
		{
			Type:      models.EntryTypeAssistant,
			Timestamp: "2026-02-01T10:00:05Z",
			SessionID: "test-session-001",
			UUID:      "entry-2",
			Message: makeMessage([]interface{}{
				map[string]interface{}{
					"type": "text",
					"text": "I'll help you create a Go CLI application. Let me start by reading the project structure.",
				},
				map[string]interface{}{
					"type":  "tool_use",
					"id":    "toolu_read001",
					"name":  "Read",
					"input": map[string]interface{}{"file_path": "/project/main.go"},
				},
			}),
		},
		{
			Type:      models.EntryTypeUser,
			Timestamp: "2026-02-01T10:00:10Z",
			SessionID: "test-session-001",
			UUID:      "entry-3",
			Message: makeMessage([]interface{}{
				map[string]interface{}{
					"type":        "tool_result",
					"tool_use_id": "toolu_read001",
					"content":     "package main\n\nfunc main() {\n\tprintln(\"Hello\")\n}",
				},
			}),
		},
		{
			Type:      models.EntryTypeAssistant,
			Timestamp: "2026-02-01T10:00:15Z",
			SessionID: "test-session-001",
			UUID:      "entry-4",
			Message: makeMessage([]interface{}{
				map[string]interface{}{
					"type": "text",
					"text": "I found the main.go file. Now let me write an updated version with proper CLI handling:\n\n```go\npackage main\n\nimport \"flag\"\n\nfunc main() {\n\tname := flag.String(\"name\", \"World\", \"Name to greet\")\n\tflag.Parse()\n\tprintln(\"Hello,\", *name)\n}\n```",
				},
				map[string]interface{}{
					"type":  "tool_use",
					"id":    "toolu_write001",
					"name":  "Write",
					"input": map[string]interface{}{"file_path": "/project/main.go", "content": "package main\n\nimport \"flag\"\n\nfunc main() {\n\tname := flag.String(\"name\", \"World\", \"Name to greet\")\n\tflag.Parse()\n\tprintln(\"Hello,\", *name)\n}\n"},
				},
			}),
		},
		{
			Type:      models.EntryTypeQueueOperation,
			Timestamp: "2026-02-01T10:00:20Z",
			SessionID: "test-session-001",
			UUID:      "entry-5",
			AgentID:   "agent-001",
		},
	}

	// Create test agents
	agents := []*agent.TreeNode{
		{
			SessionID:  "test-session-001",
			AgentID:    "agent-001",
			EntryCount: 5,
		},
	}

	// Test HTML rendering
	html, err := RenderConversation(entries, agents)
	if err != nil {
		t.Fatalf("RenderConversation failed: %v", err)
	}

	// Verify all Phase 10 features are present

	// Sprint 10a: CSS Variable System
	t.Run("CSSVariableSystem", func(t *testing.T) {
		if !strings.Contains(html, "style.css") {
			t.Error("Missing CSS file reference")
		}
	})

	// Sprint 10b: Chat Bubble Layout
	t.Run("ChatBubbleLayout", func(t *testing.T) {
		if !strings.Contains(html, "message-row") {
			t.Error("Missing message-row class for chat bubble layout")
		}
		if !strings.Contains(html, "message-bubble") {
			t.Error("Missing message-bubble class")
		}
		if !strings.Contains(html, `class="avatar`) {
			t.Error("Missing avatar element")
		}
	})

	// Sprint 10c: Copy-to-clipboard
	t.Run("CopyToClipboard", func(t *testing.T) {
		if !strings.Contains(html, "clipboard.js") {
			t.Error("Missing clipboard.js script reference")
		}
		if !strings.Contains(html, "copy-btn") {
			t.Error("Missing copy button class")
		}
		if !strings.Contains(html, "data-copy-text") {
			t.Error("Missing data-copy-text attribute for clipboard")
		}
	})

	// Sprint 10d: Color-coded overlays (tool calls)
	t.Run("ToolCallOverlays", func(t *testing.T) {
		if !strings.Contains(html, "tool-call") {
			t.Error("Missing tool-call class for tool overlays")
		}
		if !strings.Contains(html, "tool-header") {
			t.Error("Missing tool-header class")
		}
	})

	// Sprint 10e: Syntax highlighting (code blocks in markdown)
	t.Run("SyntaxHighlighting", func(t *testing.T) {
		// The markdown renderer should produce code-block elements
		if !strings.Contains(html, "code-block") && !strings.Contains(html, "<pre") {
			t.Error("Missing code block rendering")
		}
	})

	// Sprint 10f: Deep dive navigation
	t.Run("DeepDiveNavigation", func(t *testing.T) {
		if !strings.Contains(html, "navigation.js") {
			t.Error("Missing navigation.js script reference")
		}
		if !strings.Contains(html, "breadcrumbs") {
			t.Error("Missing breadcrumbs navigation")
		}
		if !strings.Contains(html, `aria-label="Navigation breadcrumbs"`) {
			t.Error("Missing ARIA label for breadcrumbs")
		}
	})

	// Sprint 10g: Interactive controls
	t.Run("InteractiveControls", func(t *testing.T) {
		if !strings.Contains(html, "controls.js") {
			t.Error("Missing controls.js script reference")
		}
		if !strings.Contains(html, "expand-all-btn") {
			t.Error("Missing expand-all button")
		}
		if !strings.Contains(html, "collapse-all-btn") {
			t.Error("Missing collapse-all button")
		}
		if !strings.Contains(html, "search-box") {
			t.Error("Missing search box")
		}
		if !strings.Contains(html, "Ctrl+K") {
			t.Error("Missing keyboard shortcut hint")
		}
	})

	// Sprint 10h: Header/footer with metadata
	t.Run("HeaderFooterMetadata", func(t *testing.T) {
		if !strings.Contains(html, "page-header") {
			t.Error("Missing page-header")
		}
		if !strings.Contains(html, "page-footer") {
			t.Error("Missing page-footer")
		}
		if !strings.Contains(html, "session-metadata") {
			t.Error("Missing session-metadata section")
		}
		if !strings.Contains(html, "Claude Code Session") {
			t.Error("Missing session title in header")
		}
		if !strings.Contains(html, "Keyboard Shortcuts") {
			t.Error("Missing keyboard shortcuts section in footer")
		}
	})

	// Accessibility
	t.Run("Accessibility", func(t *testing.T) {
		if !strings.Contains(html, `role="toolbar"`) {
			t.Error("Missing ARIA role for controls")
		}
		if !strings.Contains(html, `aria-label="Search messages"`) {
			t.Error("Missing ARIA label for search")
		}
		if !strings.Contains(html, `aria-live="polite"`) {
			t.Error("Missing aria-live for search results")
		}
		if !strings.Contains(html, `aria-hidden="true"`) {
			t.Error("Missing aria-hidden for decorative elements")
		}
	})

	// HTML structure
	t.Run("HTMLStructure", func(t *testing.T) {
		if !strings.Contains(html, "<!DOCTYPE html>") {
			t.Error("Missing DOCTYPE declaration")
		}
		if !strings.Contains(html, `<meta charset="UTF-8">`) {
			t.Error("Missing charset meta tag")
		}
		if !strings.Contains(html, "</html>") {
			t.Error("Missing closing html tag")
		}
	})
}

// TestIntegration_AgentResurrectionWorkflow tests the agent ID copy flow.
func TestIntegration_AgentResurrectionWorkflow(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			Type:      models.EntryTypeQueueOperation,
			Timestamp: "2026-02-01T10:00:00Z",
			SessionID: "test-session-001",
			UUID:      "queue-1",
			AgentID:   "resurrection-agent-abcd1234",
		},
	}

	agents := []*agent.TreeNode{
		{
			SessionID:  "test-session-001",
			AgentID:    "resurrection-agent-abcd1234",
			EntryCount: 10,
		},
	}

	html, err := RenderConversation(entries, agents)
	if err != nil {
		t.Fatalf("RenderConversation failed: %v", err)
	}

	// Verify agent ID is copyable
	if !strings.Contains(html, `data-copy-text="resurrection-agent-abcd1234"`) {
		t.Error("Agent ID should be copyable via data-copy-text attribute")
	}

	// Verify copy button exists for agent
	if !strings.Contains(html, `data-copy-type="agent-id"`) {
		t.Error("Missing agent-id copy button")
	}
}

// TestIntegration_SearchFunctionality verifies search components are present.
func TestIntegration_SearchFunctionality(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			Type:      models.EntryTypeUser,
			Timestamp: "2026-02-01T10:00:00Z",
			SessionID: "test-session-001",
			UUID:      "entry-1",
			Message:   makeMessage("Search test message with unique keyword: findme123"),
		},
		{
			Type:      models.EntryTypeAssistant,
			Timestamp: "2026-02-01T10:00:05Z",
			SessionID: "test-session-001",
			UUID:      "entry-2",
			Message:   makeMessage("Response with different content"),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation failed: %v", err)
	}

	// Search box elements
	if !strings.Contains(html, `id="search-box"`) {
		t.Error("Missing search box with ID")
	}
	if !strings.Contains(html, `type="search"`) {
		t.Error("Search input should have type=search")
	}
	if !strings.Contains(html, `placeholder="Search messages..."`) {
		t.Error("Missing search placeholder")
	}

	// Navigation buttons
	if !strings.Contains(html, `id="search-prev-btn"`) {
		t.Error("Missing previous match button")
	}
	if !strings.Contains(html, `id="search-next-btn"`) {
		t.Error("Missing next match button")
	}

	// Results display
	if !strings.Contains(html, `class="search-results"`) {
		t.Error("Missing search results display")
	}
}

// TestIntegration_KeyboardShortcuts verifies keyboard shortcut documentation.
func TestIntegration_KeyboardShortcuts(t *testing.T) {
	html, err := RenderConversation(nil, nil)
	if err != nil {
		t.Fatalf("RenderConversation failed: %v", err)
	}

	// Shortcut hints on buttons
	if !strings.Contains(html, `data-shortcut="Ctrl+K"`) {
		t.Error("Missing Ctrl+K shortcut on expand button")
	}
	if !strings.Contains(html, `data-shortcut="Ctrl+F"`) {
		t.Error("Missing Ctrl+F shortcut on search")
	}

	// Footer documentation
	if !strings.Contains(html, "<kbd>Ctrl</kbd>") {
		t.Error("Missing keyboard shortcut documentation in footer")
	}
	if !strings.Contains(html, "<kbd>Esc</kbd>") {
		t.Error("Missing Escape key documentation")
	}
}

// TestIntegration_SessionStats verifies session statistics computation.
func TestIntegration_SessionStats(t *testing.T) {
	entries := []models.ConversationEntry{
		{Type: models.EntryTypeUser, UUID: "1"},
		{Type: models.EntryTypeAssistant, UUID: "2", Message: makeMessage([]interface{}{
			map[string]interface{}{"type": "text", "text": "Hello"},
			map[string]interface{}{"type": "tool_use", "id": "tool1", "name": "Bash", "input": map[string]interface{}{}},
			map[string]interface{}{"type": "tool_use", "id": "tool2", "name": "Read", "input": map[string]interface{}{}},
		})},
		{Type: models.EntryTypeUser, UUID: "3"},
		{Type: models.EntryTypeAssistant, UUID: "4", Message: makeMessage([]interface{}{
			map[string]interface{}{"type": "tool_use", "id": "tool3", "name": "Write", "input": map[string]interface{}{}},
		})},
		{Type: models.EntryTypeSystem, UUID: "5"},
	}

	agents := []*agent.TreeNode{
		{AgentID: "agent-1", EntryCount: 5},
		{AgentID: "agent-2", EntryCount: 3},
	}

	stats := ComputeSessionStats(entries, agents)

	// 2 user + 2 assistant = 4 messages
	if stats.MessageCount != 4 {
		t.Errorf("Expected MessageCount=4, got %d", stats.MessageCount)
	}

	// 3 tool calls total
	if stats.ToolCallCount != 3 {
		t.Errorf("Expected ToolCallCount=3, got %d", stats.ToolCallCount)
	}

	// 2 agents
	if stats.AgentCount != 2 {
		t.Errorf("Expected AgentCount=2, got %d", stats.AgentCount)
	}

	// Export time should be set
	if stats.ExportTime == "" {
		t.Error("ExportTime should be set")
	}
}

// TestIntegration_StaticAssetsComplete verifies all static assets are written.
func TestIntegration_StaticAssetsComplete(t *testing.T) {
	tmpDir := t.TempDir()

	err := WriteStaticAssets(tmpDir)
	if err != nil {
		t.Fatalf("WriteStaticAssets failed: %v", err)
	}

	staticDir := filepath.Join(tmpDir, "static")

	// All required files
	requiredFiles := []string{
		"style.css",
		"script.js",
		"clipboard.js",
		"controls.js",
		"navigation.js",
	}

	for _, file := range requiredFiles {
		filePath := filepath.Join(staticDir, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("Missing required static file: %s", file)
		}

		// Verify non-empty
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Errorf("Failed to read %s: %v", file, err)
		}
		if len(content) == 0 {
			t.Errorf("Static file %s is empty", file)
		}
	}
}

// TestIntegration_CSSPrintStyles verifies print stylesheet is present.
func TestIntegration_CSSPrintStyles(t *testing.T) {
	css := GetStyleCSS()

	// Multiple @media print blocks should exist
	if !strings.Contains(css, "@media print") {
		t.Error("Missing @media print styles")
	}

	// Print styles should hide controls
	if !strings.Contains(css, ".controls") {
		t.Error("Missing controls style rule")
	}

	// Print should expand collapsibles
	if !strings.Contains(css, "display: block") || !strings.Contains(css, ".tool-body.hidden") {
		t.Error("Print styles should expand hidden tool bodies")
	}

	// Page breaks
	if !strings.Contains(css, "break-inside: avoid") && !strings.Contains(css, "page-break") {
		t.Error("Print styles should handle page breaks")
	}

	// Remove animations
	if !strings.Contains(css, "animation: none") {
		t.Error("Print styles should disable animations")
	}
}

// TestIntegration_CSSResponsiveStyles verifies responsive design.
func TestIntegration_CSSResponsiveStyles(t *testing.T) {
	css := GetStyleCSS()

	// Mobile breakpoint
	if !strings.Contains(css, "@media (max-width: 768px)") {
		t.Error("Missing mobile responsive styles")
	}

	// Dark mode
	if !strings.Contains(css, "@media (prefers-color-scheme: dark)") {
		t.Error("Missing dark mode styles")
	}
}

// TestIntegration_CSSAccessibility verifies accessibility-related styles.
func TestIntegration_CSSAccessibility(t *testing.T) {
	css := GetStyleCSS()

	// Screen reader utility class
	if !strings.Contains(css, ".sr-only") {
		t.Error("Missing .sr-only utility class for screen readers")
	}

	// Focus styles
	if !strings.Contains(css, ":focus") {
		t.Error("Missing focus styles for keyboard navigation")
	}
}

// TestIntegration_JSModulesComplete verifies all JS modules have required functions.
func TestIntegration_JSModulesComplete(t *testing.T) {
	// Clipboard.js
	clipboard := GetClipboardJS()
	clipboardFunctions := []string{
		"copyToClipboard",
		"handleCopyClick",
		"initCopyButtons",
		"showCopySuccess",
		"showCopyError",
	}
	for _, fn := range clipboardFunctions {
		if !strings.Contains(clipboard, fn) {
			t.Errorf("clipboard.js missing function: %s", fn)
		}
	}

	// Controls.js
	controls := GetControlsJS()
	controlsFunctions := []string{
		"expandAllTools",
		"collapseAllTools",
		"performSearch",
		"clearSearch",
		"initKeyboardShortcuts",
		"ControlsAPI",
	}
	for _, fn := range controlsFunctions {
		if !strings.Contains(controls, fn) {
			t.Errorf("controls.js missing function: %s", fn)
		}
	}

	// Navigation.js
	navigation := GetNavigationJS()
	navigationFunctions := []string{
		"updateBreadcrumbs",
		"expandSubagent",
		"collapseSubagent",
		"scrollToAgent",
		"NavigationAPI",
		"deepDiveAgent",
	}
	for _, fn := range navigationFunctions {
		if !strings.Contains(navigation, fn) {
			t.Errorf("navigation.js missing function: %s", fn)
		}
	}
}

// TestIntegration_JSDebounce verifies search debounce is implemented.
func TestIntegration_JSDebounce(t *testing.T) {
	controls := GetControlsJS()

	// Look for debounce implementation
	if !strings.Contains(controls, "debounce") && !strings.Contains(controls, "setTimeout") {
		t.Error("Search should have debounce/setTimeout for performance")
	}
}

// TestIntegration_ManifestGeneration verifies manifest structure via session stats.
func TestIntegration_ManifestGeneration(t *testing.T) {
	entries := []models.ConversationEntry{
		{Type: models.EntryTypeUser, UUID: "1", Timestamp: "2026-02-01T10:00:00Z"},
		{Type: models.EntryTypeAssistant, UUID: "2", Timestamp: "2026-02-01T10:00:05Z"},
	}

	agents := []*agent.TreeNode{
		{AgentID: "agent-1", EntryCount: 5},
	}

	// Test using ComputeSessionStats which is used internally for manifests
	stats := ComputeSessionStats(entries, agents)

	// Verify stats are computed correctly
	if stats.MessageCount != 2 {
		t.Errorf("Expected MessageCount=2, got %d", stats.MessageCount)
	}
	if stats.AgentCount != 1 {
		t.Errorf("Expected AgentCount=1, got %d", stats.AgentCount)
	}
	if stats.ExportTime == "" {
		t.Error("ExportTime should be set")
	}
}

// TestIntegration_ToolResultMatching verifies tool calls are matched with results.
func TestIntegration_ToolResultMatching(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			Type:      models.EntryTypeAssistant,
			Timestamp: "2026-02-01T10:00:00Z",
			UUID:      "assistant-1",
			Message: makeMessage([]interface{}{
				map[string]interface{}{
					"type":  "tool_use",
					"id":    "toolu_match001",
					"name":  "Read",
					"input": map[string]interface{}{"file_path": "/test.go"},
				},
			}),
		},
		{
			Type:      models.EntryTypeUser,
			Timestamp: "2026-02-01T10:00:05Z",
			UUID:      "user-1",
			Message: makeMessage([]interface{}{
				map[string]interface{}{
					"type":        "tool_result",
					"tool_use_id": "toolu_match001",
					"content":     "File contents here",
				},
			}),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation failed: %v", err)
	}

	// Tool result should be rendered
	if !strings.Contains(html, "File contents here") {
		t.Error("Tool result content should be included in the rendered HTML")
	}
}

// TestIntegration_XSSPrevention verifies HTML escaping.
func TestIntegration_XSSPrevention(t *testing.T) {
	maliciousContent := `<script>alert('XSS')</script>`
	entries := []models.ConversationEntry{
		{
			Type:      models.EntryTypeUser,
			Timestamp: "2026-02-01T10:00:00Z",
			UUID:      "xss-1",
			Message:   makeMessage(maliciousContent),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation failed: %v", err)
	}

	// Raw script tag should not appear
	if strings.Contains(html, "<script>alert") {
		t.Error("XSS vulnerability: unescaped script tag in output")
	}

	// Should be escaped
	if !strings.Contains(html, "&lt;script&gt;") {
		t.Error("Script tag should be HTML escaped")
	}
}

// TestIntegration_LargeConversation tests performance with many entries.
func TestIntegration_LargeConversation(t *testing.T) {
	// Create 1000 entries
	entries := make([]models.ConversationEntry, 1000)
	for i := 0; i < 1000; i++ {
		entryType := models.EntryTypeUser
		if i%2 == 1 {
			entryType = models.EntryTypeAssistant
		}
		entries[i] = models.ConversationEntry{
			Type:      entryType,
			Timestamp: time.Now().Format(time.RFC3339Nano),
			UUID:      "entry-" + string(rune('0'+i%10)),
			Message:   makeMessage("Message content " + string(rune('0'+i%10))),
		}
	}

	start := time.Now()
	html, err := RenderConversation(entries, nil)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("RenderConversation failed: %v", err)
	}

	// Should complete in reasonable time (under 5 seconds)
	if duration > 5*time.Second {
		t.Errorf("Rendering 1000 entries took too long: %v", duration)
	}

	// Output should be non-empty
	if len(html) < 1000 {
		t.Error("Output HTML seems too small for 1000 entries")
	}
}

// TestIntegration_EmptyConversation tests handling of empty input.
func TestIntegration_EmptyConversation(t *testing.T) {
	html, err := RenderConversation(nil, nil)
	if err != nil {
		t.Fatalf("RenderConversation with nil entries failed: %v", err)
	}

	// Should still produce valid HTML structure
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("Empty conversation should still have DOCTYPE")
	}
	if !strings.Contains(html, "</html>") {
		t.Error("Empty conversation should still have closing html tag")
	}
}

// TestIntegration_SpecialCharacters tests handling of special characters.
func TestIntegration_SpecialCharacters(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			Type:    models.EntryTypeUser,
			UUID:    "special-1",
			Message: makeMessage("Test with special chars: & < > \" ' \u00e9 \u4e2d\u6587 emoji: \U0001f600"),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation failed: %v", err)
	}

	// Ampersand should be escaped
	if !strings.Contains(html, "&amp;") {
		t.Error("Ampersand should be escaped to &amp;")
	}

	// Unicode should be preserved
	if !strings.Contains(html, "\u00e9") && !strings.Contains(html, "&#233;") {
		t.Error("Unicode characters should be preserved")
	}
}

// TestIntegration_AllEntryTypes tests rendering of all entry types.
func TestIntegration_AllEntryTypes(t *testing.T) {
	entries := []models.ConversationEntry{
		{Type: models.EntryTypeUser, UUID: "1", Message: makeMessage("User message")},
		{Type: models.EntryTypeAssistant, UUID: "2", Message: makeMessage("Assistant message")},
		{Type: models.EntryTypeSystem, UUID: "3", Message: makeMessage("System message")},
		{Type: models.EntryTypeQueueOperation, UUID: "4", AgentID: "test-agent"},
		{Type: models.EntryTypeSummary, UUID: "5", Message: makeMessage("Summary")},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation failed: %v", err)
	}

	// Entry types with content should have corresponding CSS classes
	entryClasses := []string{
		"message-row user",
		"message-row assistant",
		"message-row system",
		// queue-operation with no message content is skipped (only subagent placeholder rendered)
		"message-row summary",
	}

	for _, class := range entryClasses {
		if !strings.Contains(html, class) {
			t.Errorf("Missing entry type class: %s", class)
		}
	}

	// Queue operation without message content should still render subagent placeholder
	if !strings.Contains(html, `class="subagent collapsible collapsed"`) {
		t.Error("Queue operation should render subagent placeholder with collapsible collapsed classes")
	}
	if !strings.Contains(html, `data-agent-id="test-agent"`) {
		t.Error("Subagent placeholder should have agent ID")
	}
}

// TestIntegration_FilePathCopyButton tests file path copy functionality.
func TestIntegration_FilePathCopyButton(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			Type: models.EntryTypeAssistant,
			UUID: "1",
			Message: makeMessage([]interface{}{
				map[string]interface{}{
					"type":  "tool_use",
					"id":    "toolu_file001",
					"name":  "Read",
					"input": map[string]interface{}{"file_path": "/home/user/project/src/main.go"},
				},
			}),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation failed: %v", err)
	}

	// File path should be copyable
	if !strings.Contains(html, `data-copy-type="file-path"`) {
		t.Error("File path copy button should be present for Read tool")
	}
	if !strings.Contains(html, "/home/user/project/src/main.go") {
		t.Error("File path should appear in the HTML")
	}
}

// TestIntegration_ExportFormatVersion verifies version is included in output.
func TestIntegration_ExportFormatVersion(t *testing.T) {
	html, err := RenderConversation(nil, nil)
	if err != nil {
		t.Fatalf("RenderConversation failed: %v", err)
	}

	// Export version should be in footer
	if !strings.Contains(html, "Export format version") {
		t.Error("Export format version should be documented in footer")
	}
	if !strings.Contains(html, ExportFormatVersion) {
		t.Errorf("Export format version %s should appear in HTML", ExportFormatVersion)
	}
}

// TestIntegration_ScriptLoadOrder verifies scripts are loaded in correct order.
func TestIntegration_ScriptLoadOrder(t *testing.T) {
	html, err := RenderConversation(nil, nil)
	if err != nil {
		t.Fatalf("RenderConversation failed: %v", err)
	}

	// Find positions of script tags
	scriptPos := strings.Index(html, `src="static/script.js"`)
	clipboardPos := strings.Index(html, `src="static/clipboard.js"`)
	controlsPos := strings.Index(html, `src="static/controls.js"`)
	navigationPos := strings.Index(html, `src="static/navigation.js"`)

	// All scripts should be present
	if scriptPos == -1 {
		t.Error("Missing script.js reference")
	}
	if clipboardPos == -1 {
		t.Error("Missing clipboard.js reference")
	}
	if controlsPos == -1 {
		t.Error("Missing controls.js reference")
	}
	if navigationPos == -1 {
		t.Error("Missing navigation.js reference")
	}

	// script.js should come first (base functionality)
	if scriptPos > clipboardPos || scriptPos > controlsPos || scriptPos > navigationPos {
		t.Error("script.js should be loaded before other JavaScript modules")
	}
}

// TestIntegration_CSSColorContrast verifies color contrast variables exist.
func TestIntegration_CSSColorContrast(t *testing.T) {
	css := GetStyleCSS()

	// HSL color palette should exist for accessibility
	hslPatterns := []string{
		"--neutral-",
		"--blue-",
		"--green-",
		"--red-",
		"--orange-",
	}

	for _, pattern := range hslPatterns {
		if !strings.Contains(css, pattern) {
			t.Errorf("Missing HSL color palette: %s", pattern)
		}
	}

	// Semantic color variables
	semanticColors := []string{
		"--text-primary",
		"--text-secondary",
		"--bg-primary",
		"--bg-secondary",
		"--border-primary",
	}

	for _, color := range semanticColors {
		if !strings.Contains(css, color) {
			t.Errorf("Missing semantic color variable: %s", color)
		}
	}
}
