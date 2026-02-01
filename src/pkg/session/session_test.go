package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/randlee/claude-history/pkg/models"
)

// mustWriteFile writes a file or fails the test
func mustWriteFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("WriteFile(%q) failed: %v", path, err)
	}
}

func TestReadSession(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.jsonl")

	content := `{"uuid":"1","sessionId":"test-session","type":"user","timestamp":"2026-02-01T18:00:00.000Z","message":"Hello"}
{"uuid":"2","sessionId":"test-session","type":"assistant","timestamp":"2026-02-01T18:00:01.000Z","message":"Hi there"}
`
	mustWriteFile(t, testFile, []byte(content))

	entries, err := ReadSession(testFile)
	if err != nil {
		t.Fatalf("ReadSession() error: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("ReadSession() returned %d entries, want 2", len(entries))
	}

	if entries[0].Type != models.EntryTypeUser {
		t.Errorf("First entry type = %v, want user", entries[0].Type)
	}
}

func TestGetSessionInfo(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "679761ba-80c0-4cd3-a586-cc6a1fc56308.jsonl")

	content := `{"uuid":"1","sessionId":"679761ba-80c0-4cd3-a586-cc6a1fc56308","type":"user","timestamp":"2026-02-01T18:00:00.000Z","message":"What is the weather?"}
{"uuid":"2","sessionId":"679761ba-80c0-4cd3-a586-cc6a1fc56308","type":"assistant","timestamp":"2026-02-01T18:00:05.000Z"}
{"uuid":"3","sessionId":"679761ba-80c0-4cd3-a586-cc6a1fc56308","type":"user","timestamp":"2026-02-01T18:01:00.000Z"}
`
	mustWriteFile(t, testFile, []byte(content))

	session, err := GetSessionInfo(testFile)
	if err != nil {
		t.Fatalf("GetSessionInfo() error: %v", err)
	}

	if session.ID != "679761ba-80c0-4cd3-a586-cc6a1fc56308" {
		t.Errorf("Session ID = %q, want expected UUID", session.ID)
	}

	if session.MessageCount != 3 {
		t.Errorf("MessageCount = %d, want 3", session.MessageCount)
	}

	if session.FirstPrompt != "What is the weather?" {
		t.Errorf("FirstPrompt = %q, want 'What is the weather?'", session.FirstPrompt)
	}
}

func TestFilterEntries(t *testing.T) {
	entries := []models.ConversationEntry{
		{UUID: "1", Type: models.EntryTypeUser, Timestamp: "2026-02-01T10:00:00.000Z"},
		{UUID: "2", Type: models.EntryTypeAssistant, Timestamp: "2026-02-01T11:00:00.000Z"},
		{UUID: "3", Type: models.EntryTypeUser, Timestamp: "2026-02-01T12:00:00.000Z"},
		{UUID: "4", Type: models.EntryTypeSystem, Timestamp: "2026-02-01T13:00:00.000Z"},
	}

	t.Run("filter by type", func(t *testing.T) {
		result := FilterEntries(entries, FilterOptions{
			Types: []models.EntryType{models.EntryTypeUser},
		})
		if len(result) != 2 {
			t.Errorf("Got %d entries, want 2", len(result))
		}
	})

	t.Run("filter by time range", func(t *testing.T) {
		start := time.Date(2026, 2, 1, 10, 30, 0, 0, time.UTC)
		end := time.Date(2026, 2, 1, 12, 30, 0, 0, time.UTC)
		result := FilterEntries(entries, FilterOptions{
			StartTime: &start,
			EndTime:   &end,
		})
		if len(result) != 2 {
			t.Errorf("Got %d entries, want 2", len(result))
		}
	})
}

func TestCountEntriesByType(t *testing.T) {
	entries := []models.ConversationEntry{
		{Type: models.EntryTypeUser},
		{Type: models.EntryTypeAssistant},
		{Type: models.EntryTypeUser},
		{Type: models.EntryTypeAssistant},
		{Type: models.EntryTypeAssistant},
		{Type: models.EntryTypeSystem},
	}

	counts := CountEntriesByType(entries)

	if counts[models.EntryTypeUser] != 2 {
		t.Errorf("User count = %d, want 2", counts[models.EntryTypeUser])
	}
	if counts[models.EntryTypeAssistant] != 3 {
		t.Errorf("Assistant count = %d, want 3", counts[models.EntryTypeAssistant])
	}
	if counts[models.EntryTypeSystem] != 1 {
		t.Errorf("System count = %d, want 1", counts[models.EntryTypeSystem])
	}
}

