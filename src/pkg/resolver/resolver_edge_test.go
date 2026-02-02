package resolver

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestResolveSessionID_EmptyPrefix tests that an empty prefix returns an error.
func TestResolveSessionID_EmptyPrefix(t *testing.T) {
	projectDir, _ := createTestProjectStructure(t)

	_, err := ResolveSessionID(projectDir, "")
	assertError(t, err, "empty")
}

// TestResolveSessionID_NonExistentPrefix tests that a non-matching prefix returns an error.
func TestResolveSessionID_NonExistentPrefix(t *testing.T) {
	projectDir, _ := createTestProjectStructure(t)

	_, err := ResolveSessionID(projectDir, "zzz")
	assertError(t, err, "no sessions found")
}

// TestResolveSessionID_FullIDPassthrough tests that a full UUID is returned as-is.
func TestResolveSessionID_FullIDPassthrough(t *testing.T) {
	projectDir, sessions := createTestProjectStructure(t)

	// Get full session ID from test data
	fullID := filepath.Base(sessions["aaa"])
	fullID = strings.TrimSuffix(fullID, ".jsonl")

	resolved, err := ResolveSessionID(projectDir, fullID)
	assertNoError(t, err)
	assertSessionIDResolved(t, resolved, fullID)
}

// TestResolveSessionID_VeryShortPrefix tests 1-2 character prefixes.
func TestResolveSessionID_VeryShortPrefix(t *testing.T) {
	tests := []struct {
		name    string
		prefix  string
		wantErr bool
	}{
		{
			name:    "one character unique",
			prefix:  "a",
			wantErr: false, // Should resolve to "aaa..." session
		},
		{
			name:    "two characters unique",
			prefix:  "aa",
			wantErr: false,
		},
		{
			name:    "one character ambiguous",
			prefix:  "c",
			wantErr: false, // Only one session starts with 'c'
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectDir, _ := createTestProjectStructure(t)

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
}

// TestResolveSessionID_CaseSensitivity tests that prefix matching is case-sensitive.
func TestResolveSessionID_CaseSensitivity(t *testing.T) {
	projectDir, _ := createTestProjectStructure(t)

	// All test sessions use lowercase, so uppercase should not match
	_, err := ResolveSessionID(projectDir, "AAA")
	assertError(t, err, "no sessions found")

	// Lowercase should work
	resolved, err := ResolveSessionID(projectDir, "aaa")
	assertNoError(t, err)
	if !strings.HasPrefix(resolved, "aaa") {
		t.Errorf("resolved ID %q does not start with 'aaa'", resolved)
	}
}

// TestResolveSessionID_SpecialCharacters tests UUIDs with dashes.
func TestResolveSessionID_SpecialCharacters(t *testing.T) {
	projectDir := t.TempDir()

	// Create session with dashes at specific positions
	sessionID := "cd2e9388-3108-40e5-b41b-79497cbb58b4"
	createTestSession(t, projectDir, sessionID, "Test prompt", time.Now())

	tests := []struct {
		name   string
		prefix string
		want   string
	}{
		{
			name:   "prefix before dash",
			prefix: "cd2e",
			want:   sessionID,
		},
		{
			name:   "prefix including dash",
			prefix: "cd2e9388-",
			want:   sessionID,
		},
		{
			name:   "prefix spanning dash",
			prefix: "cd2e9388-3108",
			want:   sessionID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved, err := ResolveSessionID(projectDir, tt.prefix)
			assertNoError(t, err)
			assertSessionIDResolved(t, resolved, tt.want)
		})
	}
}

// TestResolveSessionID_PrefixAtHyphenBoundary tests prefixes ending at dash positions.
func TestResolveSessionID_PrefixAtHyphenBoundary(t *testing.T) {
	projectDir := t.TempDir()

	// UUID format: 8-4-4-4-12
	sessionID := "12345678-1234-5678-9abc-def012345678"
	createTestSession(t, projectDir, sessionID, "Test", time.Now())

	tests := []struct {
		prefix string
	}{
		{"12345678-"},       // After first dash
		{"12345678-1234-"},  // After second dash
		{"12345678-1234-5"}, // Mid-segment after dash
	}

	for _, tt := range tests {
		t.Run("prefix_"+tt.prefix, func(t *testing.T) {
			resolved, err := ResolveSessionID(projectDir, tt.prefix)
			assertNoError(t, err)
			assertSessionIDResolved(t, resolved, sessionID)
		})
	}
}

// TestResolveSessionID_EmptySession tests resolution with empty session file.
func TestResolveSessionID_EmptySession(t *testing.T) {
	projectDir := t.TempDir()

	sessionID := "empty123-1234-1234-1234-123456789abc"
	createEmptySession(t, projectDir, sessionID)

	// Should still resolve even if empty
	resolved, err := ResolveSessionID(projectDir, "empty")
	assertNoError(t, err)
	assertSessionIDResolved(t, resolved, sessionID)
}

// TestResolveSessionID_SessionWithNoUserMessages tests session with only system entries.
func TestResolveSessionID_SessionWithNoUserMessages(t *testing.T) {
	projectDir := t.TempDir()

	sessionID := "system12-1234-1234-1234-123456789abc"
	createSessionWithNoUserMessages(t, projectDir, sessionID)

	// Should still resolve
	resolved, err := ResolveSessionID(projectDir, "system")
	assertNoError(t, err)
	assertSessionIDResolved(t, resolved, sessionID)
}

// TestResolveSessionID_MalformedJSON tests handling of malformed session files.
func TestResolveSessionID_MalformedJSON(t *testing.T) {
	projectDir := t.TempDir()

	sessionID := "malform1-1234-1234-1234-123456789abc"
	createMalformedSession(t, projectDir, sessionID)

	// Should still find the session file by name, even if content is bad
	resolved, err := ResolveSessionID(projectDir, "malform")
	assertNoError(t, err)
	assertSessionIDResolved(t, resolved, sessionID)
}

// TestResolveSessionID_AmbiguousPrefixErrorFormat tests the error message format for ambiguous prefixes.
func TestResolveSessionID_AmbiguousPrefixErrorFormat(t *testing.T) {
	projectDir := t.TempDir()

	// Create two sessions with same prefix
	sessionIDs := createAmbiguousSessions(t, projectDir, "cd2e", 2)

	_, err := ResolveSessionID(projectDir, "cd2e")
	if err == nil {
		t.Fatal("expected ambiguous prefix error, got nil")
	}

	errMsg := err.Error()

	// Check error message contains expected elements
	requiredSubstrings := []string{
		"ambiguous",
		"provide more characters",
		sessionIDs[0], // First full ID
		sessionIDs[1], // Second full ID
	}

	for _, substr := range requiredSubstrings {
		if !strings.Contains(errMsg, substr) {
			t.Errorf("error message missing %q: %s", substr, errMsg)
		}
	}
}

// TestResolveSessionID_NonExistentDirectory tests resolution in non-existent directory.
func TestResolveSessionID_NonExistentDirectory(t *testing.T) {
	projectDir := filepath.Join(t.TempDir(), "nonexistent")

	_, err := ResolveSessionID(projectDir, "test")
	assertError(t, err, "")
}

// TestResolveAgentID_EmptyPrefix tests that empty agent prefix returns error.
func TestResolveAgentID_EmptyPrefix(t *testing.T) {
	projectDir, sessions := createTestProjectStructure(t)

	sessionID := filepath.Base(sessions["aaa"])
	sessionID = strings.TrimSuffix(sessionID, ".jsonl")

	_, err := ResolveAgentID(projectDir, sessionID, "")
	assertError(t, err, "empty")
}

// TestResolveAgentID_NonExistentPrefix tests non-matching agent prefix.
func TestResolveAgentID_NonExistentPrefix(t *testing.T) {
	projectDir, sessions := createTestProjectStructure(t)

	sessionID := filepath.Base(sessions["aaa"])
	sessionID = strings.TrimSuffix(sessionID, ".jsonl")

	// Create an agent
	sessionDir := filepath.Join(projectDir, sessionID)
	createTestAgent(t, sessionDir, "agent123", "Test agent")

	// Try non-existent prefix
	_, err := ResolveAgentID(projectDir, sessionID, "zzz")
	assertError(t, err, "no agents found")
}

// TestResolveAgentID_FullIDPassthrough tests full agent ID pass-through.
func TestResolveAgentID_FullIDPassthrough(t *testing.T) {
	projectDir, sessions := createTestProjectStructure(t)

	sessionID := filepath.Base(sessions["aaa"])
	sessionID = strings.TrimSuffix(sessionID, ".jsonl")

	// Create agent with specific ID
	sessionDir := filepath.Join(projectDir, sessionID)
	fullAgentID := "agent12345678"
	createTestAgent(t, sessionDir, fullAgentID, "Test agent")

	resolved, err := ResolveAgentID(projectDir, sessionID, fullAgentID)
	assertNoError(t, err)
	if resolved != fullAgentID {
		t.Errorf("ResolveAgentID() = %q, want %q", resolved, fullAgentID)
	}
}

// TestResolveAgentID_AmbiguousPrefix tests ambiguous agent prefix error.
func TestResolveAgentID_AmbiguousPrefix(t *testing.T) {
	projectDir, sessions := createTestProjectStructure(t)

	sessionID := filepath.Base(sessions["aaa"])
	sessionID = strings.TrimSuffix(sessionID, ".jsonl")

	sessionDir := filepath.Join(projectDir, sessionID)

	// Create multiple agents with same prefix
	createTestAgent(t, sessionDir, "agent123-aaa", "Agent A")
	createTestAgent(t, sessionDir, "agent123-bbb", "Agent B")

	_, err := ResolveAgentID(projectDir, sessionID, "agent123")
	assertError(t, err, "ambiguous")
}

// TestResolveAgentID_CaseSensitive tests case-sensitive agent matching.
func TestResolveAgentID_CaseSensitive(t *testing.T) {
	projectDir, sessions := createTestProjectStructure(t)

	sessionID := filepath.Base(sessions["aaa"])
	sessionID = strings.TrimSuffix(sessionID, ".jsonl")

	sessionDir := filepath.Join(projectDir, sessionID)
	createTestAgent(t, sessionDir, "agentabc", "Test agent")

	// Uppercase should not match
	_, err := ResolveAgentID(projectDir, sessionID, "AGENT")
	assertError(t, err, "no agents found")

	// Lowercase should work
	resolved, err := ResolveAgentID(projectDir, sessionID, "agent")
	assertNoError(t, err)
	if !strings.HasPrefix(resolved, "agent") {
		t.Errorf("resolved agent %q does not start with 'agent'", resolved)
	}
}

// TestResolveAgentID_SessionNotFound tests resolution when session doesn't exist.
func TestResolveAgentID_SessionNotFound(t *testing.T) {
	projectDir := t.TempDir()

	_, err := ResolveAgentID(projectDir, "nonexistent-session-id", "agent")
	if err == nil {
		t.Error("expected error for non-existent session, got nil")
	}
}

// TestResolveBoth_SessionAndAgent tests resolving both session and agent prefixes.
func TestResolveBoth_SessionAndAgent(t *testing.T) {
	projectDir, _ := createTestProjectStructure(t)

	// Create a session with known prefix
	sessionID := "test1234-1234-1234-1234-123456789abc"
	createTestSession(t, projectDir, sessionID, "Test", time.Now())

	// Create agent in that session
	sessionDir := filepath.Join(projectDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create session dir: %v", err)
	}
	createTestAgent(t, sessionDir, "agent9999", "Test agent")

	// Resolve session prefix
	resolvedSession, err := ResolveSessionID(projectDir, "test")
	assertNoError(t, err)
	assertSessionIDResolved(t, resolvedSession, sessionID)

	// Resolve agent prefix
	resolvedAgent, err := ResolveAgentID(projectDir, resolvedSession, "agent9")
	assertNoError(t, err)
	if resolvedAgent != "agent9999" {
		t.Errorf("ResolveAgentID() = %q, want %q", resolvedAgent, "agent9999")
	}
}

// TestResolveSessionID_MixedLineEndings tests handling of different line endings.
func TestResolveSessionID_MixedLineEndings(t *testing.T) {
	projectDir := t.TempDir()

	sessionID := "mixed123-1234-1234-1234-123456789abc"
	filePath := filepath.Join(projectDir, sessionID+".jsonl")

	// Create file with Windows line endings
	content := `{"type":"user","sessionId":"` + sessionID + `","timestamp":1234567890,"content":"test"}` + "\r\n"
	content += `{"type":"assistant","sessionId":"` + sessionID + `","timestamp":1234567891,"content":"response"}` + "\r\n"

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write session file: %v", err)
	}

	// Should resolve correctly despite \r\n endings
	resolved, err := ResolveSessionID(projectDir, "mixed")
	assertNoError(t, err)
	assertSessionIDResolved(t, resolved, sessionID)
}
