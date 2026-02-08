package export

import (
	"encoding/json"
	"testing"

	"github.com/randlee/claude-history/pkg/models"
)

// TestHasContentEdgeCases tests the hasContent function with various edge cases
func TestHasContentEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		setupEntry  func() models.ConversationEntry
		wantContent bool
		reason      string
	}{
		{
			name: "User message with only tool results should be filtered",
			setupEntry: func() models.ConversationEntry {
				message := map[string]any{
					"role": "user",
					"content": []map[string]any{
						{
							"type":        "tool_result",
							"tool_use_id": "test-id",
							"content":     "some output",
						},
					},
				}
				msgJSON, _ := json.Marshal(message)
				return models.ConversationEntry{
					Type:    models.EntryTypeUser,
					Message: msgJSON,
				}
			},
			wantContent: false,
			reason:      "Tool results in user messages are not rendered in HTML",
		},
		{
			name: "User message with text and tool results should be rendered",
			setupEntry: func() models.ConversationEntry {
				message := map[string]any{
					"role": "user",
					"content": []map[string]any{
						{
							"type": "text",
							"text": "Here is the result:",
						},
						{
							"type":        "tool_result",
							"tool_use_id": "test-id",
							"content":     "some output",
						},
					},
				}
				msgJSON, _ := json.Marshal(message)
				return models.ConversationEntry{
					Type:    models.EntryTypeUser,
					Message: msgJSON,
				}
			},
			wantContent: true,
			reason:      "Has text content",
		},
		{
			name: "Assistant message with only whitespace text should be filtered",
			setupEntry: func() models.ConversationEntry {
				message := map[string]any{
					"role": "assistant",
					"content": []map[string]any{
						{
							"type": "text",
							"text": "   \n\n\r\n\t   ",
						},
					},
				}
				msgJSON, _ := json.Marshal(message)
				return models.ConversationEntry{
					Type:    models.EntryTypeAssistant,
					Message: msgJSON,
				}
			},
			wantContent: false,
			reason:      "Only whitespace, no real content",
		},
		{
			name: "Assistant message with tool calls should be rendered",
			setupEntry: func() models.ConversationEntry {
				message := map[string]any{
					"role": "assistant",
					"content": []map[string]any{
						{
							"type":  "tool_use",
							"id":    "test-id",
							"name":  "Bash",
							"input": map[string]any{"command": "ls"},
						},
					},
				}
				msgJSON, _ := json.Marshal(message)
				return models.ConversationEntry{
					Type:    models.EntryTypeAssistant,
					Message: msgJSON,
				}
			},
			wantContent: true,
			reason:      "Has tool calls",
		},
		{
			name: "Empty message should be filtered",
			setupEntry: func() models.ConversationEntry {
				return models.ConversationEntry{
					Type:    models.EntryTypeUser,
					Message: json.RawMessage(`{"role":"user","content":""}`),
				}
			},
			wantContent: false,
			reason:      "Empty message content",
		},
		{
			name: "Message with newlines only should be filtered",
			setupEntry: func() models.ConversationEntry {
				return models.ConversationEntry{
					Type:    models.EntryTypeAssistant,
					Message: json.RawMessage(`{"role":"assistant","content":"\n\n\n"}`),
				}
			},
			wantContent: false,
			reason:      "Only newlines, no real content",
		},
		{
			name: "Message with single space should be filtered",
			setupEntry: func() models.ConversationEntry {
				return models.ConversationEntry{
					Type:    models.EntryTypeUser,
					Message: json.RawMessage(`{"role":"user","content":" "}`),
				}
			},
			wantContent: false,
			reason:      "Only whitespace",
		},
		{
			name: "Message with actual text (even short) should be rendered",
			setupEntry: func() models.ConversationEntry {
				return models.ConversationEntry{
					Type:    models.EntryTypeAssistant,
					Message: json.RawMessage(`{"role":"assistant","content":"ok"}`),
				}
			},
			wantContent: true,
			reason:      "Has actual text content",
		},
		{
			name: "System message with no text should be filtered",
			setupEntry: func() models.ConversationEntry {
				return models.ConversationEntry{
					Type:    models.EntryTypeSystem,
					Message: json.RawMessage(`{"content":""}`),
				}
			},
			wantContent: false,
			reason:      "System message with no content",
		},
		{
			name: "Queue operation with no content should be filtered",
			setupEntry: func() models.ConversationEntry {
				return models.ConversationEntry{
					Type:    models.EntryTypeQueueOperation,
					Message: json.RawMessage(`{}`),
				}
			},
			wantContent: false,
			reason:      "Queue operation with no content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := tt.setupEntry()
			got := hasContent(entry)

			if got != tt.wantContent {
				t.Errorf("hasContent() = %v, want %v", got, tt.wantContent)
				t.Logf("Reason: %s", tt.reason)
				t.Logf("Entry type: %s", entry.Type)
				t.Logf("Text content: %q", entry.GetTextContent())
				if entry.Type == models.EntryTypeAssistant {
					t.Logf("Tool calls: %d", len(entry.ExtractToolCalls()))
				}
				if entry.Type == models.EntryTypeUser {
					t.Logf("Tool results: %d", len(entry.ExtractToolResults()))
				}
			}
		})
	}
}

// TestHasContentDocumentation documents the expected behavior of hasContent()
func TestHasContentDocumentation(t *testing.T) {
	t.Log("hasContent() should return FALSE for:")
	t.Log("  - User messages with ONLY tool results (no text)")
	t.Log("  - Any message with ONLY whitespace (spaces, tabs, newlines)")
	t.Log("  - Empty messages")
	t.Log("  - System/Queue messages with no content")
	t.Log("")
	t.Log("hasContent() should return TRUE for:")
	t.Log("  - User messages with actual text (even if they also have tool results)")
	t.Log("  - Assistant messages with actual text")
	t.Log("  - Assistant messages with tool calls (even if no text)")
	t.Log("  - Any message with non-whitespace text content")
}