func TestReadSessionIndex(t *testing.T) {
	tmpDir := t.TempDir()
	indexFile := filepath.Join(tmpDir, "sessions-index.json")

	content := `{
  "version": 1,
  "entries": [
    {
      "sessionId": "679761ba-80c0-4cd3-a586-cc6a1fc56308",
      "fullPath": "/test/path/session.jsonl",
      "projectPath": "/Users/test/project",
      "messageCount": 10,
      "created": "2026-02-01T18:00:00.000Z",
      "modified": "2026-02-01T19:00:00.000Z"
    }
  ]
}`
	mustWriteFile(t, indexFile, []byte(content))

	index, err := ReadSessionIndex(indexFile)
	if err != nil {
		t.Fatalf("ReadSessionIndex() error: %v", err)
	}

	if index.Version != 1 {
		t.Errorf("Version = %d, want 1", index.Version)
	}

	if len(index.Entries) != 1 {
		t.Errorf("Entries count = %d, want 1", len(index.Entries))
	}

	if index.Entries[0].SessionID != "679761ba-80c0-4cd3-a586-cc6a1fc56308" {
		t.Error("SessionID mismatch")
	}
}

// Helper to create an assistant entry with tool calls
func makeAssistantWithTools(uuid string, tools ...struct{ name, input string }) models.ConversationEntry {
	var content []map[string]any
	for i, tool := range tools {
		var inputMap map[string]any
		_ = json.Unmarshal([]byte(tool.input), &inputMap)
		content = append(content, map[string]any{
			"type":  "tool_use",
			"id":    "toolu_" + string(rune('0'+i)),
			"name":  tool.name,
			"input": inputMap,
		})
	}

	wrapper := map[string]any{
		"role":    "assistant",
		"content": content,
	}
	msgBytes, _ := json.Marshal(wrapper)

	return models.ConversationEntry{
		UUID:      uuid,
		Type:      models.EntryTypeAssistant,
		Timestamp: "2026-02-01T10:00:00.000Z",
		Message:   json.RawMessage(msgBytes),
	}
}

func TestFilterEntries_ToolTypes(t *testing.T) {
	bashEntry := makeAssistantWithTools("1", struct{ name, input string }{"Bash", `{"command":"git status"}`})
	readEntry := makeAssistantWithTools("2", struct{ name, input string }{"Read", `{"file_path":"/path/to/file.go"}`})
	multiEntry := makeAssistantWithTools("3",
		struct{ name, input string }{"Bash", `{"command":"npm install"}`},
		struct{ name, input string }{"Write", `{"file_path":"/tmp/test.txt"}`},
	)
	userEntry := models.ConversationEntry{UUID: "4", Type: models.EntryTypeUser, Timestamp: "2026-02-01T10:00:00.000Z"}
	systemEntry := models.ConversationEntry{UUID: "5", Type: models.EntryTypeSystem, Timestamp: "2026-02-01T10:00:00.000Z"}
	assistantNoTools := models.ConversationEntry{
		UUID:      "6",
		Type:      models.EntryTypeAssistant,
		Timestamp: "2026-02-01T10:00:00.000Z",
		Message:   json.RawMessage(`{"role":"assistant","content":"Just text, no tools"}`),
	}

	entries := []models.ConversationEntry{bashEntry, readEntry, multiEntry, userEntry, systemEntry, assistantNoTools}

	tests := []struct {
		name      string
		toolTypes []string
		wantCount int
		wantUUIDs []string
	}{
		{
			name:      "single tool type",
			toolTypes: []string{"Bash"},
			wantCount: 2,
			wantUUIDs: []string{"1", "3"},
		},
		{
			name:      "case insensitive matching",
			toolTypes: []string{"bash"},
			wantCount: 2,
			wantUUIDs: []string{"1", "3"},
		},
		{
			name:      "mixed case matching",
			toolTypes: []string{"BaSh"},
			wantCount: 2,
			wantUUIDs: []string{"1", "3"},
		},
		{
			name:      "multiple tool types OR logic",
			toolTypes: []string{"Read", "Write"},
			wantCount: 2,
			wantUUIDs: []string{"2", "3"},
		},
		{
			name:      "multiple tool types with case insensitive",
			toolTypes: []string{"read", "WRITE"},
			wantCount: 2,
			wantUUIDs: []string{"2", "3"},
		},
		{
			name:      "non-existent tool",
			toolTypes: []string{"NonExistent"},
			wantCount: 0,
			wantUUIDs: []string{},
		},
		{
			name:      "empty tool types does not filter",
			toolTypes: []string{},
			wantCount: 6,
			wantUUIDs: []string{"1", "2", "3", "4", "5", "6"},
		},
		{
			name:      "nil tool types does not filter",
			toolTypes: nil,
			wantCount: 6,
			wantUUIDs: []string{"1", "2", "3", "4", "5", "6"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterEntries(entries, FilterOptions{
				ToolTypes: tt.toolTypes,
			})
			if len(result) != tt.wantCount {
				t.Errorf("Got %d entries, want %d", len(result), tt.wantCount)
			}
			// Verify correct entries returned
			if len(result) == len(tt.wantUUIDs) {
				for i, uuid := range tt.wantUUIDs {
					if result[i].UUID != uuid {
						t.Errorf("Entry %d: got UUID %s, want %s", i, result[i].UUID, uuid)
					}
				}
			}
		})
	}
}

