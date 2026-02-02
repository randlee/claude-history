// Package resolver provides git-style prefix matching for session and agent IDs.
package resolver

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/randlee/claude-history/internal/jsonl"
	"github.com/randlee/claude-history/pkg/agent"
	"github.com/randlee/claude-history/pkg/models"
	"github.com/randlee/claude-history/pkg/paths"
	"github.com/randlee/claude-history/pkg/session"
)

// SessionMatch represents a session that matches a prefix.
type SessionMatch struct {
	ID          string    // Full session ID
	Path        string    // Path to JSONL file
	FirstPrompt string    // Truncated to 60 chars
	Timestamp   time.Time // First entry timestamp
}

// AgentMatch represents an agent that matches a prefix.
type AgentMatch struct {
	ID          string    // Full agent ID
	Path        string    // Path to agent JSONL file
	Description string    // From Task spawn or first prompt
	Timestamp   time.Time // First entry timestamp
}

// ResolveSessionID finds a session by prefix in a project directory.
// If exactly 1 match is found, returns the full session ID.
// If 0 matches, returns error.
// If 2+ matches, returns detailed ambiguity error.
func ResolveSessionID(projectDir, prefix string) (string, error) {
	if prefix == "" {
		return "", fmt.Errorf("session ID prefix cannot be empty")
	}

	matches, err := findMatchingSessionIDs(projectDir, prefix)
	if err != nil {
		return "", err
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("no session found with prefix '%s'", prefix)
	}

	if len(matches) == 1 {
		return matches[0].ID, nil
	}

	// Multiple matches - return ambiguity error
	return "", formatSessionAmbiguityError(prefix, matches)
}

// ResolveAgentID finds an agent by prefix in a session.
// If exactly 1 match is found, returns the full agent ID.
// If 0 matches, returns error.
// If 2+ matches, returns detailed ambiguity error.
func ResolveAgentID(projectDir, sessionID, prefix string) (string, error) {
	if prefix == "" {
		return "", fmt.Errorf("agent ID prefix cannot be empty")
	}

	matches, err := findMatchingAgentIDs(projectDir, sessionID, prefix)
	if err != nil {
		return "", err
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("no agent found with prefix '%s' in session %s", prefix, sessionID)
	}

	if len(matches) == 1 {
		return matches[0].ID, nil
	}

	// Multiple matches - return ambiguity error
	return "", formatAgentAmbiguityError(prefix, matches)
}

// findMatchingSessionIDs finds all sessions in projectDir that start with prefix.
func findMatchingSessionIDs(projectDir, prefix string) ([]SessionMatch, error) {
	sessionFiles, err := paths.ListSessionFiles(projectDir)
	if err != nil {
		return nil, fmt.Errorf("failed to list session files: %w", err)
	}

	var matches []SessionMatch
	for sessionID, filePath := range sessionFiles {
		// Case-sensitive prefix matching (like git)
		if strings.HasPrefix(sessionID, prefix) {
			match := SessionMatch{
				ID:   sessionID,
				Path: filePath,
			}

			// Extract first prompt and timestamp
			var foundFirst bool
			err := session.ScanSession(filePath, func(entry models.ConversationEntry) error {
				if !foundFirst {
					// Get timestamp from first entry
					if ts, err := entry.GetTimestamp(); err == nil {
						match.Timestamp = ts
					}
					foundFirst = true
				}

				// Find first user message for prompt
				if match.FirstPrompt == "" && entry.IsUser() {
					prompt := entry.GetTextContent()
					if len(prompt) > 60 {
						match.FirstPrompt = prompt[:60] + "..."
					} else {
						match.FirstPrompt = prompt
					}
					// Stop scanning after we have both timestamp and prompt
					if !match.Timestamp.IsZero() {
						return session.StopScan
					}
				}

				return nil
			})

			if err != nil {
				// Skip sessions that can't be read
				continue
			}

			matches = append(matches, match)
		}
	}

	return matches, nil
}

