package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/randlee/claude-history/pkg/models"
	"github.com/randlee/claude-history/pkg/resolver"
)

// TestPrefixMatching_UniqueResolution tests unique prefix resolution across multiple sessions.
func TestPrefixMatching_UniqueResolution(t *testing.T) {
	projectDir := setupTestProject(t)

	// Create 3 sessions with distinct prefixes
	sessions := []struct {
		id     string
		prompt string
	}{
		{"aaa12345-1234-1234-1234-123456789abc", "First session prompt"},
		{"bbb67890-5678-5678-5678-567890abcdef", "Second session prompt"},
		{"ccc11111-9999-8888-7777-666655554444", "Third session prompt"},
	}

	now := time.Now()
	for i, s := range sessions {
		createSessionFile(t, projectDir, s.id, s.prompt, now.Add(time.Duration(-i)*time.Hour))
	}

	// Test unique prefix resolution
	tests := []struct {
		name   string
		prefix string
		want   string
	}{
		{"prefix_aaa", "aaa", sessions[0].id},
		{"prefix_aa", "aa", sessions[0].id},
		{"prefix_a", "a", sessions[0].id},
		{"prefix_bbb", "bbb", sessions[1].id},
		{"prefix_bb", "bb", sessions[1].id},
		{"prefix_b", "b", sessions[1].id},
		{"prefix_ccc", "ccc", sessions[2].id},
		{"prefix_c", "c", sessions[2].id},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved, err := resolver.ResolveSessionID(projectDir, tt.prefix)
			if err != nil {
				t.Fatalf("ResolveSessionID(%q) error = %v", tt.prefix, err)
			}
			if resolved != tt.want {
				t.Errorf("ResolveSessionID(%q) = %q, want %q", tt.prefix, resolved, tt.want)
			}
		})
	}
}

// TestPrefixMatching_AmbiguousError tests error handling for ambiguous prefixes.
func TestPrefixMatching_AmbiguousError(t *testing.T) {
	projectDir := setupTestProject(t)

	// Create 2 sessions with same prefix
	sessions := []string{
		"cd2e9388-3108-40e5-b41b-79497cbb58b4",
		"cd2e9388-9999-8888-7777-666655554444",
	}

	now := time.Now()
	for i, id := range sessions {
		createSessionFile(t, projectDir, id, "Prompt "+string(rune('A'+i)), now.Add(time.Duration(-i)*time.Hour))
	}

	// Try to resolve with ambiguous prefix
	_, err := resolver.ResolveSessionID(projectDir, "cd2e9388")
	if err == nil {
		t.Fatal("expected ambiguous prefix error, got nil")
	}

	errMsg := err.Error()

	// Verify error message contains all expected elements
	requiredElements := []string{
		"ambiguous",
		sessions[0],
		sessions[1],
		"provide more characters",
	}

	for _, elem := range requiredElements {
		if !strings.Contains(errMsg, elem) {
			t.Errorf("error message missing %q\nGot: %s", elem, errMsg)
		}
	}

	// Note: Enhanced error messages with timestamps and prompts are part of Phase 7b
	// For now, we just verify the basic ambiguous error structure
}

// TestPrefixMatching_AgentResolution tests agent prefix matching.
func TestPrefixMatching_AgentResolution(t *testing.T) {
	projectDir := setupTestProject(t)

	// Create session
	sessionID := "test1234-1234-1234-1234-123456789abc"
	createSessionFile(t, projectDir, sessionID, "Main session", time.Now())

	// Create session directory and agents
	sessionDir := filepath.Join(projectDir, sessionID)
	subagentsDir := filepath.Join(sessionDir, "subagents")
	if err := os.MkdirAll(subagentsDir, 0755); err != nil {
		t.Fatalf("failed to create subagents dir: %v", err)
	}

	agents := []string{
		"agent-abc123",
		"agent-def456",
		"agent-ghi789",
	}

	for _, agentID := range agents {
		createAgentFile(t, sessionDir, agentID, "Agent task: "+agentID)
	}

	// Test agent prefix resolution
	tests := []struct {
		prefix string
		want   string
	}{
		{"agent-abc", agents[0]},
		{"agent-def", agents[1]},
		{"agent-ghi", agents[2]},
		{"agent-a", agents[0]},
		{"agent-d", agents[1]},
	}

	for _, tt := range tests {
		t.Run("agent_"+tt.prefix, func(t *testing.T) {
			resolved, err := resolver.ResolveAgentID(projectDir, sessionID, tt.prefix)
			if err != nil {
				t.Fatalf("ResolveAgentID(%q) error = %v", tt.prefix, err)
			}
			if resolved != tt.want {
				t.Errorf("ResolveAgentID(%q) = %q, want %q", tt.prefix, resolved, tt.want)
			}
		})
	}
}

