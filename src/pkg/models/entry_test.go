// Package models defines data structures for Claude Code history entries.
package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestToolUseResultParsing(t *testing.T) {
	// Real Claude Code agent spawn entry format
	jsonData := `{
		"type": "user",
		"uuid": "7bd059eb-60fb-49b4-92ea-5b6be2a6cfce",
		"sessionId": "926ef72c-163e-4022-bc68-49fcca61ba80",
		"parentUuid": "4e08ee78-a494-47ce-a82c-cf565114a15e",
		"sourceToolAssistantUUID": "4e08ee78-a494-47ce-a82c-cf565114a15e",
		"timestamp": "2026-01-15T10:30:00.000Z",
		"message": {
			"role": "user",
			"content": [{"type": "tool_result", "tool_use_id": "toolu_01Won", "content": []}]
		},
		"toolUseResult": {
			"isAsync": true,
			"status": "async_launched",
			"agentId": "a6f6578",
			"description": "Review PR #217 workflow changes",
			"prompt": "Review GitHub PR #217...",
			"outputFile": "/tmp/claude/tasks/a6f6578.output"
		}
	}`

	var entry ConversationEntry
	err := json.Unmarshal([]byte(jsonData), &entry)
	if err != nil {
		t.Fatalf("Failed to parse entry: %v", err)
	}

	// Verify basic fields
	if entry.Type != EntryTypeUser {
		t.Errorf("Expected type 'user', got '%s'", entry.Type)
	}
	if entry.UUID != "7bd059eb-60fb-49b4-92ea-5b6be2a6cfce" {
		t.Errorf("UUID mismatch: got '%s'", entry.UUID)
	}
	if entry.SessionID != "926ef72c-163e-4022-bc68-49fcca61ba80" {
		t.Errorf("SessionID mismatch: got '%s'", entry.SessionID)
	}
	if entry.SourceToolAssistantUUID != "4e08ee78-a494-47ce-a82c-cf565114a15e" {
		t.Errorf("SourceToolAssistantUUID mismatch: got '%s'", entry.SourceToolAssistantUUID)
	}

	// Verify ToolUseResult parsing
	if entry.ToolUseResult == nil {
		t.Fatal("ToolUseResult should not be nil")
	}
	if !entry.ToolUseResult.IsAsync {
		t.Error("Expected IsAsync to be true")
	}
	if entry.ToolUseResult.Status != "async_launched" {
		t.Errorf("Expected status 'async_launched', got '%s'", entry.ToolUseResult.Status)
	}
	if entry.ToolUseResult.AgentID != "a6f6578" {
		t.Errorf("Expected agentId 'a6f6578', got '%s'", entry.ToolUseResult.AgentID)
	}
	if entry.ToolUseResult.Description != "Review PR #217 workflow changes" {
		t.Errorf("Description mismatch: got '%s'", entry.ToolUseResult.Description)
	}
	if entry.ToolUseResult.Prompt != "Review GitHub PR #217..." {
		t.Errorf("Prompt mismatch: got '%s'", entry.ToolUseResult.Prompt)
	}
	if entry.ToolUseResult.OutputFile != "/tmp/claude/tasks/a6f6578.output" {
		t.Errorf("OutputFile mismatch: got '%s'", entry.ToolUseResult.OutputFile)
	}
}

