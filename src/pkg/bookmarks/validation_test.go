package bookmarks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestValidateBookmarkName tests bookmark name validation
func TestValidateBookmarkName(t *testing.T) {
	tests := []struct {
		name      string
		bookmarkName string
		wantErr   bool
		errContains string
	}{
		{
			name:         "valid simple name",
			bookmarkName: "my-bookmark",
			wantErr:      false,
		},
		{
			name:         "valid with underscores",
			bookmarkName: "my_bookmark_123",
			wantErr:      false,
		},
		{
			name:         "valid with hyphens",
			bookmarkName: "beads-architecture-expert",
			wantErr:      false,
		},
		{
			name:         "valid alphanumeric",
			bookmarkName: "bookmark123",
			wantErr:      false,
		},
		{
			name:         "valid single character",
			bookmarkName: "a",
			wantErr:      false,
		},
		{
			name:         "valid max length (64 chars)",
			bookmarkName: strings.Repeat("a", 64),
			wantErr:      false,
		},
		{
			name:         "empty name",
			bookmarkName: "",
			wantErr:      true,
			errContains:  "at least",
		},
		{
			name:         "too long",
			bookmarkName: strings.Repeat("a", 65),
			wantErr:      true,
			errContains:  "exceeds maximum length",
		},
		{
			name:         "contains spaces",
			bookmarkName: "my bookmark",
			wantErr:      true,
			errContains:  "alphanumeric",
		},
		{
			name:         "contains special chars",
			bookmarkName: "my@bookmark",
			wantErr:      true,
			errContains:  "alphanumeric",
		},
		{
			name:         "contains dots",
			bookmarkName: "my.bookmark",
			wantErr:      true,
			errContains:  "alphanumeric",
		},
		{
			name:         "contains slashes",
			bookmarkName: "my/bookmark",
			wantErr:      true,
			errContains:  "alphanumeric",
		},
		{
			name:         "unicode characters",
			bookmarkName: "my-书签",
			wantErr:      true,
			errContains:  "alphanumeric",
		},
		{
			name:         "emoji",
			bookmarkName: "bookmark-🔖",
			wantErr:      true,
			errContains:  "alphanumeric",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBookmarkName(tt.bookmarkName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBookmarkName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("ValidateBookmarkName() error = %v, want error containing %q", err, tt.errContains)
			}
		})
	}
}

// TestValidateBookmark tests full bookmark validation
func TestValidateBookmark(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	validPath := filepath.Join(tempDir, "test-project")
	if err := os.MkdirAll(validPath, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	now := time.Now()

	tests := []struct {
		name        string
		bookmark    Bookmark
		wantErr     bool
		errContains string
	}{
		{
			name: "valid bookmark with all fields",
			bookmark: Bookmark{
				BookmarkID:        "bmk-2026-02-09-001",
				Name:              "test-bookmark",
				Description:       "Test description",
				AgentID:           "agent-123",
				SessionID:         "session-456",
				ProjectPath:       validPath,
				OriginalTimestamp: now,
				Hostname:          "localhost",
				BookmarkedAt:      now,
				BookmarkedBy:      "testuser",
				Scope:             "global",
				Tags:              []string{"test"},
			},
			wantErr: false,
		},
		{
			name: "valid bookmark without optional fields",
			bookmark: Bookmark{
				Name:      "test-bookmark",
				AgentID:   "agent-123",
				SessionID: "session-456",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			bookmark: Bookmark{
				AgentID:   "agent-123",
				SessionID: "session-456",
			},
			wantErr:     true,
			errContains: "name is required",
		},
		{
			name: "missing agent_id",
			bookmark: Bookmark{
				Name:      "test-bookmark",
				SessionID: "session-456",
			},
			wantErr:     true,
			errContains: "agent_id is required",
		},
		{
			name: "missing session_id",
			bookmark: Bookmark{
				Name:    "test-bookmark",
				AgentID: "agent-123",
			},
			wantErr:     true,
			errContains: "session_id is required",
		},
		{
			name: "invalid name format",
			bookmark: Bookmark{
				Name:      "test bookmark",
				AgentID:   "agent-123",
				SessionID: "session-456",
			},
			wantErr:     true,
			errContains: "alphanumeric",
		},
		{
			name: "invalid project path",
			bookmark: Bookmark{
				Name:        "test-bookmark",
				AgentID:     "agent-123",
				SessionID:   "session-456",
				ProjectPath: filepath.Join(tempDir, "nonexistent"),
			},
			wantErr:     true,
			errContains: "does not exist",
		},
		{
			name: "hostname too long",
			bookmark: Bookmark{
				Name:      "test-bookmark",
				AgentID:   "agent-123",
				SessionID: "session-456",
				Hostname:  strings.Repeat("a", 254),
			},
			wantErr:     true,
			errContains: "hostname exceeds",
		},
		{
			name: "valid hostname at boundary",
			bookmark: Bookmark{
				Name:      "test-bookmark",
				AgentID:   "agent-123",
				SessionID: "session-456",
				Hostname:  strings.Repeat("a", 253),
			},
			wantErr: false,
		},
		{
			name: "empty hostname is valid",
			bookmark: Bookmark{
				Name:      "test-bookmark",
				AgentID:   "agent-123",
				SessionID: "session-456",
				Hostname:  "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBookmark(tt.bookmark)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBookmark() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("ValidateBookmark() error = %v, want error containing %q", err, tt.errContains)
			}
		})
	}
}

// TestValidateAgentExists tests agent existence validation
func TestValidateAgentExists(t *testing.T) {
	// This test requires a real session structure with agent files
	// We'll create a minimal test setup
	tempDir := t.TempDir()

	// Create a test project structure:
	// tempDir/
	//   .claude/
	//     projects/
	//       -test-project/
	//         session-123/
	//           subagents/
	//             agent-abc.jsonl

	claudeDir := filepath.Join(tempDir, ".claude")
	projectsDir := filepath.Join(claudeDir, "projects")
	encodedProject := "-test-project"
	projectDir := filepath.Join(projectsDir, encodedProject)
	sessionDir := filepath.Join(projectDir, "session-123")
	subagentsDir := filepath.Join(sessionDir, "subagents")

	// Create directories
	if err := os.MkdirAll(subagentsDir, 0755); err != nil {
		t.Fatalf("failed to create test directory structure: %v", err)
	}

	// Create an agent file
	agentFile := filepath.Join(subagentsDir, "agent-abc.jsonl")
	if err := os.WriteFile(agentFile, []byte(`{"type":"user","content":"test"}`+"\n"), 0644); err != nil {
		t.Fatalf("failed to create test agent file: %v", err)
	}

	// Set HOME to tempDir so paths.DefaultClaudeDir() uses our test .claude
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	tests := []struct {
		name        string
		agentID     string
		sessionID   string
		projectPath string
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid agent exists",
			agentID:     "abc",
			sessionID:   "session-123",
			projectPath: "/test/project",
			wantErr:     false,
		},
		{
			name:        "agent with prefix",
			agentID:     "agent-abc",
			sessionID:   "session-123",
			projectPath: "/test/project",
			wantErr:     false,
		},
		{
			name:        "agent does not exist",
			agentID:     "nonexistent",
			sessionID:   "session-123",
			projectPath: "/test/project",
			wantErr:     true,
			errContains: "agent not found",
		},
		{
			name:        "session does not exist",
			agentID:     "abc",
			sessionID:   "session-999",
			projectPath: "/test/project",
			wantErr:     true,
			errContains: "session directory does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAgentExists(tt.agentID, tt.sessionID, tt.projectPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAgentExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("ValidateAgentExists() error = %v, want error containing %q", err, tt.errContains)
			}
		})
	}
}