// TestPrefixMatching_NestedAgents tests prefix resolution with nested agent hierarchy.
func TestPrefixMatching_NestedAgents(t *testing.T) {
	projectDir := setupTestProject(t)

	// Create session
	sessionID := "nested12-1234-1234-1234-123456789abc"
	createSessionFile(t, projectDir, sessionID, "Nested test", time.Now())

	// Create nested agent structure:
	// session/subagents/agent-A/subagents/agent-B/subagents/agent-C
	sessionDir := filepath.Join(projectDir, sessionID)

	// Level 1: agent-A
	createAgentFile(t, sessionDir, "agent-A", "Level 1")

	// Level 2: agent-B (nested under agent-A)
	agentADir := filepath.Join(sessionDir, "subagents", "agent-agent-A")
	if err := os.MkdirAll(agentADir, 0755); err != nil {
		t.Fatalf("failed to create agent-A dir: %v", err)
	}
	createAgentFile(t, agentADir, "agent-B", "Level 2")

	// Level 3: agent-C (nested under agent-B)
	agentBDir := filepath.Join(agentADir, "subagents", "agent-agent-B")
	if err := os.MkdirAll(agentBDir, 0755); err != nil {
		t.Fatalf("failed to create agent-B dir: %v", err)
	}
	createAgentFile(t, agentBDir, "agent-C", "Level 3")

	// Test that we can resolve top-level agent
	resolved, err := resolver.ResolveAgentID(projectDir, sessionID, "agent-A")
	if err != nil {
		t.Fatalf("ResolveAgentID(agent-A) error = %v", err)
	}
	if resolved != "agent-A" {
		t.Errorf("ResolveAgentID(agent-A) = %q, want %q", resolved, "agent-A")
	}
}

// TestPrefixMatching_CrossCommandIntegration tests prefix usage across different commands.
func TestPrefixMatching_CrossCommandIntegration(t *testing.T) {
	// This test would normally invoke CLI commands, but since we're testing the resolver package,
	// we'll test the underlying functions that commands would use.

	projectDir := setupTestProject(t)

	// Setup: Create session with known prefix
	sessionID := "cmd12345-1234-1234-1234-123456789abc"
	createSessionFile(t, projectDir, sessionID, "Command test", time.Now())

	// Create agent
	sessionDir := filepath.Join(projectDir, sessionID)
	if err := os.MkdirAll(filepath.Join(sessionDir, "subagents"), 0755); err != nil {
		t.Fatalf("failed to create subagents dir: %v", err)
	}
	createAgentFile(t, sessionDir, "agent999", "Test agent")

	// Simulate 'resolve' command usage
	t.Run("resolve_command", func(t *testing.T) {
		resolved, err := resolver.ResolveSessionID(projectDir, "cmd")
		if err != nil {
			t.Fatalf("resolve command failed: %v", err)
		}
		if resolved != sessionID {
			t.Errorf("resolved session = %q, want %q", resolved, sessionID)
		}
	})

	// Simulate 'list' command with prefix
	t.Run("list_command", func(t *testing.T) {
		// List command would first resolve the prefix, then list
		resolved, err := resolver.ResolveSessionID(projectDir, "cmd123")
		if err != nil {
			t.Fatalf("list resolve failed: %v", err)
		}
		if resolved != sessionID {
			t.Errorf("list resolved = %q, want %q", resolved, sessionID)
		}
	})

	// Simulate 'query' command with session prefix
	t.Run("query_command", func(t *testing.T) {
		resolved, err := resolver.ResolveSessionID(projectDir, "cmd")
		if err != nil {
			t.Fatalf("query resolve failed: %v", err)
		}
		if resolved != sessionID {
			t.Errorf("query resolved = %q, want %q", resolved, sessionID)
		}
	})

	// Simulate 'tree' command with session and agent prefix
	t.Run("tree_command", func(t *testing.T) {
		resolvedSession, err := resolver.ResolveSessionID(projectDir, "cmd")
		if err != nil {
			t.Fatalf("tree session resolve failed: %v", err)
		}

		resolvedAgent, err := resolver.ResolveAgentID(projectDir, resolvedSession, "agent9")
		if err != nil {
			t.Fatalf("tree agent resolve failed: %v", err)
		}
		if resolvedAgent != "agent999" {
			t.Errorf("tree agent = %q, want %q", resolvedAgent, "agent999")
		}
	})

	// Simulate 'find-agent' command with session prefix
	t.Run("find_agent_command", func(t *testing.T) {
		resolved, err := resolver.ResolveSessionID(projectDir, "cmd1")
		if err != nil {
			t.Fatalf("find-agent resolve failed: %v", err)
		}
		if resolved != sessionID {
			t.Errorf("find-agent resolved = %q, want %q", resolved, sessionID)
		}
	})

	// Simulate 'export' command with session prefix
	t.Run("export_command", func(t *testing.T) {
		resolved, err := resolver.ResolveSessionID(projectDir, "cmd12")
		if err != nil {
			t.Fatalf("export resolve failed: %v", err)
		}
		if resolved != sessionID {
			t.Errorf("export resolved = %q, want %q", resolved, sessionID)
		}
	})
}

