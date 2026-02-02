package resolver

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/randlee/claude-history/pkg/models"
)

// createTestSession creates a mock session JSONL file with test data.
// Returns the full path to the created session file.
//
// Note: This creates a minimal session without agent spawning. For sessions
// with agents, use createTestSessionWithAgentSpawn or manually add spawn entries.
func createTestSession(t *testing.T, dir, sessionID, firstPrompt string, timestamp time.Time) string {
	t.Helper()

	filePath := filepath.Join(dir, sessionID+".jsonl")

	// Create user message entry
	userEntry := models.ConversationEntry{
		Type:      models.EntryTypeUser,
		SessionID: sessionID,
		Timestamp: timestamp.Format(time.RFC3339Nano),
		Message:   json.RawMessage(`{"role":"user","content":"` + firstPrompt + `"}`),
	}

	// Create assistant response entry
	assistantEntry := models.ConversationEntry{
		Type:      models.EntryTypeAssistant,
		SessionID: sessionID,
		Timestamp: timestamp.Add(1 * time.Second).Format(time.RFC3339Nano),
		Message:   json.RawMessage(`{"role":"assistant","content":[{"type":"text","text":"Response"}]}`),
	}

	// Write entries to file
	f, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("failed to create session file: %v", err)
	}
	defer func() { _ = f.Close() }()

	encoder := json.NewEncoder(f)
	if err := encoder.Encode(userEntry); err != nil {
		t.Fatalf("failed to write user entry: %v", err)
	}
	if err := encoder.Encode(assistantEntry); err != nil {
		t.Fatalf("failed to write assistant entry: %v", err)
	}

	return filePath
}

// createTestAgent creates a mock agent JSONL file.
// Returns the full path to the created agent file.
//
// The agent file entry includes:
// - agentId: The agent's identifier (matches real Claude Code format)
// - parentUuid: null (indicating this is the first entry in the agent)
// - sessionId: Extracted from the session directory path
//
// Note: In real Claude Code, agent files contain entries with agentId field set,
// and the first entry typically has parentUuid: null.
func createTestAgent(t *testing.T, sessionDir, agentID, description string) string {
	t.Helper()

	subagentsDir := filepath.Join(sessionDir, "subagents")
	if err := os.MkdirAll(subagentsDir, 0755); err != nil {
		t.Fatalf("failed to create subagents dir: %v", err)
	}

	filePath := filepath.Join(subagentsDir, "agent-"+agentID+".jsonl")

	// Extract session ID from directory name (last 36 characters is UUID)
	sessionID := sessionDir[len(sessionDir)-36:]

	// Create agent entry with proper agentId field (matching real Claude Code format)
	// In real Claude Code, agent file entries have:
	// - agentId set to the agent's ID
	// - parentUuid: null for the first entry
	// - sessionId matching the parent session
	agentEntry := models.ConversationEntry{
		Type:      models.EntryTypeUser,
		SessionID: sessionID,
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
	if err := encoder.Encode(agentEntry); err != nil {
		t.Fatalf("failed to write agent entry: %v", err)
	}

	return filePath
}

// createAmbiguousSessions creates N sessions with the same prefix.
// Returns slice of full session IDs.
func createAmbiguousSessions(t *testing.T, dir, prefix string, count int) []string {
	t.Helper()

	var sessionIDs []string
	baseTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	for i := 0; i < count; i++ {
		// Generate unique session ID with same prefix
		sessionID := prefix + generateUniqueUUIDSuffix(i)
		sessionIDs = append(sessionIDs, sessionID)

		// Create session with unique timestamp and prompt
		timestamp := baseTime.Add(time.Duration(i) * time.Hour)
		prompt := "Test prompt " + string(rune('A'+i))
		createTestSession(t, dir, sessionID, prompt, timestamp)
	}

	return sessionIDs
}

// generateUniqueUUIDSuffix generates a unique UUID suffix for testing.
// Format: prefix + this_suffix to create valid UUID format.
// For a 4-char prefix, we need 4 more chars for first group (8 total), then 4-4-4-12.
func generateUniqueUUIDSuffix(index int) string {
	// Generate hex string from index
	chars := "0123456789abcdef"

	// Create a base 16 representation with leading zeros
	hex := ""
	num := index
	for i := 0; i < 8; i++ {
		hex = string(chars[num%16]) + hex
		num = num / 16
	}

	// Format as UUID suffix: 4-4-4-4-12
	// If prefix is 4 chars (like "cd2e"), we need 4 more for first group
	// Then 4-4-4-12 for the remaining groups
	return hex[0:4] + "-" + hex[4:8] + "-1111-1111-111111111111"
}

// createTestProjectStructure creates a complete test project with sessions and agents.
func createTestProjectStructure(t *testing.T) (string, map[string]string) {
	t.Helper()

	// Create temp directory for test project
	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, "test-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}

	sessions := make(map[string]string)

	// Create test sessions with various prefixes
	now := time.Now()
	sessions["aaa"] = createTestSession(t, projectDir, "aaa12345-1234-1234-1234-123456789abc", "First test prompt", now)
	sessions["bbb"] = createTestSession(t, projectDir, "bbb12345-1234-1234-1234-123456789abc", "Second test prompt", now.Add(-1*time.Hour))
	sessions["ccc"] = createTestSession(t, projectDir, "ccc12345-1234-1234-1234-123456789abc", "Third test prompt", now.Add(-2*time.Hour))

	// Create session directory for agent testing
	sessionID := "aaa12345-1234-1234-1234-123456789abc"
	sessionDir := filepath.Join(projectDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create session dir: %v", err)
	}

	return projectDir, sessions
}

// createEmptySession creates a session file with no entries.
func createEmptySession(t *testing.T, dir, sessionID string) string {
	t.Helper()

	filePath := filepath.Join(dir, sessionID+".jsonl")
	f, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("failed to create empty session file: %v", err)
	}
	_ = f.Close()

	return filePath
}

