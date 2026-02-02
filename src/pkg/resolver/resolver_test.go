package resolver

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// setupTestProject creates a temporary project directory with test sessions.
func setupTestProject(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()

	// Create session files with different prefixes
	sessions := []struct {
		id     string
		prompt string
		ts     string
	}{
		{
			id:     "cd2e9388-3108-40e5-b41b-79497cbb58b4",
			prompt: "read CLAUDE.md and docs/project-plan let's do some work",
			ts:     "2026-02-02T01:50:37.000Z",
		},
		{
			id:     "cd2e4f21-9a14-4b29-8d3c-f5e8a9c1d7e2",
			prompt: "fix the bug in auth handler",
			ts:     "2026-01-30T14:23:11.000Z",
		},
		{
			id:     "ab123456-7890-1234-5678-90abcdef1234",
			prompt: "implement new feature for user management",
			ts:     "2026-01-28T10:00:00.000Z",
		},
		{
			id:     "xyz98765-4321-9876-5432-10fedcba9876",
			prompt: "refactor database layer for better performance",
			ts:     "2026-01-25T08:30:00.000Z",
		},
	}

	for _, s := range sessions {
		filePath := filepath.Join(tmpDir, s.id+".jsonl")
		content := `{"uuid":"entry-1","sessionId":"` + s.id + `","type":"user","timestamp":"` + s.ts + `","message":"` + s.prompt + `","isSidechain":false}
{"uuid":"entry-2","sessionId":"` + s.id + `","type":"assistant","timestamp":"` + s.ts + `","message":[{"type":"text","text":"I'll help with that."}],"isSidechain":false}
`
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test session file: %v", err)
		}
	}

	return tmpDir
}

// setupTestAgents creates test agents in a session directory.
func setupTestAgents(t *testing.T, projectDir, sessionID string) {
	t.Helper()

	sessionDir := filepath.Join(projectDir, sessionID)
	subagentsDir := filepath.Join(sessionDir, "subagents")
	if err := os.MkdirAll(subagentsDir, 0755); err != nil {
		t.Fatalf("Failed to create subagents dir: %v", err)
	}

	// Create session file with queue-operation entries
	sessionFile := filepath.Join(projectDir, sessionID+".jsonl")
	sessionContent := `{"uuid":"entry-1","sessionId":"` + sessionID + `","type":"user","timestamp":"2026-02-01T10:00:00.000Z","message":"start task","isSidechain":false}
{"uuid":"entry-2","sessionId":"` + sessionID + `","type":"queue-operation","timestamp":"2026-02-01T10:00:01.000Z","agentId":"a12eb64","message":"Explore codebase structure","isSidechain":false}
{"uuid":"entry-3","sessionId":"` + sessionID + `","type":"queue-operation","timestamp":"2026-02-01T10:00:02.000Z","agentId":"a12ef99","message":"Analyze dependencies","isSidechain":false}
{"uuid":"entry-4","sessionId":"` + sessionID + `","type":"queue-operation","timestamp":"2026-02-01T10:00:03.000Z","agentId":"bcd4567","message":"Generate report","isSidechain":false}
`
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0644); err != nil {
		t.Fatalf("Failed to create session file: %v", err)
	}

	// Create agent files
	agents := []struct {
		id     string
		prompt string
		ts     string
	}{
		{
			id:     "a12eb64",
			prompt: "exploring the pkg/ directory to understand the structure",
			ts:     "2026-02-01T10:00:05.000Z",
		},
		{
			id:     "a12ef99",
			prompt: "analyzing go.mod and go.sum for dependencies",
			ts:     "2026-02-01T10:00:10.000Z",
		},
		{
			id:     "bcd4567",
			prompt: "generating markdown report of findings",
			ts:     "2026-02-01T10:00:15.000Z",
		},
	}

	for _, ag := range agents {
		filePath := filepath.Join(subagentsDir, "agent-"+ag.id+".jsonl")
		content := `{"uuid":"agent-entry-1","sessionId":"` + sessionID + `","agentId":"` + ag.id + `","type":"user","timestamp":"` + ag.ts + `","message":"` + ag.prompt + `","isSidechain":true}
{"uuid":"agent-entry-2","sessionId":"` + sessionID + `","agentId":"` + ag.id + `","type":"assistant","timestamp":"` + ag.ts + `","message":[{"type":"text","text":"Working on it..."}],"isSidechain":true}
`
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create agent file: %v", err)
		}
	}
}

