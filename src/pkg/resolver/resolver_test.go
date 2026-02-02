package resolver

import (
	"strings"
	"testing"
	"time"
)

// TestResolveSessionID_Basic tests basic session ID resolution.
func TestResolveSessionID_Basic(t *testing.T) {
	projectDir, sessions := createTestProjectStructure(t)

	tests := []struct {
		name    string
		prefix  string
		wantErr bool
	}{
		{"unique prefix aaa", "aaa", false},
		{"unique prefix bbb", "bbb", false},
		{"unique prefix ccc", "ccc", false},
		{"non-existent prefix", "zzz", true},
		{"empty prefix", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved, err := ResolveSessionID(projectDir, tt.prefix)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got sessionID: %s", resolved)
				}
			} else {
				assertNoError(t, err)
				if !strings.HasPrefix(resolved, tt.prefix) {
					t.Errorf("resolved ID %q does not start with prefix %q", resolved, tt.prefix)
				}
			}
		})
	}

	_ = sessions // Use the variable to avoid compiler warning
}

// TestResolveSessionID_FullID tests that full session IDs are returned as-is.
func TestResolveSessionID_FullID(t *testing.T) {
	projectDir := t.TempDir()

	sessionID := "full1234-5678-90ab-cdef-123456789012"
	createTestSession(t, projectDir, sessionID, "Test", time.Now())

	resolved, err := ResolveSessionID(projectDir, sessionID)
	assertNoError(t, err)
	assertSessionIDResolved(t, resolved, sessionID)
}

// TestResolveSessionID_Ambiguous tests handling of ambiguous prefixes.
func TestResolveSessionID_Ambiguous(t *testing.T) {
	projectDir := t.TempDir()

	// Create two sessions that both start with "ambig"
	sessionID1 := "ambigaaa-1111-1111-1111-111111111111"
	sessionID2 := "ambigbbb-2222-2222-2222-222222222222"
	createTestSession(t, projectDir, sessionID1, "First", time.Now())
	createTestSession(t, projectDir, sessionID2, "Second", time.Now().Add(-1*time.Hour))

	_, err := ResolveSessionID(projectDir, "ambig")
	if err == nil {
		t.Fatal("expected ambiguous prefix error, got nil")
	}

	// Check error message contains both IDs
	errMsg := err.Error()
	if !strings.Contains(errMsg, sessionID1) {
		t.Errorf("error message missing session ID %q: %s", sessionID1, errMsg)
	}
	if !strings.Contains(errMsg, sessionID2) {
		t.Errorf("error message missing session ID %q: %s", sessionID2, errMsg)
	}
}

// TestResolveAgentID_Basic tests basic agent ID resolution.
func TestResolveAgentID_Basic(t *testing.T) {
	projectDir, sessions := createTestProjectStructure(t)

	// Get session ID
	sessionID := "aaa12345-1234-1234-1234-123456789abc"

	// Create agents
	sessionDir := projectDir + "/" + sessionID
	createTestAgent(t, sessionDir, "agent123", "Test agent 1")
	createTestAgent(t, sessionDir, "agent456", "Test agent 2")

	tests := []struct {
		name    string
		prefix  string
		want    string
		wantErr bool
	}{
		{"prefix agent1", "agent1", "agent123", false},
		{"prefix agent4", "agent4", "agent456", false},
		{"non-existent", "zzz", "", true},
		{"empty prefix", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved, err := ResolveAgentID(projectDir, sessionID, tt.prefix)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got agentID: %s", resolved)
				}
			} else {
				assertNoError(t, err)
				if resolved != tt.want {
					t.Errorf("ResolveAgentID() = %q, want %q", resolved, tt.want)
				}
			}
		})
	}

	_ = sessions // Use the variable
}

// TestPrefixError_Error tests PrefixError error message formatting.
func TestPrefixError_Error(t *testing.T) {
	err := &PrefixError{
		Prefix:  "test",
		Message: "no sessions found",
	}

	want := "no sessions found: test"
	if got := err.Error(); got != want {
		t.Errorf("PrefixError.Error() = %q, want %q", got, want)
	}
}

// TestAmbiguousPrefixError_Error tests AmbiguousPrefixError formatting.
func TestAmbiguousPrefixError_Error(t *testing.T) {
	err := &AmbiguousPrefixError{
		Prefix:   "abc",
		Matches:  []string{"abc123", "abc456"},
		ItemType: "session",
	}

	errMsg := err.Error()

	// Check required elements
	required := []string{"ambiguous", "abc", "abc123", "abc456", "provide more characters"}
	for _, substr := range required {
		if !strings.Contains(errMsg, substr) {
			t.Errorf("error message missing %q: %s", substr, errMsg)
		}
	}
}

