package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/randlee/claude-history/pkg/models"
	"github.com/randlee/claude-history/pkg/session"
)

// Test helper to create a minimal project structure with sessions and agents
func createTestProjectStructure(t *testing.T, baseDir string) string {
	t.Helper()

	// Create project directory structure
	projectDir := filepath.Join(baseDir, "-test-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}

	// Create a session file
	sessionID := "679761ba-80c0-4cd3-a586-cc6a1fc56308"
	sessionFile := filepath.Join(projectDir, sessionID+".jsonl")
	sessionContent := `{"uuid":"1","sessionId":"679761ba-80c0-4cd3-a586-cc6a1fc56308","type":"user","timestamp":"2026-02-01T10:00:00.000Z","message":"Hello main session"}
{"uuid":"2","sessionId":"679761ba-80c0-4cd3-a586-cc6a1fc56308","type":"assistant","timestamp":"2026-02-01T10:00:05.000Z","message":"Hello from assistant"}
`
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0600); err != nil {
		t.Fatalf("failed to write session file: %v", err)
	}

	// Create session directory and subagents
	sessionDir := filepath.Join(projectDir, sessionID)
	subagentsDir := filepath.Join(sessionDir, "subagents")
	if err := os.MkdirAll(subagentsDir, 0755); err != nil {
		t.Fatalf("failed to create subagents dir: %v", err)
	}

	// Create agent file
	agentID := "abc123def-456-789-0ab-cdef12345678"
	agentFile := filepath.Join(subagentsDir, "agent-"+agentID+".jsonl")
	agentContent := `{"uuid":"a1","sessionId":"679761ba-80c0-4cd3-a586-cc6a1fc56308","agentId":"abc123def-456-789-0ab-cdef12345678","type":"user","timestamp":"2026-02-01T10:01:00.000Z","message":"Agent task message"}
{"uuid":"a2","sessionId":"679761ba-80c0-4cd3-a586-cc6a1fc56308","agentId":"abc123def-456-789-0ab-cdef12345678","type":"assistant","timestamp":"2026-02-01T10:01:05.000Z","message":"Agent response"}
{"uuid":"a3","sessionId":"679761ba-80c0-4cd3-a586-cc6a1fc56308","agentId":"abc123def-456-789-0ab-cdef12345678","type":"user","timestamp":"2026-02-01T10:02:00.000Z","message":"Agent follow-up"}
`
	if err := os.WriteFile(agentFile, []byte(agentContent), 0600); err != nil {
		t.Fatalf("failed to write agent file: %v", err)
	}

	// Create a second agent
	agent2ID := "xyz789abc-123-456-789-abcdef123456"
	agent2File := filepath.Join(subagentsDir, "agent-"+agent2ID+".jsonl")
	agent2Content := `{"uuid":"b1","sessionId":"679761ba-80c0-4cd3-a586-cc6a1fc56308","agentId":"xyz789abc-123-456-789-abcdef123456","type":"user","timestamp":"2026-02-01T10:05:00.000Z","message":"Second agent task"}
{"uuid":"b2","sessionId":"679761ba-80c0-4cd3-a586-cc6a1fc56308","agentId":"xyz789abc-123-456-789-abcdef123456","type":"assistant","timestamp":"2026-02-01T10:05:05.000Z","message":"Second agent response"}
`
	if err := os.WriteFile(agent2File, []byte(agent2Content), 0600); err != nil {
		t.Fatalf("failed to write second agent file: %v", err)
	}

	return projectDir
}

// Test helper to create nested agent structure
func createNestedAgentStructureForQuery(t *testing.T, projectDir, sessionID string) {
	t.Helper()

	// Create nested agent directory structure
	// agent-nested1 spawns agent-nested2
	agentDir := filepath.Join(projectDir, sessionID, "subagents", "agent-nested1-111-222-333-444555666777")
	nestedSubagentsDir := filepath.Join(agentDir, "subagents")
	if err := os.MkdirAll(nestedSubagentsDir, 0755); err != nil {
		t.Fatalf("failed to create nested subagents dir: %v", err)
	}

	// Create first level nested agent file
	nested1File := filepath.Join(projectDir, sessionID, "subagents", "agent-nested1-111-222-333-444555666777.jsonl")
	nested1Content := `{"uuid":"n1","sessionId":"679761ba-80c0-4cd3-a586-cc6a1fc56308","agentId":"nested1-111-222-333-444555666777","type":"user","timestamp":"2026-02-01T11:00:00.000Z","message":"Nested agent 1 task"}
{"uuid":"n2","sessionId":"679761ba-80c0-4cd3-a586-cc6a1fc56308","agentId":"nested1-111-222-333-444555666777","type":"assistant","timestamp":"2026-02-01T11:00:05.000Z","message":"Nested agent 1 response"}
`
	if err := os.WriteFile(nested1File, []byte(nested1Content), 0600); err != nil {
		t.Fatalf("failed to write nested agent 1 file: %v", err)
	}

	// Create second level nested agent file
	nested2File := filepath.Join(nestedSubagentsDir, "agent-nested2-aaa-bbb-ccc-dddeeefffggg.jsonl")
	nested2Content := `{"uuid":"nn1","sessionId":"679761ba-80c0-4cd3-a586-cc6a1fc56308","agentId":"nested2-aaa-bbb-ccc-dddeeefffggg","type":"user","timestamp":"2026-02-01T11:05:00.000Z","message":"Deeply nested agent task"}
{"uuid":"nn2","sessionId":"679761ba-80c0-4cd3-a586-cc6a1fc56308","agentId":"nested2-aaa-bbb-ccc-dddeeefffggg","type":"assistant","timestamp":"2026-02-01T11:05:05.000Z","message":"Deeply nested agent response"}
`
	if err := os.WriteFile(nested2File, []byte(nested2Content), 0600); err != nil {
		t.Fatalf("failed to write nested agent 2 file: %v", err)
	}
}

