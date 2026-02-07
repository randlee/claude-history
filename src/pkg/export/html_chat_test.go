package export

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/randlee/claude-history/pkg/models"
)

// TestChatBubble_UserMessageLayout verifies user messages use left-aligned chat bubble layout.
func TestChatBubble_UserMessageLayout(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-user-001",
			SessionID: "session-001",
			Type:      models.EntryTypeUser,
			Timestamp: "2026-01-31T14:30:00Z",
			Message:   json.RawMessage(`"Hello, Claude!"`),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	// Check for message-row with user class
	if !strings.Contains(html, `class="message-row user"`) {
		t.Error("User message missing message-row user class")
	}

	// Check for avatar placeholder
	if !strings.Contains(html, `class="avatar user"`) {
		t.Error("User message missing avatar placeholder")
	}

	// Check for message-bubble
	if !strings.Contains(html, `class="message-bubble"`) {
		t.Error("User message missing message-bubble class")
	}

	// Check for message-header
	if !strings.Contains(html, `class="message-header"`) {
		t.Error("User message missing message-header class")
	}

	// Check for message-content
	if !strings.Contains(html, `class="message-content"`) {
		t.Error("User message missing message-content class")
	}

	// Check for role label
	if !strings.Contains(html, `<span class="role">User</span>`) {
		t.Error("User message missing role label")
	}

	// Check content is present
	if !strings.Contains(html, "Hello, Claude!") {
		t.Error("User message content not rendered")
	}
}

// TestChatBubble_AssistantMessageLayout verifies assistant messages use right-aligned chat bubble layout.
func TestChatBubble_AssistantMessageLayout(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-assistant-001",
			SessionID: "session-001",
			Type:      models.EntryTypeAssistant,
			Timestamp: "2026-01-31T14:31:00Z",
			Message:   json.RawMessage(`{"role": "assistant", "content": [{"type": "text", "text": "Hello! How can I help you?"}]}`),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	// Check for message-row with assistant class
	if !strings.Contains(html, `class="message-row assistant"`) {
		t.Error("Assistant message missing message-row assistant class")
	}

	// Check for avatar placeholder
	if !strings.Contains(html, `class="avatar assistant"`) {
		t.Error("Assistant message missing avatar placeholder")
	}

	// Check for role label
	if !strings.Contains(html, `<span class="role">Assistant</span>`) {
		t.Error("Assistant message missing role label")
	}

	// Check content is present
	if !strings.Contains(html, "Hello! How can I help you?") {
		t.Error("Assistant message content not rendered")
	}
}

// TestChatBubble_SystemMessageLayout verifies system messages use appropriate layout.
func TestChatBubble_SystemMessageLayout(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-system-001",
			SessionID: "session-001",
			Type:      models.EntryTypeSystem,
			Timestamp: "2026-01-31T14:00:00Z",
			Message:   json.RawMessage(`"System initialized successfully"`),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	// Check for message-row with system class
	if !strings.Contains(html, `class="message-row system"`) {
		t.Error("System message missing message-row system class")
	}

	// Check for avatar placeholder
	if !strings.Contains(html, `class="avatar system"`) {
		t.Error("System message missing avatar placeholder")
	}

	// Check for role label
	if !strings.Contains(html, `<span class="role">System</span>`) {
		t.Error("System message missing role label")
	}
}

// TestChatBubble_QueueOperationLayout verifies queue operation messages use appropriate layout.
func TestChatBubble_QueueOperationLayout(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-queue-001",
			SessionID: "session-001",
			Type:      models.EntryTypeQueueOperation,
			AgentID:   "agent-abc123",
			Timestamp: "2026-01-31T14:32:00Z",
			Message:   json.RawMessage(`"Spawning agent for task"`),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	// Check for message-row with queue-operation class
	if !strings.Contains(html, `class="message-row queue-operation"`) {
		t.Error("Queue operation message missing message-row queue-operation class")
	}

	// Check for avatar placeholder
	if !strings.Contains(html, `class="avatar queue-operation"`) {
		t.Error("Queue operation message missing avatar placeholder")
	}

	// Check for role label
	if !strings.Contains(html, `<span class="role">Agent Task</span>`) {
		t.Error("Queue operation message missing role label")
	}
}

