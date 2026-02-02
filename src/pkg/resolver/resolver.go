// Package resolver provides session and agent ID prefix resolution.
package resolver

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/randlee/claude-history/pkg/paths"
)

var (
	// ErrEmptyPrefix is returned when an empty prefix is provided.
	ErrEmptyPrefix = errors.New("prefix cannot be empty")

	// ErrAmbiguousPrefix is returned when a prefix matches multiple IDs.
	ErrAmbiguousPrefix = errors.New("ambiguous prefix")

	// ErrNoMatch is returned when no IDs match the prefix.
	ErrNoMatch = errors.New("no matching IDs found")
)

// ResolveSessionID resolves a session ID prefix to a full session ID.
// If the prefix is already a full UUID, it returns it unchanged.
// If the prefix matches multiple sessions, it returns an error with details.
// If no sessions match, it returns ErrNoMatch.
func ResolveSessionID(projectDir, prefix string) (string, error) {
	// Validate input
	if prefix == "" {
		return "", ErrEmptyPrefix
	}

	// Check if projectDir exists
	if !paths.Exists(projectDir) {
		return "", &os.PathError{Op: "resolve", Path: projectDir, Err: os.ErrNotExist}
	}

	// List all session files
	sessionFiles, err := paths.ListSessionFiles(projectDir)
	if err != nil {
		return "", err
	}

	// Find matching sessions
	var matches []string
	for sessionID := range sessionFiles {
		if strings.HasPrefix(sessionID, prefix) {
			matches = append(matches, sessionID)
		}
	}

	// Handle results
	switch len(matches) {
	case 0:
		return "", &PrefixError{
			Prefix:  prefix,
			Message: "no sessions found matching prefix",
		}
	case 1:
		return matches[0], nil
	default:
		return "", &AmbiguousPrefixError{
			Prefix:   prefix,
			Matches:  matches,
			ItemType: "session",
		}
	}
}

// ResolveAgentID resolves an agent ID prefix to a full agent ID within a session.
// If the prefix is already a full ID, it returns it unchanged.
// If the prefix matches multiple agents, it returns an error with details.
func ResolveAgentID(projectDir, sessionID, prefix string) (string, error) {
	// Validate input
	if prefix == "" {
		return "", ErrEmptyPrefix
	}

	// Check session directory exists
	sessionDir := filepath.Join(projectDir, sessionID)
	if !paths.Exists(sessionDir) {
		return "", &os.PathError{Op: "resolve agent", Path: sessionDir, Err: os.ErrNotExist}
	}

	// List all agent files
	agentFiles, err := paths.ListAgentFiles(sessionDir)
	if err != nil {
		return "", err
	}

	// Find matching agents
	var matches []string
	for agentID := range agentFiles {
		if strings.HasPrefix(agentID, prefix) {
			matches = append(matches, agentID)
		}
	}

	// Handle results
	switch len(matches) {
	case 0:
		return "", &PrefixError{
			Prefix:  prefix,
			Message: "no agents found matching prefix",
		}
	case 1:
		return matches[0], nil
	default:
		return "", &AmbiguousPrefixError{
			Prefix:   prefix,
			Matches:  matches,
			ItemType: "agent",
		}
	}
}

// PrefixError represents an error when a prefix doesn't match anything.
type PrefixError struct {
	Prefix  string
	Message string
}

func (e *PrefixError) Error() string {
	return e.Message + ": " + e.Prefix
}

// AmbiguousPrefixError represents an error when a prefix matches multiple items.
type AmbiguousPrefixError struct {
	Prefix   string
	Matches  []string
	ItemType string // "session" or "agent"
}

func (e *AmbiguousPrefixError) Error() string {
	var sb strings.Builder
	sb.WriteString("ambiguous ")
	sb.WriteString(e.ItemType)
	sb.WriteString(" prefix '")
	sb.WriteString(e.Prefix)
	sb.WriteString("' matches ")
	sb.WriteString(strings.Join(e.Matches, ", "))
	sb.WriteString("; provide more characters to uniquely identify the ")
	sb.WriteString(e.ItemType)
	return sb.String()
}