func TestFilterEntries_ToolTypes_WithOtherFilters(t *testing.T) {
	// Test ToolTypes combined with existing filters
	bashEntry1 := makeAssistantWithTools("1", struct{ name, input string }{"Bash", `{"command":"git status"}`})
	bashEntry1.Timestamp = "2026-02-01T10:00:00.000Z"
	bashEntry1.Type = models.EntryTypeAssistant

	bashEntry2 := makeAssistantWithTools("2", struct{ name, input string }{"Bash", `{"command":"npm install"}`})
	bashEntry2.Timestamp = "2026-02-01T11:00:00.000Z"
	bashEntry2.Type = models.EntryTypeAssistant

	readEntry := makeAssistantWithTools("3", struct{ name, input string }{"Read", `{"file_path":"/path/to/file.go"}`})
	readEntry.Timestamp = "2026-02-01T12:00:00.000Z"
	readEntry.Type = models.EntryTypeAssistant

	userEntry := models.ConversationEntry{UUID: "4", Type: models.EntryTypeUser, Timestamp: "2026-02-01T13:00:00.000Z"}

	entries := []models.ConversationEntry{bashEntry1, bashEntry2, readEntry, userEntry}

	t.Run("tool types with entry type filter", func(t *testing.T) {
		result := FilterEntries(entries, FilterOptions{
			ToolTypes: []string{"Bash"},
			Types:     []models.EntryType{models.EntryTypeAssistant},
		})
		if len(result) != 2 {
			t.Errorf("Got %d entries, want 2", len(result))
		}
	})

	t.Run("tool types with time range filter", func(t *testing.T) {
		start := time.Date(2026, 2, 1, 10, 30, 0, 0, time.UTC)
		end := time.Date(2026, 2, 1, 12, 30, 0, 0, time.UTC)
		result := FilterEntries(entries, FilterOptions{
			ToolTypes: []string{"Bash"},
			StartTime: &start,
			EndTime:   &end,
		})
		if len(result) != 1 {
			t.Errorf("Got %d entries, want 1 (bashEntry2 only)", len(result))
		}
		if len(result) > 0 && result[0].UUID != "2" {
			t.Errorf("Got UUID %s, want 2", result[0].UUID)
		}
	})

	t.Run("tool types with all filters combined", func(t *testing.T) {
		start := time.Date(2026, 2, 1, 9, 0, 0, 0, time.UTC)
		end := time.Date(2026, 2, 1, 11, 30, 0, 0, time.UTC)
		result := FilterEntries(entries, FilterOptions{
			ToolTypes: []string{"Bash"},
			Types:     []models.EntryType{models.EntryTypeAssistant},
			StartTime: &start,
			EndTime:   &end,
		})
		if len(result) != 2 {
			t.Errorf("Got %d entries, want 2", len(result))
		}
	})
}

