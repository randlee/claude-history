package export

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/randlee/claude-history/pkg/agent"
	"github.com/randlee/claude-history/pkg/models"
)

// TestComputeSessionStats_Empty tests stats computation with no entries.
func TestComputeSessionStats_Empty(t *testing.T) {
	stats := ComputeSessionStats(nil, nil)

	if stats == nil {
		t.Fatal("ComputeSessionStats should not return nil")
	}
	if stats.MessageCount != 0 {
		t.Errorf("MessageCount = %d, want 0", stats.MessageCount)
	}
	if stats.AgentCount != 0 {
		t.Errorf("AgentCount = %d, want 0", stats.AgentCount)
	}
	if stats.ToolCallCount != 0 {
		t.Errorf("ToolCallCount = %d, want 0", stats.ToolCallCount)
	}
	if stats.ExportTime == "" {
		t.Error("ExportTime should not be empty")
	}
	if stats.SessionStart != "" {
		t.Error("SessionStart should be empty when no entries")
	}
	if stats.SessionEnd != "" {
		t.Error("SessionEnd should be empty when no entries")
	}
	if stats.Duration != "" {
		t.Error("Duration should be empty when no entries")
	}
}

// TestComputeSessionStats_WithMessages tests message counting.
func TestComputeSessionStats_WithMessages(t *testing.T) {
	entries := []models.ConversationEntry{
		{Type: models.EntryTypeUser, SessionID: "session-123", Timestamp: "2026-02-06T14:23:00Z"},
		{Type: models.EntryTypeAssistant, Timestamp: "2026-02-06T14:23:30Z"},
		{Type: models.EntryTypeUser, Timestamp: "2026-02-06T16:58:00Z"},
		{Type: models.EntryTypeAssistant, Timestamp: "2026-02-06T16:58:30Z"},
		{Type: models.EntryTypeSystem, Timestamp: "2026-02-06T16:59:00Z"},         // Should not count
		{Type: models.EntryTypeQueueOperation, Timestamp: "2026-02-06T16:59:30Z"}, // Should not count
	}

	stats := ComputeSessionStats(entries, nil)

	if stats.MessageCount != 4 {
		t.Errorf("MessageCount = %d, want 4", stats.MessageCount)
	}
	if stats.UserMessages != 2 {
		t.Errorf("UserMessages = %d, want 2", stats.UserMessages)
	}
	if stats.AssistantMessages != 2 {
		t.Errorf("AssistantMessages = %d, want 2", stats.AssistantMessages)
	}
	if stats.SessionID != "session-123" {
		t.Errorf("SessionID = %q, want %q", stats.SessionID, "session-123")
	}
	if stats.SessionStart != "2026-02-06 14:23" {
		t.Errorf("SessionStart = %q, want %q", stats.SessionStart, "2026-02-06 14:23")
	}
	if stats.SessionEnd != "2026-02-06 16:59" {
		t.Errorf("SessionEnd = %q, want %q", stats.SessionEnd, "2026-02-06 16:59")
	}
	if stats.Duration != "2h 36m" {
		t.Errorf("Duration = %q, want %q", stats.Duration, "2h 36m")
	}
}

// TestComputeSessionStats_WithToolCalls tests tool call counting.
func TestComputeSessionStats_WithToolCalls(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			Type: models.EntryTypeAssistant,
			Message: json.RawMessage(`{
				"role": "assistant",
				"content": [
					{"type": "tool_use", "id": "toolu_01", "name": "Read", "input": {}},
					{"type": "tool_use", "id": "toolu_02", "name": "Bash", "input": {}}
				]
			}`),
		},
		{
			Type: models.EntryTypeAssistant,
			Message: json.RawMessage(`{
				"role": "assistant",
				"content": [
					{"type": "tool_use", "id": "toolu_03", "name": "Write", "input": {}}
				]
			}`),
		},
	}

	stats := ComputeSessionStats(entries, nil)

	if stats.ToolCallCount != 3 {
		t.Errorf("ToolCallCount = %d, want 3", stats.ToolCallCount)
	}
}

