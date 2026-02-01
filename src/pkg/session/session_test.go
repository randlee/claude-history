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

	entries := []models.ConversationEntry{bashEntry, readEntry, multiEntry, userEntry}

	t.Run("filter for Bash", func(t *testing.T) {
		result := FilterEntries(entries, FilterOptions{
			ToolTypes: []string{"Bash"},
		})
		if len(result) != 2 {
			t.Errorf("Got %d entries, want 2 (bash and multi)", len(result))
		}
	})

	t.Run("filter for bash lowercase", func(t *testing.T) {
		result := FilterEntries(entries, FilterOptions{
			ToolTypes: []string{"bash"},
		})
		if len(result) != 2 {
			t.Errorf("Got %d entries, want 2 (case-insensitive)", len(result))
		}
	})

	t.Run("filter for multiple tools", func(t *testing.T) {
		result := FilterEntries(entries, FilterOptions{
			ToolTypes: []string{"Read", "Write"},
		})
		if len(result) != 2 {
			t.Errorf("Got %d entries, want 2 (read and multi)", len(result))
		}
	})

	t.Run("filter for non-existent tool", func(t *testing.T) {
		result := FilterEntries(entries, FilterOptions{
			ToolTypes: []string{"NonExistent"},
		})
		if len(result) != 0 {
			t.Errorf("Got %d entries, want 0", len(result))
		}
	})
}

func TestFilterEntries_ToolMatch(t *testing.T) {
	gitEntry := makeAssistantWithTools("1", struct{ name, input string }{"Bash", `{"command":"git status"}`})
	npmEntry := makeAssistantWithTools("2", struct{ name, input string }{"Bash", `{"command":"npm install"}`})
	goFileEntry := makeAssistantWithTools("3", struct{ name, input string }{"Read", `{"file_path":"/Users/test/main.go"}`})

	entries := []models.ConversationEntry{gitEntry, npmEntry, goFileEntry}

	t.Run("match git commands", func(t *testing.T) {
		result := FilterEntries(entries, FilterOptions{
			ToolMatch: "git",
		})
		if len(result) != 1 {
			t.Errorf("Got %d entries, want 1", len(result))
		}
	})

	t.Run("match .go files", func(t *testing.T) {
		result := FilterEntries(entries, FilterOptions{
			ToolMatch: `\.go`,
		})
		if len(result) != 1 {
			t.Errorf("Got %d entries, want 1", len(result))
		}
	})

	t.Run("invalid regex returns no matches", func(t *testing.T) {
		result := FilterEntries(entries, FilterOptions{
			ToolMatch: "[invalid",
		})
		if len(result) != 0 {
			t.Errorf("Got %d entries, want 0 for invalid regex", len(result))
		}
	})
}

func TestFilterEntries_ToolTypeAndMatch(t *testing.T) {
	bashGit := makeAssistantWithTools("1", struct{ name, input string }{"Bash", `{"command":"git status"}`})
	bashNpm := makeAssistantWithTools("2", struct{ name, input string }{"Bash", `{"command":"npm install"}`})
	readGo := makeAssistantWithTools("3", struct{ name, input string }{"Read", `{"file_path":"/test/main.go"}`})

	entries := []models.ConversationEntry{bashGit, bashNpm, readGo}

	t.Run("both filters must match", func(t *testing.T) {
		result := FilterEntries(entries, FilterOptions{
			ToolTypes: []string{"Bash"},
			ToolMatch: "git",
		})
		if len(result) != 1 {
			t.Errorf("Got %d entries, want 1 (only bashGit matches both)", len(result))
		}
		if len(result) > 0 && result[0].UUID != "1" {
			t.Errorf("Expected entry 1 (bashGit), got %s", result[0].UUID)
		}
	})

	t.Run("tool type matches but pattern doesn't", func(t *testing.T) {
		result := FilterEntries(entries, FilterOptions{
			ToolTypes: []string{"Read"},
			ToolMatch: "git",
		})
		if len(result) != 0 {
			t.Errorf("Got %d entries, want 0", len(result))
		}
	})
}

// Verify the json import is used
var _ = json.Marshal