func TestFilterEntries_ToolMatch(t *testing.T) {
	gitEntry := makeAssistantWithTools("1", struct{ name, input string }{"Bash", `{"command":"git status"}`})
	npmEntry := makeAssistantWithTools("2", struct{ name, input string }{"Bash", `{"command":"npm install"}`})
	goFileEntry := makeAssistantWithTools("3", struct{ name, input string }{"Read", `{"file_path":"/Users/test/main.go"}`})
	pyFileEntry := makeAssistantWithTools("4", struct{ name, input string }{"Read", `{"file_path":"/home/user/script.py"}`})
	writeEntry := makeAssistantWithTools("5", struct{ name, input string }{"Write", `{"file_path":"/tmp/output.txt","content":"test"}`})
	multiToolEntry := makeAssistantWithTools("6",
		struct{ name, input string }{"Bash", `{"command":"ls -la"}`},
		struct{ name, input string }{"Grep", `{"pattern":"TODO"}`},
	)
	userEntry := models.ConversationEntry{UUID: "7", Type: models.EntryTypeUser, Timestamp: "2026-02-01T10:00:00.000Z"}

	entries := []models.ConversationEntry{gitEntry, npmEntry, goFileEntry, pyFileEntry, writeEntry, multiToolEntry, userEntry}

	tests := []struct {
		name      string
		pattern   string
		wantCount int
		wantUUIDs []string
	}{
		{
			name:      "simple substring match",
			pattern:   "git",
			wantCount: 1,
			wantUUIDs: []string{"1"},
		},
		{
			name:      "regex pattern for commands",
			pattern:   `git.*status`,
			wantCount: 1,
			wantUUIDs: []string{"1"},
		},
		{
			name:      "regex pattern for npm",
			pattern:   `npm\s+install`,
			wantCount: 1,
			wantUUIDs: []string{"2"},
		},
		{
			name:      "match .go file paths",
			pattern:   `\.go`,
			wantCount: 1,
			wantUUIDs: []string{"3"},
		},
		{
			name:      "match .py file paths",
			pattern:   `\.py`,
			wantCount: 1,
			wantUUIDs: []string{"4"},
		},
		{
			name:      "match any file extension",
			pattern:   `\.\w+$`,
			wantCount: 0,
			wantUUIDs: []string{},
		},
		{
			name:      "match file_path key",
			pattern:   `file_path`,
			wantCount: 3,
			wantUUIDs: []string{"3", "4", "5"},
		},
		{
			name:      "match pattern in multi-tool entry",
			pattern:   `TODO`,
			wantCount: 1,
			wantUUIDs: []string{"6"},
		},
		{
			name:      "no match returns empty",
			pattern:   "nonexistent",
			wantCount: 0,
			wantUUIDs: []string{},
		},
		{
			name:      "invalid regex returns empty",
			pattern:   "[invalid",
			wantCount: 0,
			wantUUIDs: []string{},
		},
		{
			name:      "unclosed group regex returns empty",
			pattern:   "(unclosed",
			wantCount: 0,
			wantUUIDs: []string{},
		},
		{
			name:      "empty pattern does not filter",
			pattern:   "",
			wantCount: 7,
			wantUUIDs: []string{"1", "2", "3", "4", "5", "6", "7"},
		},
		{
			name:      "match /tmp paths",
			pattern:   `/tmp/`,
			wantCount: 1,
			wantUUIDs: []string{"5"},
		},
		{
			name:      "match Users or home in paths",
			pattern:   `(Users|home)`,
			wantCount: 2,
			wantUUIDs: []string{"3", "4"},
		},
		{
			name:      "case sensitive pattern",
			pattern:   `Bash`,
			wantCount: 0,
			wantUUIDs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterEntries(entries, FilterOptions{
				ToolMatch: tt.pattern,
			})
			if len(result) != tt.wantCount {
				t.Errorf("Got %d entries, want %d", len(result), tt.wantCount)
				for i, r := range result {
					t.Logf("  Result[%d]: UUID=%s", i, r.UUID)
				}
			}
			// Verify correct entries returned when count matches
			if len(result) == len(tt.wantUUIDs) {
				for i, uuid := range tt.wantUUIDs {
					if result[i].UUID != uuid {
						t.Errorf("Entry %d: got UUID %s, want %s", i, result[i].UUID, uuid)
					}
				}
			}
		})
	}
}

