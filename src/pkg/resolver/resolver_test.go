package resolver

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/randlee/claude-history/pkg/encoding"
)

// TestResolveSessionID tests session ID prefix resolution.
func TestResolveSessionID(t *testing.T) {
	// Create a temporary project structure for testing
	tmpDir := t.TempDir()

	// Create a mock Claude projects directory
	projectPath := filepath.Join(tmpDir, "test-project")
	err := os.MkdirAll(projectPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	// Create encoded project directory
	encodedPath := encoding.EncodePath(projectPath)
	claudeProjectDir := filepath.Join(tmpDir, ".claude", "projects", encodedPath)
	err = os.MkdirAll(claudeProjectDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create Claude project dir: %v", err)
	}

	// Create test session files
	sessions := []string{
		"cd2e9388-3108-40e5-b41b-79497cbb58b4",
		"cd2e4f21-9a14-4b29-8d3c-f5e8a9c1d7e2",
		"ab123456-7890-1234-5678-123456789abc",
	}

	for _, sessionID := range sessions {
		sessionFile := filepath.Join(claudeProjectDir, sessionID+".jsonl")
		// Create minimal JSONL content with a user message
		content := `{"type":"user","role":"user","content":"test prompt","timestamp":"2026-02-01T00:00:00Z","uuid":"test-uuid","sessionId":"` + sessionID + `"}` + "\n"
		err := os.WriteFile(sessionFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create session file: %v", err)
		}
	}

	// Override home dir for testing
	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	// Set claudeDir to the temporary .claude directory
	claudeDir := filepath.Join(tmpDir, ".claude")

	tests := []struct {
		name        string
		prefix      string
		wantID      string
		wantErr     bool
		errContains string
	}{
		{
			name:        "empty prefix",
			prefix:      "",
			wantErr:     true,
			errContains: "cannot be empty",
		},
		{
			name:   "unique prefix - first 8 chars",
			prefix: "cd2e9388",
			wantID: "cd2e9388-3108-40e5-b41b-79497cbb58b4",
		},
		{
			name:   "unique prefix - short",
			prefix: "ab123",
			wantID: "ab123456-7890-1234-5678-123456789abc",
		},
		{
			name:        "ambiguous prefix",
			prefix:      "cd2e",
			wantErr:     true,
			errContains: "ambiguous session ID prefix",
		},
		{
			name:        "no match",
			prefix:      "xyz999",
			wantErr:     true,
			errContains: "no session found",
		},
		{
			name:   "full ID",
			prefix: "cd2e9388-3108-40e5-b41b-79497cbb58b4",
			wantID: "cd2e9388-3108-40e5-b41b-79497cbb58b4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveSessionID(claudeDir, projectPath, tt.prefix)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ResolveSessionID() expected error, got nil")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ResolveSessionID() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("ResolveSessionID() unexpected error: %v", err)
				return
			}

			if got != tt.wantID {
				t.Errorf("ResolveSessionID() = %v, want %v", got, tt.wantID)
			}
		})
	}
}