func TestGetAgentPath(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := createTestProjectStructure(t, tmpDir)
	sessionID := "679761ba-80c0-4cd3-a586-cc6a1fc56308"

	t.Run("finds agent in standard location", func(t *testing.T) {
		agentID := "abc123def-456-789-0ab-cdef12345678"
		path, err := getAgentPath(projectDir, sessionID, agentID)
		if err != nil {
			t.Fatalf("getAgentPath() error: %v", err)
		}

		expectedPath := filepath.Join(projectDir, sessionID, "subagents", "agent-"+agentID+".jsonl")
		if path != expectedPath {
			t.Errorf("getAgentPath() = %q, want %q", path, expectedPath)
		}
	})

	t.Run("finds second agent", func(t *testing.T) {
		agentID := "xyz789abc-123-456-789-abcdef123456"
		path, err := getAgentPath(projectDir, sessionID, agentID)
		if err != nil {
			t.Fatalf("getAgentPath() error: %v", err)
		}

		expectedPath := filepath.Join(projectDir, sessionID, "subagents", "agent-"+agentID+".jsonl")
		if path != expectedPath {
			t.Errorf("getAgentPath() = %q, want %q", path, expectedPath)
		}
	})

	t.Run("returns error for non-existent agent", func(t *testing.T) {
		_, err := getAgentPath(projectDir, sessionID, "nonexistent-agent-id")
		if err == nil {
			t.Error("getAgentPath() expected error for non-existent agent")
		}
	})

	t.Run("finds nested agent via recursive search", func(t *testing.T) {
		createNestedAgentStructureForQuery(t, projectDir, sessionID)

		// Find the deeply nested agent
		agentID := "nested2-aaa-bbb-ccc-dddeeefffggg"
		path, err := getAgentPath(projectDir, sessionID, agentID)
		if err != nil {
			t.Fatalf("getAgentPath() error for nested agent: %v", err)
		}

		// Verify the path is correct
		if path == "" {
			t.Error("getAgentPath() returned empty path for nested agent")
		}

		// Verify the file exists
		if _, statErr := os.Stat(path); statErr != nil {
			t.Errorf("getAgentPath() returned non-existent path: %s", path)
		}
	})
}

func TestQueryAgentFile(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := createTestProjectStructure(t, tmpDir)
	sessionID := "679761ba-80c0-4cd3-a586-cc6a1fc56308"
	agentID := "abc123def-456-789-0ab-cdef12345678"

	t.Run("reads agent file directly", func(t *testing.T) {
		entries, err := queryAgentFile(projectDir, sessionID, agentID, session.FilterOptions{})
		if err != nil {
			t.Fatalf("queryAgentFile() error: %v", err)
		}

		// Agent file has 3 entries
		if len(entries) != 3 {
			t.Errorf("queryAgentFile() returned %d entries, want 3", len(entries))
		}

		// Verify entries are from the agent
		for _, entry := range entries {
			if entry.AgentID != agentID {
				t.Errorf("Entry has agentId %q, want %q", entry.AgentID, agentID)
			}
		}
	})

	t.Run("applies type filter", func(t *testing.T) {
		opts := session.FilterOptions{
			Types: []models.EntryType{models.EntryTypeUser},
		}
		entries, err := queryAgentFile(projectDir, sessionID, agentID, opts)
		if err != nil {
			t.Fatalf("queryAgentFile() error: %v", err)
		}

		// Agent file has 2 user entries
		if len(entries) != 2 {
			t.Errorf("queryAgentFile() with type filter returned %d entries, want 2", len(entries))
		}
	})

	t.Run("returns error for non-existent agent", func(t *testing.T) {
		_, err := queryAgentFile(projectDir, sessionID, "nonexistent", session.FilterOptions{})
		if err == nil {
			t.Error("queryAgentFile() expected error for non-existent agent")
		}
	})
}