func TestFilterEntries_ToolMatch_WithOtherFilters(t *testing.T) {
	// Test ToolMatch combined with existing filters
	gitEntry := makeAssistantWithTools("1", struct{ name, input string }{"Bash", `{"command":"git status"}`})
	gitEntry.Timestamp = "2026-02-01T10:00:00.000Z"
	gitEntry.Type = models.EntryTypeAssistant

	npmEntry := makeAssistantWithTools("2", struct{ name, input string }{"Bash", `{"command":"npm install"}`})
	npmEntry.Timestamp = "2026-02-01T11:00:00.000Z"
	npmEntry.Type = models.EntryTypeAssistant

	goFileEntry := makeAssistantWithTools("3", struct{ name, input string }{"Read", `{"file_path":"/test/main.go"}`})
	goFileEntry.Timestamp = "2026-02-01T12:00:00.000Z"
	goFileEntry.Type = models.EntryTypeAssistant

	userEntry := models.ConversationEntry{UUID: "4", Type: models.EntryTypeUser, Timestamp: "2026-02-01T13:00:00.000Z"}

	entries := []models.ConversationEntry{gitEntry, npmEntry, goFileEntry, userEntry}

	t.Run("tool match with entry type filter", func(t *testing.T) {
		result := FilterEntries(entries, FilterOptions{
			ToolMatch: "git",
			Types:     []models.EntryType{models.EntryTypeAssistant},
		})
		if len(result) != 1 {
			t.Errorf("Got %d entries, want 1", len(result))
		}
	})

	t.Run("tool match with time range filter", func(t *testing.T) {
		start := time.Date(2026, 2, 1, 10, 30, 0, 0, time.UTC)
		end := time.Date(2026, 2, 1, 13, 0, 0, 0, time.UTC)
		result := FilterEntries(entries, FilterOptions{
			ToolMatch: `\.(go|js)`,
			StartTime: &start,
			EndTime:   &end,
		})
		if len(result) != 1 {
			t.Errorf("Got %d entries, want 1 (goFileEntry only)", len(result))
		}
		if len(result) > 0 && result[0].UUID != "3" {
			t.Errorf("Got UUID %s, want 3", result[0].UUID)
		}
	})

	t.Run("tool match with all filters combined", func(t *testing.T) {
		start := time.Date(2026, 2, 1, 9, 0, 0, 0, time.UTC)
		end := time.Date(2026, 2, 1, 11, 30, 0, 0, time.UTC)
		result := FilterEntries(entries, FilterOptions{
			ToolMatch: "command",
			Types:     []models.EntryType{models.EntryTypeAssistant},
			StartTime: &start,
			EndTime:   &end,
		})
		if len(result) != 2 {
			t.Errorf("Got %d entries, want 2", len(result))
		}
	})

	t.Run("tool match with AgentID filter", func(t *testing.T) {
		// Create fresh entries with AgentID set
		gitEntry2 := makeAssistantWithTools("1", struct{ name, input string }{"Bash", `{"command":"git status"}`})
		gitEntry2.Timestamp = "2026-02-01T10:00:00.000Z"
		gitEntry2.Type = models.EntryTypeAssistant
		gitEntry2.AgentID = "agent-1"

		npmEntry2 := makeAssistantWithTools("2", struct{ name, input string }{"Bash", `{"command":"npm install"}`})
		npmEntry2.Timestamp = "2026-02-01T11:00:00.000Z"
		npmEntry2.Type = models.EntryTypeAssistant
		npmEntry2.AgentID = "agent-2"

		goFileEntry2 := makeAssistantWithTools("3", struct{ name, input string }{"Read", `{"file_path":"/test/main.go"}`})
		goFileEntry2.Timestamp = "2026-02-01T12:00:00.000Z"
		goFileEntry2.Type = models.EntryTypeAssistant
		goFileEntry2.AgentID = "agent-1"

		entries2 := []models.ConversationEntry{gitEntry2, npmEntry2, goFileEntry2}

		result := FilterEntries(entries2, FilterOptions{
			ToolMatch: "command",
			AgentID:   "agent-1",
		})
		if len(result) != 1 {
			t.Errorf("Got %d entries, want 1 (gitEntry only)", len(result))
		}
	})
}