// TestFormatDuration tests duration formatting.
func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration string
		expected string
	}{
		{"less than minute", "30s", "30s"},
		{"one minute", "1m", "1m"},
		{"minutes only", "45m", "45m"},
		{"one hour", "1h", "1h 0m"},
		{"hours and minutes", "2h35m", "2h 35m"},
		{"long session", "5h23m", "5h 23m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, _ := time.ParseDuration(tt.duration)
			result := formatDuration(d)
			if result != tt.expected {
				t.Errorf("formatDuration(%s) = %q, want %q", tt.duration, result, tt.expected)
			}
		})
	}
}

// TestComputeSessionStats_WithAgents tests agent counting.
func TestComputeSessionStats_WithAgents(t *testing.T) {
	agents := []*agent.TreeNode{
		{AgentID: "agent-001", EntryCount: 10},
		{AgentID: "agent-002", EntryCount: 20},
		{
			AgentID:    "agent-003",
			EntryCount: 5,
			Children: []*agent.TreeNode{
				{AgentID: "agent-004", EntryCount: 3},
			},
		},
	}

	stats := ComputeSessionStats(nil, agents)

	if stats.AgentCount != 4 {
		t.Errorf("AgentCount = %d, want 4", stats.AgentCount)
	}
	if stats.TotalAgentMessages != 38 {
		t.Errorf("TotalAgentMessages = %d, want 38 (10+20+5+3)", stats.TotalAgentMessages)
	}
	if stats.SubagentMessages != 38 {
		t.Errorf("SubagentMessages = %d, want 38", stats.SubagentMessages)
	}
}

