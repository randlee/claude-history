package export

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/randlee/claude-history/pkg/models"
)

// TestUserContentWithBashOutput tests the full rendering pipeline for user messages with bash output
func TestUserContentWithBashOutput(t *testing.T) {
	// Simulate a real user message with bash output
	userMessage := `<bash-stdout>beads
beads-state-machine
claude-code-viewer</bash-stdout><bash-stderr></bash-stderr>`

	entries := []models.ConversationEntry{
		{
			UUID:      "test-001",
			SessionID: "session-001",
			Type:      models.EntryTypeUser,
			Timestamp: "2026-02-07T10:00:00Z",
			Message:   json.RawMessage(`"` + userMessage + `"`),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	// Verify bash-stdout is rendered with proper formatting
	if !strings.Contains(html, `xml-tag-block`) {
		t.Error("HTML should contain xml-tag-block div")
	}
	if !strings.Contains(html, `&lt;bash-stdout&gt;`) {
		t.Error("HTML should contain escaped bash-stdout tag")
	}
	if !strings.Contains(html, `xml-tag-content`) {
		t.Error("HTML should contain xml-tag-content div")
	}
	if !strings.Contains(html, "beads") {
		t.Error("HTML should contain bash output content")
	}

	// Verify bash-stderr is NOT rendered (empty tag)
	if strings.Contains(html, "bash-stderr") {
		t.Error("HTML should not contain empty bash-stderr tag")
	}

	// Verify user-content class is applied
	if !strings.Contains(html, `class="text user-content"`) {
		t.Error("HTML should have user-content class on text div")
	}
}

// TestUserContentPlainText tests that plain user text without XML tags works correctly
func TestUserContentPlainText(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "test-002",
			SessionID: "session-001",
			Type:      models.EntryTypeUser,
			Timestamp: "2026-02-07T10:00:00Z",
			Message:   json.RawMessage(`"Hello, this is plain text"`),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	// Verify plain text is escaped and rendered
	if !strings.Contains(html, "Hello, this is plain text") {
		t.Error("HTML should contain plain text content")
	}

	// Verify no XML formatting is applied
	if strings.Contains(html, "xml-tag-block") {
		t.Error("HTML should not contain xml-tag-block for plain text")
	}
}

// TestUserContentMixedContent tests user messages with text before and after XML tags
func TestUserContentMixedContent(t *testing.T) {
	userMessage := `Running command:
<bash-stdout>output here</bash-stdout>
Command completed.`

	entries := []models.ConversationEntry{
		{
			UUID:      "test-003",
			SessionID: "session-001",
			Type:      models.EntryTypeUser,
			Timestamp: "2026-02-07T10:00:00Z",
			Message:   json.RawMessage(`"` + userMessage + `"`),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	// Verify all parts are present
	if !strings.Contains(html, "Running command:") {
		t.Error("HTML should contain text before XML tag")
	}
	if !strings.Contains(html, "xml-tag-block") {
		t.Error("HTML should contain xml-tag-block")
	}
	if !strings.Contains(html, "output here") {
		t.Error("HTML should contain XML tag content")
	}
	if !strings.Contains(html, "Command completed.") {
		t.Error("HTML should contain text after XML tag")
	}
}