func TestFilterEntries_ToolTypeAndMatch(t *testing.T) {
	bashGit := makeAssistantWithTools("1", struct{ name, input string }{"Bash", `{"command":"git status"}`})
	bashNpm := makeAssistantWithTools("2", struct{ name, input string }{"Bash", `{"command":"npm install"}`})
	readGo := makeAssistantWithTools("3", struct{ name, input string }{"Read", `{"file_path":"/test/main.go"}`})
	readPy := makeAssistantWithTools("4", struct{ name, input string }{"Read", `{"file_path":"/test/script.py"}`})
	writeGo := makeAssistantWithTools("5", struct{ name, input string }{"Write", `{"file_path":"/test/output.go"}`})
	bashLs := makeAssistantWithTools("6", struct{ name, input string }{"Bash", `{"command":"ls -la"}`})
	multiTool := makeAssistantWithTools("7",
		struct{ name, input string }{"Bash", `{"command":"git log"}`},
		struct{ name, input string }{"Read", `{"file_path":"/test/README.md"}`},
	)

	entries := []models.ConversationEntry{bashGit, bashNpm, readGo, readPy, writeGo, bashLs, multiTool}

	tests := []struct {
		name      string
		toolTypes []string
		toolMatch string
		wantCount int
		wantUUIDs []string
		desc      string
	}{
		{
			name:      "both filters match",
			toolTypes: []string{"Bash"},
			toolMatch: "git",
			wantCount: 2,
			wantUUIDs: []string{"1", "7"},
			desc:      "Bash tools with 'git' in input",
		},
		{
			name:      "tool type matches but pattern doesn't",
			toolTypes: []string{"Read"},
			toolMatch: "git",
			wantCount: 1,
			wantUUIDs: []string{"7"},
			desc:      "Multi-tool entry has Read AND git (in different tools)",
		},
		{
			name:      "pattern matches but tool type doesn't",
			toolTypes: []string{"Write"},
			toolMatch: "npm",
			wantCount: 0,
			wantUUIDs: []string{},
			desc:      "Write tools have no 'npm' in input",
		},
		{
			name:      "both match different tools",
			toolTypes: []string{"Read", "Write"},
			toolMatch: `\.go`,
			wantCount: 2,
			wantUUIDs: []string{"3", "5"},
			desc:      "Read or Write tools with .go files",
		},
		{
			name:      "multiple tool types one matches pattern",
			toolTypes: []string{"Bash", "Read"},
			toolMatch: "status",
			wantCount: 1,
			wantUUIDs: []string{"1"},
			desc:      "Only Bash git status matches",
		},
		{
			name:      "case insensitive tool type with pattern",
			toolTypes: []string{"bash"},
			toolMatch: "npm",
			wantCount: 1,
			wantUUIDs: []string{"2"},
			desc:      "Case-insensitive Bash with npm pattern",
		},
		{
			name:      "multi-tool entry matches both filters",
			toolTypes: []string{"Bash"},
			toolMatch: "log",
			wantCount: 1,
			wantUUIDs: []string{"7"},
			desc:      "Multi-tool entry has Bash with 'log'",
		},
		{
			name:      "multi-tool entry matches type but not pattern",
			toolTypes: []string{"Read"},
			toolMatch: "status",
			wantCount: 0,
			wantUUIDs: []string{},
			desc:      "Multi-tool has Read but only Bash has 'status'",
		},
		{
			name:      "empty tool type ignores pattern",
			toolTypes: []string{},
			toolMatch: "git",
			wantCount: 2,
			wantUUIDs: []string{"1", "7"},
			desc:      "Empty ToolTypes, only ToolMatch applied",
		},
		{
			name:      "empty pattern ignores tool type",
			toolTypes: []string{"Bash"},
			toolMatch: "",
			wantCount: 4,
			wantUUIDs: []string{"1", "2", "6", "7"},
			desc:      "Empty ToolMatch, only ToolTypes applied",
		},
		{
			name:      "both empty returns all",
			toolTypes: []string{},
			toolMatch: "",
			wantCount: 7,
			wantUUIDs: []string{"1", "2", "3", "4", "5", "6", "7"},
			desc:      "No filtering applied",
		},
		{
			name:      "complex regex with multiple tool types",
			toolTypes: []string{"Read", "Write"},
			toolMatch: `\.(go|py)`,
			wantCount: 3,
			wantUUIDs: []string{"3", "4", "5"},
			desc:      "Read/Write tools with .go or .py files",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterEntries(entries, FilterOptions{
				ToolTypes: tt.toolTypes,
				ToolMatch: tt.toolMatch,
			})
			if len(result) != tt.wantCount {
				t.Errorf("%s: Got %d entries, want %d", tt.desc, len(result), tt.wantCount)
				for i, r := range result {
					t.Logf("  Result[%d]: UUID=%s", i, r.UUID)
				}
			}
			// Verify correct entries returned when count matches
			if len(result) == len(tt.wantUUIDs) {
				for i, uuid := range tt.wantUUIDs {
					if result[i].UUID != uuid {
						t.Errorf("Entry %d: got UUID %s, want %s", i, result[i].UUID, uuid)
					}
				}
			}
		})
	}
}