func TestIsAgentSpawn(t *testing.T) {
	tests := []struct {
		name     string
		entry    ConversationEntry
		expected bool
	}{
		{
			name: "valid agent spawn",
			entry: ConversationEntry{
				Type: EntryTypeUser,
				ToolUseResult: &ToolUseResult{
					IsAsync: true,
					Status:  "async_launched",
					AgentID: "a6f6578",
				},
			},
			expected: true,
		},
		{
			name: "completed status - not a spawn",
			entry: ConversationEntry{
				Type: EntryTypeUser,
				ToolUseResult: &ToolUseResult{
					IsAsync: false,
					Status:  "completed",
					AgentID: "a6f6578",
				},
			},
			expected: false,
		},
		{
			name: "async_launched but empty agentId",
			entry: ConversationEntry{
				Type: EntryTypeUser,
				ToolUseResult: &ToolUseResult{
					IsAsync: true,
					Status:  "async_launched",
					AgentID: "",
				},
			},
			expected: false,
		},
		{
			name: "no toolUseResult",
			entry: ConversationEntry{
				Type:          EntryTypeUser,
				ToolUseResult: nil,
			},
			expected: false,
		},
		{
			name: "assistant entry with toolUseResult - should still work",
			entry: ConversationEntry{
				Type: EntryTypeAssistant,
				ToolUseResult: &ToolUseResult{
					IsAsync: true,
					Status:  "async_launched",
					AgentID: "b7g7689",
				},
			},
			expected: true,
		},
		{
			name:     "empty entry",
			entry:    ConversationEntry{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.entry.IsAgentSpawn()
			if result != tt.expected {
				t.Errorf("IsAgentSpawn() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestGetSpawnedAgentID(t *testing.T) {
	tests := []struct {
		name     string
		entry    ConversationEntry
		expected string
	}{
		{
			name: "valid agent spawn returns agentId",
			entry: ConversationEntry{
				Type: EntryTypeUser,
				ToolUseResult: &ToolUseResult{
					IsAsync: true,
					Status:  "async_launched",
					AgentID: "a6f6578",
				},
			},
			expected: "a6f6578",
		},
		{
			name: "completed status returns empty",
			entry: ConversationEntry{
				Type: EntryTypeUser,
				ToolUseResult: &ToolUseResult{
					Status:  "completed",
					AgentID: "a6f6578",
				},
			},
			expected: "",
		},
		{
			name: "no toolUseResult returns empty",
			entry: ConversationEntry{
				Type: EntryTypeUser,
			},
			expected: "",
		},
		{
			name:     "empty entry returns empty",
			entry:    ConversationEntry{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.entry.GetSpawnedAgentID()
			if result != tt.expected {
				t.Errorf("GetSpawnedAgentID() = '%s', expected '%s'", result, tt.expected)
			}
		})
	}
}

func TestHasToolUseResult(t *testing.T) {
	tests := []struct {
		name     string
		entry    ConversationEntry
		expected bool
	}{
		{
			name: "with toolUseResult",
			entry: ConversationEntry{
				ToolUseResult: &ToolUseResult{
					Status:  "async_launched",
					AgentID: "test123",
				},
			},
			expected: true,
		},
		{
			name: "without toolUseResult",
			entry: ConversationEntry{
				Type: EntryTypeUser,
			},
			expected: false,
		},
		{
			name:     "empty entry",
			entry:    ConversationEntry{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.entry.HasToolUseResult()
			if result != tt.expected {
				t.Errorf("HasToolUseResult() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestGetToolUseResult(t *testing.T) {
	// Test with toolUseResult present
	tur := &ToolUseResult{
		IsAsync:     true,
		Status:      "async_launched",
		AgentID:     "test123",
		Description: "Test task",
	}
	entry := ConversationEntry{
		ToolUseResult: tur,
	}
	result := entry.GetToolUseResult()
	if result != tur {
		t.Errorf("GetToolUseResult() did not return the expected pointer")
	}
	if result.AgentID != "test123" {
		t.Errorf("AgentID mismatch: got '%s'", result.AgentID)
	}

	// Test without toolUseResult
	emptyEntry := ConversationEntry{}
	if emptyEntry.GetToolUseResult() != nil {
		t.Error("GetToolUseResult() should return nil for empty entry")
	}
}

func TestBackwardCompatibility(t *testing.T) {
	// Test that entries without the new fields still parse correctly
	oldFormatJSON := `{
		"type": "user",
		"uuid": "old-uuid-123",
		"sessionId": "old-session-456",
		"timestamp": "2026-01-15T10:30:00.000Z",
		"message": "Hello, Claude!"
	}`

	var entry ConversationEntry
	err := json.Unmarshal([]byte(oldFormatJSON), &entry)
	if err != nil {
		t.Fatalf("Failed to parse old format entry: %v", err)
	}

	// Verify basic parsing still works
	if entry.Type != EntryTypeUser {
		t.Errorf("Expected type 'user', got '%s'", entry.Type)
	}
	if entry.UUID != "old-uuid-123" {
		t.Errorf("UUID mismatch: got '%s'", entry.UUID)
	}

	// Verify new fields default to zero values
	if entry.SourceToolAssistantUUID != "" {
		t.Errorf("SourceToolAssistantUUID should be empty, got '%s'", entry.SourceToolAssistantUUID)
	}
	if entry.ToolUseResult != nil {
		t.Error("ToolUseResult should be nil for old format entries")
	}

	// Verify helper methods work correctly on old format entries
	if entry.HasToolUseResult() {
		t.Error("HasToolUseResult() should return false for old format entries")
	}
	if entry.IsAgentSpawn() {
		t.Error("IsAgentSpawn() should return false for old format entries")
	}
	if entry.GetSpawnedAgentID() != "" {
		t.Error("GetSpawnedAgentID() should return empty string for old format entries")
	}
}

func TestEntryTypeHelpers(t *testing.T) {
	tests := []struct {
		entryType        EntryType
		isUser           bool
		isAssistant      bool
		isSystem         bool
		isQueueOperation bool
	}{
		{EntryTypeUser, true, false, false, false},
		{EntryTypeAssistant, false, true, false, false},
		{EntryTypeSystem, false, false, true, false},
		{EntryTypeQueueOperation, false, false, false, true},
		{EntryTypeSummary, false, false, false, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.entryType), func(t *testing.T) {
			entry := ConversationEntry{Type: tt.entryType}

			if entry.IsUser() != tt.isUser {
				t.Errorf("IsUser() = %v, expected %v", entry.IsUser(), tt.isUser)
			}
			if entry.IsAssistant() != tt.isAssistant {
				t.Errorf("IsAssistant() = %v, expected %v", entry.IsAssistant(), tt.isAssistant)
			}
			if entry.IsSystem() != tt.isSystem {
				t.Errorf("IsSystem() = %v, expected %v", entry.IsSystem(), tt.isSystem)
			}
			if entry.IsQueueOperation() != tt.isQueueOperation {
				t.Errorf("IsQueueOperation() = %v, expected %v", entry.IsQueueOperation(), tt.isQueueOperation)
			}
		})
	}
}

func TestGetTimestamp(t *testing.T) {
	entry := ConversationEntry{
		Timestamp: "2026-01-15T10:30:00.123456789Z",
	}

	ts, err := entry.GetTimestamp()
	if err != nil {
		t.Fatalf("GetTimestamp() failed: %v", err)
	}

	expected := time.Date(2026, 1, 15, 10, 30, 0, 123456789, time.UTC)
	if !ts.Equal(expected) {
		t.Errorf("Timestamp mismatch: got %v, expected %v", ts, expected)
	}
}

func TestGetTimestampInvalid(t *testing.T) {
	entry := ConversationEntry{
		Timestamp: "invalid-timestamp",
	}

	_, err := entry.GetTimestamp()
	if err == nil {
		t.Error("GetTimestamp() should fail for invalid timestamp")
	}
}

func TestParseMessageContent(t *testing.T) {
	tests := []struct {
		name        string
		messageJSON string
		expectText  string
	}{
		{
			name:        "plain string message",
			messageJSON: `"Hello, Claude!"`,
			expectText:  "Hello, Claude!",
		},
		{
			name:        "wrapped string content",
			messageJSON: `{"role": "user", "content": "Hello from wrapper"}`,
			expectText:  "Hello from wrapper",
		},
		{
			name:        "array content with text",
			messageJSON: `{"role": "assistant", "content": [{"type": "text", "text": "Response text"}]}`,
			expectText:  "Response text",
		},
		{
			name:        "empty message",
			messageJSON: ``,
			expectText:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := ConversationEntry{
				Message: json.RawMessage(tt.messageJSON),
			}
			text := entry.GetTextContent()
			if text != tt.expectText {
				t.Errorf("GetTextContent() = '%s', expected '%s'", text, tt.expectText)
			}
		})
	}
}

func TestToolUseResultJSONSerialization(t *testing.T) {
	entry := ConversationEntry{
		UUID:                    "test-uuid",
		SessionID:               "test-session",
		Type:                    EntryTypeUser,
		SourceToolAssistantUUID: "parent-uuid",
		ToolUseResult: &ToolUseResult{
			IsAsync:     true,
			Status:      "async_launched",
			AgentID:     "agent123",
			Description: "Test agent task",
			Prompt:      "Do something",
			OutputFile:  "/tmp/output.txt",
		},
	}

	// Serialize to JSON
	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("Failed to marshal entry: %v", err)
	}

	// Deserialize back
	var parsed ConversationEntry
	err = json.Unmarshal(data, &parsed)
	if err != nil {
		t.Fatalf("Failed to unmarshal entry: %v", err)
	}

	// Verify round-trip
	if parsed.SourceToolAssistantUUID != entry.SourceToolAssistantUUID {
		t.Errorf("SourceToolAssistantUUID mismatch after round-trip")
	}
	if !parsed.HasToolUseResult() {
		t.Fatal("ToolUseResult should be present after round-trip")
	}
	if parsed.ToolUseResult.AgentID != entry.ToolUseResult.AgentID {
		t.Errorf("AgentID mismatch after round-trip")
	}
	if !parsed.IsAgentSpawn() {
		t.Error("IsAgentSpawn() should return true after round-trip")
	}
}

func TestMultipleAgentSpawnDetection(t *testing.T) {
	// Test a sequence of entries to ensure we can identify spawns correctly
	entries := []ConversationEntry{
		{
			Type: EntryTypeUser,
			UUID: "entry1",
			ToolUseResult: &ToolUseResult{
				Status:  "async_launched",
				AgentID: "agent1",
			},
		},
		{
			Type: EntryTypeAssistant,
			UUID: "entry2",
		},
		{
			Type: EntryTypeUser,
			UUID: "entry3",
			ToolUseResult: &ToolUseResult{
				Status:  "completed",
				AgentID: "agent1",
			},
		},
		{
			Type: EntryTypeUser,
			UUID: "entry4",
			ToolUseResult: &ToolUseResult{
				Status:  "async_launched",
				AgentID: "agent2",
			},
		},
	}

	spawnCount := 0
	var spawnedIDs []string

	for _, e := range entries {
		if e.IsAgentSpawn() {
			spawnCount++
			spawnedIDs = append(spawnedIDs, e.GetSpawnedAgentID())
		}
	}

	if spawnCount != 2 {
		t.Errorf("Expected 2 agent spawns, found %d", spawnCount)
	}

	expectedIDs := []string{"agent1", "agent2"}
	for i, id := range spawnedIDs {
		if id != expectedIDs[i] {
			t.Errorf("SpawnedID[%d] = '%s', expected '%s'", i, id, expectedIDs[i])
		}
	}
}