// TestResolveAgentID tests agent ID prefix resolution.
func TestResolveAgentID(t *testing.T) {
	// Create a temporary project structure for testing
	tmpDir := t.TempDir()

	// Create a mock Claude projects directory
	projectPath := filepath.Join(tmpDir, "test-project")
	err := os.MkdirAll(projectPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	// Create encoded project directory
	encodedPath := encoding.EncodePath(projectPath)
	claudeProjectDir := filepath.Join(tmpDir, ".claude", "projects", encodedPath)
	err = os.MkdirAll(claudeProjectDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create Claude project dir: %v", err)
	}

	// Create a test session
	sessionID := "cd2e9388-3108-40e5-b41b-79497cbb58b4"
	sessionFile := filepath.Join(claudeProjectDir, sessionID+".jsonl")
	sessionContent := `{"type":"user","role":"user","content":"test","timestamp":"2026-02-01T00:00:00Z","uuid":"test-uuid","sessionId":"` + sessionID + `"}` + "\n"
	err = os.WriteFile(sessionFile, []byte(sessionContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create session file: %v", err)
	}

	// Create subagents directory
	subagentsDir := filepath.Join(claudeProjectDir, sessionID, "subagents")
	err = os.MkdirAll(subagentsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subagents dir: %v", err)
	}

	// Create test agent files
	agents := []string{
		"a12eb64",
		"a12ef99",
		"a68b8c0",
	}

	for _, agentID := range agents {
		agentFile := filepath.Join(subagentsDir, "agent-"+agentID+".jsonl")
		content := `{"type":"user","role":"user","content":"agent test","timestamp":"2026-02-01T00:00:00Z","uuid":"test-uuid","sessionId":"` + sessionID + `"}` + "\n"
		err := os.WriteFile(agentFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create agent file: %v", err)
		}
	}

	// Override home dir for testing
	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	// Set claudeDir to the temporary .claude directory
	claudeDir := filepath.Join(tmpDir, ".claude")

	tests := []struct {
		name        string
		sessionID   string
		prefix      string
		wantID      string
		wantErr     bool
		errContains string
	}{
		{
			name:        "empty prefix",
			sessionID:   sessionID,
			prefix:      "",
			wantErr:     true,
			errContains: "cannot be empty",
		},
		{
			name:        "empty session ID",
			sessionID:   "",
			prefix:      "a12",
			wantErr:     true,
			errContains: "session ID is required",
		},
		{
			name:      "unique prefix",
			sessionID: sessionID,
			prefix:    "a68",
			wantID:    "a68b8c0",
		},
		{
			name:        "ambiguous prefix",
			sessionID:   sessionID,
			prefix:      "a12",
			wantErr:     true,
			errContains: "ambiguous agent ID prefix",
		},
		{
			name:        "no match",
			sessionID:   sessionID,
			prefix:      "z99",
			wantErr:     true,
			errContains: "no agent found",
		},
		{
			name:      "full ID",
			sessionID: sessionID,
			prefix:    "a12eb64",
			wantID:    "a12eb64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveAgentID(claudeDir, projectPath, tt.sessionID, tt.prefix)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ResolveAgentID() expected error, got nil")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ResolveAgentID() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("ResolveAgentID() unexpected error: %v", err)
				return
			}

			if got != tt.wantID {
				t.Errorf("ResolveAgentID() = %v, want %v", got, tt.wantID)
			}
		})
	}
}

// TestFormatSessionAmbiguityError tests the session ambiguity error formatting.
func TestFormatSessionAmbiguityError(t *testing.T) {
	matches := []SessionMatch{
		{
			ID:          "cd2e9388-3108-40e5-b41b-79497cbb58b4",
			ProjectPath: "/Users/test/project",
			Created:     "2026-02-01T10:00:00Z",
			FirstPrompt: "test prompt 1",
		},
		{
			ID:          "cd2e4f21-9a14-4b29-8d3c-f5e8a9c1d7e2",
			ProjectPath: "/Users/test/project",
			Created:     "2026-02-01T09:00:00Z",
			FirstPrompt: "test prompt 2",
		},
	}

	err := formatSessionAmbiguityError("cd2e", matches)

	if err == nil {
		t.Fatal("formatSessionAmbiguityError() expected error, got nil")
	}

	errMsg := err.Error()

	// Check for required elements in error message
	required := []string{
		"ambiguous session ID prefix",
		"cd2e",
		"2 sessions",
		"cd2e9388-3108-40e5-b41b-79497cbb58b4",
		"cd2e4f21-9a14-4b29-8d3c-f5e8a9c1d7e2",
		"/Users/test/project",
		"2026-02-01T10:00:00Z",
		"test prompt 1",
		"Please provide more characters",
	}

	for _, req := range required {
		if !strings.Contains(errMsg, req) {
			t.Errorf("formatSessionAmbiguityError() error missing %q, got: %s", req, errMsg)
		}
	}
}

// TestFormatAgentAmbiguityError tests the agent ambiguity error formatting.
func TestFormatAgentAmbiguityError(t *testing.T) {
	sessionID := "cd2e9388-3108-40e5-b41b-79497cbb58b4"
	matches := []AgentMatch{
		{
			ID:         "a12eb64",
			SessionID:  sessionID,
			AgentType:  "explore",
			EntryCount: 42,
		},
		{
			ID:         "a12ef99",
			SessionID:  sessionID,
			AgentType:  "",
			EntryCount: 15,
		},
	}

	err := formatAgentAmbiguityError("a12", sessionID, matches)

	if err == nil {
		t.Fatal("formatAgentAmbiguityError() expected error, got nil")
	}

	errMsg := err.Error()

	// Check for required elements in error message
	required := []string{
		"ambiguous agent ID prefix",
		"a12",
		"2 agents",
		sessionID,
		"a12eb64",
		"a12ef99",
		"Type: explore",
		"Entries: 42",
		"Entries: 15",
		"Please provide more characters",
	}

	for _, req := range required {
		if !strings.Contains(errMsg, req) {
			t.Errorf("formatAgentAmbiguityError() error missing %q, got: %s", req, errMsg)
		}
	}
}
