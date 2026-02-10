package bookmarks

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/randlee/claude-history/pkg/agent"
	"github.com/randlee/claude-history/pkg/paths"
)

const (
	// MinNameLength is the minimum length for a bookmark name
	MinNameLength = 1
	// MaxNameLength is the maximum length for a bookmark name
	MaxNameLength = 64
)

var (
	// nameRegex validates bookmark name format: alphanumeric, hyphens, underscores only
	nameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

// ValidateBookmark validates all fields of a bookmark.
// Returns an error if any validation fails.
func ValidateBookmark(bookmark Bookmark) error {
	// Check required fields are non-empty
	if bookmark.Name == "" {
		return fmt.Errorf("bookmark name is required")
	}
	if bookmark.AgentID == "" {
		return fmt.Errorf("agent_id is required")
	}
	if bookmark.SessionID == "" {
		return fmt.Errorf("session_id is required")
	}

	// Validate name format
	if err := ValidateBookmarkName(bookmark.Name); err != nil {
		return err
	}

	// Validate project path exists if provided
	if bookmark.ProjectPath != "" {
		if _, err := os.Stat(bookmark.ProjectPath); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("project_path does not exist: %s", bookmark.ProjectPath)
			}
			return fmt.Errorf("cannot access project_path: %w", err)
		}
	}

	// Validate hostname format (basic validation - just check it's not empty and reasonable length)
	if bookmark.Hostname != "" && len(bookmark.Hostname) > 253 {
		return fmt.Errorf("hostname exceeds maximum length of 253 characters")
	}

	return nil
}

// ValidateBookmarkName validates the format of a bookmark name.
// Name must be alphanumeric with hyphens and underscores only,
// between 1 and 64 characters.
func ValidateBookmarkName(name string) error {
	if len(name) < MinNameLength {
		return fmt.Errorf("bookmark name must be at least %d character", MinNameLength)
	}
	if len(name) > MaxNameLength {
		return fmt.Errorf("bookmark name exceeds maximum length of %d characters", MaxNameLength)
	}
	if !nameRegex.MatchString(name) {
		return fmt.Errorf("bookmark name must contain only alphanumeric characters, hyphens, and underscores")
	}
	return nil
}

// ValidateAgentExists checks that an agent exists in the session history.
// It uses the existing agent discovery code to verify the agent file exists.
func ValidateAgentExists(agentID, sessionID, projectPath string) error {
	// Get the session directory path
	sessionDir, err := getSessionDir(projectPath, sessionID)
	if err != nil {
		return fmt.Errorf("cannot resolve session directory: %w", err)
	}

	// Check if session directory exists
	if _, err := os.Stat(sessionDir); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("session directory does not exist: %s", sessionDir)
		}
		return fmt.Errorf("cannot access session directory: %w", err)
	}

	// Try to get the agent
	agentInfo, err := agent.GetAgent(sessionDir, agentID)
	if err != nil {
		return fmt.Errorf("error checking agent: %w", err)
	}

	if agentInfo == nil {
		return fmt.Errorf("agent not found: %s (session: %s)", agentID, sessionID)
	}

	return nil
}

// getSessionDir constructs the session directory path.
// This is a helper function that uses the paths package to resolve the directory.
func getSessionDir(projectPath, sessionID string) (string, error) {
	// Get the project directory
	projectDir, err := paths.ProjectDir("", projectPath)
	if err != nil {
		return "", err
	}

	// Session directory is {projectDir}/{sessionID}
	return filepath.Join(projectDir, sessionID), nil
}
