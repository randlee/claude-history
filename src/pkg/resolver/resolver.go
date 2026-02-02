// Package resolver provides git-style prefix matching for session and agent IDs.
package resolver

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/randlee/claude-history/pkg/agent"
	"github.com/randlee/claude-history/pkg/paths"
	"github.com/randlee/claude-history/pkg/session"
)

// SessionMatch represents a matched session for ambiguity reporting.
type SessionMatch struct {
	ID          string
	ProjectPath string
	Created     string
	FirstPrompt string
}

// AgentMatch represents a matched agent for ambiguity reporting.
type AgentMatch struct {
	ID         string
	SessionID  string
	AgentType  string
	EntryCount int
}

// ResolveSessionID resolves a session ID prefix to a full session ID.
// Returns an error if the prefix is ambiguous or matches no sessions.
// claudeDir can be empty string to use default (~/.claude).
func ResolveSessionID(claudeDir, projectPath, prefix string) (string, error) {
	if prefix == "" {
		return "", fmt.Errorf("session ID prefix cannot be empty")
	}

	matches, err := findMatchingSessionIDs(claudeDir, projectPath, prefix)
	if err != nil {
		return "", fmt.Errorf("failed to list sessions: %w", err)
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("no session found with ID prefix %q", prefix)
	}

	if len(matches) == 1 {
		return matches[0].ID, nil
	}

	// Multiple matches - return ambiguity error
	return "", formatSessionAmbiguityError(prefix, matches)
}

// ResolveAgentID resolves an agent ID prefix to a full agent ID within a session.
// Returns an error if the prefix is ambiguous or matches no agents.
// claudeDir can be empty string to use default (~/.claude).
func ResolveAgentID(claudeDir, projectPath, sessionID, prefix string) (string, error) {
	if prefix == "" {
		return "", fmt.Errorf("agent ID prefix cannot be empty")
	}

	if sessionID == "" {
		return "", fmt.Errorf("session ID is required to resolve agent ID")
	}

	matches, err := findMatchingAgentIDs(claudeDir, projectPath, sessionID, prefix)
	if err != nil {
		return "", fmt.Errorf("failed to list agents: %w", err)
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("no agent found with ID prefix %q in session %s", prefix, sessionID)
	}

	if len(matches) == 1 {
		return matches[0].ID, nil
	}

	// Multiple matches - return ambiguity error
	return "", formatAgentAmbiguityError(prefix, sessionID, matches)
}

// findMatchingSessionIDs finds all sessions matching the given prefix.
func findMatchingSessionIDs(claudeDir, projectPath, prefix string) ([]SessionMatch, error) {
	// Get the encoded project directory
	projectDir, err := paths.ProjectDir(claudeDir, projectPath)
	if err != nil {
		return nil, err
	}

	// List all sessions in the project
	sessions, err := session.ListSessions(projectDir)
	if err != nil {
		return nil, err
	}

	var matches []SessionMatch
	for _, s := range sessions {
		if strings.HasPrefix(s.ID, prefix) {
			match := SessionMatch{
				ID:          s.ID,
				ProjectPath: projectPath,
				Created:     s.Created.Format("2006-01-02T15:04:05Z"),
				FirstPrompt: s.FirstPrompt,
			}
			matches = append(matches, match)
		}
	}

	// Sort by creation time (most recent first)
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Created > matches[j].Created
	})

	return matches, nil
}

// findMatchingAgentIDs finds all agents matching the given prefix within a session.
func findMatchingAgentIDs(claudeDir, projectPath, sessionID, prefix string) ([]AgentMatch, error) {
	// Get the encoded project directory
	projectDir, err := paths.ProjectDir(claudeDir, projectPath)
	if err != nil {
		return nil, err
	}

	// Get the session directory (project dir + session ID)
	sessionDir := filepath.Join(projectDir, sessionID)

	// Discover all agents in the session
	agents, err := agent.DiscoverAgents(sessionDir)
	if err != nil {
		return nil, err
	}

	var matches []AgentMatch
	for _, a := range agents {
		if strings.HasPrefix(a.ID, prefix) {
			match := AgentMatch{
				ID:         a.ID,
				SessionID:  a.SessionID,
				AgentType:  a.AgentType,
				EntryCount: a.EntryCount,
			}
			matches = append(matches, match)
		}
	}

	// Sort by ID for consistent output
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].ID < matches[j].ID
	})

	return matches, nil
}

// formatSessionAmbiguityError creates a user-friendly error message for ambiguous session prefixes.
func formatSessionAmbiguityError(prefix string, matches []SessionMatch) error {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("ambiguous session ID prefix %q matches %d sessions:\n\n", prefix, len(matches)))

	for _, m := range matches {
		sb.WriteString(fmt.Sprintf("  %s\n", m.ID))
		sb.WriteString(fmt.Sprintf("    Project: %s\n", m.ProjectPath))
		sb.WriteString(fmt.Sprintf("    Date: %s\n", m.Created))
		if m.FirstPrompt != "" {
			sb.WriteString(fmt.Sprintf("    Prompt: %s\n", m.FirstPrompt))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("Please provide more characters to uniquely identify the session.")
	return fmt.Errorf("%s", sb.String())
}

// formatAgentAmbiguityError creates a user-friendly error message for ambiguous agent prefixes.
func formatAgentAmbiguityError(prefix, sessionID string, matches []AgentMatch) error {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("ambiguous agent ID prefix %q matches %d agents in session %s:\n\n",
		prefix, len(matches), sessionID))

	for _, m := range matches {
		sb.WriteString(fmt.Sprintf("  %s\n", m.ID))
		if m.AgentType != "" {
			sb.WriteString(fmt.Sprintf("    Type: %s\n", m.AgentType))
		}
		sb.WriteString(fmt.Sprintf("    Entries: %d\n", m.EntryCount))
		sb.WriteString("\n")
	}

	sb.WriteString("Please provide more characters to uniquely identify the agent.")
	return fmt.Errorf("%s", sb.String())
}