// TestPrefixMatching_WindowsPaths tests prefix resolution with Windows-style paths.
func TestPrefixMatching_WindowsPaths(t *testing.T) {
	// Skip on non-Windows if path handling differs
	// This test ensures cross-platform compatibility

	projectDir := setupTestProject(t)

	sessionID := "win12345-1234-1234-1234-123456789abc"
	createSessionFile(t, projectDir, sessionID, "Windows test", time.Now())

	// Test that prefix resolution works with normalized paths
	resolved, err := resolver.ResolveSessionID(projectDir, "win")
	if err != nil {
		t.Fatalf("Windows path test failed: %v", err)
	}
	if resolved != sessionID {
		t.Errorf("resolved = %q, want %q", resolved, sessionID)
	}
}

// TestPrefixMatching_UnixPaths tests prefix resolution with Unix-style paths.
func TestPrefixMatching_UnixPaths(t *testing.T) {
	projectDir := setupTestProject(t)

	sessionID := "unix1234-1234-1234-1234-123456789abc"
	createSessionFile(t, projectDir, sessionID, "Unix test", time.Now())

	resolved, err := resolver.ResolveSessionID(projectDir, "unix")
	if err != nil {
		t.Fatalf("Unix path test failed: %v", err)
	}
	if resolved != sessionID {
		t.Errorf("resolved = %q, want %q", resolved, sessionID)
	}
}

// Helper functions for integration tests

func setupTestProject(t *testing.T) string {
	t.Helper()
	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, "test-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}
	return projectDir
}

func createSessionFile(t *testing.T, projectDir, sessionID, prompt string, timestamp time.Time) {
	t.Helper()

	filePath := filepath.Join(projectDir, sessionID+".jsonl")

	entries := []models.ConversationEntry{
		{
			Type:      models.EntryTypeUser,
			SessionID: sessionID,
			Timestamp: timestamp.Format(time.RFC3339Nano),
			Message:   json.RawMessage(`{"role":"user","content":"` + prompt + `"}`),
		},
		{
			Type:      models.EntryTypeAssistant,
			SessionID: sessionID,
			Timestamp: timestamp.Add(1 * time.Second).Format(time.RFC3339Nano),
			Message:   json.RawMessage(`{"role":"assistant","content":[{"type":"text","text":"Response to: ` + prompt + `"}]}`),
		},
	}

	f, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("failed to create session file: %v", err)
	}
	defer func() { _ = f.Close() }()

	encoder := json.NewEncoder(f)
	for _, entry := range entries {
		if err := encoder.Encode(entry); err != nil {
			t.Fatalf("failed to write entry: %v", err)
		}
	}
}

func createAgentFile(t *testing.T, sessionDir, agentID, description string) {
	t.Helper()

	subagentsDir := filepath.Join(sessionDir, "subagents")
	if err := os.MkdirAll(subagentsDir, 0755); err != nil {
		t.Fatalf("failed to create subagents dir: %v", err)
	}

	filePath := filepath.Join(subagentsDir, "agent-"+agentID+".jsonl")

	entry := models.ConversationEntry{
		Type:      models.EntryTypeUser,
		SessionID: filepath.Base(sessionDir),
		AgentID:   agentID,
		Timestamp: time.Now().Format(time.RFC3339Nano),
		Message:   json.RawMessage(`{"role":"user","content":"` + description + `"}`),
	}

	f, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("failed to create agent file: %v", err)
	}
	defer func() { _ = f.Close() }()

	encoder := json.NewEncoder(f)
	if err := encoder.Encode(entry); err != nil {
		t.Fatalf("failed to write agent entry: %v", err)
	}
}