// TestChatBubble_SummaryMessageLayout verifies summary messages use appropriate layout.
func TestChatBubble_SummaryMessageLayout(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-summary-001",
			SessionID: "session-001",
			Type:      models.EntryTypeSummary,
			Timestamp: "2026-01-31T15:00:00Z",
			Message:   json.RawMessage(`"Conversation summary: Discussed project requirements."`),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	// Check for message-row with summary class
	if !strings.Contains(html, `class="message-row summary"`) {
		t.Error("Summary message missing message-row summary class")
	}

	// Check for avatar placeholder
	if !strings.Contains(html, `class="avatar summary"`) {
		t.Error("Summary message missing avatar placeholder")
	}

	// Check for role label
	if !strings.Contains(html, `<span class="role">Summary</span>`) {
		t.Error("Summary message missing role label")
	}
}

// TestChatBubble_TimestampFormatting verifies timestamps are formatted as readable times.
func TestChatBubble_TimestampFormatting(t *testing.T) {
	tests := []struct {
		name      string
		timestamp string
		expected  string
	}{
		{
			name:      "morning time",
			timestamp: "2026-01-31T09:30:00Z",
			expected:  "9:30 AM",
		},
		{
			name:      "afternoon time",
			timestamp: "2026-01-31T14:30:00Z",
			expected:  "2:30 PM",
		},
		{
			name:      "midnight",
			timestamp: "2026-01-31T00:00:00Z",
			expected:  "12:00 AM",
		},
		{
			name:      "noon",
			timestamp: "2026-01-31T12:00:00Z",
			expected:  "12:00 PM",
		},
		{
			name:      "with nanoseconds",
			timestamp: "2026-01-31T15:45:30.123456789Z",
			expected:  "3:45 PM",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTimestampReadable(tt.timestamp)
			if result != tt.expected {
				t.Errorf("formatTimestampReadable(%q) = %q, want %q", tt.timestamp, result, tt.expected)
			}
		})
	}
}

// TestChatBubble_TimestampInvalid verifies invalid timestamps are handled gracefully.
func TestChatBubble_TimestampInvalid(t *testing.T) {
	result := formatTimestampReadable("not-a-valid-timestamp")
	if result != "not-a-valid-timestamp" {
		t.Errorf("Invalid timestamp should return original string, got %q", result)
	}
}

// TestChatBubble_TimestampEmpty verifies empty timestamps are handled gracefully.
func TestChatBubble_TimestampEmpty(t *testing.T) {
	result := formatTimestampReadable("")
	if result != "" {
		t.Errorf("Empty timestamp should return empty string, got %q", result)
	}
}

// TestChatBubble_TimestampInHTML verifies timestamp appears in rendered HTML.
func TestChatBubble_TimestampInHTML(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-001",
			SessionID: "session-001",
			Type:      models.EntryTypeUser,
			Timestamp: "2026-01-31T14:30:45Z",
			Message:   json.RawMessage(`"Test message"`),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	if !strings.Contains(html, `class="timestamp"`) {
		t.Error("HTML missing timestamp class")
	}
	if !strings.Contains(html, "2:30 PM") {
		t.Error("HTML missing formatted timestamp")
	}
}

// TestChatBubble_AgentIDInHeader verifies agent ID appears in message header.
func TestChatBubble_AgentIDInHeader(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-001",
			SessionID: "session-001",
			AgentID:   "agent-xyz789",
			Type:      models.EntryTypeAssistant,
			Timestamp: "2026-01-31T14:30:00Z",
			Message:   json.RawMessage(`{"role": "assistant", "content": [{"type": "text", "text": "Agent response"}]}`),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	// Agent ID should be in header
	if !strings.Contains(html, "[agent-xyz789]") {
		t.Error("HTML missing agent ID in header")
	}

	// Copy button should be present
	if !strings.Contains(html, `data-copy-text="agent-xyz789"`) {
		t.Error("HTML missing copy button for agent ID")
	}
}

// TestChatBubble_EmptyContent verifies empty content is skipped.
func TestChatBubble_EmptyContent(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-001",
			SessionID: "session-001",
			Type:      models.EntryTypeUser,
			Timestamp: "2026-01-31T14:30:00Z",
			Message:   json.RawMessage(`""`),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	// Empty content should be skipped (no message-row rendered)
	if strings.Contains(html, `class="message-row user"`) {
		t.Error("Empty content should not render message-row")
	}
	if strings.Contains(html, `class="message-bubble"`) {
		t.Error("Empty content should not render message-bubble")
	}
}