// findMatchingAgentIDs finds all agents in a session that start with prefix.
func findMatchingAgentIDs(projectDir, sessionID, prefix string) ([]AgentMatch, error) {
	sessionDir := filepath.Join(projectDir, sessionID)

	agents, err := agent.DiscoverAgents(sessionDir)
	if err != nil {
		return nil, fmt.Errorf("failed to discover agents: %w", err)
	}

	// Get agent spawn descriptions from session file
	sessionFile := filepath.Join(projectDir, sessionID+".jsonl")
	spawnDescs := extractAgentSpawnDescriptions(sessionFile)

	var matches []AgentMatch
	for _, ag := range agents {
		// Case-sensitive prefix matching (like git)
		if strings.HasPrefix(ag.ID, prefix) {
			match := AgentMatch{
				ID:   ag.ID,
				Path: ag.FilePath,
			}

			// Try to get description from spawn operation first
			if desc, ok := spawnDescs[ag.ID]; ok {
				if len(desc) > 60 {
					match.Description = desc[:60] + "..."
				} else {
					match.Description = desc
				}
			}

			// Get timestamp and fallback description from first entry
			_ = jsonl.ScanInto(ag.FilePath, func(entry models.ConversationEntry) error {
				// Get timestamp
				if match.Timestamp.IsZero() {
					if ts, err := entry.GetTimestamp(); err == nil {
						match.Timestamp = ts
					}
				}

				// Fallback: use first user message if no spawn description
				if match.Description == "" && entry.IsUser() {
					prompt := entry.GetTextContent()
					if len(prompt) > 60 {
						match.Description = prompt[:60] + "..."
					} else {
						match.Description = prompt
					}
					return agent.StopIteration
				}

				// Stop after first entry if we have everything
				if !match.Timestamp.IsZero() && match.Description != "" {
					return agent.StopIteration
				}

				return nil
			})

			matches = append(matches, match)
		}
	}

	return matches, nil
}

// extractAgentSpawnDescriptions extracts descriptions from queue-operation entries.
func extractAgentSpawnDescriptions(sessionFile string) map[string]string {
	descriptions := make(map[string]string)

	_ = session.ScanSession(sessionFile, func(entry models.ConversationEntry) error {
		if entry.Type == models.EntryTypeQueueOperation && entry.AgentID != "" {
			// Extract description from message content
			text := entry.GetTextContent()
			if text != "" {
				descriptions[entry.AgentID] = text
			}
		}
		return nil
	})

	return descriptions
}

// formatSessionAmbiguityError formats a detailed error message for ambiguous session prefixes.
func formatSessionAmbiguityError(prefix string, matches []SessionMatch) error {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Error: ambiguous session ID prefix \"%s\" matches %d sessions:\n", prefix, len(matches)))

	for _, match := range matches {
		sb.WriteString(fmt.Sprintf("\n  %s\n", match.ID))
		sb.WriteString(fmt.Sprintf("    Date: %s\n", match.Timestamp.Format(time.RFC3339)))
		if match.FirstPrompt != "" {
			sb.WriteString(fmt.Sprintf("    Prompt: %s\n", match.FirstPrompt))
		}
	}

	sb.WriteString("\nPlease provide more characters to uniquely identify the session.")

	return fmt.Errorf("%s", sb.String())
}

// formatAgentAmbiguityError formats a detailed error message for ambiguous agent prefixes.
func formatAgentAmbiguityError(prefix string, matches []AgentMatch) error {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Error: ambiguous agent ID prefix \"%s\" matches %d agents:\n", prefix, len(matches)))

	for _, match := range matches {
		sb.WriteString(fmt.Sprintf("\n  %s\n", match.ID))
		sb.WriteString(fmt.Sprintf("    Date: %s\n", match.Timestamp.Format(time.RFC3339)))
		if match.Description != "" {
			sb.WriteString(fmt.Sprintf("    Description: %s\n", match.Description))
		}
	}

	sb.WriteString("\nPlease provide more characters to uniquely identify the agent.")

	return fmt.Errorf("%s", sb.String())
}