func TestResolveSessionID_UniqueMatch(t *testing.T) {
	projectDir := setupTestProject(t)

	tests := []struct {
		name     string
		prefix   string
		expected string
	}{
		{
			name:     "unique prefix ab",
			prefix:   "ab",
			expected: "ab123456-7890-1234-5678-90abcdef1234",
		},
		{
			name:     "unique prefix xyz",
			prefix:   "xyz",
			expected: "xyz98765-4321-9876-5432-10fedcba9876",
		},
		{
			name:     "full UUID",
			prefix:   "ab123456-7890-1234-5678-90abcdef1234",
			expected: "ab123456-7890-1234-5678-90abcdef1234",
		},
		{
			name:     "longer unique prefix",
			prefix:   "cd2e93",
			expected: "cd2e9388-3108-40e5-b41b-79497cbb58b4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ResolveSessionID(projectDir, tt.prefix)
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestResolveSessionID_AmbiguousMatch(t *testing.T) {
	projectDir := setupTestProject(t)

	prefix := "cd2e"
	_, err := ResolveSessionID(projectDir, prefix)

	if err == nil {
		t.Fatal("Expected ambiguity error, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "ambiguous session ID prefix") {
		t.Errorf("Error should contain 'ambiguous session ID prefix', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, `"cd2e"`) {
		t.Errorf("Error should contain the prefix 'cd2e', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "matches 2 sessions") {
		t.Errorf("Error should mention 2 matches, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "cd2e9388-3108-40e5-b41b-79497cbb58b4") {
		t.Errorf("Error should list first matching session ID, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "cd2e4f21-9a14-4b29-8d3c-f5e8a9c1d7e2") {
		t.Errorf("Error should list second matching session ID, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "Date:") {
		t.Errorf("Error should include timestamps, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "Prompt:") {
		t.Errorf("Error should include prompts, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "Please provide more characters") {
		t.Errorf("Error should suggest providing more characters, got: %s", errMsg)
	}
}

func TestResolveSessionID_NoMatch(t *testing.T) {
	projectDir := setupTestProject(t)

	prefix := "nonexistent"
	_, err := ResolveSessionID(projectDir, prefix)

	if err == nil {
		t.Fatal("Expected error for no matches, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "no sessions found with prefix") {
		t.Errorf("Error should mention 'no sessions found', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, prefix) {
		t.Errorf("Error should include the prefix, got: %s", errMsg)
	}
}

func TestResolveSessionID_CaseSensitive(t *testing.T) {
	projectDir := setupTestProject(t)

	// Lowercase prefix should match
	result, err := ResolveSessionID(projectDir, "cd2e93")
	if err != nil {
		t.Fatalf("Expected match for lowercase, got error: %v", err)
	}
	if !strings.HasPrefix(result, "cd2e93") {
		t.Errorf("Expected result to start with cd2e93, got: %s", result)
	}

	// Uppercase prefix should NOT match (case-sensitive like git)
	_, err = ResolveSessionID(projectDir, "CD2E93")
	if err == nil {
		t.Fatal("Expected no match for uppercase prefix")
	}
	if !strings.Contains(err.Error(), "no sessions found") {
		t.Errorf("Expected 'no sessions found' error, got: %v", err)
	}
}

func TestResolveAgentID_UniqueMatch(t *testing.T) {
	projectDir := setupTestProject(t)
	sessionID := "test-session-12345678-1234-1234-1234-123456789abc"
	setupTestAgents(t, projectDir, sessionID)

	tests := []struct {
		name     string
		prefix   string
		expected string
	}{
		{
			name:     "unique prefix b",
			prefix:   "b",
			expected: "bcd4567",
		},
		{
			name:     "unique prefix a12eb",
			prefix:   "a12eb",
			expected: "a12eb64",
		},
		{
			name:     "full agent ID",
			prefix:   "a12ef99",
			expected: "a12ef99",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ResolveAgentID(projectDir, sessionID, tt.prefix)
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestResolveAgentID_AmbiguousMatch(t *testing.T) {
	projectDir := setupTestProject(t)
	sessionID := "test-session-12345678-1234-1234-1234-123456789abc"
	setupTestAgents(t, projectDir, sessionID)

	prefix := "a12e"
	_, err := ResolveAgentID(projectDir, sessionID, prefix)

	if err == nil {
		t.Fatal("Expected ambiguity error, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "ambiguous agent ID prefix") {
		t.Errorf("Error should contain 'ambiguous agent ID prefix', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, `"a12e"`) {
		t.Errorf("Error should contain the prefix 'a12e', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "matches 2 agents") {
		t.Errorf("Error should mention 2 matches, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "a12eb64") {
		t.Errorf("Error should list first matching agent ID, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "a12ef99") {
		t.Errorf("Error should list second matching agent ID, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "Date:") {
		t.Errorf("Error should include timestamps, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "Description:") {
		t.Errorf("Error should include descriptions, got: %s", errMsg)
	}
}

func TestResolveAgentID_NoMatch(t *testing.T) {
	projectDir := setupTestProject(t)
	sessionID := "test-session-12345678-1234-1234-1234-123456789abc"
	setupTestAgents(t, projectDir, sessionID)

	prefix := "xyz999"
	_, err := ResolveAgentID(projectDir, sessionID, prefix)

	if err == nil {
		t.Fatal("Expected error for no matches, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "no agents found with prefix") {
		t.Errorf("Error should mention 'no agents found', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, prefix) {
		t.Errorf("Error should include the prefix, got: %s", errMsg)
	}
}

func TestResolveAgentID_NoSubagentsDir(t *testing.T) {
	projectDir := setupTestProject(t)
	sessionID := "test-session-12345678-1234-1234-1234-123456789abc"
	// Don't create subagents - should gracefully handle missing directory

	_, err := ResolveAgentID(projectDir, sessionID, "a12")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// Should get "no agents found" error (not a directory read error)
	if !strings.Contains(err.Error(), "no agents found") {
		t.Errorf("Expected 'no agents found' error, got: %v", err)
	}
}

func TestFindMatchingSessionIDs(t *testing.T) {
	projectDir := setupTestProject(t)

	matches, err := findMatchingSessionIDs(projectDir, "cd2e")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(matches) != 2 {
		t.Fatalf("Expected 2 matches, got %d", len(matches))
	}

	// Verify match structure
	for _, match := range matches {
		if !strings.HasPrefix(match.ID, "cd2e") {
			t.Errorf("Match ID should start with cd2e, got: %s", match.ID)
		}
		if match.Path == "" {
			t.Error("Match path should not be empty")
		}
		if match.FirstPrompt == "" {
			t.Error("Match first prompt should not be empty")
		}
		if match.Timestamp.IsZero() {
			t.Error("Match timestamp should not be zero")
		}
		if !filepath.IsAbs(match.Path) {
			t.Errorf("Match path should be absolute, got: %s", match.Path)
		}
	}
}

func TestFindMatchingAgentIDs(t *testing.T) {
	projectDir := setupTestProject(t)
	sessionID := "test-session-12345678-1234-1234-1234-123456789abc"
	setupTestAgents(t, projectDir, sessionID)

	matches, err := findMatchingAgentIDs(projectDir, sessionID, "a12e")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(matches) != 2 {
		t.Fatalf("Expected 2 matches, got %d", len(matches))
	}

	// Verify match structure
	for _, match := range matches {
		if !strings.HasPrefix(match.ID, "a12e") {
			t.Errorf("Match ID should start with a12e, got: %s", match.ID)
		}
		if match.Path == "" {
			t.Error("Match path should not be empty")
		}
		if match.Description == "" {
			t.Error("Match description should not be empty")
		}
		if match.Timestamp.IsZero() {
			t.Error("Match timestamp should not be zero")
		}
		if !filepath.IsAbs(match.Path) {
			t.Errorf("Match path should be absolute, got: %s", match.Path)
		}
	}
}

func TestExtractAgentSpawnDescriptions(t *testing.T) {
	projectDir := setupTestProject(t)
	sessionID := "test-session-12345678-1234-1234-1234-123456789abc"
	setupTestAgents(t, projectDir, sessionID)

	sessionFile := filepath.Join(projectDir, sessionID+".jsonl")
	descs := extractAgentSpawnDescriptions(sessionFile)

	if len(descs) != 3 {
		t.Fatalf("Expected 3 spawn descriptions, got %d", len(descs))
	}

	expectedDescs := map[string]string{
		"a12eb64": "Explore codebase structure",
		"a12ef99": "Analyze dependencies",
		"bcd4567": "Generate report",
	}

	for agentID, expected := range expectedDescs {
		if desc, ok := descs[agentID]; !ok {
			t.Errorf("Missing description for agent %s", agentID)
		} else if desc != expected {
			t.Errorf("Expected description '%s' for agent %s, got '%s'", expected, agentID, desc)
		}
	}
}

func TestSessionMatch_PromptTruncation(t *testing.T) {
	tmpDir := t.TempDir()

	// Create session with long prompt (>60 chars)
	// Use proper UUID format that looksLikeUUID will accept
	sessionID := "12345678-1234-1234-1234-123456789abc"
	longPrompt := "This is a very long prompt that should be truncated to exactly sixty characters plus ellipsis"
	filePath := filepath.Join(tmpDir, sessionID+".jsonl")
	content := `{"uuid":"entry-1","sessionId":"` + sessionID + `","type":"user","timestamp":"2026-02-01T10:00:00.000Z","message":"` + longPrompt + `","isSidechain":false}
`
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	matches, err := findMatchingSessionIDs(tmpDir, "12345")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(matches) != 1 {
		t.Fatalf("Expected 1 match, got %d", len(matches))
	}

	prompt := matches[0].FirstPrompt
	if len(prompt) != 63 { // 60 chars + "..."
		t.Errorf("Expected truncated prompt length 63, got %d: %s", len(prompt), prompt)
	}
	if !strings.HasSuffix(prompt, "...") {
		t.Errorf("Expected prompt to end with '...', got: %s", prompt)
	}
}

func TestAgentMatch_DescriptionTruncation(t *testing.T) {
	projectDir := t.TempDir()
	sessionID := "test-session-12345678-1234-1234-1234-123456789abc"

	// Create session with long description
	sessionDir := filepath.Join(projectDir, sessionID)
	subagentsDir := filepath.Join(sessionDir, "subagents")
	if err := os.MkdirAll(subagentsDir, 0755); err != nil {
		t.Fatalf("Failed to create subagents dir: %v", err)
	}

	longDesc := "This is a very long agent description that should be truncated to exactly sixty characters plus ellipsis"
	sessionFile := filepath.Join(projectDir, sessionID+".jsonl")
	sessionContent := `{"uuid":"entry-1","sessionId":"` + sessionID + `","type":"queue-operation","timestamp":"2026-02-01T10:00:00.000Z","agentId":"testlongdesc","message":"` + longDesc + `","isSidechain":false}
`
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0644); err != nil {
		t.Fatalf("Failed to create session file: %v", err)
	}

	agentFile := filepath.Join(subagentsDir, "agent-testlongdesc.jsonl")
	agentContent := `{"uuid":"agent-1","sessionId":"` + sessionID + `","agentId":"testlongdesc","type":"user","timestamp":"2026-02-01T10:00:05.000Z","message":"agent prompt","isSidechain":true}
`
	if err := os.WriteFile(agentFile, []byte(agentContent), 0644); err != nil {
		t.Fatalf("Failed to create agent file: %v", err)
	}

	matches, err := findMatchingAgentIDs(projectDir, sessionID, "test")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(matches) != 1 {
		t.Fatalf("Expected 1 match, got %d", len(matches))
	}

	desc := matches[0].Description
	if len(desc) != 63 { // 60 chars + "..."
		t.Errorf("Expected truncated description length 63, got %d: %s", len(desc), desc)
	}
	if !strings.HasSuffix(desc, "...") {
		t.Errorf("Expected description to end with '...', got: %s", desc)
	}
}

func TestFormatSessionAmbiguityError(t *testing.T) {
	matches := []SessionMatch{
		{
			ID:          "cd2e9388-3108-40e5-b41b-79497cbb58b4",
			Path:        "/path/to/session1.jsonl",
			FirstPrompt: "first prompt here",
			Timestamp:   time.Date(2026, 2, 2, 1, 50, 37, 0, time.UTC),
		},
		{
			ID:          "cd2e4f21-9a14-4b29-8d3c-f5e8a9c1d7e2",
			Path:        "/path/to/session2.jsonl",
			FirstPrompt: "second prompt here",
			Timestamp:   time.Date(2026, 1, 30, 14, 23, 11, 0, time.UTC),
		},
	}

	err := formatSessionAmbiguityError("cd2e", matches)
	errMsg := err.Error()

	// Verify all components are present
	requiredStrings := []string{
		"ambiguous session ID prefix",
		`"cd2e"`,
		"matches 2 sessions",
		"cd2e9388-3108-40e5-b41b-79497cbb58b4",
		"cd2e4f21-9a14-4b29-8d3c-f5e8a9c1d7e2",
		"Date: 2026-02-02T01:50:37Z",
		"Date: 2026-01-30T14:23:11Z",
		"Prompt: first prompt here",
		"Prompt: second prompt here",
		"Please provide more characters",
	}

	for _, required := range requiredStrings {
		if !strings.Contains(errMsg, required) {
			t.Errorf("Error message should contain '%s', got:\n%s", required, errMsg)
		}
	}
}

func TestFormatAgentAmbiguityError(t *testing.T) {
	matches := []AgentMatch{
		{
			ID:          "a12eb64",
			Path:        "/path/to/agent1.jsonl",
			Description: "Explore codebase",
			Timestamp:   time.Date(2026, 2, 1, 10, 0, 5, 0, time.UTC),
		},
		{
			ID:          "a12ef99",
			Path:        "/path/to/agent2.jsonl",
			Description: "Analyze dependencies",
			Timestamp:   time.Date(2026, 2, 1, 10, 0, 10, 0, time.UTC),
		},
	}

	err := formatAgentAmbiguityError("a12e", matches)
	errMsg := err.Error()

	// Verify all components are present
	requiredStrings := []string{
		"ambiguous agent ID prefix",
		`"a12e"`,
		"matches 2 agents",
		"a12eb64",
		"a12ef99",
		"Date: 2026-02-01T10:00:05Z",
		"Date: 2026-02-01T10:00:10Z",
		"Description: Explore codebase",
		"Description: Analyze dependencies",
		"Please provide more characters",
	}

	for _, required := range requiredStrings {
		if !strings.Contains(errMsg, required) {
			t.Errorf("Error message should contain '%s', got:\n%s", required, errMsg)
		}
	}
}