// TestChatBubble_NullContent verifies null content is skipped.
func TestChatBubble_NullContent(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-001",
			SessionID: "session-001",
			Type:      models.EntryTypeUser,
			Timestamp: "2026-01-31T14:30:00Z",
			Message:   nil,
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	// Null content should be skipped (no message-row rendered)
	if strings.Contains(html, `class="message-row user"`) {
		t.Error("Null content should not render message-row")
	}
}

// TestChatBubble_VeryLongContent verifies very long content is handled correctly.
func TestChatBubble_VeryLongContent(t *testing.T) {
	longText := strings.Repeat("This is a very long message. ", 1000)
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-001",
			SessionID: "session-001",
			Type:      models.EntryTypeUser,
			Timestamp: "2026-01-31T14:30:00Z",
			Message:   json.RawMessage(`"` + longText + `"`),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	// Content should be present
	if !strings.Contains(html, "This is a very long message.") {
		t.Error("Long content should be rendered")
	}

	// Should have message-content wrapper for proper word-break handling
	if !strings.Contains(html, `class="message-content"`) {
		t.Error("Long content should have message-content wrapper")
	}
}

// TestChatBubble_MultilineContent verifies multiline content is preserved.
func TestChatBubble_MultilineContent(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-001",
			SessionID: "session-001",
			Type:      models.EntryTypeUser,
			Timestamp: "2026-01-31T14:30:00Z",
			Message:   json.RawMessage(`"Line 1\nLine 2\nLine 3"`),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	// Newlines should be preserved (CSS will handle display with white-space: pre-wrap)
	if !strings.Contains(html, "Line 1\nLine 2\nLine 3") {
		t.Error("Multiline content should preserve newlines")
	}
}

// TestChatBubble_SpecialCharacters verifies special characters are properly escaped.
func TestChatBubble_SpecialCharacters(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-001",
			SessionID: "session-001",
			Type:      models.EntryTypeUser,
			Timestamp: "2026-01-31T14:30:00Z",
			Message:   json.RawMessage(`"Test <script>alert('xss')</script> & \"quotes\""`),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	// Script tags should be escaped
	if strings.Contains(html, "<script>") {
		t.Error("Script tags should be HTML escaped")
	}
	if !strings.Contains(html, "&lt;script&gt;") {
		t.Error("Script tags should be escaped to &lt;script&gt;")
	}
}

// TestGetRoleLabel verifies all entry types have appropriate role labels.
func TestGetRoleLabel(t *testing.T) {
	tests := []struct {
		entryType models.EntryType
		expected  string
	}{
		{models.EntryTypeUser, "User"},
		{models.EntryTypeAssistant, "Assistant"},
		{models.EntryTypeSystem, "System"},
		{models.EntryTypeQueueOperation, "Agent Task"},
		{models.EntryTypeSummary, "Summary"},
		{models.EntryType("unknown-type"), "unknown-type"},
	}

	for _, tt := range tests {
		t.Run(string(tt.entryType), func(t *testing.T) {
			result := getRoleLabel(tt.entryType, "User", "Assistant")
			if result != tt.expected {
				t.Errorf("getRoleLabel(%q) = %q, want %q", tt.entryType, result, tt.expected)
			}
		})
	}
}

// TestChatBubble_AllMessageTypesHaveAvatar verifies all message types include avatar placeholders.
func TestChatBubble_AllMessageTypesHaveAvatar(t *testing.T) {
	entryTypes := []models.EntryType{
		models.EntryTypeUser,
		models.EntryTypeAssistant,
		models.EntryTypeSystem,
		models.EntryTypeQueueOperation,
		models.EntryTypeSummary,
	}

	for _, entryType := range entryTypes {
		t.Run(string(entryType), func(t *testing.T) {
			entries := []models.ConversationEntry{
				{
					UUID:      "uuid-test",
					SessionID: "session-001",
					Type:      entryType,
					Timestamp: "2026-01-31T14:30:00Z",
					Message:   json.RawMessage(`"test content"`),
				},
			}

			html, err := RenderConversation(entries, nil)
			if err != nil {
				t.Fatalf("RenderConversation() error = %v", err)
			}

			expectedClass := getEntryClass(entryType)
			avatarClass := `class="avatar ` + expectedClass + `"`
			if !strings.Contains(html, avatarClass) {
				t.Errorf("Entry type %s should have avatar with class %q", entryType, avatarClass)
			}
		})
	}
}