// createMalformedSession creates a session file with invalid JSON.
func createMalformedSession(t *testing.T, dir, sessionID string) string {
	t.Helper()

	filePath := filepath.Join(dir, sessionID+".jsonl")
	if err := os.WriteFile(filePath, []byte("{invalid json\n{more invalid"), 0644); err != nil {
		t.Fatalf("failed to create malformed session: %v", err)
	}

	return filePath
}

// createSessionWithNoUserMessages creates a session with only system entries.
func createSessionWithNoUserMessages(t *testing.T, dir, sessionID string) string {
	t.Helper()

	filePath := filepath.Join(dir, sessionID+".jsonl")

	systemEntry := models.ConversationEntry{
		Type:      models.EntryTypeSystem,
		SessionID: sessionID,
		Timestamp: time.Now().Format(time.RFC3339Nano),
		Message:   json.RawMessage(`{"type":"system","content":"System event"}`),
	}

	f, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("failed to create session file: %v", err)
	}
	defer func() { _ = f.Close() }()

	encoder := json.NewEncoder(f)
	if err := encoder.Encode(systemEntry); err != nil {
		t.Fatalf("failed to write system entry: %v", err)
	}

	return filePath
}

// createLargeTestProject creates a project with many sessions for performance testing.
func createLargeTestProject(tb testing.TB, sessionCount int) string {
	tb.Helper()

	tempDir := tb.TempDir()
	projectDir := filepath.Join(tempDir, "large-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		tb.Fatalf("failed to create project dir: %v", err)
	}

	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	for i := 0; i < sessionCount; i++ {
		sessionID := generateRandomUUID(i)
		timestamp := baseTime.Add(time.Duration(i) * time.Minute)
		prompt := "Test prompt for performance testing"
		// For large project creation, bypass the t.Helper check
		createTestSessionDirect(tb, projectDir, sessionID, prompt, timestamp)
	}

	return projectDir
}