// TestResolveSessionID_MultipleMatches tests behavior with multiple matching prefixes.
func TestResolveSessionID_MultipleMatches(t *testing.T) {
	projectDir := t.TempDir()

	// Create 5 sessions that all start with "multi"
	// Using valid UUID format: 8-4-4-4-12
	var sessionIDs []string
	baseTime := time.Now()
	for i := 0; i < 5; i++ {
		// Create valid UUIDs that all share "multi" prefix (first 5 chars)
		sessionID := "multi" + string(rune('a'+i)) + "11-2222-3333-4444-555555555555"
		sessionIDs = append(sessionIDs, sessionID)
		createTestSession(t, projectDir, sessionID, "Test "+string(rune('A'+i)), baseTime.Add(time.Duration(-i)*time.Hour))
	}

	_, err := ResolveSessionID(projectDir, "multi")
	assertError(t, err, "ambiguous")

	// Verify all IDs are in error message
	errMsg := err.Error()
	for _, id := range sessionIDs {
		if !strings.Contains(errMsg, id) {
			t.Errorf("error missing session ID %q", id)
		}
	}
}

// TestResolveSessionID_ShortestPrefix tests resolution with minimal prefix.
func TestResolveSessionID_ShortestPrefix(t *testing.T) {
	projectDir := t.TempDir()

	// Create sessions starting with different first characters
	createTestSession(t, projectDir, "a0000000-0000-0000-0000-000000000000", "A", time.Now())
	createTestSession(t, projectDir, "b0000000-0000-0000-0000-000000000000", "B", time.Now())

	tests := []struct {
		prefix string
		want   string
	}{
		{"a", "a0000000-0000-0000-0000-000000000000"},
		{"b", "b0000000-0000-0000-0000-000000000000"},
	}

	for _, tt := range tests {
		t.Run("prefix_"+tt.prefix, func(t *testing.T) {
			resolved, err := ResolveSessionID(projectDir, tt.prefix)
			assertNoError(t, err)
			assertSessionIDResolved(t, resolved, tt.want)
		})
	}
}

// TestResolveAgentID_NoSubagentsDir tests resolution when subagents dir doesn't exist.
func TestResolveAgentID_NoSubagentsDir(t *testing.T) {
	projectDir := t.TempDir()

	sessionID := "nosub123-1234-1234-1234-123456789abc"
	createTestSession(t, projectDir, sessionID, "Test", time.Now())

	// Don't create subagents directory - should return empty result (no agents found)

	_, err := ResolveAgentID(projectDir, sessionID, "agent")
	if err == nil {
		t.Error("expected error when no agents exist")
	}
	// Error could be "no agents found" or path error, both are acceptable
}

// TestResolveSessionID_SpecialCharactersInPrefix tests handling of dashes in session IDs.
func TestResolveSessionID_SpecialCharactersInPrefix(t *testing.T) {
	projectDir := t.TempDir()

	// Use valid UUID format with dashes
	sessionID := "special1-ch4r-ac7e-rs12-123456789abc"
	createTestSession(t, projectDir, sessionID, "Special", time.Now())

	tests := []struct {
		prefix string
	}{
		{"spec"},
		{"special"},
		{"special1"},
		{"special1-"},
		{"special1-ch4r"},
	}

	for _, tt := range tests {
		t.Run("prefix_"+tt.prefix, func(t *testing.T) {
			resolved, err := ResolveSessionID(projectDir, tt.prefix)
			assertNoError(t, err)
			assertSessionIDResolved(t, resolved, sessionID)
		})
	}
}

// TestResolveSessionID_PrefixLength tests various prefix lengths.
func TestResolveSessionID_PrefixLength(t *testing.T) {
	projectDir := t.TempDir()

	sessionID := "abcdef12-3456-7890-abcd-ef0123456789"
	createTestSession(t, projectDir, sessionID, "Length test", time.Now())

	prefixes := []string{
		"a",        // 1 char
		"ab",       // 2 chars
		"abc",      // 3 chars
		"abcdef",   // 6 chars
		"abcdef12", // 8 chars (first segment)
		sessionID,  // Full ID
	}

	for _, prefix := range prefixes {
		t.Run("len_"+prefix, func(t *testing.T) {
			resolved, err := ResolveSessionID(projectDir, prefix)
			assertNoError(t, err)
			assertSessionIDResolved(t, resolved, sessionID)
		})
	}
}

// TestResolveAgentID_MultipleAgents tests resolution with many agents.
func TestResolveAgentID_MultipleAgents(t *testing.T) {
	projectDir, _ := createTestProjectStructure(t)

	sessionID := "multi123-1234-1234-1234-123456789abc"
	createTestSession(t, projectDir, sessionID, "Multi agent", time.Now())

	sessionDir := projectDir + "/" + sessionID

	// Create 10 agents with different prefixes
	agents := []string{
		"agent-aaa", "agent-bbb", "agent-ccc",
		"agent-ddd", "agent-eee", "agent-fff",
		"agent-ggg", "agent-hhh", "agent-iii", "agent-jjj",
	}

	for _, agentID := range agents {
		createTestAgent(t, sessionDir, agentID, "Agent: "+agentID)
	}

	// Test each can be resolved uniquely
	for _, expected := range agents {
		prefix := expected[:8] // "agent-a", "agent-b", etc
		t.Run("resolve_"+prefix, func(t *testing.T) {
			resolved, err := ResolveAgentID(projectDir, sessionID, prefix)
			assertNoError(t, err)
			if resolved != expected {
				t.Errorf("ResolveAgentID(%q) = %q, want %q", prefix, resolved, expected)
			}
		})
	}
}