// TestTruncateSessionID tests session ID truncation.
func TestTruncateSessionID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"fbd51e2b-1234-5678-90ab-cdef12345678", "fbd51e2b"},
		{"short", "short"},
		{"exactly8", "exactly8"},
		{"", ""},
		{"1234567890", "12345678"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := TruncateSessionID(tt.input)
			if result != tt.expected {
				t.Errorf("TruncateSessionID(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestRenderHTMLHeader_WithStats tests header generation with stats.
func TestRenderHTMLHeader_WithStats(t *testing.T) {
	stats := &SessionStats{
		SessionID:          "fbd51e2b-1234-5678-90ab-cdef12345678",
		ProjectPath:        "/Users/name/project",
		SessionStart:       "2026-02-01 14:23",
		Duration:           "2h 35m",
		MessageCount:       914,
		UserMessages:       42,
		AssistantMessages:  128,
		AgentCount:         11,
		TotalAgentMessages: 488,
		ToolCallCount:      247,
	}

	html := renderHTMLHeader(stats, nil)

	// Check structure
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("Missing DOCTYPE")
	}
	if !strings.Contains(html, "<header class=\"page-header\">") {
		t.Error("Missing header element with page-header class")
	}
	if !strings.Contains(html, "<h1>Claude Code Session") {
		t.Error("Missing h1 title")
	}
	if !strings.Contains(html, "class=\"session-metadata\"") {
		t.Error("Missing session-metadata class")
	}

	// Check metadata items
	if !strings.Contains(html, "Session:") {
		t.Error("Missing Session label")
	}
	if !strings.Contains(html, "fbd51e2b") {
		t.Error("Missing truncated session ID")
	}
	// Note: "Project:" label removed - session folder is now in h1
	if !strings.Contains(html, "Started: 2026-02-01 14:23") {
		t.Error("Missing session start time")
	}
	if !strings.Contains(html, "Duration: 2h 35m") {
		t.Error("Missing duration")
	}
	// Verify export time is NOT displayed
	if strings.Contains(html, "Exported:") {
		t.Error("Export time should not be displayed in header")
	}
	// Check enhanced message statistics
	if !strings.Contains(html, "User: 42") {
		t.Error("Missing user message count")
	}
	if !strings.Contains(html, "Assistant: 128") {
		t.Error("Missing assistant message count")
	}
	if !strings.Contains(html, "Subagents[11]: 488 messages") {
		t.Error("Missing subagent message count")
	}
	if !strings.Contains(html, "Tools: 247 calls") {
		t.Error("Missing tool call count")
	}

	// Check copy button for session ID
	if !strings.Contains(html, "data-copy-text=\"fbd51e2b-1234-5678-90ab-cdef12345678\"") {
		t.Error("Missing copy button with full session ID")
	}
}

// TestRenderHTMLHeader_NilStats tests header generation without stats.
func TestRenderHTMLHeader_NilStats(t *testing.T) {
	html := renderHTMLHeader(nil, nil)

	// Should still have basic structure
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("Missing DOCTYPE")
	}
	if !strings.Contains(html, "<h1>Claude Code Session</h1>") {
		t.Error("Missing h1 title")
	}
	if !strings.Contains(html, "class=\"controls\"") {
		t.Error("Missing controls")
	}
}

// TestRenderHTMLHeader_EmptyStats tests header with empty stats.
func TestRenderHTMLHeader_EmptyStats(t *testing.T) {
	stats := &SessionStats{}
	html := renderHTMLHeader(stats, nil)

	// Should have basic structure
	if !strings.Contains(html, "<header class=\"page-header\">") {
		t.Error("Missing header element")
	}
	// Should have zero counts in enhanced format
	if !strings.Contains(html, "User: 0 | Assistant: 0 | Subagents[0]: 0 messages") {
		t.Error("Missing enhanced zero message counts")
	}
	if !strings.Contains(html, "Tools: 0 calls") {
		t.Error("Missing zero tool count")
	}
}

// TestRenderHTMLFooter_WithStats tests footer generation with stats.
func TestRenderHTMLFooter_WithStats(t *testing.T) {
	stats := &SessionStats{
		ProjectPath: "/Users/name/project",
	}

	html := renderHTMLFooter(stats)

	// Check footer structure
	if !strings.Contains(html, "<footer class=\"page-footer\">") {
		t.Error("Missing footer element with page-footer class")
	}
	if !strings.Contains(html, "class=\"footer-info\"") {
		t.Error("Missing footer-info class")
	}
	if !strings.Contains(html, "class=\"footer-help\"") {
		t.Error("Missing footer-help class")
	}

	// Check info content
	if !strings.Contains(html, "Exported from <strong>claude-history</strong> CLI") {
		t.Error("Missing export attribution")
	}
	if !strings.Contains(html, "Export format version: "+ExportFormatVersion) {
		t.Error("Missing format version")
	}

	// Check source path
	if !strings.Contains(html, "Source:") {
		t.Error("Missing Source label")
	}
	if !strings.Contains(html, "~/.claude/projects/") {
		t.Error("Missing source path prefix")
	}

	// Check keyboard shortcuts
	if !strings.Contains(html, "<details>") {
		t.Error("Missing details element for shortcuts")
	}
	if !strings.Contains(html, "Keyboard Shortcuts") {
		t.Error("Missing keyboard shortcuts summary")
	}
	if !strings.Contains(html, "<kbd>Ctrl</kbd>") {
		t.Error("Missing kbd element")
	}
	if !strings.Contains(html, "Expand/Collapse All") {
		t.Error("Missing Ctrl+K shortcut description")
	}
	if !strings.Contains(html, "Search") {
		t.Error("Missing Ctrl+F shortcut description")
	}
	if !strings.Contains(html, "Clear Search") {
		t.Error("Missing Esc shortcut description")
	}

	// Check scripts are included
	if !strings.Contains(html, "<script src=\"static/script.js\"></script>") {
		t.Error("Missing script.js")
	}
	if !strings.Contains(html, "<script src=\"static/clipboard.js\"></script>") {
		t.Error("Missing clipboard.js")
	}
	if !strings.Contains(html, "<script src=\"static/controls.js\"></script>") {
		t.Error("Missing controls.js")
	}
}

// TestRenderHTMLFooter_NilStats tests footer generation without stats.
func TestRenderHTMLFooter_NilStats(t *testing.T) {
	html := renderHTMLFooter(nil)

	// Should have basic structure
	if !strings.Contains(html, "<footer class=\"page-footer\">") {
		t.Error("Missing footer element")
	}
	if !strings.Contains(html, "Exported from <strong>claude-history</strong> CLI") {
		t.Error("Missing export attribution")
	}
	// Should NOT have source path when stats is nil
	if strings.Contains(html, "Source:") {
		t.Error("Should not have Source when ProjectPath is empty")
	}
}

// TestRenderConversationWithStats_Integration tests full rendering with stats.
func TestRenderConversationWithStats_Integration(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-001",
			SessionID: "test-session-123",
			Type:      models.EntryTypeUser,
			Timestamp: "2026-01-31T10:00:00Z",
			Message:   json.RawMessage(`"Hello!"`),
		},
		{
			UUID:      "uuid-002",
			SessionID: "test-session-123",
			Type:      models.EntryTypeAssistant,
			Timestamp: "2026-01-31T10:00:05Z",
			Message: json.RawMessage(`{
				"role": "assistant",
				"content": [
					{"type": "text", "text": "Hi there!"},
					{"type": "tool_use", "id": "toolu_01", "name": "Read", "input": {"file_path": "/test.go"}}
				]
			}`),
		},
	}

	stats := &SessionStats{
		SessionID:          "test-session-123",
		ProjectPath:        "/test/project",
		SessionStart:       "2026-01-31 10:00",
		Duration:           "5s",
		MessageCount:       2,
		UserMessages:       1,
		AssistantMessages:  1,
		AgentCount:         0,
		TotalAgentMessages: 0,
		ToolCallCount:      1,
	}

	html, err := RenderConversationWithStats(entries, nil, stats)
	if err != nil {
		t.Fatalf("RenderConversationWithStats() error = %v", err)
	}

	// Check header content
	if !strings.Contains(html, "<h1>Claude Code Session") {
		t.Error("Missing header title")
	}
	if !strings.Contains(html, "test-ses") {
		t.Error("Missing truncated session ID in header")
	}
	// Note: Project path moved to h1 as session folder link
	if !strings.Contains(html, "Started: 2026-01-31 10:00") {
		t.Error("Missing session start time in header")
	}
	if !strings.Contains(html, "Duration: 5s") {
		t.Error("Missing duration in header")
	}
	if !strings.Contains(html, "User: 1 | Assistant: 1 | Subagents[0]: 0 messages") {
		t.Error("Missing enhanced message counts in header")
	}

	// Check conversation content
	if !strings.Contains(html, "Hello!") {
		t.Error("Missing user message")
	}
	if !strings.Contains(html, "Hi there!") {
		t.Error("Missing assistant message")
	}
	if !strings.Contains(html, "[Read] /test.go") {
		t.Error("Missing tool call")
	}

	// Check footer content
	if !strings.Contains(html, "<footer class=\"page-footer\">") {
		t.Error("Missing footer")
	}
	if !strings.Contains(html, "Export format version:") {
		t.Error("Missing version in footer")
	}
}

