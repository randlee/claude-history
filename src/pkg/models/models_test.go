package models

import (
	"encoding/json"
	"testing"
)

func TestConversationEntry_GetTimestamp(t *testing.T) {
	entry := ConversationEntry{
		Timestamp: "2026-02-01T18:57:51.729Z",
	}

	ts, err := entry.GetTimestamp()
	if err != nil {
		t.Fatalf("GetTimestamp() error: %v", err)
	}

	if ts.Year() != 2026 || ts.Month() != 2 || ts.Day() != 1 {
		t.Errorf("GetTimestamp() = %v, unexpected date", ts)
	}
}

func TestConversationEntry_TypeChecks(t *testing.T) {
	tests := []struct {
		entryType   EntryType
		isUser      bool
		isAssistant bool
		isSystem    bool
		isQueueOp   bool
	}{
		{EntryTypeUser, true, false, false, false},
		{EntryTypeAssistant, false, true, false, false},
		{EntryTypeSystem, false, false, true, false},
		{EntryTypeQueueOperation, false, false, false, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.entryType), func(t *testing.T) {
			entry := ConversationEntry{Type: tt.entryType}

			if entry.IsUser() != tt.isUser {
				t.Errorf("IsUser() = %v, want %v", entry.IsUser(), tt.isUser)
			}
			if entry.IsAssistant() != tt.isAssistant {
				t.Errorf("IsAssistant() = %v, want %v", entry.IsAssistant(), tt.isAssistant)
			}
			if entry.IsSystem() != tt.isSystem {
				t.Errorf("IsSystem() = %v, want %v", entry.IsSystem(), tt.isSystem)
			}
			if entry.IsQueueOperation() != tt.isQueueOp {
				t.Errorf("IsQueueOperation() = %v, want %v", entry.IsQueueOperation(), tt.isQueueOp)
			}
		})
	}
}

func TestConversationEntry_ParseMessageContent(t *testing.T) {
	tests := []struct {
		name      string
		message   string
		wantCount int
		wantType  string
		wantText  string
	}{
		{
			name:      "array of content",
			message:   `[{"type": "text", "text": "Hello"}]`,
			wantCount: 1,
			wantType:  "text",
			wantText:  "Hello",
		},
		{
			name:      "single object",
			message:   `{"type": "text", "text": "World"}`,
			wantCount: 1,
			wantType:  "text",
			wantText:  "World",
		},
		{
			name:      "plain string",
			message:   `"Just text"`,
			wantCount: 1,
			wantType:  "text",
			wantText:  "Just text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := ConversationEntry{
				Message: json.RawMessage(tt.message),
			}

			contents, err := entry.ParseMessageContent()
			if err != nil {
				t.Fatalf("ParseMessageContent() error: %v", err)
			}

			if len(contents) != tt.wantCount {
				t.Errorf("ParseMessageContent() returned %d items, want %d", len(contents), tt.wantCount)
			}

			if len(contents) > 0 {
				if contents[0].Type != tt.wantType {
					t.Errorf("content type = %q, want %q", contents[0].Type, tt.wantType)
				}
				if contents[0].Text != tt.wantText {
					t.Errorf("content text = %q, want %q", contents[0].Text, tt.wantText)
				}
			}
		})
	}
}

func TestConversationEntry_GetTextContent(t *testing.T) {
	entry := ConversationEntry{
		Message: json.RawMessage(`[{"type": "text", "text": "Line 1"}, {"type": "tool_use", "name": "test"}, {"type": "text", "text": "Line 2"}]`),
	}

	text := entry.GetTextContent()
	expected := "Line 1\nLine 2"

	if text != expected {
		t.Errorf("GetTextContent() = %q, want %q", text, expected)
	}
}

func TestSessionIndexEntry_ToSession(t *testing.T) {
	entry := SessionIndexEntry{
		SessionID:    "679761ba-80c0-4cd3-a586-cc6a1fc56308",
		FullPath:     "/Users/test/.claude/projects/-test/session.jsonl",
		ProjectPath:  "/Users/test/project",
		FirstPrompt:  "Hello",
		MessageCount: 10,
		Created:      "2026-02-01T18:00:00.000Z",
		Modified:     "2026-02-01T19:00:00.000Z",
		GitBranch:    "main",
	}

	session := entry.ToSession()

	if session.ID != entry.SessionID {
		t.Errorf("ID = %q, want %q", session.ID, entry.SessionID)
	}
	if session.ProjectPath != entry.ProjectPath {
		t.Errorf("ProjectPath = %q, want %q", session.ProjectPath, entry.ProjectPath)
	}
	if session.MessageCount != entry.MessageCount {
		t.Errorf("MessageCount = %d, want %d", session.MessageCount, entry.MessageCount)
	}
	if session.Created.Year() != 2026 {
		t.Errorf("Created year = %d, want 2026", session.Created.Year())
	}
}