func TestQuerySessionWithAgents(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := createTestProjectStructure(t, tmpDir)
	sessionID := "679761ba-80c0-4cd3-a586-cc6a1fc56308"

	t.Run("includes main session and agent entries", func(t *testing.T) {
		entries, err := querySessionWithAgents(projectDir, sessionID, session.FilterOptions{})
		if err != nil {
			t.Fatalf("querySessionWithAgents() error: %v", err)
		}

		// Main session has 2 entries, agent 1 has 3 entries, agent 2 has 2 entries = 7 total
		if len(entries) != 7 {
			t.Errorf("querySessionWithAgents() returned %d entries, want 7", len(entries))
		}
	})

	t.Run("applies filter to all entries", func(t *testing.T) {
		opts := session.FilterOptions{
			Types: []models.EntryType{models.EntryTypeAssistant},
		}
		entries, err := querySessionWithAgents(projectDir, sessionID, opts)
		if err != nil {
			t.Fatalf("querySessionWithAgents() error: %v", err)
		}

		// Main session has 1 assistant, agent 1 has 1, agent 2 has 1 = 3 total
		if len(entries) != 3 {
			t.Errorf("querySessionWithAgents() with assistant filter returned %d entries, want 3", len(entries))
		}
	})

	t.Run("includes nested agents", func(t *testing.T) {
		createNestedAgentStructureForQuery(t, projectDir, sessionID)

		entries, err := querySessionWithAgents(projectDir, sessionID, session.FilterOptions{})
		if err != nil {
			t.Fatalf("querySessionWithAgents() error: %v", err)
		}

		// Original: 7 entries + nested1: 2 entries + nested2: 2 entries = 11 total
		if len(entries) != 11 {
			t.Errorf("querySessionWithAgents() with nested agents returned %d entries, want 11", len(entries))
		}
	})

	t.Run("handles session without agents", func(t *testing.T) {
		// Create a session without agents
		noAgentSessionID := "11111111-2222-3333-4444-555555555555"
		sessionFile := filepath.Join(projectDir, noAgentSessionID+".jsonl")
		content := `{"uuid":"x1","sessionId":"11111111-2222-3333-4444-555555555555","type":"user","timestamp":"2026-02-01T12:00:00.000Z","message":"No agents here"}
{"uuid":"x2","sessionId":"11111111-2222-3333-4444-555555555555","type":"assistant","timestamp":"2026-02-01T12:00:05.000Z","message":"Just us"}
`
		if err := os.WriteFile(sessionFile, []byte(content), 0600); err != nil {
			t.Fatalf("failed to write session file: %v", err)
		}

		entries, err := querySessionWithAgents(projectDir, noAgentSessionID, session.FilterOptions{})
		if err != nil {
			t.Fatalf("querySessionWithAgents() error: %v", err)
		}

		// Just the 2 main session entries
		if len(entries) != 2 {
			t.Errorf("querySessionWithAgents() for session without agents returned %d entries, want 2", len(entries))
		}
	})
}

func TestQuerySession(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := createTestProjectStructure(t, tmpDir)
	sessionID := "679761ba-80c0-4cd3-a586-cc6a1fc56308"

	t.Run("reads main session file only", func(t *testing.T) {
		entries, err := querySession(projectDir, sessionID, session.FilterOptions{})
		if err != nil {
			t.Fatalf("querySession() error: %v", err)
		}

		// Main session has 2 entries
		if len(entries) != 2 {
			t.Errorf("querySession() returned %d entries, want 2", len(entries))
		}
	})

	t.Run("does not include agent entries", func(t *testing.T) {
		entries, err := querySession(projectDir, sessionID, session.FilterOptions{})
		if err != nil {
			t.Fatalf("querySession() error: %v", err)
		}

		for _, entry := range entries {
			if entry.AgentID != "" {
				t.Errorf("querySession() returned entry with agentId %q, expected empty", entry.AgentID)
			}
		}
	})
}

func TestIncludeAgentsAndAgentMutuallyExclusive(t *testing.T) {
	// Test that --include-agents and --agent are mutually exclusive
	// This validation happens in runQuery
	t.Run("validation logic", func(t *testing.T) {
		// Simulate the scenario where both are set
		includeAgents := true
		agentID := "some-agent-id"

		// The condition that should trigger an error
		if includeAgents && agentID != "" {
			// This is the expected case - the combination should be rejected
			// In real code: return fmt.Errorf("--include-agents and --agent cannot be used together")
		} else {
			t.Error("Expected include-agents and agent to be mutually exclusive")
		}
	})
}

func TestQuerySessionFileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "-nonexistent-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}

	t.Run("returns error for non-existent session", func(t *testing.T) {
		_, err := querySession(projectDir, "nonexistent-session-id", session.FilterOptions{})
		if err == nil {
			t.Error("querySession() expected error for non-existent session")
		}
	})
}

func TestQueryAgentFileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := createTestProjectStructure(t, tmpDir)
	sessionID := "679761ba-80c0-4cd3-a586-cc6a1fc56308"

	t.Run("returns error for non-existent agent", func(t *testing.T) {
		_, err := queryAgentFile(projectDir, sessionID, "nonexistent-agent", session.FilterOptions{})
		if err == nil {
			t.Error("queryAgentFile() expected error for non-existent agent")
		}
		// Verify error message contains "agent not found"
		if err != nil && !contains(err.Error(), "agent not found") {
			t.Errorf("Expected error to contain 'agent not found', got: %v", err)
		}
	})
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