// TestRenderConversationWithStats_AutoComputeStats tests auto-computation of stats.
func TestRenderConversationWithStats_AutoComputeStats(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-001",
			SessionID: "auto-session",
			Type:      models.EntryTypeUser,
			Timestamp: "2026-01-31T10:00:00Z",
			Message:   json.RawMessage(`"Test"`),
		},
		{
			UUID:      "uuid-002",
			SessionID: "auto-session",
			Type:      models.EntryTypeAssistant,
			Timestamp: "2026-01-31T10:00:05Z",
			Message: json.RawMessage(`{
				"role": "assistant",
				"content": [{"type": "text", "text": "Response"}]
			}`),
		},
	}

	// Pass nil for stats - should auto-compute
	html, err := RenderConversationWithStats(entries, nil, nil)
	if err != nil {
		t.Fatalf("RenderConversationWithStats() error = %v", err)
	}

	// Should have auto-computed enhanced message counts
	if !strings.Contains(html, "User: 1 | Assistant: 1 | Subagents[0]: 0 messages") {
		t.Error("Missing auto-computed enhanced message counts")
	}
	// Session ID should be extracted from entries
	if !strings.Contains(html, "auto-ses") {
		t.Error("Missing auto-extracted session ID")
	}
}

// TestExportFormatVersion tests that the version constant is set.
func TestExportFormatVersion(t *testing.T) {
	if ExportFormatVersion == "" {
		t.Error("ExportFormatVersion should not be empty")
	}
	if ExportFormatVersion != "2.0" {
		t.Errorf("ExportFormatVersion = %q, want %q", ExportFormatVersion, "2.0")
	}
}

