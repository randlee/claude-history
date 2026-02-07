package export

import (
	"strings"
	"testing"

	"github.com/randlee/claude-history/pkg/models"
)

func TestTaskNotificationDetection(t *testing.T) {
	tests := []struct {
		name     string
		entry    models.ConversationEntry
		expected bool
	}{
		{
			name: "user entry with task-notification",
			entry: models.ConversationEntry{
				Type: models.EntryTypeUser,
				Message: []byte(`{"role":"user","content":"<task-notification><task-id>abc123</task-id><status>completed</status><summary>Test task</summary><result>Success</result></task-notification>"}`),
			},
			expected: true,
		},
		{
			name: "user entry without task-notification",
			entry: models.ConversationEntry{
				Type:    models.EntryTypeUser,
				Message: []byte(`{"role":"user","content":"Regular user message"}`),
			},
			expected: false,
		},
		{
			name: "assistant entry with task-notification (should not match)",
			entry: models.ConversationEntry{
				Type:    models.EntryTypeAssistant,
				Message: []byte(`{"role":"assistant","content":"<task-notification>test</task-notification>"}`),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			textContent := tt.entry.GetTextContent()
			isTaskNotif := tt.entry.Type == models.EntryTypeUser && strings.Contains(textContent, "<task-notification>")

			if isTaskNotif != tt.expected {
				t.Errorf("expected isTaskNotif=%v, got %v", tt.expected, isTaskNotif)
			}
		})
	}
}

func TestParseTaskNotification(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected *TaskNotificationData
	}{
		{
			name: "complete task notification",
			content: `<task-notification>
<task-id>abc123</task-id>
<status>completed</status>
<summary>Agent "Test Agent" completed</summary>
<result>Perfect! The task is done.</result>
</task-notification>`,
			expected: &TaskNotificationData{
				TaskID:  "abc123",
				Status:  "completed",
				Summary: `Agent "Test Agent" completed`,
				Result:  "Perfect! The task is done.",
			},
		},
		{
			name: "task notification without result",
			content: `<task-notification>
<task-id>xyz789</task-id>
<status>running</status>
<summary>Agent task in progress</summary>
</task-notification>`,
			expected: &TaskNotificationData{
				TaskID:  "xyz789",
				Status:  "running",
				Summary: "Agent task in progress",
				Result:  "",
			},
		},
		{
			name:     "not a task notification",
			content:  "Regular text content",
			expected: nil,
		},
		{
			name: "multiline result",
			content: `<task-notification>
<task-id>multi123</task-id>
<status>completed</status>
<summary>Complex task</summary>
<result>Line 1
Line 2
Line 3</result>
</task-notification>`,
			expected: &TaskNotificationData{
				TaskID:  "multi123",
				Status:  "completed",
				Summary: "Complex task",
				Result:  "Line 1\nLine 2\nLine 3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseTaskNotification(tt.content)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected nil result, got %+v", result)
				}
				return
			}

			if result == nil {
				t.Fatalf("expected non-nil result, got nil")
			}

			if result.TaskID != tt.expected.TaskID {
				t.Errorf("TaskID: expected %q, got %q", tt.expected.TaskID, result.TaskID)
			}
			if result.Status != tt.expected.Status {
				t.Errorf("Status: expected %q, got %q", tt.expected.Status, result.Status)
			}
			if result.Summary != tt.expected.Summary {
				t.Errorf("Summary: expected %q, got %q", tt.expected.Summary, result.Summary)
			}
			if result.Result != tt.expected.Result {
				t.Errorf("Result: expected %q, got %q", tt.expected.Result, result.Result)
			}
		})
	}
}

func TestRenderTaskNotification(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		expectContains []string
	}{
		{
			name: "completed task",
			content: `<task-notification>
<task-id>abc123</task-id>
<status>completed</status>
<summary>Agent "Test Agent" completed</summary>
<result>Success!</result>
</task-notification>`,
			expectContains: []string{
				`class="task-notification"`,
				`status-completed`,
				"✓",
				`Agent &#34;Test Agent&#34; completed`, // HTML-escaped quotes
				"abc123",
				"Success!",
			},
		},
		{
			name: "failed task",
			content: `<task-notification>
<task-id>xyz789</task-id>
<status>failed</status>
<summary>Agent task failed</summary>
<result>Error occurred</result>
</task-notification>`,
			expectContains: []string{
				`class="task-notification"`,
				`status-failed`,
				"✗",
				"Agent task failed",
			},
		},
		{
			name: "running task",
			content: `<task-notification>
<task-id>run123</task-id>
<status>running</status>
<summary>Agent task in progress</summary>
</task-notification>`,
			expectContains: []string{
				`class="task-notification"`,
				`status-running`,
				"⏳",
				"Agent task in progress",
			},
		},
		{
			name: "long result (should be collapsible)",
			content: `<task-notification>
<task-id>long123</task-id>
<status>completed</status>
<summary>Long task</summary>
<result>` + strings.Repeat("a", 400) + `</result>
</task-notification>`,
			expectContains: []string{
				`class="task-notification"`,
				"<details",
				"View result",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html := renderTaskNotification(tt.content)

			for _, expected := range tt.expectContains {
				if !strings.Contains(html, expected) {
					t.Errorf("expected HTML to contain %q, but it doesn't.\nGot: %s", expected, html)
				}
			}
		})
	}
}

func TestRenderEntry_WithTaskNotification(t *testing.T) {
	entry := models.ConversationEntry{
		UUID:      "test-uuid",
		Type:      models.EntryTypeUser,
		Timestamp: "2026-02-07T12:00:00Z",
		Message:   []byte(`{"role":"user","content":"<task-notification><task-id>abc123</task-id><status>completed</status><summary>Agent completed</summary><result>Done!</result></task-notification>"}`),
	}

	html := renderEntry(entry, make(map[string]models.ToolResult))

	// Should render as system entry, not user
	if !strings.Contains(html, `class="message-row system"`) {
		t.Error("expected task-notification to render as system entry")
	}

	// Should have Agent Notification label
	if !strings.Contains(html, "Agent Notification") {
		t.Error("expected 'Agent Notification' role label")
	}

	// Should contain task-notification styling
	if !strings.Contains(html, `class="task-notification"`) {
		t.Error("expected task-notification div")
	}

	// Should have completion icon
	if !strings.Contains(html, "✓") {
		t.Error("expected completion icon")
	}
}