// TestChatBubble_MessageBubbleStructure verifies the complete message bubble structure.
func TestChatBubble_MessageBubbleStructure(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-001",
			SessionID: "session-001",
			Type:      models.EntryTypeUser,
			Timestamp: "2026-01-31T14:30:00Z",
			Message:   json.RawMessage(`"Test content"`),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	// Verify structure ordering (message-row -> avatar + message-bubble)
	messageRowIdx := strings.Index(html, `class="message-row user"`)
	avatarIdx := strings.Index(html, `class="avatar user"`)
	bubbleIdx := strings.Index(html, `class="message-bubble"`)
	headerIdx := strings.Index(html, `class="message-header"`)
	contentIdx := strings.Index(html, `class="message-content"`)

	if messageRowIdx == -1 || avatarIdx == -1 || bubbleIdx == -1 || headerIdx == -1 || contentIdx == -1 {
		t.Fatal("Missing required structural elements")
	}

	// Avatar and bubble should be inside message-row
	if avatarIdx < messageRowIdx {
		t.Error("Avatar should be inside message-row")
	}
	if bubbleIdx < messageRowIdx {
		t.Error("Bubble should be inside message-row")
	}

	// Header and content should be inside bubble
	if headerIdx < bubbleIdx {
		t.Error("Header should be inside message-bubble")
	}
	if contentIdx < bubbleIdx {
		t.Error("Content should be inside message-bubble")
	}

	// Header should come before content
	if headerIdx > contentIdx {
		t.Error("Header should come before content")
	}
}

// TestChatBubble_ConversationFlow verifies a full conversation renders correctly.
func TestChatBubble_ConversationFlow(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-001",
			SessionID: "session-001",
			Type:      models.EntryTypeUser,
			Timestamp: "2026-01-31T14:30:00Z",
			Message:   json.RawMessage(`"What is Go?"`),
		},
		{
			UUID:      "uuid-002",
			SessionID: "session-001",
			Type:      models.EntryTypeAssistant,
			Timestamp: "2026-01-31T14:30:30Z",
			Message:   json.RawMessage(`{"role": "assistant", "content": [{"type": "text", "text": "Go is a programming language designed at Google."}]}`),
		},
		{
			UUID:      "uuid-003",
			SessionID: "session-001",
			Type:      models.EntryTypeUser,
			Timestamp: "2026-01-31T14:31:00Z",
			Message:   json.RawMessage(`"Can you show me an example?"`),
		},
		{
			UUID:      "uuid-004",
			SessionID: "session-001",
			Type:      models.EntryTypeAssistant,
			Timestamp: "2026-01-31T14:31:30Z",
			Message:   json.RawMessage(`{"role": "assistant", "content": [{"type": "text", "text": "Here is a simple Hello World example in Go."}]}`),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	// Check all messages are present
	if !strings.Contains(html, "What is Go?") {
		t.Error("First user message missing")
	}
	if !strings.Contains(html, "Go is a programming language") {
		t.Error("First assistant message missing")
	}
	if !strings.Contains(html, "Can you show me an example?") {
		t.Error("Second user message missing")
	}
	if !strings.Contains(html, "Hello World example") {
		t.Error("Second assistant message missing")
	}

	// Check message counts
	userCount := strings.Count(html, `class="message-row user"`)
	assistantCount := strings.Count(html, `class="message-row assistant"`)

	if userCount != 2 {
		t.Errorf("Expected 2 user messages, got %d", userCount)
	}
	if assistantCount != 2 {
		t.Errorf("Expected 2 assistant messages, got %d", assistantCount)
	}

	// Check timestamps are formatted
	if !strings.Contains(html, "2:30 PM") {
		t.Error("Timestamps should be formatted as readable times")
	}
}