// TestRenderHTMLHeader_XSSPrevention tests XSS prevention in header.
func TestRenderHTMLHeader_XSSPrevention(t *testing.T) {
	stats := &SessionStats{
		SessionID:   "<script>alert('xss')</script>",
		ProjectPath: "<img onerror='alert(1)'>",
	}

	html := renderHTMLHeader(stats, nil)

	// Script and img tags should be escaped
	if strings.Contains(html, "<script>alert") {
		t.Error("XSS vulnerability: unescaped script tag")
	}
	if strings.Contains(html, "<img onerror") {
		t.Error("XSS vulnerability: unescaped img tag")
	}
	if !strings.Contains(html, "&lt;script&gt;") {
		t.Error("Script tag should be escaped")
	}
}

// TestRenderHTMLFooter_XSSPrevention tests XSS prevention in footer.
func TestRenderHTMLFooter_XSSPrevention(t *testing.T) {
	stats := &SessionStats{
		ProjectPath: "<script>evil()</script>",
	}

	html := renderHTMLFooter(stats)

	if strings.Contains(html, "<script>evil") {
		t.Error("XSS vulnerability: unescaped script tag in footer")
	}
}

// TestCopyButtonIntegration tests copy button generation in header.
func TestCopyButtonIntegration(t *testing.T) {
	stats := &SessionStats{
		SessionID:   "full-session-id-12345",
		ProjectPath: "/path/to/project",
	}

	html := renderHTMLHeader(stats, nil)

	// Check session ID copy button
	if !strings.Contains(html, "class=\"copy-btn\"") {
		t.Error("Missing copy button class")
	}
	if !strings.Contains(html, "data-copy-text=\"full-session-id-12345\"") {
		t.Error("Missing copy text for full session ID")
	}
	if !strings.Contains(html, "data-copy-type=\"session-id\"") {
		t.Error("Missing copy type for session ID")
	}
	if !strings.Contains(html, "title=\"Copy full session ID\"") {
		t.Error("Missing tooltip for session copy button")
	}
}

// TestCopyButtonIntegration_Footer tests copy button in footer.
func TestCopyButtonIntegration_Footer(t *testing.T) {
	stats := &SessionStats{
		ProjectPath: "/my/project/path",
	}

	html := renderHTMLFooter(stats)

	// Check source path copy button
	if !strings.Contains(html, "data-copy-type=\"source-path\"") {
		t.Error("Missing copy type for source path")
	}
	if !strings.Contains(html, "data-copy-text=\"/my/project/path\"") {
		t.Error("Missing copy text for source path")
	}
}