// createTestSessionDirect creates a session without requiring *testing.T specifically.
func createTestSessionDirect(tb testing.TB, dir, sessionID, firstPrompt string, timestamp time.Time) string {
	tb.Helper()

	filePath := filepath.Join(dir, sessionID+".jsonl")

	// Create user message entry
	userEntry := models.ConversationEntry{
		Type:      models.EntryTypeUser,
		SessionID: sessionID,
		Timestamp: timestamp.Format(time.RFC3339Nano),
		Message:   json.RawMessage(`{"role":"user","content":"` + firstPrompt + `"}`),
	}

	// Create assistant response entry
	assistantEntry := models.ConversationEntry{
		Type:      models.EntryTypeAssistant,
		SessionID: sessionID,
		Timestamp: timestamp.Add(1 * time.Second).Format(time.RFC3339Nano),
		Message:   json.RawMessage(`{"role":"assistant","content":[{"type":"text","text":"Response"}]}`),
	}

	// Write entries to file
	f, err := os.Create(filePath)
	if err != nil {
		tb.Fatalf("failed to create session file: %v", err)
	}
	defer func() { _ = f.Close() }()

	encoder := json.NewEncoder(f)
	if err := encoder.Encode(userEntry); err != nil {
		tb.Fatalf("failed to write user entry: %v", err)
	}
	if err := encoder.Encode(assistantEntry); err != nil {
		tb.Fatalf("failed to write assistant entry: %v", err)
	}

	return filePath
}

// generateRandomUUID generates a pseudo-random UUID for testing.
// Each index produces a unique UUID.
func generateRandomUUID(seed int) string {
	// Convert seed to 8-digit hex (32 bits)
	chars := "0123456789abcdef"
	hex1 := make([]byte, 8)
	hex2 := make([]byte, 4)
	hex3 := make([]byte, 4)
	hex4 := make([]byte, 4)
	hex5 := make([]byte, 12)

	// First part: encode seed
	n := seed
	for i := 7; i >= 0; i-- {
		hex1[i] = chars[n%16]
		n = n / 16
	}

	// Remaining parts: fill with 1s (deterministic)
	for i := range hex2 {
		hex2[i] = '1'
	}
	for i := range hex3 {
		hex3[i] = '1'
	}
	for i := range hex4 {
		hex4[i] = '1'
	}
	for i := range hex5 {
		hex5[i] = '1'
	}

	// Format as UUID: 8-4-4-4-12
	return string(hex1) + "-" + string(hex2) + "-" + string(hex3) + "-" + string(hex4) + "-" + string(hex5)
}

// createDeeplyNestedAgents creates agents with deep nesting for testing.
func createDeeplyNestedAgents(t *testing.T, projectDir, sessionID string, depth int) {
	t.Helper()

	sessionDir := filepath.Join(projectDir, sessionID)
	currentDir := sessionDir

	for i := 0; i < depth; i++ {
		agentID := "agent" + string(rune('A'+i))

		// Create agent file
		subagentsDir := filepath.Join(currentDir, "subagents")
		if err := os.MkdirAll(subagentsDir, 0755); err != nil {
			t.Fatalf("failed to create subagents dir: %v", err)
		}

		filePath := filepath.Join(subagentsDir, "agent-"+agentID+".jsonl")
		agentEntry := models.ConversationEntry{
			Type:      models.EntryTypeUser,
			SessionID: sessionID,
			AgentID:   agentID,
			Timestamp: time.Now().Format(time.RFC3339Nano),
			Message:   json.RawMessage(`{"role":"user","content":"Agent at depth ` + string(rune('0'+i)) + `"}`),
		}

		f, err := os.Create(filePath)
		if err != nil {
			t.Fatalf("failed to create agent file: %v", err)
		}

		encoder := json.NewEncoder(f)
		if err := encoder.Encode(agentEntry); err != nil {
			_ = f.Close()
			t.Fatalf("failed to write agent entry: %v", err)
		}
		_ = f.Close()

		// Move to next level
		currentDir = filepath.Join(currentDir, "subagents", "agent-"+agentID)
		if err := os.MkdirAll(currentDir, 0755); err != nil {
			t.Fatalf("failed to create nested dir: %v", err)
		}
	}
}

// assertSessionIDResolved checks that a session ID was resolved correctly.
func assertSessionIDResolved(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("ResolveSessionID() = %q, want %q", got, want)
	}
}

// assertError checks that an error occurred with the expected message substring.
func assertError(t *testing.T, err error, wantSubstring string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error containing %q, got nil", wantSubstring)
	}
	if wantSubstring != "" && !contains(err.Error(), wantSubstring) {
		t.Errorf("error = %q, want substring %q", err.Error(), wantSubstring)
	}
}

// assertNoError checks that no error occurred.
func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// contains checks if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && (s[0:len(substr)] == substr || contains(s[1:], substr))))
}