// TestChatBubble_ToolCallsInsideBubble verifies tool calls appear inside the message bubble.
func TestChatBubble_ToolCallsInsideBubble(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-001",
			SessionID: "session-001",
			Type:      models.EntryTypeAssistant,
			Timestamp: "2026-01-31T14:30:00Z",
			Message: json.RawMessage(`{
				"role": "assistant",
				"content": [
					{"type": "text", "text": "Let me check the file."},
					{"type": "tool_use", "id": "toolu_01", "name": "Read", "input": {"file_path": "/path/to/file.go"}}
				]
			}`),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	// Tool call should be present (with collapsible collapsed classes)
	if !strings.Contains(html, `class="tool-call collapsible collapsed"`) {
		t.Error("Tool call should be rendered with collapsible collapsed classes")
	}

	// Verify tool call is inside message-content
	contentIdx := strings.Index(html, `class="message-content"`)
	toolIdx := strings.Index(html, `class="tool-call collapsible collapsed"`)

	if toolIdx < contentIdx {
		t.Error("Tool call should be inside message-content")
	}
}

// TestChatBubble_DataUUIDAttribute verifies data-uuid attribute is present.
func TestChatBubble_DataUUIDAttribute(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "test-uuid-12345",
			SessionID: "session-001",
			Type:      models.EntryTypeUser,
			Timestamp: "2026-01-31T14:30:00Z",
			Message:   json.RawMessage(`"test"`),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	if !strings.Contains(html, `data-uuid="test-uuid-12345"`) {
		t.Error("Message row should have data-uuid attribute")
	}
}

// TestChatBubble_AriaHiddenOnAvatar verifies avatar has aria-hidden for accessibility.
func TestChatBubble_AriaHiddenOnAvatar(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-001",
			SessionID: "session-001",
			Type:      models.EntryTypeUser,
			Timestamp: "2026-01-31T14:30:00Z",
			Message:   json.RawMessage(`"test"`),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	if !strings.Contains(html, `aria-hidden="true"`) {
		t.Error("Avatar should have aria-hidden attribute for accessibility")
	}
}

// TestRenderEntry_DirectCall tests the renderEntry function directly.
func TestRenderEntry_DirectCall(t *testing.T) {
	entry := models.ConversationEntry{
		UUID:      "uuid-direct",
		SessionID: "session-001",
		Type:      models.EntryTypeUser,
		Timestamp: "2026-01-31T14:30:00Z",
		Message:   json.RawMessage(`"Direct test"`),
	}

	html := renderEntry(entry, nil, "", "User", "Assistant")

	// Should produce valid HTML structure
	if !strings.Contains(html, `class="message-row user"`) {
		t.Error("renderEntry should produce message-row")
	}
	if !strings.Contains(html, "Direct test") {
		t.Error("renderEntry should include content")
	}
}

// TestChatBubble_MarkdownInAssistantMessage verifies markdown is rendered for assistant messages.
func TestChatBubble_MarkdownInAssistantMessage(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-001",
			SessionID: "session-001",
			Type:      models.EntryTypeAssistant,
			Timestamp: "2026-01-31T14:30:00Z",
			Message:   json.RawMessage(`{"role": "assistant", "content": [{"type": "text", "text": "Here is some **bold** text."}]}`),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	// Should have markdown-content class
	if !strings.Contains(html, `class="text markdown-content"`) {
		t.Error("Assistant message should have markdown-content class")
	}
}

// TestChatBubble_PlainTextInUserMessage verifies user messages don't get markdown rendering.
func TestChatBubble_PlainTextInUserMessage(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-001",
			SessionID: "session-001",
			Type:      models.EntryTypeUser,
			Timestamp: "2026-01-31T14:30:00Z",
			Message:   json.RawMessage(`"Some **text** here"`),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	// User messages should NOT have markdown-content class on the text div
	// They should have class="text" without markdown-content
	if strings.Contains(html, `class="text markdown-content"`) {
		// This would only be a problem if it's in a user message context
		// Let's check if the user message has plain text class
		userIdx := strings.Index(html, `class="message-row user"`)
		markdownIdx := strings.Index(html, `class="text markdown-content"`)

		// Find the closing of the user message
		closingIdx := strings.Index(html[userIdx:], "</div>\n</div>\n")
		if markdownIdx > userIdx && markdownIdx < userIdx+closingIdx {
			t.Error("User message should not have markdown-content rendering")
		}
	}
}