// TestGetSessionDir tests the helper function for resolving session directories
func TestGetSessionDir(t *testing.T) {
	tests := []struct {
		name        string
		projectPath string
		sessionID   string
		wantSuffix  string // Check the suffix since the full path depends on HOME
	}{
		{
			name:        "unix path",
			projectPath: "/test/project",
			sessionID:   "session-123",
			wantSuffix:  filepath.Join("-test-project", "session-123"),
		},
		{
			name:        "path with hyphens",
			projectPath: "/my-test/project",
			sessionID:   "session-456",
			wantSuffix:  filepath.Join("-my-test-project", "session-456"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getSessionDir(tt.projectPath, tt.sessionID)
			if err != nil {
				t.Errorf("getSessionDir() error = %v", err)
				return
			}
			if !strings.HasSuffix(got, tt.wantSuffix) {
				t.Errorf("getSessionDir() = %v, want suffix %v", got, tt.wantSuffix)
			}
		})
	}
}

// TestValidateBookmarkEdgeCases tests edge cases and boundary conditions
func TestValidateBookmarkEdgeCases(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		bookmark    Bookmark
		wantErr     bool
		errContains string
	}{
		{
			name: "very long description",
			bookmark: Bookmark{
				Name:        "test",
				AgentID:     "agent-123",
				SessionID:   "session-456",
				Description: strings.Repeat("a", 10000),
			},
			wantErr: false, // Description length is not limited
		},
		{
			name: "many tags",
			bookmark: Bookmark{
				Name:      "test",
				AgentID:   "agent-123",
				SessionID: "session-456",
				Tags:      make([]string, 100),
			},
			wantErr: false, // Tag count is not limited
		},
		{
			name: "zero timestamp",
			bookmark: Bookmark{
				Name:              "test",
				AgentID:           "agent-123",
				SessionID:         "session-456",
				OriginalTimestamp: time.Time{},
				BookmarkedAt:      time.Time{},
			},
			wantErr: false, // Zero timestamps are valid
		},
		{
			name: "future timestamp",
			bookmark: Bookmark{
				Name:              "test",
				AgentID:           "agent-123",
				SessionID:         "session-456",
				OriginalTimestamp: now.Add(24 * time.Hour),
			},
			wantErr: false, // Future timestamps are valid
		},
		{
			name: "name at boundary (64 chars)",
			bookmark: Bookmark{
				Name:      strings.Repeat("a", 64),
				AgentID:   "agent-123",
				SessionID: "session-456",
			},
			wantErr: false,
		},
		{
			name: "name over boundary (65 chars)",
			bookmark: Bookmark{
				Name:      strings.Repeat("a", 65),
				AgentID:   "agent-123",
				SessionID: "session-456",
			},
			wantErr:     true,
			errContains: "exceeds maximum length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBookmark(tt.bookmark)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBookmark() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("ValidateBookmark() error = %v, want error containing %q", err, tt.errContains)
			}
		})
	}
}