// TestRenderConversation_BackwardCompatibility tests the original function still works.
func TestRenderConversation_BackwardCompatibility(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-001",
			SessionID: "compat-session",
			Type:      models.EntryTypeUser,
			Timestamp: "2026-01-31T10:00:00Z",
			Message:   json.RawMessage(`"Test message"`),
		},
	}

	// Original function should still work
	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	// Should have header and footer
	if !strings.Contains(html, "<header class=\"page-header\">") {
		t.Error("Missing header in backward compatible output")
	}
	if !strings.Contains(html, "<footer class=\"page-footer\">") {
		t.Error("Missing footer in backward compatible output")
	}
	if !strings.Contains(html, "Test message") {
		t.Error("Missing message content")
	}
}

// TestSessionStats_Struct tests the SessionStats struct fields.
func TestSessionStats_Struct(t *testing.T) {
	stats := SessionStats{
		SessionID:     "test-id",
		ProjectPath:   "/test/path",
		ExportTime:    "2026-01-01 00:00:00",
		MessageCount:  100,
		AgentCount:    5,
		ToolCallCount: 50,
	}

	if stats.SessionID != "test-id" {
		t.Errorf("SessionID mismatch")
	}
	if stats.ProjectPath != "/test/path" {
		t.Errorf("ProjectPath mismatch")
	}
	if stats.ExportTime != "2026-01-01 00:00:00" {
		t.Errorf("ExportTime mismatch")
	}
	if stats.MessageCount != 100 {
		t.Errorf("MessageCount mismatch")
	}
	if stats.AgentCount != 5 {
		t.Errorf("AgentCount mismatch")
	}
	if stats.ToolCallCount != 50 {
		t.Errorf("ToolCallCount mismatch")
	}
}

// TestRenderHTMLHeader_Controls tests that controls are included in header.
func TestRenderHTMLHeader_Controls(t *testing.T) {
	html := renderHTMLHeader(nil, nil)

	// Check controls are present
	if !strings.Contains(html, "id=\"expand-all-btn\"") {
		t.Error("Missing expand all button")
	}
	if !strings.Contains(html, "id=\"collapse-all-btn\"") {
		t.Error("Missing collapse all button")
	}
	if !strings.Contains(html, "id=\"search-box\"") {
		t.Error("Missing search box")
	}
	if !strings.Contains(html, "id=\"search-prev-btn\"") {
		t.Error("Missing search prev button")
	}
	if !strings.Contains(html, "id=\"search-next-btn\"") {
		t.Error("Missing search next button")
	}
	if !strings.Contains(html, "class=\"search-results\"") {
		t.Error("Missing search results span")
	}
}

// TestRenderHTMLHeader_Accessibility tests accessibility attributes.
func TestRenderHTMLHeader_Accessibility(t *testing.T) {
	html := renderHTMLHeader(nil, nil)

	if !strings.Contains(html, "role=\"toolbar\"") {
		t.Error("Missing toolbar role")
	}
	if !strings.Contains(html, "aria-label=\"Conversation controls\"") {
		t.Error("Missing aria-label for controls")
	}
	if !strings.Contains(html, "aria-label=\"Search messages\"") {
		t.Error("Missing aria-label for search")
	}
	if !strings.Contains(html, "aria-live=\"polite\"") {
		t.Error("Missing aria-live for search results")
	}
}

// TestHtmlHeaderConstant_BackwardCompatibility tests the deprecated constant still exists.
func TestHtmlHeaderConstant_BackwardCompatibility(t *testing.T) {
	if !strings.Contains(htmlHeader, "<!DOCTYPE html>") {
		t.Error("htmlHeader constant should contain DOCTYPE")
	}
	if !strings.Contains(htmlHeader, "<header class=\"page-header\">") {
		t.Error("htmlHeader constant should have header element")
	}
}

// TestHtmlFooterConstant_BackwardCompatibility tests the deprecated constant still exists.
func TestHtmlFooterConstant_BackwardCompatibility(t *testing.T) {
	if !strings.Contains(htmlFooter, "<footer class=\"page-footer\">") {
		t.Error("htmlFooter constant should have footer element")
	}
	if !strings.Contains(htmlFooter, "</html>") {
		t.Error("htmlFooter constant should close html tag")
	}
}