func TestFilterEntries_ToolTypeAndMatch_WithAllFilters(t *testing.T) {
	// Test ToolTypes + ToolMatch + Types + Time + AgentID all together
	bashGit := makeAssistantWithTools("1", struct{ name, input string }{"Bash", `{"command":"git status"}`})
	bashGit.Timestamp = "2026-02-01T10:00:00.000Z"
	bashGit.Type = models.EntryTypeAssistant
	bashGit.AgentID = "agent-1"

	bashNpm := makeAssistantWithTools("2", struct{ name, input string }{"Bash", `{"command":"npm install"}`})
	bashNpm.Timestamp = "2026-02-01T11:00:00.000Z"
	bashNpm.Type = models.EntryTypeAssistant
	bashNpm.AgentID = "agent-2"

	readGo := makeAssistantWithTools("3", struct{ name, input string }{"Read", `{"file_path":"/test/main.go"}`})
	readGo.Timestamp = "2026-02-01T12:00:00.000Z"
	readGo.Type = models.EntryTypeAssistant
	readGo.AgentID = "agent-1"

	userEntry := models.ConversationEntry{
		UUID:      "4",
		Type:      models.EntryTypeUser,
		Timestamp: "2026-02-01T13:00:00.000Z",
		AgentID:   "agent-1",
	}

	entries := []models.ConversationEntry{bashGit, bashNpm, readGo, userEntry}

	t.Run("all filters combined", func(t *testing.T) {
		start := time.Date(2026, 2, 1, 9, 0, 0, 0, time.UTC)
		end := time.Date(2026, 2, 1, 10, 30, 0, 0, time.UTC)
		result := FilterEntries(entries, FilterOptions{
			ToolTypes: []string{"Bash"},
			ToolMatch: "git",
			Types:     []models.EntryType{models.EntryTypeAssistant},
			StartTime: &start,
			EndTime:   &end,
			AgentID:   "agent-1",
		})
		if len(result) != 1 {
			t.Errorf("Got %d entries, want 1 (only bashGit matches all)", len(result))
		}
		if len(result) > 0 && result[0].UUID != "1" {
			t.Errorf("Expected entry 1 (bashGit), got %s", result[0].UUID)
		}
	})

	t.Run("all filters but wrong agent", func(t *testing.T) {
		start := time.Date(2026, 2, 1, 9, 0, 0, 0, time.UTC)
		end := time.Date(2026, 2, 1, 10, 30, 0, 0, time.UTC)
		result := FilterEntries(entries, FilterOptions{
			ToolTypes: []string{"Bash"},
			ToolMatch: "git",
			Types:     []models.EntryType{models.EntryTypeAssistant},
			StartTime: &start,
			EndTime:   &end,
			AgentID:   "agent-2",
		})
		if len(result) != 0 {
			t.Errorf("Got %d entries, want 0 (wrong agent)", len(result))
		}
	})

	t.Run("all filters but wrong time range", func(t *testing.T) {
		start := time.Date(2026, 2, 1, 11, 0, 0, 0, time.UTC)
		end := time.Date(2026, 2, 1, 13, 0, 0, 0, time.UTC)
		result := FilterEntries(entries, FilterOptions{
			ToolTypes: []string{"Bash"},
			ToolMatch: "git",
			Types:     []models.EntryType{models.EntryTypeAssistant},
			StartTime: &start,
			EndTime:   &end,
			AgentID:   "agent-1",
		})
		if len(result) != 0 {
			t.Errorf("Got %d entries, want 0 (bashGit outside time range)", len(result))
		}
	})

	t.Run("tool filters with multiple matching entries", func(t *testing.T) {
		start := time.Date(2026, 2, 1, 9, 0, 0, 0, time.UTC)
		end := time.Date(2026, 2, 1, 12, 30, 0, 0, time.UTC)
		result := FilterEntries(entries, FilterOptions{
			ToolTypes: []string{"Bash", "Read"},
			ToolMatch: "(git|main)",
			Types:     []models.EntryType{models.EntryTypeAssistant},
			StartTime: &start,
			EndTime:   &end,
			AgentID:   "agent-1",
		})
		if len(result) != 2 {
			t.Errorf("Got %d entries, want 2 (bashGit and readGo)", len(result))
		}
	})
}

// Verify the json import is used
var _ = json.Marshal
